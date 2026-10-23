//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database"
)

func TestFreshMigration(t *testing.T) {
	pool, url := migratedDatabase(t)
	ctx := context.Background()

	var defaults int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM tenants WHERE slug = 'default'`).Scan(&defaults); err != nil || defaults != 1 {
		t.Fatalf("default tenant: count=%d err=%v", defaults, err)
	}
	for _, table := range []string{"customers", "services", "admins", "appointments", "tenants", "tenant_settings"} {
		var exists bool
		if err := pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)`, table).Scan(&exists); err != nil || !exists {
			t.Fatalf("table %s missing: %v", table, err)
		}
	}
	var pk string
	_ = pool.QueryRow(ctx, `SELECT string_agg(column_name, ',' ORDER BY ordinal_position)
		FROM information_schema.key_column_usage WHERE constraint_name =
		(SELECT conname FROM pg_constraint WHERE conrelid = 'customers'::regclass AND contype = 'p')`).Scan(&pk)
	if pk != "tenant_id,phone" {
		t.Fatalf("customers PK = %q, want tenant_id,phone", pk)
	}
	// Migrating twice is a no-op.
	if err := database.RunMigrations(ctx, url, quietLogger()); err != nil {
		t.Fatalf("re-migrate: %v", err)
	}
}

// TestAdoptsHibernateSchema simulates a database created by the former
// Spring/Hibernate backend (Spanish names, no tenant columns) with data, and
// checks the migrations make it multi-tenant and English without data loss.
func TestAdoptsHibernateSchema(t *testing.T) {
	url := emptyDatabase(t)
	pool := openPool(t, url)
	ctx := context.Background()

	for _, stmt := range []string{
		`CREATE TABLE clientes (telefono TEXT PRIMARY KEY, nombre TEXT, apellido TEXT, email TEXT)`,
		`CREATE TABLE servicios (id BIGSERIAL PRIMARY KEY, tipo TEXT, precio REAL)`,
		`CREATE TABLE administradores (id BIGSERIAL PRIMARY KEY, nombre TEXT, apellido TEXT, email TEXT, contrasenia TEXT)`,
		`CREATE INDEX idx_admin_email ON administradores (email)`,
		`CREATE TABLE turnos (id BIGSERIAL PRIMARY KEY, fecha_hora TIMESTAMP NOT NULL, cliente_id TEXT REFERENCES clientes(telefono), servicio_id BIGINT REFERENCES servicios(id))`,
		`CREATE INDEX idx_turno_fechahora ON turnos (fecha_hora)`,
		`INSERT INTO clientes VALUES ('1150000001', 'Juan', 'Perez', 'juan@example.com')`,
		`INSERT INTO servicios (tipo, precio) VALUES ('Cut + beard', 7000)`,
		`INSERT INTO turnos (fecha_hora, cliente_id, servicio_id) VALUES ('2026-10-22 15:00:00', '1150000001', 1)`,
	} {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("ddl: %v (%s)", err, stmt)
		}
	}

	if err := database.RunMigrations(ctx, url, quietLogger()); err != nil {
		t.Fatalf("adoption migration: %v", err)
	}

	var phone, firstName, serviceName string
	if err := pool.QueryRow(ctx, `
		SELECT c.phone, c.first_name, s.name
		FROM appointments a
		JOIN customers c ON c.phone = a.customer_phone AND c.tenant_id = a.tenant_id
		JOIN services s ON s.id = a.service_id
		WHERE a.tenant_id = (SELECT id FROM tenants WHERE slug = 'default')`).Scan(&phone, &firstName, &serviceName); err != nil {
		t.Fatalf("data lost during adoption: %v", err)
	}
	if phone != "1150000001" || firstName != "Juan" || serviceName != "Cut + beard" {
		t.Fatalf("unexpected data: %s %s %s", phone, firstName, serviceName)
	}

	// The unique index now rejects double bookings.
	_, err := pool.Exec(ctx, `INSERT INTO appointments (starts_at, customer_phone, service_id, tenant_id)
		VALUES ('2026-10-22 15:00:00', '1150000001', 1, (SELECT id FROM tenants WHERE slug = 'default'))`)
	if err == nil || !strings.Contains(err.Error(), "23505") {
		t.Fatalf("expected 23505, got %v", err)
	}
}

// TestEnglishSchemaMigrationIsReversible runs 000007 down and up again on a
// database with data: names round-trip and no row is lost.
func TestEnglishSchemaMigrationIsReversible(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	serviceID := e.createService(t, token, "Cut", 5000)
	e.call(t, "POST", "/api/appointments", bookingBody(futureAt(3, 15, 0), serviceID, "1150000009", "Ana", "Paz"), "").
		expect(201, "book")

	ctx := context.Background()
	down, err := readMigration("000007_english_schema.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	up, err := readMigration("000007_english_schema.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.pool.Exec(ctx, down); err != nil {
		t.Fatalf("down: %v", err)
	}
	var spanish int
	if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM turnos t JOIN clientes c ON c.telefono = t.cliente_id`).Scan(&spanish); err != nil || spanish != 1 {
		t.Fatalf("after down: rows=%d err=%v", spanish, err)
	}
	if _, err := e.pool.Exec(ctx, up); err != nil {
		t.Fatalf("up: %v", err)
	}
	var english int
	if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM appointments a JOIN customers c ON c.phone = a.customer_phone`).Scan(&english); err != nil || english != 1 {
		t.Fatalf("after up: rows=%d err=%v", english, err)
	}
}

func readMigration(name string) (string, error) {
	b, err := migrationsFS.ReadFile("migrations/" + name)
	return string(b), err
}
