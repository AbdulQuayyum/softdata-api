package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/security"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

// TokenVerifier validates bearer access tokens without exposing parser details.
type TokenVerifier interface {
	ValidateAccessToken(token string) (*security.AccessTokenClaims, error)
}

type tokenVerifierFunc func(string) (*security.AccessTokenClaims, error)

func (f tokenVerifierFunc) ValidateAccessToken(token string) (*security.AccessTokenClaims, error) {
	return f(token)
}

// Authentication returns middleware that requires a valid bearer access token.
func Authentication(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID, _ := RequestIDFromContext(r.Context())
			if verifier == nil {
				_ = response.Error(w, errors.New("internal server error"), requestID)
				return
			}

			token, ok := bearerTokenFromRequest(r)
			if !ok {
				_ = response.Error(w, services.ErrInvalidCredentials, requestID)
				return
			}

			claims, err := verifier.ValidateAccessToken(token)
			if err != nil {
				switch {
				case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
					_ = response.Error(w, err, requestID)
				default:
					_ = response.Error(w, services.ErrInvalidCredentials, requestID)
				}
				return
			}
			if claims == nil || strings.TrimSpace(claims.Subject) == "" || strings.TrimSpace(claims.TokenType) != "access" {
				_ = response.Error(w, services.ErrInvalidCredentials, requestID)
				return
			}

			ctx := WithAccountIdentity(r.Context(), AccountIdentity{AccountID: claims.Subject})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerTokenFromRequest(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	values := r.Header.Values("Authorization")
	if len(values) != 1 {
		return "", false
	}

	raw := strings.TrimSpace(values[0])
	if raw == "" || strings.Contains(raw, ",") {
		return "", false
	}

	parts := strings.Fields(raw)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	if strings.ContainsAny(parts[1], " \t\r\n") {
		return "", false
	}
	return parts[1], true
}
