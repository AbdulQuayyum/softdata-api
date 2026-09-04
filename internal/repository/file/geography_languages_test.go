package file

import (
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

func TestNormalizeCountryLanguageRowsBaseRowWinsAndAliases(t *testing.T) {
	rows := []countryLanguageSourceRow{
		{CountryAreaID: "gb", LanguageID: "en-Latn", Status: "used", BaseLanguage: false},
		{CountryAreaID: "gb", LanguageID: "en", Status: "official", BaseLanguage: true},
		{CountryAreaID: "ng", LanguageID: "fat", Status: "used", BaseLanguage: true},
		{CountryAreaID: "ng", LanguageID: "sh", Status: "used", BaseLanguage: true},
	}
	got := normalizeCountryLanguageRows(rows)
	want := []models.CountryLanguage{
		{CountryAreaID: "gb", LanguageID: "en", Status: "official"},
		{CountryAreaID: "ng", LanguageID: "ak", Status: "used"},
		{CountryAreaID: "ng", LanguageID: "sr", Status: "used"},
	}
	if len(got) != len(want) {
		t.Fatalf("normalizeCountryLanguageRows() returned %d rows, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("row %d = %#v, want %#v", i, got[i], want[i])
		}
	}

	got = normalizeCountryLanguageRows([]countryLanguageSourceRow{
		{CountryAreaID: "gb", LanguageID: "sh", Status: "used", BaseLanguage: false},
		{CountryAreaID: "gb", LanguageID: "sr", Status: "official", BaseLanguage: true},
	})
	if len(got) != 1 || got[0].LanguageID != "sr" || got[0].Status != "official" {
		t.Fatalf("base row did not win collapsed pair: %#v", got)
	}
}
