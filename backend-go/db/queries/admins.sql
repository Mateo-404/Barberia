-- name: ListAdmins :many
SELECT id, first_name, last_name, email, password_hash, role
FROM admins
WHERE tenant_id = sqlc.arg('tenant_id')
ORDER BY id;

-- name: PageAdmins :many
SELECT id, first_name, last_name, email, password_hash, role
FROM admins
WHERE tenant_id = sqlc.arg('tenant_id')
ORDER BY id
LIMIT sqlc.arg('row_limit') OFFSET sqlc.arg('row_offset');

-- name: CountAdmins :one
SELECT count(*) AS total
FROM admins
WHERE tenant_id = sqlc.arg('tenant_id');

-- name: GetAdmin :one
SELECT id, first_name, last_name, email, password_hash, role
FROM admins
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: GetAdminByEmail :one
-- Exact, case-sensitive match (same as the original Java findByEmail).
SELECT id, first_name, last_name, email, password_hash, role
FROM admins
WHERE tenant_id = sqlc.arg('tenant_id') AND email = sqlc.arg('email');

-- name: CreateAdmin :one
INSERT INTO admins (first_name, last_name, email, password_hash, tenant_id, role)
VALUES (sqlc.arg('first_name'), sqlc.arg('last_name'), sqlc.arg('email'), sqlc.arg('password_hash'), sqlc.arg('tenant_id'), sqlc.arg('role'))
RETURNING id, first_name, last_name, email, password_hash, role;

-- name: UpdateAdmin :one
UPDATE admins
SET first_name = sqlc.arg('first_name'),
    last_name = sqlc.arg('last_name'),
    email = sqlc.arg('email'),
    password_hash = sqlc.arg('password_hash'),
    role = sqlc.arg('role')
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id')
RETURNING id, first_name, last_name, email, password_hash, role;

-- name: DeleteAdmin :execrows
DELETE FROM admins
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');
