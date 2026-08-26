package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const rotateSessionTokens = `-- name: RotateSessionTokens :one
UPDATE sessions
SET
    refresh_token_hash = $3,
    access_token_jti = $4,
    last_used_at = now(),
    updated_at = now()
WHERE id = $1
  AND refresh_token_hash = $2
  AND revoked_at IS NULL
  AND expires_at > now()
RETURNING id, account_id, refresh_token_hash, access_token_jti, user_agent, ip_address, expires_at, revoked_at, last_used_at, created_at, updated_at
`

type RotateSessionTokensParams struct {
	SessionID               pgtype.UUID `json:"session_id"`
	CurrentRefreshTokenHash string      `json:"current_refresh_token_hash"`
	NewRefreshTokenHash     string      `json:"new_refresh_token_hash"`
	NewAccessTokenJti       pgtype.UUID `json:"new_access_token_jti"`
}

func (q *Queries) RotateSessionTokens(ctx context.Context, arg RotateSessionTokensParams) (Session, error) {
	row := q.db.QueryRow(ctx, rotateSessionTokens,
		arg.SessionID,
		arg.CurrentRefreshTokenHash,
		arg.NewRefreshTokenHash,
		arg.NewAccessTokenJti,
	)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.AccountID,
		&i.RefreshTokenHash,
		&i.AccessTokenJti,
		&i.UserAgent,
		&i.IpAddress,
		&i.ExpiresAt,
		&i.RevokedAt,
		&i.LastUsedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
