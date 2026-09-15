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

const nhiaHCPRoute = "/v1/healthcare/nhia-active-accredited-healthcare-providers"

func TestNHIAActiveAccreditedHealthcareProviderHandlerList(t *testing.T) {
	for _, tc := range []struct {
		name  string
		query string
		want  interfaces.NHIAActiveAccreditedHealthcareProviderQuery
	}{
		{"default", "", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 50}},
		{"explicit", "page=2&page_size=10", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 2, PageSize: 10}},
		{"max", "page_size=100", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 100}},
		{"provider code", "provider_code=FCT/0001/P", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 50, ProviderCode: "FCT/0001/P"}},
		{"trimmed provider code", "provider_code=%20FCT/0001/P%20", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 50, ProviderCode: "FCT/0001/P"}},
		{"facility primary", "facility_type=primary", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 50, FacilityType: "primary"}},
		{"facility primary secondary", "facility_type=primary_and_secondary", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 50, FacilityType: "primary_and_secondary"}},
		{"listing status", "listing_status=active_accredited", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 50, ListingStatus: "active_accredited"}},
		{"search", "search=clinic", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 50, Search: "clinic"}},
		{"trimmed search", "search=%20WILDOT%20", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 50, Search: "WILDOT"}},
		{"combined", "provider_code=FCT/0001/P&facility_type=primary&listing_status=active_accredited&search=clinic", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 50, ProviderCode: "FCT/0001/P", FacilityType: "primary", ListingStatus: "active_accredited", Search: "clinic"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &nhiaHCPHandlerServiceStub{list: interfaces.NHIAActiveAccreditedHealthcareProviderListResult{Records: []models.NHIAActiveAccreditedHealthcareProvider{sampleNHIAHCP()}, Page: tc.want.Page, PageSize: tc.want.PageSize, Total: 6536, TotalPages: 131}}
			handler := mustNewNHIAHCPHandler(t, stub)
			target := nhiaHCPRoute
			if tc.query != "" {
				target += "?" + tc.query
			}
			w := httptest.NewRecorder()
			handler.ListNHIAActiveAccreditedHealthcareProviders(w, httptest.NewRequest(http.MethodGet, target, nil))
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
			assertNHIAHCPResponseFields(t, body.Data[0])
		})
	}
}

func TestNHIAActiveAccreditedHealthcareProviderHandlerListValidationErrors(t *testing.T) {
	for _, query := range []string{
		"page=abc", "page=999999999999999999999999999999", "page=0", "page=-1",
		"page_size=0", "page_size=101", "provider_code=FCT/1/P", "provider_code=fct/0001/p",
		"facility_type=secondary", "listing_status=licensed", "search=%20", "search=" + strings.Repeat("a", 101),
		"address=true", "page=1&page=2",
	} {
		stub := &nhiaHCPHandlerServiceStub{}
		handler := mustNewNHIAHCPHandler(t, stub)
		w := httptest.NewRecorder()
		handler.ListNHIAActiveAccreditedHealthcareProviders(w, httptest.NewRequest(http.MethodGet, nhiaHCPRoute+"?"+query, nil))
		if w.Code != http.StatusBadRequest || stub.calls != 0 {
			t.Fatalf("query=%q status=%d calls=%d body=%s", query, w.Code, stub.calls, w.Body.String())
		}
	}
}

func TestNHIAActiveAccreditedHealthcareProviderHandlerNilAndBeyondFinalList(t *testing.T) {
	for _, records := range [][]models.NHIAActiveAccreditedHealthcareProvider{nil, {}} {
		stub := &nhiaHCPHandlerServiceStub{list: interfaces.NHIAActiveAccreditedHealthcareProviderListResult{Records: records, Page: 132, PageSize: 50, Total: 6536, TotalPages: 131}}
		handler := mustNewNHIAHCPHandler(t, stub)
		w := httptest.NewRecorder()
		handler.ListNHIAActiveAccreditedHealthcareProviders(w, httptest.NewRequest(http.MethodGet, nhiaHCPRoute+"?page=132", nil))
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"data":[]`) {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	}
}

func TestNHIAActiveAccreditedHealthcareProviderHandlerDetail(t *testing.T) {
	for _, id := range []string{"ab-0001-p", "fct-0001-p"} {
		stub := &nhiaHCPHandlerServiceStub{record: sampleNHIAHCP()}
		stub.record.ID = id
		handler := mustNewNHIAHCPHandler(t, stub)
		req := httptest.NewRequest(http.MethodGet, nhiaHCPRoute+"/"+id, nil)
		req.SetPathValue("provider_id", id)
		w := httptest.NewRecorder()
		handler.GetNHIAActiveAccreditedHealthcareProvider(w, req)
		if w.Code != http.StatusOK || stub.calls != 1 || stub.id != id {
			t.Fatalf("status=%d calls=%d id=%q body=%s", w.Code, stub.calls, stub.id, w.Body.String())
		}
		var body struct {
			Success bool           `json:"success"`
			Data    map[string]any `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		assertNHIAHCPResponseFields(t, body.Data)
	}
}

func TestNHIAActiveAccreditedHealthcareProviderHandlerDetailErrors(t *testing.T) {
	for _, id := range []string{"", "Bad", "bad_id", "bad/id", "bad%2Fid", "../bad", strings.Repeat("a", 256)} {
		stub := &nhiaHCPHandlerServiceStub{}
		handler := mustNewNHIAHCPHandler(t, stub)
		req := httptest.NewRequest(http.MethodGet, nhiaHCPRoute+"/"+id, nil)
		req.SetPathValue("provider_id", id)
		w := httptest.NewRecorder()
		handler.GetNHIAActiveAccreditedHealthcareProvider(w, req)
		if w.Code != http.StatusBadRequest || stub.calls != 0 {
			t.Fatalf("id=%q status=%d calls=%d body=%s", id, w.Code, stub.calls, w.Body.String())
		}
	}
	stub := &nhiaHCPHandlerServiceStub{err: services.ErrNHIAActiveAccreditedHealthcareProviderNotFound}
	handler := mustNewNHIAHCPHandler(t, stub)
	req := httptest.NewRequest(http.MethodGet, nhiaHCPRoute+"/unknown-valid-id", nil)
	req.SetPathValue("provider_id", "unknown-valid-id")
	w := httptest.NewRecorder()
	handler.GetNHIAActiveAccreditedHealthcareProvider(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestNHIAActiveAccreditedHealthcareProviderHandlerServiceErrors(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want int
	}{
		{errors.New("secret decoder address path"), http.StatusInternalServerError},
		{context.Canceled, http.StatusServiceUnavailable},
		{context.DeadlineExceeded, http.StatusServiceUnavailable},
		{services.ErrInvalidNHIAActiveAccreditedHealthcareProviderCode, http.StatusBadRequest},
	} {
		stub := &nhiaHCPHandlerServiceStub{err: tc.err}
		handler := mustNewNHIAHCPHandler(t, stub)
		w := httptest.NewRecorder()
		handler.ListNHIAActiveAccreditedHealthcareProviders(w, httptest.NewRequest(http.MethodGet, nhiaHCPRoute, nil))
		if w.Code != tc.want || strings.Contains(w.Body.String(), "secret") || strings.Contains(w.Body.String(), "address") {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	}
}

func mustNewNHIAHCPHandler(t testing.TB, service nhiaActiveAccreditedHealthcareProviderService) *NHIAActiveAccreditedHealthcareProviderHandler {
	t.Helper()
	handler, err := NewNHIAActiveAccreditedHealthcareProviderHandler(service)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func assertNHIAHCPResponseFields(t testing.TB, record map[string]any) {
	t.Helper()
	expected := map[string]bool{"id": true, "name": true, "country_code": true, "provider_code": true, "facility_type": true, "listing_status": true}
	if len(record) != len(expected) {
		t.Fatalf("unexpected response fields: %#v", record)
	}
	for field := range record {
		if !expected[field] {
			t.Fatalf("forbidden field %q in response", field)
		}
	}
}

type nhiaHCPHandlerServiceStub struct {
	query  interfaces.NHIAActiveAccreditedHealthcareProviderQuery
	id     string
	list   interfaces.NHIAActiveAccreditedHealthcareProviderListResult
	record models.NHIAActiveAccreditedHealthcareProvider
	err    error
	calls  int
}

func (s *nhiaHCPHandlerServiceStub) ListNHIAActiveAccreditedHealthcareProviders(_ context.Context, query interfaces.NHIAActiveAccreditedHealthcareProviderQuery) (interfaces.NHIAActiveAccreditedHealthcareProviderListResult, error) {
	s.calls++
	s.query = query
	if s.err != nil {
		return interfaces.NHIAActiveAccreditedHealthcareProviderListResult{}, s.err
	}
	return s.list, nil
}

func (s *nhiaHCPHandlerServiceStub) GetNHIAActiveAccreditedHealthcareProvider(_ context.Context, id string) (models.NHIAActiveAccreditedHealthcareProvider, error) {
	s.calls++
	s.id = id
	if s.err != nil {
		return models.NHIAActiveAccreditedHealthcareProvider{}, s.err
	}
	return s.record, nil
}

func sampleNHIAHCP() models.NHIAActiveAccreditedHealthcareProvider {
	return models.NHIAActiveAccreditedHealthcareProvider{ID: "fct-0001-p", Name: "WILDOT CLINIC", CountryCode: "NG", ProviderCode: "FCT/0001/P", FacilityType: "primary", ListingStatus: "active_accredited"}
}
