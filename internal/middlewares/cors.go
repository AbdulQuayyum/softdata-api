package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// CORSOptions configures immutable CORS behavior.
type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
}

// NewCORS returns middleware that enforces exact-origin CORS policy.
func NewCORS(opts CORSOptions) (Middleware, error) {
	origins := make([]string, 0, len(opts.AllowedOrigins))
	for _, origin := range opts.AllowedOrigins {
		normalized, err := normalizeOrigin(origin)
		if err != nil {
			return nil, err
		}
		if normalized == "*" {
			return nil, errors.New("wildcard origins are not supported")
		}
		origins = append(origins, normalized)
	}
	if opts.AllowCredentials {
		for _, origin := range origins {
			if origin == "*" {
				return nil, errors.New("credentials cannot be used with wildcard origins")
			}
		}
	}

	methods := normalizeTokens(opts.AllowedMethods)
	if len(methods) == 0 {
		methods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions}
	}
	headers := normalizeTokens(opts.AllowedHeaders)
	if len(headers) == 0 {
		headers = []string{"Accept", "Authorization", "Content-Type", "X-API-Key", "X-Request-ID"}
	}
	exposed := normalizeTokens(opts.ExposedHeaders)

	allowedOriginSet := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		allowedOriginSet[origin] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			normalizedOrigin, err := normalizeOrigin(origin)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if _, ok := allowedOriginSet[normalizedOrigin]; !ok {
				next.ServeHTTP(w, r)
				return
			}

			allowHeaders := strings.Join(headers, ", ")
			allowMethods := strings.Join(methods, ", ")
			exposeHeaders := strings.Join(exposed, ", ")

			h := w.Header()
			h.Add("Vary", "Origin")
			h.Set("Access-Control-Allow-Origin", normalizedOrigin)
			if opts.AllowCredentials {
				h.Set("Access-Control-Allow-Credentials", "true")
			}
			if exposeHeaders != "" {
				h.Set("Access-Control-Expose-Headers", exposeHeaders)
			}

			if r.Method != http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			requestMethod := strings.TrimSpace(r.Header.Get("Access-Control-Request-Method"))
			if requestMethod == "" || !tokenInList(requestMethod, methods) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			requestedHeaders := parseHeaderList(r.Header.Get("Access-Control-Request-Headers"))
			for _, requested := range requestedHeaders {
				if !headerInList(requested, headers) {
					http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
					return
				}
			}

			h.Set("Access-Control-Allow-Methods", allowMethods)
			if allowHeaders != "" {
				h.Set("Access-Control-Allow-Headers", allowHeaders)
			}
			w.WriteHeader(http.StatusNoContent)
		})
	}, nil
}

func normalizeOrigin(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", errors.New("origin is required")
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("invalid origin %q", value)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid origin %q", value)
	}
	if parsed.User != nil || parsed.Path != "" && parsed.Path != "/" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("invalid origin %q", value)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("invalid origin %q", value)
	}

	return strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host), nil
}

func normalizeTokens(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}
	return normalized
}

func parseHeaderList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	headers := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		headers = append(headers, part)
	}
	return headers
}

func tokenInList(value string, list []string) bool {
	for _, item := range list {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}

func headerInList(value string, list []string) bool {
	for _, item := range list {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}
