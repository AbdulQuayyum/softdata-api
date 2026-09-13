package app

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/datasets"
	"github.com/AbdulQuayyum/softdata-api/internal/handlers"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	fileRepo "github.com/AbdulQuayyum/softdata-api/internal/repository/file"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

func TestNHIAStateSocialHealthInsuranceAgencyServiceBuilderConstructsWithoutEagerDecode(t *testing.T) {
	jsonRepository := &nhiaSSHIAAppJSONStub{}
	var gotPath string
	service, err := buildNHIASSHIAServiceFromJSONRepository(
		context.Background(),
		jsonRepository,
		func(repository interfaces.JSONFileRepository, recordsPath string) (interfaces.NHIAStateSocialHealthInsuranceAgencyRepository, error) {
			if repository != jsonRepository {
				t.Fatal("unexpected json repository")
			}
			gotPath = recordsPath
			return &nhiaSSHIAAppRepositoryStub{}, nil
		},
		func(repository interfaces.NHIAStateSocialHealthInsuranceAgencyRepository) (nhiaSSHIAService, error) {
			return services.NewNHIAStateSocialHealthInsuranceAgencyService(repository)
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if service == nil || gotPath != healthcareNHIAStateSocialHealthInsuranceAgenciesPath || jsonRepository.decodeCalls != 0 {
		t.Fatalf("unexpected builder state: service=%T path=%q decodes=%d", service, gotPath, jsonRepository.decodeCalls)
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyServiceBuilderErrors(t *testing.T) {
	validRepositoryFactory := func(interfaces.JSONFileRepository, string) (interfaces.NHIAStateSocialHealthInsuranceAgencyRepository, error) {
		return &nhiaSSHIAAppRepositoryStub{}, nil
	}
	validServiceFactory := func(repository interfaces.NHIAStateSocialHealthInsuranceAgencyRepository) (nhiaSSHIAService, error) {
		return services.NewNHIAStateSocialHealthInsuranceAgencyService(repository)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := buildNHIASSHIAServiceFromJSONRepository(ctx, &nhiaSSHIAAppJSONStub{}, validRepositoryFactory, validServiceFactory); !errors.Is(err, context.Canceled) {
		t.Fatalf("context error = %v", err)
	}
	if _, err := buildNHIASSHIAServiceFromJSONRepository(context.Background(), nil, validRepositoryFactory, validServiceFactory); err == nil {
		t.Fatal("nil json repository accepted")
	}
	if _, err := buildNHIASSHIAServiceFromJSONRepository(context.Background(), &nhiaSSHIAAppJSONStub{}, nil, validServiceFactory); err == nil {
		t.Fatal("nil repository factory accepted")
	}
	if _, err := buildNHIASSHIAServiceFromJSONRepository(context.Background(), &nhiaSSHIAAppJSONStub{}, validRepositoryFactory, nil); err == nil {
		t.Fatal("nil service factory accepted")
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyAppDependencyTypeKeepsService(t *testing.T) {
	depsType := reflect.TypeOf(appDependencies{})
	if _, ok := depsType.FieldByName("nhiaSSHIAService"); !ok {
		t.Fatal("NHIA SSHIA service dependency missing")
	}
}

type startupNHIASSHIAStub struct {
	pages   map[string]interfaces.NHIAStateSocialHealthInsuranceAgencyListResult
	anchors map[string]models.NHIAStateSocialHealthInsuranceAgency
	calls   []interfaces.NHIAStateSocialHealthInsuranceAgencyQuery
	ids     []string
	failAt  int
	err     error
}

func (s *startupNHIASSHIAStub) ListNHIAStateSocialHealthInsuranceAgencies(_ context.Context, q interfaces.NHIAStateSocialHealthInsuranceAgencyQuery) (interfaces.NHIAStateSocialHealthInsuranceAgencyListResult, error) {
	s.calls = append(s.calls, q)
	if len(s.calls)+len(s.ids) == s.failAt {
		return interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{}, s.err
	}
	key := q.StateID
	if key == "" {
		key = "default"
	}
	if q.Page == nhiaSSHIAExpectedCount+1 {
		key = "beyond"
	}
	return s.pages[key], nil
}

func (s *startupNHIASSHIAStub) GetNHIAStateSocialHealthInsuranceAgency(_ context.Context, id string) (models.NHIAStateSocialHealthInsuranceAgency, error) {
	s.ids = append(s.ids, id)
	if len(s.calls)+len(s.ids) == s.failAt {
		return models.NHIAStateSocialHealthInsuranceAgency{}, s.err
	}
	record, ok := s.anchors[id]
	if !ok {
		return record, errors.New("missing anchor")
	}
	return record, nil
}

func validStartupNHIASSHIA() *startupNHIASSHIAStub {
	first := models.NHIAStateSocialHealthInsuranceAgency{ID: nhiaSSHIAFirstAnchor, Name: nhiaSSHIAFirstAnchorName, StateID: "abia", CountryCode: "NG", OrganisationType: nhiaSSHIAOrganisationType}
	last := models.NHIAStateSocialHealthInsuranceAgency{ID: nhiaSSHIALastAnchor, Name: nhiaSSHIALastAnchorName, StateID: "zamfara", CountryCode: "NG", OrganisationType: nhiaSSHIAOrganisationType}
	lagos := models.NHIAStateSocialHealthInsuranceAgency{ID: nhiaSSHIACheckStateAnchor, Name: nhiaSSHIACheckStateName, StateID: nhiaSSHIACheckState, CountryCode: "NG", OrganisationType: nhiaSSHIAOrganisationType}
	fct := models.NHIAStateSocialHealthInsuranceAgency{ID: nhiaSSHIAFCTAnchor, Name: nhiaSSHIAFCTAnchorName, StateID: "fct", CountryCode: "NG", OrganisationType: nhiaSSHIAOrganisationType}
	return &startupNHIASSHIAStub{
		pages: map[string]interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{
			"default":           {Records: []models.NHIAStateSocialHealthInsuranceAgency{first}, Page: 1, PageSize: 1, Total: 37, TotalPages: 37},
			nhiaSSHIACheckState: {Records: []models.NHIAStateSocialHealthInsuranceAgency{lagos}, Page: 1, PageSize: 1, Total: 1, TotalPages: 1},
			"fct":               {Records: []models.NHIAStateSocialHealthInsuranceAgency{fct}, Page: 1, PageSize: 1, Total: 1, TotalPages: 1},
			"beyond":            {Records: []models.NHIAStateSocialHealthInsuranceAgency{}, Page: 38, PageSize: 1, Total: 37, TotalPages: 37},
		},
		anchors: map[string]models.NHIAStateSocialHealthInsuranceAgency{first.ID: first, last.ID: last},
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyStartupBoundedQueries(t *testing.T) {
	s := validStartupNHIASSHIA()
	if err := verifyNHIASSHIAs(nil, s); err != nil {
		t.Fatal(err)
	}
	wantCalls := []interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{
		{Page: 1, PageSize: 1},
		{Page: 1, PageSize: 1, StateID: nhiaSSHIACheckState},
		{Page: 1, PageSize: 1, StateID: "fct"},
		{Page: 38, PageSize: 1},
	}
	if !reflect.DeepEqual(s.calls, wantCalls) || !reflect.DeepEqual(s.ids, []string{nhiaSSHIAFirstAnchor, nhiaSSHIALastAnchor}) {
		t.Fatalf("calls %v ids %v", s.calls, s.ids)
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyStartupRejectsBadPagesAndRecords(t *testing.T) {
	for _, tc := range []struct {
		name, key string
		mutate    func(*interfaces.NHIAStateSocialHealthInsuranceAgencyListResult)
	}{
		{"wrong total", "default", func(p *interfaces.NHIAStateSocialHealthInsuranceAgencyListResult) { p.Total = 36 }},
		{"wrong total pages", "default", func(p *interfaces.NHIAStateSocialHealthInsuranceAgencyListResult) { p.TotalPages = 1 }},
		{"nil first", "default", func(p *interfaces.NHIAStateSocialHealthInsuranceAgencyListResult) { p.Records = nil }},
		{"empty first", "default", func(p *interfaces.NHIAStateSocialHealthInsuranceAgencyListResult) {
			p.Records = []models.NHIAStateSocialHealthInsuranceAgency{}
		}},
		{"invalid first country", "default", func(p *interfaces.NHIAStateSocialHealthInsuranceAgencyListResult) { p.Records[0].CountryCode = "GH" }},
		{"invalid first type", "default", func(p *interfaces.NHIAStateSocialHealthInsuranceAgencyListResult) {
			p.Records[0].OrganisationType = "agency"
		}},
		{"state zero matches", nhiaSSHIACheckState, func(p *interfaces.NHIAStateSocialHealthInsuranceAgencyListResult) {
			p.Records = []models.NHIAStateSocialHealthInsuranceAgency{}
		}},
		{"state too many matches", nhiaSSHIACheckState, func(p *interfaces.NHIAStateSocialHealthInsuranceAgencyListResult) {
			p.Records = append(p.Records, p.Records[0])
		}},
		{"state wrong total", nhiaSSHIACheckState, func(p *interfaces.NHIAStateSocialHealthInsuranceAgencyListResult) { p.Total = 2 }},
		{"state wrong id", nhiaSSHIACheckState, func(p *interfaces.NHIAStateSocialHealthInsuranceAgencyListResult) {
			p.Records[0].ID = nhiaSSHIAFCTAnchor
		}},
		{"state wrong state", nhiaSSHIACheckState, func(p *interfaces.NHIAStateSocialHealthInsuranceAgencyListResult) { p.Records[0].StateID = "fct" }},
		{"fct wrong state", "fct", func(p *interfaces.NHIAStateSocialHealthInsuranceAgencyListResult) { p.Records[0].StateID = "lagos" }},
		{"beyond nonempty", "beyond", func(p *interfaces.NHIAStateSocialHealthInsuranceAgencyListResult) {
			p.Records = []models.NHIAStateSocialHealthInsuranceAgency{{ID: "unexpected"}}
		}},
		{"beyond nil", "beyond", func(p *interfaces.NHIAStateSocialHealthInsuranceAgencyListResult) { p.Records = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := validStartupNHIASSHIA()
			page := s.pages[tc.key]
			tc.mutate(&page)
			s.pages[tc.key] = page
			if err := verifyNHIASSHIAs(context.Background(), s); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("error = %v", err)
			}
		})
	}
	for name, mutate := range map[string]func(*models.NHIAStateSocialHealthInsuranceAgency){
		"id":      func(r *models.NHIAStateSocialHealthInsuranceAgency) { r.ID = "bad_id" },
		"name":    func(r *models.NHIAStateSocialHealthInsuranceAgency) { r.Name = " " },
		"state":   func(r *models.NHIAStateSocialHealthInsuranceAgency) { r.StateID = " " },
		"country": func(r *models.NHIAStateSocialHealthInsuranceAgency) { r.CountryCode = "US" },
		"type":    func(r *models.NHIAStateSocialHealthInsuranceAgency) { r.OrganisationType = "insurer" },
	} {
		t.Run(name, func(t *testing.T) {
			record := validStartupNHIASSHIA().anchors[nhiaSSHIAFirstAnchor]
			mutate(&record)
			if err := validateStartupNHIASSHIA(record); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("error = %v", err)
			}
		})
	}
	for name, mutate := range map[string]func(*models.NHIAStateSocialHealthInsuranceAgency){
		"id":      func(r *models.NHIAStateSocialHealthInsuranceAgency) { r.ID = "wrong-anchor" },
		"name":    func(r *models.NHIAStateSocialHealthInsuranceAgency) { r.Name = "Wrong" },
		"country": func(r *models.NHIAStateSocialHealthInsuranceAgency) { r.CountryCode = "GH" },
		"type":    func(r *models.NHIAStateSocialHealthInsuranceAgency) { r.OrganisationType = "scheme" },
	} {
		t.Run("anchor "+name, func(t *testing.T) {
			s := validStartupNHIASSHIA()
			record := s.anchors[nhiaSSHIAFirstAnchor]
			mutate(&record)
			s.anchors[nhiaSSHIAFirstAnchor] = record
			if err := verifyNHIASSHIAs(context.Background(), s); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyStartupErrors(t *testing.T) {
	if err := verifyNHIASSHIAs(context.Background(), nil); err == nil {
		t.Fatal("nil service accepted")
	}
	for _, deadline := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		if deadline {
			cancel()
			ctx, cancel = context.WithDeadline(context.Background(), time.Unix(0, 0))
		} else {
			cancel()
		}
		s := validStartupNHIASSHIA()
		if err := verifyNHIASSHIAs(ctx, s); !errors.Is(err, ctx.Err()) || len(s.calls) != 0 {
			t.Fatalf("context %v", err)
		}
		cancel()
	}
	for step := 1; step <= 6; step++ {
		for _, cause := range []error{context.Canceled, context.DeadlineExceeded, errors.New("secret decoder path")} {
			s := validStartupNHIASSHIA()
			s.failAt = step
			s.err = fmt.Errorf("secret wrapper: %w", cause)
			err := verifyNHIASSHIAs(context.Background(), s)
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

type nhiaSSHIACountingJSON struct {
	interfaces.JSONFileRepository
	counts map[string]int
}

func (s *nhiaSSHIACountingJSON) Decode(ctx context.Context, path string, dst any) error {
	s.counts[path]++
	return s.JSONFileRepository.Decode(ctx, path, dst)
}

func TestNHIAStateSocialHealthInsuranceAgencyProductionBuilderVerifiesAndReusesCache(t *testing.T) {
	base, err := fileRepo.NewEmbeddedJSONRepository(datasets.Files(), 64<<20)
	if err != nil {
		t.Fatal(err)
	}
	counting := &nhiaSSHIACountingJSON{JSONFileRepository: base, counts: map[string]int{}}
	service, handler, err := buildNHIASSHIAHandler(
		context.Background(),
		counting,
		func(repository interfaces.JSONFileRepository, recordsPath string) (interfaces.NHIAStateSocialHealthInsuranceAgencyRepository, error) {
			return fileRepo.NewNHIAStateSocialHealthInsuranceAgencyRepository(repository, recordsPath)
		},
		func(repository interfaces.NHIAStateSocialHealthInsuranceAgencyRepository) (nhiaSSHIAService, error) {
			return services.NewNHIAStateSocialHealthInsuranceAgencyService(repository)
		},
		func(service nhiaSSHIAService) (*handlers.NHIAStateSocialHealthInsuranceAgencyHandler, error) {
			return handlers.NewNHIAStateSocialHealthInsuranceAgencyHandler(service)
		},
	)
	if err != nil || service == nil || handler == nil {
		t.Fatalf("builder service=%T handler=%T err=%v", service, handler, err)
	}
	if _, err := service.GetNHIAStateSocialHealthInsuranceAgency(context.Background(), nhiaSSHIAFirstAnchor); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(counting.counts, map[string]int{healthcareNHIAStateSocialHealthInsuranceAgenciesPath: 1}) {
		t.Fatalf("unexpected reads: %v", counting.counts)
	}
	if _, err := datasets.Files().Open("metadata/healthcare/nhia_state_social_health_insurance_agencies_reconciliation/index.json"); err == nil {
		t.Fatal("reconciliation embedded")
	}
}

type nhiaSSHIAAppJSONStub struct {
	decodeCalls int
}

func (s *nhiaSSHIAAppJSONStub) Decode(context.Context, string, any) error {
	s.decodeCalls++
	return errors.New("decode should not be called during construction")
}

type nhiaSSHIAAppRepositoryStub struct{}

func (s *nhiaSSHIAAppRepositoryStub) ListNHIAStateSocialHealthInsuranceAgencies(_ context.Context, query interfaces.NHIAStateSocialHealthInsuranceAgencyQuery) (interfaces.NHIAStateSocialHealthInsuranceAgencyListResult, error) {
	return interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{
		Records:  make([]models.NHIAStateSocialHealthInsuranceAgency, 0),
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (s *nhiaSSHIAAppRepositoryStub) GetNHIAStateSocialHealthInsuranceAgency(context.Context, string) (models.NHIAStateSocialHealthInsuranceAgency, error) {
	return models.NHIAStateSocialHealthInsuranceAgency{}, interfaces.ErrNHIAStateSocialHealthInsuranceAgencyNotFound
}
