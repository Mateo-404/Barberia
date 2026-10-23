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

-- name: GetAdminByGoogleSub :one
SELECT id, first_name, last_name, email, password_hash, role,
       google_sub, auth_provider, totp_secret, totp_enabled, totp_verified_at
FROM admins
WHERE tenant_id = sqlc.arg('tenant_id') AND google_sub = sqlc.arg('google_sub');

-- name: CreateAdminGoogle :one
INSERT INTO admins (first_name, last_name, email, password_hash, tenant_id, role, google_sub, auth_provider)
VALUES (sqlc.arg('first_name'), sqlc.arg('last_name'), sqlc.arg('email'), sqlc.arg('password_hash'),
        sqlc.arg('tenant_id'), sqlc.arg('role'), sqlc.arg('google_sub'), sqlc.arg('auth_provider'))
RETURNING id, first_name, last_name, email, password_hash, role, google_sub, auth_provider, totp_secret, totp_enabled, totp_verified_at;

-- name: LinkAdminGoogle :one
UPDATE admins
SET google_sub = sqlc.arg('google_sub'),
    auth_provider = CASE WHEN auth_provider = 'password' THEN 'both' ELSE auth_provider END
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id')
RETURNING id, first_name, last_name, email, password_hash, role, google_sub, auth_provider, totp_secret, totp_enabled, totp_verified_at;

-- name: GetAdminTOTP :one
SELECT totp_secret, totp_enabled, totp_verified_at
FROM admins
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: SetAdminTOTPSecret :exec
UPDATE admins
SET totp_secret = sqlc.arg('totp_secret'),
    totp_enabled = FALSE,
    totp_verified_at = NULL
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: EnableAdminTOTP :exec
UPDATE admins
SET totp_enabled = TRUE,
    totp_verified_at = now()
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: DisableAdminTOTP :exec
UPDATE admins
SET totp_secret = NULL,
    totp_enabled = FALSE,
    totp_verified_at = NULL,
    auth_provider = CASE WHEN google_sub IS NOT NULL THEN 'google' ELSE 'password' END
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: DeleteAdmin :execrows
DELETE FROM admins
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');
