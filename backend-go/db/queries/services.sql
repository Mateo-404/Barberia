-- name: ListServices :many
SELECT id, name, price
FROM services
WHERE tenant_id = sqlc.arg('tenant_id')
ORDER BY id;

-- name: PageServices :many
SELECT id, name, price
FROM services
WHERE tenant_id = sqlc.arg('tenant_id')
ORDER BY id
LIMIT sqlc.arg('row_limit') OFFSET sqlc.arg('row_offset');

-- name: CountServices :one
SELECT count(*) AS total
FROM services
WHERE tenant_id = sqlc.arg('tenant_id');

-- name: GetService :one
SELECT id, name, price
FROM services
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: CreateService :one
INSERT INTO services (name, price, tenant_id)
VALUES (sqlc.arg('name'), sqlc.arg('price'), sqlc.arg('tenant_id'))
RETURNING id, name, price;

-- name: UpdateService :one
UPDATE services
SET name = sqlc.arg('name'),
    price = sqlc.arg('price')
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id')
RETURNING id, name, price;

-- name: DeleteService :execrows
DELETE FROM services
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');
