package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

// APIKeyAuthenticator authenticates a plaintext API key and returns safe identity data.
type APIKeyAuthenticator interface {
	Authenticate(ctx context.Context, plaintext string) (services.APIKeyIdentity, error)
}

// OptionalAPIKey returns middleware that identifies API-key traffic when a key is supplied.
func OptionalAPIKey(authenticator APIKeyAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID, _ := RequestIDFromContext(r.Context())
			if authenticator == nil {
				_ = response.Error(w, errors.New("internal server error"), requestID)
				return
			}

			values := r.Header.Values("X-API-Key")
			if len(values) == 0 {
				next.ServeHTTP(w, r)
				return
			}
			if len(values) != 1 {
				_ = response.Error(w, services.ErrAPIKeyNotFound, requestID)
				return
			}

			raw := values[0]

			identity, err := authenticator.Authenticate(r.Context(), raw)
			if err != nil {
				switch {
				case errors.Is(err, services.ErrAPIKeyNotFound):
					_ = response.Error(w, services.ErrAPIKeyNotFound, requestID)
				case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
					_ = response.Error(w, err, requestID)
				default:
					_ = response.Error(w, errors.New("internal server error"), requestID)
				}
				return
			}
			if identity.APIKeyID == "" || identity.AccountID == "" {
				_ = response.Error(w, services.ErrAPIKeyNotFound, requestID)
				return
			}

			ctx := WithAPIKeyIdentity(r.Context(), identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
