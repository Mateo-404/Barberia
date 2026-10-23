-- 000006_composite_fks.down.sql
ALTER TABLE turnos DROP CONSTRAINT IF EXISTS fk_turnos_servicio_tenant;
ALTER TABLE turnos DROP CONSTRAINT IF EXISTS fk_turnos_cliente_tenant;
ALTER TABLE turnos ADD CONSTRAINT turnos_cliente_id_fkey FOREIGN KEY (cliente_id) REFERENCES clientes (telefono);
ALTER TABLE turnos ADD CONSTRAINT turnos_servicio_id_fkey FOREIGN KEY (servicio_id) REFERENCES servicios (id);
ALTER TABLE servicios DROP CONSTRAINT IF EXISTS uq_servicios_tenant_id;
ALTER TABLE clientes DROP CONSTRAINT IF EXISTS uq_clientes_tenant_telefono;
