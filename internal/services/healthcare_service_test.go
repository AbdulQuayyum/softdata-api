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

func TestHealthFacilityServiceRetrievesEveryPublishedID(t *testing.T) {
	data, err := os.ReadFile("../../datasets/healthcare/health_facilities.json")
	if err != nil {
		t.Fatal(err)
	}
	var facilities []models.HealthFacility
	if err := json.Unmarshal(data, &facilities); err != nil {
		t.Fatal(err)
	}
	jsonRepository, err := fileRepo.NewJSONRepository("../../datasets", 64<<20)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := fileRepo.NewHealthFacilityRepository(jsonRepository, "healthcare/health_facilities.json", "geography/states.json", "geography/lgas.json")
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewHealthFacilityService(repository)
	if err != nil {
		t.Fatal(err)
	}
	for _, facility := range facilities {
		got, err := service.GetHealthFacility(context.Background(), facility.ID)
		if err != nil || got.ID != facility.ID {
			t.Fatalf("published ID %q is not retrievable: got %q, error %v", facility.ID, got.ID, err)
		}
	}
}

type healthFacilityServiceRepositoryStub struct {
	listResult interfaces.HealthFacilityListResult
	listErr    error
	getResult  models.HealthFacility
	getErr     error
	listCalls  int
	getCalls   int
}

func (s *healthFacilityServiceRepositoryStub) ListHealthFacilities(_ context.Context, _ interfaces.HealthFacilityQuery) (interfaces.HealthFacilityListResult, error) {
	s.listCalls++
	return s.listResult, s.listErr
}

func (s *healthFacilityServiceRepositoryStub) GetHealthFacility(_ context.Context, _ string) (models.HealthFacility, error) {
	s.getCalls++
	return s.getResult, s.getErr
}

func TestHealthFacilityServiceNormalizesAndTranslates(t *testing.T) {
	latitude := 6.5
	repository := &healthFacilityServiceRepositoryStub{listResult: interfaces.HealthFacilityListResult{
		Facilities: []models.HealthFacility{{ID: "facility-one", Name: "Facility One", Latitude: &latitude}},
		Page:       1, PageSize: 50, Total: 1, TotalPages: 1,
	}, getResult: models.HealthFacility{ID: "facility-one", Name: "Facility One", Latitude: &latitude}}
	service, err := NewHealthFacilityService(repository)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.ListHealthFacilities(context.Background(), HealthFacilityQuery{Page: 0, PageSize: 0, StateID: " lagos ", Search: " facility "})
	if err != nil || result.Page != 1 || len(result.Facilities) != 1 {
		t.Fatalf("normalized list = %#v, %v", result, err)
	}
	result.Facilities[0].Latitude = nil
	again, err := service.ListHealthFacilities(context.Background(), HealthFacilityQuery{})
	if err != nil || again.Facilities[0].Latitude == nil {
		t.Fatalf("service returned shared pointer state: %#v, %v", again, err)
	}
	got, err := service.GetHealthFacility(context.Background(), " facility-one ")
	if err != nil || got.ID != "facility-one" {
		t.Fatalf("detail = %#v, %v", got, err)
	}
	if repository.listCalls != 2 || repository.getCalls != 1 {
		t.Fatalf("unexpected repository calls: list=%d get=%d", repository.listCalls, repository.getCalls)
	}
}

func TestHealthFacilityServiceValidation(t *testing.T) {
	repository := &healthFacilityServiceRepositoryStub{}
	service, err := NewHealthFacilityService(repository)
	if err != nil {
		t.Fatal(err)
	}
	listCases := []struct {
		name  string
		query HealthFacilityQuery
		want  error
	}{
		{"page", HealthFacilityQuery{Page: -1}, ErrInvalidHealthFacilityPagination},
		{"page size", HealthFacilityQuery{PageSize: 101}, ErrInvalidHealthFacilityPagination},
		{"state", HealthFacilityQuery{StateID: "Lagos"}, ErrInvalidHealthFacilityStateID},
		{"lga syntax", HealthFacilityQuery{LGAID: "lagos/ikeja"}, ErrInvalidHealthFacilityLGAID},
		{"type", HealthFacilityQuery{FacilityType: "hospital"}, ErrInvalidHealthFacilityType},
		{"level", HealthFacilityQuery{FacilityLevel: "quaternary"}, ErrInvalidHealthFacilityLevel},
		{"ownership", HealthFacilityQuery{OwnershipType: "public"}, ErrInvalidHealthFacilityOwnership},
		{"search", HealthFacilityQuery{Search: strings.Repeat("x", 101)}, ErrInvalidHealthFacilitySearch},
	}
	for _, test := range listCases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := service.ListHealthFacilities(context.Background(), test.query); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	for _, id := range []string{"", "Upper-case", "bad/id", "bad--id", "-leading", "trailing-"} {
		if _, err := service.GetHealthFacility(context.Background(), id); !errors.Is(err, ErrInvalidHealthFacilityID) {
			t.Fatalf("GetHealthFacility(%q) error = %v", id, err)
		}
	}
	if repository.listCalls != 0 || repository.getCalls != 0 {
		t.Fatal("invalid requests reached repository")
	}
}

func TestHealthFacilityServiceErrorTranslationAndContext(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		want    error
	}{
		{"not found", interfaces.ErrHealthFacilityNotFound, ErrHealthFacilityNotFound},
		{"unexpected", errors.New("/private/health facility internals"), nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &healthFacilityServiceRepositoryStub{getErr: test.repoErr}
			service, _ := NewHealthFacilityService(repository)
			_, err := service.GetHealthFacility(context.Background(), "valid-id")
			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if test.want == nil && (err == nil || strings.Contains(err.Error(), "/private/health")) {
				t.Fatalf("unexpected error was not sanitized: %v", err)
			}
		})
	}
	service, _ := NewHealthFacilityService(&healthFacilityServiceRepositoryStub{listErr: context.Canceled})
	if _, err := service.ListHealthFacilities(context.Background(), HealthFacilityQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled list error = %v", err)
	}
	service, _ = NewHealthFacilityService(&healthFacilityServiceRepositoryStub{getErr: context.DeadlineExceeded})
	if _, err := service.GetHealthFacility(context.Background(), "valid-id"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline detail error = %v", err)
	}
}

func TestHealthFacilityServiceIDLengthBoundaries(t *testing.T) {
	repository := &healthFacilityServiceRepositoryStub{getErr: interfaces.ErrHealthFacilityNotFound}
	service, err := NewHealthFacilityService(repository)
	if err != nil {
		t.Fatal(err)
	}
	for _, size := range []int{129, models.HealthFacilityIDMaxLength} {
		if _, err := service.GetHealthFacility(context.Background(), strings.Repeat("a", size)); !errors.Is(err, ErrHealthFacilityNotFound) {
			t.Fatalf("valid unknown ID length %d: %v", size, err)
		}
	}
	calls := repository.getCalls
	for _, id := range []string{strings.Repeat("a", models.HealthFacilityIDMaxLength+1), "valid/invalid", "UPPER"} {
		if _, err := service.GetHealthFacility(context.Background(), id); !errors.Is(err, ErrInvalidHealthFacilityID) {
			t.Fatalf("invalid ID accepted: %v", err)
		}
	}
	if repository.getCalls != calls {
		t.Fatal("invalid IDs reached repository")
	}
}
