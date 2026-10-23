-- 000006_composite_fks.up.sql
-- FK compuestas (tenant_id, ...) para que la integridad cross-tenant se
-- verifique en la base, no solo en la aplicación (B-04).
-- Requiere que los padres tengan UNIQUE(tenant_id, pk).

-- clientes: UNIQUE(tenant_id, telefono)
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='uq_clientes_tenant_telefono') THEN
    ALTER TABLE clientes ADD CONSTRAINT uq_clientes_tenant_telefono UNIQUE (tenant_id, telefono);
  END IF;
END $$;

-- servicios: UNIQUE(tenant_id, id)
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='uq_servicios_tenant_id') THEN
    ALTER TABLE servicios ADD CONSTRAINT uq_servicios_tenant_id UNIQUE (tenant_id, id);
  END IF;
END $$;

-- turnos: drop FK simples y crear compuestas
ALTER TABLE turnos DROP CONSTRAINT IF EXISTS turnos_cliente_id_fkey;
ALTER TABLE turnos DROP CONSTRAINT IF EXISTS turnos_servicio_id_fkey;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='fk_turnos_cliente_tenant') THEN
    ALTER TABLE turnos ADD CONSTRAINT fk_turnos_cliente_tenant
      FOREIGN KEY (tenant_id, cliente_id) REFERENCES clientes (tenant_id, telefono) ON DELETE RESTRICT;
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='fk_turnos_servicio_tenant') THEN
    ALTER TABLE turnos ADD CONSTRAINT fk_turnos_servicio_tenant
      FOREIGN KEY (tenant_id, servicio_id) REFERENCES servicios (tenant_id, id) ON DELETE RESTRICT;
  END IF;
END $$;
