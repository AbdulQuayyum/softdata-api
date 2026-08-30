package validators

import (
	"errors"
	"net/url"
	"testing"
)

func TestValidatePaymentServiceProviderID(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "trimmed", value: " paystack-limited ", want: "paystack-limited"},
		{name: "hyphenated", value: "moniepoint-mfb", want: "moniepoint-mfb"},
		{name: "numbers", value: "provider-2", want: "provider-2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidatePaymentServiceProviderID(tc.value)
			if err != nil {
				t.Fatalf("ValidatePaymentServiceProviderID() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidatePaymentServiceProviderIDRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
		{name: "uppercase", value: "Paystack-Limited"},
		{name: "mixedcase", value: "payStack-limited"},
		{name: "underscore", value: "paystack_limited"},
		{name: "spaces", value: "paystack limited"},
		{name: "leading hyphen", value: "-paystack-limited"},
		{name: "trailing hyphen", value: "paystack-limited-"},
		{name: "double hyphen", value: "paystack--limited"},
		{name: "uuid", value: "550e8400-e29b-41d4-a716-446655440000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidatePaymentServiceProviderID(tc.value)
			if err == nil {
				t.Fatalf("ValidatePaymentServiceProviderID() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidatePaymentServiceProviderID() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "provider_id" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidatePaymentServiceProviderType(t *testing.T) {
	for _, value := range []string{
		"mobile_money_operator",
		"switching_and_processing_company",
		"payment_solution_service_provider",
		"payment_terminal_service_provider",
		"super_agent",
		"payment_service_holding_company",
		"payment_terminal_service_aggregator",
	} {
		t.Run(value, func(t *testing.T) {
			got, err := ValidatePaymentServiceProviderType(" " + value + " ")
			if err != nil {
				t.Fatalf("ValidatePaymentServiceProviderType() error = %v", err)
			}
			if got != value {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidatePaymentServiceProviderTypeRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
		{name: "display label", value: "Payment Service Provider"},
		{name: "display plural", value: "Mobile Money Operators"},
		{name: "hyphenated", value: "mobile-money-operator"},
		{name: "uppercase", value: "SUPER_AGENT"},
		{name: "unknown", value: "merchant_service_provider"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidatePaymentServiceProviderType(tc.value)
			if err == nil {
				t.Fatalf("ValidatePaymentServiceProviderType() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidatePaymentServiceProviderType() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "institution_type" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidatePaymentServiceProviderListQuery(t *testing.T) {
	values := url.Values{}
	query, err := ValidatePaymentServiceProviderListQuery(values)
	if err != nil {
		t.Fatalf("ValidatePaymentServiceProviderListQuery() error = %v", err)
	}
	if query.InstitutionType != nil {
		t.Fatalf("unexpected filter: %#v", query.InstitutionType)
	}

	values = url.Values{"institution_type": []string{" mobile_money_operator "}}
	want := values.Encode()
	query, err = ValidatePaymentServiceProviderListQuery(values)
	if err != nil {
		t.Fatalf("ValidatePaymentServiceProviderListQuery() error = %v", err)
	}
	if query.InstitutionType == nil || *query.InstitutionType != "mobile_money_operator" {
		t.Fatalf("unexpected filter: %#v", query.InstitutionType)
	}
	if got := values.Encode(); got != want {
		t.Fatalf("query values mutated: got %q want %q", got, want)
	}
}

func TestValidatePaymentServiceProviderListQueryRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value []string
	}{
		{name: "empty", value: []string{""}},
		{name: "whitespace", value: []string{"   "}},
		{name: "uppercase", value: []string{"Mobile Money Operator"}},
		{name: "duplicate", value: []string{"mobile_money_operator", "super_agent"}},
		{name: "display label", value: []string{"Mobile Money Operators"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidatePaymentServiceProviderListQuery(url.Values{"institution_type": tc.value})
			if err == nil {
				t.Fatal("expected validation error")
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidatePaymentServiceProviderListQuery() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) == 0 || validationErr.Fields[0].Field != "institution_type" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}
