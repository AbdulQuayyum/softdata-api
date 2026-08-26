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
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type apiKeyAuthenticatorStub struct {
	authFn func(context.Context, string) (services.APIKeyIdentity, error)
	last   string
	calls  int
	mu     sync.Mutex
}

func (s *apiKeyAuthenticatorStub) Authenticate(ctx context.Context, plaintext string) (services.APIKeyIdentity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.last = plaintext
	if s.authFn == nil {
		return services.APIKeyIdentity{}, nil
	}
	return s.authFn(ctx, plaintext)
}

func TestOptionalAPIKeySkipsAbsentHeader(t *testing.T) {
	authenticator := &apiKeyAuthenticatorStub{authFn: func(context.Context, string) (services.APIKeyIdentity, error) {
		t.Fatal("authenticator should not be called")
		return services.APIKeyIdentity{}, nil
	}}

	handler := OptionalAPIKey(authenticator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := APIKeyIdentityFromContext(r.Context()); ok {
			t.Fatal("unexpected api key identity in anonymous request")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if authenticator.calls != 0 {
		t.Fatalf("unexpected authenticator call count: %d", authenticator.calls)
	}
}

func TestOptionalAPIKeyAllowsValidKey(t *testing.T) {
	authenticator := &apiKeyAuthenticatorStub{
		authFn: func(_ context.Context, plaintext string) (services.APIKeyIdentity, error) {
			if plaintext != "sd_live_example" {
				t.Fatalf("unexpected plaintext key: %q", plaintext)
			}
			return services.APIKeyIdentity{APIKeyID: "key_123", AccountID: "acc_123"}, nil
		},
	}

	handler := OptionalAPIKey(authenticator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := APIKeyIdentityFromContext(r.Context())
		if !ok {
			t.Fatal("api key identity missing")
		}
		if identity.APIKeyID != "key_123" || identity.AccountID != "acc_123" {
			t.Fatalf("unexpected api key identity: %#v", identity)
		}
		if _, ok := AccountIdentityFromContext(r.Context()); ok {
			t.Fatal("api key middleware should not fabricate bearer identity")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "sd_live_example")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if authenticator.calls != 1 {
		t.Fatalf("unexpected authenticator call count: %d", authenticator.calls)
	}
	if authenticator.last != "sd_live_example" {
		t.Fatalf("unexpected plaintext key received: %q", authenticator.last)
	}
}

func TestOptionalAPIKeyRejectsInvalidKeys(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		present    bool
		wantStatus int
	}{
		{name: "not found", err: services.ErrAPIKeyNotFound, present: true, wantStatus: http.StatusUnauthorized},
		{name: "revoked", err: services.ErrAPIKeyNotFound, present: true, wantStatus: http.StatusUnauthorized},
		{name: "expired", err: services.ErrAPIKeyNotFound, present: true, wantStatus: http.StatusUnauthorized},
		{name: "blank header", err: services.ErrAPIKeyNotFound, present: true, wantStatus: http.StatusUnauthorized},
		{name: "unexpected failure", err: errors.New("revoked"), present: true, wantStatus: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			authenticator := &apiKeyAuthenticatorStub{
				authFn: func(context.Context, string) (services.APIKeyIdentity, error) {
					return services.APIKeyIdentity{}, tc.err
				},
			}
			handler := RequestID(OptionalAPIKey(authenticator)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("next handler should not be called")
			})))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.name == "blank header" {
				req.Header.Set("X-API-Key", "")
			} else if tc.present {
				req.Header.Set("X-API-Key", "sd_live_example")
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("unexpected status: %d", rr.Code)
			}

			var body response.ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if tc.wantStatus == http.StatusUnauthorized && body.Error.Code != "INVALID_API_KEY" {
				t.Fatalf("unexpected error code: %q", body.Error.Code)
			}
			if body.Error.RequestID == "" {
				t.Fatal("request id missing from error response")
			}
		})
	}
}

func TestOptionalAPIKeyPreservesExistingContextValues(t *testing.T) {
	authenticator := &apiKeyAuthenticatorStub{
		authFn: func(_ context.Context, plaintext string) (services.APIKeyIdentity, error) {
			return services.APIKeyIdentity{APIKeyID: "key_123", AccountID: "acc_123"}, nil
		},
	}

	type ctxKey struct{}
	seedCtx := context.WithValue(context.Background(), ctxKey{}, "seed")

	handler := OptionalAPIKey(authenticator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Context().Value(ctxKey{}); got != "seed" {
			t.Fatalf("context value lost: %#v", got)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(seedCtx)
	req.Header.Set("X-API-Key", "sd_live_example")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}
