package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type apiKeyRepoStub struct {
	createFn            func(context.Context, string, models.APIKeyCreateInput, string, string, string, *time.Time, *time.Time, *time.Time) (models.APIKey, error)
	getByIDFn           func(context.Context, string) (models.APIKey, error)
	listFn              func(context.Context, string, int32, int32) ([]models.APIKey, error)
	countFn             func(context.Context, string) (int64, error)
	revokeFn            func(context.Context, string) (models.APIKey, error)
	rotateFn            func(context.Context, string) (models.APIKey, error)
	createCalls         int
	getByIDCalls        int
	listCalls           int
	countCalls          int
	revokeCalls         int
	rotateCalls         int
	lastCreateAccountID string
	lastCreateInput     models.APIKeyCreateInput
	lastCreatePrefix    string
	lastCreateHash      string
	lastCreateLast4     string
	lastCreateExpiresAt *time.Time
	lastListAccountID   string
	lastListLimit       int32
	lastListOffset      int32
	lastRevokeID        string
	keysByID            map[string]models.APIKey
}

func newAPIKeyRepoStub() *apiKeyRepoStub {
	return &apiKeyRepoStub{
		keysByID: map[string]models.APIKey{},
	}
}

func (s *apiKeyRepoStub) Create(ctx context.Context, accountID string, input models.APIKeyCreateInput, keyPrefix, keyHash, keyLast4 string, expiresAt, lastUsedAt, revokedAt *time.Time) (models.APIKey, error) {
	s.createCalls++
	s.lastCreateAccountID = accountID
	s.lastCreateInput = input
	s.lastCreatePrefix = keyPrefix
	s.lastCreateHash = keyHash
	s.lastCreateLast4 = keyLast4
	s.lastCreateExpiresAt = expiresAt
	if s.createFn != nil {
		return s.createFn(ctx, accountID, input, keyPrefix, keyHash, keyLast4, expiresAt, lastUsedAt, revokedAt)
	}
	key := models.APIKey{
		ID:        "key-created",
		AccountID: accountID,
		Name:      input.Name,
		KeyPrefix: keyPrefix,
		KeyHash:   keyHash,
		KeyLast4:  keyLast4,
		Status:    models.APIKeyStatusActive,
		ExpiresAt: expiresAt,
		CreatedAt: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC),
	}
	s.keysByID[key.ID] = key
	return key, nil
}

func (s *apiKeyRepoStub) GetByID(ctx context.Context, id string) (models.APIKey, error) {
	s.getByIDCalls++
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	if key, ok := s.keysByID[id]; ok {
		return key, nil
	}
	return models.APIKey{}, interfaces.ErrNotFound
}

func (s *apiKeyRepoStub) GetByKeyHash(ctx context.Context, keyHash string) (models.APIKey, error) {
	for _, key := range s.keysByID {
		if key.KeyHash == keyHash {
			return key, nil
		}
	}
	return models.APIKey{}, interfaces.ErrNotFound
}

func (s *apiKeyRepoStub) ListByAccountID(ctx context.Context, accountID string, limit, offset int32) ([]models.APIKey, error) {
	s.listCalls++
	s.lastListAccountID = accountID
	s.lastListLimit = limit
	s.lastListOffset = offset
	if s.listFn != nil {
		return s.listFn(ctx, accountID, limit, offset)
	}
	items := make([]models.APIKey, 0, len(s.keysByID))
	for _, key := range s.keysByID {
		if key.AccountID == accountID {
			items = append(items, key)
		}
	}
	return items, nil
}

func (s *apiKeyRepoStub) CountActiveByAccountID(ctx context.Context, accountID string) (int64, error) {
	s.countCalls++
	if s.countFn != nil {
		return s.countFn(ctx, accountID)
	}
	return 0, nil
}

func (s *apiKeyRepoStub) Touch(ctx context.Context, id string) (models.APIKey, error) {
	if key, ok := s.keysByID[id]; ok {
		return key, nil
	}
	return models.APIKey{}, interfaces.ErrNotFound
}

func (s *apiKeyRepoStub) Revoke(ctx context.Context, id string) (models.APIKey, error) {
	s.revokeCalls++
	s.lastRevokeID = id
	if s.revokeFn != nil {
		return s.revokeFn(ctx, id)
	}
	key, ok := s.keysByID[id]
	if !ok {
		return models.APIKey{}, interfaces.ErrNotFound
	}
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	key.Status = models.APIKeyStatusRevoked
	key.RevokedAt = &now
	key.UpdatedAt = now
	s.keysByID[id] = key
	return key, nil
}

func (s *apiKeyRepoStub) Rotate(ctx context.Context, id string) (models.APIKey, error) {
	s.rotateCalls++
	if s.rotateFn != nil {
		return s.rotateFn(ctx, id)
	}
	return s.Revoke(ctx, id)
}

func (s *apiKeyRepoStub) DeleteExpired(ctx context.Context) error {
	return nil
}

type apiKeyGeneratorStub struct {
	plaintext string
	hash      string
	prefix    string
	last4     string
	err       error
	calls     int
}

func (g *apiKeyGeneratorStub) Generate() (string, string, string, string, error) {
	g.calls++
	return g.plaintext, g.hash, g.prefix, g.last4, g.err
}

func TestAPIKeyServiceCreateKey(t *testing.T) {
	repo := newAPIKeyRepoStub()
	repo.countFn = func(context.Context, string) (int64, error) {
		return 1, nil
	}
	repo.createFn = func(_ context.Context, accountID string, input models.APIKeyCreateInput, prefix, hash, last4 string, expiresAt, lastUsedAt, revokedAt *time.Time) (models.APIKey, error) {
		if accountID != "acct-1" {
			t.Fatalf("unexpected account id: %q", accountID)
		}
		if input.Name != "Portfolio Application" {
			t.Fatalf("name was not trimmed correctly: %q", input.Name)
		}
		if prefix != "sd_live_" || hash != "stored-hash" || last4 != "abcd" {
			t.Fatalf("unexpected persisted key parts: %q %q %q", prefix, hash, last4)
		}
		if expiresAt != nil || lastUsedAt != nil || revokedAt != nil {
			t.Fatalf("unexpected timestamps: expires=%v lastUsed=%v revoked=%v", expiresAt, lastUsedAt, revokedAt)
		}
		now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
		return models.APIKey{
			ID:        "key-1",
			AccountID: accountID,
			Name:      input.Name,
			KeyPrefix: prefix,
			KeyHash:   hash,
			KeyLast4:  last4,
			Status:    models.APIKeyStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	svc, err := NewAPIKeyService(repo, &apiKeyGeneratorStub{
		plaintext: "sd_live_test-plaintext",
		hash:      "stored-hash",
		prefix:    "sd_live_",
		last4:     "abcd",
	})
	if err != nil {
		t.Fatalf("NewAPIKeyService() error = %v", err)
	}

	result, err := svc.CreateKey(context.Background(), "acct-1", models.APIKeyCreateInput{Name: " Portfolio Application "})
	if err != nil {
		t.Fatalf("CreateKey() error = %v", err)
	}
	if result.Key != "sd_live_test-plaintext" {
		t.Fatalf("unexpected plaintext key: %q", result.Key)
	}
	if result.APIKey.KeyPrefix != "sd_live_" || result.APIKey.KeyLast4 != "abcd" {
		t.Fatalf("unexpected API key metadata: %#v", result.APIKey)
	}
	if repo.lastCreateHash != "stored-hash" {
		t.Fatalf("plaintext key reached repository: %q", repo.lastCreateHash)
	}
	if repo.countCalls != 1 || repo.createCalls != 1 {
		t.Fatalf("unexpected repo calls: count=%d create=%d", repo.countCalls, repo.createCalls)
	}
	marshaled, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if strings.Count(string(marshaled), "sd_live_test-plaintext") != 1 {
		t.Fatalf("plaintext key was not returned exactly once: %s", marshaled)
	}
}

func TestAPIKeyServiceCreateKeyLimitReached(t *testing.T) {
	repo := newAPIKeyRepoStub()
	repo.countFn = func(context.Context, string) (int64, error) {
		return apiKeyActiveLimit, nil
	}
	gen := &apiKeyGeneratorStub{
		plaintext: "sd_live_limit",
		hash:      "hash",
		prefix:    "sd_live_",
		last4:     "1234",
	}
	svc, err := NewAPIKeyService(repo, gen)
	if err != nil {
		t.Fatalf("NewAPIKeyService() error = %v", err)
	}

	_, err = svc.CreateKey(context.Background(), "acct-1", models.APIKeyCreateInput{Name: "Key"})
	if !errors.Is(err, ErrAPIKeyLimitReached) {
		t.Fatalf("CreateKey() error = %v, want ErrAPIKeyLimitReached", err)
	}
	if gen.calls != 0 {
		t.Fatalf("generator should not be called when limit is reached")
	}
}

func TestAPIKeyServiceCreateKeyFailureDoesNotExposePlaintext(t *testing.T) {
	repo := newAPIKeyRepoStub()
	repo.countFn = func(context.Context, string) (int64, error) {
		return 0, nil
	}
	repo.createFn = func(context.Context, string, models.APIKeyCreateInput, string, string, string, *time.Time, *time.Time, *time.Time) (models.APIKey, error) {
		return models.APIKey{}, errors.New("repository unavailable")
	}
	gen := &apiKeyGeneratorStub{
		plaintext: "sd_live_secret-key",
		hash:      "hash",
		prefix:    "sd_live_",
		last4:     "4321",
	}
	svc, err := NewAPIKeyService(repo, gen)
	if err != nil {
		t.Fatalf("NewAPIKeyService() error = %v", err)
	}

	_, err = svc.CreateKey(context.Background(), "acct-1", models.APIKeyCreateInput{Name: "Key"})
	if err == nil {
		t.Fatal("CreateKey() error = nil, want error")
	}
	if strings.Contains(err.Error(), gen.plaintext) {
		t.Fatalf("error exposed plaintext key: %v", err)
	}
}

func TestAPIKeyServiceListKeysFiltersOwnership(t *testing.T) {
	repo := newAPIKeyRepoStub()
	repo.listFn = func(context.Context, string, int32, int32) ([]models.APIKey, error) {
		now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
		return []models.APIKey{
			{
				ID:        "key-1",
				AccountID: "acct-1",
				Name:      "Primary",
				KeyPrefix: "sd_live_",
				KeyLast4:  "1111",
				Status:    models.APIKeyStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:        "key-2",
				AccountID: "acct-2",
				Name:      "Wrong Owner",
				KeyPrefix: "sd_live_",
				KeyLast4:  "2222",
				Status:    models.APIKeyStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}, nil
	}
	svc, err := NewAPIKeyService(repo, &apiKeyGeneratorStub{})
	if err != nil {
		t.Fatalf("NewAPIKeyService() error = %v", err)
	}

	keys, err := svc.ListKeys(context.Background(), "acct-1")
	if err != nil {
		t.Fatalf("ListKeys() error = %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected one owned key, got %d", len(keys))
	}
	if keys[0].ID != "key-1" || keys[0].Name != "Primary" {
		t.Fatalf("unexpected key metadata: %#v", keys[0])
	}
	marshaled, err := json.Marshal(keys)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if strings.Contains(string(marshaled), "key_hash") || strings.Contains(string(marshaled), "sd_live_secret") {
		t.Fatalf("unsafe data leaked from list response: %s", marshaled)
	}
}

func TestAPIKeyServiceRevokeKeyEnforcesOwnership(t *testing.T) {
	repo := newAPIKeyRepoStub()
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	repo.keysByID["key-1"] = models.APIKey{
		ID:        "key-1",
		AccountID: "acct-1",
		Name:      "Primary",
		KeyPrefix: "sd_live_",
		KeyLast4:  "1111",
		Status:    models.APIKeyStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	svc, err := NewAPIKeyService(repo, &apiKeyGeneratorStub{})
	if err != nil {
		t.Fatalf("NewAPIKeyService() error = %v", err)
	}

	if err := svc.RevokeKey(context.Background(), "acct-2", "key-1"); !errors.Is(err, ErrAPIKeyNotFound) {
		t.Fatalf("RevokeKey() error = %v, want ErrAPIKeyNotFound", err)
	}
	if repo.revokeCalls != 0 {
		t.Fatalf("revoke should not be called for a non-owned key")
	}
}

func TestAPIKeyServiceRevokeKeyIsIdempotent(t *testing.T) {
	repo := newAPIKeyRepoStub()
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	repo.keysByID["key-1"] = models.APIKey{
		ID:        "key-1",
		AccountID: "acct-1",
		Name:      "Primary",
		KeyPrefix: "sd_live_",
		KeyLast4:  "1111",
		Status:    models.APIKeyStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	svc, err := NewAPIKeyService(repo, &apiKeyGeneratorStub{})
	if err != nil {
		t.Fatalf("NewAPIKeyService() error = %v", err)
	}

	if err := svc.RevokeKey(context.Background(), "acct-1", "key-1"); err != nil {
		t.Fatalf("first RevokeKey() error = %v", err)
	}
	if err := svc.RevokeKey(context.Background(), "acct-1", "key-1"); err != nil {
		t.Fatalf("second RevokeKey() error = %v", err)
	}
	if repo.revokeCalls != 1 {
		t.Fatalf("revoke should only be called once, got %d", repo.revokeCalls)
	}
}

func TestAPIKeyServiceRotateKeyCreatesReplacement(t *testing.T) {
	repo := newAPIKeyRepoStub()
	expiresAt := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	repo.keysByID["key-1"] = models.APIKey{
		ID:        "key-1",
		AccountID: "acct-1",
		Name:      "Primary",
		KeyPrefix: "sd_live_old_",
		KeyLast4:  "1111",
		Status:    models.APIKeyStatusActive,
		ExpiresAt: &expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
	gen := &apiKeyGeneratorStub{
		plaintext: "sd_live_rotated",
		hash:      "rotated-hash",
		prefix:    "sd_live_",
		last4:     "9999",
	}
	repo.createFn = func(_ context.Context, accountID string, input models.APIKeyCreateInput, prefix, hash, last4 string, expiresAtArg, lastUsedAt, revokedAt *time.Time) (models.APIKey, error) {
		if accountID != "acct-1" || input.Name != "Primary" {
			t.Fatalf("unexpected rotate create input: %q %#v", accountID, input)
		}
		if expiresAtArg == nil || !expiresAtArg.Equal(expiresAt) {
			t.Fatalf("rotate should preserve expiry: %#v", expiresAtArg)
		}
		return models.APIKey{
			ID:        "key-2",
			AccountID: accountID,
			Name:      input.Name,
			KeyPrefix: prefix,
			KeyHash:   hash,
			KeyLast4:  last4,
			Status:    models.APIKeyStatusActive,
			ExpiresAt: expiresAtArg,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}
	svc, err := NewAPIKeyService(repo, gen)
	if err != nil {
		t.Fatalf("NewAPIKeyService() error = %v", err)
	}

	result, err := svc.RotateKey(context.Background(), "acct-1", "key-1")
	if err != nil {
		t.Fatalf("RotateKey() error = %v", err)
	}
	if result.Key != "sd_live_rotated" {
		t.Fatalf("unexpected rotated plaintext: %q", result.Key)
	}
	if repo.revokeCalls != 1 || repo.createCalls != 1 {
		t.Fatalf("unexpected rotate calls: revoke=%d create=%d", repo.revokeCalls, repo.createCalls)
	}
}
