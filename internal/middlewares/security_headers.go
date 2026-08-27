package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SecurityHeadersOptions configures the immutable security-header middleware.
type SecurityHeadersOptions struct {
	EnableHSTS            bool
	HSTSMaxAge            time.Duration
	HSTSIncludeSubdomains bool
	HSTSPreload           bool
	ContentSecurityPolicy string
	PermissionsPolicy     string
	ReferrerPolicy        string
}

// NewSecurityHeaders returns middleware that applies safe security headers.
func NewSecurityHeaders(opts SecurityHeadersOptions) (Middleware, error) {
	if opts.EnableHSTS && opts.HSTSMaxAge <= 0 {
		return nil, errors.New("hsts max age must be positive")
	}
	if strings.TrimSpace(opts.ReferrerPolicy) == "" {
		opts.ReferrerPolicy = "no-referrer"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			setIfEmpty(h, "X-Content-Type-Options", "nosniff")
			setIfEmpty(h, "X-Frame-Options", "DENY")
			setIfEmpty(h, "Referrer-Policy", opts.ReferrerPolicy)
			if csp := strings.TrimSpace(opts.ContentSecurityPolicy); csp != "" {
				setIfEmpty(h, "Content-Security-Policy", csp)
			}
			if policy := strings.TrimSpace(opts.PermissionsPolicy); policy != "" {
				setIfEmpty(h, "Permissions-Policy", policy)
			}
			if opts.EnableHSTS && r.TLS != nil {
				value := buildHSTSValue(opts)
				if value != "" {
					setIfEmpty(h, "Strict-Transport-Security", value)
				}
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

func setIfEmpty(header http.Header, key, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	if existing := strings.TrimSpace(header.Get(key)); existing == "" {
		header.Set(key, value)
	}
}

func buildHSTSValue(opts SecurityHeadersOptions) string {
	value := fmt.Sprintf("max-age=%d", int64(opts.HSTSMaxAge/time.Second))
	if opts.HSTSIncludeSubdomains {
		value += "; includeSubDomains"
	}
	if opts.HSTSPreload {
		value += "; preload"
	}
	return value
}
