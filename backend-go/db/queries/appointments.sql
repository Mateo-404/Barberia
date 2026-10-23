-- name: ListAppointments :many
-- Flat row with customer and service (LEFT JOIN: a PATCH may have cleared
-- an association). Used by GET /appointments and by the paged view, whose
-- nested shape is built in Go.
SELECT a.id, a.starts_at,
       c.phone, c.first_name, c.last_name, c.email,
       s.id AS service_id, s.name AS service_name, s.price AS service_price
FROM appointments a
LEFT JOIN customers c ON a.customer_phone = c.phone AND c.tenant_id = a.tenant_id
LEFT JOIN services s ON a.service_id = s.id AND s.tenant_id = a.tenant_id
WHERE a.tenant_id = sqlc.arg('tenant_id')
ORDER BY a.id;

-- name: PageAppointments :many
SELECT a.id, a.starts_at,
       c.phone, c.first_name, c.last_name, c.email,
       s.id AS service_id, s.name AS service_name, s.price AS service_price
FROM appointments a
LEFT JOIN customers c ON a.customer_phone = c.phone AND c.tenant_id = a.tenant_id
LEFT JOIN services s ON a.service_id = s.id AND s.tenant_id = a.tenant_id
WHERE a.tenant_id = sqlc.arg('tenant_id')
ORDER BY a.id
LIMIT sqlc.arg('row_limit') OFFSET sqlc.arg('row_offset');

-- name: CountAppointments :one
SELECT count(*) AS total
FROM appointments
WHERE tenant_id = sqlc.arg('tenant_id');

-- name: GetAppointment :one
SELECT a.id, a.starts_at,
       c.phone, c.first_name, c.last_name, c.email,
       s.id AS service_id, s.name AS service_name, s.price AS service_price
FROM appointments a
LEFT JOIN customers c ON a.customer_phone = c.phone AND c.tenant_id = a.tenant_id
LEFT JOIN services s ON a.service_id = s.id AND s.tenant_id = a.tenant_id
WHERE a.tenant_id = sqlc.arg('tenant_id') AND a.id = sqlc.arg('id');

-- name: CreateAppointment :one
INSERT INTO appointments (starts_at, customer_phone, service_id, tenant_id)
VALUES (sqlc.arg('starts_at'), sqlc.arg('customer_phone'), sqlc.arg('service_id'), sqlc.arg('tenant_id'))
RETURNING id, starts_at, customer_phone, service_id;

-- name: UpdateAppointment :one
UPDATE appointments
SET starts_at = sqlc.arg('starts_at'),
    customer_phone = sqlc.arg('customer_phone'),
    service_id = sqlc.arg('service_id')
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id')
RETURNING id, starts_at, customer_phone, service_id;

-- name: DeleteAppointment :execrows
DELETE FROM appointments
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: ListBookedSlotsFrom :many
-- Booked start times from a given instant, ascending. Formatting (no
-- seconds) happens in Go.
SELECT starts_at
FROM appointments
WHERE tenant_id = sqlc.arg('tenant_id') AND starts_at >= sqlc.arg('from_ts')
ORDER BY starts_at;

-- name: AppointmentExistsAt :one
SELECT EXISTS (
    SELECT 1 FROM appointments
    WHERE tenant_id = sqlc.arg('tenant_id') AND starts_at = sqlc.arg('starts_at')
) AS exists_at;
