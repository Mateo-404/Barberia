-- 000008_theme_v2.down.sql
-- The business data of v1 documents is still inside draft/published, so
-- dropping the columns loses only what was saved after the migration.
ALTER TABLE tenant_settings
    DROP COLUMN IF EXISTS business,
    DROP COLUMN IF EXISTS onboarding_enabled,
    DROP COLUMN IF EXISTS theme_editor_enabled;
