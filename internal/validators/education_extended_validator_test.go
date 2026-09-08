package validators

import (
	"net/url"
	"strings"
	"testing"
)

func TestValidateEducationInstitutionID(t *testing.T) {
	got, err := ValidateEducationInstitutionID("institution_id", "  sample-college  ")
	if err != nil || got != "sample-college" {
		t.Fatalf("got %q, %v", got, err)
	}
	for _, value := range []string{"", " ", "Sample-college", "-sample", "sample-", "sample--college", "sample/college", "sample%2fcollege", "sample?x"} {
		if _, err := ValidateEducationInstitutionID("institution_id", value); err == nil {
			t.Fatalf("value %q was accepted", value)
		}
	}
}

func TestValidateSchoolListQuery(t *testing.T) {
	query, err := ValidateSchoolListQuery(url.Values{})
	if err != nil || query.Page != 1 || query.PageSize != 50 {
		t.Fatalf("defaults: %#v, %v", query, err)
	}
	query, err = ValidateSchoolListQuery(url.Values{
		"page": {"2"}, "page_size": {"100"}, "state_id": {" lagos "}, "lga_id": {" ikeja "},
		"education_level": {"primary"}, "ownership_type": {"private"}, "search": {"  Community School  "},
	})
	if err != nil || query.Page != 2 || query.PageSize != 100 || query.StateID != "lagos" || query.LGAID != "ikeja" || query.Search != "Community School" {
		t.Fatalf("query: %#v, %v", query, err)
	}
	query, err = ValidateSchoolListQuery(url.Values{"state_id": {"  "}, "lga_id": {"  "}, "education_level": {"  "}, "ownership_type": {"  "}})
	if err != nil || query.StateID != "" || query.LGAID != "" || query.EducationLevel != "" || query.OwnershipType != "" {
		t.Fatalf("blank filters: %#v, %v", query, err)
	}
	for _, values := range []url.Values{
		{"page": {"0"}}, {"page": {"-1"}}, {"page": {"1.5"}}, {"page_size": {"101"}},
		{"page": {"999999999999999999999999"}}, {"education_level": {"tertiary"}},
		{"ownership_type": {"government"}}, {"search": {strings.Repeat("x", 101)}},
	} {
		if _, err := ValidateSchoolListQuery(values); err == nil {
			t.Fatalf("values accepted: %#v", values)
		}
	}
}
