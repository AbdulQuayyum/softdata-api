package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAccountRoutesWireApiKeysAndUsageEndpoints(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v1/account/api-keys/key_123", nil)
	req.Header.Set("Authorization", "Bearer access-token")
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	got := strings.Join(rec.snapshot(), ",")
	for _, want := range []string{"optional_api_key", "authentication", "rate_limit", "usage:/v1/account/api-keys/{key_id}|", "apikey.delete:key_123"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in middleware sequence: %v", want, rec.snapshot())
		}
	}

	rec = &routerRecorder{}
	router = newTestRouter(t, rec)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/account/usage/history", nil)
	req.Header.Set("Authorization", "Bearer access-token")
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	got = strings.Join(rec.snapshot(), ",")
	if !strings.Contains(got, "usage:/v1/account/usage/history|") {
		t.Fatalf("usage route did not receive expected endpoint metadata: %v", rec.snapshot())
	}
}
