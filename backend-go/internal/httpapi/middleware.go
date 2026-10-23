package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
	"github.com/Mateo-404/Barberia/backend-go/internal/tenant"
)

type ctxKey int

const (
	ctxKeyClaims ctxKey = iota
	ctxKeyTenant
	ctxKeyAdminTenantID
	ctxKeyRequestID
)

// tenantDe r devuelve el tenant resuelto (Host o JWT).
func tenantDe(r *http.Request) (tenant.Tenant, bool) {
	t, ok := r.Context().Value(ctxKeyTenant).(tenant.Tenant)
	return t, ok
}

// esPublica indica si la ruta no requiere autenticación (SecurityConfig Java).
func esPublica(r *http.Request) bool {
	p := strings.TrimSuffix(r.URL.Path, "/")
	if p == "" {
		p = "/"
	}
	switch {
	case r.Method == http.MethodOptions:
		return true
	case p == "/ping" || p == "/healthz" || p == "/readyz":
		return true
	case r.Method == http.MethodPost && p == "/turnos":
		return true
	case r.Method == http.MethodGet && p == "/turnos/findDateTimes":
		return true
	case r.Method == http.MethodGet && (p == "/servicios" || strings.HasPrefix(p, "/servicios/")):
		return true
	case r.Method == http.MethodPost && p == "/administradores/login":
		return true
	case strings.HasPrefix(p, "/api/public"):
		return true
	}
	return false
}

// middlewareAuth exige/valida Bearer como el filtro de Spring: un token Bearer
// inválido rechaza incluso en rutas públicas; sin token en ruta protegida
// devuelve 401 sin instance.
func (s *Server) middlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		var claims *auth.Claims
		if strings.HasPrefix(authz, "Bearer ") {
			c, err := s.jwt.Parse(strings.TrimPrefix(authz, "Bearer "))
			if err != nil {
				unauthorizedTokenInvalido(w)
				return
			}
			claims = c
		} else if !esPublica(r) {
			unauthorizedSinToken(w)
			return
		}

		// Las respuestas autenticadas o de credenciales (login) no deben
		// quedar en cachés intermedias.
		if claims != nil || (r.Method == http.MethodPost && r.URL.Path == "/administradores/login") {
			w.Header().Set("Cache-Control", "no-store")
		}

		ctx := r.Context()
		if claims != nil {
			// Tokens legados sin claim de tenant: rechazar con 401 (re-login), sin fallback
			if claims.TenantID == 0 {
				unauthorizedTokenInvalido(w)
				return
			}
			ctx = context.WithValue(ctx, ctxKeyClaims, claims)
			resolved, err := s.resolver.Resolve(ctx, r.Host)
			if err != nil {
				// tenant desconocido/suspendido -> 404 genérico
				notFoundRecurso(w, r)
				return
			}
			// JWT tenant != tenant del host -> 403
			if resolved.ID != claims.TenantID {
				forbidden(w, r)
				return
			}
			ctx = context.WithValue(ctx, ctxKeyTenant, resolved)
			ctx = tenant.WithID(ctx, resolved.ID)
		} else {
			resolved, err := s.resolver.Resolve(ctx, r.Host)
			if err != nil {
				// desconocido/suspendido -> 404 genérico
				notFoundRecurso(w, r)
				return
			}
			ctx = context.WithValue(ctx, ctxKeyTenant, resolved)
			ctx = tenant.WithID(ctx, resolved.ID)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// nuevoRequestID genera un id corto para correlación de logs.
func nuevoRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "0000000000000000"
	}
	return hex.EncodeToString(b)
}

// middlewareRequestID asigna un request_id al context y lo devuelve en el header.
func (s *Server) middlewareRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = nuevoRequestID()
		}
		w.Header().Set("X-Request-ID", rid)
		ctx := context.WithValue(r.Context(), ctxKeyRequestID, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		tenantID, _ := r.Context().Value(ctxKeyTenant).(tenant.Tenant)
		rid, _ := r.Context().Value(ctxKeyRequestID).(string)
		s.logger.Info("http",
			"method", r.Method, "path", r.URL.Path, "status", sw.status,
			"duracion_ms", time.Since(inicio).Milliseconds(),
			"remote", s.clienteIP(r), "tenant_id", tenantID.ID,
			"request_id", rid)
	})
}

// limitaEscritura aplica el rate limit a las escrituras públicas (POST /turnos):
// sin control, cualquiera podría crear clientes y turnos masivamente.
func (s *Server) limitaEscritura(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost && strings.TrimSuffix(r.URL.Path, "/") == "/turnos" {
		if !s.escrituraLimiter.Allow(s.clienteIP(r)) {
			rateLimited(w, r)
			return false
		}
	}
	return true
}

// middlewareHeaders agrega cabeceras de seguridad a todas las respuestas
func (s *Server) middlewareHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-site")
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) middlewareRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.logger.Error("panic recuperado", "panic", rec, "path", r.URL.Path,
					"stack", string(debug.Stack()))
				internalServerError(w, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// middlewareCORS replica el comportamiento por perfil de Java:
// orígenes configurados exactos con credenciales.
func (s *Server) middlewareCORS(next http.Handler) http.Handler {
	permitidos := make(map[string]bool, len(s.cfg.CORSAllowOrigins))
	for _, o := range s.cfg.CORSAllowOrigins {
		permitidos[strings.TrimSuffix(o, "/")] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && permitidos[strings.TrimSuffix(origin, "/")] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Max-Age", "3600")
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clienteIP devuelve la IP del cliente para el rate limit. Por defecto es el
// peer directo (RemoteAddr) y NO se confía en X-Forwarded-For: un cliente
// podría falsificarla para evadir el límite. Solo cuando TRUST_PROXY_HEADERS
// está activo (despliegue detrás de un proxy de confianza) se usa el primer
// valor de X-Forwarded-For.
func (s *Server) clienteIP(r *http.Request) string {
	if s.trustProxyHeaders {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if i := strings.IndexByte(xff, ','); i >= 0 {
				xff = xff[:i]
			}
			if ip := strings.TrimSpace(xff); ip != "" {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
