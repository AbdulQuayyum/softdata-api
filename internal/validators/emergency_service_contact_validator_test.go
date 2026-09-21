package validators

import (
	"net/url"
	"strings"
	"testing"
)

func TestEmergencyServiceContactListQueryValidator(t *testing.T) {
	got, err := ValidateEmergencyServiceContactListQuery(url.Values{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Page != 1 || got.PageSize != 50 {
		t.Fatalf("defaults=%#v", got)
	}
	values := url.Values{
		"page": {"2"}, "page_size": {"100"}, "service_type": {" fire "}, "contact_type": {" short_code "},
		"coverage_type": {" national "}, "contact_value": {" 112 "}, "search": {" Federal Fire "},
	}
	got, err = ValidateEmergencyServiceContactListQuery(values)
	if err != nil {
		t.Fatal(err)
	}
	if got.Page != 2 || got.PageSize != 100 || got.ServiceType != "fire" || got.ContactType != "short_code" || got.CoverageType != "national" || got.ContactValue != "112" || got.Search != "Federal Fire" {
		t.Fatalf("normalized=%#v", got)
	}
	for _, serviceType := range []string{"general_emergency", "disaster_management", "road_emergency", "police", "fire", "ambulance", "other"} {
		if _, err := ValidateEmergencyServiceContactListQuery(url.Values{"service_type": {serviceType}}); err != nil {
			t.Fatalf("service_type=%q err=%v", serviceType, err)
		}
	}
	for _, contactType := range []string{"short_code", "telephone"} {
		if _, err := ValidateEmergencyServiceContactListQuery(url.Values{"contact_type": {contactType}}); err != nil {
			t.Fatalf("contact_type=%q err=%v", contactType, err)
		}
	}
	for _, coverageType := range []string{"national", "state"} {
		if _, err := ValidateEmergencyServiceContactListQuery(url.Values{"coverage_type": {coverageType}}); err != nil {
			t.Fatalf("coverage_type=%q err=%v", coverageType, err)
		}
	}
	for _, contactValue := range []string{"112", "01234567890", "080022556362", "+2348032003557"} {
		if got, err := ValidateEmergencyServiceContactListQuery(url.Values{"contact_value": {" " + contactValue + " "}}); err != nil || got.ContactValue != contactValue {
			t.Fatalf("contact_value=%q got=%#v err=%v", contactValue, got, err)
		}
	}
	if got, err := ValidateEmergencyServiceContactListQuery(url.Values{"search": {"   "}}); err != nil || got.Search != "" {
		t.Fatalf("empty search should be omitted: got=%#v err=%v", got, err)
	}
	if _, err := ValidateEmergencyServiceContactListQuery(url.Values{"search": {strings.Repeat("a", 100)}}); err != nil {
		t.Fatalf("max search err=%v", err)
	}
	for _, tc := range []struct {
		name  string
		query string
	}{
		{"page zero", "page=0"},
		{"page signed", "page=-1"},
		{"page decimal", "page=1.5"},
		{"page malformed", "page=abc"},
		{"page overflow", "page=999999999999999999999999999999"},
		{"page size zero", "page_size=0"},
		{"page size too large", "page_size=101"},
		{"page repeated", "page=1&page=2"},
		{"service type empty", "service_type=%20"},
		{"service type invalid", "service_type=office"},
		{"contact type empty", "contact_type=%20"},
		{"contact type invalid", "contact_type=email"},
		{"coverage type empty", "coverage_type=%20"},
		{"coverage type invalid", "coverage_type=lga"},
		{"contact value empty", "contact_value=%20"},
		{"contact value invalid", "contact_value=abc"},
		{"search too long", "search=" + strings.Repeat("a", 101)},
		{"unsupported", "state_id=fct"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tc.query)
			if _, err := ValidateEmergencyServiceContactListQuery(values); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestEmergencyServiceContactIDValidator(t *testing.T) {
	id := strings.Repeat("a", 255)
	if got, err := ValidateEmergencyServiceContactID("contact_id", id); err != nil || got != id {
		t.Fatalf("valid 255 id got=%q err=%v", got, err)
	}
	for _, id := range []string{"", "Bad", "bad_id", "has space", "bad%2Fid", "bad/id", "..", "id?x=1", "id#frag", strings.Repeat("a", 256)} {
		if _, err := ValidateEmergencyServiceContactID("contact_id", id); err == nil {
			t.Fatalf("expected error for %q", id)
		}
	}
}
