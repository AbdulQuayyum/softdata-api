package models

import (
	"encoding/json"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

type countryAreaMetadata struct {
	DatasetKey   string              `json:"dataset_key"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	CountryCode  string              `json:"country_code"`
	DatasetGroup string              `json:"dataset_group"`
	Format       string              `json:"format"`
	RelativePath string              `json:"relative_path"`
	SchemaPath   string              `json:"schema_path"`
	RecordCount  int                 `json:"record_count"`
	Version      string              `json:"version"`
	LicenseID    string              `json:"license_id"`
	LicenseURL   string              `json:"license_url"`
	Methodology  string              `json:"methodology"`
	Sources      []countryAreaSource `json:"sources"`
	VerifiedAt   string              `json:"verified_at"`
}

type countryAreaSource struct {
	Organization string `json:"organization"`
	Title        string `json:"title"`
	URL          string `json:"url"`
	Purpose      string `json:"purpose"`
	AccessedAt   string `json:"accessed_at"`
}

type countryAreaSchema struct {
	Schema      string `json:"$schema"`
	ID          string `json:"$id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	MinItems    int    `json:"minItems"`
	MaxItems    int    `json:"maxItems"`
	UniqueItems bool   `json:"uniqueItems"`
	Items       struct {
		Type                 string   `json:"type"`
		AdditionalProperties bool     `json:"additionalProperties"`
		Required             []string `json:"required"`
		Properties           map[string]struct {
			Type      string   `json:"type"`
			Pattern   string   `json:"pattern,omitempty"`
			MinLength int      `json:"minLength,omitempty"`
			Enum      []string `json:"enum,omitempty"`
			Const     string   `json:"const,omitempty"`
		} `json:"properties"`
		AllOf []struct {
			If struct {
				Required []string `json:"required"`
			} `json:"if"`
			Then struct {
				Required []string `json:"required"`
			} `json:"then"`
		} `json:"allOf"`
	} `json:"items"`
}

type schemaConditionalRule struct {
	If struct {
		Required []string `json:"required"`
	} `json:"if"`
	Then struct {
		Required []string `json:"required"`
	} `json:"then"`
}

func TestWorldCountriesAndAreasDatasetMatchesApprovedManifest(t *testing.T) {
	rows := loadCountryOrAreaDataset(t)

	if rows == nil {
		t.Fatal("decoded dataset is nil")
	}
	if got := len(rows); got != 248 {
		t.Fatalf("unexpected record count: got %d want 248", got)
	}

	allowedFields := map[string]struct{}{
		"id":                       {},
		"name":                     {},
		"alpha_2_code":             {},
		"alpha_3_code":             {},
		"numeric_code":             {},
		"region_code":              {},
		"region_name":              {},
		"subregion_code":           {},
		"subregion_name":           {},
		"intermediate_region_code": {},
		"intermediate_region_name": {},
	}
	allowedNames := map[string]struct{}{}
	for field := range allowedFields {
		allowedNames[field] = struct{}{}
	}

	seenIDs := make(map[string]struct{}, len(rows))
	seenAlpha2 := make(map[string]struct{}, len(rows))
	seenAlpha3 := make(map[string]struct{}, len(rows))
	seenNumeric := make(map[string]struct{}, len(rows))
	seenNames := make(map[string]struct{}, len(rows))
	withoutRegion := 0

	for i, row := range rows {
		if row.ID == "" || row.Name == "" || row.Alpha2Code == "" || row.Alpha3Code == "" || row.NumericCode == "" {
			t.Fatalf("record %d has empty required field: %#v", i, row)
		}
		if row.ID != strings.ToLower(row.Alpha2Code) {
			t.Fatalf("record %d id mismatch: got %q want %q", i, row.ID, strings.ToLower(row.Alpha2Code))
		}
		if !regexp.MustCompile(`^[a-z]{2}$`).MatchString(row.ID) {
			t.Fatalf("record %d has invalid id %q", i, row.ID)
		}
		if !regexp.MustCompile(`^[A-Z]{2}$`).MatchString(row.Alpha2Code) {
			t.Fatalf("record %d has invalid alpha-2 code %q", i, row.Alpha2Code)
		}
		if !regexp.MustCompile(`^[A-Z]{3}$`).MatchString(row.Alpha3Code) {
			t.Fatalf("record %d has invalid alpha-3 code %q", i, row.Alpha3Code)
		}
		if !regexp.MustCompile(`^[0-9]{3}$`).MatchString(row.NumericCode) {
			t.Fatalf("record %d has invalid numeric code %q", i, row.NumericCode)
		}
		if _, ok := seenIDs[row.ID]; ok {
			t.Fatalf("duplicate id found: %q", row.ID)
		}
		if _, ok := seenAlpha2[row.Alpha2Code]; ok {
			t.Fatalf("duplicate alpha-2 code found: %q", row.Alpha2Code)
		}
		if _, ok := seenAlpha3[row.Alpha3Code]; ok {
			t.Fatalf("duplicate alpha-3 code found: %q", row.Alpha3Code)
		}
		if _, ok := seenNumeric[row.NumericCode]; ok {
			t.Fatalf("duplicate numeric code found: %q", row.NumericCode)
		}
		if _, ok := seenNames[row.Name]; ok {
			t.Fatalf("duplicate name found: %q", row.Name)
		}
		seenIDs[row.ID] = struct{}{}
		seenAlpha2[row.Alpha2Code] = struct{}{}
		seenAlpha3[row.Alpha3Code] = struct{}{}
		seenNumeric[row.NumericCode] = struct{}{}
		seenNames[row.Name] = struct{}{}

		if i > 0 {
			prev := rows[i-1]
			if prev.Name > row.Name || (prev.Name == row.Name && prev.Alpha2Code > row.Alpha2Code) {
				t.Fatalf("records are not sorted by name then alpha-2: %q before %q", prev.Name, row.Name)
			}
		}

		if row.RegionCode == "" || row.RegionName == "" {
			withoutRegion++
		}
		if (row.RegionCode == "") != (row.RegionName == "") {
			t.Fatalf("record %d has inconsistent region pair: %#v", i, row)
		}
		if (row.SubregionCode == "") != (row.SubregionName == "") {
			t.Fatalf("record %d has inconsistent subregion pair: %#v", i, row)
		}
		if (row.IntermediateRegionCode == "") != (row.IntermediateRegionName == "") {
			t.Fatalf("record %d has inconsistent intermediate-region pair: %#v", i, row)
		}
		if row.RegionCode != "" && row.SubregionCode == "" {
			t.Fatalf("record %d has region without subregion: %#v", i, row)
		}
		if row.IntermediateRegionCode != "" && row.SubregionCode == "" {
			t.Fatalf("record %d has intermediate region without subregion: %#v", i, row)
		}

		marshaled, err := json.Marshal(row)
		if err != nil {
			t.Fatalf("marshal record %d: %v", i, err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(marshaled, &raw); err != nil {
			t.Fatalf("unmarshal marshaled record %d: %v", i, err)
		}
		for key := range raw {
			if _, ok := allowedNames[key]; !ok {
				t.Fatalf("record %d contains unexpected field %q", i, key)
			}
		}
	}

	if len(seenIDs) != 248 || len(seenAlpha2) != 248 || len(seenAlpha3) != 248 || len(seenNumeric) != 248 {
		t.Fatalf("unexpected uniqueness totals: ids=%d alpha2=%d alpha3=%d numeric=%d", len(seenIDs), len(seenAlpha2), len(seenAlpha3), len(seenNumeric))
	}
	if withoutRegion != 1 {
		t.Fatalf("unexpected number of records without region hierarchy: got %d want 1", withoutRegion)
	}

	checkCountryOrAreaRecord(t, rows, "NG", "Nigeria", "NGA", "566")
	checkCountryOrAreaRecord(t, rows, "DZ", "Algeria", "DZA", "012")
	checkCountryOrAreaRecord(t, rows, "VA", "Holy See", "VAT", "336")
	checkCountryOrAreaRecord(t, rows, "PS", "State of Palestine", "PSE", "275")
	checkCountryOrAreaRecord(t, rows, "EH", "Western Sahara", "ESH", "732")
	checkCountryOrAreaRecord(t, rows, "HK", "China, Hong Kong Special Administrative Region", "HKG", "344")
	checkCountryOrAreaRecord(t, rows, "MO", "China, Macao Special Administrative Region", "MAC", "446")
	checkCountryOrAreaRecord(t, rows, "AQ", "Antarctica", "ATA", "010")

	if hasCountryOrAreaRecord(rows, "XK") {
		t.Fatal("kosovo should be absent from the approved current manifest")
	}
}

func TestWorldCountriesAndAreasMetadataSchemaAndNotice(t *testing.T) {
	metadata := loadCountryOrAreaMetadata(t)
	schema := loadCountryOrAreaSchema(t)

	if metadata.DatasetKey != "world-countries-and-areas" {
		t.Fatalf("unexpected dataset key: %q", metadata.DatasetKey)
	}
	if metadata.Title != "World Countries and Areas" {
		t.Fatalf("unexpected title: %q", metadata.Title)
	}
	if metadata.CountryCode != "001" || metadata.DatasetGroup != "geography" || metadata.Format != "json" {
		t.Fatalf("unexpected metadata classification: %#v", metadata)
	}
	if metadata.RelativePath != "geography/countries_and_areas.json" {
		t.Fatalf("unexpected relative path: %q", metadata.RelativePath)
	}
	if metadata.SchemaPath != "schemas/geography/countries_and_areas.schema.json" {
		t.Fatalf("unexpected schema path: %q", metadata.SchemaPath)
	}
	if metadata.RecordCount != 248 {
		t.Fatalf("unexpected record count: %d", metadata.RecordCount)
	}
	if metadata.Version != "1.0.0" {
		t.Fatalf("unexpected version: %q", metadata.Version)
	}
	if metadata.LicenseID != "CC-BY-4.0" || metadata.LicenseURL != "https://creativecommons.org/licenses/by/4.0/" {
		t.Fatalf("unexpected license metadata: %#v", metadata)
	}
	if metadata.VerifiedAt != "2026-08-30" {
		t.Fatalf("unexpected verified_at: %q", metadata.VerifiedAt)
	}
	verifiedAt, err := time.Parse("2006-01-02", metadata.VerifiedAt)
	if err != nil {
		t.Fatalf("verified_at is not a valid date: %v", err)
	}
	if verifiedAt.After(time.Now().UTC()) {
		t.Fatalf("verified_at is in the future: %s", metadata.VerifiedAt)
	}
	if len(metadata.Sources) != 3 {
		t.Fatalf("unexpected source count: %d", len(metadata.Sources))
	}

	wantTimestamps := []string{
		"2026-08-30T22:54:39Z",
		"2026-08-30T22:54:40Z",
		"2026-08-30T22:54:42Z",
	}
	for i, source := range metadata.Sources {
		if source.Organization != "United Nations Statistics Division" || source.Title != "Standard country or area codes for statistical use (M49)" || source.URL != "https://unstats.un.org/unsd/methodology/m49/overview/" {
			t.Fatalf("unexpected provenance source %d: %#v", i, source)
		}
		if source.Purpose == "" || source.AccessedAt == "" {
			t.Fatalf("incomplete provenance source %d: %#v", i, source)
		}
		if source.AccessedAt != wantTimestamps[i] {
			t.Fatalf("unexpected accessed_at for source %d: got %q want %q", i, source.AccessedAt, wantTimestamps[i])
		}
	}

	requiredSnippets := []string{
		"248 rows",
		"retrieved three times",
		"deterministic name order",
		"statistical reference only",
		"political recognition",
		"ISO-derived alpha codes are included as published",
		"UN source material retains its own rights",
		"licensed under CC BY 4.0",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(metadata.Methodology, snippet) {
			t.Fatalf("missing metadata methodology snippet: %q", snippet)
		}
	}

	if schema.Schema == "" || schema.ID == "" || schema.Title == "" || schema.Description == "" {
		t.Fatalf("schema missing required top-level fields: %#v", schema)
	}
	if schema.Type != "array" || schema.MinItems != 248 || schema.MaxItems != 248 || !schema.UniqueItems {
		t.Fatalf("unexpected schema constraints: %#v", schema)
	}
	if schema.Items.Type != "object" {
		t.Fatalf("unexpected item type: %q", schema.Items.Type)
	}
	if schema.Items.AdditionalProperties {
		t.Fatal("schema should forbid additional properties")
	}

	wantRequired := []string{"id", "name", "alpha_2_code", "alpha_3_code", "numeric_code"}
	sort.Strings(wantRequired)
	if !reflect.DeepEqual(sortedStrings(schema.Items.Required), wantRequired) {
		t.Fatalf("unexpected required fields: got %v want %v", sortedStrings(schema.Items.Required), wantRequired)
	}

	for _, field := range wantRequired {
		if _, ok := schema.Items.Properties[field]; !ok {
			t.Fatalf("schema missing required property %q", field)
		}
	}

	for _, field := range []string{"region_code", "region_name", "subregion_code", "subregion_name", "intermediate_region_code", "intermediate_region_name"} {
		prop, ok := schema.Items.Properties[field]
		if !ok {
			t.Fatalf("schema missing optional property %q", field)
		}
		if prop.Type != "string" {
			t.Fatalf("schema optional property %q has unexpected type %q", field, prop.Type)
		}
	}
	if len(schema.Items.AllOf) != 6 {
		t.Fatalf("unexpected conditional count: %d", len(schema.Items.AllOf))
	}
	assertPairRule(t, schema.Items.AllOf[0], "region_code", "region_name")
	assertPairRule(t, schema.Items.AllOf[1], "region_name", "region_code")
	assertPairRule(t, schema.Items.AllOf[2], "subregion_code", "subregion_name")
	assertPairRule(t, schema.Items.AllOf[3], "subregion_name", "subregion_code")
	assertPairRule(t, schema.Items.AllOf[4], "intermediate_region_code", "intermediate_region_name")
	assertPairRule(t, schema.Items.AllOf[5], "intermediate_region_name", "intermediate_region_code")
}

func loadCountryOrAreaDataset(t *testing.T) []CountryOrArea {
	t.Helper()

	data, err := os.ReadFile(datasetPath("geography", "countries_and_areas.json"))
	if err != nil {
		t.Fatalf("read dataset: %v", err)
	}
	var rows []CountryOrArea
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatalf("unmarshal dataset: %v", err)
	}
	return rows
}

func loadCountryOrAreaMetadata(t *testing.T) countryAreaMetadata {
	t.Helper()

	var metadata countryAreaMetadata
	if err := json.Unmarshal([]byte(readTextFile(t, datasetPath("metadata", "geography", "countries_and_areas.json"))), &metadata); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	return metadata
}

func loadCountryOrAreaSchema(t *testing.T) countryAreaSchema {
	t.Helper()

	var schema countryAreaSchema
	if err := json.Unmarshal([]byte(readTextFile(t, datasetPath("schemas", "geography", "countries_and_areas.schema.json"))), &schema); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	return schema
}

func checkCountryOrAreaRecord(t *testing.T, rows []CountryOrArea, alpha2, name, alpha3, numeric string) {
	t.Helper()

	for _, row := range rows {
		if row.Alpha2Code == alpha2 {
			if row.ID != strings.ToLower(alpha2) || row.Name != name || row.Alpha3Code != alpha3 || row.NumericCode != numeric {
				t.Fatalf("unexpected record for %s: %#v", alpha2, row)
			}
			return
		}
	}
	t.Fatalf("missing expected record for alpha-2 code %q", alpha2)
}

func hasCountryOrAreaRecord(rows []CountryOrArea, alpha2 string) bool {
	for _, row := range rows {
		if row.Alpha2Code == alpha2 {
			return true
		}
	}
	return false
}

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func assertPairRule(t *testing.T, rule schemaConditionalRule, wantIf, wantThen string) {
	t.Helper()

	if len(rule.If.Required) != 1 || rule.If.Required[0] != wantIf {
		t.Fatalf("unexpected if-required pair: %#v", rule)
	}
	if len(rule.Then.Required) != 1 || rule.Then.Required[0] != wantThen {
		t.Fatalf("unexpected then-required pair: %#v", rule)
	}
}
