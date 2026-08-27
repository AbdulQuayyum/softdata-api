package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestRequestIDGeneratesAndStoresContextValue(t *testing.T) {
	var gotHeader string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := RequestIDFromContext(r.Context())
		if !ok {
			t.Fatal("request id missing from context")
		}
		gotHeader = w.Header().Get(requestIDHeader)
		if id != gotHeader {
			t.Fatalf("context id and header differ: %q vs %q", id, gotHeader)
		}
		if !isSafeRequestID(id) {
			t.Fatalf("generated request id is not safe: %q", id)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if gotHeader == "" {
		t.Fatal("request id header was not written")
	}
}

func TestRequestIDPreservesSafeIncomingID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(requestIDHeader, "req_123-abc")
	rr := httptest.NewRecorder()

	RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := RequestIDFromContext(r.Context())
		if !ok || id != "req_123-abc" {
			t.Fatalf("unexpected request id in context: %q %v", id, ok)
		}
	})).ServeHTTP(rr, req)

	if got := rr.Header().Get(requestIDHeader); got != "req_123-abc" {
		t.Fatalf("unexpected request id header: %q", got)
	}
}

func TestRequestIDReplacesUnsafeValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(requestIDHeader, "bad id with spaces")
	rr := httptest.NewRecorder()

	RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := RequestIDFromContext(r.Context())
		if id == "bad id with spaces" {
			t.Fatal("unsafe request id was preserved")
		}
		if !strings.HasPrefix(id, requestIDPrefix) {
			t.Fatalf("generated id missing prefix: %q", id)
		}
	})).ServeHTTP(rr, req)
}

func TestRequestIDRejectsOverlongValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(requestIDHeader, strings.Repeat("a", maxRequestIDLen+1))
	rr := httptest.NewRecorder()

	RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := RequestIDFromContext(r.Context())
		if len(id) > maxRequestIDLen {
			t.Fatalf("overlong request id was preserved: %q", id)
		}
	})).ServeHTTP(rr, req)
}

func TestRequestIDUsesTypedContextKey(t *testing.T) {
	ctx := context.WithValue(context.Background(), "request_id", "string-key")
	if _, ok := RequestIDFromContext(ctx); ok {
		t.Fatal("plain string context key unexpectedly resolved")
	}
}

func TestRequestIDConcurrentRequestsAreIndependent(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	const requests = 32
	var wg sync.WaitGroup
	ids := make(chan string, requests)

	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			ids <- rr.Header().Get(requestIDHeader)
		}()
	}

	wg.Wait()
	close(ids)

	seen := make(map[string]struct{}, requests)
	for id := range ids {
		if id == "" {
			t.Fatal("generated empty request id")
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate request id generated: %q", id)
		}
		seen[id] = struct{}{}
	}
}
