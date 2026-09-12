package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestNHIAAccreditedHMOServiceListValidationAndPropagation(t *testing.T) {
	stub := &nhiaAccreditedHMOServiceRepositoryStub{
		listResult: interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{
			Records: []models.NHIAAccreditedHealthMaintenanceOrganisation{{ID: "a-and-m-healthcare-trust-limited-102", Name: "A&M HEALTHCARE TRUST LIMITED", CountryCode: "NG", OrganisationType: "health_maintenance_organisation", AccreditationStatus: "accredited", HMOID: "102"}},
			Total:   1,
		},
	}
	service := mustNewNHIAAccreditedHMOService(t, stub)

	result, err := service.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), NHIAAccreditedHealthMaintenanceOrganisationQuery{
		AccreditationStatus: " accredited ",
		HMOID:               " HMO-001 ",
		Search:              " trust ",
	})
	if err != nil {
		t.Fatalf("ListNHIAAccreditedHealthMaintenanceOrganisations() error = %v", err)
	}
	if stub.listQuery.Page != 1 || stub.listQuery.PageSize != 50 || stub.listQuery.AccreditationStatus != "accredited" || stub.listQuery.HMOID != "HMO-001" || stub.listQuery.Search != "trust" {
		t.Fatalf("query was not normalized and propagated: %#v", stub.listQuery)
	}
	result.Records[0].Name = "mutated"
	if stub.listResult.Records[0].Name == "mutated" {
		t.Fatal("service returned repository-owned slice")
	}
}

func TestNHIAAccreditedHMOServiceRejectsInvalidListInputs(t *testing.T) {
	service := mustNewNHIAAccreditedHMOService(t, &nhiaAccreditedHMOServiceRepositoryStub{})
	for _, test := range []struct {
		name  string
		query NHIAAccreditedHealthMaintenanceOrganisationQuery
		want  error
	}{
		{"bad page", NHIAAccreditedHealthMaintenanceOrganisationQuery{Page: -1}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationPagination},
		{"bad page size low", NHIAAccreditedHealthMaintenanceOrganisationQuery{PageSize: -1}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationPagination},
		{"bad page size high", NHIAAccreditedHealthMaintenanceOrganisationQuery{PageSize: 101}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationPagination},
		{"blank status", NHIAAccreditedHealthMaintenanceOrganisationQuery{AccreditationStatus: "  "}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationStatus},
		{"bad status", NHIAAccreditedHealthMaintenanceOrganisationQuery{AccreditationStatus: "expired"}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationStatus},
		{"blank hmo id", NHIAAccreditedHealthMaintenanceOrganisationQuery{HMOID: "  "}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationHMOID},
		{"bad hmo id", NHIAAccreditedHealthMaintenanceOrganisationQuery{HMOID: "12\n0"}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationHMOID},
		{"blank search", NHIAAccreditedHealthMaintenanceOrganisationQuery{Search: "  "}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationSearch},
		{"long search", NHIAAccreditedHealthMaintenanceOrganisationQuery{Search: strings.Repeat("x", 101)}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationSearch},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := service.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), test.query); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestNHIAAccreditedHMOServiceDetailAndErrors(t *testing.T) {
	record := models.NHIAAccreditedHealthMaintenanceOrganisation{ID: "a-and-m-healthcare-trust-limited-102", Name: "A&M HEALTHCARE TRUST LIMITED", CountryCode: "NG", OrganisationType: "health_maintenance_organisation", AccreditationStatus: "accredited", HMOID: "102"}
	stub := &nhiaAccreditedHMOServiceRepositoryStub{getRecord: record}
	service := mustNewNHIAAccreditedHMOService(t, stub)
	got, err := service.GetNHIAAccreditedHealthMaintenanceOrganisation(context.Background(), " "+record.ID+" ")
	if err != nil || got.ID != record.ID || stub.getID != record.ID {
		t.Fatalf("detail = %#v, id=%q, err=%v", got, stub.getID, err)
	}
	if _, err := service.GetNHIAAccreditedHealthMaintenanceOrganisation(context.Background(), "Bad ID"); !errors.Is(err, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationID) {
		t.Fatalf("invalid id error = %v", err)
	}

	stub.getErr = interfaces.ErrNHIAAccreditedHealthMaintenanceOrganisationNotFound
	if _, err := service.GetNHIAAccreditedHealthMaintenanceOrganisation(context.Background(), record.ID); !errors.Is(err, ErrNHIAAccreditedHealthMaintenanceOrganisationNotFound) {
		t.Fatalf("not found error = %v", err)
	}
	stub.getErr = errors.New("secret /tmp/path")
	if _, err := service.GetNHIAAccreditedHealthMaintenanceOrganisation(context.Background(), record.ID); err == nil || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "/tmp/path") {
		t.Fatalf("unexpected error was not sanitized: %v", err)
	}
}

func TestNHIAAccreditedHMOServicePreservesContextErrors(t *testing.T) {
	service := mustNewNHIAAccreditedHMOService(t, &nhiaAccreditedHMOServiceRepositoryStub{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.ListNHIAAccreditedHealthMaintenanceOrganisations(ctx, NHIAAccreditedHealthMaintenanceOrganisationQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled list error = %v", err)
	}

	stub := &nhiaAccreditedHMOServiceRepositoryStub{listErr: context.DeadlineExceeded}
	service = mustNewNHIAAccreditedHMOService(t, stub)
	if _, err := service.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), NHIAAccreditedHealthMaintenanceOrganisationQuery{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline list error = %v", err)
	}
}

func mustNewNHIAAccreditedHMOService(t testing.TB, repo interfaces.NHIAAccreditedHealthMaintenanceOrganisationRepository) *NHIAAccreditedHealthMaintenanceOrganisationService {
	t.Helper()
	service, err := NewNHIAAccreditedHealthMaintenanceOrganisationService(repo)
	if err != nil {
		t.Fatal(err)
	}
	return service
}

type nhiaAccreditedHMOServiceRepositoryStub struct {
	listQuery  interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery
	listResult interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult
	listErr    error
	getID      string
	getRecord  models.NHIAAccreditedHealthMaintenanceOrganisation
	getErr     error
}

func (s *nhiaAccreditedHMOServiceRepositoryStub) ListNHIAAccreditedHealthMaintenanceOrganisations(_ context.Context, query interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery) (interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult, error) {
	s.listQuery = query
	if s.listErr != nil {
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{}, s.listErr
	}
	if s.listResult.Records == nil {
		return interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult{Records: make([]models.NHIAAccreditedHealthMaintenanceOrganisation, 0), Page: query.Page, PageSize: query.PageSize}, nil
	}
	return s.listResult, nil
}

func (s *nhiaAccreditedHMOServiceRepositoryStub) GetNHIAAccreditedHealthMaintenanceOrganisation(_ context.Context, id string) (models.NHIAAccreditedHealthMaintenanceOrganisation, error) {
	s.getID = id
	if s.getErr != nil {
		return models.NHIAAccreditedHealthMaintenanceOrganisation{}, s.getErr
	}
	if s.getRecord.ID == "" {
		return models.NHIAAccreditedHealthMaintenanceOrganisation{}, fmt.Errorf("%w", interfaces.ErrNHIAAccreditedHealthMaintenanceOrganisationNotFound)
	}
	return s.getRecord, nil
}
