package models

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
