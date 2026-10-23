-- 000003_tenancy_rls.down.sql (reversible)
DROP POLICY IF EXISTS tenant_isolation ON tenant_settings;
DROP POLICY IF EXISTS tenant_isolation ON turnos;
DROP POLICY IF EXISTS tenant_isolation ON administradores;
DROP POLICY IF EXISTS tenant_isolation ON servicios;
DROP POLICY IF EXISTS tenant_isolation ON clientes;

ALTER TABLE tenant_settings DISABLE ROW LEVEL SECURITY;
ALTER TABLE turnos DISABLE ROW LEVEL SECURITY;
ALTER TABLE administradores DISABLE ROW LEVEL SECURITY;
ALTER TABLE servicios DISABLE ROW LEVEL SECURITY;
ALTER TABLE clientes DISABLE ROW LEVEL SECURITY;

DROP TABLE IF EXISTS tenant_settings;
DROP TABLE IF EXISTS tenant_domains;

DROP FUNCTION IF EXISTS enable_tenant_rls(regclass);
DROP FUNCTION IF EXISTS current_tenant_id();
