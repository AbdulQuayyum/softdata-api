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
