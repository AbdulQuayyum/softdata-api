package middlewares

import (
	"errors"
	"fmt"
	"net/http"
)

// NewBodyLimit returns middleware that bounds request bodies for body-bearing methods.
func NewBodyLimit(maxBytes int64) (Middleware, error) {
	if maxBytes <= 0 {
		return nil, errors.New("body limit must be positive")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if methodAllowsBody(r.Method) {
				body := r.Body
				if body == nil {
					body = http.NoBody
				}
				r.Body = http.MaxBytesReader(w, body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

func methodAllowsBody(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

// MaxBytesLimitError describes an invalid body-limit configuration.
type MaxBytesLimitError struct {
	Limit int64
}

func (e MaxBytesLimitError) Error() string {
	return fmt.Sprintf("invalid body limit: %d", e.Limit)
}
