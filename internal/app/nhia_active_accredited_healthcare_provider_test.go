package app

import (
	"context"
	"errors"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
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
	deps := appDependencies{nhiaHCPService: &nhiaHCPAppServiceStub{}}
	if deps.nhiaHCPService == nil {
		t.Fatal("nhia HCP service was not retained")
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
