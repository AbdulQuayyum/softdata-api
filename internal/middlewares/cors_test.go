package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSAllowsConfiguredOriginsAndPreflight(t *testing.T) {
	mw, err := NewCORS(CORSOptions{
		AllowedOrigins:   []string{"https://example.com"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-API-Key", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Total-Count"},
		AllowCredentials: true,
	})
	if err != nil {
		t.Fatalf("NewCORS() error = %v", err)
	}

	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler was not called for allowed origin")
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Fatalf("unexpected allow-origin header: %q", rr.Header().Get("Access-Control-Allow-Origin"))
	}
	if rr.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("missing credentials header")
	}

	preflight := httptest.NewRequest(http.MethodOptions, "/", nil)
	preflight.Header.Set("Origin", "https://example.com")
	preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
	preflight.Header.Set("Access-Control-Request-Headers", "Authorization, X-API-Key")
	preflightRR := httptest.NewRecorder()
	called = false
	handler.ServeHTTP(preflightRR, preflight)

	if called {
		t.Fatal("preflight reached downstream handler")
	}
	if preflightRR.Code != http.StatusNoContent {
		t.Fatalf("unexpected preflight status: %d", preflightRR.Code)
	}
	if preflightRR.Header().Get("Access-Control-Allow-Methods") != "GET, POST" {
		t.Fatalf("unexpected allow-methods header: %q", preflightRR.Header().Get("Access-Control-Allow-Methods"))
	}
	if preflightRR.Header().Get("Access-Control-Allow-Headers") != "Authorization, Content-Type, X-API-Key, X-Request-ID" {
		t.Fatalf("unexpected allow-headers header: %q", preflightRR.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestCORSRejectsUnconfiguredOriginsWithoutCallingHandler(t *testing.T) {
	mw, err := NewCORS(CORSOptions{AllowedOrigins: []string{"https://example.com"}})
	if err != nil {
		t.Fatalf("NewCORS() error = %v", err)
	}

	called := false
	handler := mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com.attacker.test")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("non-CORS request should still reach the handler")
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("disallowed origin received permissive CORS headers")
	}
}

func TestCORSNoOriginPassesThrough(t *testing.T) {
	mw, err := NewCORS(CORSOptions{AllowedOrigins: []string{"https://example.com"}})
	if err != nil {
		t.Fatalf("NewCORS() error = %v", err)
	}

	called := false
	handler := mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("request without origin did not reach handler")
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("no-origin request received CORS headers")
	}
}

func TestCORSRejectsInvalidPreflight(t *testing.T) {
	mw, err := NewCORS(CORSOptions{AllowedOrigins: []string{"https://example.com"}})
	if err != nil {
		t.Fatalf("NewCORS() error = %v", err)
	}

	called := false
	handler := mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodTrace)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if called {
		t.Fatal("invalid preflight reached downstream handler")
	}
	if rr.Code != http.StatusForbidden {
		t.Fatalf("unexpected invalid-preflight status: %d", rr.Code)
	}
}

func TestCORSRejectsWildcardCredentialsConfiguration(t *testing.T) {
	if _, err := NewCORS(CORSOptions{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	}); err == nil {
		t.Fatal("NewCORS() error = nil, want error")
	}
}
