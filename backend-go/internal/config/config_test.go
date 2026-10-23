package config

import (
	"strings"
	"testing"
	"time"
)

// entornoLimpio borra todas las variables que Load lee para que cada test
// parta de un estado conocido (t.Setenv impide t.Parallel, suficiente aquí).
func entornoLimpio(t *testing.T) {
	t.Helper()
	for _, v := range []string{
		"HTTP_ADDR", "DATABASE_URL", "JWT_SECRET", "DEFAULT_TENANT_SLUG",
		"CORS_ALLOW_ORIGINS", "AUTO_MIGRATE", "ADMIN_SEED", "ADMIN_SEED_PASSWORD",
		"ADMIN_SEED_EMAIL", "ADMIN_SEED_NOMBRE", "ADMIN_SEED_APELLIDO",
		"LOGIN_RATE_LIMIT_PER_MINUTE", "POOL_MAX_CONNS", "CONN_TIMEOUT", "MAX_REQUEST_BODY",
		"TURNO_RATE_LIMIT_PER_MINUTE", "TRUST_PROXY_HEADERS", "STATEMENT_TIMEOUT", "APP_ENV",
	} {
		t.Setenv(v, "")
	}
}

const secretoValido = "una-clave-secreta-de-al-menos-32-bytes!"

func TestLoadValoresPorDefecto(t *testing.T) {
	entornoLimpio(t)
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/barber?sslmode=disable")
	t.Setenv("JWT_SECRET", secretoValido)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTPAddr != ":8080" || cfg.DefaultTenantSlug != "default" {
		t.Errorf("defaults HTTP/tenant incorrectos: %+v", cfg)
	}
	if cfg.JWTExpiry != 24*time.Hour {
		t.Errorf("JWTExpiry = %v, want 24h", cfg.JWTExpiry)
	}
	if cfg.AppEnv != "prod" {
		t.Errorf("AppEnv = %q, want prod (fail-closed)", cfg.AppEnv)
	}
	if cfg.LoginRateLimitPerMinute != 10 || cfg.PoolMaxConns != 10 || cfg.ConnTimeout != 5*time.Second {
		t.Errorf("defaults pool/rate incorrectos: %+v", cfg)
	}
	if cfg.TurnoRateLimitPerMinute != 30 || cfg.TrustProxyHeaders {
		t.Errorf("defaults escritura/proxy incorrectos: turnos=%d trustProxy=%v",
			cfg.TurnoRateLimitPerMinute, cfg.TrustProxyHeaders)
	}
	if cfg.StatementTimeout != 10*time.Second {
		t.Errorf("StatementTimeout = %v, want 10s", cfg.StatementTimeout)
	}
	if cfg.MaxRequestBody != 1<<20 {
		t.Errorf("MaxRequestBody = %d, want 1MiB", cfg.MaxRequestBody)
	}
	if cfg.AdminSeed {
		t.Error("AdminSeed no debería habilitarse sin password")
	}
}

func TestLoadFailClosedSecretos(t *testing.T) {
	casos := []struct {
		nombre, db, secret, wantErr string
	}{
		{"sin DATABASE_URL", "", secretoValido, "DATABASE_URL"},
		{"sin JWT_SECRET", "postgres://x", "", "JWT_SECRET"},
		{"JWT_SECRET corta", "postgres://x", "corta", "JWT_SECRET"},
		{"JWT_SECRET con solo espacios", "postgres://x", "                                ", "JWT_SECRET"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			entornoLimpio(t)
			t.Setenv("DATABASE_URL", c.db)
			t.Setenv("JWT_SECRET", c.secret)
			_, err := Load()
			if err == nil {
				t.Fatal("esperaba error fail-closed")
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("error %q no menciona %q", err, c.wantErr)
			}
		})
	}
}

func TestLoadRecortaSecretos(t *testing.T) {
	entornoLimpio(t)
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", "  "+secretoValido+"  ")
	if _, err := Load(); err != nil {
		t.Fatalf("secretos con espacios deberían recortarse: %v", err)
	}
}

func TestLoadCORSLista(t *testing.T) {
	entornoLimpio(t)
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", secretoValido)
	t.Setenv("CORS_ALLOW_ORIGINS", "http://a.com , https://b.com,")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.CORSAllowOrigins) != 2 || cfg.CORSAllowOrigins[0] != "http://a.com" || cfg.CORSAllowOrigins[1] != "https://b.com" {
		t.Fatalf("CORSAllowOrigins = %v", cfg.CORSAllowOrigins)
	}
}

func TestLoadSeedImplicitaConPassword(t *testing.T) {
	entornoLimpio(t)
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", secretoValido)
	t.Setenv("ADMIN_SEED_PASSWORD", "admin1234")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	// Paridad con el AdminSeedRunner condicional de Java.
	if !cfg.AdminSeed {
		t.Fatal("ADMIN_SEED debería habilitarse implícitamente con password")
	}
}

func TestLoadSeedSinPasswordFallida(t *testing.T) {
	entornoLimpio(t)
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", secretoValido)
	t.Setenv("ADMIN_SEED", "true")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "ADMIN_SEED_PASSWORD") {
		t.Fatalf("esperaba error por ADMIN_SEED sin password, got %v", err)
	}
}

func TestLoadNumericasInvalidas(t *testing.T) {
	casos := []struct{ nombre, key, valor, wantErr string }{
		{"AUTO_MIGRATE no booleana", "AUTO_MIGRATE", "quizas", "AUTO_MIGRATE"},
		{"ADMIN_SEED no booleana", "ADMIN_SEED", "quizas", "ADMIN_SEED"},
		{"rate limit cero", "LOGIN_RATE_LIMIT_PER_MINUTE", "0", "LOGIN_RATE_LIMIT"},
		{"rate limit no numérica", "LOGIN_RATE_LIMIT_PER_MINUTE", "x", "LOGIN_RATE_LIMIT"},
		{"pool cero", "POOL_MAX_CONNS", "0", "POOL_MAX_CONNS"},
		{"timeout no duración", "CONN_TIMEOUT", "rápido", "CONN_TIMEOUT"},
		{"timeout negativo", "CONN_TIMEOUT", "-5s", "CONN_TIMEOUT"},
		{"max body cero", "MAX_REQUEST_BODY", "0", "MAX_REQUEST_BODY"},
		{"turno rate cero", "TURNO_RATE_LIMIT_PER_MINUTE", "0", "TURNO_RATE_LIMIT"},
		{"turno rate no numérica", "TURNO_RATE_LIMIT_PER_MINUTE", "muchas", "TURNO_RATE_LIMIT"},
		{"trust proxy no booleano", "TRUST_PROXY_HEADERS", "si", "TRUST_PROXY_HEADERS"},
		{"statement timeout no duración", "STATEMENT_TIMEOUT", "rápido", "STATEMENT_TIMEOUT"},
		{"statement timeout negativo", "STATEMENT_TIMEOUT", "-1s", "STATEMENT_TIMEOUT"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			entornoLimpio(t)
			t.Setenv("DATABASE_URL", "postgres://x")
			t.Setenv("JWT_SECRET", secretoValido)
			t.Setenv(c.key, c.valor)
			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("esperaba error por %s, got %v", c.key, err)
			}
		})
	}
}

func TestLoadNumericasValidas(t *testing.T) {
	entornoLimpio(t)
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", secretoValido)
	t.Setenv("LOGIN_RATE_LIMIT_PER_MINUTE", "3")
	t.Setenv("POOL_MAX_CONNS", "25")
	t.Setenv("CONN_TIMEOUT", "30s")
	t.Setenv("MAX_REQUEST_BODY", "2048")
	t.Setenv("AUTO_MIGRATE", "true")
	t.Setenv("TURNO_RATE_LIMIT_PER_MINUTE", "60")
	t.Setenv("TRUST_PROXY_HEADERS", "true")
	t.Setenv("STATEMENT_TIMEOUT", "3s")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LoginRateLimitPerMinute != 3 || cfg.PoolMaxConns != 25 ||
		cfg.ConnTimeout != 30*time.Second || cfg.MaxRequestBody != 2048 || !cfg.AutoMigrate {
		t.Fatalf("valores no aplicados: %+v", cfg)
	}
	if cfg.TurnoRateLimitPerMinute != 60 || !cfg.TrustProxyHeaders || cfg.StatementTimeout != 3*time.Second {
		t.Fatalf("valores nuevos no aplicados: %+v", cfg)
	}
}
