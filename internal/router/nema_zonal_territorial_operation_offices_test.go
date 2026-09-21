package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/handlers"
	"github.com/AbdulQuayyum/softdata-api/internal/middlewares"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

const (
	nemaOfficeRoute        = "/v1/emergency/nema-zonal-territorial-operation-offices"
	nemaOfficeRouteFirstID = "nema-abuja-zonal-territorial-operation-office"
)

type routerNEMAOfficeStub struct {
	rec    *routerRecorder
	mu     sync.Mutex
	query  interfaces.NEMAZonalTerritorialOperationOfficeQuery
	id     string
	apiKey bool
}

func (s *routerNEMAOfficeStub) ListNEMAZonalTerritorialOperationOffices(ctx context.Context, q interfaces.NEMAZonalTerritorialOperationOfficeQuery) (interfaces.NEMAZonalTerritorialOperationOfficeListResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.query = q
	if _, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.apiKey = true
	}
	if s.rec != nil {
		s.rec.add("nema:list")
	}
	records := []models.NEMAZonalTerritorialOperationOffice{nemaOfficeRouteFirstRecord()}
	total := 17
	if q.StateID == "missing" || q.Search == "missing" {
		records = nil
		total = 0
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + q.PageSize - 1) / q.PageSize
	}
	if q.Page > totalPages {
		records = nil
	}
	return interfaces.NEMAZonalTerritorialOperationOfficeListResult{Records: records, Page: q.Page, PageSize: q.PageSize, Total: total, TotalPages: totalPages}, nil
}

func (s *routerNEMAOfficeStub) GetNEMAZonalTerritorialOperationOffice(ctx context.Context, id string) (models.NEMAZonalTerritorialOperationOffice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.id = id
	if _, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.apiKey = true
	}
	if s.rec != nil {
		s.rec.add("nema:detail")
	}
	if id == "unknown-valid-id" {
		return models.NEMAZonalTerritorialOperationOffice{}, services.ErrNEMAZonalTerritorialOperationOfficeNotFound
	}
	record := nemaOfficeRouteFirstRecord()
	record.ID = id
	return record, nil
}

func testNEMAOfficeHandler(t *testing.T, s *routerNEMAOfficeStub) *handlers.NEMAZonalTerritorialOperationOfficeHandler {
	t.Helper()
	h, err := handlers.NewNEMAZonalTerritorialOperationOfficeHandler(s)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestNEMAZonalTerritorialOperationOfficeProductionRoutes(t *testing.T) {
	for _, tc := range []struct {
		name       string
		path       string
		status     int
		page       int
		size       int
		total      int
		stateID    string
		officeType string
		search     string
		id         string
	}{
		{name: "default list", status: 200, page: 1, size: 50, total: 17},
		{name: "pagination", path: "?page=2&page_size=5", status: 200, page: 2, size: 5, total: 17},
		{name: "state", path: "?state_id=fct", status: 200, page: 1, size: 50, total: 17, stateID: "fct"},
		{name: "office type", path: "?office_type=zonal_territorial_operation_office", status: 200, page: 1, size: 50, total: 17, officeType: "zonal_territorial_operation_office"},
		{name: "search", path: "?search=%20Abuja%20", status: 200, page: 1, size: 50, total: 17, search: "Abuja"},
		{name: "combined", path: "?state_id=fct&office_type=zonal_territorial_operation_office&search=Abuja&page=1&page_size=1", status: 200, page: 1, size: 1, total: 17, stateID: "fct", officeType: "zonal_territorial_operation_office", search: "Abuja"},
		{name: "empty list", path: "?search=missing", status: 200, page: 1, size: 50, total: 0, search: "missing"},
		{name: "beyond final", path: "?page=18&page_size=1", status: 200, page: 18, size: 1, total: 17},
		{name: "bad page", path: "?page=abc", status: 400},
		{name: "bad size", path: "?page_size=101", status: 400},
		{name: "bad state", path: "?state_id=not-a-state", status: 400},
		{name: "bad office type", path: "?office_type=regional_office", status: 400},
		{name: "detail", path: "/" + nemaOfficeRouteFirstID, status: 200, id: nemaOfficeRouteFirstID},
		{name: "unknown", path: "/unknown-valid-id", status: 404, id: "unknown-valid-id"},
		{name: "too long", path: "/" + strings.Repeat("a", 256), status: 400},
		{name: "extra segment", path: "/" + nemaOfficeRouteFirstID + "/extra", status: 404},
		{name: "trailing slash", path: "/", status: 404},
		{name: "detail trailing slash", path: "/" + nemaOfficeRouteFirstID + "/", status: 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var rec routerRecorder
			stub := &routerNEMAOfficeStub{rec: &rec}
			h := testHandlers(t, &rec)
			h.NEMAZonalTerritorialOperationOffices = testNEMAOfficeHandler(t, stub)
			r, err := New(h, testMiddleware(&rec))
			if err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, nemaOfficeRoute+tc.path, nil)
			req.Header.Set("X-API-Key", "test-key")
			r.ServeHTTP(w, req)
			if w.Code != tc.status || !json.Valid(w.Body.Bytes()) || !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
			if tc.id != "" && stub.id != tc.id {
				t.Fatalf("ID %s", stub.id)
			}
			if tc.page > 0 {
				want := interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: tc.page, PageSize: tc.size, StateID: tc.stateID, OfficeType: tc.officeType, Search: tc.search}
				if stub.query != want {
					t.Fatalf("query %+v", stub.query)
				}
				var body struct {
					Success bool
					Data    []models.NEMAZonalTerritorialOperationOffice
					Meta    map[string]int
				}
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if !body.Success || body.Data == nil || body.Meta["page"] != tc.page || body.Meta["page_size"] != tc.size || body.Meta["total"] != tc.total {
					t.Fatalf("envelope %s", w.Body.String())
				}
				if (tc.search == "missing" || tc.path == "?page=18&page_size=1") && len(body.Data) != 0 {
					t.Fatalf("expected empty list: %s", w.Body.String())
				}
			}
			if tc.status == 200 {
				template := nemaOfficeRoute
				call := "nema:list"
				if tc.id != "" {
					template += "/{office_id}"
					call = "nema:detail"
				}
				want := []string{"request_id", "recovery", "logger", "security_headers", "cors", "body_limit", "timeout", "optional_api_key", "rate_limit", "usage:" + template + "|emergency", call}
				if !reflect.DeepEqual(rec.snapshot(), want) {
					t.Fatalf("middleware %v", rec.snapshot())
				}
				if !stub.apiKey {
					t.Fatal("optional API key identity did not reach handler")
				}
				joined := strings.Join(rec.snapshot(), ",")
				for _, literal := range []string{nemaOfficeRouteFirstID + "|emergency", "fct|emergency", "Abuja|emergency"} {
					if strings.Contains(joined, literal) {
						t.Fatalf("literal value leaked into usage dimensions: %v", rec.snapshot())
					}
				}
			}
		})
	}
}

func TestNEMAZonalTerritorialOperationOfficeProductionMethodsAndAliases(t *testing.T) {
	router := newTestRouter(t, &routerRecorder{})
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead} {
		for _, path := range []string{nemaOfficeRoute, nemaOfficeRoute + "/" + nemaOfficeRouteFirstID} {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(method, path, nil))
			if w.Code != http.StatusMethodNotAllowed || w.Header().Get("Allow") != http.MethodGet || !json.Valid(w.Body.Bytes()) || !strings.Contains(w.Body.String(), `"success":false`) {
				t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
			}
		}
	}
	for _, path := range []string{"/v1/emergency/nema-offices", "/v1/emergency/nema-zonal-offices", "/v1/emergency/nema-zonal-territorial-operation-office", "/v1/emergency/nema-zonal-territorial-operation-offices/" + nemaOfficeRouteFirstID + "/extra"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("alias route exists: %s", path)
		}
	}
}

func TestNEMAZonalTerritorialOperationOfficeHandlerDependencyRequired(t *testing.T) {
	handlers := testHandlers(t, &routerRecorder{})
	handlers.NEMAZonalTerritorialOperationOffices = nil
	if _, err := New(handlers, testMiddleware(&routerRecorder{})); err == nil || !strings.Contains(err.Error(), "nema zonal territorial operation office handler is required") {
		t.Fatalf("New() error = %v", err)
	}
}

func TestNEMAZonalTerritorialOperationOfficeRoutesDoNotChangeEmergencyContacts(t *testing.T) {
	router := newTestRouter(t, &routerRecorder{})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, emergencyServiceContactRoute, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("emergency contacts route changed: %d %s", w.Code, w.Body.String())
	}
}

func nemaOfficeRouteFirstRecord() models.NEMAZonalTerritorialOperationOffice {
	return models.NEMAZonalTerritorialOperationOffice{
		ID:          nemaOfficeRouteFirstID,
		Name:        "NEMA Abuja Office",
		OfficeType:  "zonal_territorial_operation_office",
		StateID:     "fct",
		CountryCode: "NG",
	}
}
