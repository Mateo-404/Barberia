-- name: GetTenantByHost :one
SELECT id, slug, name, host
FROM tenants
WHERE host = lower(sqlc.arg('host'));

-- name: GetTenantBySlug :one
SELECT id, slug, name, host
FROM tenants
WHERE slug = sqlc.arg('slug');

-- name: GetTenantByDomain :one
SELECT t.id, t.slug, t.name, t.host
FROM tenant_domains td
JOIN tenants t ON t.id = td.tenant_id
WHERE td.domain = lower(sqlc.arg('domain')) AND td.verified_at IS NOT NULL;
