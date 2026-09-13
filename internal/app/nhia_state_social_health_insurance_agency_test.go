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

func TestNHIAStateSocialHealthInsuranceAgencyAppDependencyTypeKeepsNoHTTPContract(t *testing.T) {
	depsType := reflect.TypeOf(appDependencies{})
	if _, ok := depsType.FieldByName("nhiaSSHIAService"); !ok {
		t.Fatal("NHIA SSHIA service dependency missing")
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
