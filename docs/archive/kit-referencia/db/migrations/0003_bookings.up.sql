-- Reservas con integridad garantizada por la base de datos.
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE services (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name         text NOT NULL,
  duration_min int  NOT NULL CHECK (duration_min BETWEEN 5 AND 480),
  price_cents  bigint NOT NULL CHECK (price_cents >= 0),
  active       boolean NOT NULL DEFAULT true,
  UNIQUE (tenant_id, id)
);

CREATE TABLE barbers (
  id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name      text NOT NULL,
  active    boolean NOT NULL DEFAULT true,
  UNIQUE (tenant_id, id)
);

CREATE TABLE bookings (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id      uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  barber_id      uuid NOT NULL,
  service_id     uuid NOT NULL,
  during         tstzrange NOT NULL,
  status         text NOT NULL DEFAULT 'confirmed'
                   CHECK (status IN ('confirmed','cancelled','completed','no_show')),
  customer_name  text NOT NULL,
  customer_phone text NOT NULL,
  created_at     timestamptz NOT NULL DEFAULT now(),
  -- FKs compuestas: impiden referenciar barberos/servicios de OTRO tenant.
  FOREIGN KEY (tenant_id, barber_id)  REFERENCES barbers  (tenant_id, id),
  FOREIGN KEY (tenant_id, service_id) REFERENCES services (tenant_id, id),
  -- Rango semiabierto [inicio, fin), acotado y de duración razonable.
  CONSTRAINT bookings_range_ok CHECK (
    NOT isempty(during) AND lower_inc(during) AND NOT upper_inc(during)
    AND lower(during) IS NOT NULL AND upper(during) IS NOT NULL
    AND upper(during) - lower(during) <= interval '12 hours'),
  -- Un barbero no puede tener dos turnos no cancelados que se solapen (barber_id es único entre tenants).
  CONSTRAINT bookings_no_overlap EXCLUDE USING gist (barber_id WITH =, during WITH &&)
    WHERE (status <> 'cancelled')
);
CREATE INDEX bookings_tenant_barber_idx ON bookings (tenant_id, barber_id);

SELECT enable_tenant_rls('services');
SELECT enable_tenant_rls('barbers');
SELECT enable_tenant_rls('bookings');

GRANT SELECT, INSERT, UPDATE, DELETE ON services, barbers, bookings TO app_user;
