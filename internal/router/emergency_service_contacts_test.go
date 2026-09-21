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

const emergencyServiceContactRoute = "/v1/emergency/emergency-service-contacts"
const emergencyRouteFirstContactID = "federal-fire-service-fire-112-national"

type routerEmergencyServiceContactStub struct {
	rec    *routerRecorder
	mu     sync.Mutex
	query  interfaces.EmergencyServiceContactQuery
	id     string
	apiKey bool
}

func (s *routerEmergencyServiceContactStub) ListEmergencyServiceContacts(ctx context.Context, q interfaces.EmergencyServiceContactQuery) (interfaces.EmergencyServiceContactListResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.query = q
	if _, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.apiKey = true
	}
	if s.rec != nil {
		s.rec.add("emergency:list")
	}
	records := []models.EmergencyServiceContact{emergencyRouteFirstContact()}
	total := 5
	if q.ContactValue == "112" {
		records = []models.EmergencyServiceContact{emergencyRouteFirstContact(), emergencyRouteLastContact()}
		total = 2
	}
	if q.Search == "missing" {
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
	return interfaces.EmergencyServiceContactListResult{Records: records, Page: q.Page, PageSize: q.PageSize, Total: total, TotalPages: totalPages}, nil
}

func (s *routerEmergencyServiceContactStub) GetEmergencyServiceContact(ctx context.Context, id string) (models.EmergencyServiceContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.id = id
	if _, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.apiKey = true
	}
	if s.rec != nil {
		s.rec.add("emergency:detail")
	}
	if id == "unknown-valid-id" {
		return models.EmergencyServiceContact{}, services.ErrEmergencyServiceContactNotFound
	}
	record := emergencyRouteFirstContact()
	record.ID = id
	return record, nil
}

func testEmergencyServiceContactHandler(t *testing.T, s *routerEmergencyServiceContactStub) *handlers.EmergencyServiceContactHandler {
	t.Helper()
	h, err := handlers.NewEmergencyServiceContactHandler(s)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestEmergencyServiceContactProductionRoutes(t *testing.T) {
	for _, tc := range []struct {
		name                string
		path                string
		status              int
		page, size, total   int
		serviceType         string
		contactType         string
		coverageType        string
		contactValue        string
		search              string
		id                  string
		repeatedContactSize int
	}{
		{name: "default list", status: 200, page: 1, size: 50, total: 5},
		{name: "pagination", path: "?page=2&page_size=5", status: 200, page: 2, size: 5, total: 5},
		{name: "page size max", path: "?page_size=100", status: 200, page: 1, size: 100, total: 5},
		{name: "service type", path: "?service_type=fire", status: 200, page: 1, size: 50, total: 5, serviceType: "fire"},
		{name: "contact type", path: "?contact_type=short_code", status: 200, page: 1, size: 50, total: 5, contactType: "short_code"},
		{name: "coverage type", path: "?coverage_type=national", status: 200, page: 1, size: 50, total: 5, coverageType: "national"},
		{name: "contact value", path: "?contact_value=%20112%20&page_size=2", status: 200, page: 1, size: 2, total: 2, contactValue: "112", repeatedContactSize: 2},
		{name: "search", path: "?search=%20missing%20", status: 200, page: 1, size: 50, total: 0, search: "missing"},
		{name: "combined filters", path: "?service_type=fire&contact_type=short_code&coverage_type=national&contact_value=112&search=Fire&page=1&page_size=2", status: 200, page: 1, size: 2, total: 2, serviceType: "fire", contactType: "short_code", coverageType: "national", contactValue: "112", search: "Fire", repeatedContactSize: 2},
		{name: "beyond final", path: "?page=6&page_size=1", status: 200, page: 6, size: 1, total: 5},
		{name: "bad page", path: "?page=abc", status: 400},
		{name: "bad size", path: "?page_size=101", status: 400},
		{name: "bad service type", path: "?service_type=office", status: 400},
		{name: "bad contact type", path: "?contact_type=email", status: 400},
		{name: "bad coverage type", path: "?coverage_type=lga", status: 400},
		{name: "bad contact value", path: "?contact_value=abc", status: 400},
		{name: "unsupported state filter", path: "?state_id=lagos", status: 400},
		{name: "detail", path: "/" + emergencyRouteFirstContactID, status: 200, id: emergencyRouteFirstContactID},
		{name: "unknown", path: "/unknown-valid-id", status: 404, id: "unknown-valid-id"},
		{name: "too long", path: "/" + strings.Repeat("a", 256), status: 400},
		{name: "extra segment", path: "/" + emergencyRouteFirstContactID + "/extra", status: 404},
		{name: "trailing slash", path: "/", status: 404},
		{name: "detail trailing slash", path: "/" + emergencyRouteFirstContactID + "/", status: 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var rec routerRecorder
			stub := &routerEmergencyServiceContactStub{rec: &rec}
			h := testHandlers(t, &rec)
			h.EmergencyServiceContacts = testEmergencyServiceContactHandler(t, stub)
			r, err := New(h, testMiddleware(&rec))
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, emergencyServiceContactRoute+tc.path, nil)
			req.Header.Set("X-API-Key", "test-key")
			r.ServeHTTP(w, req)
			if w.Code != tc.status || !json.Valid(w.Body.Bytes()) || !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
			if tc.id != "" && stub.id != tc.id {
				t.Fatalf("ID %s", stub.id)
			}
			if tc.page > 0 {
				want := interfaces.EmergencyServiceContactQuery{Page: tc.page, PageSize: tc.size, ServiceType: tc.serviceType, ContactType: tc.contactType, CoverageType: tc.coverageType, ContactValue: tc.contactValue, Search: tc.search}
				if stub.query != want {
					t.Fatalf("query %+v", stub.query)
				}
				var body struct {
					Success bool
					Data    []models.EmergencyServiceContact
					Meta    map[string]int
				}
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if !body.Success || body.Data == nil || body.Meta["page"] != tc.page || body.Meta["page_size"] != tc.size || body.Meta["total"] != tc.total {
					t.Fatalf("envelope %s", w.Body.String())
				}
				if tc.repeatedContactSize > 0 && len(body.Data) != tc.repeatedContactSize {
					t.Fatalf("repeated contact rows = %d", len(body.Data))
				}
				if (tc.search == "missing" || tc.path == "?page=6&page_size=1") && len(body.Data) != 0 {
					t.Fatalf("expected empty list: %s", w.Body.String())
				}
			}
			if tc.status == 200 {
				template := emergencyServiceContactRoute
				call := "emergency:list"
				if tc.id != "" {
					template += "/{contact_id}"
					call = "emergency:detail"
				}
				want := []string{"request_id", "recovery", "logger", "security_headers", "cors", "body_limit", "timeout", "optional_api_key", "rate_limit", "usage:" + template + "|emergency", call}
				if !reflect.DeepEqual(rec.snapshot(), want) {
					t.Fatalf("middleware %v", rec.snapshot())
				}
				if !stub.apiKey {
					t.Fatal("optional API key identity did not reach handler")
				}
				if strings.Contains(strings.Join(rec.snapshot(), ","), emergencyRouteFirstContactID+"|emergency") {
					t.Fatalf("literal contact ID leaked into usage dimensions: %v", rec.snapshot())
				}
			}
		})
	}
}

func TestEmergencyServiceContactProductionMethodsAndAliases(t *testing.T) {
	router := newTestRouter(t, &routerRecorder{})
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead} {
		for _, path := range []string{emergencyServiceContactRoute, emergencyServiceContactRoute + "/" + emergencyRouteFirstContactID} {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(method, path, nil))
			if w.Code != http.StatusMethodNotAllowed || w.Header().Get("Allow") != http.MethodGet || !json.Valid(w.Body.Bytes()) || !strings.Contains(w.Body.String(), `"success":false`) {
				t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
			}
		}
	}
	for _, path := range []string{"/v1/emergency/contacts", "/v1/emergency/service-contacts", "/v1/emergency/nema-offices", "/v1/emergency/fire-service-commands", "/v1/emergency/ambulance-services"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("alias route exists: %s", path)
		}
	}
}

func emergencyRouteFirstContact() models.EmergencyServiceContact {
	return models.EmergencyServiceContact{
		ID:           emergencyRouteFirstContactID,
		ServiceName:  "Federal Fire Service Emergency Response",
		AgencyName:   "Federal Fire Service",
		ServiceType:  "fire",
		ContactType:  "short_code",
		ContactValue: "112",
		CoverageType: "national",
		CountryCode:  "NG",
	}
}

func emergencyRouteLastContact() models.EmergencyServiceContact {
	return models.EmergencyServiceContact{
		ID:           "nigerian-communications-commission-general-emergency-112-national",
		ServiceName:  "112 Emergency Number",
		AgencyName:   "Nigerian Communications Commission",
		ServiceType:  "general_emergency",
		ContactType:  "short_code",
		ContactValue: "112",
		CoverageType: "national",
		CountryCode:  "NG",
	}
}
