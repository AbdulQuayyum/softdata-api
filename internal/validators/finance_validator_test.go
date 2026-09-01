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

func TestValidateInternationalMoneyTransferOperatorID(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "trimmed", value: " olive-monies-express-limited ", want: "olive-monies-express-limited"},
		{name: "single token", value: "nouveau", want: "nouveau"},
		{name: "multi token", value: "nouveau-mobile-limited", want: "nouveau-mobile-limited"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateInternationalMoneyTransferOperatorID(tc.value)
			if err != nil {
				t.Fatalf("ValidateInternationalMoneyTransferOperatorID() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidateInternationalMoneyTransferOperatorIDRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
		{name: "uppercase", value: "Olive-Monies-Express-Limited"},
		{name: "mixedcase", value: "olive-Monies-express-limited"},
		{name: "underscore", value: "olive_monies_express_limited"},
		{name: "spaces", value: "olive monies express limited"},
		{name: "leading hyphen", value: "-olive-monies-express-limited"},
		{name: "trailing hyphen", value: "olive-monies-express-limited-"},
		{name: "double hyphen", value: "olive--monies-express-limited"},
		{name: "uuid", value: "550e8400-e29b-41d4-a716-446655440000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateInternationalMoneyTransferOperatorID(tc.value)
			if err == nil {
				t.Fatalf("ValidateInternationalMoneyTransferOperatorID() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateInternationalMoneyTransferOperatorID() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "operator_id" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateCurrencyID(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "trimmed", value: " ngn ", want: "ngn"},
		{name: "usd", value: "usd", want: "usd"},
		{name: "twd", value: "twd", want: "twd"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateCurrencyID(tc.value)
			if err != nil {
				t.Fatalf("ValidateCurrencyID() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidateCurrencyIDRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
		{name: "uppercase", value: "NGN"},
		{name: "mixedcase", value: "NgN"},
		{name: "underscore", value: "ng_n"},
		{name: "hyphen", value: "ng-n"},
		{name: "digit", value: "ng1"},
		{name: "uuid", value: "550e8400-e29b-41d4-a716-446655440000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateCurrencyID(tc.value)
			if err == nil {
				t.Fatalf("ValidateCurrencyID() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateCurrencyID() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "currency_id" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateCurrencyCountryAreaID(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "trimmed", value: " ng ", want: "ng"},
		{name: "antarctica", value: "aq", want: "aq"},
		{name: "palestine", value: "ps", want: "ps"},
		{name: "south georgia", value: "gs", want: "gs"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateCurrencyCountryAreaID(tc.value)
			if err != nil {
				t.Fatalf("ValidateCurrencyCountryAreaID() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidateCurrencyCountryAreaIDRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
		{name: "uppercase", value: "NG"},
		{name: "mixedcase", value: "nG"},
		{name: "unknown", value: "zz"},
		{name: "digit", value: "1g"},
		{name: "underscore", value: "n_g"},
		{name: "hyphen", value: "n-g"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateCurrencyCountryAreaID(tc.value)
			if err == nil {
				t.Fatalf("ValidateCurrencyCountryAreaID() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateCurrencyCountryAreaID() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "country_area_id" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateCurrencyListQuery(t *testing.T) {
	values := url.Values{}
	query, err := ValidateCurrencyListQuery(values)
	if err != nil {
		t.Fatalf("ValidateCurrencyListQuery() error = %v", err)
	}
	if query.CountryAreaID != "" {
		t.Fatalf("unexpected filter: %#v", query)
	}

	values = url.Values{"country_area_id": []string{" ng "}}
	want := values.Encode()
	query, err = ValidateCurrencyListQuery(values)
	if err != nil {
		t.Fatalf("ValidateCurrencyListQuery() error = %v", err)
	}
	if query.CountryAreaID != "ng" {
		t.Fatalf("unexpected filter: %#v", query)
	}
	if got := values.Encode(); got != want {
		t.Fatalf("query values mutated: got %q want %q", got, want)
	}
}

func TestValidateCurrencyListQueryRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name   string
		values url.Values
	}{
		{name: "empty explicit", values: url.Values{"country_area_id": []string{""}}},
		{name: "whitespace", values: url.Values{"country_area_id": []string{"   "}}},
		{name: "uppercase", values: url.Values{"country_area_id": []string{"NG"}}},
		{name: "unknown", values: url.Values{"country_area_id": []string{"zz"}}},
		{name: "duplicate", values: url.Values{"country_area_id": []string{"ng", "bt"}}},
		{name: "extra unknown parameter", values: url.Values{"country_area_id": []string{"ng"}, "page": []string{"1"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query, err := ValidateCurrencyListQuery(tc.values)
			if tc.name == "extra unknown parameter" {
				if err != nil {
					t.Fatalf("ValidateCurrencyListQuery() error = %v", err)
				}
				if query.CountryAreaID != "ng" {
					t.Fatalf("unexpected filter: %#v", query)
				}
				return
			}
			if err == nil {
				t.Fatal("expected validation error")
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateCurrencyListQuery() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) == 0 || validationErr.Fields[0].Field != "country_area_id" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}
