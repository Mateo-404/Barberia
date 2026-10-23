-- SOLO DESARROLLO/TEST (docker-entrypoint-initdb.d).
-- En producción: creá el rol con contraseña gestionada por secretos, sin SUPERUSER ni BYPASSRLS.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user') THEN
    CREATE ROLE app_user LOGIN PASSWORD 'app_user_dev'
      NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE;
  ELSE
    -- La migración 000003 crea app_user sin LOGIN si aún no existía.
    ALTER ROLE app_user LOGIN PASSWORD 'app_user_dev';
  END IF;
END $$;
