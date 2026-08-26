package postgres

import (
	"context"
	"fmt"
	"time"

	sqlc "github.com/AbdulQuayyum/softdata-api/internal/database/sqlc"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var _ interfaces.APIKeyRepository = (*APIKeyRepository)(nil)

type APIKeyRepository struct {
	queries *sqlc.Queries
}

func NewAPIKeyRepository(dbt sqlc.DBTX) *APIKeyRepository {
	return &APIKeyRepository{queries: sqlc.New(dbt)}
}

func (r *APIKeyRepository) Create(ctx context.Context, accountID string, input models.APIKeyCreateInput, keyPrefix, keyHash, keyLast4 string, expiresAt, lastUsedAt, revokedAt *time.Time) (models.APIKey, error) {
	uid, err := uuidFromString(accountID)
	if err != nil {
		return models.APIKey{}, fmt.Errorf("create api key: %w", err)
	}

	row, err := r.queries.CreateAPIKey(ctx, sqlc.CreateAPIKeyParams{
		AccountID:  uid,
		Name:       input.Name,
		KeyPrefix:  keyPrefix,
		KeyHash:    keyHash,
		KeyLast4:   keyLast4,
		Column6:    nil,
		LastUsedAt: timestamptzFromTimePtr(lastUsedAt),
		ExpiresAt:  timestamptzFromTimePtr(expiresAt),
		RevokedAt:  timestamptzFromTimePtr(revokedAt),
	})
	if err != nil {
		return models.APIKey{}, translateError("create api key", err)
	}
	return apiKeyFromRow(row), nil
}

func (r *APIKeyRepository) GetByID(ctx context.Context, id string) (models.APIKey, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.APIKey{}, fmt.Errorf("get api key by id: %w", err)
	}

	row, err := r.queries.GetAPIKeyByID(ctx, uid)
	if err != nil {
		return models.APIKey{}, translateError("get api key by id", err)
	}
	return apiKeyFromRow(row), nil
}

func (r *APIKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (models.APIKey, error) {
	row, err := r.queries.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		return models.APIKey{}, translateError("get api key by hash", err)
	}
	return apiKeyFromRow(row), nil
}

func (r *APIKeyRepository) ListByAccountID(ctx context.Context, accountID string, limit, offset int32) ([]models.APIKey, error) {
	uid, err := uuidFromString(accountID)
	if err != nil {
		return nil, fmt.Errorf("list api keys by account id: %w", err)
	}

	rows, err := r.queries.ListAPIKeysByAccountID(ctx, sqlc.ListAPIKeysByAccountIDParams{
		AccountID: uid,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, translateError("list api keys by account id", err)
	}

	items := make([]models.APIKey, 0, len(rows))
	for _, row := range rows {
		items = append(items, apiKeyFromRow(row))
	}
	return items, nil
}

func (r *APIKeyRepository) CountActiveByAccountID(ctx context.Context, accountID string) (int64, error) {
	uid, err := uuidFromString(accountID)
	if err != nil {
		return 0, fmt.Errorf("count active api keys by account id: %w", err)
	}

	count, err := r.queries.CountActiveAPIKeysByAccountID(ctx, uid)
	if err != nil {
		return 0, translateError("count active api keys by account id", err)
	}
	return count, nil
}

func (r *APIKeyRepository) Touch(ctx context.Context, id string) (models.APIKey, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.APIKey{}, fmt.Errorf("touch api key: %w", err)
	}

	row, err := r.queries.TouchAPIKey(ctx, uid)
	if err != nil {
		return models.APIKey{}, translateError("touch api key", err)
	}
	return apiKeyFromRow(row), nil
}

func (r *APIKeyRepository) Revoke(ctx context.Context, id string) (models.APIKey, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.APIKey{}, fmt.Errorf("revoke api key: %w", err)
	}

	row, err := r.queries.RevokeAPIKey(ctx, uid)
	if err != nil {
		return models.APIKey{}, translateError("revoke api key", err)
	}
	return apiKeyFromRow(row), nil
}

func (r *APIKeyRepository) Rotate(ctx context.Context, id string) (models.APIKey, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.APIKey{}, fmt.Errorf("rotate api key: %w", err)
	}

	row, err := r.queries.RotateAPIKey(ctx, uid)
	if err != nil {
		return models.APIKey{}, translateError("rotate api key", err)
	}
	return apiKeyFromRow(row), nil
}

func (r *APIKeyRepository) DeleteExpired(ctx context.Context) error {
	return translateError("delete expired api keys", r.queries.DeleteExpiredAPIKeys(ctx))
}
