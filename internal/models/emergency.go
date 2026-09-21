package models

// EmergencyServiceContactIDMaxLength bounds public ASCII slug IDs for emergency contact records.
const EmergencyServiceContactIDMaxLength = 255

// EmergencyServiceContact represents one source-verified institutional emergency contact observation.
// It is published as a dated snapshot, not a live operational-status guarantee.
type EmergencyServiceContact struct {
	ID           string `json:"id"`
	ServiceName  string `json:"service_name"`
	AgencyName   string `json:"agency_name"`
	ServiceType  string `json:"service_type"`
	ContactType  string `json:"contact_type"`
	ContactValue string `json:"contact_value"`
	CoverageType string `json:"coverage_type"`
	CountryCode  string `json:"country_code"`
	StateID      string `json:"state_id,omitempty"`
	Availability string `json:"availability,omitempty"`
	CallCost     string `json:"call_cost,omitempty"`
	Notes        string `json:"notes,omitempty"`
}

// NEMAZonalTerritorialOperationOfficeIDMaxLength bounds public ASCII slug IDs for NEMA office records.
const NEMAZonalTerritorialOperationOfficeIDMaxLength = 255

// NEMAZonalTerritorialOperationOffice represents one source-verified NEMA office location.
// It is published as a dated snapshot, not a live operational-status guarantee.
type NEMAZonalTerritorialOperationOffice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	OfficeType  string `json:"office_type"`
	StateID     string `json:"state_id"`
	CountryCode string `json:"country_code"`
}

// FRSCZonalCommandIDMaxLength bounds public ASCII slug IDs for FRSC zonal command records.
const FRSCZonalCommandIDMaxLength = 255

// FRSCZonalCommand represents one source-verified Federal Road Safety Corps zonal command.
// It is published as a dated snapshot, not a live operational-status guarantee.
type FRSCZonalCommand struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CommandType string `json:"command_type"`
	CommandCode string `json:"command_code"`
	StateID     string `json:"state_id"`
	CountryCode string `json:"country_code"`
}
