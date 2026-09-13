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

func TestNHIAAccreditedHMOServiceBuilderConstructsWithoutEagerDecode(t *testing.T) {
	jsonRepository := &nhiaHMOAppJSONStub{}
	var gotPath string
	service, err := buildNHIAAccreditedHMOServiceFromJSONRepository(
		context.Background(),
		jsonRepository,
		func(repository interfaces.JSONFileRepository, recordsPath string) (interfaces.NHIAAccreditedHealthMaintenanceOrganisationRepository, error) {
			if repository != jsonRepository {
				t.Fatal("unexpected json repository")
			}
			gotPath = recordsPath
			return &nhiaHMOAppRepositoryStub{}, nil
		},
		func(repository interfaces.NHIAAccreditedHealthMaintenanceOrganisationRepository) (nhiaAccreditedHMOService, error) {
			return services.NewNHIAAccreditedHealthMaintenanceOrganisationService(repository)
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if service == nil || gotPath != healthcareNHIAAccreditedHealthMaintenanceOrganisationsPath || jsonRepository.decodeCalls != 0 {
		t.Fatalf("unexpected builder state: service=%T path=%q decodes=%d", service, gotPath, jsonRepository.decodeCalls)
	}
}

func TestNHIAAccreditedHMOServiceBuilderErrors(t *testing.T) {
	validRepositoryFactory := func(interfaces.JSONFileRepository, string) (interfaces.NHIAAccreditedHealthMaintenanceOrganisationRepository, error) {
		return &nhiaHMOAppRepositoryStub{}, nil
	}
	validServiceFactory := func(repository interfaces.NHIAAccreditedHealthMaintenanceOrganisationRepository) (nhiaAccreditedHMOService, error) {
		return services.NewNHIAAccreditedHealthMaintenanceOrganisationService(repository)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := buildNHIAAccreditedHMOServiceFromJSONRepository(ctx, &nhiaHMOAppJSONStub{}, validRepositoryFactory, validServiceFactory); !errors.Is(err, context.Canceled) {
		t.Fatalf("context error = %v", err)
	}
	if _, err := buildNHIAAccreditedHMOServiceFromJSONRepository(context.Background(), nil, validRepositoryFactory, validServiceFactory); err == nil {
		t.Fatal("nil json repository accepted")
	}
	if _, err := buildNHIAAccreditedHMOServiceFromJSONRepository(context.Background(), &nhiaHMOAppJSONStub{}, nil, validServiceFactory); err == nil {
		t.Fatal("nil repository factory accepted")
	}
	if _, err := buildNHIAAccreditedHMOServiceFromJSONRepository(context.Background(), &nhiaHMOAppJSONStub{}, validRepositoryFactory, nil); err == nil {
		t.Fatal("nil service factory accepted")
	}
}

func TestNHIAAccreditedHMOAppDependencyTypeKeepsNoHTTPContract(t *testing.T) {
	depsType := reflect.TypeOf(appDependencies{})
	if _, ok := depsType.FieldByName("nhiaHMOService"); !ok {
		t.Fatal("NHIA HMO service dependency missing")
	}
}

type startupNHIAHMOStub struct {
	pages   map[string]interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult
	anchors map[string]models.NHIAAccreditedHealthMaintenanceOrganisation
	calls   []interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery
	ids     []string
	failAt  int
	err     error
}

func (s *startupNHIAHMOStub) ListNHIAAccreditedHealthMaintenanceOrganisations(_ context.Context, q interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery) (interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult, error) {
	s.calls = append(s.calls, q)
	if len(s.calls)+len(s.ids) == s.failAt {
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{}, s.err
	}
	key := q.AccreditationStatus
	if q.HMOID != "" {
		key = "hmo_id"
	}
	if q.Page == nhiaAccreditedHMOExpectedCount+1 {
		key = "beyond"
	}
	return s.pages[key], nil
}

func (s *startupNHIAHMOStub) GetNHIAAccreditedHealthMaintenanceOrganisation(_ context.Context, id string) (models.NHIAAccreditedHealthMaintenanceOrganisation, error) {
	s.ids = append(s.ids, id)
	if len(s.calls)+len(s.ids) == s.failAt {
		return models.NHIAAccreditedHealthMaintenanceOrganisation{}, s.err
	}
	record, ok := s.anchors[id]
	if !ok {
		return record, errors.New("missing anchor")
	}
	return record, nil
}

func validStartupNHIAHMO() *startupNHIAHMOStub {
	first := models.NHIAAccreditedHealthMaintenanceOrganisation{ID: nhiaAccreditedHMOFirstAnchor, Name: nhiaAccreditedHMOAnchorName, CountryCode: "NG", OrganisationType: "health_maintenance_organisation", AccreditationStatus: "accredited", HMOID: nhiaAccreditedHMOAnchorHMOID}
	last := first
	last.ID = nhiaAccreditedHMOLastAnchor
	last.Name = "ZUMA HEALTH TRUST"
	last.HMOID = "28"
	s := &startupNHIAHMOStub{
		pages:   map[string]interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{},
		anchors: map[string]models.NHIAAccreditedHealthMaintenanceOrganisation{first.ID: first, last.ID: last},
	}
	for _, key := range []string{"", "accredited"} {
		s.pages[key] = interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{Records: []models.NHIAAccreditedHealthMaintenanceOrganisation{first}, Page: 1, PageSize: 1, Total: 94, TotalPages: 94}
	}
	s.pages["hmo_id"] = interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{Records: []models.NHIAAccreditedHealthMaintenanceOrganisation{first}, Page: 1, PageSize: 1, Total: 1, TotalPages: 1}
	s.pages["beyond"] = interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{Records: []models.NHIAAccreditedHealthMaintenanceOrganisation{}, Page: 95, PageSize: 1, Total: 94, TotalPages: 94}
	return s
}

func TestNHIAAccreditedHMOStartupBoundedQueries(t *testing.T) {
	s := validStartupNHIAHMO()
	if err := verifyNHIAAccreditedHMOs(nil, s); err != nil {
		t.Fatal(err)
	}
	want := []interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{
		{Page: 1, PageSize: 1},
		{Page: 1, PageSize: 1, AccreditationStatus: "accredited"},
		{Page: 1, PageSize: 1, HMOID: nhiaAccreditedHMOAnchorHMOID},
		{Page: 95, PageSize: 1},
	}
	if !reflect.DeepEqual(s.calls, want) || !reflect.DeepEqual(s.ids, []string{nhiaAccreditedHMOFirstAnchor, nhiaAccreditedHMOLastAnchor}) {
		t.Fatalf("calls %v ids %v", s.calls, s.ids)
	}
}

func TestNHIAAccreditedHMOStartupRejectsBadPagesAndRecords(t *testing.T) {
	for _, tc := range []struct {
		name, key string
		mutate    func(*interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult)
	}{
		{"wrong total", "", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) {
			p.Total = 93
			p.TotalPages = 93
		}},
		{"wrong status total", "accredited", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) {
			p.Total = 93
			p.TotalPages = 93
		}},
		{"wrong pages", "", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) { p.TotalPages = 1 }},
		{"nil first", "", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) { p.Records = nil }},
		{"empty first", "", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) {
			p.Records = []models.NHIAAccreditedHealthMaintenanceOrganisation{}
		}},
		{"too many", "", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) {
			p.Records = append(p.Records, p.Records[0])
		}},
		{"invalid status", "", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) {
			p.Records[0].AccreditationStatus = "licensed"
		}},
		{"beyond nonempty", "beyond", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) {
			p.Records = []models.NHIAAccreditedHealthMaintenanceOrganisation{{ID: "unexpected"}}
		}},
		{"beyond nil", "beyond", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) { p.Records = nil }},
		{"hmo id zero matches", "hmo_id", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) {
			p.Records = []models.NHIAAccreditedHealthMaintenanceOrganisation{}
		}},
		{"hmo id too many matches", "hmo_id", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) {
			p.Records = append(p.Records, p.Records[0])
			p.Total = 2
		}},
		{"hmo id wrong total", "hmo_id", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) { p.Total = 2 }},
		{"hmo id wrong public id", "hmo_id", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) {
			p.Records[0].ID = nhiaAccreditedHMOLastAnchor
		}},
		{"hmo id wrong hmo id", "hmo_id", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) {
			p.Records[0].HMOID = "103"
		}},
		{"hmo id wrong country", "hmo_id", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) {
			p.Records[0].CountryCode = "GH"
		}},
		{"hmo id wrong type", "hmo_id", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) {
			p.Records[0].OrganisationType = "insurer"
		}},
		{"hmo id wrong status", "hmo_id", func(p *interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult) {
			p.Records[0].AccreditationStatus = "registered"
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := validStartupNHIAHMO()
			page := s.pages[tc.key]
			tc.mutate(&page)
			s.pages[tc.key] = page
			if err := verifyNHIAAccreditedHMOs(context.Background(), s); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("error = %v", err)
			}
		})
	}
	for name, mutate := range map[string]func(*models.NHIAAccreditedHealthMaintenanceOrganisation){
		"id": func(r *models.NHIAAccreditedHealthMaintenanceOrganisation) { r.ID = "bad_id" },
		"name": func(r *models.NHIAAccreditedHealthMaintenanceOrganisation) {
			r.Name = " "
		},
		"country": func(r *models.NHIAAccreditedHealthMaintenanceOrganisation) { r.CountryCode = "US" },
		"type":    func(r *models.NHIAAccreditedHealthMaintenanceOrganisation) { r.OrganisationType = "insurer" },
		"status":  func(r *models.NHIAAccreditedHealthMaintenanceOrganisation) { r.AccreditationStatus = "expired" },
		"hmo_id":  func(r *models.NHIAAccreditedHealthMaintenanceOrganisation) { r.HMOID = " " },
	} {
		t.Run(name, func(t *testing.T) {
			record := validStartupNHIAHMO().anchors[nhiaAccreditedHMOFirstAnchor]
			mutate(&record)
			if err := validateStartupNHIAAccreditedHMO(record); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestNHIAAccreditedHMOStartupErrors(t *testing.T) {
	if err := verifyNHIAAccreditedHMOs(context.Background(), nil); err == nil {
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
		s := validStartupNHIAHMO()
		if err := verifyNHIAAccreditedHMOs(ctx, s); !errors.Is(err, ctx.Err()) || len(s.calls) != 0 {
			t.Fatalf("context %v", err)
		}
		cancel()
	}
	for step := 1; step <= 6; step++ {
		for _, cause := range []error{context.Canceled, context.DeadlineExceeded, errors.New("secret decoder path")} {
			s := validStartupNHIAHMO()
			s.failAt = step
			s.err = fmt.Errorf("secret wrapper: %w", cause)
			err := verifyNHIAAccreditedHMOs(context.Background(), s)
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

type nhiaHMOCountingJSON struct {
	interfaces.JSONFileRepository
	counts map[string]int
}

func (s *nhiaHMOCountingJSON) Decode(ctx context.Context, path string, dst any) error {
	s.counts[path]++
	return s.JSONFileRepository.Decode(ctx, path, dst)
}

func TestNHIAAccreditedHMOProductionBuilderVerifiesAndReusesCache(t *testing.T) {
	base, err := fileRepo.NewEmbeddedJSONRepository(datasets.Files(), 64<<20)
	if err != nil {
		t.Fatal(err)
	}
	counting := &nhiaHMOCountingJSON{JSONFileRepository: base, counts: map[string]int{}}
	service, handler, err := buildNHIAAccreditedHMOHandler(
		context.Background(),
		counting,
		func(repository interfaces.JSONFileRepository, recordsPath string) (interfaces.NHIAAccreditedHealthMaintenanceOrganisationRepository, error) {
			return fileRepo.NewNHIAAccreditedHealthMaintenanceOrganisationRepository(repository, recordsPath)
		},
		func(repository interfaces.NHIAAccreditedHealthMaintenanceOrganisationRepository) (nhiaAccreditedHMOService, error) {
			return services.NewNHIAAccreditedHealthMaintenanceOrganisationService(repository)
		},
		func(service nhiaAccreditedHMOService) (*handlers.NHIAAccreditedHealthMaintenanceOrganisationHandler, error) {
			return handlers.NewNHIAAccreditedHealthMaintenanceOrganisationHandler(service)
		},
	)
	if err != nil || service == nil || handler == nil {
		t.Fatalf("builder service=%T handler=%T err=%v", service, handler, err)
	}
	if !reflect.DeepEqual(counting.counts, map[string]int{healthcareNHIAAccreditedHealthMaintenanceOrganisationsPath: 1}) {
		t.Fatalf("unexpected reads: %v", counting.counts)
	}
	if _, err := datasets.Files().Open("metadata/healthcare/nhia_accredited_health_maintenance_organisations_reconciliation/index.json"); err == nil {
		t.Fatal("reconciliation embedded")
	}
}

type nhiaHMOAppJSONStub struct {
	decodeCalls int
}

func (s *nhiaHMOAppJSONStub) Decode(context.Context, string, any) error {
	s.decodeCalls++
	return errors.New("decode should not be called during construction")
}

type nhiaHMOAppRepositoryStub struct{}

func (s *nhiaHMOAppRepositoryStub) ListNHIAAccreditedHealthMaintenanceOrganisations(_ context.Context, query interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery) (interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult, error) {
	return interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{
		Records:  make([]models.NHIAAccreditedHealthMaintenanceOrganisation, 0),
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (s *nhiaHMOAppRepositoryStub) GetNHIAAccreditedHealthMaintenanceOrganisation(context.Context, string) (models.NHIAAccreditedHealthMaintenanceOrganisation, error) {
	return models.NHIAAccreditedHealthMaintenanceOrganisation{}, interfaces.ErrNHIAAccreditedHealthMaintenanceOrganisationNotFound
}
