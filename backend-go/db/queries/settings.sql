-- name: GetTenantSettings :one
SELECT tenant_id, draft, published, published_version, onboarded, updated_at, published_at, theme_editor_enabled, onboarding_enabled, business
FROM tenant_settings
WHERE tenant_id = sqlc.arg('tenant_id');

-- name: UpsertTenantSettingsDraft :one
INSERT INTO tenant_settings (tenant_id, draft, updated_at)
VALUES (sqlc.arg('tenant_id'), sqlc.arg('draft'), now())
ON CONFLICT (tenant_id) DO UPDATE
SET draft = EXCLUDED.draft, updated_at = now()
RETURNING tenant_id, draft, published, published_version, onboarded, updated_at, published_at, theme_editor_enabled, onboarding_enabled, business;

-- name: PublishTenantSettings :one
UPDATE tenant_settings
SET published = draft,
    published_version = published_version + 1,
    onboarded = TRUE,
    published_at = now(),
    updated_at = now()
WHERE tenant_id = sqlc.arg('tenant_id')
RETURNING tenant_id, draft, published, published_version, onboarded, updated_at, published_at, theme_editor_enabled, onboarding_enabled, business;

-- name: ListSettingsVersions :many
SELECT version, hash, created_at
FROM tenant_settings_versions
WHERE tenant_id = sqlc.arg('tenant_id')
ORDER BY version DESC
LIMIT 20;

-- name: InsertSettingsVersion :one
INSERT INTO tenant_settings_versions (tenant_id, version, hash, settings)
VALUES (sqlc.arg('tenant_id'), sqlc.arg('version'), sqlc.arg('hash'), sqlc.arg('settings'))
ON CONFLICT (tenant_id, version) DO UPDATE
SET hash = EXCLUDED.hash, settings = EXCLUDED.settings
RETURNING version, hash, created_at;

-- name: GetSettingsVersion :one
SELECT version, hash, settings, created_at
FROM tenant_settings_versions
WHERE tenant_id = sqlc.arg('tenant_id') AND version = sqlc.arg('version');

-- name: UpdateSettingsPreferences :one
-- Creates the row when the tenant has none; null arguments keep the value.
INSERT INTO tenant_settings (tenant_id, theme_editor_enabled, onboarding_enabled)
VALUES (
    sqlc.arg('tenant_id'),
    COALESCE(sqlc.narg('theme_editor_enabled')::boolean, TRUE),
    COALESCE(sqlc.narg('onboarding_enabled')::boolean, TRUE)
)
ON CONFLICT (tenant_id) DO UPDATE
SET theme_editor_enabled = COALESCE(sqlc.narg('theme_editor_enabled')::boolean, tenant_settings.theme_editor_enabled),
    onboarding_enabled   = COALESCE(sqlc.narg('onboarding_enabled')::boolean, tenant_settings.onboarding_enabled),
    updated_at = now()
RETURNING tenant_id, draft, published, published_version, onboarded, updated_at, published_at, theme_editor_enabled, onboarding_enabled, business;

-- name: UpsertBusiness :one
INSERT INTO tenant_settings (tenant_id, business)
VALUES (sqlc.arg('tenant_id'), sqlc.arg('business'))
ON CONFLICT (tenant_id) DO UPDATE
SET business = EXCLUDED.business, updated_at = now()
RETURNING tenant_id, draft, published, published_version, onboarded, updated_at, published_at, theme_editor_enabled, onboarding_enabled, business;
