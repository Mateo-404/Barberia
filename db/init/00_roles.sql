-- DEVELOPMENT/TEST ONLY (docker-entrypoint-initdb.d).
-- In production: create the role with a secret-managed password, without SUPERUSER or BYPASSRLS.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user') THEN
    CREATE ROLE app_user LOGIN PASSWORD 'app_user_dev'
      NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE;
  ELSE
    -- Migration 000003 creates app_user without LOGIN when it did not exist.
    ALTER ROLE app_user LOGIN PASSWORD 'app_user_dev';
  END IF;
END $$;
