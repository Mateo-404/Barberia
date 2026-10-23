-- name: CountTurnosDelDia :one
-- Réplica de countByFecha: CAST(fechaHora AS date) = CAST(:fecha AS date).
SELECT count(*) AS total
FROM turnos
WHERE tenant_id = sqlc.arg('tenant_id')
  AND fecha_hora::date = sqlc.arg('fecha')::date;

-- name: SumIngresosRango :one
-- Réplica de findIngresosByFecha: fecha_hora >= inicio AND < fin.
SELECT COALESCE(sum(s.precio), 0)::double precision AS total
FROM turnos t
JOIN servicios s ON t.servicio_id = s.id AND s.tenant_id = t.tenant_id
WHERE t.tenant_id = sqlc.arg('tenant_id')
  AND t.fecha_hora >= sqlc.arg('inicio')
  AND t.fecha_hora < sqlc.arg('fin');

-- name: CountServiciosRango :many
-- Réplica de countServiciosByFecha (agrupado): cantidad por servicio, desc.
SELECT s.id, s.tipo, count(*) AS cantidad
FROM turnos t
JOIN servicios s ON t.servicio_id = s.id AND s.tenant_id = t.tenant_id
WHERE t.tenant_id = sqlc.arg('tenant_id')
  AND t.fecha_hora >= sqlc.arg('desde')
  AND t.fecha_hora < sqlc.arg('hasta')
GROUP BY s.id, s.tipo
ORDER BY cantidad DESC, s.id;

-- name: CountTurnosPorHorario :many
-- Réplica de countTurnosByHorario: agrupado por hora, desde el 1/1.
SELECT EXTRACT(HOUR FROM fecha_hora)::int AS hora, count(*) AS cantidad
FROM turnos
WHERE tenant_id = sqlc.arg('tenant_id')
  AND fecha_hora >= sqlc.arg('desde')
GROUP BY 1
ORDER BY 1;
