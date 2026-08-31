package validators

import (
	"errors"
	"net/url"
	"testing"
)

func TestValidateStateID(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "abia", value: "abia", want: "abia"},
		{name: "akwa-ibom", value: " akwa-ibom ", want: "akwa-ibom"},
		{name: "cross-river", value: "cross-river", want: "cross-river"},
		{name: "fct", value: "fct", want: "fct"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateStateID(tc.value)
			if err != nil {
				t.Fatalf("ValidateStateID() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidateStateIDRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		field string
	}{
		{name: "empty", value: "", field: "state_id"},
		{name: "whitespace", value: "   ", field: "state_id"},
		{name: "uppercase", value: "Abia", field: "state_id"},
		{name: "mixedcase", value: "Akwa-ibom", field: "state_id"},
		{name: "underscore", value: "akwa_ibom", field: "state_id"},
		{name: "spaces", value: "akwa ibom", field: "state_id"},
		{name: "dots", value: "abia.json", field: "state_id"},
		{name: "traversal", value: "../abia", field: "state_id"},
		{name: "slash", value: "abia/state", field: "state_id"},
		{name: "backslash", value: "abia\\state", field: "state_id"},
		{name: "numeric", value: "0", field: "state_id"},
		{name: "iso", value: "NG-AB", field: "state_id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateStateID(tc.value)
			if err == nil {
				t.Fatalf("ValidateStateID() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateStateID() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != tc.field {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateGeopoliticalZoneID(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "north-central", value: "north-central", want: "north-central"},
		{name: "north east trimmed", value: " north-east ", want: "north-east"},
		{name: "south-west", value: "south-west", want: "south-west"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateGeopoliticalZoneID(tc.value)
			if err != nil {
				t.Fatalf("ValidateGeopoliticalZoneID() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidateGeopoliticalZoneIDRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
		{name: "uppercase", value: "North Central"},
		{name: "mixedcase", value: "North-central"},
		{name: "underscore", value: "north_central"},
		{name: "slash", value: "north/central"},
		{name: "space", value: "north central"},
		{name: "unsupported", value: "central"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateGeopoliticalZoneID(tc.value)
			if err == nil {
				t.Fatalf("ValidateGeopoliticalZoneID() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateGeopoliticalZoneID() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "zone_id" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateLocalGovernmentUnitListQuery(t *testing.T) {
	values := url.Values{"state_id": []string{" lagos "}}
	want := values.Encode()

	query, err := ValidateLocalGovernmentUnitListQuery(url.Values{})
	if err != nil {
		t.Fatalf("ValidateLocalGovernmentUnitListQuery() error = %v", err)
	}
	if query.StateID != nil {
		t.Fatalf("unexpected state filter: %#v", query.StateID)
	}

	query, err = ValidateLocalGovernmentUnitListQuery(values)
	if err != nil {
		t.Fatalf("ValidateLocalGovernmentUnitListQuery() error = %v", err)
	}
	if query.StateID == nil || *query.StateID != "lagos" {
		t.Fatalf("unexpected state filter: %#v", query.StateID)
	}
	if got := values.Encode(); got != want {
		t.Fatalf("query values mutated: got %q want %q", got, want)
	}
}

func TestValidateLocalGovernmentUnitListQueryRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value []string
		field string
	}{
		{name: "empty", value: []string{""}, field: "state_id"},
		{name: "whitespace", value: []string{"   "}, field: "state_id"},
		{name: "uppercase", value: []string{"Lagos"}, field: "state_id"},
		{name: "mixedcase", value: []string{"Akwa-ibom"}, field: "state_id"},
		{name: "name", value: []string{"Lagos State"}, field: "state_id"},
		{name: "uuid", value: []string{"550e8400-e29b-41d4-a716-446655440000"}, field: "state_id"},
		{name: "unsupported", value: []string{"north-central"}, field: "state_id"},
		{name: "duplicate", value: []string{"lagos", "fct"}, field: "state_id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateLocalGovernmentUnitListQuery(url.Values{"state_id": tc.value})
			if err == nil {
				t.Fatal("expected validation error")
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateLocalGovernmentUnitListQuery() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) == 0 || validationErr.Fields[0].Field != tc.field {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateLocalGovernmentUnitID(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "lagos", value: "lagos-ikeja", want: "lagos-ikeja"},
		{name: "fct", value: " fct-abuja-municipal ", want: "fct-abuja-municipal"},
		{name: "hyphenated state", value: "akwa-ibom-urue-offong-oruko", want: "akwa-ibom-urue-offong-oruko"},
		{name: "punctuated slug", value: "bauchi-jama-are", want: "bauchi-jama-are"},
		{name: "multiword slug", value: "plateau-qua-an-pan", want: "plateau-qua-an-pan"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateLocalGovernmentUnitID(tc.value)
			if err != nil {
				t.Fatalf("ValidateLocalGovernmentUnitID() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidateLocalGovernmentUnitIDRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
		{name: "uppercase", value: "LAGOS-IKEJA"},
		{name: "mixedcase", value: "Lagos-ikeja"},
		{name: "underscore", value: "lagos_ikeja"},
		{name: "slash", value: "lagos/ikeja"},
		{name: "apostrophe", value: "bauchi-jama'are"},
		{name: "embedded whitespace", value: "lagos ikeja"},
		{name: "state only", value: "fct"},
		{name: "uuid", value: "550e8400-e29b-41d4-a716-446655440000"},
		{name: "human name", value: "Aba North"},
		{name: "unsupported state", value: "north-central-ikeja"},
		{name: "trailing hyphen", value: "lagos-"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateLocalGovernmentUnitID(tc.value)
			if err == nil {
				t.Fatalf("ValidateLocalGovernmentUnitID() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateLocalGovernmentUnitID() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "lga_id" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateCountryOrAreaID(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "ng", value: "ng", want: "ng"},
		{name: "us trimmed", value: " us ", want: "us"},
		{name: "xk", value: "xk", want: "xk"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateCountryOrAreaID(tc.value)
			if err != nil {
				t.Fatalf("ValidateCountryOrAreaID() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidateCountryOrAreaIDRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		field string
	}{
		{name: "empty", value: "", field: "country_id"},
		{name: "whitespace", value: "   ", field: "country_id"},
		{name: "uppercase", value: "NG", field: "country_id"},
		{name: "mixedcase", value: "Ng", field: "country_id"},
		{name: "too short", value: "n", field: "country_id"},
		{name: "too long", value: "nga", field: "country_id"},
		{name: "numeric", value: "56", field: "country_id"},
		{name: "with hyphen", value: "n-g", field: "country_id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateCountryOrAreaID(tc.value)
			if err == nil {
				t.Fatalf("ValidateCountryOrAreaID() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateCountryOrAreaID() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != tc.field {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateCountryOrAreaRegionAndSubregionCodes(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "region", value: "002", want: "002"},
		{name: "region trimmed", value: " 150 ", want: "150"},
		{name: "subregion", value: "015", want: "015"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateCountryOrAreaRegionCode(tc.value)
			if err == nil {
				if got != tc.want {
					t.Fatalf("unexpected region code: %q", got)
				}
			}
			got, err = ValidateCountryOrAreaSubregionCode(tc.value)
			if err == nil {
				if got != tc.want {
					t.Fatalf("unexpected subregion code: %q", got)
				}
			}
		})
	}
}

func TestValidateCountryOrAreaRegionAndSubregionCodesRejectInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		field string
	}{
		{name: "empty region", value: "", field: "region_code"},
		{name: "whitespace region", value: "   ", field: "region_code"},
		{name: "bad region", value: "02", field: "region_code"},
		{name: "alpha region", value: "ABC", field: "region_code"},
		{name: "signed region", value: "+02", field: "region_code"},
		{name: "decimal region", value: "02.0", field: "region_code"},
		{name: "empty subregion", value: "", field: "subregion_code"},
		{name: "whitespace subregion", value: "   ", field: "subregion_code"},
		{name: "bad subregion", value: "2", field: "subregion_code"},
		{name: "alpha subregion", value: "ABC", field: "subregion_code"},
		{name: "signed subregion", value: "-02", field: "subregion_code"},
		{name: "decimal subregion", value: "02.0", field: "subregion_code"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			if tc.field == "region_code" {
				_, err = ValidateCountryOrAreaRegionCode(tc.value)
			} else {
				_, err = ValidateCountryOrAreaSubregionCode(tc.value)
			}
			if err == nil {
				t.Fatal("expected validation error")
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("unexpected error type: %v", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != tc.field {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateCountryOrAreaListQuery(t *testing.T) {
	query, err := ValidateCountryOrAreaListQuery(url.Values{})
	if err != nil {
		t.Fatalf("ValidateCountryOrAreaListQuery() error = %v", err)
	}
	if query.RegionCode != "" || query.SubregionCode != "" {
		t.Fatalf("unexpected empty query: %#v", query)
	}

	values := url.Values{"region_code": []string{" 002 ", "ignored"}, "subregion_code": []string{" 015 "}}
	query, err = ValidateCountryOrAreaListQuery(values)
	if err == nil {
		t.Fatal("expected validation error for duplicate region_code")
	}
	var validationErr ValidationErrors
	if !errors.As(err, &validationErr) {
		t.Fatalf("unexpected error type: %v", err)
	}
	if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "region_code" {
		t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
	}

	values = url.Values{"region_code": []string{"002"}, "subregion_code": []string{"015"}}
	query, err = ValidateCountryOrAreaListQuery(values)
	if err != nil {
		t.Fatalf("ValidateCountryOrAreaListQuery() error = %v", err)
	}
	if query.RegionCode != "002" || query.SubregionCode != "015" {
		t.Fatalf("unexpected query: %#v", query)
	}

	values = url.Values{"subregion_code": []string{"019"}}
	query, err = ValidateCountryOrAreaListQuery(values)
	if err != nil {
		t.Fatalf("ValidateCountryOrAreaListQuery() error = %v", err)
	}
	if query.RegionCode != "" || query.SubregionCode != "019" {
		t.Fatalf("unexpected subregion-only query: %#v", query)
	}
}

func TestValidateCountryOrAreaListQueryRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name   string
		values url.Values
		field  string
	}{
		{name: "empty region", values: url.Values{"region_code": []string{""}}, field: "region_code"},
		{name: "whitespace region", values: url.Values{"region_code": []string{"   "}}, field: "region_code"},
		{name: "bad region", values: url.Values{"region_code": []string{"02"}}, field: "region_code"},
		{name: "empty subregion", values: url.Values{"subregion_code": []string{""}}, field: "subregion_code"},
		{name: "whitespace subregion", values: url.Values{"subregion_code": []string{"   "}}, field: "subregion_code"},
		{name: "bad subregion", values: url.Values{"subregion_code": []string{"2"}}, field: "subregion_code"},
		{name: "duplicate subregion", values: url.Values{"subregion_code": []string{"002", "019"}}, field: "subregion_code"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateCountryOrAreaListQuery(tc.values)
			if err == nil {
				t.Fatal("expected validation error")
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("unexpected error type: %v", err)
			}
			if len(validationErr.Fields) == 0 || validationErr.Fields[0].Field != tc.field {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}
