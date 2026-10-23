package config

import (
	"strings"
	"testing"
	"time"
)

// cleanEnv clears every variable Load reads so each test starts from a known
// state (t.Setenv rules out t.Parallel, which is fine here).
func cleanEnv(t *testing.T) {
	t.Helper()
	for _, v := range []string{
		"HTTP_ADDR", "DATABASE_URL", "JWT_SECRET", "DEFAULT_TENANT_SLUG",
		"CORS_ALLOW_ORIGINS", "AUTO_MIGRATE", "ADMIN_SEED", "ADMIN_SEED_PASSWORD",
		"ADMIN_SEED_EMAIL", "ADMIN_SEED_FIRST_NAME", "ADMIN_SEED_LAST_NAME",
		"LOGIN_RATE_LIMIT_PER_MINUTE", "POOL_MAX_CONNS", "CONN_TIMEOUT", "MAX_REQUEST_BODY",
		"BOOKING_RATE_LIMIT_PER_MINUTE", "TRUST_PROXY_HEADERS", "STATEMENT_TIMEOUT", "APP_ENV",
	} {
		t.Setenv(v, "")
	}
}

const validSecret = "a-secret-key-of-at-least-32-bytes!!"

func TestLoadDefaults(t *testing.T) {
	cleanEnv(t)
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/barber?sslmode=disable")
	t.Setenv("JWT_SECRET", validSecret)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTPAddr != ":8080" || cfg.DefaultTenantSlug != "default" {
		t.Errorf("wrong HTTP/tenant defaults: %+v", cfg)
	}
	if cfg.JWTExpiry != 24*time.Hour {
		t.Errorf("JWTExpiry = %v, want 24h", cfg.JWTExpiry)
	}
	if cfg.AppEnv != "prod" {
		t.Errorf("AppEnv = %q, want prod (fail-closed)", cfg.AppEnv)
	}
	if cfg.LoginRateLimitPerMinute != 10 || cfg.PoolMaxConns != 10 || cfg.ConnTimeout != 5*time.Second {
		t.Errorf("wrong pool/rate defaults: %+v", cfg)
	}
	if cfg.BookingRateLimitPerMinute != 30 || cfg.TrustProxyHeaders {
		t.Errorf("wrong booking/proxy defaults: booking=%d trustProxy=%v",
			cfg.BookingRateLimitPerMinute, cfg.TrustProxyHeaders)
	}
	if cfg.StatementTimeout != 10*time.Second {
		t.Errorf("StatementTimeout = %v, want 10s", cfg.StatementTimeout)
	}
	if cfg.MaxRequestBody != 1<<20 {
		t.Errorf("MaxRequestBody = %d, want 1MiB", cfg.MaxRequestBody)
	}
	if cfg.AdminSeed {
		t.Error("AdminSeed must not be enabled without a password")
	}
}

func TestLoadFailsClosedOnSecrets(t *testing.T) {
	cases := []struct {
		name, db, secret, wantErr string
	}{
		{"no DATABASE_URL", "", validSecret, "DATABASE_URL"},
		{"no JWT_SECRET", "postgres://x", "", "JWT_SECRET"},
		{"short JWT_SECRET", "postgres://x", "short", "JWT_SECRET"},
		{"JWT_SECRET with only spaces", "postgres://x", "                                ", "JWT_SECRET"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cleanEnv(t)
			t.Setenv("DATABASE_URL", c.db)
			t.Setenv("JWT_SECRET", c.secret)
			_, err := Load()
			if err == nil {
				t.Fatal("expected a fail-closed error")
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("error %q does not mention %q", err, c.wantErr)
			}
		})
	}
}

func TestLoadTrimsSecrets(t *testing.T) {
	cleanEnv(t)
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", "  "+validSecret+"  ")
	if _, err := Load(); err != nil {
		t.Fatalf("secrets with spaces should be trimmed: %v", err)
	}
}

func TestLoadCORSList(t *testing.T) {
	cleanEnv(t)
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", validSecret)
	t.Setenv("CORS_ALLOW_ORIGINS", "http://a.com , https://b.com,")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.CORSAllowOrigins) != 2 || cfg.CORSAllowOrigins[0] != "http://a.com" || cfg.CORSAllowOrigins[1] != "https://b.com" {
		t.Fatalf("CORSAllowOrigins = %v", cfg.CORSAllowOrigins)
	}
}

func TestLoadSeedImpliedByPassword(t *testing.T) {
	cleanEnv(t)
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", validSecret)
	t.Setenv("ADMIN_SEED_PASSWORD", "admin1234")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.AdminSeed {
		t.Fatal("ADMIN_SEED should be enabled implicitly by a password")
	}
}

func TestLoadSeedWithoutPasswordFails(t *testing.T) {
	cleanEnv(t)
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", validSecret)
	t.Setenv("ADMIN_SEED", "true")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "ADMIN_SEED_PASSWORD") {
		t.Fatalf("expected an error for ADMIN_SEED without password, got %v", err)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	cases := []struct{ name, key, value, wantErr string }{
		{"AUTO_MIGRATE not boolean", "AUTO_MIGRATE", "maybe", "AUTO_MIGRATE"},
		{"ADMIN_SEED not boolean", "ADMIN_SEED", "maybe", "ADMIN_SEED"},
		{"zero rate limit", "LOGIN_RATE_LIMIT_PER_MINUTE", "0", "LOGIN_RATE_LIMIT"},
		{"non-numeric rate limit", "LOGIN_RATE_LIMIT_PER_MINUTE", "x", "LOGIN_RATE_LIMIT"},
		{"zero pool", "POOL_MAX_CONNS", "0", "POOL_MAX_CONNS"},
		{"timeout not a duration", "CONN_TIMEOUT", "fast", "CONN_TIMEOUT"},
		{"negative timeout", "CONN_TIMEOUT", "-5s", "CONN_TIMEOUT"},
		{"zero max body", "MAX_REQUEST_BODY", "0", "MAX_REQUEST_BODY"},
		{"zero booking rate", "BOOKING_RATE_LIMIT_PER_MINUTE", "0", "BOOKING_RATE_LIMIT"},
		{"non-numeric booking rate", "BOOKING_RATE_LIMIT_PER_MINUTE", "many", "BOOKING_RATE_LIMIT"},
		{"trust proxy not boolean", "TRUST_PROXY_HEADERS", "yes", "TRUST_PROXY_HEADERS"},
		{"statement timeout not a duration", "STATEMENT_TIMEOUT", "fast", "STATEMENT_TIMEOUT"},
		{"negative statement timeout", "STATEMENT_TIMEOUT", "-1s", "STATEMENT_TIMEOUT"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cleanEnv(t)
			t.Setenv("DATABASE_URL", "postgres://x")
			t.Setenv("JWT_SECRET", validSecret)
			t.Setenv(c.key, c.value)
			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("expected an error for %s, got %v", c.key, err)
			}
		})
	}
}

func TestLoadAppliesValidValues(t *testing.T) {
	cleanEnv(t)
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", validSecret)
	t.Setenv("LOGIN_RATE_LIMIT_PER_MINUTE", "3")
	t.Setenv("POOL_MAX_CONNS", "25")
	t.Setenv("CONN_TIMEOUT", "30s")
	t.Setenv("MAX_REQUEST_BODY", "2048")
	t.Setenv("AUTO_MIGRATE", "true")
	t.Setenv("BOOKING_RATE_LIMIT_PER_MINUTE", "60")
	t.Setenv("TRUST_PROXY_HEADERS", "true")
	t.Setenv("STATEMENT_TIMEOUT", "3s")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LoginRateLimitPerMinute != 3 || cfg.PoolMaxConns != 25 ||
		cfg.ConnTimeout != 30*time.Second || cfg.MaxRequestBody != 2048 || !cfg.AutoMigrate {
		t.Fatalf("values not applied: %+v", cfg)
	}
	if cfg.BookingRateLimitPerMinute != 60 || !cfg.TrustProxyHeaders || cfg.StatementTimeout != 3*time.Second {
		t.Fatalf("newer values not applied: %+v", cfg)
	}
}
