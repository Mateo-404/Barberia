//go:build integration

// Tests de integración complementarios: CRUD end-to-end de cada recurso,
// reglas de negocio de turnos, panel con datos reales, tokens legados,
// seed idempotente y aislamiento de findDateTimes.
package integration

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/administradores"
)

// fechaFutura devuelve una fecha futura a las 14:00 (horario laboral).
func fechaFutura(dias int) time.Time {
	f := time.Now().AddDate(0, 0, dias)
	return time.Date(f.Year(), f.Month(), f.Day(), 14, 0, 0, 0, time.Local)
}

func crearServicio(t *testing.T, e *entornoHTTP, token, tipo string, precio float64) {
	t.Helper()
	rec := e.req(t, http.MethodPost, "/servicios",
		fmt.Sprintf(`{"tipo":%q,"precio":%v}`, tipo, precio), token)
	if rec.Code != 201 {
		t.Fatalf("crear servicio %q: %d %s", tipo, rec.Code, rec.Body.String())
	}
}

// crearClienteDirecto inserta un cliente en el tenant por defecto (para tests
// que solo necesitan que la FK exista).
func crearClienteDirecto(t *testing.T, e *entornoHTTP, telefono string) {
	t.Helper()
	_, err := e.pool.Exec(context.Background(),
		`INSERT INTO clientes (telefono, nombre, apellido, tenant_id)
		 VALUES ($1, 'Test', 'Fixtures', (SELECT id FROM tenants WHERE slug='default'))`, telefono)
	if err != nil {
		t.Fatalf("crear cliente directo: %v", err)
	}
}

// jwtDeTest construye un manager con el mismo secreto que el entorno.
func jwtDeTest(t *testing.T) *auth.JWTManager {
	t.Helper()
	m, err := auth.NewJWTManager(secretDefecto(), 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// ---- CRUD de clientes end-to-end ----

func TestCRUDClientesEndToEnd(t *testing.T) {
	e, _ := nuevoEntorno(t)
	token := e.login(t)

	// Crear.
	rec := e.req(t, http.MethodPost, "/clientes",
		`{"telefono":"1151111111","nombre":"Ana","apellido":"Gomez","email":"ana@email.com"}`, token)
	if rec.Code != 201 || !strings.Contains(rec.Body.String(), `"telefono":"1151111111"`) {
		t.Fatalf("crear: %d %s", rec.Code, rec.Body.String())
	}

	// Duplicado → 409 con la forma exacta.
	rec = e.req(t, http.MethodPost, "/clientes",
		`{"telefono":"1151111111","nombre":"Otro","apellido":"Nombre"}`, token)
	if rec.Code != 409 {
		t.Fatalf("duplicado: %d %s", rec.Code, rec.Body.String())
	}
	want := `{"type":"/errors/resource-already-exists","title":"El recurso ya existe","status":409,"detail":"Ya existe un cliente con el teléfono 1151111111","instance":"/clientes"}`
	if strings.TrimSpace(rec.Body.String()) != want {
		t.Fatalf("409 body = %s", rec.Body.String())
	}

	// GET.
	rec = e.req(t, http.MethodGet, "/clientes/1151111111", "", token)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"nombre":"Ana"`) {
		t.Fatalf("get: %d %s", rec.Code, rec.Body.String())
	}

	// PATCH parcial: solo nombre; el resto se conserva (copyProperties).
	rec = e.req(t, http.MethodPatch, "/clientes/1151111111", `{"nombre":"Ana Maria"}`, token)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"nombre":"Ana Maria"`) ||
		!strings.Contains(rec.Body.String(), `"apellido":"Gomez"`) ||
		!strings.Contains(rec.Body.String(), `"email":"ana@email.com"`) {
		t.Fatalf("patch parcial: %d %s", rec.Code, rec.Body.String())
	}

	// PATCH con telefono en el body: la PK no cambia (decisión documentada).
	rec = e.req(t, http.MethodPatch, "/clientes/1151111111", `{"telefono":"999999999","nombre":"Ana M."}`, token)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"telefono":"1151111111"`) {
		t.Fatalf("patch con telefono no debería cambiar la PK: %d %s", rec.Code, rec.Body.String())
	}
	if rec = e.req(t, http.MethodGet, "/clientes/999999999", "", token); rec.Code != 404 {
		t.Fatalf("la ruta vieja no debería existir: %d", rec.Code)
	}

	// exists.
	rec = e.req(t, http.MethodGet, "/clientes/exists/1151111111", "", token)
	if strings.TrimSpace(rec.Body.String()) != "true" {
		t.Fatalf("exists: %s", rec.Body.String())
	}

	// Paginación: forma Spring Page.
	rec = e.req(t, http.MethodGet, "/clientes/0/10", "", token)
	for _, campo := range []string{`"content"`, `"pageable"`, `"totalElements"`, `"totalPages"`,
		`"last"`, `"first"`, `"size"`, `"number"`, `"sort"`, `"numberOfElements"`, `"empty"`} {
		if !strings.Contains(rec.Body.String(), campo) {
			t.Fatalf("page sin %s: %s", campo, rec.Body.String())
		}
	}

	// DELETE → 204; repetir → 404 vacío.
	if rec = e.req(t, http.MethodDelete, "/clientes/1151111111", "", token); rec.Code != 204 {
		t.Fatalf("delete: %d", rec.Code)
	}
	rec = e.req(t, http.MethodDelete, "/clientes/1151111111", "", token)
	if rec.Code != 404 || rec.Body.Len() != 0 {
		t.Fatalf("delete repetido: %d %q", rec.Code, rec.Body.String())
	}
}

// TestClienteConTurnosNoSeBorra: FK restrict → 500 (como el Java).
func TestClienteConTurnosNoSeBorra(t *testing.T) {
	e, _ := nuevoEntorno(t)
	token := e.login(t)
	crearServicio(t, e, token, "Corte", 5000)
	crearClienteDirecto(t, e, "1152222222")
	futura := fechaFutura(3)
	body := fmt.Sprintf(`{"fechaHora":"%s","idServicio":1,"telefonoCliente":"1152222222","nombreCliente":"Luis","apellidoCliente":"Sosa"}`,
		futura.Format("2006-01-02T15:04:05"))
	if rec := e.req(t, http.MethodPost, "/turnos", body, ""); rec.Code != 201 {
		t.Fatalf("crear turno: %d %s", rec.Code, rec.Body.String())
	}
	rec := e.req(t, http.MethodDelete, "/clientes/1152222222", "", token)
	if rec.Code != 500 {
		t.Fatalf("delete con FK restrict debería dar 500, got %d %s", rec.Code, rec.Body.String())
	}
}

// ---- CRUD de servicios end-to-end ----

func TestCRUDServiciosEndToEnd(t *testing.T) {
	e, _ := nuevoEntorno(t)
	token := e.login(t)

	rec := e.req(t, http.MethodPost, "/servicios", `{"tipo":"Barba","precio":2500}`, token)
	if rec.Code != 201 || !strings.Contains(rec.Body.String(), `"tipo":"Barba"`) {
		t.Fatalf("crear: %d %s", rec.Code, rec.Body.String())
	}
	// GET público sin token.
	rec = e.req(t, http.MethodGet, "/servicios/1", "", "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"precio":2500`) {
		t.Fatalf("get publico: %d %s", rec.Code, rec.Body.String())
	}
	// PATCH.
	rec = e.req(t, http.MethodPatch, "/servicios/1", `{"precio":3000}`, token)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"precio":3000`) ||
		!strings.Contains(rec.Body.String(), `"tipo":"Barba"`) {
		t.Fatalf("patch: %d %s", rec.Code, rec.Body.String())
	}
	// Validaciones.
	if rec = e.req(t, http.MethodPost, "/servicios", `{"tipo":"","precio":1000}`, token); rec.Code != 400 {
		t.Fatalf("tipo en blanco debería dar 400: %d", rec.Code)
	}
	if rec = e.req(t, http.MethodPost, "/servicios", `{"tipo":"X","precio":-1}`, token); rec.Code != 400 {
		t.Fatalf("precio negativo debería dar 400: %d", rec.Code)
	}
	// DELETE → 204; existe → false.
	if rec = e.req(t, http.MethodDelete, "/servicios/1", "", token); rec.Code != 204 {
		t.Fatalf("delete: %d", rec.Code)
	}
	if rec = e.req(t, http.MethodGet, "/servicios/exists/1", "", ""); strings.TrimSpace(rec.Body.String()) != "false" {
		t.Fatalf("exists tras delete: %s", rec.Body.String())
	}
}

// ---- CRUD de administradores + login con hash (ADR-004) ----

func TestCRUDAdministradoresEndToEnd(t *testing.T) {
	e, _ := nuevoEntorno(t)
	token := e.login(t)

	// Crear admin (nunca expone hash).
	rec := e.req(t, http.MethodPost, "/administradores",
		`{"nombre":"Carla","apellido":"Ruiz","email":"carla@barber.com","contrasenia":"clave12345"}`, token)
	if rec.Code != 201 || strings.Contains(rec.Body.String(), "argon2") || strings.Contains(rec.Body.String(), "contrasenia") {
		t.Fatalf("crear admin: %d %s", rec.Code, rec.Body.String())
	}

	// Email duplicado → 409.
	rec = e.req(t, http.MethodPost, "/administradores",
		`{"nombre":"Otra","apellido":"Persona","email":"carla@barber.com","contrasenia":"clave12345"}`, token)
	if rec.Code != 409 || !strings.Contains(rec.Body.String(), "carla@barber.com") {
		t.Fatalf("duplicado: %d %s", rec.Code, rec.Body.String())
	}

	// Login con el admin nuevo.
	rec = e.req(t, http.MethodPost, "/administradores/login",
		`{"email":"carla@barber.com","contrasenia":"clave12345"}`, "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"email":"carla@barber.com"`) ||
		!strings.Contains(rec.Body.String(), `"token"`) {
		t.Fatalf("login admin nuevo: %d %s", rec.Code, rec.Body.String())
	}

	// PATCH de contraseña: se hashea (ADR-004); la vieja deja de servir.
	rec = e.req(t, http.MethodGet, "/administradores", "", token)
	if !strings.Contains(rec.Body.String(), "carla@barber.com") {
		t.Fatalf("el admin no aparece en la lista: %s", rec.Body.String())
	}
	idCarla := extraerIDPorEmail(t, rec.Body.String(), "carla@barber.com")
	rec = e.req(t, http.MethodPatch, "/administradores/"+idCarla, `{"contrasenia":"nuevaClave123"}`, token)
	if rec.Code != 200 || strings.Contains(rec.Body.String(), "nuevaClave123") {
		t.Fatalf("patch contraseña: %d %s", rec.Code, rec.Body.String())
	}
	if rec = e.req(t, http.MethodPost, "/administradores/login",
		`{"email":"carla@barber.com","contrasenia":"clave12345"}`, ""); rec.Code != 401 {
		t.Fatalf("la contraseña vieja debería fallar: %d %s", rec.Code, rec.Body.String())
	}
	if rec = e.req(t, http.MethodPost, "/administradores/login",
		`{"email":"carla@barber.com","contrasenia":"nuevaClave123"}`, ""); rec.Code != 200 {
		t.Fatalf("la contraseña nueva debería funcionar: %d %s", rec.Code, rec.Body.String())
	}

	// Login con contraseña corta → 400 (validación).
	if rec = e.req(t, http.MethodPost, "/administradores/login",
		`{"email":"carla@barber.com","contrasenia":"corta"}`, ""); rec.Code != 400 {
		t.Fatalf("contraseña corta: %d", rec.Code)
	}

	// DELETE.
	if rec = e.req(t, http.MethodDelete, "/administradores/"+idCarla, "", token); rec.Code != 204 {
		t.Fatalf("delete: %d", rec.Code)
	}
}

// extraerIDPorEmail extrae el id numérico del objeto admin que contiene el
// email dado dentro de un array JSON.
func extraerIDPorEmail(t *testing.T, cuerpo, email string) string {
	t.Helper()
	i := strings.Index(cuerpo, `"email":"`+email+`"`)
	if i < 0 {
		t.Fatalf("email %q no encontrado en %s", email, cuerpo)
	}
	j := strings.LastIndex(cuerpo[:i], `"id":`)
	if j < 0 {
		t.Fatalf("id no encontrado antes de %s", email)
	}
	resto := cuerpo[j+len(`"id":`):]
	k := strings.Index(resto, ",")
	if k < 0 {
		k = len(resto)
	}
	return strings.TrimSpace(resto[:k])
}

// ---- Reglas de negocio de turnos ----

func TestReglasDeTurnoEndToEnd(t *testing.T) {
	e, _ := nuevoEntorno(t)
	token := e.login(t)
	crearServicio(t, e, token, "Corte", 5000)

	futura := fechaFutura(5)

	casos := []struct {
		nombre     string
		hora       string
		wantStatus int
		wantDetail string
	}{
		{"apertura 13:00 válida", "13:00", 201, ""},
		{"13:30 válida", "13:30", 201, ""},
		{"cierre 20:00 válido", "20:00", 201, ""},
		{"12:30 fuera de rango", "12:30", 400, "La hora debe estar entre 13:00 y 20:00."},
		{"20:30 fuera de rango", "20:30", 400, "La hora debe estar entre 13:00 y 20:00."},
		{"14:15 no es intervalo", "14:15", 400, "La hora debe ser en intervalos de 30 minutos."},
	}
	for i, c := range casos {
		h, m := parseHora(c.hora)
		f := time.Date(futura.Year(), futura.Month(), futura.Day(), h, m, 0, 0, time.Local)
		body := fmt.Sprintf(`{"fechaHora":"%s","idServicio":1,"telefonoCliente":"115%d%d","nombreCliente":"N%d","apellidoCliente":"A%d"}`,
			f.Format("2006-01-02T15:04:05"), 30000000, i, i, i)
		rec := e.req(t, http.MethodPost, "/turnos", body, "")
		if rec.Code != c.wantStatus {
			t.Fatalf("%s: status = %d, want %d (%s)", c.nombre, rec.Code, c.wantStatus, rec.Body.String())
		}
		if c.wantDetail != "" && !strings.Contains(rec.Body.String(), c.wantDetail) {
			t.Fatalf("%s: detail sin %q: %s", c.nombre, c.wantDetail, rec.Body.String())
		}
	}

	// Fecha pasada → 400 (validación @Future, no regla de negocio).
	pasada := time.Now().AddDate(0, 0, -1)
	body := fmt.Sprintf(`{"fechaHora":"%s","idServicio":1,"telefonoCliente":"1150000000","nombreCliente":"V","apellidoCliente":"P"}`,
		pasada.Format("2006-01-02T15:04:05"))
	if rec := e.req(t, http.MethodPost, "/turnos", body, ""); rec.Code != 400 {
		t.Fatalf("fecha pasada: %d", rec.Code)
	}

	// Servicio inexistente → 404 con cuerpo.
	f2 := fechaFutura(6)
	body = fmt.Sprintf(`{"fechaHora":"%s","idServicio":9999,"telefonoCliente":"1150000000","nombreCliente":"V","apellidoCliente":"P"}`,
		f2.Format("2006-01-02T15:04:05"))
	rec := e.req(t, http.MethodPost, "/turnos", body, "")
	if rec.Code != 404 || !strings.Contains(rec.Body.String(), "No existe un servicio con el id 9999") {
		t.Fatalf("servicio inexistente: %d %s", rec.Code, rec.Body.String())
	}
}

func parseHora(s string) (int, int) {
	var h, m int
	_, _ = fmt.Sscanf(s, "%d:%d", &h, &m)
	return h, m
}

// TestFindOrCreateClienteReutiliza: como Java, un turno con un teléfono
// existente NO sobrescribe nombre/apellido/email del cliente.
func TestFindOrCreateClienteReutiliza(t *testing.T) {
	e, _ := nuevoEntorno(t)
	token := e.login(t)
	crearServicio(t, e, token, "Corte", 5000)

	f1 := fechaFutura(7)
	body := fmt.Sprintf(`{"fechaHora":"%s","idServicio":1,"telefonoCliente":"1154444444","nombreCliente":"Original","apellidoCliente":"Apellido","emailCliente":"original@email.com"}`,
		f1.Format("2006-01-02T15:04:05"))
	if rec := e.req(t, http.MethodPost, "/turnos", body, ""); rec.Code != 201 {
		t.Fatalf("turno 1: %d %s", rec.Code, rec.Body.String())
	}

	// Mismo teléfono, otro nombre: el cliente existente manda.
	f2 := fechaFutura(8)
	body = fmt.Sprintf(`{"fechaHora":"%s","idServicio":1,"telefonoCliente":"1154444444","nombreCliente":"Distinto","apellidoCliente":"Ignorado"}`,
		f2.Format("2006-01-02T15:04:05"))
	if rec := e.req(t, http.MethodPost, "/turnos", body, ""); rec.Code != 201 {
		t.Fatalf("turno 2: %d %s", rec.Code, rec.Body.String())
	}

	rec := e.req(t, http.MethodGet, "/clientes/1154444444", "", token)
	if !strings.Contains(rec.Body.String(), `"nombre":"Original"`) ||
		!strings.Contains(rec.Body.String(), `"email":"original@email.com"`) {
		t.Fatalf("find-or-create sobrescribió el cliente: %s", rec.Body.String())
	}
}

// TestTurnosAllTransaccional: ADR-005 — crea con cliente nuevo (el Java daba
// 500) y revierte todo si un elemento es inválido.
func TestTurnosAllTransaccional(t *testing.T) {
	e, _ := nuevoEntorno(t)
	token := e.login(t)
	crearServicio(t, e, token, "Corte", 5000)

	// Caso feliz con cliente nuevo (bug 500 del Java corregido).
	f1 := fechaFutura(9)
	f2 := fechaFutura(10)
	body := fmt.Sprintf(`[{"fechaHora":"%s","idServicio":1,"telefonoCliente":"1155555551","nombreCliente":"Nuevo","apellidoCliente":"Cliente"},{"fechaHora":"%s","idServicio":1,"telefonoCliente":"1155555552","nombreCliente":"Otro","apellidoCliente":"Nuevo"}]`,
		f1.Format("2006-01-02T15:04:05"), f2.Format("2006-01-02T15:04:05"))
	rec := e.req(t, http.MethodPost, "/turnos/all", body, token)
	if rec.Code != 201 || !strings.Contains(rec.Body.String(), `"id":1`) || !strings.Contains(rec.Body.String(), `"id":2`) {
		t.Fatalf("all feliz: %d %s", rec.Code, rec.Body.String())
	}

	// Con duplicado interno: todo se revierte (transaccional).
	f3 := fechaFutura(11)
	f4 := fechaFutura(11) // misma fecha que f3 → duplicado
	body = fmt.Sprintf(`[{"fechaHora":"%s","idServicio":1,"telefonoCliente":"1156666666","nombreCliente":"A","apellidoCliente":"A"},{"fechaHora":"%s","idServicio":1,"telefonoCliente":"1157777777","nombreCliente":"B","apellidoCliente":"B"}]`,
		f3.Format("2006-01-02T15:04:05"), f4.Format("2006-01-02T15:04:05"))
	rec = e.req(t, http.MethodPost, "/turnos/all", body, token)
	if rec.Code != 409 || !strings.Contains(rec.Body.String(), "slot_taken") {
		t.Fatalf("all con duplicado: %d %s", rec.Code, rec.Body.String())
	}
	// Ninguno de los dos se persistió.
	if rec = e.req(t, http.MethodGet, "/turnos", "", token); strings.Contains(rec.Body.String(), "1156666666") || strings.Contains(rec.Body.String(), "1157777777") {
		t.Fatalf("all debería revertir todo: %s", rec.Body.String())
	}
}

// TestDuplicadoPasadoRechazadoPorIndice: desviación documentada (ADR-006) —
// el índice único rechaza también duplicados en el pasado.
func TestDuplicadoPasadoRechazadoPorIndice(t *testing.T) {
	e, ctx := nuevoEntorno(t)
	token := e.login(t)
	crearServicio(t, e, token, "Corte", 5000)
	crearClienteDirecto(t, e, "1150000001")

	pasada := time.Now().AddDate(0, 0, -7)
	pasada = time.Date(pasada.Year(), pasada.Month(), pasada.Day(), 15, 0, 0, 0, time.Local)

	if _, err := e.pool.Exec(ctx,
		`INSERT INTO turnos (fecha_hora, cliente_id, servicio_id, tenant_id)
		 VALUES ($1, '1150000001', 1, (SELECT id FROM tenants WHERE slug='default'))`, pasada); err != nil {
		t.Fatal(err)
	}
	// Mover otro turno a la misma fecha pasada viola el índice único → 400.
	otra := pasada.AddDate(0, 0, 1)
	if _, err := e.pool.Exec(ctx,
		`INSERT INTO turnos (fecha_hora, cliente_id, servicio_id, tenant_id)
		 VALUES ($1, '1150000001', 1, (SELECT id FROM tenants WHERE slug='default'))`, otra); err != nil {
		t.Fatal(err)
	}
	var idOtra int64
	if err := e.pool.QueryRow(ctx, `SELECT id FROM turnos WHERE fecha_hora = $1`, otra).Scan(&idOtra); err != nil {
		t.Fatal(err)
	}
	rec := e.req(t, http.MethodPatch, "/turnos/"+fmt.Sprint(idOtra),
		fmt.Sprintf(`{"fechaHora":"%s"}`, pasada.Format("2006-01-02T15:04:05")), token)
	if rec.Code != 409 || !strings.Contains(rec.Body.String(), "slot_taken") {
		t.Fatalf("duplicado pasado vía PATCH: %d %s", rec.Code, rec.Body.String())
	}
}

// ---- Panel con datos reales ----

func TestPanelConDatos(t *testing.T) {
	e, ctx := nuevoEntorno(t)
	token := e.login(t)
	crearServicio(t, e, token, "Corte", 5000)
	crearServicio(t, e, token, "Barba", 2500)
	crearClienteDirecto(t, e, "1150000001")

	var tid int64
	if err := e.pool.QueryRow(ctx, `SELECT id FROM tenants WHERE slug='default'`).Scan(&tid); err != nil {
		t.Fatal(err)
	}
	hoy := time.Now()
	hoy14 := time.Date(hoy.Year(), hoy.Month(), hoy.Day(), 14, 0, 0, 0, time.Local)
	ayer15 := hoy14.AddDate(0, 0, -1).Add(time.Hour)
	for _, v := range []struct {
		fecha time.Time
		serv  int64
	}{{hoy14, 1}, {hoy14.Add(2 * time.Hour), 1}, {ayer15, 2}} {
		if _, err := e.pool.Exec(ctx,
			`INSERT INTO turnos (fecha_hora, cliente_id, servicio_id, tenant_id) VALUES ($1, '1150000001', $2, $3)`,
			v.fecha, v.serv, tid); err != nil {
			t.Fatal(err)
		}
	}

	// Los ingresos del mes dependen de si ayer cae en el mismo mes (día 1).
	mismoMes := hoy.Day() != 1
	ingresosMes, ingresosAnterior := 10000.0, 0.0 // 2 × 5000 hoy
	if mismoMes {
		ingresosMes += 2500 // ayer (Barba) suma al mismo mes
	} else {
		ingresosAnterior = 2500
	}

	rec := e.req(t, http.MethodGet, "/estadisticas/panel", "", token)
	if rec.Code != 200 {
		t.Fatalf("panel: %d %s", rec.Code, rec.Body.String())
	}
	cuerpo := rec.Body.String()
	esperados := []string{
		`"turnosHoy":2`,
		`"turnosAyer":1`,
		fmt.Sprintf(`"ingresosMes":%v`, ingresosMes),
		fmt.Sprintf(`"ingresosMesAnterior":%v`, ingresosAnterior),
		`"cantClientes":1`,
		`"ingresosDiarios"`,
		`"ingresoTotal"`,
		`"servicios"`,
		`"cantidadRealizado":2`, // Corte: 2 turnos en el último mes
		`"horarios"`,
		`"hora":14`,
	}
	for _, campo := range esperados {
		if !strings.Contains(cuerpo, campo) {
			t.Fatalf("panel sin %s: %s", campo, cuerpo)
		}
	}
	// El bucket diario más reciente es el de hoy (formato Jackson).
	if i := strings.Index(cuerpo, `"ingresosDiarios":[{"fecha":"`); i < 0 {
		t.Fatalf("ingresosDiarios vacío: %s", cuerpo)
	} else if !strings.Contains(cuerpo[i:i+80], hoy14.Format("2006-01-02")) {
		t.Fatalf("el primer ingreso diario no es hoy: %s", cuerpo[i:i+80])
	}
}

// ---- Token legado del Java (sin claim tenant) ----

func TestTokenLegadoSinTenantRechazado(t *testing.T) {
	e, _ := nuevoEntorno(t)
	// Los tokens sin claim de tenant se rechazan con 401 (re-login); no hay
	// fallback por sub (decisión ADR-002).
	legacy, err := jwtDeTest(t).Issue(1, "admin@admin.com", 0, "owner")
	if err != nil {
		t.Fatal(err)
	}
	rec := e.req(t, http.MethodGet, "/administradores", "", legacy)
	if rec.Code != 401 {
		t.Fatalf("token sin claim tenant debe ser 401: %d %s", rec.Code, rec.Body.String())
	}
}

// ---- Seed idempotente ----

func TestSeedIdempotente(t *testing.T) {
	e, ctx := nuevoEntorno(t)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	seed := administradores.New(e.pool, e.q, jwtDeTest(t), logger)

	// El entorno ya sembró admin@admin.com; un segundo seed no debe duplicar.
	if err := seed.Seed(ctx, 1, "admin@admin.com", "admin1234", "Admin", "Por Defecto"); err != nil {
		t.Fatal(err)
	}
	var total int
	if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM administradores`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("seed duplicó: total = %d", total)
	}
	// Password corta → error (como el Java).
	if err := seed.Seed(ctx, 999, "x@y.com", "corta", "X", "Y"); err == nil || !strings.Contains(err.Error(), "8 caracteres") {
		t.Fatalf("seed con password corta: %v", err)
	}
}

// ---- Aislamiento de findDateTimes ----

func TestFindDateTimesAislado(t *testing.T) {
	e, ctx := nuevoEntorno(t)
	token := e.login(t)
	crearServicio(t, e, token, "Corte", 5000)
	crearClienteDirecto(t, e, "1150000001")
	crearClienteDirecto(t, e, "1150000002")

	// Segundo tenant con su servicio y su turno.
	var tid2 int64
	if err := e.pool.QueryRow(ctx,
		`INSERT INTO tenants (slug, name, host) VALUES ('salon2', 'Segundo', 'salon2.example.com') RETURNING id`).Scan(&tid2); err != nil {
		t.Fatal(err)
	}
	if _, err := e.pool.Exec(ctx,
		`INSERT INTO servicios (tipo, precio, tenant_id) VALUES ('Corte2', 9000, $1)`, tid2); err != nil {
		t.Fatal(err)
	}
	// El cliente debe existir en el MISMO tenant (FK compuesta tenant_id+telefono).
	if _, err := e.pool.Exec(ctx,
		`INSERT INTO clientes (telefono, nombre, apellido, tenant_id) VALUES ('1150000001', 'Test', 'Fixtures', $1)`, tid2); err != nil {
		t.Fatal(err)
	}

	// Turnos a la misma hora en cada tenant (el aislamiento lo permite).
	futura := fechaFutura(12)
	if _, err := e.pool.Exec(ctx,
		`INSERT INTO turnos (fecha_hora, cliente_id, servicio_id, tenant_id) VALUES ($1, '1150000001', (SELECT id FROM servicios WHERE tenant_id=$2), $2)`,
		futura, tid2); err != nil {
		t.Fatal(err)
	}
	if _, err := e.pool.Exec(ctx,
		`INSERT INTO turnos (fecha_hora, cliente_id, servicio_id, tenant_id)
		 VALUES ($1, '1150000002', 1, (SELECT id FROM tenants WHERE slug='default'))`, futura); err != nil {
		t.Fatal(err)
	}

	// El default ve solo el suyo (misma hora, otro tenant → no filtrable por string).
	rec := e.req(t, http.MethodGet, "/turnos/findDateTimes", "", "")
	if got := strings.Count(rec.Body.String(), futura.Format("2006-01-02T15:04")); got != 1 {
		t.Fatalf("findDateTimes default debería tener 1 entrada, got %d: %s", got, rec.Body.String())
	}
	// El host del segundo tenant ve exactamente el suyo (también 1, pero de su tenant).
	r := httptest.NewRequest(http.MethodGet, "/turnos/findDateTimes", nil)
	r.Host = "salon2.example.com"
	rec2 := httptest.NewRecorder()
	e.handler.ServeHTTP(rec2, r)
	if got := strings.Count(rec2.Body.String(), futura.Format("2006-01-02T15:04")); got != 1 {
		t.Fatalf("findDateTimes salon2 debería tener 1 entrada, got %d: %s", got, rec2.Body.String())
	}
}
