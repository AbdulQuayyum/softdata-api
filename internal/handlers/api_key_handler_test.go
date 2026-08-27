package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/middlewares"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type apiKeyHandlerStub struct {
	createFn func(context.Context, string, models.APIKeyCreateInput) (models.APIKeyCreatedResponse, error)
	listFn   func(context.Context, string) ([]models.APIKeyMetadata, error)
	revokeFn func(context.Context, string, string) error
	rotateFn func(context.Context, string, string) (models.APIKeyCreatedResponse, error)

	createCalls int
	listCalls   int
	revokeCalls int
	rotateCalls int

	lastAccountID string
	lastKeyID     string
	lastCreate    models.APIKeyCreateInput
}

func (s *apiKeyHandlerStub) CreateKey(ctx context.Context, accountID string, input models.APIKeyCreateInput) (models.APIKeyCreatedResponse, error) {
	s.createCalls++
	s.lastAccountID = accountID
	s.lastCreate = input
	if s.createFn != nil {
		return s.createFn(ctx, accountID, input)
	}
	return models.APIKeyCreatedResponse{}, nil
}

func (s *apiKeyHandlerStub) ListKeys(ctx context.Context, accountID string) ([]models.APIKeyMetadata, error) {
	s.listCalls++
	s.lastAccountID = accountID
	if s.listFn != nil {
		return s.listFn(ctx, accountID)
	}
	return nil, nil
}

func (s *apiKeyHandlerStub) RevokeKey(ctx context.Context, accountID, keyID string) error {
	s.revokeCalls++
	s.lastAccountID = accountID
	s.lastKeyID = keyID
	if s.revokeFn != nil {
		return s.revokeFn(ctx, accountID, keyID)
	}
	return nil
}

func (s *apiKeyHandlerStub) RotateKey(ctx context.Context, accountID, keyID string) (models.APIKeyCreatedResponse, error) {
	s.rotateCalls++
	s.lastAccountID = accountID
	s.lastKeyID = keyID
	if s.rotateFn != nil {
		return s.rotateFn(ctx, accountID, keyID)
	}
	return models.APIKeyCreatedResponse{}, nil
}

func TestNewAPIKeyHandlerRejectsNilService(t *testing.T) {
	if _, err := NewAPIKeyHandler(nil); err == nil {
		t.Fatal("NewAPIKeyHandler(nil) error = nil, want error")
	}
}

func TestAPIKeyHandlerCreateKeyReturnsOneTimePlaintextKey(t *testing.T) {
	stub := &apiKeyHandlerStub{
		createFn: func(ctx context.Context, accountID string, input models.APIKeyCreateInput) (models.APIKeyCreatedResponse, error) {
			if accountID != "acc_123" {
				t.Fatalf("unexpected account id: %q", accountID)
			}
			if input.Name != "Portfolio Application" {
				t.Fatalf("unexpected key name: %q", input.Name)
			}
			now := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
			return models.APIKeyCreatedResponse{
				Key: "sd_live_abc123",
				APIKey: models.APIKeyMetadata{
					ID:        "key_123",
					Name:      input.Name,
					KeyPrefix: "sd_live_",
					KeyLast4:  "c123",
					Status:    models.APIKeyStatusActive,
					CreatedAt: now,
					UpdatedAt: now,
				},
			}, nil
		},
	}
	h, err := NewAPIKeyHandler(stub)
	if err != nil {
		t.Fatalf("NewAPIKeyHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/account/api-keys", strings.NewReader(`{"name":" Portfolio Application "}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	rr := httptest.NewRecorder()

	h.CreateKey(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type: %q", got)
	}
	if stub.createCalls != 1 {
		t.Fatalf("unexpected create call count: %d", stub.createCalls)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body["success"] != true {
		t.Fatalf("unexpected success flag: %#v", body["success"])
	}
	data := body["data"].(map[string]any)
	if data["key"] != "sd_live_abc123" {
		t.Fatalf("unexpected key payload: %#v", data)
	}
	apiKey := data["api_key"].(map[string]any)
	if apiKey["id"] != "key_123" || apiKey["name"] != "Portfolio Application" {
		t.Fatalf("unexpected api_key payload: %#v", apiKey)
	}
	if _, exists := apiKey["key_hash"]; exists {
		t.Fatal("key hash leaked in create response")
	}
}

func TestAPIKeyHandlerListKeysReturnsSafeMetadataAndEmptySlice(t *testing.T) {
	stub := &apiKeyHandlerStub{
		listFn: func(ctx context.Context, accountID string) ([]models.APIKeyMetadata, error) {
			if accountID != "acc_123" {
				t.Fatalf("unexpected account id: %q", accountID)
			}
			return nil, nil
		},
	}
	h, err := NewAPIKeyHandler(stub)
	if err != nil {
		t.Fatalf("NewAPIKeyHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/account/api-keys", nil)
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	rr := httptest.NewRecorder()

	h.ListKeys(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.listCalls != 1 {
		t.Fatalf("unexpected list call count: %d", stub.listCalls)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	data := body["data"].([]any)
	if len(data) != 0 {
		t.Fatalf("expected empty array, got %#v", data)
	}
}

func TestAPIKeyHandlerRevokeKeyUsesPathValueAndReturnsNoContent(t *testing.T) {
	stub := &apiKeyHandlerStub{
		revokeFn: func(ctx context.Context, accountID, keyID string) error {
			if accountID != "acc_123" {
				t.Fatalf("unexpected account id: %q", accountID)
			}
			if keyID != "key_123" {
				t.Fatalf("unexpected key id: %q", keyID)
			}
			return nil
		},
	}
	h, err := NewAPIKeyHandler(stub)
	if err != nil {
		t.Fatalf("NewAPIKeyHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/account/api-keys/key_123", nil)
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	req.SetPathValue("key_id", " key_123 ")
	rr := httptest.NewRecorder()

	h.RevokeKey(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("revoke should not write a body: %q", rr.Body.String())
	}
	if stub.revokeCalls != 1 {
		t.Fatalf("unexpected revoke call count: %d", stub.revokeCalls)
	}
}

func TestAPIKeyHandlerRotateKeyReturnsReplacementKey(t *testing.T) {
	stub := &apiKeyHandlerStub{
		rotateFn: func(ctx context.Context, accountID, keyID string) (models.APIKeyCreatedResponse, error) {
			if accountID != "acc_123" {
				t.Fatalf("unexpected account id: %q", accountID)
			}
			if keyID != "key_123" {
				t.Fatalf("unexpected key id: %q", keyID)
			}
			now := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
			return models.APIKeyCreatedResponse{
				Key: "sd_live_newkey",
				APIKey: models.APIKeyMetadata{
					ID:        "key_456",
					Name:      "Portfolio Application",
					KeyPrefix: "sd_live_",
					KeyLast4:  "wkey",
					Status:    models.APIKeyStatusActive,
					CreatedAt: now,
					UpdatedAt: now,
				},
			}, nil
		},
	}
	h, err := NewAPIKeyHandler(stub)
	if err != nil {
		t.Fatalf("NewAPIKeyHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/account/api-keys/key_123/rotate", nil)
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	req.SetPathValue("key_id", "key_123")
	rr := httptest.NewRecorder()

	h.RotateKey(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.rotateCalls != 1 {
		t.Fatalf("unexpected rotate call count: %d", stub.rotateCalls)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	data := body["data"].(map[string]any)
	if data["key"] != "sd_live_newkey" {
		t.Fatalf("unexpected key payload: %#v", data)
	}
}

func TestAPIKeyHandlerRejectsInvalidInputsAndMissingIdentity(t *testing.T) {
	stub := &apiKeyHandlerStub{}
	h, err := NewAPIKeyHandler(stub)
	if err != nil {
		t.Fatalf("NewAPIKeyHandler() error = %v", err)
	}

	createCases := []struct {
		name string
		body string
	}{
		{name: "empty body", body: ""},
		{name: "unknown field", body: `{"name":"Portfolio","unexpected":true}`},
		{name: "blank name", body: `{"name":"   "}`},
	}

	for _, tc := range createCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/account/api-keys", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
			rr := httptest.NewRecorder()

			h.CreateKey(rr, req)

			switch tc.name {
			case "blank name":
				if rr.Code != http.StatusUnprocessableEntity {
					t.Fatalf("unexpected status: %d", rr.Code)
				}
			default:
				if rr.Code != http.StatusBadRequest {
					t.Fatalf("unexpected status: %d", rr.Code)
				}
			}
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/account/api-keys", nil)
	rr := httptest.NewRecorder()
	h.ListKeys(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status for missing identity: %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, "/v1/account/api-keys/", nil)
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	rr = httptest.NewRecorder()
	h.RevokeKey(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status for missing key id: %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/account/api-keys/key_123/rotate", nil)
	req.SetPathValue("key_id", "key_123")
	rr = httptest.NewRecorder()
	h.RotateKey(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status for missing identity on rotate: %d", rr.Code)
	}
}

func TestAPIKeyHandlerMapsServiceErrors(t *testing.T) {
	stub := &apiKeyHandlerStub{
		createFn: func(context.Context, string, models.APIKeyCreateInput) (models.APIKeyCreatedResponse, error) {
			return models.APIKeyCreatedResponse{}, services.ErrAPIKeyLimitReached
		},
	}
	h, err := NewAPIKeyHandler(stub)
	if err != nil {
		t.Fatalf("NewAPIKeyHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/account/api-keys", strings.NewReader(`{"name":"Portfolio"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	rr := httptest.NewRecorder()

	h.CreateKey(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("unexpected status: %d", rr.Code)
	}

	var body response.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Error.Code != "RESOURCE_CONFLICT" {
		t.Fatalf("unexpected error code: %q", body.Error.Code)
	}
}
