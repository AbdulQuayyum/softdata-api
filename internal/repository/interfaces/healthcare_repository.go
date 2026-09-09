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

// MedicalLaboratoryAccreditationQuery captures supported medical-laboratory accreditation filters.
type MedicalLaboratoryAccreditationQuery struct {
	Page                int
	PageSize            int
	StateID             string
	AccreditationStatus string
	Search              string
}

// MedicalLaboratoryAccreditationListResult contains a page of accreditation records and metadata.
type MedicalLaboratoryAccreditationListResult struct {
	Records    []models.MedicalLaboratoryAccreditation
	Page       int
	PageSize   int
	Total      int
	TotalPages int
}

// MedicalLaboratoryAccreditationRepository defines paginated accreditation snapshot access.
type MedicalLaboratoryAccreditationRepository interface {
	ListMedicalLaboratoryAccreditations(context.Context, MedicalLaboratoryAccreditationQuery) (MedicalLaboratoryAccreditationListResult, error)
	GetMedicalLaboratoryAccreditation(context.Context, string) (models.MedicalLaboratoryAccreditation, error)
}
