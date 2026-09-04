package validators

import (
	"errors"
	"net/url"
	"testing"
)

func TestValidateLanguageID(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  string
		valid bool
	}{
		{value: "en", want: "en", valid: true},
		{value: "  fil ", want: "fil", valid: true},
		{value: "fat", valid: false},
		{value: "sr-Latn", valid: false},
		{value: "EN", valid: false},
		{value: "", valid: false},
	} {
		got, err := ValidateLanguageID(tc.value)
		if tc.valid {
			if err != nil || got != tc.want {
				t.Fatalf("ValidateLanguageID(%q) = %q, %v", tc.value, got, err)
			}
		} else if err == nil {
			t.Fatalf("ValidateLanguageID(%q) unexpectedly succeeded", tc.value)
		}
	}
}

func TestValidateCountryLanguageListQuery(t *testing.T) {
	query, err := ValidateCountryLanguageListQuery(url.Values{
		"country_area_id": {" ng "},
		"language_id":     {" yo "},
		"status":          {" official "},
	})
	if err != nil {
		t.Fatalf("ValidateCountryLanguageListQuery() error = %v", err)
	}
	if query.CountryAreaID != "ng" || query.LanguageID != "yo" || query.Status != "official" {
		t.Fatalf("unexpected normalized query: %#v", query)
	}

	for _, values := range []url.Values{
		{"language_id": {"en", "yo"}},
		{"language_id": {""}},
		{"language_id": {"en-US"}},
		{"country_area_id": {"NG"}},
		{"status": {"Official"}},
	} {
		if _, err := ValidateCountryLanguageListQuery(values); err == nil {
			t.Fatalf("ValidateCountryLanguageListQuery(%v) unexpectedly succeeded", values)
		} else {
			var validationErr ValidationErrors
			if !errors.As(err, &validationErr) {
				t.Fatalf("error %v is not a validation error", err)
			}
		}
	}
}
