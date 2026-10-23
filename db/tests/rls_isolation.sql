-- Test de aislamiento multi-tenant (RLS). Ejecutar como superusuario contra una DB migrada:
--   psql "$DB_SUPER_URL" -v ON_ERROR_STOP=1 -q -f db/tests/rls_isolation.sql
-- Cualquier fallo lanza una excepción => psql sale con código != 0.
\set ON_ERROR_STOP on
\echo '== rls_isolation =='

-- Limpieza previa y fixtures (superusuario: ignora RLS)
DELETE FROM administradores WHERE tenant_id IN (99001,99002);
DELETE FROM servicios WHERE tenant_id IN (99001,99002);
DELETE FROM clientes WHERE tenant_id IN (99001,99002);
DELETE FROM turnos WHERE tenant_id IN (99001,99002);
DELETE FROM tenant_settings WHERE tenant_id IN (99001,99002);
DELETE FROM tenants WHERE id IN (99001,99002);
INSERT INTO tenants (id, slug, name) VALUES
  (99001,'tenant-a','Tenant A'),
  (99002,'tenant-b','Tenant B');
INSERT INTO servicios (tenant_id, tipo, precio) VALUES
  (99001,'Corte A', 8000),
  (99001,'Barba A',  5000),
  (99002,'Corte B',  9000);
INSERT INTO tenant_settings (tenant_id, draft) VALUES
  (99001,'{"name":"A"}'),
  (99002,'{"name":"B"}');
-- El mismo email puede existir en tenants distintos (unicidad por tenant).
INSERT INTO administradores (tenant_id, nombre, apellido, email, contrasenia) VALUES
  (99001,'Dueno','A','dueno@mail.com','x'),
  (99002,'Dueno','B','dueno@mail.com','x');

-- T-04: el rol de la app no puede saltarse RLS
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user' AND (rolsuper OR rolbypassrls)) THEN
    RAISE EXCEPTION 'T-04 FAIL: app_user es superusuario o tiene BYPASSRLS';
  END IF;
END $$;

-- T-07: toda tabla con tenant_id (salvo tenant_domains) debe tener RLS habilitada Y forzada
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
    RAISE EXCEPTION 'T-07 FAIL: la tabla % tiene tenant_id pero no RLS forzada', t;
  END LOOP;
END $$;

-- ===== Tx 1: sin tenant => falla cerrado =====
BEGIN;
SET LOCAL ROLE app_user;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM servicios;         IF n <> 0 THEN RAISE EXCEPTION 'T-01 FAIL: sin tenant se vieron % servicios', n; END IF;
  SELECT count(*) INTO n FROM tenant_settings;  IF n <> 0 THEN RAISE EXCEPTION 'T-01 FAIL: sin tenant se vieron % settings', n; END IF;
  SELECT count(*) INTO n FROM administradores;  IF n <> 0 THEN RAISE EXCEPTION 'T-01 FAIL: sin tenant se vieron % administradores', n; END IF;
END $$;

DO $$
BEGIN
  INSERT INTO servicios (tenant_id, tipo, precio)
  VALUES (99001,'x',1);
  RAISE EXCEPTION 'T-01b FAIL: se pudo insertar sin tenant seteado';
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
  SELECT count(*) INTO n FROM servicios;
  IF n <> 2 THEN RAISE EXCEPTION 'T-02 FAIL: A debería ver 2 servicios y vio %', n; END IF;
  SELECT count(*) INTO n FROM servicios WHERE tenant_id <> 99001;
  IF n <> 0 THEN RAISE EXCEPTION 'T-02 FAIL: A vio filas de otro tenant'; END IF;
  SELECT count(*) INTO n FROM tenant_settings;
  IF n <> 1 THEN RAISE EXCEPTION 'T-02 FAIL: A debería ver 1 settings y vio %', n; END IF;
  SELECT count(*) INTO n FROM administradores;
  IF n <> 1 THEN RAISE EXCEPTION 'T-02 FAIL: A debería ver 1 administrador y vio %', n; END IF;
END $$;

-- T-02b: UPDATE/DELETE sobre filas de B afectan 0 filas
DO $$
DECLARE n int;
BEGIN
  UPDATE servicios SET tipo = 'hack' WHERE tenant_id = 99002;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 0 THEN RAISE EXCEPTION 'T-02b FAIL: A modificó % filas de B', n; END IF;
  DELETE FROM servicios WHERE tenant_id = 99002;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 0 THEN RAISE EXCEPTION 'T-02b FAIL: A borró % filas de B', n; END IF;
END $$;

-- T-03: WITH CHECK impide insertar con tenant_id ajeno
DO $$
BEGIN
  INSERT INTO servicios (tenant_id, tipo, precio)
  VALUES (99002,'x',1);
  RAISE EXCEPTION 'T-03 FAIL: A insertó una fila en B';
EXCEPTION WHEN insufficient_privilege THEN NULL;
END $$;

-- T-03b: WITH CHECK impide "mover" una fila propia a otro tenant
DO $$
BEGIN
  UPDATE servicios SET tenant_id = 99002
  WHERE tenant_id = 99001;
  RAISE EXCEPTION 'T-03b FAIL: A movió filas hacia B';
EXCEPTION WHEN insufficient_privilege THEN NULL;
END $$;
COMMIT;

-- ===== Tx 3: tenant B ve solo lo suyo =====
BEGIN;
SET LOCAL ROLE app_user;
SELECT set_config('app.tenant_id', '99002', true);
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM servicios;
  IF n <> 1 THEN RAISE EXCEPTION 'T-02 FAIL: B debería ver 1 servicio y vio %', n; END IF;
END $$;
COMMIT;

-- ===== T-05: el tenant NO se filtra a la siguiente transacción de la misma conexión (pool) =====
BEGIN;
SET LOCAL ROLE app_user;
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM servicios;
  IF n <> 0 THEN RAISE EXCEPTION 'T-05 FAIL: el tenant se filtró a la transacción siguiente (% filas)', n; END IF;
END $$;
COMMIT;

-- ===== T-06a: tenant vacío => 0 filas, sin error =====
BEGIN;
SET LOCAL ROLE app_user;
SELECT set_config('app.tenant_id', '', true);
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM servicios;
  IF n <> 0 THEN RAISE EXCEPTION 'T-06a FAIL: tenant vacío vio % filas', n; END IF;
END $$;
COMMIT;

-- ===== T-06b: tenant inválido => error (falla cerrado, nunca "sin filtro") =====
BEGIN;
SET LOCAL ROLE app_user;
SELECT set_config('app.tenant_id', 'no-es-un-numero', true);
DO $$
BEGIN
  PERFORM count(*) FROM servicios;
  RAISE EXCEPTION 'T-06b FAIL: tenant inválido no lanzó error';
EXCEPTION
  WHEN invalid_text_representation OR others THEN NULL;
END $$;
COMMIT;

\echo '== rls_isolation: OK =='
