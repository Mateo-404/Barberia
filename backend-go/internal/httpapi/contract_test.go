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
	"github.com/go-chi/chi/v5"
)

// ---- construirPagina: réplica de Page<T> de Spring Data ----

func TestConstruirPaginaSpring(t *testing.T) {
	content := []string{"a", "b", "c"}
	p := construirPagina(content, 25, 0, 10)
	if p.TotalElements != 25 || p.TotalPages != 3 {
		t.Errorf("totales: %+v", p)
	}
	if !p.First || p.Last {
		t.Errorf("first/last: first=%v last=%v", p.First, p.Last)
	}
	if p.Pageable.PageNumber != 0 || p.Pageable.PageSize != 10 || p.Pageable.Offset != 0 {
		t.Errorf("pageable: %+v", p.Pageable)
	}
	if !p.Pageable.Paged || p.Pageable.Unpaged {
		t.Errorf("paged: %+v", p.Pageable)
	}
	if p.Sort.Unsorted != true || p.Sort.Sorted != false || p.Sort.Empty != true {
		t.Errorf("sort default de Spring: %+v", p.Sort)
	}
	if p.NumberOfElements != 3 || p.Empty {
		t.Errorf("numberOfElements/empty: %+v", p)
	}

	// Última página (25 elementos, size 10 → página 2 tiene 5).
	p2 := construirPagina(content, 25, 2, 10)
	if !p2.Last || p2.First {
		t.Errorf("página final: first=%v last=%v", p2.First, p2.Last)
	}
	if p2.Pageable.Offset != 20 {
		t.Errorf("offset = %d, want 20", p2.Pageable.Offset)
	}
}

func TestConstruirPaginaVacia(t *testing.T) {
	p := construirPagina([]string{}, 0, 0, 10)
	if p.TotalElements != 0 || p.TotalPages != 0 || !p.Empty || p.NumberOfElements != 0 {
		t.Errorf("página vacía: %+v", p)
	}
	// Spring: con total=0 la página 0 es también la última.
	if !p.First || !p.Last {
		t.Errorf("vacía debería ser first y last: first=%v last=%v", p.First, p.Last)
	}
	// content null no: JSON con array vacío.
	b, _ := json.Marshal(p)
	if !strings.Contains(string(b), `"content":[]`) {
		t.Errorf("content debería serializarse como [] en página vacía: %s", b)
	}
}

// ---- decodificarJSONObligatorio: semántica Jackson ----

func TestDecodificarJSONIgnoraDesconocidos(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/servicios", strings.NewReader(`{"tipo":"Corte","precio":5000,"campoRaro":true}`))
	var dst struct {
		Tipo   string   `json:"tipo"`
		Precio *float32 `json:"precio"`
	}
	if !decodificarJSONObligatorio(rec, req, &dst) {
		t.Fatalf("campos desconocidos deberían ignorarse (FAIL_ON_UNKNOWN_PROPERTIES=false): %s", rec.Body.String())
	}
	if dst.Tipo != "Corte" || dst.Precio == nil || *dst.Precio != 5000 {
		t.Errorf("dst = %+v", dst)
	}
}

func TestDecodificarJSONRechazaBodyVacio(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/servicios", strings.NewReader(""))
	var dst struct{}
	if decodificarJSONObligatorio(rec, req, &dst) {
		t.Fatal("body vacío debería dar 400")
	}
	if !strings.Contains(rec.Body.String(), "Invalid request content.") {
		t.Errorf("body = %s", rec.Body.String())
	}
}

func TestDecodificarJSONRechazaContenidoExtra(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/servicios", strings.NewReader(`{"tipo":"A"}{"tipo":"B"}`))
	var dst struct{}
	if decodificarJSONObligatorio(rec, req, &dst) {
		t.Fatal("dos valores JSON deberían dar 400 (Jackson no acepta trailing)")
	}
}

func TestDecodificarJSONRechazaCuerpoGigante(t *testing.T) {
	gigante := `{"tipo":"` + strings.Repeat("a", 2<<20) + `"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/servicios", strings.NewReader(gigante))
	var dst struct{}
	if decodificarJSONObligatorio(rec, req, &dst) {
		t.Fatal("cuerpo > 1MiB debería dar 400")
	}
}

// ---- validarEmail: réplica de @Email de Hibernate Validator ----

func TestValidarEmail(t *testing.T) {
	validos := []string{"a@b.com", "user.name+tag@ejemplo.com.ar", "x@y"}
	invalidos := []string{"plain", "a@", "@b.com", "a@b.c.", "a b@c.com", "a@@b.com"}
	for _, e := range validos {
		v := e
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		if !validarEmail(rec, req, &v) {
			t.Errorf("%q debería ser válida", e)
		}
	}
	for _, e := range invalidos {
		v := e
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		if validarEmail(rec, req, &v) {
			t.Errorf("%q debería ser inválida", e)
		}
	}
	// null es válido (@Email sin mandatory).
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	if !validarEmail(rec, req, nil) {
		t.Error("nil debería ser válido")
	}
}

// ---- clienteIP: rate limit por el peer, sin X-Forwarded-For por defecto ----

func TestClienteIPPordefectoIgnoraForwarded(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "192.168.1.5:4242"
	if got := s.clienteIP(req); got != "192.168.1.5" {
		t.Errorf("clienteIP = %q", got)
	}
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	if got := s.clienteIP(req); got != "192.168.1.5" {
		t.Errorf("no debería confiar en X-Forwarded-For, got %q", got)
	}
}

func TestClienteIPConProxyConfiable(t *testing.T) {
	s := &Server{trustProxyHeaders: true}
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "10.0.0.9:4242"
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 10.0.0.1")
	if got := s.clienteIP(req); got != "1.2.3.4" {
		t.Errorf("con TRUST_PROXY_HEADERS debería usar la primera IP: %q", got)
	}
}

// ---- errores de formas HTTP ----

func TestMethodNotAllowedShape(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodDelete, "/ping", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d", rec.Code)
	}
	want := `{"type":"about:blank","title":"Method Not Allowed","status":405,"detail":"Method 'DELETE' is not supported.","instance":"/ping"}`
	if strings.TrimSpace(rec.Body.String()) != want {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestContentTypeErrorEsProblemJSON(t *testing.T) {
	h := nuevoServerDeTest(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/turnos", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("content-type = %q", ct)
	}
}

func TestParamIDNoNumericoDa400(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodGet, "/turnos/abc", nil)
	req.Header.Set("Authorization", "Bearer "+tokenValido(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Invalid request content.") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

// Spring PageRequest.of(-1, 10) lanza IllegalArgumentException sin handler → 500.
func TestParamPaginaInvalidoDa500(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodGet, "/turnos/-1/10", nil)
	req.Header.Set("Authorization", "Bearer "+tokenValido(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

// Un size desmedido se acota antes de tocar la BD (regresión de seguridad:
// evitaba amplificación de trabajo por petición).
func TestParamPaginaSizeDesmedidoDa500(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodGet, "/turnos/0/100000", nil)
	req.Header.Set("Authorization", "Bearer "+tokenValido(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("size desmedido debería cortarse: status = %d", rec.Code)
	}
}

// Un page fuera del rango int32 replica el MethodArgumentTypeMismatch → 400.
func TestParamPaginaFueraDeInt32Da400(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodGet, "/turnos/99999999999999999999/10", nil)
	req.Header.Set("Authorization", "Bearer "+tokenValido(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("page fuera de int32: status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

// ---- cabeceras de seguridad (regresión frente a la auditoría) ----

func TestHeadersDeSeguridad(t *testing.T) {
	h := nuevoServerDeTest(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	quiero := map[string]string{
		"X-Content-Type-Options":       "nosniff",
		"X-Frame-Options":              "DENY",
		"Referrer-Policy":              "no-referrer",
		"Cross-Origin-Resource-Policy": "same-site",
	}
	for k, v := range quiero {
		if got := rec.Header().Get(k); got != v {
			t.Errorf("%s = %q, quiero %q", k, got, v)
		}
	}
}

func TestCacheControlNoStoreEnLoginYAutenticadas(t *testing.T) {
	h := nuevoServerDeTest(t)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/administradores/login",
		strings.NewReader(`{"email":"a@b.com","contrasenia":"12345678"}`)))
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("login Cache-Control = %q", got)
	}

	req := httptest.NewRequest(http.MethodGet, "/turnos", nil)
	req.Header.Set("Authorization", "Bearer "+tokenValido(t))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("autenticada Cache-Control = %q", got)
	}

	// Pública y sin token: no se fuerza no-store.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Errorf("pública Cache-Control = %q", got)
	}
}

// ---- rate limit de escrituras públicas ----

func TestRateLimitDeTurnosPublicos(t *testing.T) {
	h := serverConLimiteTurnos(t, 2)
	body := `{"fechaHora":"2030-01-01T14:00:00","idServicio":1,"telefonoCliente":"115","nombreCliente":"Juan","apellidoCliente":"Perez"}`
	codigos := make([]int, 3)
	for i := range codigos {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/turnos", strings.NewReader(body)))
		codigos[i] = rec.Code
	}
	// Las dos primeras pasan la capa de rate limit (500 por falta de BD en el
	// test) y la tercera se corta con 429.
	if codigos[0] == http.StatusTooManyRequests || codigos[1] == http.StatusTooManyRequests {
		t.Fatalf("las dos primeras no deberían limitarse: %v", codigos)
	}
	if codigos[2] != http.StatusTooManyRequests {
		t.Fatalf("la tercera debería dar 429: %v", codigos)
	}
}

func serverConLimiteTurnos(t *testing.T, limite int) http.Handler {
	t.Helper()
	jwtManager, err := auth.NewJWTManager("test-secret-de-al-menos-32-bytes!!", 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{
		cfg: &config.Config{LoginRateLimitPerMinute: limite, TurnoRateLimitPerMinute: limite,
			MaxRequestBody: 1 << 20},
		logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
		jwt:              jwtManager,
		limiter:          auth.NewLoginLimiter(limite),
		escrituraLimiter: auth.NewLoginLimiter(limite),
		resolver:         resolverFalso{t: tenant.Tenant{ID: 1, Slug: "default"}},
		now:              time.Now,
	}
	return s.Handler()
}

// ---- validación de PATCH ----

func TestPatchAdminContraseniaCortaDa400(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodPatch, "/administradores/1", strings.NewReader(`{"contrasenia":"corta"}`))
	req.Header.Set("Authorization", "Bearer "+tokenValido(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("contraseña <8 en PATCH: status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPatchServicioPrecioNoPositivoDa400(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodPatch, "/servicios/1", strings.NewReader(`{"precio":-5}`))
	req.Header.Set("Authorization", "Bearer "+tokenValido(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("precio ≤0 en PATCH: status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPatchClienteEmailInvalidoDa400(t *testing.T) {
	h := nuevoServerDeTest(t)
	req := httptest.NewRequest(http.MethodPatch, "/clientes/1151111111", strings.NewReader(`{"email":"no-es-email"}`))
	req.Header.Set("Authorization", "Bearer "+tokenValido(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("email inválido en PATCH: status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestCreateMaxLengthDa400(t *testing.T) {
	h := nuevoServerDeTest(t)
	largo := strings.Repeat("a", maxNombre+1)
	body := `{"telefono":"1151111111","nombre":"` + largo + `","apellido":"Perez"}`
	req := httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tokenValido(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("nombre sobre maxNombre: status = %d", rec.Code)
	}
}

// ---- rate limit de login: 429 réplica ----

func serverConLimite(t *testing.T, limite int) http.Handler {
	t.Helper()
	jwtManager, err := auth.NewJWTManager("test-secret-de-al-menos-32-bytes!!", 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{
		cfg:      &config.Config{LoginRateLimitPerMinute: limite, MaxRequestBody: 1 << 20},
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		jwt:      jwtManager,
		limiter:  auth.NewLoginLimiter(limite),
		resolver: resolverFalso{t: tenant.Tenant{ID: 1, Slug: "default"}},
		now:      time.Now,
	}
	return s.Handler()
}

func TestLoginRateLimit429(t *testing.T) {
	h := serverConLimite(t, 2)
	body := `{"email":"a@b.com","contrasenia":"12345678"}`
	// Agota la cuota (los intentos llegan al servicio; en este server de
	// prueba sin DB terminan en 500, pero consumen el cupo del limiter).
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/administradores/login", strings.NewReader(body))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}
	// El siguiente intento debe devolver la forma exacta del 429.
	req := httptest.NewRequest(http.MethodPost, "/administradores/login", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, esperaba 429 tras agotar el límite", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"type":"/errors/rate-limited"`) ||
		!strings.Contains(rec.Body.String(), `"status":429`) ||
		!strings.Contains(rec.Body.String(), `"instance":"/administradores/login"`) {
		t.Fatalf("body 429 = %s", rec.Body.String())
	}
}

// ---- recover middleware: un panic interno no tumba el proceso ----

func TestPanicRecuperadoDa500(t *testing.T) {
	h := nuevoServerDeTest(t)
	// Los servicios son nil en este server de prueba: el handler panickea al
	// tocar el pool y el middlewareRecover debe convertirlo en 500 réplica.
	req := httptest.NewRequest(http.MethodGet, "/turnos", nil)
	req.Header.Set("Authorization", "Bearer "+tokenValido(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `/errors/internal-server-error`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

// ---- esPublica: matriz del SecurityConfig Java ----

func TestEsPublica(t *testing.T) {
	casos := []struct {
		metodo, path string
		publica      bool
	}{
		{http.MethodGet, "/ping", true},
		{http.MethodGet, "/healthz", true},
		{http.MethodGet, "/readyz", true},
		{http.MethodPost, "/turnos", true},
		{http.MethodGet, "/turnos/findDateTimes", true},
		{http.MethodGet, "/servicios", true},
		{http.MethodGet, "/servicios/1", true},
		{http.MethodGet, "/servicios/0/10", true},
		{http.MethodPost, "/administradores/login", true},
		{http.MethodOptions, "/turnos", true},
		{http.MethodGet, "/turnos", false},
		{http.MethodPost, "/servicios", false},
		{http.MethodPatch, "/servicios/1", false},
		{http.MethodGet, "/administradores", false},
		{http.MethodGet, "/estadisticas/panel", false},
		{http.MethodDelete, "/turnos/1", false},
	}
	for _, c := range casos {
		req := httptest.NewRequest(c.metodo, c.path, nil)
		if got := esPublica(req); got != c.publica {
			t.Errorf("esPublica(%s %s) = %v, want %v", c.metodo, c.path, got, c.publica)
		}
	}
}

// ---- paramPagina: límites y overflow (unidad) ----

// conURLParams inyecta los parámetros de ruta de chi sin pasar por el router.
func conURLParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestParamPaginaLimites(t *testing.T) {
	s := &Server{now: time.Now}
	casos := []struct {
		page, size string
		wantStatus int
		wantOK     bool
	}{
		{"0", "10", 0, true},
		{"0", "100", 0, true}, // tope exacto
		{"0", "101", http.StatusInternalServerError, false},
		{"0", "100000", http.StatusInternalServerError, false},
		{"-1", "10", http.StatusInternalServerError, false},
		{"0", "0", http.StatusInternalServerError, false},
		{"2147483647", "100", 0, true},                      // int32 máximo, sin overflow (offset en int64)
		{"2147483648", "100", http.StatusBadRequest, false}, // fuera de int32
		{"99999999999999999999", "10", http.StatusBadRequest, false},
		{"abc", "10", http.StatusBadRequest, false},
	}
	for _, c := range casos {
		req := httptest.NewRequest(http.MethodGet, "/turnos/x/y", nil)
		req = conURLParams(req, map[string]string{"page": c.page, "size": c.size})
		rec := httptest.NewRecorder()
		page, size, ok := s.paramPagina(rec, req)
		if ok != c.wantOK {
			t.Errorf("page=%q size=%q: ok = %v, want %v", c.page, c.size, ok, c.wantOK)
			continue
		}
		if c.wantStatus != 0 && rec.Code != c.wantStatus {
			t.Errorf("page=%q size=%q: status = %d, want %d", c.page, c.size, rec.Code, c.wantStatus)
		}
		if ok && (page < 0 || size < 1 || size > maxPaginaSize) {
			t.Errorf("page=%q size=%q: valores fuera de rango devueltos (%d,%d)", c.page, c.size, page, size)
		}
	}
}

// ---- validación opcional (PATCH): solo rechaza lo presente ----

func TestValidacionesOpcionales(t *testing.T) {
	vacio := ""
	corta := "corta"
	invalido := "no-es-email"
	negativo := float32(-1)

	casos := []struct {
		nombre string
		fn     func(w http.ResponseWriter, r *http.Request) bool
		want   bool
	}{
		{"nil email válido", func(w http.ResponseWriter, r *http.Request) bool {
			return validarOpcionalEmail(w, r, nil)
		}, true},
		{"email inválido", func(w http.ResponseWriter, r *http.Request) bool {
			return validarOpcionalEmail(w, r, &invalido)
		}, false},
		{"contrasenia corta", func(w http.ResponseWriter, r *http.Request) bool {
			return validarOpcionalMinLength(w, r, &corta, 8)
		}, false},
		{"contrasenia nil ok", func(w http.ResponseWriter, r *http.Request) bool {
			return validarOpcionalMinLength(w, r, nil, 8)
		}, true},
		{"precio negativo", func(w http.ResponseWriter, r *http.Request) bool {
			return validarOpcionalPositivo(w, r, &negativo)
		}, false},
		{"precio nil ok", func(w http.ResponseWriter, r *http.Request) bool {
			return validarOpcionalPositivo(w, r, nil)
		}, true},
		{"nombre vacío presente", func(w http.ResponseWriter, r *http.Request) bool {
			return validarNotBlank(w, r, &vacio)
		}, false},
	}
	for _, c := range casos {
		req := httptest.NewRequest(http.MethodPatch, "/x", nil)
		rec := httptest.NewRecorder()
		if got := c.fn(rec, req); got != c.want {
			t.Errorf("%s: = %v, want %v", c.nombre, got, c.want)
		}
	}
}
