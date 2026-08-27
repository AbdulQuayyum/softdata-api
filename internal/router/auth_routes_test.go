package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthRoutesWirePublicAndBearerEndpoints(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", strings.NewReader(`{"username":"alice","password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "auth.register") {
		t.Fatalf("register route did not call auth handler: %v", rec.snapshot())
	}

	rec = &routerRecorder{}
	router = newTestRouter(t, rec)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/auth/logout", strings.NewReader(`{"refresh_token":"opaque"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer access-token")
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	got := rec.snapshot()
	if !strings.Contains(strings.Join(got, ","), "authentication") {
		t.Fatalf("logout route did not apply bearer auth middleware: %v", got)
	}
	if !strings.Contains(strings.Join(got, ","), "auth.logout") {
		t.Fatalf("logout route did not call auth handler: %v", got)
	}
}
