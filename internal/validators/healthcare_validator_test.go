package validators

import (
	"net/url"
	"strings"
	"testing"
)

func TestValidateHealthFacilityID(t *testing.T) {
	if got, err := ValidateHealthFacilityID("facility_id", " example-facility "); err != nil || got != "example-facility" {
		t.Fatalf("valid ID = %q, %v", got, err)
	}
	for _, value := range []string{"", "UPPER", "leading-", "-trailing", "double--dash", "a/b", "../secret", "a?b", "a#b"} {
		if _, err := ValidateHealthFacilityID("facility_id", value); err == nil {
			t.Fatalf("ValidateHealthFacilityID(%q) accepted invalid value", value)
		}
	}
}

func TestValidateHealthFacilityListQuery(t *testing.T) {
	query, err := ValidateHealthFacilityListQuery(url.Values{
		"page": {"2"}, "page_size": {"100"}, "state_id": {" lagos "}, "lga_id": {" ikeja "},
		"facility_type": {" clinic "}, "facility_level": {" primary "}, "ownership_type": {" private "}, "search": {" clinic "},
	})
	if err != nil {
		t.Fatal(err)
	}
	if query.Page != 2 || query.PageSize != 100 || query.StateID != "lagos" || query.LGAID != "ikeja" || query.FacilityType != "clinic" || query.FacilityLevel != "primary" || query.OwnershipType != "private" || query.Search != "clinic" {
		t.Fatalf("unexpected normalized query: %#v", query)
	}
	defaults, err := ValidateHealthFacilityListQuery(url.Values{"search": {"   "}})
	if err != nil || defaults.Page != 1 || defaults.PageSize != 50 || defaults.Search != "" {
		t.Fatalf("unexpected defaults: %#v, %v", defaults, err)
	}
	for _, raw := range []url.Values{
		{"page": {"0"}}, {"page": {"-1"}}, {"page": {"1.5"}}, {"page": {"+1"}}, {"page": {"999999999999999999999"}},
		{"page_size": {"101"}}, {"page_size": {"0"}}, {"state_id": {"Lagos"}}, {"lga_id": {"lagos/ikeja"}},
		{"facility_type": {"hospital"}}, {"facility_level": {"quaternary"}}, {"ownership_type": {"public"}}, {"search": {strings.Repeat("x", 101)}},
	} {
		if _, err := ValidateHealthFacilityListQuery(raw); err == nil {
			t.Fatalf("accepted invalid query: %#v", raw)
		}
	}
	if _, err := ValidateHealthFacilityListQuery(url.Values{"page": {"1", "2"}}); err == nil {
		t.Fatal("accepted duplicate page")
	}
}
