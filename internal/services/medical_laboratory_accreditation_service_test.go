package services

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	fileRepo "github.com/AbdulQuayyum/softdata-api/internal/repository/file"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestMedicalLaboratoryAccreditationServiceRealDataset(t *testing.T) {
	data, err := os.ReadFile("../../datasets/healthcare/medical_laboratory_accreditations.json")
	if err != nil {
		t.Fatal(err)
	}
	var records []models.MedicalLaboratoryAccreditation
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatal(err)
	}
	jsonRepository, err := fileRepo.NewJSONRepository("../../datasets", 64<<20)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := fileRepo.NewMedicalLaboratoryAccreditationRepository(jsonRepository, "healthcare/medical_laboratory_accreditations.json", "geography/states.json")
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewMedicalLaboratoryAccreditationService(repository)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range records {
		got, err := service.GetMedicalLaboratoryAccreditation(context.Background(), record.ID)
		if err != nil || got.ID != record.ID {
			t.Fatalf("published ID %q is not retrievable: got %q, error %v", record.ID, got.ID, err)
		}
	}
}

type medicalLabAccreditationServiceRepositoryStub struct {
	listResult interfaces.MedicalLaboratoryAccreditationListResult
	listErr    error
	getResult  models.MedicalLaboratoryAccreditation
	getErr     error
	lastQuery  interfaces.MedicalLaboratoryAccreditationQuery
	lastID     string
	listCalls  int
	getCalls   int
}

func (s *medicalLabAccreditationServiceRepositoryStub) ListMedicalLaboratoryAccreditations(_ context.Context, query interfaces.MedicalLaboratoryAccreditationQuery) (interfaces.MedicalLaboratoryAccreditationListResult, error) {
	s.listCalls++
	s.lastQuery = query
	return s.listResult, s.listErr
}

func (s *medicalLabAccreditationServiceRepositoryStub) GetMedicalLaboratoryAccreditation(_ context.Context, id string) (models.MedicalLaboratoryAccreditation, error) {
	s.getCalls++
	s.lastID = id
	return s.getResult, s.getErr
}

func TestMedicalLaboratoryAccreditationServiceNormalizesAndClones(t *testing.T) {
	repository := &medicalLabAccreditationServiceRepositoryStub{listResult: interfaces.MedicalLaboratoryAccreditationListResult{
		Records: []models.MedicalLaboratoryAccreditation{{ID: "record-one", Name: "Record One"}},
		Page:    1, PageSize: 50, Total: 1, TotalPages: 1,
	}, getResult: models.MedicalLaboratoryAccreditation{ID: "record-one", Name: "Record One"}}
	service, err := NewMedicalLaboratoryAccreditationService(repository)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.ListMedicalLaboratoryAccreditations(context.Background(), MedicalLaboratoryAccreditationQuery{StateID: " lagos ", AccreditationStatus: " accredited ", Search: " record "})
	if err != nil || len(result.Records) != 1 {
		t.Fatalf("list = %#v, %v", result, err)
	}
	if repository.lastQuery.Page != 1 || repository.lastQuery.PageSize != 50 || repository.lastQuery.StateID != "lagos" || repository.lastQuery.AccreditationStatus != "accredited" || repository.lastQuery.Search != "record" {
		t.Fatalf("query was not normalized: %#v", repository.lastQuery)
	}
	result.Records[0].Name = "mutated"
	again, err := service.ListMedicalLaboratoryAccreditations(context.Background(), MedicalLaboratoryAccreditationQuery{})
	if err != nil || again.Records[0].Name == "mutated" {
		t.Fatalf("service returned shared slice state: %#v, %v", again, err)
	}
	got, err := service.GetMedicalLaboratoryAccreditation(context.Background(), " record-one ")
	if err != nil || got.ID != "record-one" || repository.lastID != "record-one" {
		t.Fatalf("detail = %#v, %v lastID=%q", got, err, repository.lastID)
	}
}

func TestMedicalLaboratoryAccreditationServiceValidation(t *testing.T) {
	repository := &medicalLabAccreditationServiceRepositoryStub{}
	service, err := NewMedicalLaboratoryAccreditationService(repository)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		query MedicalLaboratoryAccreditationQuery
		want  error
	}{
		{"page", MedicalLaboratoryAccreditationQuery{Page: -1}, ErrInvalidMedicalLaboratoryAccreditationPagination},
		{"page size", MedicalLaboratoryAccreditationQuery{PageSize: 101}, ErrInvalidMedicalLaboratoryAccreditationPagination},
		{"state", MedicalLaboratoryAccreditationQuery{StateID: "Lagos"}, ErrInvalidMedicalLaboratoryAccreditationStateID},
		{"status", MedicalLaboratoryAccreditationQuery{AccreditationStatus: "licensed"}, ErrInvalidMedicalLaboratoryAccreditationStatus},
		{"search too long", MedicalLaboratoryAccreditationQuery{Search: strings.Repeat("x", 101)}, ErrInvalidMedicalLaboratoryAccreditationSearch},
		{"search newline", MedicalLaboratoryAccreditationQuery{Search: "bad\nsearch"}, ErrInvalidMedicalLaboratoryAccreditationSearch},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := service.ListMedicalLaboratoryAccreditations(context.Background(), test.query); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	for _, id := range []string{"", "Upper-case", "bad/id", "bad--id", "-leading", "trailing-"} {
		if _, err := service.GetMedicalLaboratoryAccreditation(context.Background(), id); !errors.Is(err, ErrInvalidMedicalLaboratoryAccreditationID) {
			t.Fatalf("GetMedicalLaboratoryAccreditation(%q) error = %v", id, err)
		}
	}
	if repository.listCalls != 0 || repository.getCalls != 0 {
		t.Fatal("invalid requests reached repository")
	}
}

func TestMedicalLaboratoryAccreditationServiceIDLengthBoundaries(t *testing.T) {
	repository := &medicalLabAccreditationServiceRepositoryStub{getErr: interfaces.ErrMedicalLaboratoryAccreditationNotFound}
	service, err := NewMedicalLaboratoryAccreditationService(repository)
	if err != nil {
		t.Fatal(err)
	}
	for _, size := range []int{129, models.MedicalLaboratoryAccreditationIDMaxLength} {
		if _, err := service.GetMedicalLaboratoryAccreditation(context.Background(), strings.Repeat("a", size)); !errors.Is(err, ErrMedicalLaboratoryAccreditationNotFound) {
			t.Fatalf("valid unknown ID length %d: %v", size, err)
		}
	}
	calls := repository.getCalls
	for _, id := range []string{strings.Repeat("a", models.MedicalLaboratoryAccreditationIDMaxLength+1), "valid/invalid", "UPPER"} {
		if _, err := service.GetMedicalLaboratoryAccreditation(context.Background(), id); !errors.Is(err, ErrInvalidMedicalLaboratoryAccreditationID) {
			t.Fatalf("invalid ID accepted: %v", err)
		}
	}
	if repository.getCalls != calls {
		t.Fatal("invalid IDs reached repository")
	}
}

func TestMedicalLaboratoryAccreditationServiceErrorTranslationAndContext(t *testing.T) {
	for _, test := range []struct {
		name    string
		repoErr error
		want    error
	}{
		{"not found", interfaces.ErrMedicalLaboratoryAccreditationNotFound, ErrMedicalLaboratoryAccreditationNotFound},
		{"invalid dataset", interfaces.ErrInvalidDatasetFile, ErrInvalidMedicalLaboratoryAccreditationDataset},
		{"unexpected", errors.New("/private/repository internals"), nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := &medicalLabAccreditationServiceRepositoryStub{getErr: test.repoErr}
			service, _ := NewMedicalLaboratoryAccreditationService(repository)
			_, err := service.GetMedicalLaboratoryAccreditation(context.Background(), "valid-id")
			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if test.want == nil && (err == nil || strings.Contains(err.Error(), "/private")) {
				t.Fatalf("unexpected error was not sanitized: %v", err)
			}
		})
	}
	service, _ := NewMedicalLaboratoryAccreditationService(&medicalLabAccreditationServiceRepositoryStub{listErr: context.Canceled})
	if _, err := service.ListMedicalLaboratoryAccreditations(context.Background(), MedicalLaboratoryAccreditationQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled list error = %v", err)
	}
	service, _ = NewMedicalLaboratoryAccreditationService(&medicalLabAccreditationServiceRepositoryStub{getErr: context.DeadlineExceeded})
	if _, err := service.GetMedicalLaboratoryAccreditation(context.Background(), "valid-id"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline detail error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service, _ = NewMedicalLaboratoryAccreditationService(&medicalLabAccreditationServiceRepositoryStub{})
	if _, err := service.ListMedicalLaboratoryAccreditations(ctx, MedicalLaboratoryAccreditationQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("pre-canceled list error = %v", err)
	}
}
