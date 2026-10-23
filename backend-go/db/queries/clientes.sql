-- name: ListClientes :many
SELECT telefono, nombre, apellido, email
FROM clientes
WHERE tenant_id = sqlc.arg('tenant_id')
ORDER BY telefono;

-- name: PageClientes :many
SELECT telefono, nombre, apellido, email
FROM clientes
WHERE tenant_id = sqlc.arg('tenant_id')
ORDER BY telefono
LIMIT sqlc.arg('limite') OFFSET sqlc.arg('offset');

-- name: CountClientes :one
SELECT count(*) AS total
FROM clientes
WHERE tenant_id = sqlc.arg('tenant_id');

-- name: GetCliente :one
SELECT telefono, nombre, apellido, email
FROM clientes
WHERE tenant_id = sqlc.arg('tenant_id') AND telefono = sqlc.arg('telefono');

-- name: CreateCliente :one
INSERT INTO clientes (telefono, nombre, apellido, email, tenant_id)
VALUES (sqlc.arg('telefono'), sqlc.arg('nombre'), sqlc.arg('apellido'), sqlc.narg('email'), sqlc.arg('tenant_id'))
RETURNING telefono, nombre, apellido, email;

-- name: UpdateCliente :one
UPDATE clientes
SET nombre = sqlc.arg('nombre'),
    apellido = sqlc.arg('apellido'),
    email = sqlc.arg('email')
WHERE tenant_id = sqlc.arg('tenant_id') AND telefono = sqlc.arg('telefono')
RETURNING telefono, nombre, apellido, email;

-- name: DeleteCliente :execrows
DELETE FROM clientes
WHERE tenant_id = sqlc.arg('tenant_id') AND telefono = sqlc.arg('telefono');
