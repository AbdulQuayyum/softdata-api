package validators

import (
	"errors"
	"net/url"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

func TestValidateCollegeOfEducationID(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "trimmed", value: "  federal-college-of-education-zaria ", want: "federal-college-of-education-zaria"},
		{name: "hyphenated", value: "saadatu-rimi-college-of-education", want: "saadatu-rimi-college-of-education"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateCollegeOfEducationID(tc.value)
			if err != nil {
				t.Fatalf("ValidateCollegeOfEducationID() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidateCollegeOfEducationIDRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
		{name: "uppercase", value: "Federal-College-Of-Education-Zaria"},
		{name: "mixedcase", value: "federal-College-of-education-zaria"},
		{name: "underscore", value: "federal_college_of_education_zaria"},
		{name: "spaces", value: "federal college of education zaria"},
		{name: "leading hyphen", value: "-federal-college-of-education-zaria"},
		{name: "trailing hyphen", value: "federal-college-of-education-zaria-"},
		{name: "double hyphen", value: "federal--college-of-education-zaria"},
		{name: "uuid", value: "550e8400-e29b-41d4-a716-446655440000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateCollegeOfEducationID(tc.value)
			if err == nil {
				t.Fatalf("ValidateCollegeOfEducationID() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateCollegeOfEducationID() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "college_id" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateCollegeOfEducationOwnershipType(t *testing.T) {
	for _, value := range []string{"federal", "state", "private"} {
		t.Run(value, func(t *testing.T) {
			got, err := ValidateCollegeOfEducationOwnershipType(" " + value + " ")
			if err != nil {
				t.Fatalf("ValidateCollegeOfEducationOwnershipType() error = %v", err)
			}
			if got != value {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidateCollegeOfEducationOwnershipTypeRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
		{name: "display label", value: "Federal College of Education"},
		{name: "uppercase", value: "FEDERAL"},
		{name: "mixedcase", value: "Federal"},
		{name: "unknown", value: "joint"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateCollegeOfEducationOwnershipType(tc.value)
			if err == nil {
				t.Fatalf("ValidateCollegeOfEducationOwnershipType() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateCollegeOfEducationOwnershipType() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "ownership_type" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateCollegeOfEducationStateID(t *testing.T) {
	for _, value := range []string{"abia", "fct", "lagos", "taraba"} {
		t.Run(value, func(t *testing.T) {
			got, err := ValidateCollegeOfEducationStateID(" " + value + " ")
			if err != nil {
				t.Fatalf("ValidateCollegeOfEducationStateID() error = %v", err)
			}
			if got != value {
				t.Fatalf("unexpected value: %q", got)
			}
		})
	}
}

func TestValidateCollegeOfEducationStateIDRejectsInvalidValues(t *testing.T) {
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
			got, err := ValidateCollegeOfEducationStateID(tc.value)
			if err == nil {
				t.Fatalf("ValidateCollegeOfEducationStateID() got %q, want error", got)
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateCollegeOfEducationStateID() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) != 1 || validationErr.Fields[0].Field != "state_id" {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}

func TestValidateCollegeOfEducationListQuery(t *testing.T) {
	values := url.Values{}
	want := values.Encode()
	query, err := ValidateCollegeOfEducationListQuery(values)
	if err != nil {
		t.Fatalf("ValidateCollegeOfEducationListQuery() error = %v", err)
	}
	if query != (services.CollegeOfEducationListInput{}) {
		t.Fatalf("unexpected query: %#v", query)
	}
	if got := values.Encode(); got != want {
		t.Fatalf("query values mutated: got %q want %q", got, want)
	}

	values = url.Values{
		"ownership_type": []string{" private "},
		"state_id":       []string{" lagos "},
	}
	query, err = ValidateCollegeOfEducationListQuery(values)
	if err != nil {
		t.Fatalf("ValidateCollegeOfEducationListQuery() error = %v", err)
	}
	if query.OwnershipType != "private" {
		t.Fatalf("unexpected ownership filter: %#v", query.OwnershipType)
	}
	if query.StateID != "lagos" {
		t.Fatalf("unexpected state filter: %#v", query.StateID)
	}
}

func TestValidateCollegeOfEducationListQueryRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value url.Values
		field string
	}{
		{name: "duplicate ownership", value: url.Values{"ownership_type": []string{"federal", "private"}}, field: "ownership_type"},
		{name: "duplicate state", value: url.Values{"state_id": []string{"abia", "fct"}}, field: "state_id"},
		{name: "empty ownership", value: url.Values{"ownership_type": []string{""}}, field: "ownership_type"},
		{name: "uppercase ownership", value: url.Values{"ownership_type": []string{"Federal"}}, field: "ownership_type"},
		{name: "display ownership", value: url.Values{"ownership_type": []string{"Federal College of Education"}}, field: "ownership_type"},
		{name: "empty state", value: url.Values{"state_id": []string{""}}, field: "state_id"},
		{name: "uppercase state", value: url.Values{"state_id": []string{"ABIA"}}, field: "state_id"},
		{name: "unknown state", value: url.Values{"state_id": []string{"zaria"}}, field: "state_id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateCollegeOfEducationListQuery(tc.value)
			if err == nil {
				t.Fatal("expected validation error")
			}
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("ValidateCollegeOfEducationListQuery() error = %v, want ValidationErrors", err)
			}
			if len(validationErr.Fields) == 0 || validationErr.Fields[0].Field != tc.field {
				t.Fatalf("unexpected validation errors: %#v", validationErr.Fields)
			}
		})
	}
}
