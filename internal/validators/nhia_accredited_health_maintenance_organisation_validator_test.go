package validators

import (
	"net/url"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestNHIAAccreditedHMOListQueryValidation(t *testing.T) {
	q, err := ValidateNHIAAccreditedHealthMaintenanceOrganisationListQuery(url.Values{
		"page":                 {"2"},
		"page_size":            {"100"},
		"accreditation_status": {" accredited "},
		"hmo_id":               {" 0012 "},
		"search":               {" Example HMO "},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Page: 2, PageSize: 100, AccreditationStatus: "accredited", HMOID: "0012", Search: "Example HMO"}
	if q != want {
		t.Fatalf("query = %#v, want %#v", q, want)
	}
}

func TestNHIAAccreditedHMOListQueryRejectsInvalidValues(t *testing.T) {
	cases := []url.Values{
		{"unexpected": {"x"}},
		{"accreditation_status": {"expired"}},
		{"accreditation_status": {"  "}},
		{"hmo_id": {"  "}},
		{"hmo_id": {strings.Repeat("x", 65)}},
		{"hmo_id": {"12\n0"}},
		{"search": {"  "}},
		{"search": {strings.Repeat("x", 101)}},
		{"search": {"a\x00b"}},
	}
	for _, field := range []string{"page", "page_size"} {
		for _, value := range []string{"", "0", "-1", "+1", "1.5", "abc", "18446744073709551616", "9223372036854775808", " 1"} {
			cases = append(cases, url.Values{field: {value}})
		}
	}
	cases = append(cases, url.Values{"page_size": {"101"}})
	for _, field := range []string{"page", "page_size", "accreditation_status", "hmo_id", "search"} {
		cases = append(cases, url.Values{field: {"1", "2"}})
	}
	for _, values := range cases {
		t.Run(values.Encode(), func(t *testing.T) {
			if _, err := ValidateNHIAAccreditedHealthMaintenanceOrganisationListQuery(values); err == nil {
				t.Fatal("invalid query accepted")
			}
		})
	}
}

func TestNHIAAccreditedHMOIDValidation(t *testing.T) {
	valid := []string{"a-and-m-healthcare-trust-limited-102", strings.Repeat("a", 255)}
	for _, id := range valid {
		got, err := ValidateNHIAAccreditedHealthMaintenanceOrganisationID("organisation_id", id)
		if err != nil || got != id {
			t.Fatalf("valid id %q = %q, %v", id, got, err)
		}
	}
	for _, id := range []string{"", "Upper", "bad_id", "bad/id", "bad%2Fid", "bad\\id", "bad--id", "-bad", "bad-", " id ", strings.Repeat("a", 256)} {
		if _, err := ValidateNHIAAccreditedHealthMaintenanceOrganisationID("organisation_id", id); err == nil {
			t.Fatalf("invalid id %q accepted", id)
		}
	}
}
