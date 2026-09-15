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

func TestNHIAActiveAccreditedHealthcareProviderServiceBuilderConstructsWithoutEagerDecode(t *testing.T) {
	var gotPath string
	jsonRepository := &nhiaHCPAppJSONStub{}
	service, err := buildNHIAActiveAccreditedHealthcareProviderServiceFromJSONRepository(
		context.Background(),
		jsonRepository,
		func(repository interfaces.JSONFileRepository, recordsPath string) (interfaces.NHIAActiveAccreditedHealthcareProviderRepository, error) {
			if repository != jsonRepository {
				t.Fatal("unexpected json repository")
			}
			gotPath = recordsPath
			return &nhiaHCPAppRepositoryStub{}, nil
		},
		func(repository interfaces.NHIAActiveAccreditedHealthcareProviderRepository) (nhiaActiveAccreditedHealthcareProviderService, error) {
			return services.NewNHIAActiveAccreditedHealthcareProviderService(repository)
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if service == nil || gotPath != healthcareNHIAActiveAccreditedHealthcareProvidersPath || jsonRepository.decodeCalls != 0 {
		t.Fatalf("service=%v path=%q decode=%d", service, gotPath, jsonRepository.decodeCalls)
	}
}

func TestNHIAActiveAccreditedHealthcareProviderServiceBuilderErrors(t *testing.T) {
	validRepositoryFactory := func(interfaces.JSONFileRepository, string) (interfaces.NHIAActiveAccreditedHealthcareProviderRepository, error) {
		return &nhiaHCPAppRepositoryStub{}, nil
	}
	validServiceFactory := func(repository interfaces.NHIAActiveAccreditedHealthcareProviderRepository) (nhiaActiveAccreditedHealthcareProviderService, error) {
		return services.NewNHIAActiveAccreditedHealthcareProviderService(repository)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := buildNHIAActiveAccreditedHealthcareProviderServiceFromJSONRepository(ctx, &nhiaHCPAppJSONStub{}, validRepositoryFactory, validServiceFactory); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled err=%v", err)
	}
	if _, err := buildNHIAActiveAccreditedHealthcareProviderServiceFromJSONRepository(context.Background(), nil, validRepositoryFactory, validServiceFactory); err == nil {
		t.Fatal("expected nil json repository error")
	}
	if _, err := buildNHIAActiveAccreditedHealthcareProviderServiceFromJSONRepository(context.Background(), &nhiaHCPAppJSONStub{}, nil, validServiceFactory); err == nil {
		t.Fatal("expected nil repository factory error")
	}
	if _, err := buildNHIAActiveAccreditedHealthcareProviderServiceFromJSONRepository(context.Background(), &nhiaHCPAppJSONStub{}, validRepositoryFactory, nil); err == nil {
		t.Fatal("expected nil service factory error")
	}
	if _, err := buildNHIAActiveAccreditedHealthcareProviderServiceFromJSONRepository(context.Background(), &nhiaHCPAppJSONStub{}, func(interfaces.JSONFileRepository, string) (interfaces.NHIAActiveAccreditedHealthcareProviderRepository, error) {
		return nil, errors.New("boom")
	}, validServiceFactory); err == nil {
		t.Fatal("expected repository factory error")
	}
}

func TestNHIAActiveAccreditedHealthcareProviderAppDependencyTypeKeepsService(t *testing.T) {
	depsType := reflect.TypeOf(appDependencies{})
	if _, ok := depsType.FieldByName("nhiaHCPService"); !ok {
		t.Fatal("nhia HCP service was not retained")
	}
}

type startupNHIAHCPStub struct {
	pages   map[string]interfaces.NHIAActiveAccreditedHealthcareProviderListResult
	anchors map[string]models.NHIAActiveAccreditedHealthcareProvider
	calls   []interfaces.NHIAActiveAccreditedHealthcareProviderQuery
	ids     []string
	failAt  int
	err     error
}

func (s *startupNHIAHCPStub) ListNHIAActiveAccreditedHealthcareProviders(_ context.Context, q interfaces.NHIAActiveAccreditedHealthcareProviderQuery) (interfaces.NHIAActiveAccreditedHealthcareProviderListResult, error) {
	s.calls = append(s.calls, q)
	if len(s.calls)+len(s.ids) == s.failAt {
		return interfaces.NHIAActiveAccreditedHealthcareProviderListResult{}, s.err
	}
	key := "default"
	switch {
	case q.Page == nhiaHCPExpectedCount+1:
		key = "beyond"
	case q.ProviderCode != "":
		key = "provider_code"
	case q.FacilityType != "":
		key = "facility_type:" + q.FacilityType
	case q.ListingStatus != "":
		key = "listing_status"
	}
	return s.pages[key], nil
}

func (s *startupNHIAHCPStub) GetNHIAActiveAccreditedHealthcareProvider(_ context.Context, id string) (models.NHIAActiveAccreditedHealthcareProvider, error) {
	s.ids = append(s.ids, id)
	if len(s.calls)+len(s.ids) == s.failAt {
		return models.NHIAActiveAccreditedHealthcareProvider{}, s.err
	}
	record, ok := s.anchors[id]
	if !ok {
		return record, errors.New("missing anchor")
	}
	return record, nil
}

func validStartupNHIAHCP() *startupNHIAHCPStub {
	first := models.NHIAActiveAccreditedHealthcareProvider{ID: nhiaHCPFirstAnchor, Name: nhiaHCPFirstAnchorName, CountryCode: "NG", ProviderCode: nhiaHCPFirstAnchorCode, FacilityType: nhiaHCPFirstAnchorFacilityType, ListingStatus: nhiaHCPListingStatus}
	last := models.NHIAActiveAccreditedHealthcareProvider{ID: nhiaHCPLastAnchor, Name: nhiaHCPLastAnchorName, CountryCode: "NG", ProviderCode: nhiaHCPLastAnchorCode, FacilityType: nhiaHCPLastAnchorFacilityType, ListingStatus: nhiaHCPListingStatus}
	providerCode := models.NHIAActiveAccreditedHealthcareProvider{ID: nhiaHCPProviderCodeCheckID, Name: nhiaHCPProviderCodeCheckName, CountryCode: "NG", ProviderCode: nhiaHCPProviderCodeCheckCode, FacilityType: nhiaHCPProviderCodeCheckFacilityType, ListingStatus: nhiaHCPListingStatus}
	primary := providerCode
	secondary := first
	return &startupNHIAHCPStub{
		pages: map[string]interfaces.NHIAActiveAccreditedHealthcareProviderListResult{
			"default":       {Records: []models.NHIAActiveAccreditedHealthcareProvider{first}, Page: 1, PageSize: 1, Total: nhiaHCPExpectedCount, TotalPages: nhiaHCPExpectedCount},
			"provider_code": {Records: []models.NHIAActiveAccreditedHealthcareProvider{providerCode}, Page: 1, PageSize: 1, Total: 1, TotalPages: 1},
			"facility_type:" + nhiaHCPPrimaryFacilityType:             {Records: []models.NHIAActiveAccreditedHealthcareProvider{primary}, Page: 1, PageSize: 1, Total: nhiaHCPPrimaryCount, TotalPages: nhiaHCPPrimaryCount},
			"facility_type:" + nhiaHCPPrimaryAndSecondaryFacilityType: {Records: []models.NHIAActiveAccreditedHealthcareProvider{secondary}, Page: 1, PageSize: 1, Total: nhiaHCPPrimaryAndSecondaryCount, TotalPages: nhiaHCPPrimaryAndSecondaryCount},
			"listing_status": {Records: []models.NHIAActiveAccreditedHealthcareProvider{first}, Page: 1, PageSize: 1, Total: nhiaHCPExpectedCount, TotalPages: nhiaHCPExpectedCount},
			"beyond":         {Records: []models.NHIAActiveAccreditedHealthcareProvider{}, Page: nhiaHCPExpectedCount + 1, PageSize: 1, Total: nhiaHCPExpectedCount, TotalPages: nhiaHCPExpectedCount},
		},
		anchors: map[string]models.NHIAActiveAccreditedHealthcareProvider{first.ID: first, last.ID: last},
	}
}

func TestNHIAActiveAccreditedHealthcareProviderStartupBoundedQueries(t *testing.T) {
	s := validStartupNHIAHCP()
	if err := verifyNHIAActiveAccreditedHealthcareProviders(nil, s); err != nil {
		t.Fatal(err)
	}
	wantCalls := []interfaces.NHIAActiveAccreditedHealthcareProviderQuery{
		{Page: 1, PageSize: 1},
		{Page: 1, PageSize: 1, ProviderCode: nhiaHCPProviderCodeCheckCode},
		{Page: 1, PageSize: 1, FacilityType: nhiaHCPPrimaryFacilityType},
		{Page: 1, PageSize: 1, FacilityType: nhiaHCPPrimaryAndSecondaryFacilityType},
		{Page: 1, PageSize: 1, ListingStatus: nhiaHCPListingStatus},
		{Page: nhiaHCPExpectedCount + 1, PageSize: 1},
	}
	if !reflect.DeepEqual(s.calls, wantCalls) || !reflect.DeepEqual(s.ids, []string{nhiaHCPFirstAnchor, nhiaHCPLastAnchor}) {
		t.Fatalf("calls %v ids %v", s.calls, s.ids)
	}
}

func TestNHIAActiveAccreditedHealthcareProviderStartupRejectsBadPagesAndRecords(t *testing.T) {
	for _, tc := range []struct {
		name, key string
		mutate    func(*interfaces.NHIAActiveAccreditedHealthcareProviderListResult)
	}{
		{"wrong total", "default", func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) { p.Total = 6535 }},
		{"wrong total pages", "default", func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) { p.TotalPages = 1 }},
		{"nil first", "default", func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) { p.Records = nil }},
		{"empty first", "default", func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) {
			p.Records = []models.NHIAActiveAccreditedHealthcareProvider{}
		}},
		{"bad first anchor", "default", func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) { p.Records[0].ID = "wrong" }},
		{"provider code zero", "provider_code", func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) {
			p.Records = []models.NHIAActiveAccreditedHealthcareProvider{}
		}},
		{"provider code multiple", "provider_code", func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) {
			p.Records = append(p.Records, p.Records[0])
		}},
		{"provider code wrong total", "provider_code", func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) { p.Total = 2 }},
		{"provider code wrong id", "provider_code", func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) {
			p.Records[0].ID = nhiaHCPFirstAnchor
		}},
		{"provider code wrong code", "provider_code", func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) {
			p.Records[0].ProviderCode = nhiaHCPFirstAnchorCode
		}},
		{"facility total", "facility_type:" + nhiaHCPPrimaryFacilityType, func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) { p.Total = 3 }},
		{"facility wrong type", "facility_type:" + nhiaHCPPrimaryFacilityType, func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) {
			p.Records[0].FacilityType = nhiaHCPPrimaryAndSecondaryFacilityType
		}},
		{"listing status total", "listing_status", func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) { p.Total = 1 }},
		{"listing wrong status", "listing_status", func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) {
			p.Records[0].ListingStatus = "listed"
		}},
		{"beyond nonempty", "beyond", func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) {
			p.Records = []models.NHIAActiveAccreditedHealthcareProvider{{ID: "unexpected"}}
		}},
		{"beyond nil", "beyond", func(p *interfaces.NHIAActiveAccreditedHealthcareProviderListResult) { p.Records = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := validStartupNHIAHCP()
			page := s.pages[tc.key]
			tc.mutate(&page)
			s.pages[tc.key] = page
			if err := verifyNHIAActiveAccreditedHealthcareProviders(context.Background(), s); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("error = %v", err)
			}
		})
	}
	for name, mutate := range map[string]func(*models.NHIAActiveAccreditedHealthcareProvider){
		"id":       func(r *models.NHIAActiveAccreditedHealthcareProvider) { r.ID = "bad_id" },
		"name":     func(r *models.NHIAActiveAccreditedHealthcareProvider) { r.Name = " " },
		"country":  func(r *models.NHIAActiveAccreditedHealthcareProvider) { r.CountryCode = "GH" },
		"code":     func(r *models.NHIAActiveAccreditedHealthcareProvider) { r.ProviderCode = "bad" },
		"type":     func(r *models.NHIAActiveAccreditedHealthcareProvider) { r.FacilityType = "secondary" },
		"status":   func(r *models.NHIAActiveAccreditedHealthcareProvider) { r.ListingStatus = "listed" },
		"trimcode": func(r *models.NHIAActiveAccreditedHealthcareProvider) { r.ProviderCode = " FCT/0001/P " },
	} {
		t.Run(name, func(t *testing.T) {
			record := validStartupNHIAHCP().anchors[nhiaHCPFirstAnchor]
			mutate(&record)
			if err := validateStartupNHIAHCP(record); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestNHIAActiveAccreditedHealthcareProviderStartupErrors(t *testing.T) {
	if err := verifyNHIAActiveAccreditedHealthcareProviders(context.Background(), nil); err == nil {
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
		s := validStartupNHIAHCP()
		if err := verifyNHIAActiveAccreditedHealthcareProviders(ctx, s); !errors.Is(err, ctx.Err()) || len(s.calls) != 0 {
			t.Fatalf("context %v", err)
		}
		cancel()
	}
	for step := 1; step <= 8; step++ {
		for _, cause := range []error{context.Canceled, context.DeadlineExceeded, errors.New("secret decoder path")} {
			s := validStartupNHIAHCP()
			s.failAt = step
			s.err = fmt.Errorf("secret wrapper: %w", cause)
			err := verifyNHIAActiveAccreditedHealthcareProviders(context.Background(), s)
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

type nhiaHCPCountingJSON struct {
	interfaces.JSONFileRepository
	counts map[string]int
}

func (s *nhiaHCPCountingJSON) Decode(ctx context.Context, path string, dst any) error {
	s.counts[path]++
	return s.JSONFileRepository.Decode(ctx, path, dst)
}

func TestNHIAActiveAccreditedHealthcareProviderProductionBuilderVerifiesAndReusesCache(t *testing.T) {
	base, err := fileRepo.NewEmbeddedJSONRepository(datasets.Files(), 64<<20)
	if err != nil {
		t.Fatal(err)
	}
	counting := &nhiaHCPCountingJSON{JSONFileRepository: base, counts: map[string]int{}}
	service, handler, err := buildNHIAActiveAccreditedHealthcareProviderHandler(
		context.Background(),
		counting,
		func(repository interfaces.JSONFileRepository, recordsPath string) (interfaces.NHIAActiveAccreditedHealthcareProviderRepository, error) {
			return fileRepo.NewNHIAActiveAccreditedHealthcareProviderRepository(repository, recordsPath)
		},
		func(repository interfaces.NHIAActiveAccreditedHealthcareProviderRepository) (nhiaActiveAccreditedHealthcareProviderService, error) {
			return services.NewNHIAActiveAccreditedHealthcareProviderService(repository)
		},
		func(service nhiaActiveAccreditedHealthcareProviderService) (*handlers.NHIAActiveAccreditedHealthcareProviderHandler, error) {
			return handlers.NewNHIAActiveAccreditedHealthcareProviderHandler(service)
		},
	)
	if err != nil || service == nil || handler == nil {
		t.Fatalf("builder service=%T handler=%T err=%v", service, handler, err)
	}
	if _, err := service.GetNHIAActiveAccreditedHealthcareProvider(context.Background(), nhiaHCPFirstAnchor); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(counting.counts, map[string]int{healthcareNHIAActiveAccreditedHealthcareProvidersPath: 1}) {
		t.Fatalf("unexpected reads: %v", counting.counts)
	}
}

type nhiaHCPAppJSONStub struct {
	decodeCalls int
}

func (s *nhiaHCPAppJSONStub) Decode(context.Context, string, any) error {
	s.decodeCalls++
	return nil
}

type nhiaHCPAppRepositoryStub struct{}

func (s *nhiaHCPAppRepositoryStub) ListNHIAActiveAccreditedHealthcareProviders(context.Context, interfaces.NHIAActiveAccreditedHealthcareProviderQuery) (interfaces.NHIAActiveAccreditedHealthcareProviderListResult, error) {
	return interfaces.NHIAActiveAccreditedHealthcareProviderListResult{Records: make([]models.NHIAActiveAccreditedHealthcareProvider, 0), Page: 1, PageSize: 50, Total: 0, TotalPages: 0}, nil
}

func (s *nhiaHCPAppRepositoryStub) GetNHIAActiveAccreditedHealthcareProvider(context.Context, string) (models.NHIAActiveAccreditedHealthcareProvider, error) {
	return models.NHIAActiveAccreditedHealthcareProvider{}, interfaces.ErrNHIAActiveAccreditedHealthcareProviderNotFound
}

type nhiaHCPAppServiceStub struct{}

func (s *nhiaHCPAppServiceStub) ListNHIAActiveAccreditedHealthcareProviders(context.Context, services.NHIAActiveAccreditedHealthcareProviderQuery) (services.NHIAActiveAccreditedHealthcareProviderListResult, error) {
	return services.NHIAActiveAccreditedHealthcareProviderListResult{}, nil
}

func (s *nhiaHCPAppServiceStub) GetNHIAActiveAccreditedHealthcareProvider(context.Context, string) (models.NHIAActiveAccreditedHealthcareProvider, error) {
	return models.NHIAActiveAccreditedHealthcareProvider{}, nil
}
