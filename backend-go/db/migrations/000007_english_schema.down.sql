-- 000007_english_schema.down.sql
-- Restores the pre-000007 (Spanish) names. Pure renames, no data loss.

ALTER INDEX IF EXISTS idx_appointments_starts_at RENAME TO idx_turno_fechahora;
ALTER INDEX IF EXISTS uq_appointments_tenant_starts_at RENAME TO uq_turno_tenant_fecha;

CREATE OR REPLACE FUNCTION pg_temp.to_spanish(name TEXT) RETURNS TEXT
LANGUAGE sql IMMUTABLE AS $$
  SELECT replace(replace(replace(replace(replace(replace(replace(replace(replace(
    name,
    'admins', 'administradores'),
    'customer_phone', 'cliente_id'),
    'customers', 'clientes'),
    'customer', 'cliente'),
    'services', 'servicios'),
    'service', 'servicio'),
    'appointments', 'turnos'),
    'appointment', 'turno'),
    'starts_at', 'fecha_hora')
$$;

DO $$
DECLARE
  r RECORD;
BEGIN
  FOR r IN
    SELECT c.conrelid::regclass AS tbl, c.conname
    FROM pg_constraint c
    WHERE c.conrelid IN ('appointments'::regclass, 'customers'::regclass,
                         'services'::regclass, 'admins'::regclass)
      AND c.conname <> pg_temp.to_spanish(c.conname)
  LOOP
    EXECUTE format('ALTER TABLE %s RENAME CONSTRAINT %I TO %I',
                   r.tbl, r.conname, replace(pg_temp.to_spanish(r.conname), '_phone', '_telefono'));
  END LOOP;

  FOR r IN
    SELECT indexname
    FROM pg_indexes
    WHERE schemaname = 'public'
      AND tablename IN ('appointments', 'customers', 'services', 'admins')
      AND indexname <> pg_temp.to_spanish(indexname)
  LOOP
    EXECUTE format('ALTER INDEX %I RENAME TO %I',
                   r.indexname, replace(pg_temp.to_spanish(r.indexname), '_phone', '_telefono'));
  END LOOP;

  FOR r IN
    SELECT sequencename
    FROM pg_sequences
    WHERE schemaname = 'public'
      AND sequencename <> pg_temp.to_spanish(sequencename)
  LOOP
    EXECUTE format('ALTER SEQUENCE %I RENAME TO %I', r.sequencename, pg_temp.to_spanish(r.sequencename));
  END LOOP;
END $$;

ALTER TABLE admins RENAME COLUMN password_hash TO contrasenia;
ALTER TABLE admins RENAME COLUMN last_name TO apellido;
ALTER TABLE admins RENAME COLUMN first_name TO nombre;
ALTER TABLE admins RENAME TO administradores;

ALTER TABLE services RENAME COLUMN price TO precio;
ALTER TABLE services RENAME COLUMN name TO tipo;
ALTER TABLE services RENAME TO servicios;

ALTER TABLE customers RENAME COLUMN last_name TO apellido;
ALTER TABLE customers RENAME COLUMN first_name TO nombre;
ALTER TABLE customers RENAME COLUMN phone TO telefono;
ALTER TABLE customers RENAME TO clientes;

ALTER TABLE appointments RENAME COLUMN service_id TO servicio_id;
ALTER TABLE appointments RENAME COLUMN customer_phone TO cliente_id;
ALTER TABLE appointments RENAME COLUMN starts_at TO fecha_hora;
ALTER TABLE appointments RENAME TO turnos;
