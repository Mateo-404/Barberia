-- name: ExistsTurnoEnFecha :one
SELECT EXISTS (
    SELECT 1 FROM turnos
    WHERE tenant_id = sqlc.arg('tenant_id') AND fecha_hora = sqlc.arg('fecha_hora')
) AS existe;
