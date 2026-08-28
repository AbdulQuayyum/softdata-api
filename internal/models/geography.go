package models

// State is the public record model for a Nigerian state or the FCT.
type State struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	OfficialName       string `json:"official_name"`
	AdministrativeType string `json:"administrative_type"`
	Capital            string `json:"capital"`
	GeopoliticalZoneID string `json:"geopolitical_zone_id"`
	CountryCode        string `json:"country_code"`
}

// GeopoliticalZone is the public record model for a Nigerian geopolitical zone.
type GeopoliticalZone struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CountryCode string `json:"country_code"`
}
