package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicRoutesServeHealthAndDiscovery(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type: %q", got)
	}
	if !strings.Contains(rr.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected health response: %s", rr.Body.String())
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"success":true`) {
		t.Fatalf("unexpected discovery response: %s", rr.Body.String())
	}
}

func TestPublicRoutesServeGeographyZones(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "geography.zone.list") {
		t.Fatalf("expected zone list handler to run: %v", rec.snapshot())
	}

	rec = &routerRecorder{}
	router = newTestRouter(t, rec)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones/north-central", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "geography.zone.get:north-central") {
		t.Fatalf("expected zone detail handler to run: %v", rec.snapshot())
	}
}

func TestPublicRoutesRejectUnsupportedMethods(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Allow"); got == "" {
		t.Fatal("expected Allow header")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodHead, "/v1/geography/geopolitical-zones/north-central", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("unexpected allow header: %q", got)
	}
}
