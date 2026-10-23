# ADR-005 — `POST /turnos/all` implementado con las reglas de creación

- **Estado**: aceptado (2026-10-21)
- **Defecto verificado**: en Java, `POST /turnos/all` llama `GenericService.saveAll`
  que salta el find-or-create de cliente y la validación de `TurnoService.save`;
  con un cliente nuevo la inserción viola la FK `cliente_id` → **500**. Solo
  funciona (201) si el cliente ya existe en la BD.
- **Decisión**: el Go procesa el batch en una transacción aplicando exactamente
  las mismas reglas que `POST /turnos` (validación de hora, duplicados,
  find-or-create de cliente, servicio existente). Forma de respuesta idéntica
  (201 + `TurnoResponseDTO[]`; 400/404 en los mismos casos).
- **Justificación**: el endpoint no está roto por contrato sino por implementación;
  "corregir sin roturas silenciosas" aplica. Ningún frontend usa la ruta.
