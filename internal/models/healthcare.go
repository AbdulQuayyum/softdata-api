package models

// HealthFacilityIDMaxLength bounds public ASCII slug IDs across all API layers.
const HealthFacilityIDMaxLength = 255

// HealthFacility represents one source-verified Nigerian health facility.
// It is published as a dated registry snapshot rather than a live status feed.
type HealthFacility struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	FacilityType     string   `json:"facility_type"`
	FacilityLevel    string   `json:"facility_level,omitempty"`
	OwnershipType    string   `json:"ownership_type,omitempty"`
	StateID          string   `json:"state_id"`
	LGAID            string   `json:"lga_id,omitempty"`
	CountryCode      string   `json:"country_code"`
	SourceFacilityID string   `json:"source_facility_id,omitempty"`
	Latitude         *float64 `json:"latitude,omitempty"`
	Longitude        *float64 `json:"longitude,omitempty"`
}

// MedicalLaboratoryAccreditationIDMaxLength bounds public ASCII slug IDs for medical laboratory accreditation records.
const MedicalLaboratoryAccreditationIDMaxLength = 255

// MedicalLaboratoryAccreditation represents one source-verified Nigerian facility accreditation record.
// It is published as a dated MLSCN accreditation snapshot, not a complete live premises licensing register.
type MedicalLaboratoryAccreditation struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	StateID             string `json:"state_id"`
	CountryCode         string `json:"country_code"`
	AccreditationStatus string `json:"accreditation_status"`
	AccreditationNumber string `json:"accreditation_number,omitempty"`
	ApprovalDate        string `json:"approval_date,omitempty"`
	ExpiryDate          string `json:"expiry_date,omitempty"`
	Address             string `json:"address,omitempty"`
}

// NHIAAccreditedHealthMaintenanceOrganisationIDMaxLength bounds public ASCII slug IDs for NHIA HMO records.
const NHIAAccreditedHealthMaintenanceOrganisationIDMaxLength = 255

// NHIAAccreditedHealthMaintenanceOrganisation represents one NHIA-listed accredited HMO observation.
// It is published as a dated organisation-level snapshot, not a complete live health-insurance register.
type NHIAAccreditedHealthMaintenanceOrganisation struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	CountryCode         string `json:"country_code"`
	OrganisationType    string `json:"organisation_type"`
	AccreditationStatus string `json:"accreditation_status"`
	HMOID               string `json:"hmo_id"`
}
