-- 000002_tenancy.up.sql
-- Multi-tenantización (ADR-002) + invariantes de unicidad.
--
-- Ejecutable tanto sobre una base fresca (creada por 000001) como sobre una
-- base existente creada por el backend Java (Hibernate ddl-auto=update):
-- todas las operaciones son idempotentes (IF NOT EXISTS / guards en DO blocks).
--
-- Si la base existente contiene turnos con la misma fecha_hora, la creación
-- del índice único falla A PROPÓSITO (sin pérdida de datos): resolver los
-- duplicados manualmente antes de reintentar (ver docs/migracion-go/03).
--
-- NOTA de rollback al backend Java: las columnas tenant_id quedan con DEFAULT
-- default_tenant(), de modo que los INSERTs de Hibernate (que no conocen la
-- columna) siguen funcionando y escriben en el tenant por defecto. Una base
-- multi-tenant volcada al Java mostrará datos de todos los tenants mezclados:
-- solo aceptable en la ventana de rollback de una instalación single-tenant.

-- Función estable para DEFAULT de tenant_id (permite INSERTs ajenos al Go).
CREATE OR REPLACE FUNCTION default_tenant() RETURNS BIGINT
    LANGUAGE plpgsql STABLE
AS
$$
DECLARE
    tid BIGINT;
BEGIN
    SELECT id INTO tid FROM tenants WHERE slug = 'default';
    RETURN tid;
END
$$;

CREATE TABLE IF NOT EXISTS tenants (
    id         BIGSERIAL PRIMARY KEY,
    slug       TEXT NOT NULL UNIQUE,
    name       TEXT NOT NULL,
    host       TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO tenants (slug, name)
VALUES ('default', 'Barbería')
ON CONFLICT (slug) DO NOTHING;

-- Columnas de ámbito (agregadas solo si no existen). DEFAULT cubre el caso
-- de INSERTs que no provienen del backend Go (rollback al Java, psql, etc.).
ALTER TABLE clientes      ADD COLUMN IF NOT EXISTS tenant_id BIGINT DEFAULT default_tenant();
ALTER TABLE servicios     ADD COLUMN IF NOT EXISTS tenant_id BIGINT DEFAULT default_tenant();
ALTER TABLE administradores ADD COLUMN IF NOT EXISTS tenant_id BIGINT DEFAULT default_tenant();
ALTER TABLE turnos        ADD COLUMN IF NOT EXISTS tenant_id BIGINT DEFAULT default_tenant();

-- Backfill al tenant por defecto (no-op si ya está completo).
DO
$$
DECLARE
    tid BIGINT;
BEGIN
    SELECT id INTO tid FROM tenants WHERE slug = 'default';
    UPDATE clientes        SET tenant_id = tid WHERE tenant_id IS NULL;
    UPDATE servicios       SET tenant_id = tid WHERE tenant_id IS NULL;
    UPDATE administradores SET tenant_id = tid WHERE tenant_id IS NULL;
    UPDATE turnos          SET tenant_id = tid WHERE tenant_id IS NULL;
END
$$;

ALTER TABLE clientes        ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE servicios       ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE administradores ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE turnos          ALTER COLUMN tenant_id SET NOT NULL;

-- FKs de tenant (guardadas por nombre en pg_constraint).
DO
$$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_clientes_tenant') THEN
        ALTER TABLE clientes ADD CONSTRAINT fk_clientes_tenant
            FOREIGN KEY (tenant_id) REFERENCES tenants (id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_servicios_tenant') THEN
        ALTER TABLE servicios ADD CONSTRAINT fk_servicios_tenant
            FOREIGN KEY (tenant_id) REFERENCES tenants (id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_administradores_tenant') THEN
        ALTER TABLE administradores ADD CONSTRAINT fk_administradores_tenant
            FOREIGN KEY (tenant_id) REFERENCES tenants (id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_turnos_tenant') THEN
        ALTER TABLE turnos ADD CONSTRAINT fk_turnos_tenant
            FOREIGN KEY (tenant_id) REFERENCES tenants (id);
    END IF;
END
$$;

-- PK de clientes pasa a ser compuesta (tenant_id, telefono): dos negocios del
-- SaaS pueden atender al mismo teléfono. Se elimina primero cualquier FK que
-- apunte a la PK simple (creada por 000001 o por Hibernate) y se recrea como
-- FK compuesta desde turnos(tenant_id, cliente_id).
DO
$$
DECLARE
    r RECORD;
BEGIN
    FOR r IN SELECT conname
             FROM pg_constraint
             WHERE contype = 'f'
               AND conrelid = 'turnos'::regclass
               AND conkey = ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid = 'turnos'::regclass AND attname = 'cliente_id')]
    LOOP
        EXECUTE format('ALTER TABLE turnos DROP CONSTRAINT %I', r.conname);
    END LOOP;

    ALTER TABLE clientes DROP CONSTRAINT IF EXISTS clientes_pkey;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = 'clientes'::regclass AND contype = 'p') THEN
        ALTER TABLE clientes ADD PRIMARY KEY (tenant_id, telefono);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_turnos_cliente') THEN
        ALTER TABLE turnos ADD CONSTRAINT fk_turnos_cliente
            FOREIGN KEY (tenant_id, cliente_id) REFERENCES clientes (tenant_id, telefono);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_turnos_servicio') THEN
        ALTER TABLE turnos ADD CONSTRAINT fk_turnos_servicio
            FOREIGN KEY (servicio_id) REFERENCES servicios (id);
    END IF;
END
$$;

-- Secuencia de turnos: en bases frescas BIGSERIAL ya creó turnos_id_seq con
-- DEFAULT; en bases de Hibernate el id no tiene DEFAULT (los ids salían de
-- hibernate_sequence). En ambos casos la secuencia local se alinea con MAX(id)
-- para no colisionar con filas existentes.
DO
$$
BEGIN
    IF NOT EXISTS (SELECT 1
                   FROM pg_attrdef
                   WHERE adrelid = 'turnos'::regclass
                     AND adnum = (SELECT attnum FROM pg_attribute
                                  WHERE attrelid = 'turnos'::regclass AND attname = 'id')) THEN
        CREATE SEQUENCE IF NOT EXISTS turnos_id_seq;
        ALTER TABLE turnos ALTER COLUMN id SET DEFAULT nextval('turnos_id_seq');
        ALTER SEQUENCE turnos_id_seq OWNED BY turnos.id;
    END IF;

    IF EXISTS (SELECT 1 FROM turnos LIMIT 1) THEN
        PERFORM setval('turnos_id_seq', (SELECT MAX(id) FROM turnos));
    END IF;
END
$$;

-- Unicidad atómica (ADR-004/S3): un turno por fecha/hora dentro de cada
-- negocio, y un email de admin por negocio (case-insensitive).
CREATE UNIQUE INDEX IF NOT EXISTS uq_turno_tenant_fecha
    ON turnos (tenant_id, fecha_hora);

CREATE UNIQUE INDEX IF NOT EXISTS uq_admin_tenant_email
    ON administradores (tenant_id, lower(email));
