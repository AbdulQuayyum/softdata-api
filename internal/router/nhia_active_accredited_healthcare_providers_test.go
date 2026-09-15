package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/handlers"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

const nhiaHCPRoute = "/v1/healthcare/nhia-active-accredited-healthcare-providers"
const longestNHIAHCPID = "fct-0001-p"

type routerNHIAHCPStub struct {
	rec   *routerRecorder
	query interfaces.NHIAActiveAccreditedHealthcareProviderQuery
	id    string
}

func (s *routerNHIAHCPStub) ListNHIAActiveAccreditedHealthcareProviders(_ context.Context, q interfaces.NHIAActiveAccreditedHealthcareProviderQuery) (interfaces.NHIAActiveAccreditedHealthcareProviderListResult, error) {
	s.query = q
	if s.rec != nil {
		s.rec.add("nhia-hcps:list")
	}
	total := 6536
	records := []models.NHIAActiveAccreditedHealthcareProvider{{ID: "fct-0001-p", Name: "WILDOT CLINIC", CountryCode: "NG", ProviderCode: "FCT/0001/P", FacilityType: "primary", ListingStatus: "active_accredited"}}
	if q.ProviderCode != "" {
		records[0].ProviderCode = q.ProviderCode
		total = 1
	}
	if q.FacilityType != "" {
		records[0].FacilityType = q.FacilityType
		total = 1
	}
	if q.Search == "missing" {
		total = 0
		records = nil
	}
	pages := 0
	if total > 0 {
		pages = (total + q.PageSize - 1) / q.PageSize
	}
	if q.Page > pages {
		records = nil
	}
	return interfaces.NHIAActiveAccreditedHealthcareProviderListResult{Records: records, Page: q.Page, PageSize: q.PageSize, Total: total, TotalPages: pages}, nil
}

func (s *routerNHIAHCPStub) GetNHIAActiveAccreditedHealthcareProvider(_ context.Context, id string) (models.NHIAActiveAccreditedHealthcareProvider, error) {
	s.id = id
	if s.rec != nil {
		s.rec.add("nhia-hcps:detail")
	}
	if id == "unknown-valid-id" {
		return models.NHIAActiveAccreditedHealthcareProvider{}, services.ErrNHIAActiveAccreditedHealthcareProviderNotFound
	}
	return models.NHIAActiveAccreditedHealthcareProvider{ID: id, Name: "Sample provider", CountryCode: "NG", ProviderCode: "FCT/0001/P", FacilityType: "primary", ListingStatus: "active_accredited"}, nil
}

func testNHIAHCPHandler(t *testing.T, s *routerNHIAHCPStub) *handlers.NHIAActiveAccreditedHealthcareProviderHandler {
	t.Helper()
	h, err := handlers.NewNHIAActiveAccreditedHealthcareProviderHandler(s)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestNHIAActiveAccreditedHealthcareProviderProductionRoutes(t *testing.T) {
	for _, tc := range []struct {
		path                                   string
		status                                 int
		page, size, total                      int
		providerCode, facilityType, statusCode string
		search                                 string
		id                                     string
	}{
		{path: "", status: 200, page: 1, size: 50, total: 6536},
		{path: "?page=2&page_size=25", status: 200, page: 2, size: 25, total: 6536},
		{path: "?page_size=100", status: 200, page: 1, size: 100, total: 6536},
		{path: "?provider_code=%20FCT/0001/P%20", status: 200, page: 1, size: 50, total: 1, providerCode: "FCT/0001/P"},
		{path: "?facility_type=primary", status: 200, page: 1, size: 50, total: 1, facilityType: "primary"},
		{path: "?facility_type=primary_and_secondary", status: 200, page: 1, size: 50, total: 1, facilityType: "primary_and_secondary"},
		{path: "?listing_status=active_accredited", status: 200, page: 1, size: 50, total: 6536, statusCode: "active_accredited"},
		{path: "?search=%20missing%20", status: 200, page: 1, size: 50, total: 0, search: "missing"},
		{path: "?provider_code=FCT/0001/P&facility_type=primary&listing_status=active_accredited&search=WILDOT&page=1&page_size=1", status: 200, page: 1, size: 1, total: 1, providerCode: "FCT/0001/P", facilityType: "primary", statusCode: "active_accredited", search: "WILDOT"},
		{path: "?page=6537&page_size=1", status: 200, page: 6537, size: 1, total: 6536},
		{path: "?page=abc", status: 400},
		{path: "?page_size=101", status: 400},
		{path: "?provider_code=../bad", status: 400},
		{path: "?facility_type=secondary", status: 400},
		{path: "?listing_status=licensed", status: 400},
		{path: "?state_id=lagos", status: 400},
		{path: "/fct-0001-p", status: 200, id: "fct-0001-p"},
		{path: "/" + longestNHIAHCPID, status: 200, id: longestNHIAHCPID},
		{path: "/unknown-valid-id", status: 404, id: "unknown-valid-id"},
		{path: "/" + strings.Repeat("a", 256), status: 400},
		{path: "/fct-0001-p/extra", status: 404},
		{path: "/", status: 404},
		{path: "/fct-0001-p/", status: 404},
	} {
		t.Run(tc.path, func(t *testing.T) {
			var rec routerRecorder
			s := &routerNHIAHCPStub{rec: &rec}
			h := testHandlers(t, &rec)
			h.NHIAActiveAccreditedHealthcareProviders = testNHIAHCPHandler(t, s)
			r, err := New(h, testMiddleware(&rec))
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, nhiaHCPRoute+tc.path, nil))
			if w.Code != tc.status || !json.Valid(w.Body.Bytes()) || !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
			if tc.id != "" && s.id != tc.id {
				t.Fatalf("ID %s", s.id)
			}
			if tc.page > 0 {
				want := interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: tc.page, PageSize: tc.size, ProviderCode: tc.providerCode, FacilityType: tc.facilityType, ListingStatus: tc.statusCode, Search: tc.search}
				if s.query != want {
					t.Fatalf("query %+v", s.query)
				}
				var body struct {
					Success bool
					Data    []models.NHIAActiveAccreditedHealthcareProvider
					Meta    map[string]int
				}
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if !body.Success || body.Data == nil || len(body.Meta) != 4 || body.Meta["page"] != tc.page || body.Meta["page_size"] != tc.size || body.Meta["total"] != tc.total {
					t.Fatalf("envelope %s", w.Body.String())
				}
				if (tc.search == "missing" || tc.page == 6537) && len(body.Data) != 0 {
					t.Fatal("expected empty page")
				}
			}
			if tc.status == 200 {
				template := nhiaHCPRoute
				call := "nhia-hcps:list"
				if tc.id != "" {
					template += "/{provider_id}"
					call = "nhia-hcps:detail"
				}
				want := []string{"request_id", "recovery", "logger", "security_headers", "cors", "body_limit", "timeout", "optional_api_key", "rate_limit", "usage:" + template + "|healthcare", call}
				if !reflect.DeepEqual(rec.snapshot(), want) {
					t.Fatalf("middleware %v", rec.snapshot())
				}
			}
		})
	}
}

func TestNHIAActiveAccreditedHealthcareProviderProductionMethodsAndAliases(t *testing.T) {
	router := newTestRouter(t, &routerRecorder{})
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead} {
		for _, path := range []string{nhiaHCPRoute, nhiaHCPRoute + "/fct-0001-p"} {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(method, path, nil))
			if w.Code != 405 || w.Header().Get("Allow") != http.MethodGet || !json.Valid(w.Body.Bytes()) || !strings.Contains(w.Body.String(), `"success":false`) {
				t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
			}
		}
	}
	for _, path := range []string{"/v1/healthcare/hcps", "/v1/healthcare/healthcare-providers", "/v1/healthcare/accredited-healthcare-providers", "/v1/healthcare/active-healthcare-providers", "/v1/healthcare/nhia-providers"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != 404 {
			t.Fatalf("alias route exists: %s", path)
		}
	}
}
