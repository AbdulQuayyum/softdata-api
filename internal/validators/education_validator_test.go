package validators

import (
	"errors"
	"net/url"
	"testing"
)

func TestValidateUniversityID(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "trimmed", value: "  abubakar-tafawa-balewa-university-bauchi ", want: "abubakar-tafawa-balewa-university-bauchi"},
		{name: "hyphenated", value: "redeemers-university-ejigbo", want: "redeemers-university-ejigbo"},
		{name: "location suffix", value: "alex-ekwueme-university-ndufu-alike-ebonyi-state", want: "alex-ekwueme-university-ndufu-alike-ebonyi-state"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateUniversityID(tc.value)
			if err != nil {
				t.Fatalf("ValidateUniversityID() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidateUniversityIDRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
		{name: "uppercase", value: "Abubakar-Tafawa-Balewa-University-Bauchi"},
		{name: "mixedcase", value: "abubakar-tafawa-Balewa-university-bauchi"},
		{name: "underscore", value: "abubakar_tafawa_balewa_university_bauchi"},
		{name: "spaces", value: "abubakar tafawa balewa university bauchi"},
		{name: "leading hyphen", value: "-abubakar-tafawa-balewa-university-bauchi"},
		{name: "trailing hyphen", value: "abubakar-tafawa-balewa-university-bauchi-"},
		{name: "double hyphen", value: "abubakar--tafawa-balewa-university-bauchi"},
		{name: "uuid", value: "550e8400-e29b-41d4-a716-446655440000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateUniversityID(tc.value)
			if err == nil {
				t.Fatalf("ValidateUniversityID() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateUniversityID() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "university_id" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateUniversityOwnershipType(t *testing.T) {
	for _, value := range []string{"federal", "state", "private"} {
		t.Run(value, func(t *testing.T) {
			got, err := ValidateUniversityOwnershipType(" " + value + " ")
			if err != nil {
				t.Fatalf("ValidateUniversityOwnershipType() error = %v", err)
			}
			if got != value {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidateUniversityOwnershipTypeRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
		{name: "display label", value: "Federal University"},
		{name: "uppercase", value: "FEDERAL"},
		{name: "mixedcase", value: "Federal"},
		{name: "unknown", value: "joint"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateUniversityOwnershipType(tc.value)
			if err == nil {
				t.Fatalf("ValidateUniversityOwnershipType() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateUniversityOwnershipType() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "ownership_type" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateUniversityStateID(t *testing.T) {
	for _, value := range []string{"abia", "fct", "akwa-ibom", "cross-river"} {
		t.Run(value, func(t *testing.T) {
			got, err := ValidateUniversityStateID(" " + value + " ")
			if err != nil {
				t.Fatalf("ValidateUniversityStateID() error = %v", err)
			}
			if got != value {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidateUniversityStateIDRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
		{name: "uppercase", value: "ABIA"},
		{name: "mixedcase", value: "Akwa-Ibom"},
		{name: "unknown", value: "zaria"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateUniversityStateID(tc.value)
			if err == nil {
				t.Fatalf("ValidateUniversityStateID() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateUniversityStateID() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "state_id" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateUniversityListQuery(t *testing.T) {
	values := url.Values{}
	want := values.Encode()
	query, err := ValidateUniversityListQuery(values)
	if err != nil {
		t.Fatalf("ValidateUniversityListQuery() error = %v", err)
	}
	if query.OwnershipType != nil || query.StateID != nil {
		t.Fatalf("unexpected query: %#v", query)
	}
	if got := values.Encode(); got != want {
		t.Fatalf("query values mutated: got %q want %q", got, want)
	}

	values = url.Values{
		"ownership_type": []string{" private "},
		"state_id":       []string{" fct "},
	}
	query, err = ValidateUniversityListQuery(values)
	if err != nil {
		t.Fatalf("ValidateUniversityListQuery() error = %v", err)
	}
	if query.OwnershipType == nil || *query.OwnershipType != "private" {
		t.Fatalf("unexpected ownership filter: %#v", query.OwnershipType)
	}
	if query.StateID == nil || *query.StateID != "fct" {
		t.Fatalf("unexpected state filter: %#v", query.StateID)
	}
}

func TestValidateUniversityListQueryRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value url.Values
		field string
	}{
		{name: "duplicate ownership", value: url.Values{"ownership_type": []string{"federal", "private"}}, field: "ownership_type"},
		{name: "duplicate state", value: url.Values{"state_id": []string{"abia", "fct"}}, field: "state_id"},
		{name: "empty ownership", value: url.Values{"ownership_type": []string{""}}, field: "ownership_type"},
		{name: "uppercase ownership", value: url.Values{"ownership_type": []string{"Federal"}}, field: "ownership_type"},
		{name: "display ownership", value: url.Values{"ownership_type": []string{"Federal University"}}, field: "ownership_type"},
		{name: "empty state", value: url.Values{"state_id": []string{""}}, field: "state_id"},
		{name: "uppercase state", value: url.Values{"state_id": []string{"ABIA"}}, field: "state_id"},
		{name: "unknown state", value: url.Values{"state_id": []string{"zaria"}}, field: "state_id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateUniversityListQuery(tc.value)
			if err == nil {
				t.Fatal("expected validation error")
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateUniversityListQuery() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) == 0 || validationErr.Fields[0].Field != tc.field {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}
