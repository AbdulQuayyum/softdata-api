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

// NHIAAccreditedHealthMaintenanceOrganisationQuery captures supported NHIA accredited HMO filters.
type NHIAAccreditedHealthMaintenanceOrganisationQuery struct {
	Page                int
	PageSize            int
	AccreditationStatus string
	HMOID               string
	Search              string
}

// NHIAAccreditedHealthMaintenanceOrganisationListResult contains a page of NHIA accredited HMO records and metadata.
type NHIAAccreditedHealthMaintenanceOrganisationListResult struct {
	Records    []models.NHIAAccreditedHealthMaintenanceOrganisation
	Page       int
	PageSize   int
	Total      int
	TotalPages int
}

// NHIAAccreditedHealthMaintenanceOrganisationRepository defines paginated NHIA accredited HMO snapshot access.
type NHIAAccreditedHealthMaintenanceOrganisationRepository interface {
	ListNHIAAccreditedHealthMaintenanceOrganisations(context.Context, NHIAAccreditedHealthMaintenanceOrganisationQuery) (NHIAAccreditedHealthMaintenanceOrganisationListResult, error)
	GetNHIAAccreditedHealthMaintenanceOrganisation(context.Context, string) (models.NHIAAccreditedHealthMaintenanceOrganisation, error)
}

// NHIAStateSocialHealthInsuranceAgencyQuery captures supported NHIA-listed SSHIA filters.
type NHIAStateSocialHealthInsuranceAgencyQuery struct {
	Page     int
	PageSize int
	StateID  string
	Search   string
}

// NHIAStateSocialHealthInsuranceAgencyListResult contains a page of NHIA-listed SSHIA records and metadata.
type NHIAStateSocialHealthInsuranceAgencyListResult struct {
	Records    []models.NHIAStateSocialHealthInsuranceAgency
	Page       int
	PageSize   int
	Total      int
	TotalPages int
}

// NHIAStateSocialHealthInsuranceAgencyRepository defines paginated NHIA-listed SSHIA snapshot access.
type NHIAStateSocialHealthInsuranceAgencyRepository interface {
	ListNHIAStateSocialHealthInsuranceAgencies(context.Context, NHIAStateSocialHealthInsuranceAgencyQuery) (NHIAStateSocialHealthInsuranceAgencyListResult, error)
	GetNHIAStateSocialHealthInsuranceAgency(context.Context, string) (models.NHIAStateSocialHealthInsuranceAgency, error)
}

// NHIAActiveAccreditedHealthcareProviderQuery captures supported NHIA active-accredited HCP filters.
type NHIAActiveAccreditedHealthcareProviderQuery struct {
	Page          int
	PageSize      int
	ProviderCode  string
	FacilityType  string
	ListingStatus string
	Search        string
}

// NHIAActiveAccreditedHealthcareProviderListResult contains a page of NHIA active-accredited HCP records and metadata.
type NHIAActiveAccreditedHealthcareProviderListResult struct {
	Records    []models.NHIAActiveAccreditedHealthcareProvider
	Page       int
	PageSize   int
	Total      int
	TotalPages int
}

// NHIAActiveAccreditedHealthcareProviderRepository defines paginated NHIA active-accredited HCP snapshot access.
type NHIAActiveAccreditedHealthcareProviderRepository interface {
	ListNHIAActiveAccreditedHealthcareProviders(context.Context, NHIAActiveAccreditedHealthcareProviderQuery) (NHIAActiveAccreditedHealthcareProviderListResult, error)
	GetNHIAActiveAccreditedHealthcareProvider(context.Context, string) (models.NHIAActiveAccreditedHealthcareProvider, error)
}
