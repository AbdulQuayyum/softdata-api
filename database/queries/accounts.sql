-- Accounts queries

-- name: CreateAccount :one
INSERT INTO accounts (
    username,
    email,
    password_hash,
    status,
    email_verified_at,
    last_login_at
) VALUES (
    $1,
    $2,
    $3,
    COALESCE($4, 'active'),
    $5,
    $6
)
RETURNING *;

-- name: GetAccountByID :one
SELECT *
FROM accounts
WHERE id = $1
  AND deleted_at IS NULL;

-- name: GetAccountByUsername :one
SELECT *
FROM accounts
WHERE lower(username) = lower($1)
  AND deleted_at IS NULL;

-- name: GetAccountByEmail :one
SELECT *
FROM accounts
WHERE lower(email) = lower($1)
  AND email IS NOT NULL
  AND deleted_at IS NULL;

-- name: ListAccounts :many
SELECT *
FROM accounts
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAccountsAfter :many
SELECT *
FROM accounts
WHERE deleted_at IS NULL
  AND (
    created_at < $2
    OR (created_at = $2 AND id < $3)
  )
ORDER BY created_at DESC, id DESC
LIMIT $1;

-- name: UpdateAccountProfile :one
UPDATE accounts
SET
    username = COALESCE($2, username),
    email = COALESCE($3, email),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateAccountPassword :one
UPDATE accounts
SET
    password_hash = $2,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: MarkAccountLogin :one
UPDATE accounts
SET
    last_login_at = now(),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: VerifyAccountEmail :one
UPDATE accounts
SET
    email_verified_at = COALESCE(email_verified_at, now()),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteAccount :one
UPDATE accounts
SET
    status = 'deleted',
    deleted_at = now(),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;
