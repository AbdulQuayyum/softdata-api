package middlewares

import (
	"context"

	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type accountIdentityContextKey struct{}
type apiKeyIdentityContextKey struct{}

// AccountIdentity stores the minimum safe account identity required by bearer-authenticated routes.
type AccountIdentity struct {
	AccountID string
}

// WithAccountIdentity stores a safe account identity in the provided context.
func WithAccountIdentity(ctx context.Context, identity AccountIdentity) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, accountIdentityContextKey{}, identity)
}

// AccountIdentityFromContext retrieves a safe account identity from context.
func AccountIdentityFromContext(ctx context.Context) (AccountIdentity, bool) {
	if ctx == nil {
		return AccountIdentity{}, false
	}
	identity, ok := ctx.Value(accountIdentityContextKey{}).(AccountIdentity)
	if !ok || identity.AccountID == "" {
		return AccountIdentity{}, false
	}
	return identity, true
}

// WithAPIKeyIdentity stores the safe API-key identity in the provided context.
func WithAPIKeyIdentity(ctx context.Context, identity services.APIKeyIdentity) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, apiKeyIdentityContextKey{}, identity)
}

// APIKeyIdentityFromContext retrieves a safe API-key identity from context.
func APIKeyIdentityFromContext(ctx context.Context) (services.APIKeyIdentity, bool) {
	if ctx == nil {
		return services.APIKeyIdentity{}, false
	}
	identity, ok := ctx.Value(apiKeyIdentityContextKey{}).(services.APIKeyIdentity)
	if !ok || identity.APIKeyID == "" || identity.AccountID == "" {
		return services.APIKeyIdentity{}, false
	}
	return identity, true
}
