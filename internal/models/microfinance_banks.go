package models

// MicrofinanceBank represents one institution in the reconciled Nigerian MFB roster.
type MicrofinanceBank struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	WebsiteURL  string `json:"website_url,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
	CountryCode string `json:"country_code"`
}
