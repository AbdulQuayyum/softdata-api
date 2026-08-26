-- API keys queries

-- name: CreateAPIKey :one
INSERT INTO api_keys (
    account_id,
    name,
    key_prefix,
    key_hash,
    key_last4,
    status,
    last_used_at,
    expires_at,
    revoked_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    COALESCE($6, 'active'),
    $7,
    $8,
    $9
)
RETURNING *;

-- name: GetAPIKeyByID :one
SELECT *
FROM api_keys
WHERE id = $1;

-- name: GetAPIKeyByPrefix :one
SELECT *
FROM api_keys
WHERE key_prefix = $1;

-- name: GetAPIKeyByHash :one
SELECT *
FROM api_keys
WHERE key_hash = $1;

-- name: ListAPIKeysByAccountID :many
SELECT *
FROM api_keys
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAPIKeysByAccountIDAfter :many
SELECT *
FROM api_keys
WHERE account_id = $1
  AND (
    created_at < $3
    OR (created_at = $3 AND id < $4)
  )
ORDER BY created_at DESC, id DESC
LIMIT $2;

-- name: CountActiveAPIKeysByAccountID :one
SELECT COUNT(*)::bigint AS active_count
FROM api_keys
WHERE account_id = $1
  AND status = 'active'
  AND revoked_at IS NULL
  AND (expires_at IS NULL OR expires_at > now());

-- name: TouchAPIKey :one
UPDATE api_keys
SET
    last_used_at = now(),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: RevokeAPIKey :one
UPDATE api_keys
SET
    status = 'revoked',
    revoked_at = now(),
    updated_at = now()
WHERE id = $1
  AND status <> 'revoked'
RETURNING *;

-- name: RotateAPIKey :one
UPDATE api_keys
SET
    status = 'revoked',
    revoked_at = now(),
    updated_at = now()
WHERE id = $1
  AND status <> 'revoked'
RETURNING *;

-- name: DeleteExpiredAPIKeys :exec
DELETE FROM api_keys
WHERE expires_at IS NOT NULL
  AND expires_at < now();
