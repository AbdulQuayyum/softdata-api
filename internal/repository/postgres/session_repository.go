package postgres

import (
	"context"
	"fmt"

	sqlc "github.com/AbdulQuayyum/softdata-api/internal/database/sqlc"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var _ interfaces.SessionRepository = (*SessionRepository)(nil)

type SessionRepository struct {
	queries *sqlc.Queries
}

func NewSessionRepository(dbt sqlc.DBTX) *SessionRepository {
	return &SessionRepository{queries: sqlc.New(dbt)}
}

func (r *SessionRepository) Create(ctx context.Context, session models.Session) (models.Session, error) {
	accountID, err := uuidFromString(session.AccountID)
	if err != nil {
		return models.Session{}, fmt.Errorf("create session: %w", err)
	}

	accessTokenJTI, err := uuidFromStringPtr(session.AccessTokenJTI)
	if err != nil {
		return models.Session{}, fmt.Errorf("create session: %w", err)
	}

	ipAddress, err := netAddrPtrFromString(session.IPAddress)
	if err != nil {
		return models.Session{}, fmt.Errorf("create session: %w", err)
	}

	row, err := r.queries.CreateSession(ctx, sqlc.CreateSessionParams{
		AccountID:        accountID,
		RefreshTokenHash: session.RefreshTokenHash,
		AccessTokenJti:   accessTokenJTI,
		UserAgent:        textFromStringPtr(session.UserAgent),
		IpAddress:        ipAddress,
		ExpiresAt:        timestamptzFromTimePtr(&session.ExpiresAt),
		RevokedAt:        timestamptzFromTimePtr(session.RevokedAt),
		LastUsedAt:       timestamptzFromTimePtr(session.LastUsedAt),
	})
	if err != nil {
		return models.Session{}, translateError("create session", err)
	}
	return sessionFromRow(row), nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (models.Session, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.Session{}, fmt.Errorf("get session by id: %w", err)
	}

	row, err := r.queries.GetSessionByID(ctx, uid)
	if err != nil {
		return models.Session{}, translateError("get session by id", err)
	}
	return sessionFromRow(row), nil
}

func (r *SessionRepository) GetByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (models.Session, error) {
	row, err := r.queries.GetSessionByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		return models.Session{}, translateError("get session by refresh token hash", err)
	}
	return sessionFromRow(row), nil
}

func (r *SessionRepository) ListByAccountID(ctx context.Context, accountID string, limit, offset int32) ([]models.Session, error) {
	uid, err := uuidFromString(accountID)
	if err != nil {
		return nil, fmt.Errorf("list sessions by account id: %w", err)
	}

	rows, err := r.queries.ListSessionsByAccountID(ctx, sqlc.ListSessionsByAccountIDParams{
		AccountID: uid,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, translateError("list sessions by account id", err)
	}

	items := make([]models.Session, 0, len(rows))
	for _, row := range rows {
		items = append(items, sessionFromRow(row))
	}
	return items, nil
}

func (r *SessionRepository) Touch(ctx context.Context, id string) (models.Session, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.Session{}, fmt.Errorf("touch session: %w", err)
	}

	row, err := r.queries.TouchSession(ctx, uid)
	if err != nil {
		return models.Session{}, translateError("touch session", err)
	}
	return sessionFromRow(row), nil
}

func (r *SessionRepository) RotateSessionTokens(ctx context.Context, sessionID string, currentRefreshTokenHash string, newRefreshTokenHash string, newAccessTokenJTI string) (models.Session, error) {
	uid, err := uuidFromString(sessionID)
	if err != nil {
		return models.Session{}, fmt.Errorf("rotate session tokens: %w", err)
	}

	jti, err := uuidFromString(newAccessTokenJTI)
	if err != nil {
		return models.Session{}, fmt.Errorf("rotate session tokens: %w", err)
	}

	row, err := r.queries.RotateSessionTokens(ctx, sqlc.RotateSessionTokensParams{
		SessionID:               uid,
		CurrentRefreshTokenHash: currentRefreshTokenHash,
		NewRefreshTokenHash:     newRefreshTokenHash,
		NewAccessTokenJti:       jti,
	})
	if err != nil {
		return models.Session{}, translateError("rotate session tokens", err)
	}
	return sessionFromRow(row), nil
}

func (r *SessionRepository) RevokeByID(ctx context.Context, id string) (models.Session, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.Session{}, fmt.Errorf("revoke session by id: %w", err)
	}

	row, err := r.queries.RevokeSessionByID(ctx, uid)
	if err != nil {
		return models.Session{}, translateError("revoke session by id", err)
	}
	return sessionFromRow(row), nil
}

func (r *SessionRepository) RevokeByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (models.Session, error) {
	row, err := r.queries.RevokeSessionByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		return models.Session{}, translateError("revoke session by refresh token hash", err)
	}
	return sessionFromRow(row), nil
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	return translateError("delete expired sessions", r.queries.DeleteExpiredSessions(ctx))
}

func (r *SessionRepository) DeleteByAccountID(ctx context.Context, accountID string) error {
	uid, err := uuidFromString(accountID)
	if err != nil {
		return fmt.Errorf("delete sessions by account id: %w", err)
	}
	return translateError("delete sessions by account id", r.queries.DeleteSessionsByAccountID(ctx, uid))
}
