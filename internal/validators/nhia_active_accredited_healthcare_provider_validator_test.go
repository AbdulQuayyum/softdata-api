package validators

import (
	"net/url"
	"strings"
	"testing"
)

func TestNHIAActiveAccreditedHealthcareProviderListQueryValidator(t *testing.T) {
	got, err := ValidateNHIAActiveAccreditedHealthcareProviderListQuery(url.Values{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Page != 1 || got.PageSize != 50 {
		t.Fatalf("defaults=%#v", got)
	}
	values := url.Values{"page": {"2"}, "page_size": {"100"}, "provider_code": {" FCT/0001/P "}, "facility_type": {" primary "}, "listing_status": {" active_accredited "}, "search": {" Clinic "}}
	got, err = ValidateNHIAActiveAccreditedHealthcareProviderListQuery(values)
	if err != nil {
		t.Fatal(err)
	}
	if got.Page != 2 || got.PageSize != 100 || got.ProviderCode != "FCT/0001/P" || got.FacilityType != "primary" || got.ListingStatus != "active_accredited" || got.Search != "Clinic" {
		t.Fatalf("normalized=%#v", got)
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
		{"provider code empty", "provider_code=%20"},
		{"provider code lowercase", "provider_code=fct/0001/p"},
		{"provider code malformed", "provider_code=FCT/1/P"},
		{"facility empty", "facility_type=%20"},
		{"facility invalid", "facility_type=secondary"},
		{"status empty", "listing_status=%20"},
		{"status invalid", "listing_status=licensed"},
		{"search empty", "search=%20"},
		{"search too long", "search=" + strings.Repeat("a", 101)},
		{"unsupported", "state_id=fct"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tc.query)
			if _, err := ValidateNHIAActiveAccreditedHealthcareProviderListQuery(values); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestNHIAActiveAccreditedHealthcareProviderIDValidator(t *testing.T) {
	id := strings.Repeat("a", 255)
	if got, err := ValidateNHIAActiveAccreditedHealthcareProviderID("provider_id", id); err != nil || got != id {
		t.Fatalf("valid 255 id got=%q err=%v", got, err)
	}
	for _, id := range []string{"", "Bad", "bad_id", "has space", "bad%2Fid", "bad/id", "..", "id?x=1", "id#frag", strings.Repeat("a", 256)} {
		if _, err := ValidateNHIAActiveAccreditedHealthcareProviderID("provider_id", id); err == nil {
			t.Fatalf("expected error for %q", id)
		}
	}
}
