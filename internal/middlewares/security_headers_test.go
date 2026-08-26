package middlewares

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSecurityHeadersApplyDocumentedDefaults(t *testing.T) {
	mw, err := NewSecurityHeaders(SecurityHeadersOptions{})
	if err != nil {
		t.Fatalf("NewSecurityHeaders() error = %v", err)
	}

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Handler", "ok")
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing X-Content-Type-Options header")
	}
	if rr.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatal("missing X-Frame-Options header")
	}
	if rr.Header().Get("Referrer-Policy") != "no-referrer" {
		t.Fatal("missing Referrer-Policy header")
	}
	if rr.Header().Get("X-Handler") != "ok" {
		t.Fatal("downstream headers not preserved")
	}
	if got := rr.Header().Get("Strict-Transport-Security"); got != "" {
		t.Fatalf("unexpected HSTS header: %q", got)
	}
}

func TestSecurityHeadersEnableHSTSOnlyOnTLS(t *testing.T) {
	mw, err := NewSecurityHeaders(SecurityHeadersOptions{
		EnableHSTS:            true,
		HSTSMaxAge:            365 * 24 * time.Hour,
		HSTSIncludeSubdomains: true,
		HSTSPreload:           true,
	})
	if err != nil {
		t.Fatalf("NewSecurityHeaders() error = %v", err)
	}

	handler := mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	plain := httptest.NewRecorder()
	handler.ServeHTTP(plain, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := plain.Header().Get("Strict-Transport-Security"); got != "" {
		t.Fatalf("unexpected HSTS for plain HTTP: %q", got)
	}

	tlsReq := httptest.NewRequest(http.MethodGet, "/", nil)
	tlsReq.TLS = &tls.ConnectionState{}
	tlsResp := httptest.NewRecorder()
	handler.ServeHTTP(tlsResp, tlsReq)
	if got := tlsResp.Header().Get("Strict-Transport-Security"); got != "max-age=31536000; includeSubDomains; preload" {
		t.Fatalf("unexpected HSTS value: %q", got)
	}
}

func TestSecurityHeadersRejectInvalidHSTSConfiguration(t *testing.T) {
	if _, err := NewSecurityHeaders(SecurityHeadersOptions{EnableHSTS: true, HSTSMaxAge: 0}); err == nil {
		t.Fatal("NewSecurityHeaders() error = nil, want error")
	}
}
