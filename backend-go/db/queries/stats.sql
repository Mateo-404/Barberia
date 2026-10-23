-- name: CountAppointmentsOnDate :one
SELECT count(*) AS total
FROM appointments
WHERE tenant_id = sqlc.arg('tenant_id')
  AND starts_at::date = sqlc.arg('day')::date;

-- name: SumRevenueBetween :one
-- Revenue of appointments with from <= starts_at < until.
SELECT COALESCE(sum(s.price), 0)::double precision AS total
FROM appointments a
JOIN services s ON a.service_id = s.id AND s.tenant_id = a.tenant_id
WHERE a.tenant_id = sqlc.arg('tenant_id')
  AND a.starts_at >= sqlc.arg('from_ts')
  AND a.starts_at < sqlc.arg('until_ts');

-- name: CountServicesBetween :many
-- Appointment count per service, most booked first.
SELECT s.id, s.name, count(*) AS total
FROM appointments a
JOIN services s ON a.service_id = s.id AND s.tenant_id = a.tenant_id
WHERE a.tenant_id = sqlc.arg('tenant_id')
  AND a.starts_at >= sqlc.arg('from_ts')
  AND a.starts_at < sqlc.arg('until_ts')
GROUP BY s.id, s.name
ORDER BY total DESC, s.id;

-- name: CountAppointmentsByHour :many
SELECT EXTRACT(HOUR FROM starts_at)::int AS hour, count(*) AS total
FROM appointments
WHERE tenant_id = sqlc.arg('tenant_id')
  AND starts_at >= sqlc.arg('from_ts')
GROUP BY 1
ORDER BY 1;
