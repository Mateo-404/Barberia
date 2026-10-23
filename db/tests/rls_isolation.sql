-- Multi-tenant isolation test (RLS). Run as a superuser against a migrated database:
--   psql "$DB_SUPER_URL" -v ON_ERROR_STOP=1 -q -f db/tests/rls_isolation.sql
-- Any failure raises an exception => psql exits with a non-zero code.
\set ON_ERROR_STOP on
\echo '== rls_isolation =='

-- Cleanup and fixtures (superuser: ignores RLS)
DELETE FROM admins WHERE tenant_id IN (99001,99002);
DELETE FROM services WHERE tenant_id IN (99001,99002);
DELETE FROM customers WHERE tenant_id IN (99001,99002);
DELETE FROM appointments WHERE tenant_id IN (99001,99002);
DELETE FROM tenant_settings WHERE tenant_id IN (99001,99002);
DELETE FROM tenants WHERE id IN (99001,99002);
INSERT INTO tenants (id, slug, name) VALUES
  (99001,'tenant-a','Tenant A'),
  (99002,'tenant-b','Tenant B');
INSERT INTO services (tenant_id, name, price) VALUES
  (99001,'Cut A', 8000),
  (99001,'Beard A', 5000),
  (99002,'Cut B', 9000);
INSERT INTO tenant_settings (tenant_id, draft) VALUES
  (99001,'{"name":"A"}'),
  (99002,'{"name":"B"}');
-- The same email may exist in different tenants (uniqueness is per tenant).
INSERT INTO admins (tenant_id, first_name, last_name, email, password_hash) VALUES
  (99001,'Owner','A','owner@mail.com','x'),
  (99002,'Owner','B','owner@mail.com','x');

-- T-04: the app role cannot bypass RLS
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user' AND (rolsuper OR rolbypassrls)) THEN
    RAISE EXCEPTION 'T-04 FAIL: app_user is a superuser or has BYPASSRLS';
  END IF;
END $$;

-- T-07: every table with tenant_id (except tenant_domains) must have RLS enabled AND forced
DO $$
DECLARE t text;
BEGIN
  FOR t IN
    SELECT c.relname
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    JOIN pg_attribute a ON a.attrelid = c.oid AND a.attname = 'tenant_id' AND NOT a.attisdropped
    WHERE n.nspname = 'public' AND c.relkind = 'r'
      AND c.relname <> 'tenant_domains'
      AND NOT (c.relrowsecurity AND c.relforcerowsecurity)
  LOOP
    RAISE EXCEPTION 'T-07 FAIL: table % has tenant_id but no forced RLS', t;
  END LOOP;
END $$;

-- ===== Tx 1: no tenant => fails closed =====
BEGIN;
SET LOCAL ROLE app_user;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM services;         IF n <> 0 THEN RAISE EXCEPTION 'T-01 FAIL: without a tenant % services were visible', n; END IF;
  SELECT count(*) INTO n FROM tenant_settings;  IF n <> 0 THEN RAISE EXCEPTION 'T-01 FAIL: without a tenant % settings were visible', n; END IF;
  SELECT count(*) INTO n FROM admins;           IF n <> 0 THEN RAISE EXCEPTION 'T-01 FAIL: without a tenant % admins were visible', n; END IF;
END $$;

DO $$
BEGIN
  INSERT INTO services (tenant_id, name, price)
  VALUES (99001,'x',1);
  RAISE EXCEPTION 'T-01b FAIL: an insert succeeded without a tenant set';
EXCEPTION WHEN insufficient_privilege THEN NULL;
END $$;
COMMIT;

-- ===== Tx 2: tenant A =====
BEGIN;
SET LOCAL ROLE app_user;
SELECT set_config('app.tenant_id', '99001', true);

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM services;
  IF n <> 2 THEN RAISE EXCEPTION 'T-02 FAIL: A should see 2 services and saw %', n; END IF;
  SELECT count(*) INTO n FROM services WHERE tenant_id <> 99001;
  IF n <> 0 THEN RAISE EXCEPTION 'T-02 FAIL: A saw rows of another tenant'; END IF;
  SELECT count(*) INTO n FROM tenant_settings;
  IF n <> 1 THEN RAISE EXCEPTION 'T-02 FAIL: A should see 1 settings row and saw %', n; END IF;
  SELECT count(*) INTO n FROM admins;
  IF n <> 1 THEN RAISE EXCEPTION 'T-02 FAIL: A should see 1 admin and saw %', n; END IF;
END $$;

-- T-02b: UPDATE/DELETE over B's rows affect 0 rows
DO $$
DECLARE n int;
BEGIN
  UPDATE services SET name = 'hack' WHERE tenant_id = 99002;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 0 THEN RAISE EXCEPTION 'T-02b FAIL: A modified % rows of B', n; END IF;
  DELETE FROM services WHERE tenant_id = 99002;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 0 THEN RAISE EXCEPTION 'T-02b FAIL: A deleted % rows of B', n; END IF;
END $$;

-- T-03: WITH CHECK prevents inserting with another tenant's id
DO $$
BEGIN
  INSERT INTO services (tenant_id, name, price)
  VALUES (99002,'x',1);
  RAISE EXCEPTION 'T-03 FAIL: A inserted a row into B';
EXCEPTION WHEN insufficient_privilege THEN NULL;
END $$;

-- T-03b: WITH CHECK prevents "moving" an own row to another tenant
DO $$
BEGIN
  UPDATE services SET tenant_id = 99002
  WHERE tenant_id = 99001;
  RAISE EXCEPTION 'T-03b FAIL: A moved rows into B';
EXCEPTION WHEN insufficient_privilege THEN NULL;
END $$;
COMMIT;

-- ===== Tx 3: tenant B only sees its own =====
BEGIN;
SET LOCAL ROLE app_user;
SELECT set_config('app.tenant_id', '99002', true);
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM services;
  IF n <> 1 THEN RAISE EXCEPTION 'T-02 FAIL: B should see 1 service and saw %', n; END IF;
END $$;
COMMIT;

-- ===== T-05: the tenant does NOT leak to the next transaction of the same connection (pool) =====
BEGIN;
SET LOCAL ROLE app_user;
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM services;
  IF n <> 0 THEN RAISE EXCEPTION 'T-05 FAIL: the tenant leaked into the next transaction (% rows)', n; END IF;
END $$;
COMMIT;

-- ===== T-06a: empty tenant => 0 rows, no error =====
BEGIN;
SET LOCAL ROLE app_user;
SELECT set_config('app.tenant_id', '', true);
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM services;
  IF n <> 0 THEN RAISE EXCEPTION 'T-06a FAIL: an empty tenant saw % rows', n; END IF;
END $$;
COMMIT;

-- ===== T-06b: invalid tenant => error (fails closed, never "no filter") =====
BEGIN;
SET LOCAL ROLE app_user;
SELECT set_config('app.tenant_id', 'not-a-number', true);
DO $$
BEGIN
  PERFORM count(*) FROM services;
  RAISE EXCEPTION 'T-06b FAIL: an invalid tenant did not raise';
EXCEPTION
  WHEN invalid_text_representation OR others THEN NULL;
END $$;
COMMIT;

\echo '== rls_isolation: OK =='
