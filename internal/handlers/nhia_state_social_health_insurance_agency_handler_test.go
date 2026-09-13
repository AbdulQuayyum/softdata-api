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

const nhiaSSHIARoute = "/v1/healthcare/nhia-state-social-health-insurance-agencies"

func TestNHIAStateSocialHealthInsuranceAgencyHandlerList(t *testing.T) {
	for _, tc := range []struct {
		name  string
		query string
		want  interfaces.NHIAStateSocialHealthInsuranceAgencyQuery
	}{
		{"default", "", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 1, PageSize: 50}},
		{"explicit", "page=2&page_size=10", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 2, PageSize: 10}},
		{"max", "page_size=100", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 1, PageSize: 100}},
		{"state", "state_id=akwa-ibom", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 1, PageSize: 50, StateID: "akwa-ibom"}},
		{"fct", "state_id=fct", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 1, PageSize: 50, StateID: "fct"}},
		{"rivers", "state_id=rivers", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 1, PageSize: 50, StateID: "rivers"}},
		{"trimmed state", "state_id=%20cross-river%20", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 1, PageSize: 50, StateID: "cross-river"}},
		{"search", "search=scheme", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 1, PageSize: 50, Search: "scheme"}},
		{"trimmed search", "search=%20LASHMA%20", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 1, PageSize: 50, Search: "LASHMA"}},
		{"combined", "state_id=lagos&search=LASHMA", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 1, PageSize: 50, StateID: "lagos", Search: "LASHMA"}},
		{"filter before pagination", "state_id=lagos&page=2&page_size=1", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 2, PageSize: 1, StateID: "lagos"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &nhiaSSHIAHandlerServiceStub{list: interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{Records: []models.NHIAStateSocialHealthInsuranceAgency{sampleNHIASSHIA()}, Page: tc.want.Page, PageSize: tc.want.PageSize, Total: 37, TotalPages: 1}}
			handler := mustNewNHIASSHIAHandler(t, stub)
			w := httptest.NewRecorder()
			target := nhiaSSHIARoute
			if tc.query != "" {
				target += "?" + tc.query
			}
			handler.ListNHIAStateSocialHealthInsuranceAgencies(w, httptest.NewRequest(http.MethodGet, target, nil))
			if w.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if stub.query != tc.want {
				t.Fatalf("query=%#v want %#v", stub.query, tc.want)
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
			for _, forbidden := range []string{"director", "phone", "email", "address", "website", "logo", "accreditation_status"} {
				if _, ok := body.Data[0][forbidden]; ok {
					t.Fatalf("forbidden field %q in response", forbidden)
				}
			}
		})
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyHandlerListValidationErrors(t *testing.T) {
	for _, query := range []string{"page=0", "page_size=0", "page_size=101", "page=abc", "page=999999999999999999999999999999", "state_id=AKS", "state_id=River+State", "search=%20", "search=" + strings.Repeat("a", 101)} {
		stub := &nhiaSSHIAHandlerServiceStub{}
		handler := mustNewNHIASSHIAHandler(t, stub)
		w := httptest.NewRecorder()
		handler.ListNHIAStateSocialHealthInsuranceAgencies(w, httptest.NewRequest(http.MethodGet, nhiaSSHIARoute+"?"+query, nil))
		if w.Code != http.StatusBadRequest || stub.listCalled {
			t.Fatalf("query=%q status=%d called=%v body=%s", query, w.Code, stub.listCalled, w.Body.String())
		}
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyHandlerNilAndBeyondFinalList(t *testing.T) {
	for _, records := range [][]models.NHIAStateSocialHealthInsuranceAgency{nil, {}} {
		stub := &nhiaSSHIAHandlerServiceStub{list: interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{Records: records, Page: 2, PageSize: 50, Total: 37, TotalPages: 1}}
		handler := mustNewNHIASSHIAHandler(t, stub)
		w := httptest.NewRecorder()
		handler.ListNHIAStateSocialHealthInsuranceAgencies(w, httptest.NewRequest(http.MethodGet, nhiaSSHIARoute+"?page=2", nil))
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"data":[]`) {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyHandlerDetail(t *testing.T) {
	for _, id := range []string{"lashma", "borno-state-contributory-health-care-management-agency-boschma"} {
		stub := &nhiaSSHIAHandlerServiceStub{record: sampleNHIASSHIA()}
		stub.record.ID = id
		handler := mustNewNHIASSHIAHandler(t, stub)
		req := httptest.NewRequest(http.MethodGet, nhiaSSHIARoute+"/"+id, nil)
		req.SetPathValue("agency_id", id)
		w := httptest.NewRecorder()
		handler.GetNHIAStateSocialHealthInsuranceAgency(w, req)
		if w.Code != http.StatusOK || stub.id != id {
			t.Fatalf("status=%d id=%q body=%s", w.Code, stub.id, w.Body.String())
		}
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyHandlerDetailErrors(t *testing.T) {
	for _, id := range []string{"", "Bad", "bad_id", "bad/id", "bad%2Fid", "../bad"} {
		stub := &nhiaSSHIAHandlerServiceStub{}
		handler := mustNewNHIASSHIAHandler(t, stub)
		req := httptest.NewRequest(http.MethodGet, nhiaSSHIARoute+"/"+id, nil)
		req.SetPathValue("agency_id", id)
		w := httptest.NewRecorder()
		handler.GetNHIAStateSocialHealthInsuranceAgency(w, req)
		if w.Code != http.StatusBadRequest || stub.detailCalled {
			t.Fatalf("id=%q status=%d called=%v body=%s", id, w.Code, stub.detailCalled, w.Body.String())
		}
	}
	stub := &nhiaSSHIAHandlerServiceStub{err: services.ErrNHIAStateSocialHealthInsuranceAgencyNotFound}
	handler := mustNewNHIASSHIAHandler(t, stub)
	req := httptest.NewRequest(http.MethodGet, nhiaSSHIARoute+"/unknown-valid-id", nil)
	req.SetPathValue("agency_id", "unknown-valid-id")
	w := httptest.NewRecorder()
	handler.GetNHIAStateSocialHealthInsuranceAgency(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyHandlerServiceErrors(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want int
	}{
		{errors.New("secret decoder director path"), http.StatusInternalServerError},
		{context.Canceled, http.StatusServiceUnavailable},
		{context.DeadlineExceeded, http.StatusServiceUnavailable},
	} {
		stub := &nhiaSSHIAHandlerServiceStub{err: tc.err}
		handler := mustNewNHIASSHIAHandler(t, stub)
		w := httptest.NewRecorder()
		handler.ListNHIAStateSocialHealthInsuranceAgencies(w, httptest.NewRequest(http.MethodGet, nhiaSSHIARoute, nil))
		if w.Code != tc.want || strings.Contains(w.Body.String(), "secret") || strings.Contains(w.Body.String(), "director") {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	}
}

func mustNewNHIASSHIAHandler(t testing.TB, service nhiaStateSocialHealthInsuranceAgencyService) *NHIAStateSocialHealthInsuranceAgencyHandler {
	t.Helper()
	handler, err := NewNHIAStateSocialHealthInsuranceAgencyHandler(service)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

type nhiaSSHIAHandlerServiceStub struct {
	query        interfaces.NHIAStateSocialHealthInsuranceAgencyQuery
	id           string
	list         interfaces.NHIAStateSocialHealthInsuranceAgencyListResult
	record       models.NHIAStateSocialHealthInsuranceAgency
	err          error
	listCalled   bool
	detailCalled bool
}

func (s *nhiaSSHIAHandlerServiceStub) ListNHIAStateSocialHealthInsuranceAgencies(_ context.Context, query interfaces.NHIAStateSocialHealthInsuranceAgencyQuery) (interfaces.NHIAStateSocialHealthInsuranceAgencyListResult, error) {
	s.listCalled = true
	s.query = query
	if s.err != nil {
		return interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{}, s.err
	}
	return s.list, nil
}

func (s *nhiaSSHIAHandlerServiceStub) GetNHIAStateSocialHealthInsuranceAgency(_ context.Context, id string) (models.NHIAStateSocialHealthInsuranceAgency, error) {
	s.detailCalled = true
	s.id = id
	if s.err != nil {
		return models.NHIAStateSocialHealthInsuranceAgency{}, s.err
	}
	return s.record, nil
}

func sampleNHIASSHIA() models.NHIAStateSocialHealthInsuranceAgency {
	return models.NHIAStateSocialHealthInsuranceAgency{ID: "lashma", Name: "LASHMA", StateID: "lagos", CountryCode: "NG", OrganisationType: "state_social_health_insurance_agency"}
}
