-- Test de integridad de reservas (unicidad por tenant + fecha_hora).
--   psql "$DB_SUPER_URL" -v ON_ERROR_STOP=1 -q -f db/tests/booking_exclusion.sql
\set ON_ERROR_STOP on
\echo '== booking_exclusion (real schema) =='

-- limpieza
DELETE FROM turnos WHERE tenant_id IN (9901, 9902);
DELETE FROM servicios WHERE tenant_id IN (9901, 9902);
DELETE FROM clientes WHERE tenant_id IN (9901, 9902);
DELETE FROM tenants WHERE id IN (9901, 9902);

INSERT INTO tenants (id, slug, name) VALUES (9901, 't-excl-1', 'Tenant Excl 1'), (9902, 't-excl-2', 'Tenant Excl 2');

-- servicio y cliente base en tenant 1
INSERT INTO servicios (id, tenant_id, tipo, precio) VALUES (99010, 9901, 'Corte', 1000);
INSERT INTO clientes (telefono, tenant_id, nombre, apellido) VALUES ('+5490000000001', 9901, 'Ana', 'Test');

BEGIN;
SET LOCAL ROLE app_user;
SELECT set_config('app.tenant_id', '9901', true);

-- B-00: turno base 10:00 (tenant 1)
INSERT INTO turnos (tenant_id, fecha_hora, cliente_id, servicio_id)
VALUES (9901, '2030-01-01 10:00:00', '+5490000000001', 99010);

-- B-01: mismo slot, mismo tenant => unique_violation (23505)
DO $$
BEGIN
  INSERT INTO turnos (tenant_id, fecha_hora, cliente_id, servicio_id)
  VALUES (9901, '2030-01-01 10:00:00', '+5490000000001', 99010);
  RAISE EXCEPTION 'B-01 FAIL: se aceptó un turno duplicado';
EXCEPTION WHEN unique_violation THEN NULL;
END $$;

-- B-02: slot distinto (10:30) es válido
INSERT INTO turnos (tenant_id, fecha_hora, cliente_id, servicio_id)
VALUES (9901, '2030-01-01 10:30:00', '+5490000000001', 99010);

-- B-03: mismo slot en otro tenant es válido (aislamiento)
-- (cambiamos de tenant)
COMMIT;
BEGIN;
SET LOCAL ROLE app_user;
SELECT set_config('app.tenant_id', '9902', true);
-- servicio + cliente en tenant 2
INSERT INTO servicios (id, tenant_id, tipo, precio) VALUES (99020, 9902, 'Corte', 1000);
INSERT INTO clientes (telefono, tenant_id, nombre, apellido) VALUES ('+5490000000002', 9902, 'Beto', 'Test');
INSERT INTO turnos (tenant_id, fecha_hora, cliente_id, servicio_id)
VALUES (9902, '2030-01-01 10:00:00', '+5490000000002', 99020);
COMMIT;

-- B-04: FK entre tenants (servicio de otro tenant) => foreign_key_violation (23503)
BEGIN;
SET LOCAL ROLE app_user;
SELECT set_config('app.tenant_id', '9901', true);
DO $$
BEGIN
  INSERT INTO turnos (tenant_id, fecha_hora, cliente_id, servicio_id)
  VALUES (9901, '2030-01-01 11:00:00', '+5490000000001', 99020);
  RAISE EXCEPTION 'B-04 FAIL: se aceptó FK cross-tenant';
EXCEPTION WHEN foreign_key_violation THEN NULL;
END $$;
COMMIT;

-- limpieza final
DELETE FROM turnos WHERE tenant_id IN (9901, 9902);
DELETE FROM servicios WHERE tenant_id IN (9901, 9902);
DELETE FROM clientes WHERE tenant_id IN (9901, 9902);
DELETE FROM tenants WHERE id IN (9901, 9902);

\echo '== booking_exclusion (real schema): OK =='
