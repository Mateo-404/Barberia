-- Configuración de marca por tenant con borrador/publicación e historial para rollback.
CREATE TABLE tenant_settings (
  tenant_id         uuid PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
  draft             jsonb NOT NULL,
  published         jsonb,
  published_version int  NOT NULL DEFAULT 0,
  published_hash    text,
  onboarded         boolean NOT NULL DEFAULT false,
  updated_at        timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tenant_settings_versions (
  tenant_id  uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  version    int  NOT NULL,
  data       jsonb NOT NULL,
  hash       text NOT NULL,
  created_by uuid,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, version)
);

SELECT enable_tenant_rls('tenant_settings');
SELECT enable_tenant_rls('tenant_settings_versions');

GRANT SELECT, INSERT, UPDATE ON tenant_settings TO app_user;
GRANT SELECT, INSERT ON tenant_settings_versions TO app_user; -- historial append-only
