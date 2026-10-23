-- name: ListAdministradores :many
SELECT id, nombre, apellido, email, contrasenia, role
FROM administradores
WHERE tenant_id = sqlc.arg('tenant_id')
ORDER BY id;

-- name: PageAdministradores :many
SELECT id, nombre, apellido, email, contrasenia, role
FROM administradores
WHERE tenant_id = sqlc.arg('tenant_id')
ORDER BY id
LIMIT sqlc.arg('limite') OFFSET sqlc.arg('offset');

-- name: CountAdministradores :one
SELECT count(*) AS total
FROM administradores
WHERE tenant_id = sqlc.arg('tenant_id');

-- name: GetAdministrador :one
SELECT id, nombre, apellido, email, contrasenia, role
FROM administradores
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: GetAdministradorByEmail :one
-- java: findByEmail (coincidencia exacta, case-sensitive).
SELECT id, nombre, apellido, email, contrasenia, role
FROM administradores
WHERE tenant_id = sqlc.arg('tenant_id') AND email = sqlc.arg('email');

-- name: CreateAdministrador :one
INSERT INTO administradores (nombre, apellido, email, contrasenia, tenant_id, role)
VALUES (sqlc.arg('nombre'), sqlc.arg('apellido'), sqlc.arg('email'), sqlc.arg('contrasenia'), sqlc.arg('tenant_id'), sqlc.arg('role'))
RETURNING id, nombre, apellido, email, contrasenia, role;

-- name: UpdateAdministrador :one
UPDATE administradores
SET nombre = sqlc.arg('nombre'),
    apellido = sqlc.arg('apellido'),
    email = sqlc.arg('email'),
    contrasenia = sqlc.arg('contrasenia'),
    role = sqlc.arg('role')
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id')
RETURNING id, nombre, apellido, email, contrasenia, role;

-- name: DeleteAdministrador :execrows
DELETE FROM administradores
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');
