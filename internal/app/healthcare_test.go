package app

import (
	"context"
	"errors"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type startupHealthFacilityStub struct {
	first    interfaces.HealthFacilityListResult
	filtered interfaces.HealthFacilityListResult
	beyond   interfaces.HealthFacilityListResult
	listErr  error
	getErr   error
	get      map[string]models.HealthFacility
}

func (s *startupHealthFacilityStub) ListHealthFacilities(_ context.Context, query interfaces.HealthFacilityQuery) (interfaces.HealthFacilityListResult, error) {
	if s.listErr != nil {
		return interfaces.HealthFacilityListResult{}, s.listErr
	}
	if query.Page > healthFacilityExpectedCount {
		return s.beyond, nil
	}
	if query.StateID != "" || query.LGAID != "" {
		return s.filtered, nil
	}
	return s.first, nil
}

func (s *startupHealthFacilityStub) GetHealthFacility(_ context.Context, id string) (models.HealthFacility, error) {
	if s.getErr != nil {
		return models.HealthFacility{}, s.getErr
	}
	if facility, ok := s.get[id]; ok {
		return facility, nil
	}
	return models.HealthFacility{}, interfaces.ErrHealthFacilityNotFound
}

func validStartupHealthFacility(id string) models.HealthFacility {
	latitude, longitude := 5.10332, 7.37854
	return models.HealthFacility{
		ID: id, Name: "222 Cliford Medical Center", FacilityType: "clinic", FacilityLevel: "primary",
		OwnershipType: "other-public", StateID: "abia", LGAID: "abia-aba-north", CountryCode: "NG",
		SourceFacilityID: "cea7a6ea-db7e-4845-8057-5caf45dc26c1", Latitude: &latitude, Longitude: &longitude,
	}
}

func validStartupHealthFacilityService() *startupHealthFacilityStub {
	first := validStartupHealthFacility(healthFacilityFirstAnchor)
	last := validStartupHealthFacility(healthFacilityLastAnchor)
	return &startupHealthFacilityStub{
		first:    interfaces.HealthFacilityListResult{Facilities: []models.HealthFacility{first}, Page: 1, PageSize: 1, Total: healthFacilityExpectedCount, TotalPages: healthFacilityExpectedCount},
		filtered: interfaces.HealthFacilityListResult{Facilities: []models.HealthFacility{first}, Page: 1, PageSize: 1, Total: 1, TotalPages: 1},
		beyond:   interfaces.HealthFacilityListResult{Facilities: []models.HealthFacility{}, Page: healthFacilityExpectedCount + 1, PageSize: 1, Total: healthFacilityExpectedCount, TotalPages: healthFacilityExpectedCount},
		get:      map[string]models.HealthFacility{first.ID: first, last.ID: last},
	}
}

func TestVerifyHealthFacilityDataset(t *testing.T) {
	if err := verifyHealthFacilityDataset(context.Background(), validStartupHealthFacilityService()); err != nil {
		t.Fatalf("verifyHealthFacilityDataset() error = %v", err)
	}
}

func TestVerifyHealthFacilityDatasetRejectsInvalidStartupResults(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*startupHealthFacilityStub)
	}{
		{name: "wrong total", mutate: func(s *startupHealthFacilityStub) { s.first.Total-- }},
		{name: "wrong total pages", mutate: func(s *startupHealthFacilityStub) { s.first.TotalPages-- }},
		{name: "empty first page", mutate: func(s *startupHealthFacilityStub) { s.first.Facilities = []models.HealthFacility{} }},
		{name: "nil first page", mutate: func(s *startupHealthFacilityStub) { s.first.Facilities = nil }},
		{name: "invalid first facility", mutate: func(s *startupHealthFacilityStub) { s.first.Facilities[0].CountryCode = "US" }},
		{name: "missing anchor", mutate: func(s *startupHealthFacilityStub) { delete(s.get, healthFacilityLastAnchor) }},
		{name: "non-empty beyond final", mutate: func(s *startupHealthFacilityStub) {
			s.beyond.Facilities = []models.HealthFacility{validStartupHealthFacility("unexpected")}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := validStartupHealthFacilityService()
			tc.mutate(service)
			if err := verifyHealthFacilityDataset(context.Background(), service); err == nil {
				t.Fatal("verifyHealthFacilityDataset() error = nil, want failure")
			}
		})
	}
}

func TestVerifyHealthFacilityDatasetPreservesContextAndWrapsFailures(t *testing.T) {
	serviceErr := errors.New("backend failure")
	tests := []struct {
		name string
		ctx  context.Context
		stub *startupHealthFacilityStub
		want error
	}{
		{name: "cancelled", ctx: cancelledContext(), stub: validStartupHealthFacilityService(), want: context.Canceled},
		{name: "deadline", ctx: deadlineContext(), stub: validStartupHealthFacilityService(), want: context.DeadlineExceeded},
		{name: "service failure", ctx: context.Background(), stub: func() *startupHealthFacilityStub {
			s := validStartupHealthFacilityService()
			s.listErr = serviceErr
			return s
		}(), want: serviceErr},
		{name: "detail failure", ctx: context.Background(), stub: func() *startupHealthFacilityStub {
			s := validStartupHealthFacilityService()
			s.getErr = serviceErr
			return s
		}(), want: serviceErr},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := verifyHealthFacilityDataset(tc.ctx, tc.stub)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want errors.Is(..., %v)", err, tc.want)
			}
		})
	}
}

func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func deadlineContext() context.Context {
	return deadlineExceededContext{Context: context.Background()}
}

type deadlineExceededContext struct {
	context.Context
}

func (deadlineExceededContext) Err() error {
	return context.DeadlineExceeded
}
