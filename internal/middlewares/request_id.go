package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
)

const (
	requestIDHeader = "X-Request-ID"
	requestIDPrefix = "req_"
	maxRequestIDLen = 64
)

type middlewareContextKey struct{}

// Middleware is the shared middleware signature used throughout the package.
type Middleware func(http.Handler) http.Handler

// RequestID returns middleware that establishes and propagates a safe request ID.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestIDFromRequest(r)
		ctx := context.WithValue(r.Context(), middlewareContextKey{}, id)
		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext retrieves the request ID when one has been established.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	value, ok := ctx.Value(middlewareContextKey{}).(string)
	if !ok || value == "" {
		return "", false
	}
	return value, true
}

func requestIDFromRequest(r *http.Request) string {
	if r != nil {
		if id, ok := RequestIDFromContext(r.Context()); ok && isSafeRequestID(id) {
			return id
		}
		if id := strings.TrimSpace(r.Header.Get(requestIDHeader)); isSafeRequestID(id) {
			return id
		}
	}
	return newRequestID()
}

func isSafeRequestID(value string) bool {
	if len(value) == 0 || len(value) > maxRequestIDLen {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.', r == '~':
		default:
			return false
		}
	}
	return true
}

func newRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return requestIDPrefix + "fallback"
	}
	return requestIDPrefix + base64.RawURLEncoding.EncodeToString(buf)
}
