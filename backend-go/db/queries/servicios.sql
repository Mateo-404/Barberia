-- name: ListServicios :many
SELECT id, tipo, precio
FROM servicios
WHERE tenant_id = sqlc.arg('tenant_id')
ORDER BY id;

-- name: PageServicios :many
SELECT id, tipo, precio
FROM servicios
WHERE tenant_id = sqlc.arg('tenant_id')
ORDER BY id
LIMIT sqlc.arg('limite') OFFSET sqlc.arg('offset');

-- name: CountServicios :one
SELECT count(*) AS total
FROM servicios
WHERE tenant_id = sqlc.arg('tenant_id');

-- name: GetServicio :one
SELECT id, tipo, precio
FROM servicios
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: CreateServicio :one
INSERT INTO servicios (tipo, precio, tenant_id)
VALUES (sqlc.arg('tipo'), sqlc.arg('precio'), sqlc.arg('tenant_id'))
RETURNING id, tipo, precio;

-- name: UpdateServicio :one
UPDATE servicios
SET tipo = sqlc.arg('tipo'),
    precio = sqlc.arg('precio')
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id')
RETURNING id, tipo, precio;

-- name: DeleteServicio :execrows
DELETE FROM servicios
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');
