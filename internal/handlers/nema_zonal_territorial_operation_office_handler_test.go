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

const nemaOfficeRoute = "/v1/emergency/nema-zonal-territorial-operation-offices"

func TestNEMAZonalTerritorialOperationOfficeHandlerList(t *testing.T) {
	for _, tc := range []struct {
		name  string
		query string
		want  interfaces.NEMAZonalTerritorialOperationOfficeQuery
	}{
		{"default", "", interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: 1, PageSize: 50}},
		{"explicit", "page=2&page_size=2", interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: 2, PageSize: 2}},
		{"max", "page_size=100", interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: 1, PageSize: 100}},
		{"state", "state_id=lagos", interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: 1, PageSize: 50, StateID: "lagos"}},
		{"office type", "office_type=zonal_territorial_operation_office", interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: 1, PageSize: 50, OfficeType: "zonal_territorial_operation_office"}},
		{"search", "search=lagos", interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: 1, PageSize: 50, Search: "lagos"}},
		{"empty search omitted", "search=%20", interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: 1, PageSize: 50}},
		{"trimmed combined", "state_id=%20rivers%20&office_type=%20zonal_territorial_operation_office%20&search=%20Port%20Harcourt%20", interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: 1, PageSize: 50, StateID: "rivers", OfficeType: "zonal_territorial_operation_office", Search: "Port Harcourt"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &nemaOfficeHandlerServiceStub{list: interfaces.NEMAZonalTerritorialOperationOfficeListResult{Records: sampleNEMAOffices(), Page: tc.want.Page, PageSize: tc.want.PageSize, Total: 17, TotalPages: 1}}
			handler := mustNewNEMAOfficeHandler(t, stub)
			target := nemaOfficeRoute
			if tc.query != "" {
				target += "?" + tc.query
			}
			w := httptest.NewRecorder()
			handler.ListNEMAZonalTerritorialOperationOffices(w, httptest.NewRequest(http.MethodGet, target, nil))
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
			assertNEMAOfficeResponseFields(t, body.Data[0])
		})
	}
}

func TestNEMAZonalTerritorialOperationOfficeHandlerListValidationErrors(t *testing.T) {
	for _, query := range []string{
		"page=abc", "page=999999999999999999999999999999", "page=0", "page=-1",
		"page_size=0", "page_size=101", "state_id=bad_state", "state_id=unknown",
		"office_type=regional", "search=" + strings.Repeat("a", 101), "address=hidden", "page=1&page=2",
	} {
		stub := &nemaOfficeHandlerServiceStub{}
		handler := mustNewNEMAOfficeHandler(t, stub)
		w := httptest.NewRecorder()
		handler.ListNEMAZonalTerritorialOperationOffices(w, httptest.NewRequest(http.MethodGet, nemaOfficeRoute+"?"+query, nil))
		if w.Code != http.StatusBadRequest || stub.calls != 0 {
			t.Fatalf("query=%q status=%d calls=%d body=%s", query, w.Code, stub.calls, w.Body.String())
		}
	}
}

func TestNEMAZonalTerritorialOperationOfficeHandlerNilAndEmptyList(t *testing.T) {
	for _, records := range [][]models.NEMAZonalTerritorialOperationOffice{nil, {}} {
		stub := &nemaOfficeHandlerServiceStub{list: interfaces.NEMAZonalTerritorialOperationOfficeListResult{Records: records, Page: 2, PageSize: 50, Total: 17, TotalPages: 1}}
		handler := mustNewNEMAOfficeHandler(t, stub)
		w := httptest.NewRecorder()
		handler.ListNEMAZonalTerritorialOperationOffices(w, httptest.NewRequest(http.MethodGet, nemaOfficeRoute+"?page=2", nil))
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"data":[]`) || !strings.Contains(w.Body.String(), `"page_size":50`) {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	}
}

func TestNEMAZonalTerritorialOperationOfficeHandlerDetailAndErrors(t *testing.T) {
	stub := &nemaOfficeHandlerServiceStub{record: sampleNEMAOffices()[0]}
	handler := mustNewNEMAOfficeHandler(t, stub)
	req := httptest.NewRequest(http.MethodGet, nemaOfficeRoute+"/nema-lagos-zonal-territorial-operation-office", nil)
	req.SetPathValue("office_id", "nema-lagos-zonal-territorial-operation-office")
	w := httptest.NewRecorder()
	handler.GetNEMAZonalTerritorialOperationOffice(w, req)
	if w.Code != http.StatusOK || stub.calls != 1 || stub.id != "nema-lagos-zonal-territorial-operation-office" {
		t.Fatalf("status=%d calls=%d id=%q body=%s", w.Code, stub.calls, stub.id, w.Body.String())
	}
	var body struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	assertNEMAOfficeResponseFields(t, body.Data)

	longID := strings.Repeat("a", 255)
	stub = &nemaOfficeHandlerServiceStub{record: models.NEMAZonalTerritorialOperationOffice{ID: longID, Name: "Synthetic NEMA Office", OfficeType: "zonal_territorial_operation_office", StateID: "lagos", CountryCode: "NG"}}
	handler = mustNewNEMAOfficeHandler(t, stub)
	req = httptest.NewRequest(http.MethodGet, nemaOfficeRoute+"/"+longID, nil)
	req.SetPathValue("office_id", longID)
	w = httptest.NewRecorder()
	handler.GetNEMAZonalTerritorialOperationOffice(w, req)
	if w.Code != http.StatusOK || stub.id != longID {
		t.Fatalf("long id status=%d id=%q body=%s", w.Code, stub.id, w.Body.String())
	}

	for _, id := range []string{"", "Bad", "bad_id", "bad/id", "bad%2Fid", "../bad", strings.Repeat("a", 256)} {
		stub := &nemaOfficeHandlerServiceStub{}
		handler := mustNewNEMAOfficeHandler(t, stub)
		req := httptest.NewRequest(http.MethodGet, nemaOfficeRoute+"/"+id, nil)
		req.SetPathValue("office_id", id)
		w := httptest.NewRecorder()
		handler.GetNEMAZonalTerritorialOperationOffice(w, req)
		if w.Code != http.StatusBadRequest || stub.calls != 0 {
			t.Fatalf("id=%q status=%d calls=%d body=%s", id, w.Code, stub.calls, w.Body.String())
		}
	}
	stub = &nemaOfficeHandlerServiceStub{err: services.ErrNEMAZonalTerritorialOperationOfficeNotFound}
	handler = mustNewNEMAOfficeHandler(t, stub)
	req = httptest.NewRequest(http.MethodGet, nemaOfficeRoute+"/unknown-valid-id", nil)
	req.SetPathValue("office_id", "unknown-valid-id")
	w = httptest.NewRecorder()
	handler.GetNEMAZonalTerritorialOperationOffice(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestNEMAZonalTerritorialOperationOfficeHandlerServiceErrorsAndMethods(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want int
	}{
		{errors.New("internal/source decoder detail"), http.StatusInternalServerError},
		{context.Canceled, http.StatusServiceUnavailable},
		{context.DeadlineExceeded, http.StatusServiceUnavailable},
		{services.ErrInvalidNEMAZonalTerritorialOperationOfficeType, http.StatusBadRequest},
		{services.ErrInvalidNEMAZonalTerritorialOperationOfficePagination, http.StatusBadRequest},
	} {
		stub := &nemaOfficeHandlerServiceStub{err: tc.err}
		handler := mustNewNEMAOfficeHandler(t, stub)
		w := httptest.NewRecorder()
		handler.ListNEMAZonalTerritorialOperationOffices(w, httptest.NewRequest(http.MethodGet, nemaOfficeRoute, nil))
		if w.Code != tc.want || strings.Contains(w.Body.String(), "internal/source") || strings.Contains(w.Body.String(), "decoder") {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	}
	stub := &nemaOfficeHandlerServiceStub{}
	handler := mustNewNEMAOfficeHandler(t, stub)
	w := httptest.NewRecorder()
	handler.ListNEMAZonalTerritorialOperationOffices(w, httptest.NewRequest(http.MethodPost, nemaOfficeRoute, nil))
	if w.Code != http.StatusMethodNotAllowed || w.Header().Get("Allow") != http.MethodGet || stub.calls != 0 {
		t.Fatalf("status=%d allow=%q body=%s", w.Code, w.Header().Get("Allow"), w.Body.String())
	}
}

func mustNewNEMAOfficeHandler(t testing.TB, service nemaZonalTerritorialOperationOfficeService) *NEMAZonalTerritorialOperationOfficeHandler {
	t.Helper()
	handler, err := NewNEMAZonalTerritorialOperationOfficeHandler(service)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func assertNEMAOfficeResponseFields(t testing.TB, record map[string]any) {
	t.Helper()
	expected := map[string]bool{"id": true, "name": true, "office_type": true, "state_id": true, "country_code": true}
	for field := range record {
		if !expected[field] {
			t.Fatalf("forbidden field %q in response", field)
		}
	}
	for _, required := range []string{"id", "name", "office_type", "state_id", "country_code"} {
		if _, ok := record[required]; !ok {
			t.Fatalf("missing field %q in %#v", required, record)
		}
	}
}

type nemaOfficeHandlerServiceStub struct {
	query  interfaces.NEMAZonalTerritorialOperationOfficeQuery
	id     string
	list   interfaces.NEMAZonalTerritorialOperationOfficeListResult
	record models.NEMAZonalTerritorialOperationOffice
	err    error
	calls  int
}

func (s *nemaOfficeHandlerServiceStub) ListNEMAZonalTerritorialOperationOffices(_ context.Context, query interfaces.NEMAZonalTerritorialOperationOfficeQuery) (interfaces.NEMAZonalTerritorialOperationOfficeListResult, error) {
	s.calls++
	s.query = query
	if s.err != nil {
		return interfaces.NEMAZonalTerritorialOperationOfficeListResult{}, s.err
	}
	return s.list, nil
}

func (s *nemaOfficeHandlerServiceStub) GetNEMAZonalTerritorialOperationOffice(_ context.Context, id string) (models.NEMAZonalTerritorialOperationOffice, error) {
	s.calls++
	s.id = id
	if s.err != nil {
		return models.NEMAZonalTerritorialOperationOffice{}, s.err
	}
	return s.record, nil
}

func sampleNEMAOffices() []models.NEMAZonalTerritorialOperationOffice {
	return []models.NEMAZonalTerritorialOperationOffice{
		{ID: "nema-lagos-zonal-territorial-operation-office", Name: "NEMA Lagos Office", OfficeType: "zonal_territorial_operation_office", StateID: "lagos", CountryCode: "NG"},
		{ID: "nema-port-harcourt-zonal-territorial-operation-office", Name: "NEMA Port Harcourt Office", OfficeType: "zonal_territorial_operation_office", StateID: "rivers", CountryCode: "NG"},
	}
}
