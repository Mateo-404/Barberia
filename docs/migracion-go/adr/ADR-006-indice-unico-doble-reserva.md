# ADR-006 — Índice único de doble reserva (turnos)

- **Estado:** Aceptado
- **Fecha:** 2026-10-15
- **Contexto:** El backend Java no tenía ninguna protección contra doble reserva:
  dos clientes podían reservar el mismo `fechaHora` en paralelo y ambos INSERTs
  prosperaban. Además, el servicio bloqueaba solo el rango del día consultado
  (`FOR UPDATE` de turnos de hoy), no la fila concreta.

## Decisión

Restricción a nivel de base de datos:

```sql
CREATE UNIQUE INDEX uq_turno_tenant_fecha ON turnos (tenant_id, fecha_hora);
```

En `POST /turnos`, la violación de esta restricción (`SQLSTATE 23505`) se
convierte en `400` con tipo `https://barber.local/errors/invalid-argument` y
detalle `Ya existe un turno en esa fecha y hora.`.

La creación ocurre dentro de una transacción con aislamiento
`READ COMMITTED` (suficiente: el índice único es el mecanismo de
serialización, la detección es instantánea y no hay window de TOCTOU).

## Consecuencias

- **Doble reserva prevenida de forma correcta y concurrencia-segura** (probado
  con 8 goroutines en paralelo: exactamente 1×201, 7×400).
- **Desviación funcional consciente:** el backend Java **permitía** crear un
  turno duplicado (y también en el pasado). Go rechaza ambos con `400`. Se
  considera mejora, no regresión (ver `02-matriz-compatibilidad.md`).
- Al **adoptar** un esquema Java existente, el backfill ejecuta esta restricción
  y **falla con datos sucios** (duplicados reales en producción). Es
  deliberado: el operador debe resolver los duplicados antes de migrar; el
  error identifica la fila exacta (`Key (tenant_id, fecha_hora)=... already exists`).
- El índice es global por tenant, por lo que el aislamiento multi-tenant se
  respeta: dos tenants pueden tener el mismo `fechaHora`.
