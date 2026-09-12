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

const nhiaHMORoute = "/v1/healthcare/nhia-accredited-health-maintenance-organisations"
const longestNHIAHMOID = "sterling-health-managed-care-services-limited-34"

type routerNHIAHMOStub struct {
	rec   *routerRecorder
	query interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery
	id    string
}

func (s *routerNHIAHMOStub) ListNHIAAccreditedHealthMaintenanceOrganisations(_ context.Context, q interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery) (interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult, error) {
	s.query = q
	if s.rec != nil {
		s.rec.add("nhia-hmos:list")
	}
	total := 94
	records := []models.NHIAAccreditedHealthMaintenanceOrganisation{{ID: "a-and-m-healthcare-trust-limited-102", Name: "A&M HEALTHCARE TRUST LIMITED", CountryCode: "NG", OrganisationType: "health_maintenance_organisation", AccreditationStatus: "accredited", HMOID: "102"}}
	if q.HMOID != "" {
		records[0].HMOID = q.HMOID
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
	return interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{Records: records, Page: q.Page, PageSize: q.PageSize, Total: total, TotalPages: pages}, nil
}

func (s *routerNHIAHMOStub) GetNHIAAccreditedHealthMaintenanceOrganisation(_ context.Context, id string) (models.NHIAAccreditedHealthMaintenanceOrganisation, error) {
	s.id = id
	if s.rec != nil {
		s.rec.add("nhia-hmos:detail")
	}
	if id == "unknown-valid-id" {
		return models.NHIAAccreditedHealthMaintenanceOrganisation{}, services.ErrNHIAAccreditedHealthMaintenanceOrganisationNotFound
	}
	return models.NHIAAccreditedHealthMaintenanceOrganisation{ID: id, Name: "Sample HMO", CountryCode: "NG", OrganisationType: "health_maintenance_organisation", AccreditationStatus: "accredited", HMOID: "102"}, nil
}

func testNHIAHMOHandler(t *testing.T, s *routerNHIAHMOStub) *handlers.NHIAAccreditedHealthMaintenanceOrganisationHandler {
	t.Helper()
	h, err := handlers.NewNHIAAccreditedHealthMaintenanceOrganisationHandler(s)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestNHIAAccreditedHMOProductionRoutes(t *testing.T) {
	for _, tc := range []struct {
		path                        string
		status                      int
		page, size, total           int
		filterStatus, hmoID, search string
		id                          string
	}{
		{path: "", status: 200, page: 1, size: 50, total: 94},
		{path: "?page=2&page_size=25", status: 200, page: 2, size: 25, total: 94},
		{path: "?page_size=100", status: 200, page: 1, size: 100, total: 94},
		{path: "?accreditation_status=accredited", status: 200, page: 1, size: 50, total: 94, filterStatus: "accredited"},
		{path: "?hmo_id=%200012%20", status: 200, page: 1, size: 50, total: 1, hmoID: "0012"},
		{path: "?search=%20missing%20", status: 200, page: 1, size: 50, total: 0, search: "missing"},
		{path: "?accreditation_status=accredited&hmo_id=102&search=A%26M&page=1&page_size=1", status: 200, page: 1, size: 1, total: 1, filterStatus: "accredited", hmoID: "102", search: "A&M"},
		{path: "?page=95&page_size=1", status: 200, page: 95, size: 1, total: 94},
		{path: "?page_size=101", status: 400},
		{path: "?accreditation_status=expired", status: 400},
		{path: "?state_id=lagos", status: 400},
		{path: "/a-and-m-healthcare-trust-limited-102", status: 200, id: "a-and-m-healthcare-trust-limited-102"},
		{path: "/" + longestNHIAHMOID, status: 200, id: longestNHIAHMOID},
		{path: "/unknown-valid-id", status: 404, id: "unknown-valid-id"},
		{path: "/" + strings.Repeat("a", 256), status: 400},
		{path: "/a-and-m-healthcare-trust-limited-102/extra", status: 404},
		{path: "/", status: 404},
		{path: "/a-and-m-healthcare-trust-limited-102/", status: 404},
	} {
		t.Run(tc.path, func(t *testing.T) {
			var rec routerRecorder
			s := &routerNHIAHMOStub{rec: &rec}
			h := testHandlers(t, &rec)
			h.NHIAAccreditedHMOs = testNHIAHMOHandler(t, s)
			r, err := New(h, testMiddleware(&rec))
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, nhiaHMORoute+tc.path, nil))
			if w.Code != tc.status || !json.Valid(w.Body.Bytes()) || !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
			if tc.id != "" && s.id != tc.id {
				t.Fatalf("ID %s", s.id)
			}
			if tc.page > 0 {
				want := interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Page: tc.page, PageSize: tc.size, AccreditationStatus: tc.filterStatus, HMOID: tc.hmoID, Search: tc.search}
				if s.query != want {
					t.Fatalf("query %+v", s.query)
				}
				var body struct {
					Success bool
					Data    []models.NHIAAccreditedHealthMaintenanceOrganisation
					Meta    map[string]int
				}
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if !body.Success || body.Data == nil || len(body.Meta) != 4 || body.Meta["page"] != tc.page || body.Meta["page_size"] != tc.size || body.Meta["total"] != tc.total {
					t.Fatalf("envelope %s", w.Body.String())
				}
				if (tc.search == "missing" || tc.page == 95) && len(body.Data) != 0 {
					t.Fatal("expected empty page")
				}
			}
			if tc.status == 200 {
				template := nhiaHMORoute
				call := "nhia-hmos:list"
				if tc.id != "" {
					template += "/{organisation_id}"
					call = "nhia-hmos:detail"
				}
				want := []string{"request_id", "recovery", "logger", "security_headers", "cors", "body_limit", "timeout", "optional_api_key", "rate_limit", "usage:" + template + "|healthcare", call}
				if !reflect.DeepEqual(rec.snapshot(), want) {
					t.Fatalf("middleware %v", rec.snapshot())
				}
			}
		})
	}
}

func TestNHIAAccreditedHMOProductionMethodsAndAliases(t *testing.T) {
	router := newTestRouter(t, &routerRecorder{})
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead} {
		for _, path := range []string{nhiaHMORoute, nhiaHMORoute + "/a-and-m-healthcare-trust-limited-102"} {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(method, path, nil))
			if w.Code != 405 || w.Header().Get("Allow") != http.MethodGet || !json.Valid(w.Body.Bytes()) || !strings.Contains(w.Body.String(), `"success":false`) {
				t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
			}
		}
	}
	for _, path := range []string{"/v1/healthcare/hmos", "/v1/healthcare/health-maintenance-organizations", "/v1/healthcare/licensed-hmos", "/v1/healthcare/registered-hmos"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != 404 {
			t.Fatalf("alias route exists: %s", path)
		}
	}
}
