package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscoveryHandlerReturnsExactApiInfoResponse(t *testing.T) {
	h := NewDiscoveryHandler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1", nil)

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type: %q", got)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body["success"] != true {
		t.Fatalf("unexpected success value: %#v", body["success"])
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data was not an object: %#v", body["data"])
	}
	if len(data) != 0 {
		t.Fatalf("expected empty discovery data, got %#v", data)
	}
	if _, exists := body["status"]; exists {
		t.Fatal("discovery response must not include health fields")
	}
}

func TestDiscoveryHandlerIsDeterministicAndImmutable(t *testing.T) {
	h := NewDiscoveryHandler()

	first := httptest.NewRecorder()
	h.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/v1", nil))

	second := httptest.NewRecorder()
	h.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/v1", nil))

	if first.Body.String() != second.Body.String() {
		t.Fatalf("discovery response changed between requests:\nfirst:  %s\nsecond: %s", first.Body.String(), second.Body.String())
	}
}

func TestDiscoveryHandlerRejectsNonGet(t *testing.T) {
	h := NewDiscoveryHandler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1", nil)

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("unexpected allow header: %q", got)
	}
}
