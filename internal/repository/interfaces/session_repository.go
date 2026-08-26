package interfaces

import (
	"context"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

// SessionRepository defines refresh-session persistence operations.
type SessionRepository interface {
	Create(ctx context.Context, session models.Session) (models.Session, error)
	GetByID(ctx context.Context, id string) (models.Session, error)
	GetByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (models.Session, error)
	ListByAccountID(ctx context.Context, accountID string, limit, offset int32) ([]models.Session, error)
	Touch(ctx context.Context, id string) (models.Session, error)
	RotateSessionTokens(ctx context.Context, sessionID string, currentRefreshTokenHash string, newRefreshTokenHash string, newAccessTokenJTI string) (models.Session, error)
	RevokeByID(ctx context.Context, id string) (models.Session, error)
	RevokeByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (models.Session, error)
	DeleteExpired(ctx context.Context) error
	DeleteByAccountID(ctx context.Context, accountID string) error
}
