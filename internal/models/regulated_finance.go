package models

// NonInterestInstitution represents one CBN non-interest financial institution.
type NonInterestInstitution struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CBNCode     string `json:"cbn_code,omitempty"`
	NIPCode     string `json:"nip_code,omitempty"`
	WebsiteURL  string `json:"website_url,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
	CountryCode string `json:"country_code"`
}

// MerchantBank represents one CBN merchant bank.
type MerchantBank NonInterestInstitution

// PaymentServiceBank represents one CBN payment service bank.
type PaymentServiceBank NonInterestInstitution

// FinancialHoldingCompany represents one CBN financial holding company.
type FinancialHoldingCompany struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	WebsiteURL  string `json:"website_url,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
	CountryCode string `json:"country_code"`
}

// DevelopmentFinanceInstitution represents one CBN development finance institution.
type DevelopmentFinanceInstitution NonInterestInstitution

// PrimaryMortgageInstitution represents one CBN primary mortgage institution.
type PrimaryMortgageInstitution NonInterestInstitution
