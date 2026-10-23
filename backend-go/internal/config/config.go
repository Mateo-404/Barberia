// Package config loads and validates configuration from environment
// variables. Critical values are fail-closed: when missing, the app does not
// start.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the effective application configuration.
type Config struct {
	// HTTPAddr is the HTTP listen address.
	HTTPAddr string
	// DatabaseURL is the PostgreSQL connection string (sslmode included).
	DatabaseURL string
	// JWTSecret is the HS256 key. Required, at least 32 bytes.
	JWTSecret string
	// JWTExpiry is the lifetime of issued tokens (default 24h).
	JWTExpiry time.Duration
	// GoogleClientID / Secret / RedirectURL enable "Sign in with Google".
	// When ClientID is empty, Google login is disabled.
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	// CORSAllowOrigins lists the allowed origins (empty = no CORS).
	CORSAllowOrigins []string
	// DefaultTenantSlug is the tenant for unknown or local Hosts.
	DefaultTenantSlug string
	// AutoMigrate applies migrations at startup.
	AutoMigrate bool
	// AdminSeed enables creating the initial admin.
	AdminSeed          bool
	AdminSeedEmail     string
	AdminSeedPassword  string
	AdminSeedFirstName string
	AdminSeedLastName  string
	// LoginRateLimitPerMinute is the max login attempts per IP.
	LoginRateLimitPerMinute int
	// BookingRateLimitPerMinute is the max public bookings
	// (POST /api/appointments) per IP.
	BookingRateLimitPerMinute int
	// TrustProxyHeaders trusts X-Forwarded-For (behind a trusted proxy only);
	// false by default so the rate limit cannot be dodged.
	TrustProxyHeaders bool
	// PoolMaxConns / ConnTimeout / StatementTimeout tune the database pool.
	PoolMaxConns     int32
	ConnTimeout      time.Duration
	StatementTimeout time.Duration
	MaxRequestBody   int64
	// AppEnv is "dev" or anything else (treated as production).
	AppEnv string
}

// Load reads the configuration from the environment and fails when anything
// critical is missing or invalid.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:                  getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:               strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:                 strings.TrimSpace(os.Getenv("JWT_SECRET")),
		JWTExpiry:                 24 * time.Hour,
		DefaultTenantSlug:         getEnv("DEFAULT_TENANT_SLUG", "default"),
		AdminSeedEmail:            getEnv("ADMIN_SEED_EMAIL", "admin@admin.com"),
		AdminSeedPassword:         os.Getenv("ADMIN_SEED_PASSWORD"),
		AdminSeedFirstName:        getEnv("ADMIN_SEED_FIRST_NAME", "Admin"),
		AdminSeedLastName:         getEnv("ADMIN_SEED_LAST_NAME", "Barber"),
		LoginRateLimitPerMinute:   10,
		BookingRateLimitPerMinute: 30,
		PoolMaxConns:              10,
		ConnTimeout:               5 * time.Second,
		StatementTimeout:          10 * time.Second,
		MaxRequestBody:            1 << 20, // 1 MiB
		// Fail-closed: without an explicit APP_ENV=dev the RLS role check runs.
		AppEnv: getEnv("APP_ENV", "prod"),
	}

	var errs []string
	invalid := func(key, value, hint string) {
		errs = append(errs, fmt.Sprintf("invalid %s: %q%s", key, value, hint))
	}

	cfg.GoogleClientID = strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	cfg.GoogleClientSecret = strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET"))
	cfg.GoogleRedirectURL = strings.TrimSpace(os.Getenv("GOOGLE_REDIRECT_URL"))
	if cfg.GoogleRedirectURL == "" {
		cfg.GoogleRedirectURL = "http://localhost:5173/auth/google/callback"
	}
	if cfg.DatabaseURL == "" {
		errs = append(errs, "DATABASE_URL is required (e.g. postgres://user:pass@localhost:5432/barber?sslmode=disable)")
	}
	if len(cfg.JWTSecret) < 32 {
		errs = append(errs, "JWT_SECRET is required and must be at least 32 bytes (generate one with: openssl rand -base64 32)")
	}

	if v := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS")); v != "" {
		for _, o := range strings.Split(v, ",") {
			if o = strings.TrimSpace(o); o != "" {
				cfg.CORSAllowOrigins = append(cfg.CORSAllowOrigins, o)
			}
		}
	}

	// parseBool reports whether key was set.
	parseBool := func(key string, dst *bool) bool {
		v := os.Getenv(key)
		if v == "" {
			return false
		}
		b, err := strconv.ParseBool(v)
		if err != nil {
			invalid(key, v, "")
		}
		*dst = b
		return true
	}
	parsePositiveInt := func(key string, dst *int) {
		if v := os.Getenv(key); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n <= 0 {
				invalid(key, v, "")
				return
			}
			*dst = n
		}
	}
	parseDuration := func(key string, dst *time.Duration, allowZero bool) {
		if v := os.Getenv(key); v != "" {
			d, err := time.ParseDuration(v)
			if err != nil || d < 0 || (d == 0 && !allowZero) {
				invalid(key, v, " (use Go duration format, e.g. 5s)")
				return
			}
			*dst = d
		}
	}

	parseBool("AUTO_MIGRATE", &cfg.AutoMigrate)
	if !parseBool("ADMIN_SEED", &cfg.AdminSeed) {
		// A configured password enables the seed implicitly.
		cfg.AdminSeed = cfg.AdminSeedPassword != ""
	}
	if cfg.AdminSeed && cfg.AdminSeedPassword == "" {
		errs = append(errs, "ADMIN_SEED_PASSWORD is required when ADMIN_SEED is enabled")
	}
	parsePositiveInt("LOGIN_RATE_LIMIT_PER_MINUTE", &cfg.LoginRateLimitPerMinute)
	parsePositiveInt("BOOKING_RATE_LIMIT_PER_MINUTE", &cfg.BookingRateLimitPerMinute)
	parseBool("TRUST_PROXY_HEADERS", &cfg.TrustProxyHeaders)

	if v := os.Getenv("POOL_MAX_CONNS"); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil || n <= 0 {
			invalid("POOL_MAX_CONNS", v, "")
		} else {
			cfg.PoolMaxConns = int32(n)
		}
	}
	parseDuration("CONN_TIMEOUT", &cfg.ConnTimeout, false)
	if v := os.Getenv("MAX_REQUEST_BODY"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n <= 0 {
			invalid("MAX_REQUEST_BODY", v, "")
		} else {
			cfg.MaxRequestBody = n
		}
	}
	parseDuration("STATEMENT_TIMEOUT", &cfg.StatementTimeout, true)

	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid configuration: %s", strings.Join(errs, "; "))
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
