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

const accreditationRoute = "/v1/healthcare/medical-laboratory-accreditations"
const longestAccreditationID = "onchocerciasis-and-soil-transmitted-helminth-oncho-sth-laboratory-public-health-and-epidemiology-department-nigeria-institute-of-medical-research-lagos-yaba-ml0025"

type routerAccreditationStub struct {
	rec   *routerRecorder
	query interfaces.MedicalLaboratoryAccreditationQuery
	id    string
}

func (s *routerAccreditationStub) ListMedicalLaboratoryAccreditations(_ context.Context, q interfaces.MedicalLaboratoryAccreditationQuery) (interfaces.MedicalLaboratoryAccreditationListResult, error) {
	s.query = q
	s.rec.add("accreditations:list")
	total := 30
	if q.AccreditationStatus == "accredited" {
		total = 26
	}
	if q.AccreditationStatus == "expired" {
		total = 4
	}
	if q.StateID != "" {
		total = 2
	}
	status := q.AccreditationStatus
	if status == "" {
		status = "accredited"
	}
	records := []models.MedicalLaboratoryAccreditation{{ID: "sample-laboratory", Name: "Sample", StateID: "lagos", CountryCode: "NG", AccreditationStatus: status}}
	if q.Search == "missing" {
		total = 0
		records = nil
	}
	pages := (total + q.PageSize - 1) / q.PageSize
	if q.Page > pages {
		records = nil
	}
	return interfaces.MedicalLaboratoryAccreditationListResult{Records: records, Page: q.Page, PageSize: q.PageSize, Total: total, TotalPages: pages}, nil
}
func (s *routerAccreditationStub) GetMedicalLaboratoryAccreditation(_ context.Context, id string) (models.MedicalLaboratoryAccreditation, error) {
	s.id = id
	s.rec.add("accreditations:detail")
	if id == "unknown-valid-id" {
		return models.MedicalLaboratoryAccreditation{}, services.ErrMedicalLaboratoryAccreditationNotFound
	}
	return models.MedicalLaboratoryAccreditation{ID: id, Name: "Sample", CountryCode: "NG", StateID: "lagos", AccreditationStatus: "expired"}, nil
}
func testAccreditationHandler(t *testing.T, s *routerAccreditationStub) *handlers.MedicalLaboratoryAccreditationHandler {
	t.Helper()
	h, err := handlers.NewMedicalLaboratoryAccreditationHandler(s)
	if err != nil {
		t.Fatal(err)
	}
	return h
}
func TestAccreditationProductionRoutes(t *testing.T) {
	for _, tc := range []struct {
		path                      string
		status                    int
		page, size, total         int
		state, filter, search, id string
	}{
		{path: "", status: 200, page: 1, size: 50, total: 30},
		{path: "?page=2&page_size=1&state_id=%20lagos%20&search=%20Lab%20", status: 200, page: 2, size: 1, total: 2, state: "lagos", search: "Lab"},
		{path: "?accreditation_status=accredited", status: 200, page: 1, size: 50, total: 26, filter: "accredited"},
		{path: "?accreditation_status=expired", status: 200, page: 1, size: 50, total: 4, filter: "expired"},
		{path: "?page_size=100", status: 200, page: 1, size: 100, total: 30},
		{path: "?page_size=101", status: 400}, {path: "?accreditation_status=licensed", status: 400},
		{path: "?search=missing", status: 200, page: 1, size: 50, search: "missing"},
		{path: "?page=31&page_size=1", status: 200, page: 31, size: 1, total: 30},
		{path: "/sample-laboratory", status: 200, id: "sample-laboratory"},
		{path: "/" + longestAccreditationID, status: 200, id: longestAccreditationID},
		{path: "/unknown-valid-id", status: 404, id: "unknown-valid-id"},
		{path: "/" + strings.Repeat("a", 256), status: 400},
		{path: "/sample-laboratory/extra", status: 404}, {path: "/", status: 404}, {path: "/sample-laboratory/", status: 404},
	} {
		t.Run(tc.path, func(t *testing.T) {
			var rec routerRecorder
			s := &routerAccreditationStub{rec: &rec}
			h := testHandlers(t, &rec)
			h.MedicalLaboratoryAccreditations = testAccreditationHandler(t, s)
			r, err := New(h, testMiddleware(&rec))
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("GET", accreditationRoute+tc.path, nil))
			if w.Code != tc.status || !json.Valid(w.Body.Bytes()) || !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
			if tc.id != "" && s.id != tc.id {
				t.Fatalf("ID %s", s.id)
			}
			if tc.page > 0 {
				want := interfaces.MedicalLaboratoryAccreditationQuery{Page: tc.page, PageSize: tc.size, StateID: tc.state, AccreditationStatus: tc.filter, Search: tc.search}
				if s.query != want {
					t.Fatalf("query %+v", s.query)
				}
				var body struct {
					Success bool
					Data    []models.MedicalLaboratoryAccreditation
					Meta    map[string]int
				}
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if !body.Success || body.Data == nil || len(body.Meta) != 4 || body.Meta["page"] != tc.page || body.Meta["page_size"] != tc.size || body.Meta["total"] != tc.total {
					t.Fatalf("envelope %s", w.Body.String())
				}
				if (tc.search == "missing" || tc.page == 31) && len(body.Data) != 0 {
					t.Fatal("expected empty page")
				}
				if tc.filter == "expired" && body.Data[0].AccreditationStatus != "expired" {
					t.Fatal("expired hidden")
				}
			}
			if tc.status == 200 {
				template := accreditationRoute
				call := "accreditations:list"
				if tc.id != "" {
					template += "/{accreditation_id}"
					call = "accreditations:detail"
				}
				want := []string{"request_id", "recovery", "logger", "security_headers", "cors", "body_limit", "timeout", "optional_api_key", "rate_limit", "usage:" + template + "|healthcare", call}
				if !reflect.DeepEqual(rec.snapshot(), want) {
					t.Fatalf("middleware %v", rec.snapshot())
				}
			}
		})
	}
}
func TestAccreditationProductionMethodsAndCompatibility(t *testing.T) {
	var rec routerRecorder
	r := newTestRouter(t, &rec)
	for _, method := range []string{"POST", "PUT", "PATCH", "DELETE", "HEAD"} {
		for _, path := range []string{accreditationRoute, accreditationRoute + "/sample-laboratory"} {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(method, path, nil))
			if w.Code != 405 || w.Header().Get("Allow") != "GET" || !json.Valid(w.Body.Bytes()) || !strings.Contains(w.Body.String(), `"success":false`) {
				t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
			}
		}
	}
	for _, path := range []string{"/v1/healthcare/licensed-medical-laboratories", "/v1/healthcare/licensed-medical-laboratories/sample"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 404 {
			t.Fatal("old route exists")
		}
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/v1/healthcare/health-facilities", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"limit":50`) || strings.Contains(w.Body.String(), `"page_size"`) {
		t.Fatalf("older envelope changed: %s", w.Body.String())
	}
}
