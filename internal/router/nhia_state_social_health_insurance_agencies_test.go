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

const nhiaSSHIARoute = "/v1/healthcare/nhia-state-social-health-insurance-agencies"

type routerNHIASSHIAStub struct {
	rec   *routerRecorder
	query interfaces.NHIAStateSocialHealthInsuranceAgencyQuery
	id    string
}

func (s *routerNHIASSHIAStub) ListNHIAStateSocialHealthInsuranceAgencies(_ context.Context, q interfaces.NHIAStateSocialHealthInsuranceAgencyQuery) (interfaces.NHIAStateSocialHealthInsuranceAgencyListResult, error) {
	s.query = q
	if s.rec != nil {
		s.rec.add("nhia-sshias:list")
	}
	total := 37
	records := []models.NHIAStateSocialHealthInsuranceAgency{{ID: "abia-state-health-insurance-agency-abshia", Name: "Abia State Health Insurance Agency (ABSHIA)", StateID: "abia", CountryCode: "NG", OrganisationType: "state_social_health_insurance_agency"}}
	if q.StateID != "" {
		records[0].StateID = q.StateID
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
	return interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{Records: records, Page: q.Page, PageSize: q.PageSize, Total: total, TotalPages: pages}, nil
}

func (s *routerNHIASSHIAStub) GetNHIAStateSocialHealthInsuranceAgency(_ context.Context, id string) (models.NHIAStateSocialHealthInsuranceAgency, error) {
	s.id = id
	if s.rec != nil {
		s.rec.add("nhia-sshias:detail")
	}
	if id == "unknown-valid-id" {
		return models.NHIAStateSocialHealthInsuranceAgency{}, services.ErrNHIAStateSocialHealthInsuranceAgencyNotFound
	}
	return models.NHIAStateSocialHealthInsuranceAgency{ID: id, Name: "Sample SSHIA", StateID: "abia", CountryCode: "NG", OrganisationType: "state_social_health_insurance_agency"}, nil
}

func testNHIASSHIAHandler(t *testing.T, s *routerNHIASSHIAStub) *handlers.NHIAStateSocialHealthInsuranceAgencyHandler {
	t.Helper()
	h, err := handlers.NewNHIAStateSocialHealthInsuranceAgencyHandler(s)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestNHIAStateSocialHealthInsuranceAgencyProductionRoutes(t *testing.T) {
	for _, tc := range []struct {
		path              string
		status            int
		page, size, total int
		stateID, search   string
		id                string
	}{
		{path: "", status: 200, page: 1, size: 50, total: 37},
		{path: "?page=2&page_size=25", status: 200, page: 2, size: 25, total: 37},
		{path: "?page_size=100", status: 200, page: 1, size: 100, total: 37},
		{path: "?state_id=%20lagos%20", status: 200, page: 1, size: 50, total: 1, stateID: "lagos"},
		{path: "?state_id=fct", status: 200, page: 1, size: 50, total: 1, stateID: "fct"},
		{path: "?search=%20missing%20", status: 200, page: 1, size: 50, total: 0, search: "missing"},
		{path: "?state_id=lagos&search=LASHMA&page=1&page_size=1", status: 200, page: 1, size: 1, total: 1, stateID: "lagos", search: "LASHMA"},
		{path: "?page=38&page_size=1", status: 200, page: 38, size: 1, total: 37},
		{path: "?page_size=101", status: 400},
		{path: "?state_id=AKS", status: 400},
		{path: "?accreditation_status=accredited", status: 400},
		{path: "/abia-state-health-insurance-agency-abshia", status: 200, id: "abia-state-health-insurance-agency-abshia"},
		{path: "/unknown-valid-id", status: 404, id: "unknown-valid-id"},
		{path: "/" + strings.Repeat("a", 256), status: 400},
		{path: "/abia-state-health-insurance-agency-abshia/extra", status: 404},
		{path: "/", status: 404},
		{path: "/abia-state-health-insurance-agency-abshia/", status: 404},
	} {
		t.Run(tc.path, func(t *testing.T) {
			var rec routerRecorder
			s := &routerNHIASSHIAStub{rec: &rec}
			h := testHandlers(t, &rec)
			h.NHIAStateSocialHealthInsuranceAgencies = testNHIASSHIAHandler(t, s)
			r, err := New(h, testMiddleware(&rec))
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, nhiaSSHIARoute+tc.path, nil))
			if w.Code != tc.status || !json.Valid(w.Body.Bytes()) || !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
			if tc.id != "" && s.id != tc.id {
				t.Fatalf("ID %s", s.id)
			}
			if tc.page > 0 {
				want := interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: tc.page, PageSize: tc.size, StateID: tc.stateID, Search: tc.search}
				if s.query != want {
					t.Fatalf("query %+v", s.query)
				}
				var body struct {
					Success bool
					Data    []models.NHIAStateSocialHealthInsuranceAgency
					Meta    map[string]int
				}
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if !body.Success || body.Data == nil || len(body.Meta) != 4 || body.Meta["page"] != tc.page || body.Meta["page_size"] != tc.size || body.Meta["total"] != tc.total {
					t.Fatalf("envelope %s", w.Body.String())
				}
				if (tc.search == "missing" || tc.page == 38) && len(body.Data) != 0 {
					t.Fatal("expected empty page")
				}
			}
			if tc.status == 200 {
				template := nhiaSSHIARoute
				call := "nhia-sshias:list"
				if tc.id != "" {
					template += "/{agency_id}"
					call = "nhia-sshias:detail"
				}
				want := []string{"request_id", "recovery", "logger", "security_headers", "cors", "body_limit", "timeout", "optional_api_key", "rate_limit", "usage:" + template + "|healthcare", call}
				if !reflect.DeepEqual(rec.snapshot(), want) {
					t.Fatalf("middleware %v", rec.snapshot())
				}
			}
		})
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyProductionMethodsAndAliases(t *testing.T) {
	router := newTestRouter(t, &routerRecorder{})
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead} {
		for _, path := range []string{nhiaSSHIARoute, nhiaSSHIARoute + "/abia-state-health-insurance-agency-abshia"} {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(method, path, nil))
			if w.Code != 405 || w.Header().Get("Allow") != http.MethodGet || !json.Valid(w.Body.Bytes()) || !strings.Contains(w.Body.String(), `"success":false`) {
				t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
			}
		}
	}
	for _, path := range []string{"/v1/healthcare/sshias", "/v1/healthcare/state-health-insurance", "/v1/healthcare/state-health-insurance-agencies", "/v1/healthcare/accredited-sshias", "/v1/healthcare/licensed-sshias"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != 404 {
			t.Fatalf("alias route exists: %s", path)
		}
	}
}
