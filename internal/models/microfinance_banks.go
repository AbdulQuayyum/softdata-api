package models

// MicrofinanceBank represents one institution in the reconciled Nigerian MFB roster.
type MicrofinanceBank struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CountryCode string `json:"country_code"`
}
