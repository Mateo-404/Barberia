-- Booking integrity test (uniqueness per tenant + starts_at).
--   psql "$DB_SUPER_URL" -v ON_ERROR_STOP=1 -q -f db/tests/booking_exclusion.sql
\set ON_ERROR_STOP on
\echo '== booking_exclusion (real schema) =='

-- cleanup
DELETE FROM appointments WHERE tenant_id IN (9901, 9902);
DELETE FROM services WHERE tenant_id IN (9901, 9902);
DELETE FROM customers WHERE tenant_id IN (9901, 9902);
DELETE FROM tenants WHERE id IN (9901, 9902);

INSERT INTO tenants (id, slug, name) VALUES (9901, 't-excl-1', 'Tenant Excl 1'), (9902, 't-excl-2', 'Tenant Excl 2');

-- base service and customer in tenant 1
INSERT INTO services (id, tenant_id, name, price) VALUES (99010, 9901, 'Cut', 1000);
INSERT INTO customers (phone, tenant_id, first_name, last_name) VALUES ('+5490000000001', 9901, 'Ana', 'Test');

BEGIN;
SET LOCAL ROLE app_user;
SELECT set_config('app.tenant_id', '9901', true);

-- B-00: base appointment at 10:00 (tenant 1)
INSERT INTO appointments (tenant_id, starts_at, customer_phone, service_id)
VALUES (9901, '2030-01-01 10:00:00', '+5490000000001', 99010);

-- B-01: same slot, same tenant => unique_violation (23505)
DO $$
BEGIN
  INSERT INTO appointments (tenant_id, starts_at, customer_phone, service_id)
  VALUES (9901, '2030-01-01 10:00:00', '+5490000000001', 99010);
  RAISE EXCEPTION 'B-01 FAIL: a duplicate appointment was accepted';
EXCEPTION WHEN unique_violation THEN NULL;
END $$;

-- B-02: a different slot (10:30) is valid
INSERT INTO appointments (tenant_id, starts_at, customer_phone, service_id)
VALUES (9901, '2030-01-01 10:30:00', '+5490000000001', 99010);

-- B-03: the same slot in another tenant is valid (isolation)
COMMIT;
BEGIN;
SET LOCAL ROLE app_user;
SELECT set_config('app.tenant_id', '9902', true);
-- service + customer in tenant 2
INSERT INTO services (id, tenant_id, name, price) VALUES (99020, 9902, 'Cut', 1000);
INSERT INTO customers (phone, tenant_id, first_name, last_name) VALUES ('+5490000000002', 9902, 'Beto', 'Test');
INSERT INTO appointments (tenant_id, starts_at, customer_phone, service_id)
VALUES (9902, '2030-01-01 10:00:00', '+5490000000002', 99020);
COMMIT;

-- B-04: cross-tenant FK (another tenant's service) => foreign_key_violation (23503)
BEGIN;
SET LOCAL ROLE app_user;
SELECT set_config('app.tenant_id', '9901', true);
DO $$
BEGIN
  INSERT INTO appointments (tenant_id, starts_at, customer_phone, service_id)
  VALUES (9901, '2030-01-01 11:00:00', '+5490000000001', 99020);
  RAISE EXCEPTION 'B-04 FAIL: a cross-tenant FK was accepted';
EXCEPTION WHEN foreign_key_violation THEN NULL;
END $$;
COMMIT;

-- final cleanup
DELETE FROM appointments WHERE tenant_id IN (9901, 9902);
DELETE FROM services WHERE tenant_id IN (9901, 9902);
DELETE FROM customers WHERE tenant_id IN (9901, 9902);
DELETE FROM tenants WHERE id IN (9901, 9902);

\echo '== booking_exclusion (real schema): OK =='
