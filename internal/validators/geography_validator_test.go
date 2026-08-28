package validators

import (
	"errors"
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
