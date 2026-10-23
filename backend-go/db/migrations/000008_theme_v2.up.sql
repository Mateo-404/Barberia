-- 000008_theme_v2.up.sql
-- Theme v2: business data leaves the versioned settings document and the
-- owner gets two preferences for the admin panel.
--
-- theme_editor_enabled  show the Theme menu (editor) in the admin panel
-- onboarding_enabled    show the welcome wizard on the first login
-- business              operating data (hours, currency, contact), saved
--                       directly: not versioned, not published
--
-- Existing draft/published documents stay as stored (v1); the application
-- upgrades them on read, so old versions remain valid for rollback.

ALTER TABLE tenant_settings
    ADD COLUMN IF NOT EXISTS theme_editor_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS onboarding_enabled   BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS business             JSONB;

UPDATE tenant_settings
SET business = COALESCE(published -> 'biz', draft -> 'biz')
WHERE business IS NULL;
