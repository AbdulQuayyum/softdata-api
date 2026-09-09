package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicRoutesServeHealthcareContract(t *testing.T) {
	var rec routerRecorder
	router := newTestRouter(t, &rec)

	tests := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/v1/healthcare/health-facilities", http.StatusOK},
		{http.MethodGet, "/v1/healthcare/health-facilities/sample-health-facility", http.StatusOK},
		{http.MethodGet, "/v1/healthcare/health-facilities?state_id=%20lagos%20&page_size=100", http.StatusOK},
		{http.MethodGet, "/v1/healthcare/health-facilities/sample-health-facility/extra", http.StatusNotFound},
	}
	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != tc.want {
			t.Fatalf("%s %s status = %d, want %d; body=%s", tc.method, tc.path, rr.Code, tc.want, rr.Body.String())
		}
	}

	joined := strings.Join(rec.snapshot(), "|")
	if !strings.Contains(joined, "usage:/v1/healthcare/health-facilities|healthcare") {
		t.Fatalf("healthcare list usage template missing: %s", joined)
	}
	if !strings.Contains(joined, "usage:/v1/healthcare/health-facilities/{facility_id}|healthcare") {
		t.Fatalf("healthcare detail usage template missing: %s", joined)
	}
	if strings.Contains(joined, "sample-health-facility|healthcare") {
		t.Fatalf("literal facility ID leaked into usage template: %s", joined)
	}
}

func TestPublicRoutesRejectUnsupportedHealthcareMethodsWithJSON405(t *testing.T) {
	router := newTestRouter(t, nil)
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		for _, path := range []string{
			"/v1/healthcare/health-facilities",
			"/v1/healthcare/health-facilities/sample-health-facility",
		} {
			req := httptest.NewRequest(method, path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			if rr.Code != http.StatusMethodNotAllowed || rr.Header().Get("Allow") != http.MethodGet {
				t.Fatalf("%s %s status=%d allow=%q", method, path, rr.Code, rr.Header().Get("Allow"))
			}
			if !strings.Contains(rr.Header().Get("Content-Type"), "application/json") || strings.Contains(rr.Body.String(), "Method Not Allowed") {
				t.Fatalf("%s %s did not return standard JSON 405: %s", method, path, rr.Body.String())
			}
		}
	}
}
