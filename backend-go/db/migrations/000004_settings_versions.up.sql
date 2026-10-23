-- 000004_settings_versions.up.sql
-- Historial append-only de versiones de marca (SPEC §8, ADR A6).

CREATE TABLE IF NOT EXISTS tenant_settings_versions (
  tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  version   INTEGER NOT NULL,
  hash      TEXT NOT NULL,
  settings  JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, version)
);
CREATE INDEX IF NOT EXISTS tenant_settings_versions_tenant_idx
  ON tenant_settings_versions (tenant_id, version DESC);

ALTER TABLE tenant_settings_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_settings_versions FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON tenant_settings_versions;
CREATE POLICY tenant_isolation ON tenant_settings_versions
  USING (tenant_id = (SELECT current_tenant_id()))
  WITH CHECK (tenant_id = (SELECT current_tenant_id()));

GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_settings_versions TO app_user;
