-- Reversión de 000002. Devuelve el esquema al estado de 000001 (sin tenant).
-- Precaución: elimina la columna tenant_id con todos sus datos de ámbito.
-- Solo utilizar con una base de un único tenant (la habitual en la ventana
-- de rollback). Ver docs/migracion-go/03.

DROP INDEX IF EXISTS uq_admin_tenant_email;
DROP INDEX IF EXISTS uq_turno_tenant_fecha;

DO
$$
DECLARE
    r RECORD;
BEGIN
    -- FK compuesta de turnos hacia clientes(tenant_id, telefono).
    FOR r IN SELECT conname
             FROM pg_constraint
             WHERE contype = 'f'
               AND conrelid = 'turnos'::regclass
               AND conname IN ('fk_turnos_cliente')
    LOOP
        EXECUTE format('ALTER TABLE turnos DROP CONSTRAINT %I', r.conname);
    END LOOP;

    ALTER TABLE clientes DROP CONSTRAINT IF EXISTS fk_clientes_tenant;
    ALTER TABLE servicios DROP CONSTRAINT IF EXISTS fk_servicios_tenant;
    ALTER TABLE administradores DROP CONSTRAINT IF EXISTS fk_administradores_tenant;
    ALTER TABLE turnos DROP CONSTRAINT IF EXISTS fk_turnos_tenant;
    ALTER TABLE turnos DROP CONSTRAINT IF EXISTS fk_turnos_servicio;

    -- Restaurar PK simple de clientes y FK simple desde turnos.
    ALTER TABLE clientes DROP CONSTRAINT IF EXISTS clientes_pkey;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = 'clientes'::regclass AND contype = 'p') THEN
        ALTER TABLE clientes ADD PRIMARY KEY (telefono);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'turnos_cliente_id_fkey') THEN
        ALTER TABLE turnos ADD CONSTRAINT turnos_cliente_id_fkey
            FOREIGN KEY (cliente_id) REFERENCES clientes (telefono);
    END IF;
END
$$;

ALTER TABLE clientes        DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE servicios       DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE administradores DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE turnos          DROP COLUMN IF EXISTS tenant_id;

DROP TABLE IF EXISTS tenants;
DROP FUNCTION IF EXISTS default_tenant();
