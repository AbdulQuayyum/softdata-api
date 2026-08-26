-- Sessions queries

-- name: CreateSession :one
INSERT INTO sessions (
    account_id,
    refresh_token_hash,
    access_token_jti,
    user_agent,
    ip_address,
    expires_at,
    revoked_at,
    last_used_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetSessionByID :one
SELECT *
FROM sessions
WHERE id = $1;

-- name: GetSessionByRefreshTokenHash :one
SELECT *
FROM sessions
WHERE refresh_token_hash = $1;

-- name: ListSessionsByAccountID :many
SELECT *
FROM sessions
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListSessionsByAccountIDAfter :many
SELECT *
FROM sessions
WHERE account_id = $1
  AND (
    created_at < $3
    OR (created_at = $3 AND id < $4)
  )
ORDER BY created_at DESC, id DESC
LIMIT $2;

-- name: TouchSession :one
UPDATE sessions
SET
    last_used_at = now(),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: RevokeSessionByID :one
UPDATE sessions
SET
    revoked_at = now(),
    updated_at = now()
WHERE id = $1
  AND revoked_at IS NULL
RETURNING *;

-- name: RevokeSessionByRefreshTokenHash :one
UPDATE sessions
SET
    revoked_at = now(),
    updated_at = now()
WHERE refresh_token_hash = $1
  AND revoked_at IS NULL
RETURNING *;

-- name: RotateSessionTokens :one
UPDATE sessions
SET
    refresh_token_hash = sqlc.arg(new_refresh_token_hash),
    access_token_jti = sqlc.arg(new_access_token_jti),
    last_used_at = now(),
    updated_at = now()
WHERE id = sqlc.arg(session_id)
  AND refresh_token_hash = sqlc.arg(current_refresh_token_hash)
  AND revoked_at IS NULL
  AND expires_at > now()
RETURNING *;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < now()
   OR revoked_at IS NOT NULL;

-- name: DeleteSessionsByAccountID :exec
DELETE FROM sessions
WHERE account_id = $1;
