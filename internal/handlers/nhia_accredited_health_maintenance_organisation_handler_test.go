package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type nhiaHMOHandlerStub struct {
	query   interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery
	id      string
	ctx     context.Context
	calls   int
	records []models.NHIAAccreditedHealthMaintenanceOrganisation
	total   int
	err     error
}

func (s *nhiaHMOHandlerStub) ListNHIAAccreditedHealthMaintenanceOrganisations(ctx context.Context, q interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery) (interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult, error) {
	s.query = q
	s.ctx = ctx
	s.calls++
	pages := 0
	if s.total > 0 {
		pages = s.total / q.PageSize
		if s.total%q.PageSize != 0 {
			pages++
		}
	}
	return interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{Records: s.records, Page: q.Page, PageSize: q.PageSize, Total: s.total, TotalPages: pages}, s.err
}

func (s *nhiaHMOHandlerStub) GetNHIAAccreditedHealthMaintenanceOrganisation(ctx context.Context, id string) (models.NHIAAccreditedHealthMaintenanceOrganisation, error) {
	s.id = id
	s.ctx = ctx
	s.calls++
	if len(s.records) > 0 {
		return s.records[0], s.err
	}
	return models.NHIAAccreditedHealthMaintenanceOrganisation{}, s.err
}

func nhiaHMORequest(t *testing.T, s *nhiaHMOHandlerStub, query string, id *string) *httptest.ResponseRecorder {
	t.Helper()
	h, err := NewNHIAAccreditedHealthMaintenanceOrganisationHandler(s)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodGet, "/v1/healthcare/nhia-accredited-health-maintenance-organisations?"+query, nil)
	w := httptest.NewRecorder()
	if id == nil {
		h.ListNHIAAccreditedHealthMaintenanceOrganisations(w, r)
	} else {
		r.SetPathValue("organisation_id", *id)
		h.GetNHIAAccreditedHealthMaintenanceOrganisation(w, r)
	}
	if !json.Valid(w.Body.Bytes()) || w.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("invalid JSON response: %s", w.Body.String())
	}
	return w
}

func TestNHIAAccreditedHMOListContract(t *testing.T) {
	for _, tc := range []struct {
		name, raw         string
		page, size, total int
		status, hmoID     string
		search            string
		records           []models.NHIAAccreditedHealthMaintenanceOrganisation
	}{
		{name: "defaults", page: 1, size: 50, total: 94},
		{name: "explicit", raw: "page=2&page_size=25", page: 2, size: 25, total: 94},
		{name: "maximum", raw: "page_size=100", page: 1, size: 100, total: 94},
		{name: "status", raw: "accreditation_status=accredited", page: 1, size: 50, status: "accredited", total: 94},
		{name: "hmo id", raw: "hmo_id=0012", page: 1, size: 50, hmoID: "0012", total: 1},
		{name: "trimmed filters", raw: "accreditation_status=%20accredited%20&hmo_id=%200012%20&search=%20Health%20", page: 1, size: 50, status: "accredited", hmoID: "0012", search: "Health", total: 1},
		{name: "search", raw: "search=hEaLtH", page: 1, size: 50, search: "hEaLtH", total: 2},
		{name: "combined", raw: "accreditation_status=accredited&hmo_id=102&search=A%26M&page=1&page_size=1", page: 1, size: 1, status: "accredited", hmoID: "102", search: "A&M", total: 1},
		{name: "beyond final", raw: "page=3&page_size=50", page: 3, size: 50, total: 94},
		{name: "nil data", page: 1, size: 50, total: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &nhiaHMOHandlerStub{total: tc.total, records: tc.records}
			w := nhiaHMORequest(t, s, tc.raw, nil)
			want := interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Page: tc.page, PageSize: tc.size, AccreditationStatus: tc.status, HMOID: tc.hmoID, Search: tc.search}
			if w.Code != 200 || s.query != want || s.calls != 1 {
				t.Fatalf("response %d %s query %+v", w.Code, w.Body.String(), s.query)
			}
			var body struct {
				Success bool
				Data    []models.NHIAAccreditedHealthMaintenanceOrganisation
				Meta    map[string]int
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			pages := 0
			if tc.total > 0 {
				pages = tc.total / tc.size
				if tc.total%tc.size != 0 {
					pages++
				}
			}
			if !body.Success || body.Data == nil || body.Meta["page"] != tc.page || body.Meta["page_size"] != tc.size || body.Meta["total"] != tc.total || body.Meta["total_pages"] != pages || strings.Contains(w.Body.String(), `"limit"`) {
				t.Fatalf("envelope: %s", w.Body.String())
			}
		})
	}
}

func TestNHIAAccreditedHMOInvalidQueries(t *testing.T) {
	cases := []string{"accreditation_status=expired", "accreditation_status=%20%20", "hmo_id=%20", "hmo_id=a%00b", "hmo_id=" + strings.Repeat("x", 65), "search=%20", "search=" + strings.Repeat("a", 101), "state_id=lagos", "lga_id=ikeja", "organisation_type=health_maintenance_organisation", "licence_status=licensed", "website=x"}
	for _, field := range []string{"page", "page_size"} {
		for _, value := range []string{"", "0", "-1", "+1", "1.5", "abc", "18446744073709551616", "9223372036854775808", " 1"} {
			cases = append(cases, field+"="+url.QueryEscape(value))
		}
	}
	cases = append(cases, "page_size=101")
	for _, field := range []string{"page", "page_size", "accreditation_status", "hmo_id", "search"} {
		cases = append(cases, field+"=1&"+field+"=2")
	}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			s := &nhiaHMOHandlerStub{}
			w := nhiaHMORequest(t, s, raw, nil)
			if w.Code != 400 || s.calls != 0 || !strings.Contains(w.Body.String(), "VALIDATION_FAILED") {
				t.Fatalf("%d %s calls %d", w.Code, w.Body.String(), s.calls)
			}
		})
	}
}

func TestNHIAAccreditedHMODetailAndSerialization(t *testing.T) {
	longest := "new-healthway-company-limited-formerly-songhai-health-trust-limited-34"
	for _, id := range []string{"a-and-m-healthcare-trust-limited-102", longest, strings.Repeat("a", 255)} {
		record := models.NHIAAccreditedHealthMaintenanceOrganisation{ID: id, Name: "A&M HEALTHCARE TRUST LIMITED", CountryCode: "NG", OrganisationType: "health_maintenance_organisation", AccreditationStatus: "accredited", HMOID: "00102"}
		s := &nhiaHMOHandlerStub{records: []models.NHIAAccreditedHealthMaintenanceOrganisation{record}, total: 1}
		w := nhiaHMORequest(t, s, "", &id)
		expected, _ := json.Marshal(record)
		if w.Code != 200 || !strings.Contains(w.Body.String(), string(expected)) || s.id != id || strings.Contains(w.Body.String(), `"meta"`) {
			t.Fatalf("detail serialization: %d %s id=%q", w.Code, w.Body.String(), s.id)
		}
		for _, forbidden := range []string{"state_id", "lga_id", "address", "website", "logo", "phone", "email", "licence_status", "registration_status", "operational_status"} {
			if strings.Contains(w.Body.String(), forbidden) {
				t.Fatalf("forbidden field %q leaked in %s", forbidden, w.Body.String())
			}
		}
	}
	for _, id := range []string{"", "Upper", "bad_id", "bad/id", "bad%2Fid", "bad\\id", "bad--id", "-bad", "bad-", " id ", strings.Repeat("a", 256)} {
		s := &nhiaHMOHandlerStub{}
		w := nhiaHMORequest(t, s, "", &id)
		if w.Code != 400 || s.calls != 0 {
			t.Fatalf("invalid ID %q: %d", id, w.Code)
		}
	}
}

func TestNHIAAccreditedHMOServiceErrors(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{services.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationID, 400, "INVALID_REQUEST"},
		{services.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationPagination, 400, "INVALID_REQUEST"},
		{services.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationStatus, 400, "INVALID_REQUEST"},
		{services.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationHMOID, 400, "INVALID_REQUEST"},
		{services.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationSearch, 400, "INVALID_REQUEST"},
		{services.ErrNHIAAccreditedHealthMaintenanceOrganisationNotFound, 404, "RESOURCE_NOT_FOUND"},
		{services.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationDataset, 500, "INTERNAL_ERROR"},
		{errors.New("decoder cache source HTML reconciliation secret-path"), 500, "INTERNAL_ERROR"},
		{context.Canceled, 503, "SERVICE_UNAVAILABLE"},
		{context.DeadlineExceeded, 503, "SERVICE_UNAVAILABLE"},
	} {
		for _, wrapped := range []bool{false, true} {
			for _, detail := range []bool{false, true} {
				err := tc.err
				if wrapped {
					err = fmt.Errorf("secret-path: %w", err)
				}
				s := &nhiaHMOHandlerStub{err: err}
				id := "unknown-valid-id"
				var pathID *string
				if detail {
					pathID = &id
				}
				w := nhiaHMORequest(t, s, "", pathID)
				if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.code) || strings.Contains(w.Body.String(), "secret-path") || strings.Contains(w.Body.String(), "services:") || strings.Contains(w.Body.String(), "decoder") {
					t.Fatalf("%v: %d %s", err, w.Code, w.Body.String())
				}
			}
		}
	}
}

func TestNHIAAccreditedHMOContextAndMethod(t *testing.T) {
	if _, err := NewNHIAAccreditedHealthMaintenanceOrganisationHandler(nil); err == nil {
		t.Fatal("nil service accepted")
	}
	s := &nhiaHMOHandlerStub{}
	h, _ := NewNHIAAccreditedHealthMaintenanceOrganisationHandler(s)
	for _, method := range []func(http.ResponseWriter, *http.Request){h.ListNHIAAccreditedHealthMaintenanceOrganisations, h.GetNHIAAccreditedHealthMaintenanceOrganisation} {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()
		method(w, r)
		if w.Code != 405 || w.Header().Get("Allow") != http.MethodGet || s.calls != 0 {
			t.Fatal("method guard")
		}
	}
	for _, deadline := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		if deadline {
			cancel()
			ctx, cancel = context.WithDeadline(context.Background(), time.Unix(0, 0))
		} else {
			cancel()
		}
		for _, method := range []func(http.ResponseWriter, *http.Request){h.ListNHIAAccreditedHealthMaintenanceOrganisations, h.GetNHIAAccreditedHealthMaintenanceOrganisation} {
			s.err = ctx.Err()
			r := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
			r.SetPathValue("organisation_id", "valid-id")
			w := httptest.NewRecorder()
			method(w, r)
			if s.ctx != ctx || w.Code != 503 {
				t.Fatal("context not preserved")
			}
		}
		cancel()
	}
}
