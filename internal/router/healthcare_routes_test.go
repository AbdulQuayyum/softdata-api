package router

import (
	"context"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type routerHealthFacilityStub struct {
	rec *routerRecorder
}

func (s *routerHealthFacilityStub) ListHealthFacilities(_ context.Context, query interfaces.HealthFacilityQuery) (interfaces.HealthFacilityListResult, error) {
	if s.rec != nil {
		s.rec.add("healthcare.list")
	}
	return interfaces.HealthFacilityListResult{
		Facilities: []models.HealthFacility{{
			ID: "sample-health-facility", Name: "Sample Health Facility", FacilityType: "clinic",
			FacilityLevel: "primary", OwnershipType: "private", StateID: "lagos", LGAID: "lagos-ikeja",
			CountryCode: "NG",
		}},
		Page: query.Page, PageSize: query.PageSize, Total: 50649,
		TotalPages: (50649 + query.PageSize - 1) / query.PageSize,
	}, nil
}

func (s *routerHealthFacilityStub) GetHealthFacility(_ context.Context, id string) (models.HealthFacility, error) {
	if s.rec != nil {
		s.rec.add("healthcare.detail:" + id)
	}
	return models.HealthFacility{
		ID: id, Name: "Sample Health Facility", FacilityType: "clinic", FacilityLevel: "primary",
		OwnershipType: "private", StateID: "lagos", LGAID: "lagos-ikeja", CountryCode: "NG",
	}, nil
}
