-- 000003_tenancy_rls.up.sql
-- Adaptado al esquema real BIGINT (mantener estructura existente). Ejecutar con rol dueño.

-- tenant_domains para resolución por host (SPEC §4)
CREATE TABLE IF NOT EXISTS tenant_domains (
  domain      TEXT PRIMARY KEY CHECK (domain = lower(domain)),
  tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  verified_at TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS tenant_domains_tenant_idx ON tenant_domains (tenant_id);

-- tenant_settings para Fase 2
CREATE TABLE IF NOT EXISTS tenant_settings (
  tenant_id        BIGINT PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
  draft            JSONB,
  published        JSONB,
  published_version INTEGER NOT NULL DEFAULT 0,
  onboarded        BOOLEAN NOT NULL DEFAULT FALSE,
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  published_at     TIMESTAMPTZ
);

-- current_tenant_id() devuelve BIGINT
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS BIGINT
LANGUAGE sql STABLE AS $$
  SELECT nullif(current_setting('app.tenant_id', true), '')::bigint
$$;

-- helper para habilitar RLS forzado
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

-- Habilitar RLS en tablas con tenant_id (salvo tenant_domains)
SELECT enable_tenant_rls('clientes');
SELECT enable_tenant_rls('servicios');
SELECT enable_tenant_rls('administradores');
SELECT enable_tenant_rls('turnos');
SELECT enable_tenant_rls('tenant_settings');

-- Rol app_user sin SUPERUSER/BYPASSRLS
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user') THEN
    CREATE ROLE app_user;
  END IF;
END $$;

-- Asegurar sin privilegios indebidos (no forzamos NOLOGIN aquí; se gestiona por despliegue)
ALTER ROLE app_user NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;

-- Grants sobre tablas existentes
GRANT USAGE ON SCHEMA public TO app_user;
GRANT SELECT ON tenants, tenant_domains TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON clientes, servicios, administradores, turnos, tenant_settings TO app_user;

-- Grants sobre secuencias
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO app_user;

-- Default privileges para futuras tablas/secuencias
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO app_user;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO app_user;
