package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
	"github.com/Mateo-404/Barberia/backend-go/internal/config"
	"github.com/Mateo-404/Barberia/backend-go/internal/tenant"
)

const testSecret = "test-secret-at-least-32-bytes-long!"

type fakeResolver struct{ t tenant.Tenant }

func (f fakeResolver) Resolve(_ context.Context, _ string) (tenant.Tenant, error) { return f.t, nil }

// newTestServer builds a server without a database: handlers that reach a
// service panic and are turned into 500s by middlewareRecover, which is
// enough to test routing, auth, validation and rate limits.
func newTestServer(t *testing.T, limit int) http.Handler {
	t.Helper()
	jwtManager, err := auth.NewJWTManager(testSecret, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{
		cfg: &config.Config{
			CORSAllowOrigins:          []string{"http://localhost:5173"},
			LoginRateLimitPerMinute:   limit,
			BookingRateLimitPerMinute: limit,
			MaxRequestBody:            1 << 20,
		},
		logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		jwt:            jwtManager,
		loginLimiter:   auth.NewLoginLimiter(limit),
		bookingLimiter: auth.NewLoginLimiter(limit),
		resolver:       fakeResolver{t: tenant.Tenant{ID: 1, Slug: "default"}},
		now:            time.Now,
	}
	return s.Handler()
}

func tokenFor(t *testing.T, tenantID int64, role string) string {
	t.Helper()
	jwtManager, _ := auth.NewJWTManager(testSecret, 24*time.Hour)
	token, err := jwtManager.Issue(1, "admin@admin.com", tenantID, role)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

// do sends a request; token may be empty.
func do(h http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeProblem(t *testing.T, rec *httptest.ResponseRecorder) problem {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("content-type = %q, body = %s", ct, rec.Body.String())
	}
	var p problem
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("decode problem: %v (%s)", err, rec.Body.String())
	}
	return p
}

func TestHealthz(t *testing.T) {
	rec := do(newTestServer(t, 5), http.MethodGet, "/healthz", "", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"ok"`) {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
}

func TestProtectedRouteWithoutTokenIs401(t *testing.T) {
	rec := do(newTestServer(t, 5), http.MethodGet, "/api/appointments", "", "")
	p := decodeProblem(t, rec)
	if rec.Code != http.StatusUnauthorized || p.Code != "auth.unauthorized" || p.Instance != "/api/appointments" {
		t.Fatalf("status = %d problem = %+v", rec.Code, p)
	}
}

func TestInvalidTokenIsRejectedEvenOnPublicRoutes(t *testing.T) {
	rec := do(newTestServer(t, 5), http.MethodGet, "/api/services", "", "garbage-token")
	if p := decodeProblem(t, rec); rec.Code != http.StatusUnauthorized || p.Code != "auth.invalid_token" {
		t.Fatalf("status = %d problem = %+v", rec.Code, p)
	}
}

func TestTokenWithoutTenantIsRejected(t *testing.T) {
	rec := do(newTestServer(t, 5), http.MethodGet, "/api/appointments", "", tokenFor(t, 0, auth.RoleOwner))
	if p := decodeProblem(t, rec); rec.Code != http.StatusUnauthorized || p.Code != "auth.invalid_token" {
		t.Fatalf("status = %d problem = %+v", rec.Code, p)
	}
}

func TestTokenOfAnotherTenantIsForbidden(t *testing.T) {
	rec := do(newTestServer(t, 5), http.MethodGet, "/api/appointments", "", tokenFor(t, 2, auth.RoleOwner))
	if p := decodeProblem(t, rec); rec.Code != http.StatusForbidden || p.Code != "auth.forbidden" {
		t.Fatalf("status = %d problem = %+v", rec.Code, p)
	}
}

func TestAdminsRequireOwner(t *testing.T) {
	h := newTestServer(t, 5)
	for _, role := range []string{auth.RoleStaff, auth.RoleAdmin} {
		rec := do(h, http.MethodPost, "/api/admins", `{}`, tokenFor(t, 1, role))
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s creating an admin: status = %d, want 403", role, rec.Code)
		}
	}
}

func TestSettingsRequireOwnerOrAdmin(t *testing.T) {
	rec := do(newTestServer(t, 5), http.MethodGet, "/api/admin/settings", "", tokenFor(t, 1, auth.RoleStaff))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("staff reading settings: status = %d, want 403", rec.Code)
	}
}

func TestMalformedJSONIs400(t *testing.T) {
	rec := do(newTestServer(t, 5), http.MethodPost, "/api/appointments", "{not-json", "")
	if p := decodeProblem(t, rec); rec.Code != http.StatusBadRequest || p.Code != "request.malformed" {
		t.Fatalf("status = %d problem = %+v", rec.Code, p)
	}
}

func TestAppointmentValidationListsEveryField(t *testing.T) {
	rec := do(newTestServer(t, 5), http.MethodPost, "/api/appointments",
		`{"startsAt":"not-a-date","customerEmail":"not-an-email"}`, "")
	p := decodeProblem(t, rec)
	if rec.Code != http.StatusBadRequest || p.Code != "request.invalid_fields" {
		t.Fatalf("status = %d problem = %+v", rec.Code, p)
	}
	got := map[string]string{}
	for _, e := range p.Errors {
		got[e.Field] = e.Code
	}
	want := map[string]string{
		"startsAt":          codeInvalidDate,
		"serviceId":         codeRequired,
		"customerPhone":     codeRequired,
		"customerFirstName": codeRequired,
		"customerLastName":  codeRequired,
		"customerEmail":     codeInvalidEmail,
	}
	for field, code := range want {
		if got[field] != code {
			t.Errorf("%s: code = %q, want %q (all: %v)", field, got[field], code, got)
		}
	}
}

func TestServicePriceMustBePositive(t *testing.T) {
	rec := do(newTestServer(t, 5), http.MethodPost, "/api/services", `{"name":"Cut","price":0}`, tokenFor(t, 1, auth.RoleOwner))
	p := decodeProblem(t, rec)
	if rec.Code != http.StatusBadRequest || len(p.Errors) != 1 || p.Errors[0] != (fieldProblem{"price", codeMustBePositive}) {
		t.Fatalf("status = %d problem = %+v", rec.Code, p)
	}
}

func TestPatchValidation(t *testing.T) {
	h := newTestServer(t, 5)
	token := tokenFor(t, 1, auth.RoleOwner)
	cases := []struct{ path, body, field, code string }{
		{"/api/admins/1", `{"password":"short"}`, "password", codeTooShort},
		{"/api/services/1", `{"price":-5}`, "price", codeMustBePositive},
		{"/api/customers/1151111111", `{"email":"not-an-email"}`, "email", codeInvalidEmail},
		{"/api/customers/1151111111", `{"firstName":"` + strings.Repeat("a", maxName+1) + `"}`, "firstName", codeTooLong},
	}
	for _, c := range cases {
		rec := do(h, http.MethodPatch, c.path, c.body, token)
		p := decodeProblem(t, rec)
		if rec.Code != http.StatusBadRequest || len(p.Errors) == 0 || p.Errors[0].Field != c.field || p.Errors[0].Code != c.code {
			t.Errorf("PATCH %s %s: status = %d problem = %+v", c.path, c.body, rec.Code, p)
		}
	}
}

func TestNonNumericIDIs404(t *testing.T) {
	rec := do(newTestServer(t, 5), http.MethodGet, "/api/appointments/abc", "", tokenFor(t, 1, auth.RoleOwner))
	if p := decodeProblem(t, rec); rec.Code != http.StatusNotFound || p.Code != "resource.not_found" {
		t.Fatalf("status = %d problem = %+v", rec.Code, p)
	}
}

func TestPageParamsAreBounded(t *testing.T) {
	h := newTestServer(t, 5)
	token := tokenFor(t, 1, auth.RoleOwner)
	for _, q := range []string{"page=-1", "size=0", "size=101", "page=abc", "size=99999999999999999999"} {
		rec := do(h, http.MethodGet, "/api/appointments/paged?"+q, "", token)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", q, rec.Code)
		}
	}
}

func TestUnknownRouteAndMethod(t *testing.T) {
	h := newTestServer(t, 5)
	token := tokenFor(t, 1, auth.RoleOwner)
	if rec := do(h, http.MethodGet, "/api/nope", "", token); decodeProblem(t, rec).Code != "request.not_found" {
		t.Errorf("unknown route: %s", rec.Body.String())
	}
	if rec := do(h, http.MethodDelete, "/healthz", "", ""); rec.Code != http.StatusMethodNotAllowed ||
		decodeProblem(t, rec).Code != "request.method_not_allowed" {
		t.Errorf("method not allowed: %d %s", rec.Code, rec.Body.String())
	}
}

func TestPanicIsRecoveredAs500(t *testing.T) {
	// Services are nil in the test server: the handler panics.
	rec := do(newTestServer(t, 5), http.MethodGet, "/api/appointments", "", tokenFor(t, 1, auth.RoleOwner))
	if p := decodeProblem(t, rec); rec.Code != http.StatusInternalServerError || p.Code != "server.internal" {
		t.Fatalf("status = %d problem = %+v", rec.Code, p)
	}
}

func TestCORS(t *testing.T) {
	h := newTestServer(t, 5)
	req := httptest.NewRequest(http.MethodOptions, "/api/appointments", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" ||
		rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("preflight: %d %v", rec.Code, rec.Header())
	}

	req = httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("a foreign origin must not be reflected, got %q", got)
	}
}

func TestSecurityHeaders(t *testing.T) {
	rec := do(newTestServer(t, 5), http.MethodGet, "/healthz", "", "")
	for k, v := range map[string]string{
		"X-Content-Type-Options":       "nosniff",
		"X-Frame-Options":              "DENY",
		"Referrer-Policy":              "no-referrer",
		"Cross-Origin-Resource-Policy": "same-site",
	} {
		if got := rec.Header().Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestNoStoreOnLoginAndAuthenticatedResponses(t *testing.T) {
	h := newTestServer(t, 5)
	if got := do(h, http.MethodPost, loginPath, `{"email":"a@b.com","password":"12345678"}`, "").Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("login Cache-Control = %q", got)
	}
	if got := do(h, http.MethodGet, "/api/appointments", "", tokenFor(t, 1, auth.RoleOwner)).Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("authenticated Cache-Control = %q", got)
	}
	if got := do(h, http.MethodGet, "/healthz", "", "").Header().Get("Cache-Control"); got != "" {
		t.Errorf("public Cache-Control = %q", got)
	}
}

func TestBookingRateLimit(t *testing.T) {
	h := newTestServer(t, 2)
	body := `{"startsAt":"2030-01-01T14:00:00","serviceId":1,"customerPhone":"115","customerFirstName":"Juan","customerLastName":"Perez"}`
	codes := make([]int, 3)
	for i := range codes {
		codes[i] = do(h, http.MethodPost, "/api/appointments", body, "").Code
	}
	// The first two pass the limiter (and fail later without a database).
	if codes[0] == http.StatusTooManyRequests || codes[1] == http.StatusTooManyRequests || codes[2] != http.StatusTooManyRequests {
		t.Fatalf("codes = %v, want only the third to be 429", codes)
	}
}

func TestLoginRateLimit(t *testing.T) {
	h := newTestServer(t, 2)
	body := `{"email":"a@b.com","password":"12345678"}`
	for i := 0; i < 2; i++ {
		do(h, http.MethodPost, loginPath, body, "")
	}
	rec := do(h, http.MethodPost, loginPath, body, "")
	if p := decodeProblem(t, rec); rec.Code != http.StatusTooManyRequests || p.Code != "request.rate_limited" || p.Instance != loginPath {
		t.Fatalf("status = %d problem = %+v", rec.Code, p)
	}
}
