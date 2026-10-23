-- 000009_auth_google_totp.up.sql
-- Google OAuth + TOTP two-factor authentication for admin panel users.
-- Customer (public) accounts get a minimal google_sub so they can sign in
-- with Google and pre-fill the booking form.

-- Admins: allow password-less (Google) accounts and TOTP.
ALTER TABLE admins
    ADD COLUMN IF NOT EXISTS google_sub      TEXT,
    ADD COLUMN IF NOT EXISTS auth_provider    TEXT NOT NULL DEFAULT 'password',
    ADD COLUMN IF NOT EXISTS totp_secret      TEXT,
    ADD COLUMN IF NOT EXISTS totp_enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS totp_verified_at TIMESTAMPTZ;

ALTER TABLE admins ALTER COLUMN password_hash DROP NOT NULL;

DO
$$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_admins_auth_provider') THEN
        ALTER TABLE admins ADD CONSTRAINT chk_admins_auth_provider
            CHECK (auth_provider IN ('password', 'google', 'both'));
    END IF;
END
$$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_admins_tenant_google
    ON admins (tenant_id, google_sub)
    WHERE google_sub IS NOT NULL;

-- Customers: optional Google identity.
ALTER TABLE customers
    ADD COLUMN IF NOT EXISTS google_sub  TEXT,
    ADD COLUMN IF NOT EXISTS avatar_url  TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS uq_customers_tenant_google
    ON customers (tenant_id, google_sub)
    WHERE google_sub IS NOT NULL;

-- RLS: customers already has the tenant policy (migration 000003).
-- Grant new columns are on existing tables, so no new grants needed;
-- but keep the sequence grants consistent (no new sequences here).
