package interfaces

import (
	"context"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

// HealthFacilityQuery captures the supported health-facility filters.
type HealthFacilityQuery struct {
	Page          int
	PageSize      int
	StateID       string
	LGAID         string
	FacilityType  string
	FacilityLevel string
	OwnershipType string
	Search        string
}

// HealthFacilityListResult contains a page of health facilities and metadata.
type HealthFacilityListResult struct {
	Facilities []models.HealthFacility
	Page       int
	PageSize   int
	Total      int
	TotalPages int
}

// HealthFacilityRepository defines paginated health-facility access.
type HealthFacilityRepository interface {
	ListHealthFacilities(context.Context, HealthFacilityQuery) (HealthFacilityListResult, error)
	GetHealthFacility(context.Context, string) (models.HealthFacility, error)
}
