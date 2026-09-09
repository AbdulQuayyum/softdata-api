package app

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/datasets"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	fileRepo "github.com/AbdulQuayyum/softdata-api/internal/repository/file"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type startupAccreditationStub struct {
	pages   map[string]interfaces.MedicalLaboratoryAccreditationListResult
	anchors map[string]models.MedicalLaboratoryAccreditation
	calls   []interfaces.MedicalLaboratoryAccreditationQuery
	ids     []string
	failAt  int
	err     error
}

func (s *startupAccreditationStub) ListMedicalLaboratoryAccreditations(_ context.Context, q interfaces.MedicalLaboratoryAccreditationQuery) (interfaces.MedicalLaboratoryAccreditationListResult, error) {
	s.calls = append(s.calls, q)
	if len(s.calls)+len(s.ids) == s.failAt {
		return interfaces.MedicalLaboratoryAccreditationListResult{}, s.err
	}
	key := q.AccreditationStatus
	if q.Page == 31 {
		key = "beyond"
	}
	return s.pages[key], nil
}
func (s *startupAccreditationStub) GetMedicalLaboratoryAccreditation(_ context.Context, id string) (models.MedicalLaboratoryAccreditation, error) {
	s.ids = append(s.ids, id)
	if len(s.calls)+len(s.ids) == s.failAt {
		return models.MedicalLaboratoryAccreditation{}, s.err
	}
	r, ok := s.anchors[id]
	if !ok {
		return r, errors.New("missing anchor")
	}
	return r, nil
}
func validStartupAccreditation() *startupAccreditationStub {
	first := models.MedicalLaboratoryAccreditation{ID: medicalLaboratoryFirstAnchor, Name: "Laboratory", StateID: "akwa-ibom", CountryCode: "NG", AccreditationStatus: "accredited"}
	last := first
	last.ID = medicalLaboratoryLastAnchor
	last.StateID = "rivers"
	s := &startupAccreditationStub{pages: map[string]interfaces.MedicalLaboratoryAccreditationListResult{}, anchors: map[string]models.MedicalLaboratoryAccreditation{first.ID: first, last.ID: last}}
	for key, total := range map[string]int{"": 30, "accredited": 26, "expired": 4} {
		r := first
		if key != "" {
			r.AccreditationStatus = key
		}
		s.pages[key] = interfaces.MedicalLaboratoryAccreditationListResult{Records: []models.MedicalLaboratoryAccreditation{r}, Page: 1, PageSize: 1, Total: total, TotalPages: total}
	}
	s.pages["beyond"] = interfaces.MedicalLaboratoryAccreditationListResult{Records: []models.MedicalLaboratoryAccreditation{}, Page: 31, PageSize: 1, Total: 30, TotalPages: 30}
	return s
}
func TestAccreditationStartupBoundedQueries(t *testing.T) {
	s := validStartupAccreditation()
	if err := verifyMedicalLaboratoryAccreditations(nil, s); err != nil {
		t.Fatal(err)
	}
	want := []interfaces.MedicalLaboratoryAccreditationQuery{{Page: 1, PageSize: 1}, {Page: 1, PageSize: 1, AccreditationStatus: "accredited"}, {Page: 1, PageSize: 1, AccreditationStatus: "expired"}, {Page: 31, PageSize: 1}}
	if !reflect.DeepEqual(s.calls, want) || !reflect.DeepEqual(s.ids, []string{medicalLaboratoryFirstAnchor, medicalLaboratoryLastAnchor}) {
		t.Fatalf("calls %v %v", s.calls, s.ids)
	}
}
func TestAccreditationStartupRejectsBadPages(t *testing.T) {
	for _, tc := range []struct {
		name, key string
		mutate    func(*interfaces.MedicalLaboratoryAccreditationListResult)
	}{
		{"wrong total", "", func(p *interfaces.MedicalLaboratoryAccreditationListResult) { p.Total = 31; p.TotalPages = 31 }},
		{"wrong accredited", "accredited", func(p *interfaces.MedicalLaboratoryAccreditationListResult) { p.Total = 25; p.TotalPages = 25 }},
		{"wrong expired", "expired", func(p *interfaces.MedicalLaboratoryAccreditationListResult) { p.Total = 5; p.TotalPages = 5 }},
		{"wrong pages", "", func(p *interfaces.MedicalLaboratoryAccreditationListResult) { p.TotalPages = 1 }},
		{"nil first", "", func(p *interfaces.MedicalLaboratoryAccreditationListResult) { p.Records = nil }},
		{"empty first", "", func(p *interfaces.MedicalLaboratoryAccreditationListResult) {
			p.Records = []models.MedicalLaboratoryAccreditation{}
		}},
		{"too many", "", func(p *interfaces.MedicalLaboratoryAccreditationListResult) {
			p.Records = append(p.Records, p.Records[0])
		}},
		{"invalid record", "", func(p *interfaces.MedicalLaboratoryAccreditationListResult) { p.Records[0].ID = "invalid_id" }},
		{"expired hidden", "expired", func(p *interfaces.MedicalLaboratoryAccreditationListResult) {
			p.Records[0].AccreditationStatus = "accredited"
		}},
		{"beyond nonempty", "beyond", func(p *interfaces.MedicalLaboratoryAccreditationListResult) {
			p.Records = []models.MedicalLaboratoryAccreditation{{ID: "unexpected"}}
		}},
		{"beyond nil", "beyond", func(p *interfaces.MedicalLaboratoryAccreditationListResult) { p.Records = nil }},
		{"beyond total", "beyond", func(p *interfaces.MedicalLaboratoryAccreditationListResult) { p.Total = 0 }},
		{"beyond pages", "beyond", func(p *interfaces.MedicalLaboratoryAccreditationListResult) { p.TotalPages = 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := validStartupAccreditation()
			p := s.pages[tc.key]
			tc.mutate(&p)
			s.pages[tc.key] = p
			if err := verifyMedicalLaboratoryAccreditations(context.Background(), s); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("%v", err)
			}
		})
	}
	for _, id := range []string{medicalLaboratoryFirstAnchor, medicalLaboratoryLastAnchor} {
		s := validStartupAccreditation()
		delete(s.anchors, id)
		if err := verifyMedicalLaboratoryAccreditations(context.Background(), s); err == nil {
			t.Fatal("missing anchor accepted")
		}
	}
	s := validStartupAccreditation()
	p := s.pages[""]
	p.Total = 31
	p.TotalPages = 31
	s.pages[""] = p
	if err := verifyMedicalLaboratoryAccreditations(context.Background(), s); err == nil || !strings.Contains(err.Error(), "arithmetic") {
		t.Fatalf("arithmetic %v", err)
	}
	s = validStartupAccreditation()
	p = s.pages["accredited"]
	p.Total = 25
	p.TotalPages = 25
	s.pages["accredited"] = p
	p = s.pages["expired"]
	p.Total = 5
	p.TotalPages = 5
	s.pages["expired"] = p
	if err := verifyMedicalLaboratoryAccreditations(context.Background(), s); err == nil {
		t.Fatal("incorrect counts with valid arithmetic accepted")
	}
}
func TestAccreditationStartupRecordValidation(t *testing.T) {
	for name, mutate := range map[string]func(*models.MedicalLaboratoryAccreditation){
		"id": func(r *models.MedicalLaboratoryAccreditation) { r.ID = "bad_id" }, "long id": func(r *models.MedicalLaboratoryAccreditation) { r.ID = strings.Repeat("a", 256) }, "name": func(r *models.MedicalLaboratoryAccreditation) { r.Name = " " }, "state": func(r *models.MedicalLaboratoryAccreditation) { r.StateID = "unknown" }, "country": func(r *models.MedicalLaboratoryAccreditation) { r.CountryCode = "US" }, "status": func(r *models.MedicalLaboratoryAccreditation) { r.AccreditationStatus = "licensed" }, "number": func(r *models.MedicalLaboratoryAccreditation) { r.AccreditationNumber = " " }, "address": func(r *models.MedicalLaboratoryAccreditation) { r.Address = " " }, "approval": func(r *models.MedicalLaboratoryAccreditation) { r.ApprovalDate = "2026-02-30" }, "expiry": func(r *models.MedicalLaboratoryAccreditation) { r.ExpiryDate = "2026-1-01" },
	} {
		t.Run(name, func(t *testing.T) {
			r := validStartupAccreditation().anchors[medicalLaboratoryFirstAnchor]
			mutate(&r)
			if err := validateStartupMedicalLaboratoryAccreditation(r); err == nil {
				t.Fatal("invalid record accepted")
			}
		})
	}
	r := validStartupAccreditation().anchors[medicalLaboratoryFirstAnchor]
	r.AccreditationNumber = "ML0018"
	r.Address = "Institutional address"
	r.ApprovalDate = "2024-02-29"
	r.ExpiryDate = "2028-02-29"
	if err := validateStartupMedicalLaboratoryAccreditation(r); err != nil {
		t.Fatal(err)
	}
	typ := reflect.TypeOf(r)
	want := []string{"id", "name", "state_id", "country_code", "accreditation_status", "accreditation_number", "approval_date", "expiry_date", "address"}
	if typ.NumField() != len(want) {
		t.Fatal("unexpected fields, privacy boundary changed")
	}
	for i, name := range want {
		if strings.Split(typ.Field(i).Tag.Get("json"), ",")[0] != name {
			t.Fatal("unexpected model field")
		}
	}
}
func TestAccreditationStartupErrors(t *testing.T) {
	if err := verifyMedicalLaboratoryAccreditations(context.Background(), nil); err == nil {
		t.Fatal("nil service")
	}
	for _, deadline := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		if deadline {
			cancel()
			ctx, cancel = context.WithDeadline(context.Background(), time.Unix(0, 0))
		} else {
			cancel()
		}
		s := validStartupAccreditation()
		if err := verifyMedicalLaboratoryAccreditations(ctx, s); !errors.Is(err, ctx.Err()) || len(s.calls) != 0 {
			t.Fatalf("context %v", err)
		}
		cancel()
	}
	for step := 1; step <= 6; step++ {
		for _, cause := range []error{context.Canceled, context.DeadlineExceeded, errors.New("secret source decoder path")} {
			s := validStartupAccreditation()
			s.failAt = step
			s.err = fmt.Errorf("secret wrapper: %w", cause)
			err := verifyMedicalLaboratoryAccreditations(context.Background(), s)
			if errors.Is(cause, context.Canceled) || errors.Is(cause, context.DeadlineExceeded) {
				if !errors.Is(err, cause) {
					t.Fatalf("context lost: %v", err)
				}
			} else if err == nil || strings.Contains(err.Error(), "secret") {
				t.Fatalf("unsanitized %v", err)
			}
		}
	}
}

type accreditationCountingJSON struct {
	interfaces.JSONFileRepository
	counts map[string]int
}

func (s *accreditationCountingJSON) Decode(ctx context.Context, path string, dst any) error {
	s.counts[path]++
	return s.JSONFileRepository.Decode(ctx, path, dst)
}
func TestAccreditationProductionBuilderReusesCache(t *testing.T) {
	base, err := fileRepo.NewEmbeddedJSONRepository(datasets.Files(), 64<<20)
	if err != nil {
		t.Fatal(err)
	}
	counting := &accreditationCountingJSON{JSONFileRepository: base, counts: map[string]int{}}
	h, err := buildMedicalLaboratoryAccreditationHandler(context.Background(), counting)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{medicalLaboratoryFirstAnchor, medicalLaboratoryLastAnchor} {
		r := httptest.NewRequest("GET", "/", nil)
		r.SetPathValue("accreditation_id", id)
		w := httptest.NewRecorder()
		h.GetMedicalLaboratoryAccreditation(w, r)
		if w.Code != 200 {
			t.Fatalf("anchor %d %s", w.Code, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	h.ListMedicalLaboratoryAccreditations(w, httptest.NewRequest("GET", "/?accreditation_status=expired", nil))
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if !reflect.DeepEqual(counting.counts, map[string]int{medicalLaboratoryAccreditationsPath: 1, geographyStatesRelativePath: 1}) {
		t.Fatalf("duplicate or unexpected reads: %v", counting.counts)
	}
	if _, err := datasets.Files().Open("metadata/healthcare/medical_laboratory_accreditations_reconciliation/index.json"); err == nil {
		t.Fatal("reconciliation embedded")
	}
	if _, err := buildMedicalLaboratoryAccreditationHandler(context.Background(), nil); err == nil {
		t.Fatal("nil repository accepted")
	}
}
