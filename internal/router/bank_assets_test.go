package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublicRoutesServeCommercialBankLogo(t *testing.T) {
	router := newTestRouter(t, &routerRecorder{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/assets/banks/ng/access-bank.png", nil)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("content type = %q, want image/png", got)
	}
	if got := rr.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("cache control = %q", got)
	}
	if len(rr.Body.Bytes()) == 0 {
		t.Fatal("logo response is empty")
	}
}

func TestPublicRoutesRejectInvalidCommercialBankLogoPaths(t *testing.T) {
	router := newTestRouter(t, &routerRecorder{})
	cases := []struct {
		name   string
		method string
		target string
		status int
	}{
		{name: "unknown", method: http.MethodGet, target: "/v1/assets/banks/ng/not-a-bank.png", status: http.StatusNotFound},
		{name: "unsupported extension", method: http.MethodGet, target: "/v1/assets/banks/ng/access-bank.svg", status: http.StatusUnprocessableEntity},
		{name: "uppercase", method: http.MethodGet, target: "/v1/assets/banks/ng/Access-Bank.png", status: http.StatusUnprocessableEntity},
		{name: "method", method: http.MethodHead, target: "/v1/assets/banks/ng/access-bank.png", status: http.StatusMethodNotAllowed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.target, nil)
			router.ServeHTTP(rr, req)
			if rr.Code != tc.status {
				t.Fatalf("status = %d, want %d", rr.Code, tc.status)
			}
		})
	}
}
