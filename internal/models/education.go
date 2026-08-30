package models

// University represents one current NUC-listed Nigerian university.
type University struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	OwnershipType string `json:"ownership_type"`
	StateID       string `json:"state_id"`
	CountryCode   string `json:"country_code"`
}
