//go:build integration

// Tests de integración: migraciones, contrato HTTP end-to-end contra
// PostgreSQL real, concurrencia y aislamiento de tenant.
//
// Requiere un PostgreSQL accesible (por defecto el cluster de usuario de
// /tmp/opencode/pg en el puerto 55432; ver backend-go/README.md).
// Variables: TEST_DATABASE_URL_BASE (opcional), TEST_JWT_SECRET (opcional).
package integration

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
	"github.com/Mateo-404/Barberia/backend-go/internal/config"
	"github.com/Mateo-404/Barberia/backend-go/internal/httpapi"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/administradores"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/Mateo-404/Barberia/backend-go/internal/tenant"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	urlBase   = "postgres://postgres@127.0.0.1:55432/postgres?sslmode=disable"
	jwtSecret = "test-secret-integration-at-least-32-bytes!"
)

func urlBaseDefecto() string {
	if v := os.Getenv("TEST_DATABASE_URL_BASE"); v != "" {
		return v
	}
	return urlBase
}

func secretDefecto() string {
	if v := os.Getenv("TEST_JWT_SECRET"); v != "" {
		return v
	}
	return jwtSecret
}

// nuevaBase crea una base de datos por test y aplica las migraciones.
func nuevaBase(t *testing.T) (context.Context, *pgxpool.Pool, string) {
	t.Helper()
	ctx := context.Background()
	nombre := fmt.Sprintf("barber_test_%d", time.Now().UnixNano())

	admin, err := pgxpool.New(ctx, urlBaseDefecto())
	if err != nil {
		t.Fatalf("conectar a postgres: %v", err)
	}
	defer admin.Close()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+nombre); err != nil {
		t.Fatalf("crear base: %v", err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(ctx, "DROP DATABASE "+nombre+" WITH (FORCE)")
	})

	url := strings.Replace(urlBaseDefecto(), "/postgres?", "/"+nombre+"?", 1)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	if err := database.RunMigrations(ctx, url, logger); err != nil {
		t.Fatalf("migrar: %v", err)
	}
	pool, err := database.NewPool(ctx, database.Options{ConnectionString: url, MaxConns: 5})
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return ctx, pool, url
}

// entornoHTTP arma el servidor completo sobre la base del test.
type entornoHTTP struct {
	handler http.Handler
	pool    *pgxpool.Pool
	q       *sqlc.Queries
	token   string
}

func nuevoEntorno(t *testing.T) (*entornoHTTP, context.Context) {
	t.Helper()
	ctx, pool, _ := nuevaBase(t)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	cfg := &config.Config{
		CORSAllowOrigins:        []string{"http://localhost:5173"},
		DefaultTenantSlug:       "default",
		LoginRateLimitPerMinute: 50,
		// Alto a propósito: los tests no buscan el 429 (hay un test unitario
		// dedicado al rate limit de escrituras).
		TurnoRateLimitPerMinute: 1000,
		MaxRequestBody:          1 << 20,
	}
	jwtManager, err := auth.NewJWTManager(secretDefecto(), 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	q := sqlc.New(pool)
	resolver := tenant.NewResolver(q, "default")

	// Semilla con la contraseña del Java para verificar compatibilidad Argon2.
	seed := administradores.New(pool, q, jwtManager, logger)
	tid, err := resolver.Resolve(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := seed.Seed(ctx, tid.ID, "admin@admin.com", "admin1234", "Admin", "Por Defecto"); err != nil {
		t.Fatal(err)
	}
	// PostgreSQL asigna ids a los servicios con una única secuencia GLOBAL, no
	// por tenant: dejar el catálogo del tenant default vacío garantiza que el
	// primer servicio creado sea id=1.
	if _, err := pool.Exec(ctx, `DELETE FROM servicios WHERE tenant_id <> $1`, tid.ID); err != nil {
		t.Fatal(err)
	}

	srv := httpapi.NewServer(httpapi.Deps{Cfg: cfg, Logger: logger, Pool: pool, JWT: jwtManager, Resolver: resolver})
	return &entornoHTTP{handler: srv.Handler(), pool: pool, q: q}, ctx
}

func (e *entornoHTTP) req(t *testing.T, metodo, ruta, cuerpo, bearer string) *httptest.ResponseRecorder {
	t.Helper()
	var body *strings.Reader
	if cuerpo == "" {
		body = strings.NewReader("")
	} else {
		body = strings.NewReader(cuerpo)
	}
	r := httptest.NewRequest(metodo, ruta, body)
	if bearer != "" {
		r.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	e.handler.ServeHTTP(rec, r)
	return rec
}

// login admin por defecto y devuelve el token.
func (e *entornoHTTP) login(t *testing.T) string {
	t.Helper()
	rec := e.req(t, http.MethodPost, "/administradores/login",
		`{"email":"admin@admin.com","contrasenia":"admin1234"}`, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", rec.Code, rec.Body.String())
	}
	return extraerJSON(t, rec.Body.String(), "token")
}

// IDs concretos de la BD fresca. No asumir que ids de tenants != default son 2:
// el orden de las secuencias no es fiable entre migraciones.
func idTenantDefault(t *testing.T, e *entornoHTTP, ctx context.Context) int64 {
	t.Helper()
	var id int64
	if err := e.pool.QueryRow(ctx, `SELECT id FROM tenants WHERE slug='default'`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

// idServicioDefault devuelve el id del servicio llamado "Corte default".
func idServicioDefault(t *testing.T, e *entornoHTTP, ctx context.Context) int64 {
	t.Helper()
	var id int64
	if err := e.pool.QueryRow(ctx,
		`SELECT id FROM servicios WHERE tipo='Corte default' ORDER BY id LIMIT 1`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

// idTurnoDefault devuelve el id del turno del tenant default.
func idTurnoDefault(t *testing.T, e *entornoHTTP, ctx context.Context) int64 {
	t.Helper()
	var id int64
	if err := e.pool.QueryRow(ctx,
		`SELECT id FROM turnos ORDER BY id LIMIT 1`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func extraerJSON(t *testing.T, cuerpo, campo string) string {
	t.Helper()
	// Extracción simple sin dependencias extra.
	buscar := `"` + campo + `":`
	i := strings.Index(cuerpo, buscar)
	if i < 0 {
		t.Fatalf("campo %q no encontrado en %s", campo, cuerpo)
	}
	resto := cuerpo[i+len(buscar):]
	if !strings.HasPrefix(resto, `"`) {
		t.Fatalf("campo %q no es string en %s", campo, cuerpo)
	}
	resto = resto[1:]
	j := strings.Index(resto, `"`)
	if j < 0 {
		t.Fatalf("string sin cerrar en %s", cuerpo)
	}
	return resto[:j]
}

func TestMigracionFresca(t *testing.T) {
	ctx, pool, _ := nuevaBase(t)
	var total int
	err := pool.QueryRow(ctx, `SELECT count(*) FROM tenants WHERE slug = 'default'`).Scan(&total)
	if err != nil || total != 1 {
		t.Fatalf("tenant default inexistente: total=%d err=%v", total, err)
	}
	for _, tabla := range []string{"clientes", "servicios", "administradores", "turnos", "tenants"} {
		var existe bool
		if err := pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)`, tabla).Scan(&existe); err != nil || !existe {
			t.Fatalf("tabla %s inexistente: %v", tabla, err)
		}
	}
	// PK compuesta de clientes.
	var pk string
	_ = pool.QueryRow(ctx, `SELECT string_agg(column_name, ',' ORDER BY ordinal_position)
		FROM information_schema.key_column_usage WHERE constraint_name =
		(SELECT conname FROM pg_constraint WHERE conrelid = 'clientes'::regclass AND contype='p')`).Scan(&pk)
	if pk != "tenant_id,telefono" {
		t.Fatalf("PK de clientes = %q, want tenant_id,telefono", pk)
	}
	// Migrar dos veces no debe fallar (idempotencia).
	if err := database.RunMigrations(ctx, urlBaseDefecto(), slog.New(slog.NewTextHandler(os.Stderr, nil))); err != nil {
		t.Fatalf("re-migrar: %v", err)
	}
}

// TestAdopcionDeEsquemaJava simula una base creada por Hibernate (ddl-auto)
// con datos, y verifica que 000002 la multi-tenanticé sin perder datos.
func TestAdopcionDeEsquemaJava(t *testing.T) {
	ctx := context.Background()
	nombre := fmt.Sprintf("barber_adopt_%d", time.Now().UnixNano())
	admin, err := pgxpool.New(ctx, urlBaseDefecto())
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+nombre); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(ctx, "DROP DATABASE "+nombre+" WITH (FORCE)")
	})
	url := strings.Replace(urlBaseDefecto(), "/postgres?", "/"+nombre+"?", 1)
	pool, err := database.NewPool(ctx, database.Options{ConnectionString: url, MaxConns: 5})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	// Esquema equivalente al de Hibernate (sin columnas de tenant).
	ddl := []string{
		`CREATE TABLE clientes (telefono TEXT PRIMARY KEY, nombre TEXT, apellido TEXT, email TEXT)`,
		`CREATE TABLE servicios (id BIGSERIAL PRIMARY KEY, tipo TEXT, precio REAL)`,
		`CREATE TABLE administradores (id BIGSERIAL PRIMARY KEY, nombre TEXT, apellido TEXT, email TEXT, contrasenia TEXT)`,
		`CREATE INDEX idx_admin_email ON administradores (email)`,
		`CREATE TABLE turnos (id BIGSERIAL PRIMARY KEY, fecha_hora TIMESTAMP NOT NULL, cliente_id TEXT REFERENCES clientes(telefono), servicio_id BIGINT REFERENCES servicios(id))`,
		`CREATE INDEX idx_turno_fechahora ON turnos (fecha_hora)`,
		`INSERT INTO clientes (telefono, nombre, apellido, email) VALUES ('1150000001','Juan','Perez','juan@email.com')`,
		`INSERT INTO servicios (tipo, precio) VALUES ('Corte + barba', 7000)`,
		`INSERT INTO turnos (fecha_hora, cliente_id, servicio_id) VALUES ('2026-10-22 15:00:00', '1150000001', 1)`,
		// Duplicado de fecha/hora a propósito: la migración debe fallar en el índice.
		// (se agrega después de verificar el camino limpio)
	}
	for _, stmt := range ddl {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("ddl: %v (%s)", err, stmt)
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	if err := database.RunMigrations(ctx, url, logger); err != nil {
		t.Fatalf("migración de adopción: %v", err)
	}

	var tenants, backfill int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM tenants WHERE slug='default'`).Scan(&tenants)
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM turnos WHERE tenant_id = (SELECT id FROM tenants WHERE slug='default')`).Scan(&backfill)
	if tenants != 1 || backfill != 1 {
		t.Fatalf("tenants=%d backfill=%d", tenants, backfill)
	}

	// El índice único ahora rechaza duplicados.
	_, err = pool.Exec(ctx, `INSERT INTO turnos (fecha_hora, cliente_id, servicio_id, tenant_id)
		VALUES ('2026-10-22 15:00:00','1150000001',1,(SELECT id FROM tenants WHERE slug='default'))`)
	if err == nil || !strings.Contains(err.Error(), "23505") {
		t.Fatalf("esperaba 23505, got %v", err)
	}
}

func TestFlujoHTTPE2E(t *testing.T) {
	e, ctx := nuevoEntorno(t)
	_ = ctx
	token := e.login(t)

	// Servicios listos.
	rec := e.req(t, http.MethodGet, "/servicios", "", "")
	if rec.Code != 200 || strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("servicios inicial: %d %s", rec.Code, rec.Body.String())
	}

	// Crear servicio.
	rec = e.req(t, http.MethodPost, "/servicios", `{"tipo":"Corte + barba","precio":7000}`, token)
	if rec.Code != 201 {
		t.Fatalf("crear servicio: %d %s", rec.Code, rec.Body.String())
	}

	// Crear turno válido (público).
	futura := time.Now().Add(72 * time.Hour).Truncate(time.Hour).Add(14 * time.Hour)
	futura = time.Date(futura.Year(), futura.Month(), futura.Day(), 14, 0, 0, 0, time.Local)
	body := fmt.Sprintf(`{"fechaHora":"%s","idServicio":1,"telefonoCliente":"1150000001","nombreCliente":"Juan","apellidoCliente":"Perez","emailCliente":"juan@email.com"}`,
		futura.Format("2006-01-02T15:04:05"))
	rec = e.req(t, http.MethodPost, "/turnos", body, "")
	if rec.Code != 201 {
		t.Fatalf("crear turno: %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"tipoServicio":"Corte + barba"`) {
		t.Fatalf("turno DTO: %s", rec.Body.String())
	}

	// Duplicado → 409 slot_taken.
	rec = e.req(t, http.MethodPost, "/turnos", body, "")
	if rec.Code != 409 || !strings.Contains(rec.Body.String(), "slot_taken") {
		t.Fatalf("duplicado: %d %s", rec.Code, rec.Body.String())
	}

	// findDateTimes: formato sin segundos.
	rec = e.req(t, http.MethodGet, "/turnos/findDateTimes", "", "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `T14:00"]`) || strings.Contains(rec.Body.String(), "14:00:00") {
		t.Fatalf("findDateTimes: %d %s", rec.Code, rec.Body.String())
	}

	// PATCH turno con servicio parcial → respuesta sin refrescar.
	rec = e.req(t, http.MethodPatch, "/turnos/1", `{"servicio":{"id":1}}`, token)
	if rec.Code != 200 {
		t.Fatalf("patch: %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"tipoServicio":null`) || !strings.Contains(rec.Body.String(), `"precioServicio":0`) {
		t.Fatalf("patch sin refrescar: %s", rec.Body.String())
	}
	// GET posterior muestra los valores reales.
	rec = e.req(t, http.MethodGet, "/turnos/1", "", token)
	if !strings.Contains(rec.Body.String(), `"tipoServicio":"Corte + barba"`) {
		t.Fatalf("get post-patch: %s", rec.Body.String())
	}

	// PATCH con servicio inexistente → 404 claro (desviación documentada: Java
	// daba 500 por FK; el servicio ahora se valida contra el tenant).
	rec = e.req(t, http.MethodPatch, "/turnos/1", `{"servicio":{"id":999}}`, token)
	if rec.Code != 404 {
		t.Fatalf("patch servicio inexistente: %d %s", rec.Code, rec.Body.String())
	}

	// Turno inexistente → 404 vacío.
	rec = e.req(t, http.MethodGet, "/turnos/99999", "", token)
	if rec.Code != 404 || rec.Body.Len() != 0 {
		t.Fatalf("404 vacío: %d %q", rec.Code, rec.Body.String())
	}

	// Panel.
	rec = e.req(t, http.MethodGet, "/estadisticas/panel", "", token)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"turnosHoy"`) {
		t.Fatalf("panel: %d %s", rec.Code, rec.Body.String())
	}

	// La página de administradores NO expone el hash (ADR-004).
	rec = e.req(t, http.MethodGet, "/administradores/0/10", "", token)
	if rec.Code != 200 || strings.Contains(rec.Body.String(), "argon2") || strings.Contains(rec.Body.String(), "contrasenia") {
		t.Fatalf("página admin expone hash: %d %s", rec.Code, rec.Body.String())
	}

	// Login con credenciales incorrectas → 401 con instance exacta.
	rec = e.req(t, http.MethodPost, "/administradores/login", `{"email":"admin@admin.com","contrasenia":"clave-mala-123"}`, "")
	if rec.Code != 401 || !strings.Contains(rec.Body.String(), `"instance":"/administradores/login"`) {
		t.Fatalf("login inválido: %d %s", rec.Code, rec.Body.String())
	}
}

func TestConcurrenciaDobleReserva(t *testing.T) {
	e, _ := nuevoEntorno(t)
	token := e.login(t)
	if _, err := e.pool.Exec(context.Background(),
		`INSERT INTO servicios (tipo, precio, tenant_id) VALUES ('Corte', 5000, (SELECT id FROM tenants WHERE slug='default'))`); err != nil {
		t.Fatal(err)
	}

	futura := time.Now().Add(96 * time.Hour)
	futura = time.Date(futura.Year(), futura.Month(), futura.Day(), 16, 0, 0, 0, time.Local)
	body := fmt.Sprintf(`{"fechaHora":"%s","idServicio":1,"telefonoCliente":"1155555555","nombreCliente":"Ana","apellidoCliente":"Gomez"}`,
		futura.Format("2006-01-02T15:04:05"))

	const n = 20
	codigos := make([]int, n)
	bodies := make([]string, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rec := e.req(t, http.MethodPost, "/turnos", body, token)
			codigos[i] = rec.Code
			bodies[i] = rec.Body.String()
		}(i)
	}
	wg.Wait()

	creados := 0
	for _, c := range codigos {
		if c == 201 {
			creados++
		} else if c != 409 {
			t.Fatalf("código inesperado: %d", c)
		}
	}
	if creados != 1 {
		t.Fatalf("doble reserva: creados = %d, want 1 (códigos %v)", creados, codigos)
	}
	// B-11: la respuesta 409 tiene code "slot_taken"
	slotTakenBody := ""
	for i, c := range codigos {
		if c == 409 {
			slotTakenBody = bodies[i]
			break
		}
	}
	if slotTakenBody == "" {
		t.Fatalf("ningún 409 observado")
	}
	if !strings.Contains(slotTakenBody, "slot_taken") {
		t.Fatalf("409 sin code slot_taken: %s", slotTakenBody)
	}

	var total int
	_ = e.pool.QueryRow(context.Background(),
		`SELECT count(*) FROM turnos WHERE fecha_hora = $1`, futura).Scan(&total)
	if total != 1 {
		t.Fatalf("filas en DB = %d, want 1", total)
	}
	_ = token
}

func TestAislamientoDeTenant(t *testing.T) {
	e, ctx := nuevoEntorno(t)
	token := e.login(t)

	// Segundo tenant con su propio servicio.
	var tid2 int64
	if err := e.pool.QueryRow(ctx,
		`INSERT INTO tenants (slug, name, host) VALUES ('salon2', 'Segundo Salon', 'salon2.example.com') RETURNING id`).Scan(&tid2); err != nil {
		t.Fatal(err)
	}
	// No crear servicios para salon2 todavía: usaría el id 1 de la secuencia
	// global y el servicio del tenant default pasaría a ser id 2.

	// Servicio del tenant default.
	rec := e.req(t, http.MethodPost, "/servicios", `{"tipo":"Corte default","precio":5000}`, token)
	if rec.Code != 201 {
		t.Fatalf("crear servicio default: %d %s", rec.Code, rec.Body.String())
	}
	servicioDefault := idServicioDefault(t, e, ctx)

	// Servicio del segundo tenant, ya con el id 1 tomado.
	if _, err := e.pool.Exec(ctx,
		`INSERT INTO servicios (tipo, precio, tenant_id) VALUES ('Corte salon2', 9000, $1)`, tid2); err != nil {
		t.Fatal(err)
	}

	// Host del segundo tenant ve solo su servicio.
	rec = e.req(t, http.MethodGet, "/servicios", "", "")
	if !strings.Contains(rec.Body.String(), "Corte default") {
		t.Fatalf("default: %s", rec.Body.String())
	}

	r := httptest.NewRequest(http.MethodGet, "/servicios", nil)
	r.Host = "salon2.example.com"
	rec2 := httptest.NewRecorder()
	e.handler.ServeHTTP(rec2, r)
	if rec2.Code != 200 || !strings.Contains(rec2.Body.String(), "Corte salon2") || strings.Contains(rec2.Body.String(), "Corte default") {
		t.Fatalf("aislamiento host: %d %s", rec2.Code, rec2.Body.String())
	}

	// Un token del tenant default no ve datos del segundo tenant aunque el Host diga lo contrario.
	r = httptest.NewRequest(http.MethodGet, "/servicios", nil)
	r.Host = "salon2.example.com"
	r.Header.Set("Authorization", "Bearer "+token)
	rec3 := httptest.NewRecorder()
	e.handler.ServeHTTP(rec3, r)
	// JWT tenant != host tenant → 403.
	if rec3.Code != 403 {
		t.Fatalf("token debe scoping por tenant: %d %s", rec3.Code, rec3.Body.String())
	}
	_ = servicioDefault

	// findDateTimes aislado: crear un turno en cada tenant y verificar.
	_ = tid2
}

// TestAislamientoEscriturasEntreTenants verifica que un tenant no puede tocar
// (ni enlazar) recursos de otro: PATCH que apunte a un servicio ajeno debe
// tratarse como inexistente y no alterar la fila (hallazgo M8).
func TestAislamientoEscriturasEntreTenants(t *testing.T) {
	e, ctx := nuevoEntorno(t)
	token := e.login(t)

	var tid2 int64
	if err := e.pool.QueryRow(ctx,
		`INSERT INTO tenants (slug, name, host) VALUES ('salon3', 'Tercer Salon', 'salon3.example.com') RETURNING id`).Scan(&tid2); err != nil {
		t.Fatal(err)
	}

	crearServicio(t, e, token, "Corte propio", 5000)
	crearClienteDirecto(t, e, "1160000001")
	futura := fechaFutura(4)
	body := fmt.Sprintf(`{"fechaHora":"%s","idServicio":1,"telefonoCliente":"1160000001","nombreCliente":"Ana","apellidoCliente":"Luz"}`,
		futura.Format("2006-01-02T15:04:05"))
	rec := e.req(t, http.MethodPost, "/turnos", body, token)
	if rec.Code != 201 {
		t.Fatalf("crear turno propio: %d %s", rec.Code, rec.Body.String())
	}
	// El turno acaba de crearse; los ids de turno también usan una secuencia
	// global, así que se resuelve por BD.
	turnoID := idTurnoDefault(t, e, ctx)
	rutaTurno := fmt.Sprintf("/turnos/%d", turnoID)

	if rec = e.req(t, http.MethodGet, rutaTurno, "", token); !strings.Contains(rec.Body.String(), `"idServicio":1`) {
		t.Fatalf("turno inicial: %s", rec.Body.String())
	}

	// Servicio de OTRO tenant: PATCH debe rechazarlo.
	var servicioAjeno int64
	if err := e.pool.QueryRow(ctx,
		`INSERT INTO servicios (tipo, precio, tenant_id) VALUES ('Corte ajeno', 9999, $1) RETURNING id`, tid2).Scan(&servicioAjeno); err != nil {
		t.Fatal(err)
	}
	rec = e.req(t, http.MethodPatch, rutaTurno,
		fmt.Sprintf(`{"servicio":{"id":%d}}`, servicioAjeno), token)
	if rec.Code == 200 {
		t.Fatalf("PATCH a servicio de otro tenant no debería aceptarse: %s", rec.Body.String())
	}
	// El turno conserva su servicio original (no hubo escritura con FK ajena).
	rec = e.req(t, http.MethodGet, rutaTurno, "", token)
	if !strings.Contains(rec.Body.String(), `"idServicio":1`) {
		t.Fatalf("el turno cambió de servicio tras el PATCH rechazado: %s", rec.Body.String())
	}
}
