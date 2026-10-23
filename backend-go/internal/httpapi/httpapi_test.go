package httpapi

import (
	"context"
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

type resolverFalso struct{ t tenant.Tenant }

func (f resolverFalso) Resolve(_ context.Context, _ string) (tenant.Tenant, error) { return f.t, nil }

func nuevoServerDeTest(t *testing.T) http.Handler {
	t.Helper()
	jwtManager, err := auth.NewJWTManager("test-secret-de-al-menos-32-bytes!!", 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		CORSAllowOrigins:        []string{"http://localhost:5173"},
		LoginRateLimitPerMinute: 5,
		TurnoRateLimitPerMinute: 5,
		MaxRequestBody:          1 << 20,
	}
	s := &Server{
		cfg:              cfg,
		logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
		jwt:              jwtManager,
		limiter:          auth.NewLoginLimiter(5),
		escrituraLimiter: auth.NewLoginLimiter(5),
		resolver:         resolverFalso{t: tenant.Tenant{ID: 1, Slug: "default"}},
		now:              time.Now,
	}
	return s.Handler()
}

func tokenValido(t *testing.T) string {
	t.Helper()
	jwtManager, _ := auth.NewJWTManager("test-secret-de-al-menos-32-bytes!!", 24*time.Hour)
	token, err := jwtManager.Issue(1, "admin@admin.com", 1, "owner")
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestPing(t *testing.T) {
	h := nuevoServerDeTest(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Body.String() != "Backend funcionando!" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestRutaProtegidaSinTokenDa401Exacto(t *testing.T) {
	h := nuevoServerDeTest(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/turnos", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
	want := `{"type":"/errors/unauthorized","title":"No autorizado","status":401,"detail":"Se requiere autenticación para acceder a este recurso"}`
	if strings.TrimSpace(rec.Body.String()) != want {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestTokenInvalidoRechazaInclusoEnRutasPublicas(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodGet, "/servicios", nil)
	req.Header.Set("Authorization", "Bearer token-basura")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Token JWT inválido o expirado") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestJSONIlegibleDa400Validacion(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodPost, "/turnos", strings.NewReader("{no-json"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	want := `{"type":"about:blank","title":"Bad Request","status":400,"detail":"Invalid request content.","instance":"/turnos"}`
	if strings.TrimSpace(rec.Body.String()) != want {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestValidacionTurnoFaltanteDa400(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodPost, "/turnos", strings.NewReader(`{"fechaHora":"2030-01-01T14:00:00","idServicio":1}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Invalid request content.") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestValidacionTurnoFechaPasadaDa400(t *testing.T) {
	h := nuevoServerDeTest(t)
	body := `{"fechaHora":"2020-01-01T14:00:00","idServicio":1,"telefonoCliente":"115","nombreCliente":"Juan","apellidoCliente":"Perez"}`
	req := httptest.NewRequest(http.MethodPost, "/turnos", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestValidacionServicioPrecioNoPositivo(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodPost, "/servicios",
		strings.NewReader(`{"tipo":"Corte","precio":0}`))
	req.Header.Set("Authorization", "Bearer "+tokenValido(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestRutaDesconocidaDa404SpringStyle(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodGet, "/no-existe", nil)
	req.Header.Set("Authorization", "Bearer "+tokenValido(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "No static resource /no-existe.") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestCORSPreflight(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodOptions, "/turnos", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("allow-origin = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("allow-credentials = %q", got)
	}
}

func TestCORSOrigenNoPermitidoNoRefleja(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "https://ejemplo-malo.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("no debería reflejar el origen, got %q", got)
	}
}
