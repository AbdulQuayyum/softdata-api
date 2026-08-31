package router

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/datasets/assets"
)

func TestPublicRoutesServeCountryFlagAssetOutsideWorkingDirectory(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := performFlagRequest(t, router, http.MethodGet, "/v1/assets/flags/ng.svg")
	if rr.Code != 200 {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "image/svg+xml" {
		t.Fatalf("unexpected content type: %q", got)
	}
	if got := rr.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("unexpected cache-control: %q", got)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "usage:/v1/assets/flags/{country_id}.svg|geography") {
		t.Fatalf("expected flag asset usage middleware to run: %v", rec.snapshot())
	}

	want, err := assets.FlagSVG("ng")
	if err != nil {
		t.Fatalf("FlagSVG(ng) error = %v", err)
	}
	if !bytes.Equal(rr.Body.Bytes(), want) {
		t.Fatal("response body did not match embedded Nigeria flag")
	}

	rr = performFlagRequest(t, router, http.MethodGet, "/v1/assets/flags/zz.svg")
	if rr.Code != 404 {
		t.Fatalf("unexpected unknown-id status: %d", rr.Code)
	}

	rr = performFlagRequest(t, router, http.MethodGet, "/v1/assets/flags/NG.svg")
	if rr.Code != 422 {
		t.Fatalf("unexpected uppercase-id status: %d", rr.Code)
	}

	rr = performFlagRequest(t, router, http.MethodGet, "/v1/assets/flags/ng.svg/extra")
	if rr.Code != 404 {
		t.Fatalf("unexpected nested-path status: %d", rr.Code)
	}

	rr = performFlagRequest(t, router, http.MethodHead, "/v1/assets/flags/ng.svg")
	if rr.Code != 405 {
		t.Fatalf("unexpected HEAD status: %d", rr.Code)
	}
	if got := rr.Header().Get("Allow"); got != "GET" {
		t.Fatalf("unexpected allow header: %q", got)
	}
}

func performFlagRequest(t *testing.T, router http.Handler, method, target string) *httptest.ResponseRecorder {
	t.Helper()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, nil)
	router.ServeHTTP(rr, req)
	return rr
}

func TestRouteCatalogMatchesOnlyFlagAssetSuffixRoute(t *testing.T) {
	var catalog routeCatalog
	if err := catalog.add("GET /v1/assets/flags/{country_id}.svg"); err != nil {
		t.Fatalf("catalog.add() error = %v", err)
	}
	if err := catalog.add("GET /v1/geography/countries/{country_id}"); err != nil {
		t.Fatalf("catalog.add() error = %v", err)
	}

	if !catalog.supports("GET", "/v1/assets/flags/ng.svg") {
		t.Fatal("expected catalog to support flag asset path")
	}
	if catalog.supports("GET", "/v1/assets/flags/ng.svg/extra") {
		t.Fatal("catalog overmatched nested flag path")
	}
	if allow := catalog.allow("/v1/assets/flags/ng.svg"); len(allow) != 1 || allow[0] != "GET" {
		t.Fatalf("unexpected allow header set: %#v", allow)
	}
	if catalog.supports("GET", "/v1/assets/flags") {
		t.Fatal("catalog should not match truncated flag path")
	}
}
