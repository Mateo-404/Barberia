-- REFERENCIA: adaptá nombres/columnas a las tablas existentes (ver docs/DISCOVERY.md).
-- Ejecutar como owner/superusuario, NUNCA como app_user. Requiere que exista el rol app_user.

CREATE TABLE tenants (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  slug       text NOT NULL CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,38}[a-z0-9]$'),
  name       text NOT NULL CHECK (length(name) BETWEEN 2 AND 60),
  plan       text NOT NULL DEFAULT 'free'   CHECK (plan IN ('free','pro','business')),
  status     text NOT NULL DEFAULT 'active' CHECK (status IN ('active','suspended')),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX tenants_slug_key ON tenants (lower(slug));

CREATE TABLE tenant_domains (
  domain      text PRIMARY KEY CHECK (domain = lower(domain)),
  tenant_id   uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  verified_at timestamptz,
  created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tenant_domains_tenant_idx ON tenant_domains (tenant_id);

CREATE TABLE users (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id     uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  email         text NOT NULL CHECK (email = lower(email)),
  password_hash text NOT NULL,
  role          text NOT NULL DEFAULT 'staff' CHECK (role IN ('owner','admin','staff')),
  created_at    timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, email)
);

-- Tenant actual de la transacción. Vacío o ausente => NULL => la política no matchea ninguna fila (falla cerrado).
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS uuid
LANGUAGE sql STABLE AS $$
  SELECT nullif(current_setting('app.tenant_id', true), '')::uuid
$$;

-- Activa RLS + política de aislamiento en una tabla con columna tenant_id.
CREATE OR REPLACE FUNCTION enable_tenant_rls(tbl regclass) RETURNS void
LANGUAGE plpgsql AS $$
BEGIN
  EXECUTE format('ALTER TABLE %s ENABLE ROW LEVEL SECURITY', tbl);
  EXECUTE format('ALTER TABLE %s FORCE ROW LEVEL SECURITY', tbl);
  EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %s', tbl);
  EXECUTE format(
    'CREATE POLICY tenant_isolation ON %s USING (tenant_id = (SELECT current_tenant_id())) WITH CHECK (tenant_id = (SELECT current_tenant_id()))',
    tbl);
END $$;

SELECT enable_tenant_rls('users');

-- tenants y tenant_domains NO tienen RLS: se consultan ANTES de conocer el tenant (resolución por host).
-- app_user solo puede leerlas. El alta de tenants la hace un rol administrativo (fuera de alcance).
GRANT SELECT ON tenants, tenant_domains TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON users TO app_user;
