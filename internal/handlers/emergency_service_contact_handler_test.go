package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

const emergencyContactRoute = "/v1/emergency/emergency-service-contacts"

func TestEmergencyServiceContactHandlerList(t *testing.T) {
	for _, tc := range []struct {
		name  string
		query string
		want  interfaces.EmergencyServiceContactQuery
	}{
		{"default", "", interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 50}},
		{"explicit", "page=2&page_size=2", interfaces.EmergencyServiceContactQuery{Page: 2, PageSize: 2}},
		{"max", "page_size=100", interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 100}},
		{"service type", "service_type=fire", interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 50, ServiceType: "fire"}},
		{"contact type", "contact_type=short_code", interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 50, ContactType: "short_code"}},
		{"coverage type", "coverage_type=national", interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 50, CoverageType: "national"}},
		{"contact value", "contact_value=112", interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 50, ContactValue: "112"}},
		{"leading zero value", "contact_value=080022556362", interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 50, ContactValue: "080022556362"}},
		{"plus value", "contact_value=%2B2348032003557", interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 50, ContactValue: "+2348032003557"}},
		{"search", "search=fire", interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 50, Search: "fire"}},
		{"empty search omitted", "search=%20", interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 50}},
		{"trimmed combined", "service_type=%20fire%20&contact_type=%20short_code%20&coverage_type=%20national%20&contact_value=%20112%20&search=%20Federal%20Fire%20", interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 50, ServiceType: "fire", ContactType: "short_code", CoverageType: "national", ContactValue: "112", Search: "Federal Fire"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &emergencyContactHandlerServiceStub{list: interfaces.EmergencyServiceContactListResult{Records: sampleEmergencyContacts(), Page: tc.want.Page, PageSize: tc.want.PageSize, Total: 5, TotalPages: 1}}
			handler := mustNewEmergencyContactHandler(t, stub)
			target := emergencyContactRoute
			if tc.query != "" {
				target += "?" + tc.query
			}
			w := httptest.NewRecorder()
			handler.ListEmergencyServiceContacts(w, httptest.NewRequest(http.MethodGet, target, nil))
			if w.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if stub.calls != 1 || stub.query != tc.want {
				t.Fatalf("calls=%d query=%#v want %#v", stub.calls, stub.query, tc.want)
			}
			var body struct {
				Success bool                       `json:"success"`
				Data    []map[string]any           `json:"data"`
				Meta    map[string]json.RawMessage `json:"meta"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if !body.Success || body.Data == nil || string(body.Meta["page_size"]) == "" || string(body.Meta["limit"]) != "" {
				t.Fatalf("bad envelope: %s", w.Body.String())
			}
			assertEmergencyContactResponseFields(t, body.Data[0])
		})
	}
}

func TestEmergencyServiceContactHandlerListValidationErrors(t *testing.T) {
	for _, query := range []string{
		"page=abc", "page=999999999999999999999999999999", "page=0", "page=-1",
		"page_size=0", "page_size=101", "service_type=office", "contact_type=email",
		"coverage_type=lga", "contact_value=abc", "search=" + strings.Repeat("a", 101),
		"availability=24_hours", "page=1&page=2",
	} {
		stub := &emergencyContactHandlerServiceStub{}
		handler := mustNewEmergencyContactHandler(t, stub)
		w := httptest.NewRecorder()
		handler.ListEmergencyServiceContacts(w, httptest.NewRequest(http.MethodGet, emergencyContactRoute+"?"+query, nil))
		if w.Code != http.StatusBadRequest || stub.calls != 0 {
			t.Fatalf("query=%q status=%d calls=%d body=%s", query, w.Code, stub.calls, w.Body.String())
		}
	}
}

func TestEmergencyServiceContactHandlerNilAndEmptyList(t *testing.T) {
	for _, records := range [][]models.EmergencyServiceContact{nil, {}} {
		stub := &emergencyContactHandlerServiceStub{list: interfaces.EmergencyServiceContactListResult{Records: records, Page: 2, PageSize: 50, Total: 5, TotalPages: 1}}
		handler := mustNewEmergencyContactHandler(t, stub)
		w := httptest.NewRecorder()
		handler.ListEmergencyServiceContacts(w, httptest.NewRequest(http.MethodGet, emergencyContactRoute+"?page=2", nil))
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"data":[]`) {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	}
}

func TestEmergencyServiceContactHandlerRepeatedContactValue(t *testing.T) {
	stub := &emergencyContactHandlerServiceStub{list: interfaces.EmergencyServiceContactListResult{Records: sampleEmergencyContacts(), Page: 1, PageSize: 50, Total: 2, TotalPages: 1}}
	handler := mustNewEmergencyContactHandler(t, stub)
	w := httptest.NewRecorder()
	handler.ListEmergencyServiceContacts(w, httptest.NewRequest(http.MethodGet, emergencyContactRoute+"?contact_value=112", nil))
	if w.Code != http.StatusOK || stub.query.ContactValue != "112" {
		t.Fatalf("status=%d query=%#v body=%s", w.Code, stub.query, w.Body.String())
	}
	var body struct {
		Data []models.EmergencyServiceContact `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data) != 2 || body.Data[0].ContactValue != "112" || body.Data[1].ContactValue != "112" || body.Data[0].ID == body.Data[1].ID {
		t.Fatalf("bad repeated result: %#v", body.Data)
	}
}

func TestEmergencyServiceContactHandlerDetailAndErrors(t *testing.T) {
	stub := &emergencyContactHandlerServiceStub{record: sampleEmergencyContacts()[0]}
	handler := mustNewEmergencyContactHandler(t, stub)
	req := httptest.NewRequest(http.MethodGet, emergencyContactRoute+"/federal-fire-service-fire-112-national", nil)
	req.SetPathValue("contact_id", "federal-fire-service-fire-112-national")
	w := httptest.NewRecorder()
	handler.GetEmergencyServiceContact(w, req)
	if w.Code != http.StatusOK || stub.calls != 1 || stub.id != "federal-fire-service-fire-112-national" {
		t.Fatalf("status=%d calls=%d id=%q body=%s", w.Code, stub.calls, stub.id, w.Body.String())
	}
	var body struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	assertEmergencyContactResponseFields(t, body.Data)

	for _, id := range []string{"", "Bad", "bad_id", "bad/id", "bad%2Fid", "../bad", strings.Repeat("a", 256)} {
		stub := &emergencyContactHandlerServiceStub{}
		handler := mustNewEmergencyContactHandler(t, stub)
		req := httptest.NewRequest(http.MethodGet, emergencyContactRoute+"/"+id, nil)
		req.SetPathValue("contact_id", id)
		w := httptest.NewRecorder()
		handler.GetEmergencyServiceContact(w, req)
		if w.Code != http.StatusBadRequest || stub.calls != 0 {
			t.Fatalf("id=%q status=%d calls=%d body=%s", id, w.Code, stub.calls, w.Body.String())
		}
	}
	stub = &emergencyContactHandlerServiceStub{err: services.ErrEmergencyServiceContactNotFound}
	handler = mustNewEmergencyContactHandler(t, stub)
	req = httptest.NewRequest(http.MethodGet, emergencyContactRoute+"/unknown-valid-id", nil)
	req.SetPathValue("contact_id", "unknown-valid-id")
	w = httptest.NewRecorder()
	handler.GetEmergencyServiceContact(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestEmergencyServiceContactHandlerServiceErrorsAndMethods(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want int
	}{
		{errors.New("secret decoder path"), http.StatusInternalServerError},
		{context.Canceled, http.StatusServiceUnavailable},
		{context.DeadlineExceeded, http.StatusServiceUnavailable},
		{services.ErrInvalidEmergencyServiceContactServiceType, http.StatusBadRequest},
	} {
		stub := &emergencyContactHandlerServiceStub{err: tc.err}
		handler := mustNewEmergencyContactHandler(t, stub)
		w := httptest.NewRecorder()
		handler.ListEmergencyServiceContacts(w, httptest.NewRequest(http.MethodGet, emergencyContactRoute, nil))
		if w.Code != tc.want || strings.Contains(w.Body.String(), "secret") || strings.Contains(w.Body.String(), "decoder") {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	}
	stub := &emergencyContactHandlerServiceStub{}
	handler := mustNewEmergencyContactHandler(t, stub)
	w := httptest.NewRecorder()
	handler.ListEmergencyServiceContacts(w, httptest.NewRequest(http.MethodPost, emergencyContactRoute, nil))
	if w.Code != http.StatusMethodNotAllowed || w.Header().Get("Allow") != http.MethodGet || stub.calls != 0 {
		t.Fatalf("status=%d allow=%q body=%s", w.Code, w.Header().Get("Allow"), w.Body.String())
	}
}

func mustNewEmergencyContactHandler(t testing.TB, service emergencyServiceContactService) *EmergencyServiceContactHandler {
	t.Helper()
	handler, err := NewEmergencyServiceContactHandler(service)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func assertEmergencyContactResponseFields(t testing.TB, record map[string]any) {
	t.Helper()
	expected := map[string]bool{"id": true, "service_name": true, "agency_name": true, "service_type": true, "contact_type": true, "contact_value": true, "coverage_type": true, "country_code": true, "availability": true, "call_cost": true, "notes": true}
	for field := range record {
		if !expected[field] {
			t.Fatalf("forbidden field %q in response", field)
		}
	}
	for _, required := range []string{"id", "service_name", "agency_name", "service_type", "contact_type", "contact_value", "coverage_type", "country_code"} {
		if _, ok := record[required]; !ok {
			t.Fatalf("missing field %q in %#v", required, record)
		}
	}
}

type emergencyContactHandlerServiceStub struct {
	query  interfaces.EmergencyServiceContactQuery
	id     string
	list   interfaces.EmergencyServiceContactListResult
	record models.EmergencyServiceContact
	err    error
	calls  int
}

func (s *emergencyContactHandlerServiceStub) ListEmergencyServiceContacts(_ context.Context, query interfaces.EmergencyServiceContactQuery) (interfaces.EmergencyServiceContactListResult, error) {
	s.calls++
	s.query = query
	if s.err != nil {
		return interfaces.EmergencyServiceContactListResult{}, s.err
	}
	return s.list, nil
}

func (s *emergencyContactHandlerServiceStub) GetEmergencyServiceContact(_ context.Context, id string) (models.EmergencyServiceContact, error) {
	s.calls++
	s.id = id
	if s.err != nil {
		return models.EmergencyServiceContact{}, s.err
	}
	return s.record, nil
}

func sampleEmergencyContacts() []models.EmergencyServiceContact {
	return []models.EmergencyServiceContact{
		{ID: "federal-fire-service-fire-112-national", ServiceName: "Federal Fire Service Emergency Response", AgencyName: "Federal Fire Service", ServiceType: "fire", ContactType: "short_code", ContactValue: "112", CoverageType: "national", CountryCode: "NG", Availability: "24_hours", Notes: "Distinct official fire-service context."},
		{ID: "nigerian-communications-commission-general-emergency-112-national", ServiceName: "112 Emergency Number", AgencyName: "Nigerian Communications Commission", ServiceType: "general_emergency", ContactType: "short_code", ContactValue: "112", CoverageType: "national", CountryCode: "NG", Availability: "24_hours", CallCost: "toll_free"},
	}
}
