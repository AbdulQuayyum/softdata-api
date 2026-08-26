package middlewares

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBodyLimitAllowsRequestsWithinLimit(t *testing.T) {
	mw, err := NewBodyLimit(8)
	if err != nil {
		t.Fatalf("NewBodyLimit() error = %v", err)
	}

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if string(data) != "12345678" {
			t.Fatalf("unexpected body: %q", data)
		}
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("12345678"))
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestBodyLimitReportsOversizedBodies(t *testing.T) {
	mw, err := NewBodyLimit(4)
	if err != nil {
		t.Fatalf("NewBodyLimit() error = %v", err)
	}

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err == nil {
			t.Fatal("expected MaxBytesError")
		}
		var maxErr *http.MaxBytesError
		if !errors.As(err, &maxErr) {
			t.Fatalf("unexpected read error: %v", err)
		}
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("12345"))
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestBodyLimitSkipsMethodsWithoutBodies(t *testing.T) {
	mw, err := NewBodyLimit(4)
	if err != nil {
		t.Fatalf("NewBodyLimit() error = %v", err)
	}

	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Body == nil {
			t.Fatal("body should remain readable or safe")
		}
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if !called {
		t.Fatal("handler not called")
	}
}

func TestBodyLimitRejectsInvalidConfiguration(t *testing.T) {
	if _, err := NewBodyLimit(0); err == nil {
		t.Fatal("NewBodyLimit() error = nil, want error")
	}
}
