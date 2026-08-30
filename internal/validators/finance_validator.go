package validators

import (
	"net/url"
	"regexp"
	"strings"
)

var financePaymentServiceProviderIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)
var financeInternationalMoneyTransferOperatorIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var financePaymentServiceProviderTypes = map[string]struct{}{
	"mobile_money_operator":               {},
	"switching_and_processing_company":    {},
	"payment_solution_service_provider":   {},
	"payment_terminal_service_provider":   {},
	"super_agent":                         {},
	"payment_service_holding_company":     {},
	"payment_terminal_service_aggregator": {},
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
