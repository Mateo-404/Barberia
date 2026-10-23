//go:build integration

// Package integration runs end-to-end tests against a real PostgreSQL:
// migrations, the HTTP contract, concurrency and tenant isolation.
//
// It needs a reachable PostgreSQL. TEST_DATABASE_URL_BASE points at its
// "postgres" database (default below); every test creates and drops its own
// database. See backend-go/README.md.
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
	"github.com/Mateo-404/Barberia/backend-go/internal/config"
	"github.com/Mateo-404/Barberia/backend-go/internal/httpapi"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/admins"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/Mateo-404/Barberia/backend-go/internal/tenant"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultBaseURL = "postgres://postgres@127.0.0.1:55432/postgres?sslmode=disable"
	jwtSecret      = "test-secret-integration-at-least-32-bytes!"
	seedEmail      = "admin@admin.com"
	seedPassword   = "admin1234"
	dateLayout     = "2006-01-02T15:04:05"
)

func baseURL() string {
	if v := os.Getenv("TEST_DATABASE_URL_BASE"); v != "" {
		return v
	}
	return defaultBaseURL
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// emptyDatabase creates a throwaway database and returns its URL.
func emptyDatabase(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	name := fmt.Sprintf("barber_test_%d", time.Now().UnixNano())
	admin, err := pgxpool.New(ctx, baseURL())
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}
	defer admin.Close()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatalf("create database: %v", err)
	}
	t.Cleanup(func() {
		cleanup, err := pgxpool.New(context.Background(), baseURL())
		if err == nil {
			_, _ = cleanup.Exec(context.Background(), "DROP DATABASE "+name+" WITH (FORCE)")
			cleanup.Close()
		}
	})
	return strings.Replace(baseURL(), "/postgres?", "/"+name+"?", 1)
}

func openPool(t *testing.T, url string) *pgxpool.Pool {
	t.Helper()
	pool, err := database.NewPool(context.Background(), database.Options{ConnectionString: url, MaxConns: 5})
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// migratedDatabase creates a database and applies every migration.
func migratedDatabase(t *testing.T) (*pgxpool.Pool, string) {
	t.Helper()
	url := emptyDatabase(t)
	if err := database.RunMigrations(context.Background(), url, quietLogger()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return openPool(t, url), url
}

// env is the full HTTP server over a fresh database with the seed admin.
type env struct {
	handler  http.Handler
	pool     *pgxpool.Pool
	q        *sqlc.Queries
	tenantID int64
}

func newEnv(t *testing.T) *env {
	t.Helper()
	pool, _ := migratedDatabase(t)
	ctx := context.Background()
	cfg := &config.Config{
		CORSAllowOrigins:        []string{"http://localhost:5173"},
		DefaultTenantSlug:       "default",
		LoginRateLimitPerMinute: 50,
		// High on purpose: rate limiting has its own unit test.
		BookingRateLimitPerMinute: 1000,
		MaxRequestBody:            1 << 20,
	}
	jwtManager, err := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	q := sqlc.New(pool)
	resolver := tenant.NewResolver(q, "default")
	def, err := resolver.Resolve(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := admins.New(pool, q, jwtManager, quietLogger()).Seed(ctx, def.ID, seedEmail, seedPassword, "Admin", "Default"); err != nil {
		t.Fatal(err)
	}
	srv := httpapi.NewServer(httpapi.Deps{Cfg: cfg, Logger: quietLogger(), Pool: pool, JWT: jwtManager, Resolver: resolver})
	return &env{handler: srv.Handler(), pool: pool, q: q, tenantID: def.ID}
}

// response wraps a recorded response with JSON helpers.
type response struct {
	*httptest.ResponseRecorder
	t *testing.T
}

// json decodes the body into v.
func (r response) json(v any) {
	r.t.Helper()
	if err := json.Unmarshal(r.Body.Bytes(), v); err != nil {
		r.t.Fatalf("decode %s: %v", r.Body.String(), err)
	}
}

// problemCode returns the "code" of a problem detail body.
func (r response) problemCode() string {
	var p struct {
		Code string `json:"code"`
	}
	r.json(&p)
	return p.Code
}

// expect fails unless the status matches.
func (r response) expect(status int, what string) response {
	r.t.Helper()
	if r.Code != status {
		r.t.Fatalf("%s: status = %d, want %d, body = %s", what, r.Code, status, r.Body.String())
	}
	return r
}

// call sends a request; host and token are optional.
func (e *env) call(t *testing.T, method, path, body, token string, host ...string) response {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if len(host) > 0 {
		req.Host = host[0]
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.handler.ServeHTTP(rec, req)
	return response{rec, t}
}

func (e *env) login(t *testing.T, email, password string) string {
	t.Helper()
	var out struct {
		Token string `json:"token"`
	}
	e.call(t, http.MethodPost, "/api/auth/login", fmt.Sprintf(`{"email":%q,"password":%q}`, email, password), "").
		expect(http.StatusOK, "login").json(&out)
	return out.Token
}

func (e *env) ownerToken(t *testing.T) string { return e.login(t, seedEmail, seedPassword) }

// createService creates a service through the API and returns its id.
func (e *env) createService(t *testing.T, token, name string, price float64) int64 {
	t.Helper()
	var out struct {
		ID int64 `json:"id"`
	}
	e.call(t, http.MethodPost, "/api/services", fmt.Sprintf(`{"name":%q,"price":%v}`, name, price), token).
		expect(http.StatusCreated, "create service").json(&out)
	return out.ID
}

// insertCustomer adds a customer to a tenant directly (fixtures only).
func (e *env) insertCustomer(t *testing.T, tenantID int64, phone string) {
	t.Helper()
	if _, err := e.pool.Exec(context.Background(),
		`INSERT INTO customers (phone, first_name, last_name, tenant_id) VALUES ($1, 'Test', 'Fixture', $2)`, phone, tenantID); err != nil {
		t.Fatalf("insert customer: %v", err)
	}
}

// insertTenant adds a tenant reachable through host and returns its id.
func (e *env) insertTenant(t *testing.T, slug, host string) int64 {
	t.Helper()
	var id int64
	if err := e.pool.QueryRow(context.Background(),
		`INSERT INTO tenants (slug, name, host) VALUES ($1, $1, $2) RETURNING id`, slug, host).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

// futureAt returns days from now at hh:mm, within opening hours.
func futureAt(days, hour, minute int) time.Time {
	d := time.Now().AddDate(0, 0, days)
	return time.Date(d.Year(), d.Month(), d.Day(), hour, minute, 0, 0, time.Local)
}

func bookingBody(at time.Time, serviceID int64, phone, firstName, lastName string) string {
	return fmt.Sprintf(`{"startsAt":%q,"serviceId":%d,"customerPhone":%q,"customerFirstName":%q,"customerLastName":%q}`,
		at.Format(dateLayout), serviceID, phone, firstName, lastName)
}

// appointment mirrors the API's appointment shape.
type appointment struct {
	ID                int64   `json:"id"`
	StartsAt          string  `json:"startsAt"`
	CustomerPhone     *string `json:"customerPhone"`
	CustomerFirstName *string `json:"customerFirstName"`
	ServiceID         *int64  `json:"serviceId"`
	ServiceName       *string `json:"serviceName"`
	ServicePrice      float32 `json:"servicePrice"`
}
