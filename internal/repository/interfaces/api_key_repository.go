package interfaces

import (
	"context"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

// APIKeyRepository defines API-key persistence operations.
type APIKeyRepository interface {
	Create(ctx context.Context, accountID string, input models.APIKeyCreateInput, keyPrefix, keyHash, keyLast4 string, expiresAt, lastUsedAt, revokedAt *time.Time) (models.APIKey, error)
	GetByID(ctx context.Context, id string) (models.APIKey, error)
	GetByKeyHash(ctx context.Context, keyHash string) (models.APIKey, error)
	ListByAccountID(ctx context.Context, accountID string, limit, offset int32) ([]models.APIKey, error)
	CountActiveByAccountID(ctx context.Context, accountID string) (int64, error)
	Touch(ctx context.Context, id string) (models.APIKey, error)
	Revoke(ctx context.Context, id string) (models.APIKey, error)
	Rotate(ctx context.Context, id string) (models.APIKey, error)
	DeleteExpired(ctx context.Context) error
}
