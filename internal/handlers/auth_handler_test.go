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
)

type authHandlerStub struct {
	registerFn func(context.Context, models.AccountCreateInput) (models.AccountResponse, error)
	loginFn    func(context.Context, string, string) (models.LoginResult, error)
	refreshFn  func(context.Context, string) (models.TokenPair, error)
	logoutFn   func(context.Context, string) error

	registerCalls int
	loginCalls    int
	refreshCalls  int
	logoutCalls   int

	lastRegisterInput models.AccountCreateInput
	lastLoginUsername string
	lastLoginPassword string
	lastRefreshToken  string
	lastLogoutToken   string
}

func (s *authHandlerStub) Register(ctx context.Context, input models.AccountCreateInput) (models.AccountResponse, error) {
	s.registerCalls++
	s.lastRegisterInput = input
	if s.registerFn != nil {
		return s.registerFn(ctx, input)
	}
	return models.AccountResponse{}, nil
}

func (s *authHandlerStub) Login(ctx context.Context, username, password string) (models.LoginResult, error) {
	s.loginCalls++
	s.lastLoginUsername = username
	s.lastLoginPassword = password
	if s.loginFn != nil {
		return s.loginFn(ctx, username, password)
	}
	return models.LoginResult{}, nil
}

func (s *authHandlerStub) Refresh(ctx context.Context, refreshToken string) (models.TokenPair, error) {
	s.refreshCalls++
	s.lastRefreshToken = refreshToken
	if s.refreshFn != nil {
		return s.refreshFn(ctx, refreshToken)
	}
	return models.TokenPair{}, nil
}

func (s *authHandlerStub) Logout(ctx context.Context, refreshToken string) error {
	s.logoutCalls++
	s.lastLogoutToken = refreshToken
	if s.logoutFn != nil {
		return s.logoutFn(ctx, refreshToken)
	}
	return nil
}

func TestAuthHandlerRegisterReturnsSafeAccountResponse(t *testing.T) {
	stub := &authHandlerStub{
		registerFn: func(ctx context.Context, input models.AccountCreateInput) (models.AccountResponse, error) {
			if got := input.Username; got != "alice" {
				t.Fatalf("unexpected username: %q", got)
			}
			if input.Email == nil || *input.Email != "alice@example.com" {
				t.Fatalf("unexpected email: %#v", input.Email)
			}
			if got := input.Password; got != "secret-password" {
				t.Fatalf("password was modified: %q", got)
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
	handler, err := NewAuthHandler(stub, stub)
	if err != nil {
		t.Fatalf("NewAuthHandler() error = %v", err)
	}

	email := " Alice@Example.com "
	body := `{"username":" Alice ","email":"` + email + `","password":"secret-password"}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	handler.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type: %q", got)
	}
	if stub.registerCalls != 1 {
		t.Fatalf("unexpected register call count: %d", stub.registerCalls)
	}

	var bodyMap map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &bodyMap); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if bodyMap["success"] != true {
		t.Fatalf("unexpected success flag: %#v", bodyMap["success"])
	}
	data, ok := bodyMap["data"].(map[string]any)
	if !ok {
		t.Fatalf("data was not an object: %#v", bodyMap["data"])
	}
	if data["id"] != "acc_123" || data["username"] != "alice" {
		t.Fatalf("unexpected account payload: %#v", data)
	}
	if _, exists := data["password"]; exists {
		t.Fatal("password leaked in register response")
	}
}

func TestAuthHandlerLoginReturnsLoginResult(t *testing.T) {
	stub := &authHandlerStub{
		loginFn: func(ctx context.Context, username, password string) (models.LoginResult, error) {
			if username != "alice" {
				t.Fatalf("unexpected username: %q", username)
			}
			if password != "  secret  " {
				t.Fatalf("password was modified: %q", password)
			}
			now := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
			return models.LoginResult{
				Account: models.AccountResponse{
					ID:        "acc_123",
					Username:  "alice",
					Status:    models.AccountStatusActive,
					CreatedAt: now,
					UpdatedAt: now,
				},
				Tokens: models.TokenPair{
					AccessToken:  "access.jwt.value",
					RefreshToken: "refresh.opaque.value",
					TokenType:    "Bearer",
					ExpiresIn:    900,
				},
			}, nil
		},
	}
	handler, err := NewAuthHandler(stub, stub)
	if err != nil {
		t.Fatalf("NewAuthHandler() error = %v", err)
	}

	body := `{"username":" Alice ","password":"  secret  "}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	handler.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.loginCalls != 1 {
		t.Fatalf("unexpected login call count: %d", stub.loginCalls)
	}

	var bodyMap map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &bodyMap); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	data := bodyMap["data"].(map[string]any)
	account := data["account"].(map[string]any)
	tokens := data["tokens"].(map[string]any)
	if account["id"] != "acc_123" || account["username"] != "alice" {
		t.Fatalf("unexpected account payload: %#v", account)
	}
	if tokens["access_token"] != "access.jwt.value" || tokens["refresh_token"] != "refresh.opaque.value" || tokens["token_type"] != "Bearer" {
		t.Fatalf("unexpected tokens payload: %#v", tokens)
	}
	if tokens["expires_in"] != float64(900) {
		t.Fatalf("unexpected expires_in: %#v", tokens["expires_in"])
	}
}

func TestAuthHandlerRefreshReturnsTokenPairOnly(t *testing.T) {
	stub := &authHandlerStub{
		refreshFn: func(ctx context.Context, refreshToken string) (models.TokenPair, error) {
			if refreshToken != "  opaque-refresh-token  " {
				t.Fatalf("refresh token was modified: %q", refreshToken)
			}
			return models.TokenPair{
				AccessToken:  "new-access",
				RefreshToken: "new-refresh",
				TokenType:    "Bearer",
				ExpiresIn:    900,
			}, nil
		},
	}
	handler, err := NewAuthHandler(stub, stub)
	if err != nil {
		t.Fatalf("NewAuthHandler() error = %v", err)
	}

	body := `{"refresh_token":"  opaque-refresh-token  "}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	handler.Refresh(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.refreshCalls != 1 {
		t.Fatalf("unexpected refresh call count: %d", stub.refreshCalls)
	}

	var bodyMap map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &bodyMap); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	data := bodyMap["data"].(map[string]any)
	if data["access_token"] != "new-access" || data["refresh_token"] != "new-refresh" || data["token_type"] != "Bearer" {
		t.Fatalf("unexpected token payload: %#v", data)
	}
	if _, exists := data["account"]; exists {
		t.Fatal("refresh response must not include an account payload")
	}
}

func TestAuthHandlerLogoutRequiresIdentityAndReturnsNoContent(t *testing.T) {
	stub := &authHandlerStub{
		logoutFn: func(ctx context.Context, refreshToken string) error {
			if refreshToken != "  opaque-refresh-token  " {
				t.Fatalf("refresh token was modified: %q", refreshToken)
			}
			return nil
		},
	}
	handler, err := NewAuthHandler(stub, stub)
	if err != nil {
		t.Fatalf("NewAuthHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", strings.NewReader(`{"refresh_token":"  opaque-refresh-token  "}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	rr := httptest.NewRecorder()

	handler.Logout(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("logout should not write a body: %q", rr.Body.String())
	}
	if stub.logoutCalls != 1 {
		t.Fatalf("unexpected logout call count: %d", stub.logoutCalls)
	}
}

func TestAuthHandlerRejectsMalformedBodiesAndUnknownFields(t *testing.T) {
	stub := &authHandlerStub{}
	handler, err := NewAuthHandler(stub, stub)
	if err != nil {
		t.Fatalf("NewAuthHandler() error = %v", err)
	}

	cases := []struct {
		name   string
		method func(http.ResponseWriter, *http.Request)
		body   string
	}{
		{name: "empty body", method: handler.Login, body: ""},
		{name: "unknown field", method: handler.Register, body: `{"username":"alice","password":"secret","unknown":true}`},
		{name: "multiple json values", method: handler.Refresh, body: `{"refresh_token":"one"}{"refresh_token":"two"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/v1/auth", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			tc.method(rr, req)

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
}

func TestAuthHandlerValidationErrorsAreConverted(t *testing.T) {
	stub := &authHandlerStub{}
	handler, err := NewAuthHandler(stub, stub)
	if err != nil {
		t.Fatalf("NewAuthHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(`{"username":"","password":""}`))
	req.Header.Set("Content-Type", "application/json")

	handler.Login(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d", rr.Code)
	}

	var body response.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Error.Code != "VALIDATION_FAILED" {
		t.Fatalf("unexpected error code: %q", body.Error.Code)
	}
	if len(body.Error.Details) != 2 || body.Error.Details[0].Field != "username" || body.Error.Details[1].Field != "password" {
		t.Fatalf("unexpected validation details: %#v", body.Error.Details)
	}
}

func TestAuthHandlerRejectsMissingIdentityOnLogout(t *testing.T) {
	stub := &authHandlerStub{}
	handler, err := NewAuthHandler(stub, stub)
	if err != nil {
		t.Fatalf("NewAuthHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", strings.NewReader(`{"refresh_token":"opaque"}`))
	req.Header.Set("Content-Type", "application/json")

	handler.Logout(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	var body response.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Error.Code != "INVALID_CREDENTIALS" {
		t.Fatalf("unexpected error code: %q", body.Error.Code)
	}
}
