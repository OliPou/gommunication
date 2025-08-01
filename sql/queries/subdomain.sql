-- name: CreateAvailableSubdomain :one
INSERT INTO available_subdomains (name)
VALUES ($1)
RETURNING *;

-- name: GetAvailableSubdomainByName :one
SELECT available_subdomain_uuid, name, created_at, is_default FROM available_subdomains
WHERE name = $1;

-- name: CreateSubdomainOwnership :one
INSERT INTO subdomain_ownerships (subdomain_ownership_uuid, subdomain_id, api_key)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSubdomainOwnershipBySubdomain :one
SELECT subdomain_ownership_uuid, subdomain_id, api_key, created_at FROM subdomain_ownerships
WHERE subdomain_id = $1;

-- name: GetSubdomainOwnershipByApiKey :one
SELECT subdomain_ownership_uuid, subdomain_id, api_key, created_at FROM subdomain_ownerships
WHERE api_key = $1;

-- name: IsSubdomainAllowedForConsumer :one
SELECT EXISTS (
  SELECT 1
  FROM available_subdomains s
  LEFT JOIN subdomain_ownerships o ON s.available_subdomain_uuid = o.subdomain_id AND o.api_key = $2
  WHERE s.name = $1 AND (o.api_key IS NOT NULL OR s.is_default = TRUE)
) AS allowed;