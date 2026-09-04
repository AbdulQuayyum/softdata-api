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

// LocalGovernmentUnit is the public record model for a Nigerian LGA or FCT area council.
type LocalGovernmentUnit struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	StateID            string `json:"state_id"`
	CountryCode        string `json:"country_code"`
	AdministrativeType string `json:"administrative_type"`
}

// Language is the public record model for a current CLDR base language identifier.
type Language struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CountryLanguage is the public record model for one country/area and language relationship.
type CountryLanguage struct {
	CountryAreaID string `json:"country_area_id"`
	LanguageID    string `json:"language_id"`
	Status        string `json:"status"`
}

// CountryOrArea is the public record model for a UN M49 country or area.
type CountryOrArea struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Alpha2Code             string   `json:"alpha_2_code"`
	Alpha3Code             string   `json:"alpha_3_code"`
	NumericCode            string   `json:"numeric_code"`
	CallingCodes           []string `json:"calling_codes,omitempty"`
	FlagEmoji              string   `json:"flag_emoji"`
	FlagSVGURL             string   `json:"flag_svg_url"`
	RegionCode             string   `json:"region_code,omitempty"`
	RegionName             string   `json:"region_name,omitempty"`
	SubregionCode          string   `json:"subregion_code,omitempty"`
	SubregionName          string   `json:"subregion_name,omitempty"`
	IntermediateRegionCode string   `json:"intermediate_region_code,omitempty"`
	IntermediateRegionName string   `json:"intermediate_region_name,omitempty"`
}

// CountryProfile is the derived public profile view for a country or area.
type CountryProfile struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Alpha2Code             string   `json:"alpha_2_code"`
	Alpha3Code             string   `json:"alpha_3_code"`
	NumericCode            string   `json:"numeric_code"`
	CallingCodes           []string `json:"calling_codes,omitempty"`
	FlagEmoji              string   `json:"flag_emoji"`
	FlagSVGURL             string   `json:"flag_svg_url"`
	RegionCode             string   `json:"region_code,omitempty"`
	RegionName             string   `json:"region_name,omitempty"`
	SubregionCode          string   `json:"subregion_code,omitempty"`
	SubregionName          string   `json:"subregion_name,omitempty"`
	IntermediateRegionCode string   `json:"intermediate_region_code,omitempty"`
	IntermediateRegionName string   `json:"intermediate_region_name,omitempty"`
	CurrencyIDs            []string `json:"currency_ids"`
	TimeZoneIDs            []string `json:"time_zone_ids"`
	LanguageIDs            []string `json:"language_ids"`
}

// TimeZone is the public record model for a canonical IANA time zone.
type TimeZone struct {
	ID             string   `json:"id"`
	CountryAreaIDs []string `json:"country_area_ids"`
}
