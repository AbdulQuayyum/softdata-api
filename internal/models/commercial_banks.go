package models

// CommercialBank represents one CBN-listed Nigerian commercial bank.
type CommercialBank struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	CBNCode            string `json:"cbn_code,omitempty"`
	NIPCode            string `json:"nip_code,omitempty"`
	OfficialWebsiteURL string `json:"official_website_url"`
	LogoURL            string `json:"logo_url"`
	CountryCode        string `json:"country_code"`
}
