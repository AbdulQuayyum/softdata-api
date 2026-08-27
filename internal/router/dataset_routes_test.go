package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDatasetRoutesWirePublicMetadataEndpoints(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/datasets/ng-states", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	got := strings.Join(rec.snapshot(), ",")
	for _, want := range []string{"optional_api_key", "rate_limit", "usage:/v1/datasets/{dataset_id}|", "dataset.get:ng-states"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in middleware sequence: %v", want, rec.snapshot())
		}
	}

	rec = &routerRecorder{}
	router = newTestRouter(t, rec)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/datasets/ng-states/sources", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	got = strings.Join(rec.snapshot(), ",")
	if !strings.Contains(got, "dataset.sources:ng-states") {
		t.Fatalf("dataset sources route did not use the dataset_id path value: %v", rec.snapshot())
	}
}
