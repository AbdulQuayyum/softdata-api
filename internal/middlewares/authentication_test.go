package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/security"
	"github.com/golang-jwt/jwt/v5"
)

type tokenVerifierStub struct {
	verifyFn func(string) (*security.AccessTokenClaims, error)
	last     string
	calls    int
	mu       sync.Mutex
}

func (s *tokenVerifierStub) ValidateAccessToken(token string) (*security.AccessTokenClaims, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.last = token
	if s.verifyFn == nil {
		return nil, nil
	}
	return s.verifyFn(token)
}

func TestAuthenticationAllowsValidBearerToken(t *testing.T) {
	verifier := &tokenVerifierStub{
		verifyFn: func(token string) (*security.AccessTokenClaims, error) {
			if token != "access.jwt.value" {
				t.Fatalf("unexpected token: %q", token)
			}
			return &security.AccessTokenClaims{
				TokenType: "access",
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "acc_123",
					ID:      "jti_123",
				},
			}, nil
		},
	}

	var gotAccount AccountIdentity
	handler := Authentication(verifier)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		gotAccount, ok = AccountIdentityFromContext(r.Context())
		if !ok {
			t.Fatal("account identity missing from context")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer access.jwt.value")
	rr := httptest.NewRecorder()

	RequestID(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if gotAccount.AccountID != "acc_123" {
		t.Fatalf("unexpected account id: %#v", gotAccount)
	}
	if verifier.calls != 1 {
		t.Fatalf("unexpected verifier call count: %d", verifier.calls)
	}
	if verifier.last != "access.jwt.value" {
		t.Fatalf("unexpected verifier input: %q", verifier.last)
	}
}

func TestAuthenticationRejectsMalformedBearerHeaders(t *testing.T) {
	cases := []struct {
		name    string
		auth    string
		present bool
	}{
		{name: "missing", auth: "", present: false},
		{name: "blank header", auth: "", present: true},
		{name: "wrong scheme", auth: "Basic abc", present: true},
		{name: "empty token", auth: "Bearer", present: true},
		{name: "ambiguous spacing", auth: "Bearer   ", present: true},
		{name: "extra fields", auth: "Bearer token extra", present: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			verifier := &tokenVerifierStub{verifyFn: func(string) (*security.AccessTokenClaims, error) {
				t.Fatal("verifier should not be called")
				return nil, nil
			}}
			handler := RequestID(Authentication(verifier)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("next handler should not be called")
			})))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.present {
				req.Header.Set("Authorization", tc.auth)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			var body response.ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if body.Error.RequestID == "" {
				t.Fatal("request id missing from error response")
			}
		})
	}
}

func TestAuthenticationRejectsWrongTokenTypeAndSanitizesVerifierErrors(t *testing.T) {
	cases := []struct {
		name   string
		claims *security.AccessTokenClaims
		err    error
	}{
		{
			name: "refresh token",
			claims: &security.AccessTokenClaims{
				TokenType: "refresh",
			},
		},
		{
			name: "invalid signature",
			err:  errors.New("validate access token: signature invalid"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			verifier := &tokenVerifierStub{
				verifyFn: func(string) (*security.AccessTokenClaims, error) {
					if tc.err != nil {
						return nil, tc.err
					}
					return &security.AccessTokenClaims{
						TokenType: tc.claims.TokenType,
						RegisteredClaims: jwt.RegisteredClaims{
							Subject: "acc_123",
							ID:      "jti_123",
						},
					}, nil
				},
			}
			handler := RequestID(Authentication(verifier)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("next handler should not be called")
			})))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer access.jwt.value")
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

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
			if body.Error.RequestID == "" {
				t.Fatal("request id missing from error response")
			}
		})
	}
}

func TestAuthenticationPreservesExistingContextValues(t *testing.T) {
	verifier := &tokenVerifierStub{
		verifyFn: func(string) (*security.AccessTokenClaims, error) {
			return &security.AccessTokenClaims{
				TokenType:        "access",
				RegisteredClaims: jwt.RegisteredClaims{Subject: "acc_123", ID: "jti_123"},
			}, nil
		},
	}

	type ctxKey struct{}
	seedCtx := context.WithValue(context.Background(), ctxKey{}, "seed")

	handler := Authentication(verifier)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Context().Value(ctxKey{}); got != "seed" {
			t.Fatalf("context value lost: %#v", got)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(seedCtx)
	req.Header.Set("Authorization", "Bearer access.jwt.value")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}
