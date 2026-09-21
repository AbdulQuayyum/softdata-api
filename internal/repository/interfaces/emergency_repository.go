package interfaces

import (
	"context"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

// EmergencyServiceContactQuery captures supported emergency-contact filters.
type EmergencyServiceContactQuery struct {
	Page         int
	PageSize     int
	ServiceType  string
	ContactType  string
	CoverageType string
	ContactValue string
	Search       string
}

// EmergencyServiceContactListResult contains a page of emergency-contact records and metadata.
type EmergencyServiceContactListResult struct {
	Records    []models.EmergencyServiceContact
	Page       int
	PageSize   int
	Total      int
	TotalPages int
}

// EmergencyServiceContactRepository defines paginated emergency-contact snapshot access.
type EmergencyServiceContactRepository interface {
	ListEmergencyServiceContacts(context.Context, EmergencyServiceContactQuery) (EmergencyServiceContactListResult, error)
	GetEmergencyServiceContact(context.Context, string) (models.EmergencyServiceContact, error)
}

// NEMAZonalTerritorialOperationOfficeQuery captures supported NEMA office filters.
type NEMAZonalTerritorialOperationOfficeQuery struct {
	Page       int
	PageSize   int
	StateID    string
	OfficeType string
	Search     string
}

// NEMAZonalTerritorialOperationOfficeListResult contains a page of NEMA office records and metadata.
type NEMAZonalTerritorialOperationOfficeListResult struct {
	Records    []models.NEMAZonalTerritorialOperationOffice
	Page       int
	PageSize   int
	Total      int
	TotalPages int
}

// NEMAZonalTerritorialOperationOfficeRepository defines paginated NEMA office snapshot access.
type NEMAZonalTerritorialOperationOfficeRepository interface {
	ListNEMAZonalTerritorialOperationOffices(context.Context, NEMAZonalTerritorialOperationOfficeQuery) (NEMAZonalTerritorialOperationOfficeListResult, error)
	GetNEMAZonalTerritorialOperationOffice(context.Context, string) (models.NEMAZonalTerritorialOperationOffice, error)
}
