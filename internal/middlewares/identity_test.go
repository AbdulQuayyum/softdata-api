package middlewares

import (
	"context"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

func TestIdentityAccessorsUseTypedContextKeys(t *testing.T) {
	ctx := context.WithValue(context.Background(), "account_identity", AccountIdentity{AccountID: "acc"})
	if _, ok := AccountIdentityFromContext(ctx); ok {
		t.Fatal("plain string key unexpectedly resolved account identity")
	}

	ctx = context.WithValue(context.Background(), "api_key_identity", services.APIKeyIdentity{APIKeyID: "key", AccountID: "acc"})
	if _, ok := APIKeyIdentityFromContext(ctx); ok {
		t.Fatal("plain string key unexpectedly resolved api key identity")
	}
}

func TestIdentityAccessorsRoundTrip(t *testing.T) {
	accountCtx := WithAccountIdentity(context.Background(), AccountIdentity{AccountID: "acc_123"})
	account, ok := AccountIdentityFromContext(accountCtx)
	if !ok {
		t.Fatal("account identity missing")
	}
	if account.AccountID != "acc_123" {
		t.Fatalf("unexpected account identity: %#v", account)
	}

	apiKeyCtx := WithAPIKeyIdentity(context.Background(), services.APIKeyIdentity{APIKeyID: "key_123", AccountID: "acc_123"})
	apiKey, ok := APIKeyIdentityFromContext(apiKeyCtx)
	if !ok {
		t.Fatal("api key identity missing")
	}
	if apiKey.APIKeyID != "key_123" || apiKey.AccountID != "acc_123" {
		t.Fatalf("unexpected api key identity: %#v", apiKey)
	}
}

func TestIdentityAccessorsDoNotCollideBetweenRequests(t *testing.T) {
	first := WithAccountIdentity(context.Background(), AccountIdentity{AccountID: "acc_1"})
	second := WithAPIKeyIdentity(context.Background(), services.APIKeyIdentity{APIKeyID: "key_1", AccountID: "acc_1"})

	if _, ok := APIKeyIdentityFromContext(first); ok {
		t.Fatal("account context unexpectedly exposed api key identity")
	}
	if _, ok := AccountIdentityFromContext(second); ok {
		t.Fatal("api key context unexpectedly exposed account identity")
	}
}
