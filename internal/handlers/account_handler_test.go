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

type accountHandlerStub struct {
	getCurrentFn    func(context.Context, string) (models.AccountResponse, error)
	updateCurrentFn func(context.Context, string, models.AccountUpdateInput) (models.AccountResponse, error)
	deactivateFn    func(context.Context, string) error

	getCurrentCalls    int
	updateCurrentCalls int
	deactivateCalls    int

	lastAccountID string
	lastUpdate    models.AccountUpdateInput
}

func (s *accountHandlerStub) GetCurrent(ctx context.Context, accountID string) (models.AccountResponse, error) {
	s.getCurrentCalls++
	s.lastAccountID = accountID
	if s.getCurrentFn != nil {
		return s.getCurrentFn(ctx, accountID)
	}
	return models.AccountResponse{}, nil
}

func (s *accountHandlerStub) UpdateCurrent(ctx context.Context, accountID string, input models.AccountUpdateInput) (models.AccountResponse, error) {
	s.updateCurrentCalls++
	s.lastAccountID = accountID
	s.lastUpdate = input
	if s.updateCurrentFn != nil {
		return s.updateCurrentFn(ctx, accountID, input)
	}
	return models.AccountResponse{}, nil
}

func (s *accountHandlerStub) DeactivateCurrent(ctx context.Context, accountID string) error {
	s.deactivateCalls++
	s.lastAccountID = accountID
	if s.deactivateFn != nil {
		return s.deactivateFn(ctx, accountID)
	}
	return nil
}

func TestAccountHandlerGetCurrentReturnsSafeAccount(t *testing.T) {
	stub := &accountHandlerStub{
		getCurrentFn: func(ctx context.Context, accountID string) (models.AccountResponse, error) {
			if accountID != "acc_123" {
				t.Fatalf("unexpected account id: %q", accountID)
			}
			now := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
			return models.AccountResponse{
				ID:        "acc_123",
				Username:  "alice",
				Status:    models.AccountStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}
	h, err := NewAccountHandler(stub)
	if err != nil {
		t.Fatalf("NewAccountHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/account", nil)
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type: %q", got)
	}
	if stub.getCurrentCalls != 1 {
		t.Fatalf("unexpected get call count: %d", stub.getCurrentCalls)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body["success"] != true {
		t.Fatalf("unexpected success flag: %#v", body["success"])
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data was not an object: %#v", body["data"])
	}
	if data["id"] != "acc_123" || data["username"] != "alice" {
		t.Fatalf("unexpected account payload: %#v", data)
	}
	if _, exists := data["password_hash"]; exists {
		t.Fatal("password hash leaked in account response")
	}
	if _, exists := data["password"]; exists {
		t.Fatal("password leaked in account response")
	}

	if stub.lastAccountID != "acc_123" {
		t.Fatalf("unexpected account id passed to service: %q", stub.lastAccountID)
	}
}

func TestAccountHandlerUpdateCurrentValidatesAndUpdates(t *testing.T) {
	stub := &accountHandlerStub{
		updateCurrentFn: func(ctx context.Context, accountID string, input models.AccountUpdateInput) (models.AccountResponse, error) {
			if accountID != "acc_123" {
				t.Fatalf("unexpected account id: %q", accountID)
			}
			if input.Username == nil || *input.Username != "alice" {
				t.Fatalf("unexpected username: %#v", input.Username)
			}
			if input.Email == nil || *input.Email != "alice@example.com" {
				t.Fatalf("unexpected email: %#v", input.Email)
			}
			now := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
			return models.AccountResponse{
				ID:        "acc_123",
				Username:  "alice",
				Email:     input.Email,
				Status:    models.AccountStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}
	h, err := NewAccountHandler(stub)
	if err != nil {
		t.Fatalf("NewAccountHandler() error = %v", err)
	}

	body := `{"username":" Alice ","email":" Alice@Example.com "}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/v1/account", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.updateCurrentCalls != 1 {
		t.Fatalf("unexpected update call count: %d", stub.updateCurrentCalls)
	}
	if stub.lastUpdate.Username == nil || *stub.lastUpdate.Username != "alice" {
		t.Fatalf("unexpected normalized username: %#v", stub.lastUpdate.Username)
	}
	if stub.lastUpdate.Email == nil || *stub.lastUpdate.Email != "alice@example.com" {
		t.Fatalf("unexpected normalized email: %#v", stub.lastUpdate.Email)
	}

	var bodyMap map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &bodyMap); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	data := bodyMap["data"].(map[string]any)
	if data["id"] != "acc_123" || data["username"] != "alice" {
		t.Fatalf("unexpected account response: %#v", data)
	}
}

func TestAccountHandlerDeleteCurrentReturnsNoContent(t *testing.T) {
	stub := &accountHandlerStub{
		deactivateFn: func(ctx context.Context, accountID string) error {
			if accountID != "acc_123" {
				t.Fatalf("unexpected account id: %q", accountID)
			}
			return nil
		},
	}
	h, err := NewAccountHandler(stub)
	if err != nil {
		t.Fatalf("NewAccountHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v1/account", nil)
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("delete should not write a body: %q", rr.Body.String())
	}
	if stub.deactivateCalls != 1 {
		t.Fatalf("unexpected deactivate call count: %d", stub.deactivateCalls)
	}
}

func TestAccountHandlerRejectsInvalidBodiesAndMissingIdentity(t *testing.T) {
	stub := &accountHandlerStub{}
	h, err := NewAccountHandler(stub)
	if err != nil {
		t.Fatalf("NewAccountHandler() error = %v", err)
	}

	patchCases := []struct {
		name string
		body string
	}{
		{name: "empty body", body: ""},
		{name: "unknown field", body: `{"username":"alice","unknown":true}`},
		{name: "multiple json values", body: `{"username":"alice"}{"email":"alice@example.com"}`},
	}

	for _, tc := range patchCases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/v1/account", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))

			h.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			var body response.ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if body.Error.Code != "INVALID_REQUEST" {
				t.Fatalf("unexpected error code: %q", body.Error.Code)
			}
		})
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/account", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status for missing identity: %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/account", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status for unsupported method: %d", rr.Code)
	}
	if got := rr.Header().Get("Allow"); got != "GET, PATCH, DELETE" {
		t.Fatalf("unexpected allow header: %q", got)
	}
}

func TestAccountHandlerConvertsValidationAndServiceErrors(t *testing.T) {
	stub := &accountHandlerStub{
		updateCurrentFn: func(ctx context.Context, accountID string, input models.AccountUpdateInput) (models.AccountResponse, error) {
			return models.AccountResponse{}, services.ErrUsernameUnavailable
		},
	}
	h, err := NewAccountHandler(stub)
	if err != nil {
		t.Fatalf("NewAccountHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/v1/account", strings.NewReader(`{"username":"","email":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected validation status: %d", rr.Code)
	}
	var validation response.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &validation); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if validation.Error.Code != "VALIDATION_FAILED" {
		t.Fatalf("unexpected validation code: %q", validation.Error.Code)
	}
	if len(validation.Error.Details) != 2 || validation.Error.Details[0].Field != "username" || validation.Error.Details[1].Field != "email" {
		t.Fatalf("unexpected validation details: %#v", validation.Error.Details)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/v1/account", strings.NewReader(`{"username":"newname"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("unexpected service error status: %d", rr.Code)
	}
	var conflict response.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &conflict); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if conflict.Error.Code != "RESOURCE_CONFLICT" {
		t.Fatalf("unexpected conflict code: %q", conflict.Error.Code)
	}
}

func TestAccountHandlerRejectsOversizedBodies(t *testing.T) {
	stub := &accountHandlerStub{}
	h, err := NewAccountHandler(stub)
	if err != nil {
		t.Fatalf("NewAccountHandler() error = %v", err)
	}

	oversized := `{"username":"` + strings.Repeat("a", accountRequestBodyLimit) + `"}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/v1/account", strings.NewReader(oversized))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	var body response.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Error.Code != "INVALID_REQUEST" {
		t.Fatalf("unexpected error code: %q", body.Error.Code)
	}
}
