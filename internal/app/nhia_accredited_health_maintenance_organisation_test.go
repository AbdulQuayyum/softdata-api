package app

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
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
