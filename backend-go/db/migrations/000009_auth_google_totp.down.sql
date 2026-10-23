-- 000009_auth_google_totp.down.sql
DROP INDEX IF EXISTS uq_customers_tenant_google;
ALTER TABLE customers DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE customers DROP COLUMN IF EXISTS google_sub;

DROP INDEX IF EXISTS uq_admins_tenant_google;
ALTER TABLE admins DROP CONSTRAINT IF EXISTS chk_admins_auth_provider;
ALTER TABLE admins DROP COLUMN IF EXISTS totp_verified_at;
ALTER TABLE admins DROP COLUMN IF EXISTS totp_enabled;
ALTER TABLE admins DROP COLUMN IF EXISTS totp_secret;
ALTER TABLE admins DROP COLUMN IF EXISTS auth_provider;
ALTER TABLE admins DROP COLUMN IF EXISTS google_sub;
ALTER TABLE admins ALTER COLUMN password_hash SET NOT NULL;
