package models

// PaymentServiceProvider represents one current CBN payment-service-provider membership.
type PaymentServiceProvider struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	InstitutionType string `json:"institution_type"`
	CountryCode     string `json:"country_code"`
}
