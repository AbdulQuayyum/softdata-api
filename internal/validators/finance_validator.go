package validators

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

var financePaymentServiceProviderIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)
var financeInternationalMoneyTransferOperatorIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var financeCurrencyIDPattern = regexp.MustCompile(`^[a-z]{3}$`)
var financeCurrencyCountryAreaIDPattern = regexp.MustCompile(`^[a-z]{2}$`)
var financeCommercialBankIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)

var approvedCommercialBankIDs = map[string]struct{}{
	"access-bank": {}, "alpha-morgan-bank": {}, "citibank-nigeria": {}, "ecobank-nigeria": {}, "fidelity-bank": {},
	"first-bank-of-nigeria": {}, "first-city-monument-bank": {}, "globus-bank": {}, "guaranty-trust-bank": {}, "keystone-bank": {},
	"nova-bank": {}, "optimus-bank": {}, "parallex-bank": {}, "polaris-bank": {}, "premium-trust-bank": {}, "providus-bank": {},
	"signature-bank": {}, "stanbic-ibtc-bank": {}, "standard-chartered-bank": {}, "sterling-bank": {}, "suntrust-bank": {}, "tatum-bank": {},
	"titan-trust-bank": {}, "union-bank": {}, "united-bank-for-africa": {}, "unity-bank": {}, "wema-bank": {}, "zenith-bank": {},
}

var financePaymentServiceProviderTypes = map[string]struct{}{
	"mobile_money_operator":               {},
	"switching_and_processing_company":    {},
	"payment_solution_service_provider":   {},
	"payment_terminal_service_provider":   {},
	"super_agent":                         {},
	"payment_service_holding_company":     {},
	"payment_terminal_service_aggregator": {},
}

var validCurrencyCountryAreaIDs = map[string]struct{}{
	"af": {},
	"al": {},
	"dz": {},
	"as": {},
	"ad": {},
	"ao": {},
	"ai": {},
	"aq": {},
	"ag": {},
	"ar": {},
	"am": {},
	"aw": {},
	"au": {},
	"at": {},
	"az": {},
	"bs": {},
	"bh": {},
	"bd": {},
	"bb": {},
	"by": {},
	"be": {},
	"bz": {},
	"bj": {},
	"bm": {},
	"bt": {},
	"bo": {},
	"bq": {},
	"ba": {},
	"bw": {},
	"bv": {},
	"br": {},
	"io": {},
	"vg": {},
	"bn": {},
	"bg": {},
	"bf": {},
	"bi": {},
	"cv": {},
	"kh": {},
	"cm": {},
	"ca": {},
	"ky": {},
	"cf": {},
	"td": {},
	"cl": {},
	"cn": {},
	"hk": {},
	"mo": {},
	"cx": {},
	"cc": {},
	"co": {},
	"km": {},
	"cg": {},
	"ck": {},
	"cr": {},
	"hr": {},
	"cu": {},
	"cw": {},
	"cy": {},
	"cz": {},
	"ci": {},
	"kp": {},
	"cd": {},
	"dk": {},
	"dj": {},
	"dm": {},
	"do": {},
	"ec": {},
	"eg": {},
	"sv": {},
	"gq": {},
	"er": {},
	"ee": {},
	"sz": {},
	"et": {},
	"fk": {},
	"fo": {},
	"fj": {},
	"fi": {},
	"fr": {},
	"gf": {},
	"pf": {},
	"tf": {},
	"ga": {},
	"gm": {},
	"ge": {},
	"de": {},
	"gh": {},
	"gi": {},
	"gr": {},
	"gl": {},
	"gd": {},
	"gp": {},
	"gu": {},
	"gt": {},
	"gg": {},
	"gn": {},
	"gw": {},
	"gy": {},
	"ht": {},
	"hm": {},
	"va": {},
	"hn": {},
	"hu": {},
	"is": {},
	"in": {},
	"id": {},
	"ir": {},
	"iq": {},
	"ie": {},
	"im": {},
	"il": {},
	"it": {},
	"jm": {},
	"jp": {},
	"je": {},
	"jo": {},
	"kz": {},
	"ke": {},
	"ki": {},
	"kw": {},
	"kg": {},
	"la": {},
	"lv": {},
	"lb": {},
	"ls": {},
	"lr": {},
	"ly": {},
	"li": {},
	"lt": {},
	"lu": {},
	"mg": {},
	"mw": {},
	"my": {},
	"mv": {},
	"ml": {},
	"mt": {},
	"mh": {},
	"mq": {},
	"mr": {},
	"mu": {},
	"yt": {},
	"mx": {},
	"fm": {},
	"mc": {},
	"mn": {},
	"me": {},
	"ms": {},
	"ma": {},
	"mz": {},
	"mm": {},
	"na": {},
	"nr": {},
	"np": {},
	"nl": {},
	"nc": {},
	"nz": {},
	"ni": {},
	"ne": {},
	"ng": {},
	"nu": {},
	"nf": {},
	"mk": {},
	"mp": {},
	"no": {},
	"om": {},
	"pk": {},
	"pw": {},
	"pa": {},
	"pg": {},
	"py": {},
	"pe": {},
	"ph": {},
	"pn": {},
	"pl": {},
	"pt": {},
	"pr": {},
	"qa": {},
	"kr": {},
	"md": {},
	"ro": {},
	"ru": {},
	"rw": {},
	"re": {},
	"bl": {},
	"sh": {},
	"mf": {},
	"pm": {},
	"vc": {},
	"ws": {},
	"sm": {},
	"st": {},
	"sa": {},
	"sn": {},
	"rs": {},
	"sc": {},
	"sl": {},
	"sg": {},
	"sx": {},
	"sk": {},
	"si": {},
	"sb": {},
	"so": {},
	"za": {},
	"gs": {},
	"ss": {},
	"es": {},
	"lk": {},
	"kn": {},
	"lc": {},
	"ps": {},
	"sd": {},
	"sr": {},
	"sj": {},
	"se": {},
	"ch": {},
	"sy": {},
	"tj": {},
	"th": {},
	"tl": {},
	"tg": {},
	"tk": {},
	"to": {},
	"tt": {},
	"tn": {},
	"tm": {},
	"tc": {},
	"tv": {},
	"tr": {},
	"ug": {},
	"ua": {},
	"ae": {},
	"gb": {},
	"tz": {},
	"um": {},
	"vi": {},
	"us": {},
	"uy": {},
	"uz": {},
	"vu": {},
	"ve": {},
	"vn": {},
	"wf": {},
	"eh": {},
	"ye": {},
	"zm": {},
	"zw": {},
	"ax": {},
}

// PaymentServiceProviderListQuery contains validated payment-service-provider list filters.
type PaymentServiceProviderListQuery struct {
	InstitutionType *string
}

// ValidatePaymentServiceProviderID validates the documented public payment-service-provider identifier.
func ValidatePaymentServiceProviderID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", requiredError("provider_id", "Provider ID is required.")
	}
	if !financePaymentServiceProviderIDPattern.MatchString(value) || uuidLikePattern.MatchString(value) {
		return "", invalidField("provider_id", "Provider ID must be a valid lowercase public slug.")
	}
	return value, nil
}

// ValidateInternationalMoneyTransferOperatorID validates the documented public IMTO identifier.
func ValidateInternationalMoneyTransferOperatorID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", requiredError("operator_id", "Operator ID is required.")
	}
	if !financeInternationalMoneyTransferOperatorIDPattern.MatchString(value) || uuidLikePattern.MatchString(value) {
		return "", invalidField("operator_id", "Operator ID must be a valid lowercase public slug.")
	}
	return value, nil
}

// ValidateCommercialBankID validates an approved lowercase commercial-bank identifier.
func ValidateCommercialBankID(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return requiredError("bank_id", "Bank ID is required.")
	}
	if !financeCommercialBankIDPattern.MatchString(value) || uuidLikePattern.MatchString(value) {
		return invalidField("bank_id", "Bank ID must be a valid lowercase public slug.")
	}
	if _, ok := approvedCommercialBankIDs[value]; !ok {
		return invalidField("bank_id", "Bank ID must reference a supported commercial bank.")
	}
	return nil
}

// ValidatePaymentServiceProviderType validates the documented public institution type.
func ValidatePaymentServiceProviderType(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", requiredError("institution_type", "Institution type is required.")
	}
	if _, ok := financePaymentServiceProviderTypes[value]; !ok {
		return "", invalidField("institution_type", "Institution type must be one of the supported payment-service-provider categories.")
	}
	return value, nil
}

// ValidatePaymentServiceProviderListQuery validates the documented finance list query.
func ValidatePaymentServiceProviderListQuery(values url.Values) (PaymentServiceProviderListQuery, error) {
	var errs ValidationErrors

	institutionTypeValues := values["institution_type"]
	if len(institutionTypeValues) > 1 {
		errs.Add("institution_type", codeMalformed, "Institution type may be provided at most once.")
		return PaymentServiceProviderListQuery{}, errs
	}
	if len(institutionTypeValues) == 0 {
		return PaymentServiceProviderListQuery{}, nil
	}

	normalized, err := ValidatePaymentServiceProviderType(institutionTypeValues[0])
	if err != nil {
		if validationErr, ok := err.(ValidationErrors); ok {
			errs.Fields = append(errs.Fields, validationErr.Fields...)
		} else {
			return PaymentServiceProviderListQuery{}, err
		}
	}
	if len(errs.Fields) > 0 {
		return PaymentServiceProviderListQuery{}, errs
	}

	query := PaymentServiceProviderListQuery{InstitutionType: &normalized}
	return query, nil
}

// ValidateCurrencyID validates the documented public currency identifier.
func ValidateCurrencyID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", requiredError("currency_id", "Currency ID is required.")
	}
	if !financeCurrencyIDPattern.MatchString(value) || uuidLikePattern.MatchString(value) {
		return "", invalidField("currency_id", "Currency ID must be a valid lowercase public slug.")
	}
	return value, nil
}

// ValidateCurrencyCountryAreaID validates the documented public currency country/area filter.
func ValidateCurrencyCountryAreaID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", requiredError("country_area_id", "Country or area ID is required.")
	}
	if !financeCurrencyCountryAreaIDPattern.MatchString(value) || uuidLikePattern.MatchString(value) {
		return "", invalidField("country_area_id", "Country or area ID must be a valid lowercase alpha-2 code.")
	}
	if _, ok := validCurrencyCountryAreaIDs[value]; !ok {
		return "", invalidField("country_area_id", "Country or area ID must reference a supported public country or area.")
	}
	return value, nil
}

// ValidateCurrencyListQuery validates the documented currency list query.
func ValidateCurrencyListQuery(values url.Values) (services.CurrencyListInput, error) {
	var errs ValidationErrors

	countryAreaValues := values["country_area_id"]
	if len(countryAreaValues) > 1 {
		errs.Add("country_area_id", codeMalformed, "Country or area ID may be provided at most once.")
		return services.CurrencyListInput{}, errs
	}
	if len(countryAreaValues) == 0 {
		return services.CurrencyListInput{}, nil
	}

	normalized, err := ValidateCurrencyCountryAreaID(countryAreaValues[0])
	if err != nil {
		if validationErr, ok := err.(ValidationErrors); ok {
			errs.Fields = append(errs.Fields, validationErr.Fields...)
		} else {
			return services.CurrencyListInput{}, err
		}
	}
	if len(errs.Fields) > 0 {
		return services.CurrencyListInput{}, errs
	}

	query := services.CurrencyListInput{CountryAreaID: normalized}
	return query, nil
}
