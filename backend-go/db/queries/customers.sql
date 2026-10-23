-- name: ListCustomers :many
SELECT phone, first_name, last_name, email
FROM customers
WHERE tenant_id = sqlc.arg('tenant_id')
ORDER BY phone;

-- name: PageCustomers :many
SELECT phone, first_name, last_name, email
FROM customers
WHERE tenant_id = sqlc.arg('tenant_id')
ORDER BY phone
LIMIT sqlc.arg('row_limit') OFFSET sqlc.arg('row_offset');

-- name: CountCustomers :one
SELECT count(*) AS total
FROM customers
WHERE tenant_id = sqlc.arg('tenant_id');

-- name: GetCustomer :one
SELECT phone, first_name, last_name, email
FROM customers
WHERE tenant_id = sqlc.arg('tenant_id') AND phone = sqlc.arg('phone');

-- name: CreateCustomer :one
INSERT INTO customers (phone, first_name, last_name, email, tenant_id)
VALUES (sqlc.arg('phone'), sqlc.arg('first_name'), sqlc.arg('last_name'), sqlc.narg('email'), sqlc.arg('tenant_id'))
RETURNING phone, first_name, last_name, email;

-- name: UpdateCustomer :one
UPDATE customers
SET first_name = sqlc.arg('first_name'),
    last_name = sqlc.arg('last_name'),
    email = sqlc.arg('email')
WHERE tenant_id = sqlc.arg('tenant_id') AND phone = sqlc.arg('phone')
RETURNING phone, first_name, last_name, email;

-- name: DeleteCustomer :execrows
DELETE FROM customers
WHERE tenant_id = sqlc.arg('tenant_id') AND phone = sqlc.arg('phone');
