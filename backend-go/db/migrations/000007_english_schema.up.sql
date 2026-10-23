-- 000007_english_schema.up.sql
-- Renames the business schema to English (see docs/glossary.md).
-- Pure renames: no data is copied or dropped. RLS policies, foreign keys and
-- column defaults follow the renamed objects because PostgreSQL tracks them
-- by OID.

ALTER TABLE turnos RENAME TO appointments;
ALTER TABLE appointments RENAME COLUMN fecha_hora TO starts_at;
ALTER TABLE appointments RENAME COLUMN cliente_id TO customer_phone;
ALTER TABLE appointments RENAME COLUMN servicio_id TO service_id;

ALTER TABLE clientes RENAME TO customers;
ALTER TABLE customers RENAME COLUMN telefono TO phone;
ALTER TABLE customers RENAME COLUMN nombre TO first_name;
ALTER TABLE customers RENAME COLUMN apellido TO last_name;

ALTER TABLE servicios RENAME TO services;
ALTER TABLE services RENAME COLUMN tipo TO name;
ALTER TABLE services RENAME COLUMN precio TO price;

ALTER TABLE administradores RENAME TO admins;
ALTER TABLE admins RENAME COLUMN nombre TO first_name;
ALTER TABLE admins RENAME COLUMN apellido TO last_name;
ALTER TABLE admins RENAME COLUMN contrasenia TO password_hash;

-- Constraint, index and sequence names. Databases adopted from Hibernate may
-- carry different names, so every object on these tables is renamed by
-- substring instead of by an exact list.
CREATE OR REPLACE FUNCTION pg_temp.to_english(name TEXT) RETURNS TEXT
LANGUAGE sql IMMUTABLE AS $$
  SELECT replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(
    name,
    'administradores', 'admins'),
    'clientes', 'customers'),
    'cliente_id', 'customer_phone'),
    'cliente', 'customer'),
    'servicios', 'services'),
    'servicio', 'service'),
    'turnos', 'appointments'),
    'turno', 'appointment'),
    'fecha_hora', 'starts_at'),
    'fechahora', 'starts_at'),
    'telefono', 'phone')
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
      AND c.conname <> pg_temp.to_english(c.conname)
  LOOP
    EXECUTE format('ALTER TABLE %s RENAME CONSTRAINT %I TO %I',
                   r.tbl, r.conname, pg_temp.to_english(r.conname));
  END LOOP;

  FOR r IN
    SELECT indexname
    FROM pg_indexes
    WHERE schemaname = 'public'
      AND tablename IN ('appointments', 'customers', 'services', 'admins')
      AND indexname <> pg_temp.to_english(indexname)
  LOOP
    EXECUTE format('ALTER INDEX %I RENAME TO %I', r.indexname, pg_temp.to_english(r.indexname));
  END LOOP;

  FOR r IN
    SELECT sequencename
    FROM pg_sequences
    WHERE schemaname = 'public'
      AND sequencename <> pg_temp.to_english(sequencename)
  LOOP
    EXECUTE format('ALTER SEQUENCE %I RENAME TO %I', r.sequencename, pg_temp.to_english(r.sequencename));
  END LOOP;
END $$;

-- Hibernate-era index names that the substring map cannot express.
ALTER INDEX IF EXISTS idx_appointment_starts_at RENAME TO idx_appointments_starts_at;
ALTER INDEX IF EXISTS uq_appointment_tenant_fecha RENAME TO uq_appointments_tenant_starts_at;
