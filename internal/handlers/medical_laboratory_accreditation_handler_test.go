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

type accreditationStub struct {
	query   interfaces.MedicalLaboratoryAccreditationQuery
	id      string
	ctx     context.Context
	calls   int
	records []models.MedicalLaboratoryAccreditation
	total   int
	err     error
}

func (s *accreditationStub) ListMedicalLaboratoryAccreditations(ctx context.Context, q interfaces.MedicalLaboratoryAccreditationQuery) (interfaces.MedicalLaboratoryAccreditationListResult, error) {
	s.query = q
	s.ctx = ctx
	s.calls++
	pages := s.total / q.PageSize
	if s.total%q.PageSize != 0 {
		pages++
	}
	return interfaces.MedicalLaboratoryAccreditationListResult{Records: s.records, Page: q.Page, PageSize: q.PageSize, Total: s.total, TotalPages: pages}, s.err
}
func (s *accreditationStub) GetMedicalLaboratoryAccreditation(ctx context.Context, id string) (models.MedicalLaboratoryAccreditation, error) {
	s.id = id
	s.ctx = ctx
	s.calls++
	if len(s.records) > 0 {
		return s.records[0], s.err
	}
	return models.MedicalLaboratoryAccreditation{}, s.err
}
func accreditationRequest(t *testing.T, s *accreditationStub, query string, id *string) *httptest.ResponseRecorder {
	t.Helper()
	h, err := NewMedicalLaboratoryAccreditationHandler(s)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodGet, "/v1/healthcare/medical-laboratory-accreditations?"+query, nil)
	w := httptest.NewRecorder()
	if id == nil {
		h.ListMedicalLaboratoryAccreditations(w, r)
	} else {
		r.SetPathValue("accreditation_id", *id)
		h.GetMedicalLaboratoryAccreditation(w, r)
	}
	if !json.Valid(w.Body.Bytes()) || w.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("invalid JSON response: %s", w.Body.String())
	}
	return w
}
func TestAccreditationListContract(t *testing.T) {
	for _, tc := range []struct {
		name, raw             string
		page, size            int
		state, status, search string
		total                 int
	}{
		{name: "defaults", page: 1, size: 50, total: 30},
		{name: "explicit", raw: "page=2&page_size=3", page: 2, size: 3, total: 26},
		{name: "maximum", raw: "page_size=100", page: 1, size: 100, total: 30},
		{name: "state", raw: "state_id=%20lagos%20", page: 1, size: 50, state: "lagos", total: 5},
		{name: "accredited", raw: "accreditation_status=accredited", page: 1, size: 50, status: "accredited", total: 26},
		{name: "expired", raw: "accreditation_status=%20expired%20", page: 1, size: 50, status: "expired", total: 4},
		{name: "search trim", raw: "search=%20Example%20", page: 1, size: 50, search: "Example", total: 1},
		{name: "blank filters", raw: "search=%20%20&state_id=%20&accreditation_status=%20", page: 1, size: 50},
		{name: "combined", raw: "state_id=oyo&accreditation_status=expired&search=%20Lab%20&page=2&page_size=1", page: 2, size: 1, state: "oyo", status: "expired", search: "Lab", total: 2},
		{name: "beyond final", raw: "page=999", page: 999, size: 50, total: 30},
		{name: "empty", page: 1, size: 50},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &accreditationStub{total: tc.total}
			w := accreditationRequest(t, s, tc.raw, nil)
			want := interfaces.MedicalLaboratoryAccreditationQuery{Page: tc.page, PageSize: tc.size, StateID: tc.state, AccreditationStatus: tc.status, Search: tc.search}
			if w.Code != 200 || s.query != want || s.calls != 1 {
				t.Fatalf("response %d %s query %+v", w.Code, w.Body.String(), s.query)
			}
			var body struct {
				Success bool
				Data    []models.MedicalLaboratoryAccreditation
				Meta    map[string]int
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			pages := tc.total / tc.size
			if tc.total%tc.size != 0 {
				pages++
			}
			if !body.Success || body.Data == nil || len(body.Meta) != 4 || body.Meta["page"] != tc.page || body.Meta["page_size"] != tc.size || body.Meta["total"] != tc.total || body.Meta["total_pages"] != pages {
				t.Fatalf("envelope: %s", w.Body.String())
			}
		})
	}
}
func TestAccreditationInvalidQueries(t *testing.T) {
	cases := []string{"state_id=Bad_State", "accreditation_status=licensed", "search=" + strings.Repeat("a", 101), "search=a%00b", "search=a%0Ab", "search=a%0Db"}
	for _, field := range []string{"page", "page_size"} {
		for _, v := range []string{"", "0", "-1", "+1", "1.5", "abc", "18446744073709551616", "9223372036854775808", " 1"} {
			cases = append(cases, field+"="+url.QueryEscape(v))
		}
	}
	cases = append(cases, "page_size=101")
	for _, field := range []string{"page", "page_size", "state_id", "accreditation_status", "search"} {
		cases = append(cases, field+"=1&"+field+"=2")
	}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			s := &accreditationStub{}
			w := accreditationRequest(t, s, raw, nil)
			if w.Code != 400 || s.calls != 0 || !strings.Contains(w.Body.String(), "VALIDATION_FAILED") {
				t.Fatalf("%d %s calls %d", w.Code, w.Body.String(), s.calls)
			}
		})
	}
}
func TestAccreditationDetailAndSerialization(t *testing.T) {
	longest := "onchocerciasis-and-soil-transmitted-helminth-oncho-sth-laboratory-public-health-and-epidemiology-department-nigeria-institute-of-medical-research-lagos-yaba-ml0025"
	for _, id := range []string{"example-laboratory", longest, strings.Repeat("a", 255)} {
		for _, optional := range []bool{false, true} {
			record := models.MedicalLaboratoryAccreditation{ID: id, Name: "Example", StateID: "lagos", CountryCode: "NG", AccreditationStatus: "expired"}
			if optional {
				record.AccreditationNumber = "ML0001"
				record.ApprovalDate = "2025-01-01"
				record.ExpiryDate = "2026-01-01"
				record.Address = "Institutional address"
			}
			s := &accreditationStub{records: []models.MedicalLaboratoryAccreditation{record}, total: 1}
			for _, detail := range []bool{false, true} {
				var pathID *string
				if detail {
					pathID = &id
				}
				w := accreditationRequest(t, s, "", pathID)
				expected, _ := json.Marshal(record)
				if w.Code != 200 || !strings.Contains(w.Body.String(), string(expected)) {
					t.Fatalf("serialization: %d %s", w.Code, w.Body.String())
				}
				if detail && (s.id != id || strings.Contains(w.Body.String(), `"meta"`)) {
					t.Fatal("detail envelope or ID mismatch")
				}
			}
		}
	}
	for _, id := range []string{"", "Upper", "bad_id", "bad/id", "bad--id", "-bad", "bad-", " id ", strings.Repeat("a", 256)} {
		s := &accreditationStub{}
		w := accreditationRequest(t, s, "", &id)
		if w.Code != 400 || s.calls != 0 {
			t.Fatalf("invalid ID %q: %d", id, w.Code)
		}
	}
}
func TestAccreditationServiceErrors(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{services.ErrInvalidMedicalLaboratoryAccreditationID, 400, "INVALID_REQUEST"},
		{services.ErrInvalidMedicalLaboratoryAccreditationPagination, 400, "INVALID_REQUEST"},
		{services.ErrInvalidMedicalLaboratoryAccreditationStateID, 400, "INVALID_REQUEST"},
		{services.ErrInvalidMedicalLaboratoryAccreditationStatus, 400, "INVALID_REQUEST"},
		{services.ErrInvalidMedicalLaboratoryAccreditationSearch, 400, "INVALID_REQUEST"},
		{services.ErrMedicalLaboratoryAccreditationNotFound, 404, "RESOURCE_NOT_FOUND"},
		{services.ErrInvalidMedicalLaboratoryAccreditationDataset, 500, "INTERNAL_ERROR"},
		{errors.New("decoder cache source HTML reconciliation secret-path"), 500, "INTERNAL_ERROR"},
		{context.Canceled, 503, "SERVICE_UNAVAILABLE"}, {context.DeadlineExceeded, 503, "SERVICE_UNAVAILABLE"},
	} {
		for _, wrapped := range []bool{false, true} {
			for _, detail := range []bool{false, true} {
				err := tc.err
				if wrapped {
					err = fmt.Errorf("secret-path: %w", err)
				}
				s := &accreditationStub{err: err}
				id := "unknown-valid-id"
				var pathID *string
				if detail {
					pathID = &id
				}
				w := accreditationRequest(t, s, "state_id=unknown-state", pathID)
				if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.code) || strings.Contains(w.Body.String(), "secret-path") || strings.Contains(w.Body.String(), "services:") {
					t.Fatalf("%v: %d %s", err, w.Code, w.Body.String())
				}
			}
		}
	}
}
func TestAccreditationContextAndMethod(t *testing.T) {
	if _, err := NewMedicalLaboratoryAccreditationHandler(nil); err == nil {
		t.Fatal("nil service accepted")
	}
	s := &accreditationStub{}
	h, _ := NewMedicalLaboratoryAccreditationHandler(s)
	for _, method := range []func(http.ResponseWriter, *http.Request){h.ListMedicalLaboratoryAccreditations, h.GetMedicalLaboratoryAccreditation} {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()
		method(w, r)
		if w.Code != 405 || s.calls != 0 {
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
		for _, method := range []func(http.ResponseWriter, *http.Request){h.ListMedicalLaboratoryAccreditations, h.GetMedicalLaboratoryAccreditation} {
			s.err = ctx.Err()
			r := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
			r.SetPathValue("accreditation_id", "valid-id")
			w := httptest.NewRecorder()
			method(w, r)
			if s.ctx != ctx || w.Code != 503 {
				t.Fatal("context not preserved")
			}
		}
		cancel()
	}
}
