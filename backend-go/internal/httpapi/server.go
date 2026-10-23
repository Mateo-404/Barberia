package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
	"github.com/Mateo-404/Barberia/backend-go/internal/config"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/administradores"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/clientes"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/estadisticas"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/servicios"
	settingsmod "github.com/Mateo-404/Barberia/backend-go/internal/modules/settings"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/turnos"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/Mateo-404/Barberia/backend-go/internal/tenant"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server agrupa el handler HTTP completo.
type Server struct {
	cfg               *config.Config
	logger            *slog.Logger
	pool              *pgxpool.Pool
	jwt               *auth.JWTManager
	limiter           *auth.LoginLimiter
	escrituraLimiter  *auth.LoginLimiter
	trustProxyHeaders bool
	resolver          TenantResolver
	clientes          *clientes.Service
	servicios         *servicios.Service
	admins            *administradores.Service
	turnos            *turnos.Service
	stats             *estadisticas.Service
	settings          *settingsmod.Service
	now               func() time.Time
}

// TenantResolver resuelve el tenant por Host de la request.
type TenantResolver interface {
	Resolve(ctx context.Context, host string) (tenant.Tenant, error)
}

// Deps son las dependencias del servidor HTTP.
type Deps struct {
	Cfg      *config.Config
	Logger   *slog.Logger
	Pool     *pgxpool.Pool
	JWT      *auth.JWTManager
	Resolver TenantResolver
}

// NewServer construye el servidor con todos los módulos.
func NewServer(d Deps) *Server {
	q := sqlc.New(d.Pool)
	return &Server{
		cfg: d.Cfg, logger: d.Logger, pool: d.Pool, jwt: d.JWT,
		limiter:           auth.NewLoginLimiter(d.Cfg.LoginRateLimitPerMinute),
		escrituraLimiter:  auth.NewLoginLimiter(d.Cfg.TurnoRateLimitPerMinute),
		trustProxyHeaders: d.Cfg.TrustProxyHeaders,
		resolver:          d.Resolver,
		clientes:          clientes.New(d.Pool, q),
		servicios:         servicios.New(d.Pool, q),
		admins:            administradores.New(d.Pool, q, d.JWT, d.Logger),
		turnos:            turnos.New(d.Pool, q),
		stats:             estadisticas.New(d.Pool, q),
		settings:          settingsmod.New(d.Pool, q),
		now:               time.Now,
	}
}

// Handler monta el router chi con el contrato del backend Java.
func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(s.middlewareRecover)
	r.Use(s.middlewareRequestID)
	r.Use(s.middlewareLog)
	r.Use(s.middlewareHeaders)
	r.Use(s.middlewareCORS)
	r.Use(s.middlewareAuth)
	r.Use(s.middlewareAdminRole)
	r.Use(s.middlewareLimitarEscrituras)

	r.NotFound(func(w http.ResponseWriter, req *http.Request) { notFoundRecurso(w, req) })
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) { methodNotAllowed(w, req) })

	r.Get("/ping", s.hPing)
	r.Get("/healthz", s.hHealthz)
	r.Get("/readyz", s.hReadyz)

	r.Route("/administradores", func(r chi.Router) {
		r.Get("/", s.hAdminList)
		r.Post("/", s.hAdminCreate)
		r.Post("/all", s.hAdminCreateAll)
		r.Post("/login", s.hAdminLogin)
		r.Get("/exists/{id}", s.hAdminExists)
		r.Get("/{page}/{size}", s.hAdminPage)
		r.Get("/{id}", s.hAdminGet)
		r.Patch("/{id}", s.hAdminPatch)
		r.Delete("/{id}", s.hAdminDelete)
	})

	r.Route("/clientes", func(r chi.Router) {
		r.Get("/", s.hClienteList)
		r.Post("/", s.hClienteCreate)
		r.Post("/all", s.hClienteCreateAll)
		r.Get("/exists/{telefono}", s.hClienteExists)
		r.Get("/{page}/{size}", s.hClientePage)
		r.Get("/{telefono}", s.hClienteGet)
		r.Patch("/{telefono}", s.hClientePatch)
		r.Delete("/{telefono}", s.hClienteDelete)
	})

	r.Route("/servicios", func(r chi.Router) {
		r.Get("/", s.hServicioList)
		r.Post("/", s.hServicioCreate)
		r.Post("/all", s.hServicioCreateAll)
		r.Get("/exists/{id}", s.hServicioExists)
		r.Get("/{page}/{size}", s.hServicioPage)
		r.Get("/{id}", s.hServicioGet)
		r.Patch("/{id}", s.hServicioPatch)
		r.Delete("/{id}", s.hServicioDelete)
	})

	r.Route("/turnos", func(r chi.Router) {
		r.Get("/", s.hTurnoList)
		r.Post("/", s.hTurnoCreate)
		r.Post("/all", s.hTurnoCreateAll)
		r.Get("/findDateTimes", s.hTurnoFindDateTimes)
		r.Get("/exists/{id}", s.hTurnoExists)
		r.Get("/{page}/{size}", s.hTurnoPage)
		r.Get("/{id}", s.hTurnoGet)
		r.Patch("/{id}", s.hTurnoPatch)
		r.Delete("/{id}", s.hTurnoDelete)
	})

	r.Get("/estadisticas/panel", s.hEstadisticasPanel)

	// Settings API
	r.Route("/api/admin/settings", func(r chi.Router) {
		r.Get("/", s.hAdminSettingsGet)
		r.Put("/draft", s.hAdminSettingsDraft)
		r.Post("/publish", s.hAdminSettingsPublish)
		r.Get("/versions", s.hAdminSettingsVersions)
		r.Post("/rollback", s.hAdminSettingsRollback)
		r.Get("/preview/theme.css", s.hAdminSettingsPreview)
	})
	r.Route("/api/public", func(r chi.Router) {
		r.Get("/settings", s.hPublicSettings)
		r.Get("/theme.css", s.hPublicThemeCSS)
	})

	return r
}

// middlewareLimitarEscrituras corta temprano (antes de tocar la BD) los POST
// públicos al endpoint de turnos que exceden el límite por IP.
func (s *Server) middlewareLimitarEscrituras(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.limitaEscritura(w, r) {
			return
		}
		next.ServeHTTP(w, r)
	})
}
