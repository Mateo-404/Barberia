// Package httpapi is the HTTP layer: routing, middleware, request
// validation and the JSON contract described in api/openapi.json.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
	"github.com/Mateo-404/Barberia/backend-go/internal/config"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/admins"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/appointments"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/customers"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/services"
	settingsmod "github.com/Mateo-404/Barberia/backend-go/internal/modules/settings"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/stats"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/Mateo-404/Barberia/backend-go/internal/tenant"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server holds the complete HTTP handler and its dependencies.
type Server struct {
	cfg               *config.Config
	logger            *slog.Logger
	pool              *pgxpool.Pool
	jwt               *auth.JWTManager
	loginLimiter      *auth.LoginLimiter
	bookingLimiter    *auth.LoginLimiter
	trustProxyHeaders bool
	resolver          TenantResolver
	customers         *customers.Service
	services          *services.Service
	admins            *admins.Service
	appointments      *appointments.Service
	stats             *stats.Service
	settings          *settingsmod.Service
	now               func() time.Time
}

// TenantResolver resolves the tenant of a request Host.
type TenantResolver interface {
	Resolve(ctx context.Context, host string) (tenant.Tenant, error)
}

// Deps are the server dependencies.
type Deps struct {
	Cfg      *config.Config
	Logger   *slog.Logger
	Pool     *pgxpool.Pool
	JWT      *auth.JWTManager
	Resolver TenantResolver
}

// NewServer builds the server with every module.
func NewServer(d Deps) *Server {
	q := sqlc.New(d.Pool)
	return &Server{
		cfg: d.Cfg, logger: d.Logger, pool: d.Pool, jwt: d.JWT,
		loginLimiter:      auth.NewLoginLimiter(d.Cfg.LoginRateLimitPerMinute),
		bookingLimiter:    auth.NewLoginLimiter(d.Cfg.BookingRateLimitPerMinute),
		trustProxyHeaders: d.Cfg.TrustProxyHeaders,
		resolver:          d.Resolver,
		customers:         customers.New(d.Pool, q),
		services:          services.New(d.Pool, q),
		admins:            admins.New(d.Pool, q, d.JWT, d.Logger, d.Cfg.JWTSecret),
		appointments:      appointments.New(d.Pool, q),
		stats:             stats.New(d.Pool, q),
		settings:          settingsmod.New(d.Pool, q),
		now:               time.Now,
	}
}

// Handler mounts the chi router.
func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(s.middlewareRecover)
	r.Use(s.middlewareRequestID)
	r.Use(s.middlewareLog)
	r.Use(s.middlewareHeaders)
	r.Use(s.middlewareCORS)
	r.Use(s.middlewareAuth)
	r.Use(s.middlewareBookingRateLimit)

	r.NotFound(routeNotFound)
	r.MethodNotAllowed(methodNotAllowed)

	r.Get("/healthz", s.handleHealthz)
	r.Get("/readyz", s.handleReadyz)

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/login", s.handleLogin)
		r.Get("/auth/providers", s.handleAuthProviders)
		r.Post("/auth/google/start", s.handleGoogleStart)
		r.Post("/auth/google/callback", s.handleGoogleCallback)
		r.Post("/auth/totp/challenge", s.handleTOTPChallenge)
		r.Route("/auth/totp", func(r chi.Router) {
			r.Use(requireRole(auth.RoleOwner, auth.RoleAdmin, auth.RoleStaff))
			r.Post("/setup", s.handleTOTPSetup)
			r.Post("/verify", s.handleTOTPVerify)
			r.Post("/disable", s.handleTOTPDisable)
		})

		r.Route("/appointments", func(r chi.Router) {
			r.Get("/", s.handleAppointmentList)
			r.Post("/", s.handleAppointmentCreate)
			r.Get("/paged", s.handleAppointmentPage)
			r.Get("/booked-slots", s.handleBookedSlots)
			r.Get("/{id}", s.handleAppointmentGet)
			r.Patch("/{id}", s.handleAppointmentPatch)
			r.Delete("/{id}", s.handleAppointmentDelete)
		})

		r.Route("/customers", func(r chi.Router) {
			r.Get("/", s.handleCustomerList)
			r.Post("/", s.handleCustomerCreate)
			r.Get("/paged", s.handleCustomerPage)
			r.Get("/{phone}", s.handleCustomerGet)
			r.Patch("/{phone}", s.handleCustomerPatch)
			r.Delete("/{phone}", s.handleCustomerDelete)
		})

		r.Route("/services", func(r chi.Router) {
			r.Get("/", s.handleServiceList)
			r.Post("/", s.handleServiceCreate)
			r.Get("/paged", s.handleServicePage)
			r.Get("/{id}", s.handleServiceGet)
			r.Patch("/{id}", s.handleServicePatch)
			r.Delete("/{id}", s.handleServiceDelete)
		})

		// Managing admins is reserved to the owner (privilege escalation guard).
		r.Route("/admins", func(r chi.Router) {
			r.Use(requireRole(auth.RoleOwner))
			r.Get("/", s.handleAdminList)
			r.Post("/", s.handleAdminCreate)
			r.Get("/paged", s.handleAdminPage)
			r.Get("/{id}", s.handleAdminGet)
			r.Patch("/{id}", s.handleAdminPatch)
			r.Delete("/{id}", s.handleAdminDelete)
		})

		r.Get("/stats/dashboard", s.handleDashboard)

		// Site settings: owner and admin only.
		r.Route("/admin", func(r chi.Router) {
			r.Use(requireRole(auth.RoleOwner, auth.RoleAdmin))
			r.Route("/settings", func(r chi.Router) {
				r.Get("/", s.handleSettingsGet)
				r.Put("/draft", s.handleSettingsDraft)
				r.Post("/publish", s.handleSettingsPublish)
				r.Get("/versions", s.handleSettingsVersions)
				r.Post("/rollback", s.handleSettingsRollback)
				r.Patch("/preferences", s.handleSettingsPreferences)
			})
			r.Route("/theme", func(r chi.Router) {
				r.Post("/preview", s.handleThemePreview)
				r.Get("/presets", s.handleThemePresets)
			})
			r.Get("/business", s.handleBusinessGet)
			r.Put("/business", s.handleBusinessPut)
		})

		r.Route("/public", func(r chi.Router) {
			r.Get("/settings", s.handlePublicSettings)
			r.Get("/theme.css", s.handlePublicThemeCSS)
		})
	})

	return r
}
