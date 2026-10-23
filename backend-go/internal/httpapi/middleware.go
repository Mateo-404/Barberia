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
	ctxKeyRequestID
)

const loginPath = "/api/auth/login"

// tenantFrom returns the tenant resolved from the Host or the JWT.
func tenantFrom(r *http.Request) (tenant.Tenant, bool) {
	t, ok := r.Context().Value(ctxKeyTenant).(tenant.Tenant)
	return t, ok
}

// isPublic reports whether the route can be called without a token.
func isPublic(r *http.Request) bool {
	p := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case r.Method == http.MethodOptions:
		return true
	case p == "/healthz" || p == "/readyz":
		return true
	case r.Method == http.MethodPost && p == "/api/appointments":
		return true
	case r.Method == http.MethodGet && p == "/api/appointments/booked-slots":
		return true
	case r.Method == http.MethodGet && (p == "/api/services" || strings.HasPrefix(p, "/api/services/")):
		return true
	case r.Method == http.MethodPost && p == loginPath:
		return true
	case strings.HasPrefix(p, "/api/public/"):
		return true
	}
	return false
}

// middlewareAuth validates the Bearer token. An invalid token is rejected
// even on public routes; a protected route without a token answers 401.
// The tenant comes from the JWT on authenticated requests and from the Host
// otherwise — never from a client-supplied header.
func (s *Server) middlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var claims *auth.Claims
		if authz := r.Header.Get("Authorization"); strings.HasPrefix(authz, "Bearer ") {
			c, err := s.jwt.Parse(strings.TrimPrefix(authz, "Bearer "))
			if err != nil || c.TenantID == 0 {
				invalidToken(w, r)
				return
			}
			claims = c
		} else if !isPublic(r) {
			unauthorized(w, r)
			return
		}

		// Authenticated and credential responses must not be cached.
		if claims != nil || r.URL.Path == loginPath {
			w.Header().Set("Cache-Control", "no-store")
		}

		ctx := r.Context()
		resolved, err := s.resolver.Resolve(ctx, r.Host)
		if err != nil {
			// Unknown or suspended tenant: generic 404.
			routeNotFound(w, r)
			return
		}
		if claims != nil {
			if resolved.ID != claims.TenantID {
				forbidden(w, r)
				return
			}
			ctx = context.WithValue(ctx, ctxKeyClaims, claims)
		}
		ctx = context.WithValue(ctx, ctxKeyTenant, resolved)
		ctx = tenant.WithID(ctx, resolved.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "0000000000000000"
	}
	return hex.EncodeToString(b)
}

// middlewareRequestID sets a correlation id in the context and the response.
func (s *Server) middlewareRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = newRequestID()
		}
		w.Header().Set("X-Request-ID", rid)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, rid)))
	})
}

func (s *Server) middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		t, _ := tenantFrom(r)
		s.logger.Info("http",
			"method", r.Method, "path", r.URL.Path, "status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote", s.clientIP(r), "tenant_id", t.ID,
			"request_id", requestID(r))
	})
}

// middlewareBookingRateLimit stops public bookings over the per-IP limit
// before they touch the database: otherwise anyone could create customers
// and appointments in bulk.
func (s *Server) middlewareBookingRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.TrimSuffix(r.URL.Path, "/") == "/api/appointments" &&
			!s.bookingLimiter.Allow(s.clientIP(r)) {
			rateLimited(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// middlewareHeaders adds security headers to every response.
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
				s.logger.Error("panic recovered", "panic", rec, "path", r.URL.Path,
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

// middlewareCORS allows the configured exact origins, with credentials.
func (s *Server) middlewareCORS(next http.Handler) http.Handler {
	allowed := make(map[string]bool, len(s.cfg.CORSAllowOrigins))
	for _, o := range s.cfg.CORSAllowOrigins {
		allowed[strings.TrimSuffix(o, "/")] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" && allowed[strings.TrimSuffix(origin, "/")] {
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

// clientIP returns the client IP used for rate limiting. By default it is
// the direct peer (RemoteAddr): X-Forwarded-For is only trusted with
// TRUST_PROXY_HEADERS, because a client could forge it to dodge the limit.
func (s *Server) clientIP(r *http.Request) string {
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
