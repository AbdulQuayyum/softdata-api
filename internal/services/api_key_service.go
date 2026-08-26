package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/security"
)

const apiKeyActiveLimit = 3

type APIKeyGenerator interface {
	Generate() (plaintext, hash, prefix, last4 string, err error)
}

type securityAPIKeyGenerator struct{}

func NewSecurityAPIKeyGenerator() APIKeyGenerator {
	return securityAPIKeyGenerator{}
}

func (securityAPIKeyGenerator) Generate() (string, string, string, string, error) {
	return security.GenerateAPIKey()
}

type APIKeyService struct {
	apiKeys   interfaces.APIKeyRepository
	generator APIKeyGenerator
}

// APIKeyIdentity is the safe authentication result for a validated API key.
type APIKeyIdentity struct {
	APIKeyID  string
	AccountID string
}

func NewAPIKeyService(apiKeys interfaces.APIKeyRepository, generator APIKeyGenerator) (*APIKeyService, error) {
	if apiKeys == nil {
		return nil, fmt.Errorf("api key repository is required")
	}
	if generator == nil {
		return nil, fmt.Errorf("api key generator is required")
	}

	return &APIKeyService{apiKeys: apiKeys, generator: generator}, nil
}

func (s *APIKeyService) CreateKey(ctx context.Context, accountID string, input models.APIKeyCreateInput) (models.APIKeyCreatedResponse, error) {
	accountID = normalizeAPIKeyAccountID(accountID)
	if accountID == "" {
		return models.APIKeyCreatedResponse{}, fmt.Errorf("create api key: account id is required")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return models.APIKeyCreatedResponse{}, ErrAPIKeyNameRequired
	}

	activeCount, err := s.apiKeys.CountActiveByAccountID(ctx, accountID)
	if err != nil {
		return models.APIKeyCreatedResponse{}, fmt.Errorf("count active api keys: %w", err)
	}
	if activeCount >= apiKeyActiveLimit {
		return models.APIKeyCreatedResponse{}, ErrAPIKeyLimitReached
	}

	plaintext, hash, prefix, last4, err := s.generator.Generate()
	if err != nil {
		return models.APIKeyCreatedResponse{}, fmt.Errorf("generate api key: %w", err)
	}

	created, err := s.apiKeys.Create(ctx, accountID, models.APIKeyCreateInput{Name: name}, prefix, hash, last4, nil, nil, nil)
	if err != nil {
		return models.APIKeyCreatedResponse{}, fmt.Errorf("create api key: %w", err)
	}

	return models.APIKeyCreatedResponse{
		Key:    plaintext,
		APIKey: apiKeyMetadataFromModel(created),
	}, nil
}

// Authenticate validates a plaintext API key and returns its safe identity.
func (s *APIKeyService) Authenticate(ctx context.Context, plaintext string) (APIKeyIdentity, error) {
	if err := ctx.Err(); err != nil {
		return APIKeyIdentity{}, err
	}
	if strings.TrimSpace(plaintext) == "" {
		return APIKeyIdentity{}, ErrAPIKeyNotFound
	}
	if _, err := security.DecodeAPIKeySuffix(plaintext); err != nil {
		return APIKeyIdentity{}, ErrAPIKeyNotFound
	}

	keyHash := security.HashAPIKey(plaintext)
	key, err := s.apiKeys.GetByKeyHash(ctx, keyHash)
	if err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return APIKeyIdentity{}, ErrAPIKeyNotFound
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return APIKeyIdentity{}, err
		}
		return APIKeyIdentity{}, fmt.Errorf("authenticate api key")
	}

	if key.ID == "" || key.AccountID == "" {
		return APIKeyIdentity{}, ErrAPIKeyNotFound
	}
	if key.Status != models.APIKeyStatusActive || key.RevokedAt != nil {
		return APIKeyIdentity{}, ErrAPIKeyNotFound
	}
	if key.ExpiresAt != nil && !key.ExpiresAt.After(time.Now().UTC()) {
		return APIKeyIdentity{}, ErrAPIKeyNotFound
	}

	return APIKeyIdentity{
		APIKeyID:  key.ID,
		AccountID: key.AccountID,
	}, nil
}

func (s *APIKeyService) ListKeys(ctx context.Context, accountID string) ([]models.APIKeyMetadata, error) {
	accountID = normalizeAPIKeyAccountID(accountID)
	if accountID == "" {
		return nil, fmt.Errorf("list api keys: account id is required")
	}

	rows, err := s.apiKeys.ListByAccountID(ctx, accountID, int32(math.MaxInt32), 0)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}

	items := make([]models.APIKeyMetadata, 0, len(rows))
	for _, key := range rows {
		if key.AccountID != accountID {
			continue
		}
		items = append(items, apiKeyMetadataFromModel(key))
	}
	return items, nil
}

func (s *APIKeyService) RevokeKey(ctx context.Context, accountID, keyID string) error {
	accountID = normalizeAPIKeyAccountID(accountID)
	keyID = normalizeAPIKeyAccountID(keyID)
	if accountID == "" || keyID == "" {
		return ErrAPIKeyNotFound
	}

	key, err := s.ownedAPIKey(ctx, accountID, keyID)
	if err != nil {
		return err
	}
	if key.RevokedAt != nil || key.Status == models.APIKeyStatusRevoked {
		return nil
	}

	if _, err := s.apiKeys.Revoke(ctx, key.ID); err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return ErrAPIKeyNotFound
		}
		return fmt.Errorf("revoke api key: %w", err)
	}

	return nil
}

func (s *APIKeyService) RotateKey(ctx context.Context, accountID, keyID string) (models.APIKeyCreatedResponse, error) {
	accountID = normalizeAPIKeyAccountID(accountID)
	keyID = normalizeAPIKeyAccountID(keyID)
	if accountID == "" || keyID == "" {
		return models.APIKeyCreatedResponse{}, ErrAPIKeyNotFound
	}

	key, err := s.ownedAPIKey(ctx, accountID, keyID)
	if err != nil {
		return models.APIKeyCreatedResponse{}, err
	}
	if key.RevokedAt != nil || key.Status == models.APIKeyStatusRevoked {
		return models.APIKeyCreatedResponse{}, ErrAPIKeyNotFound
	}

	if _, err := s.apiKeys.Revoke(ctx, key.ID); err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return models.APIKeyCreatedResponse{}, ErrAPIKeyNotFound
		}
		return models.APIKeyCreatedResponse{}, fmt.Errorf("revoke api key: %w", err)
	}

	plaintext, hash, prefix, last4, err := s.generator.Generate()
	if err != nil {
		return models.APIKeyCreatedResponse{}, fmt.Errorf("generate api key: %w", err)
	}

	expiresAt := key.ExpiresAt
	created, err := s.apiKeys.Create(ctx, accountID, models.APIKeyCreateInput{Name: key.Name}, prefix, hash, last4, expiresAt, nil, nil)
	if err != nil {
		return models.APIKeyCreatedResponse{}, fmt.Errorf("create rotated api key: %w", err)
	}

	return models.APIKeyCreatedResponse{
		Key:    plaintext,
		APIKey: apiKeyMetadataFromModel(created),
	}, nil
}

func (s *APIKeyService) ownedAPIKey(ctx context.Context, accountID, keyID string) (models.APIKey, error) {
	key, err := s.apiKeys.GetByID(ctx, keyID)
	if err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return models.APIKey{}, ErrAPIKeyNotFound
		}
		return models.APIKey{}, fmt.Errorf("get api key: %w", err)
	}
	if normalizeAPIKeyAccountID(key.AccountID) != accountID {
		return models.APIKey{}, ErrAPIKeyNotFound
	}
	return key, nil
}

func apiKeyMetadataFromModel(key models.APIKey) models.APIKeyMetadata {
	return models.APIKeyMetadata{
		ID:         key.ID,
		Name:       key.Name,
		KeyPrefix:  key.KeyPrefix,
		KeyLast4:   key.KeyLast4,
		Status:     key.Status,
		LastUsedAt: key.LastUsedAt,
		ExpiresAt:  key.ExpiresAt,
		RevokedAt:  key.RevokedAt,
		CreatedAt:  key.CreatedAt,
		UpdatedAt:  key.UpdatedAt,
	}
}

func normalizeAPIKeyAccountID(value string) string {
	return strings.TrimSpace(value)
}
