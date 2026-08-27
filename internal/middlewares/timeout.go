package middlewares

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// NewTimeout returns middleware that applies a bounded request deadline.
func NewTimeout(timeout time.Duration) (Middleware, error) {
	if timeout <= 0 {
		return nil, errors.New("timeout must be positive")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}, nil
}
