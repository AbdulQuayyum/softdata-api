package validators

import (
	"net/url"
	"strings"
	"testing"
)

func TestNEMAZonalTerritorialOperationOfficeListQueryValidator(t *testing.T) {
	got, err := ValidateNEMAZonalTerritorialOperationOfficeListQuery(url.Values{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Page != 1 || got.PageSize != 50 {
		t.Fatalf("defaults=%#v", got)
	}
	values := url.Values{"page": {"2"}, "page_size": {"100"}, "state_id": {" lagos "}, "office_type": {" zonal_territorial_operation_office "}, "search": {" Lagos "}}
	got, err = ValidateNEMAZonalTerritorialOperationOfficeListQuery(values)
	if err != nil {
		t.Fatal(err)
	}
	if got.Page != 2 || got.PageSize != 100 || got.StateID != "lagos" || got.OfficeType != "zonal_territorial_operation_office" || got.Search != "Lagos" {
		t.Fatalf("normalized=%#v", got)
	}
	for _, stateID := range []string{"fct", "akwa-ibom", "rivers"} {
		if _, err := ValidateNEMAZonalTerritorialOperationOfficeListQuery(url.Values{"state_id": {stateID}}); err != nil {
			t.Fatalf("state_id=%q err=%v", stateID, err)
		}
	}
	if got, err := ValidateNEMAZonalTerritorialOperationOfficeListQuery(url.Values{"search": {"   "}}); err != nil || got.Search != "" {
		t.Fatalf("empty search should be omitted: got=%#v err=%v", got, err)
	}
	if _, err := ValidateNEMAZonalTerritorialOperationOfficeListQuery(url.Values{"search": {strings.Repeat("a", 100)}}); err != nil {
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
		{"state empty", "state_id=%20"},
		{"state invalid syntax", "state_id=bad_state"},
		{"state unknown", "state_id=unknown"},
		{"office type empty", "office_type=%20"},
		{"office type invalid", "office_type=regional"},
		{"search too long", "search=" + strings.Repeat("a", 101)},
		{"unsupported", "contact_value=112"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tc.query)
			if _, err := ValidateNEMAZonalTerritorialOperationOfficeListQuery(values); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestNEMAZonalTerritorialOperationOfficeIDValidator(t *testing.T) {
	id := strings.Repeat("a", 255)
	if got, err := ValidateNEMAZonalTerritorialOperationOfficeID("office_id", id); err != nil || got != id {
		t.Fatalf("valid 255 id got=%q err=%v", got, err)
	}
	if got, err := ValidateNEMAZonalTerritorialOperationOfficeID("office_id", " nema-lagos-zonal-territorial-operation-office "); err != nil || got != "nema-lagos-zonal-territorial-operation-office" {
		t.Fatalf("trimmed id got=%q err=%v", got, err)
	}
	for _, id := range []string{"", "Bad", "bad_id", "has space", "bad%2Fid", "bad/id", "..", "id?x=1", "id#frag", strings.Repeat("a", 256)} {
		if _, err := ValidateNEMAZonalTerritorialOperationOfficeID("office_id", id); err == nil {
			t.Fatalf("expected error for %q", id)
		}
	}
}
