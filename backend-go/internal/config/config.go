// Package config carga y valida la configuración desde variables de entorno.
// Todos los valores críticos son fail-closed: si faltan, la app no arranca.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config es la configuración efectiva de la aplicación.
type Config struct {
	// HTTPAddr es la dirección de escucha del servidor HTTP.
	HTTPAddr string
	// DatabaseURL es la cadena de conexión PostgreSQL (sslmode incluido).
	DatabaseURL string
	// JWTSecret es la clave HS256. Obligatoria, mínimo 32 bytes.
	JWTSecret string
	// JWTExpiry es la validez de los tokens emitidos (default 24h).
	JWTExpiry time.Duration
	// CORSAllowOrigins lista de orígenes permitidos (vacío = sin CORS).
	CORSAllowOrigins []string
	// DefaultTenantSlug es el tenant para rutas públicas o Host desconocido.
	DefaultTenantSlug string
	// AutoMigrate aplica migraciones al arrancar.
	AutoMigrate bool
	// AdminSeed habilita la creación del admin inicial.
	AdminSeed         bool
	AdminSeedEmail    string
	AdminSeedPassword string
	AdminSeedNombre   string
	AdminSeedApellido string
	// LoginRateLimitPerMinute es el máximo de intentos de login por IP.
	LoginRateLimitPerMinute int
	// TurnoRateLimitPerMinute es el máximo de POST /turnos por IP (endpoint
	// público de reserva).
	TurnoRateLimitPerMinute int
	// TrustProxyHeaders indica si se confía en X-Forwarded-For (proxy de
	// confianza); por defecto false para no permitir evadir el rate limit.
	TrustProxyHeaders bool
	// PoolMaxConns / ConnTimeout controlan el pool de conexiones.
	PoolMaxConns     int32
	ConnTimeout      time.Duration
	StatementTimeout time.Duration
	MaxRequestBody   int64
	AppEnv           string
}

// Load lee la configuración del entorno. Devuelve error si falta algo crítico.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:                getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:             strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:               strings.TrimSpace(os.Getenv("JWT_SECRET")),
		JWTExpiry:               24 * time.Hour,
		DefaultTenantSlug:       getEnv("DEFAULT_TENANT_SLUG", "default"),
		AdminSeedEmail:          getEnv("ADMIN_SEED_EMAIL", "admin@admin.com"),
		AdminSeedPassword:       os.Getenv("ADMIN_SEED_PASSWORD"),
		AdminSeedNombre:         getEnv("ADMIN_SEED_NOMBRE", "Admin"),
		AdminSeedApellido:       getEnv("ADMIN_SEED_APELLIDO", "Barber"),
		LoginRateLimitPerMinute: 10,
		TurnoRateLimitPerMinute: 30,
		PoolMaxConns:            10,
		ConnTimeout:             5 * time.Second,
		StatementTimeout:        10 * time.Second,
		MaxRequestBody:          1 << 20, // 1 MiB
	}

	var errs []string

	if cfg.DatabaseURL == "" {
		errs = append(errs, "DATABASE_URL es obligatoria (ej: postgres://user:pass@localhost:5432/barber?sslmode=disable)")
	}
	// Fail-closed: sin secreto firmemente establecido no se arranca.
	if len(cfg.JWTSecret) < 32 {
		errs = append(errs, "JWT_SECRET es obligatoria y debe tener al menos 32 bytes (genere con: openssl rand -base64 32)")
	}

	if v := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS")); v != "" {
		for _, o := range strings.Split(v, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				cfg.CORSAllowOrigins = append(cfg.CORSAllowOrigins, o)
			}
		}
	}
	if v := os.Getenv("AUTO_MIGRATE"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("AUTO_MIGRATE inválida: %q", v))
		}
		cfg.AutoMigrate = b
	}
	if v := os.Getenv("ADMIN_SEED"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("ADMIN_SEED inválida: %q", v))
		}
		cfg.AdminSeed = b
	} else {
		// Si hay password configurada, se habilita implícitamente (paridad con
		// la lógica condicional de AdminSeedRunner de Java).
		cfg.AdminSeed = cfg.AdminSeedPassword != ""
	}
	if cfg.AdminSeed && cfg.AdminSeedPassword == "" {
		errs = append(errs, "ADMIN_SEED_PASSWORD es obligatoria cuando ADMIN_SEED está activa")
	}
	if v := os.Getenv("LOGIN_RATE_LIMIT_PER_MINUTE"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			errs = append(errs, fmt.Sprintf("LOGIN_RATE_LIMIT_PER_MINUTE inválida: %q", v))
		} else {
			cfg.LoginRateLimitPerMinute = n
		}
	}
	if v := os.Getenv("TURNO_RATE_LIMIT_PER_MINUTE"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			errs = append(errs, fmt.Sprintf("TURNO_RATE_LIMIT_PER_MINUTE inválida: %q", v))
		} else {
			cfg.TurnoRateLimitPerMinute = n
		}
	}
	if v := os.Getenv("TRUST_PROXY_HEADERS"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("TRUST_PROXY_HEADERS inválida: %q", v))
		}
		cfg.TrustProxyHeaders = b
	}
	if v := os.Getenv("POOL_MAX_CONNS"); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil || n <= 0 {
			errs = append(errs, fmt.Sprintf("POOL_MAX_CONNS inválida: %q", v))
		} else {
			cfg.PoolMaxConns = int32(n)
		}
	}
	if v := os.Getenv("CONN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			errs = append(errs, fmt.Sprintf("CONN_TIMEOUT inválida: %q (use formato Go, ej 5s)", v))
		} else {
			cfg.ConnTimeout = d
		}
	}
	if v := os.Getenv("MAX_REQUEST_BODY"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n <= 0 {
			errs = append(errs, fmt.Sprintf("MAX_REQUEST_BODY inválida: %q", v))
		} else {
			cfg.MaxRequestBody = n
		}
	}
	if v := os.Getenv("STATEMENT_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d < 0 {
			errs = append(errs, fmt.Sprintf("STATEMENT_TIMEOUT inválida: %q (use formato Go, ej 10s)", v))
		} else {
			cfg.StatementTimeout = d
		}
	}

	// Fail-closed: sin APP_ENV=dev explícito se exige el chequeo de rol (RLS).
	cfg.AppEnv = getEnv("APP_ENV", "prod")

	if len(errs) > 0 {
		return nil, fmt.Errorf("configuración inválida: %s", strings.Join(errs, "; "))
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
