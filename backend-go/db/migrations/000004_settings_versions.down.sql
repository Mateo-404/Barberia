-- 000004_settings_versions.down.sql
DROP POLICY IF EXISTS tenant_isolation ON tenant_settings_versions;
ALTER TABLE tenant_settings_versions DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS tenant_settings_versions;
