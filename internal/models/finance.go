package models

// PaymentServiceProvider represents one current CBN payment-service-provider membership.
type PaymentServiceProvider struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	InstitutionType string `json:"institution_type"`
	CountryCode     string `json:"country_code"`
}

// InternationalMoneyTransferOperator represents one current CBN-listed IMTO entry.
type InternationalMoneyTransferOperator struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Currency represents one current monetary ISO 4217 currency.
type Currency struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	AlphabeticCode string   `json:"alphabetic_code"`
	NumericCode    string   `json:"numeric_code"`
	MinorUnit      int      `json:"minor_unit"`
	CountryAreaIDs []string `json:"country_area_ids"`
}
