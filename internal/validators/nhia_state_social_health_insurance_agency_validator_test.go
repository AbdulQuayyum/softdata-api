package validators

import (
	"net/url"
	"strings"
	"testing"
)

func TestNHIAStateSocialHealthInsuranceAgencyListQueryValidator(t *testing.T) {
	got, err := ValidateNHIAStateSocialHealthInsuranceAgencyListQuery(url.Values{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Page != 1 || got.PageSize != 50 {
		t.Fatalf("defaults=%#v", got)
	}
	valid := url.Values{"page": {"2"}, "page_size": {"100"}, "state_id": {" fct "}, "search": {" Scheme "}}
	got, err = ValidateNHIAStateSocialHealthInsuranceAgencyListQuery(valid)
	if err != nil {
		t.Fatal(err)
	}
	if got.Page != 2 || got.PageSize != 100 || got.StateID != "fct" || got.Search != "Scheme" {
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
		{"state alias aks", "state_id=AKS"},
		{"state alias crs", "state_id=CRS"},
		{"state alias river", "state_id=River+State"},
		{"state syntax", "state_id=akwa_ibom"},
		{"state unsupported", "state_id=unknown-state"},
		{"state empty", "state_id=%20"},
		{"search empty", "search=%20"},
		{"search too long", "search=" + strings.Repeat("a", 101)},
		{"unsupported", "website=true"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tc.query)
			if _, err := ValidateNHIAStateSocialHealthInsuranceAgencyListQuery(values); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyIDValidator(t *testing.T) {
	id := strings.Repeat("a", 255)
	if got, err := ValidateNHIAStateSocialHealthInsuranceAgencyID("agency_id", id); err != nil || got != id {
		t.Fatalf("valid 255 id got=%q err=%v", got, err)
	}
	for _, id := range []string{"", "Bad", "bad_id", "has space", "bad%2Fid", "bad/id", "..", "id?x=1", "id#frag", strings.Repeat("a", 256)} {
		if _, err := ValidateNHIAStateSocialHealthInsuranceAgencyID("agency_id", id); err == nil {
			t.Fatalf("expected error for %q", id)
		}
	}
}
