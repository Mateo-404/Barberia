-- name: ListTurnos :many
-- Fila plana con cliente y servicio (LEFT JOIN: puede haber asociaciones
-- anuladas vía PATCH). Sirve para GET /turnos (DTO plano) y para la entidad
-- anidada de la paginación (se compone en Go).
SELECT t.id, t.fecha_hora,
       c.telefono, c.nombre, c.apellido, c.email,
       s.id AS servicio_id, s.tipo AS servicio_tipo, s.precio AS servicio_precio
FROM turnos t
LEFT JOIN clientes c ON t.cliente_id = c.telefono AND c.tenant_id = t.tenant_id
LEFT JOIN servicios s ON t.servicio_id = s.id AND s.tenant_id = t.tenant_id
WHERE t.tenant_id = sqlc.arg('tenant_id')
ORDER BY t.id;

-- name: PageTurnos :many
SELECT t.id, t.fecha_hora,
       c.telefono, c.nombre, c.apellido, c.email,
       s.id AS servicio_id, s.tipo AS servicio_tipo, s.precio AS servicio_precio
FROM turnos t
LEFT JOIN clientes c ON t.cliente_id = c.telefono AND c.tenant_id = t.tenant_id
LEFT JOIN servicios s ON t.servicio_id = s.id AND s.tenant_id = t.tenant_id
WHERE t.tenant_id = sqlc.arg('tenant_id')
ORDER BY t.id
LIMIT sqlc.arg('limite') OFFSET sqlc.arg('offset');

-- name: CountTurnos :one
SELECT count(*) AS total
FROM turnos
WHERE tenant_id = sqlc.arg('tenant_id');

-- name: GetTurno :one
SELECT t.id, t.fecha_hora,
       c.telefono, c.nombre, c.apellido, c.email,
       s.id AS servicio_id, s.tipo AS servicio_tipo, s.precio AS servicio_precio
FROM turnos t
LEFT JOIN clientes c ON t.cliente_id = c.telefono AND c.tenant_id = t.tenant_id
LEFT JOIN servicios s ON t.servicio_id = s.id AND s.tenant_id = t.tenant_id
WHERE t.tenant_id = sqlc.arg('tenant_id') AND t.id = sqlc.arg('id');

-- name: CreateTurno :one
INSERT INTO turnos (fecha_hora, cliente_id, servicio_id, tenant_id)
VALUES (sqlc.arg('fecha_hora'), sqlc.arg('cliente_id'), sqlc.arg('servicio_id'), sqlc.arg('tenant_id'))
RETURNING id, fecha_hora, cliente_id, servicio_id;

-- name: UpdateTurno :one
UPDATE turnos
SET fecha_hora = sqlc.arg('fecha_hora'),
    cliente_id = sqlc.arg('cliente_id'),
    servicio_id = sqlc.arg('servicio_id')
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id')
RETURNING id, fecha_hora, cliente_id, servicio_id;

-- name: DeleteTurno :execrows
DELETE FROM turnos
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: FindDateTimesDesde :many
-- Réplica de findDateTimes(desde) de Java: solo fecha_hora >= ahora, orden
-- ascendente. La serialización sin segundos se hace en Go (LocalDateTime.toString).
SELECT fecha_hora
FROM turnos
WHERE tenant_id = sqlc.arg('tenant_id') AND fecha_hora >= sqlc.arg('desde')
ORDER BY fecha_hora;
