package models

import (
	"bytes"
	"encoding/json"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

type currencyMetadata struct {
	DatasetKey   string          `json:"dataset_key"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	CountryCode  string          `json:"country_code"`
	DatasetGroup string          `json:"dataset_group"`
	Format       string          `json:"format"`
	RelativePath string          `json:"relative_path"`
	SchemaPath   string          `json:"schema_path"`
	RecordCount  int             `json:"record_count"`
	Version      string          `json:"version"`
	LicenseID    string          `json:"license_id"`
	LicenseURL   string          `json:"license_url"`
	Methodology  string          `json:"methodology"`
	Sources      []datasetSource `json:"sources"`
	VerifiedAt   string          `json:"verified_at"`
}

type currencySchema struct {
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
			Type        string `json:"type"`
			Pattern     string `json:"pattern,omitempty"`
			MinLength   int    `json:"minLength,omitempty"`
			UniqueItems bool   `json:"uniqueItems,omitempty"`
			Enum        []any  `json:"enum,omitempty"`
			Items       *struct {
				Type    string   `json:"type"`
				Enum    []string `json:"enum,omitempty"`
				Pattern string   `json:"pattern,omitempty"`
			} `json:"items,omitempty"`
		} `json:"properties"`
	} `json:"items"`
}

func TestCurrenciesDatasetMatchesApprovedSnapshot(t *testing.T) {
	currencies := loadCurrenciesDataset(t)
	if currencies == nil {
		t.Fatal("decoded dataset is nil")
	}
	if got := len(currencies); got != 155 {
		t.Fatalf("unexpected record count: got %d want 155", got)
	}

	countries := loadCountryOrAreaDataset(t)
	countryIDs := make(map[string]struct{}, len(countries))
	for _, country := range countries {
		countryIDs[country.ID] = struct{}{}
	}

	allowedCodes := map[string]struct{}{
		"BOV": {}, "CHE": {}, "CHW": {}, "CLF": {}, "COU": {}, "MXV": {}, "USN": {}, "UYI": {}, "UYW": {},
		"XAD": {}, "XAG": {}, "XAU": {}, "XBA": {}, "XBB": {}, "XBC": {}, "XBD": {}, "XDR": {}, "XPD": {},
		"XPT": {}, "XSU": {}, "XTS": {}, "XUA": {}, "XXX": {},
	}

	seenIDs := make(map[string]struct{}, len(currencies))
	seenAlphabetic := make(map[string]struct{}, len(currencies))
	seenNumeric := make(map[string]struct{}, len(currencies))
	countryToCodes := make(map[string][]string, len(countryIDs))
	zeroCount := 0
	relationships := 0
	prevName := ""
	prevAlphabetic := ""

	for i, currency := range currencies {
		if currency.ID == "" || currency.Name == "" || currency.AlphabeticCode == "" || currency.NumericCode == "" {
			t.Fatalf("record %d has empty required field: %#v", i, currency)
		}
		if currency.ID != strings.ToLower(currency.AlphabeticCode) {
			t.Fatalf("record %d id mismatch: got %q want %q", i, currency.ID, strings.ToLower(currency.AlphabeticCode))
		}
		if len(currency.CountryAreaIDs) == 0 {
			zeroCount++
		}
		if currency.CountryAreaIDs == nil {
			t.Fatalf("record %d has nil country_area_ids slice", i)
		}
		if currency.Name != strings.TrimSpace(currency.Name) {
			t.Fatalf("record %d name is not trimmed: %q", i, currency.Name)
		}
		if !matchString(`^[a-z]{3}$`, currency.ID) {
			t.Fatalf("record %d has invalid id %q", i, currency.ID)
		}
		if !matchString(`^[A-Z]{3}$`, currency.AlphabeticCode) {
			t.Fatalf("record %d has invalid alphabetic code %q", i, currency.AlphabeticCode)
		}
		if !matchString(`^[0-9]{3}$`, currency.NumericCode) {
			t.Fatalf("record %d has invalid numeric code %q", i, currency.NumericCode)
		}
		if currency.MinorUnit != 0 && currency.MinorUnit != 2 && currency.MinorUnit != 3 {
			t.Fatalf("record %d has unsupported minor unit %d", i, currency.MinorUnit)
		}
		if _, forbidden := allowedCodes[currency.AlphabeticCode]; forbidden {
			t.Fatalf("forbidden code leaked into dataset: %s", currency.AlphabeticCode)
		}
		if _, ok := seenIDs[currency.ID]; ok {
			t.Fatalf("duplicate id found: %q", currency.ID)
		}
		if _, ok := seenAlphabetic[currency.AlphabeticCode]; ok {
			t.Fatalf("duplicate alphabetic code found: %q", currency.AlphabeticCode)
		}
		if _, ok := seenNumeric[currency.NumericCode]; ok {
			t.Fatalf("duplicate numeric code found: %q", currency.NumericCode)
		}
		seenIDs[currency.ID] = struct{}{}
		seenAlphabetic[currency.AlphabeticCode] = struct{}{}
		seenNumeric[currency.NumericCode] = struct{}{}

		if i > 0 {
			if prevName > currency.Name || (prevName == currency.Name && prevAlphabetic > currency.AlphabeticCode) {
				t.Fatalf("records are not sorted by name then alphabetic code at %d: %q before %q", i, prevName, currency.Name)
			}
		}
		prevName = currency.Name
		prevAlphabetic = currency.AlphabeticCode

		if len(currency.CountryAreaIDs) > 0 {
			if !reflect.DeepEqual(currency.CountryAreaIDs, sortedStrings(currency.CountryAreaIDs)) {
				t.Fatalf("record %d country_area_ids are not sorted: %#v", i, currency.CountryAreaIDs)
			}
			seenCountry := make(map[string]struct{}, len(currency.CountryAreaIDs))
			for _, countryID := range currency.CountryAreaIDs {
				if _, ok := countryIDs[countryID]; !ok {
					t.Fatalf("record %d references unknown country_area_id %q", i, countryID)
				}
				if _, ok := seenCountry[countryID]; ok {
					t.Fatalf("record %d contains duplicate country_area_id %q", i, countryID)
				}
				seenCountry[countryID] = struct{}{}
				countryToCodes[countryID] = append(countryToCodes[countryID], currency.AlphabeticCode)
			}
		}

		relationships += len(currency.CountryAreaIDs)

		serialized, err := json.Marshal(currency)
		if err != nil {
			t.Fatalf("marshal currency %q: %v", currency.AlphabeticCode, err)
		}
		var public map[string]any
		if err := json.Unmarshal(serialized, &public); err != nil {
			t.Fatalf("unmarshal public currency %q: %v", currency.AlphabeticCode, err)
		}
		if len(public) != 6 {
			t.Fatalf("unexpected public field count for %q: %#v", currency.AlphabeticCode, public)
		}
		for _, field := range []string{"id", "name", "alphabetic_code", "numeric_code", "minor_unit", "country_area_ids"} {
			if _, ok := public[field]; !ok {
				t.Fatalf("public field %q missing for %q", field, currency.AlphabeticCode)
			}
		}
	}

	if zeroCount != 1 {
		t.Fatalf("unexpected zero-mapping count: got %d want 1", zeroCount)
	}
	if relationships != 252 {
		t.Fatalf("unexpected relationship count: got %d want 252", relationships)
	}
	if len(countryToCodes) != 245 {
		t.Fatalf("unexpected unique mapped country/area count: got %d want 245", len(countryToCodes))
	}

	for _, code := range []string{"BOV", "CHE", "CHW", "CLF", "COU", "MXV", "USN", "UYI", "UYW", "XAD", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XDR", "XPD", "XPT", "XSU", "XTS", "XUA", "XXX"} {
		if containsCurrencyCode(currencies, code) {
			t.Fatalf("excluded code unexpectedly present: %s", code)
		}
	}

	for _, code := range []string{"XAF", "XCD", "XCG", "XOF", "XPF"} {
		if !containsCurrencyCode(currencies, code) {
			t.Fatalf("expected operational currency missing: %s", code)
		}
	}

	for _, code := range []string{"EUR", "USD", "GBP", "XAF", "XOF", "XCD", "XPF", "XCG", "DKK", "NZD", "AUD"} {
		if got := len(currencyCountryIDs(currencies, code)); got != map[string]int{"EUR": 36, "USD": 19, "GBP": 4, "XAF": 6, "XOF": 8, "XCD": 8, "XPF": 3, "XCG": 2, "DKK": 3, "NZD": 5, "AUD": 8}[code] {
			t.Fatalf("unexpected mapped country count for %s", code)
		}
	}

	if got := currencyCountryIDs(currencies, "TWD"); !reflect.DeepEqual(got, []string{}) {
		t.Fatalf("unexpected TWD country_area_ids: %#v", got)
	}

	requireCurrencyCountries(t, countryToCodes, "bt", []string{"BTN", "INR"})
	requireCurrencyCountries(t, countryToCodes, "sv", []string{"SVC", "USD"})
	requireCurrencyCountries(t, countryToCodes, "ht", []string{"HTG", "USD"})
	requireCurrencyCountries(t, countryToCodes, "ls", []string{"LSL", "ZAR"})
	requireCurrencyCountries(t, countryToCodes, "na", []string{"NAD", "ZAR"})
	requireCurrencyCountries(t, countryToCodes, "pa", []string{"PAB", "USD"})
	requireCurrencyCountries(t, countryToCodes, "ve", []string{"VED", "VES"})

	for _, countryID := range []string{"aq", "ps", "gs"} {
		if got := countryToCodes[countryID]; len(got) != 0 {
			t.Fatalf("unexpected reverse currency mapping for %s: %#v", countryID, got)
		}
	}

	requireCurrencyCountries(t, countryToCodes, "za", []string{"ZAR"})

}

func TestCurrenciesMetadataSchemaAndAttribution(t *testing.T) {
	metadata := loadCurrencyMetadata(t)
	if metadata.DatasetKey != "world-currencies" {
		t.Fatalf("unexpected dataset key: %q", metadata.DatasetKey)
	}
	if metadata.Title != "World Currencies" {
		t.Fatalf("unexpected title: %q", metadata.Title)
	}
	if metadata.DatasetGroup != "finance" || metadata.Format != "json" || metadata.CountryCode != "001" {
		t.Fatalf("unexpected metadata classification: %#v", metadata)
	}
	if metadata.RelativePath != "finance/currencies.json" {
		t.Fatalf("unexpected relative path: %q", metadata.RelativePath)
	}
	if metadata.SchemaPath != "schemas/finance/currencies.schema.json" {
		t.Fatalf("unexpected schema path: %q", metadata.SchemaPath)
	}
	if metadata.RecordCount != 155 {
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
	if len(metadata.Sources) != 5 {
		t.Fatalf("unexpected source count: %d", len(metadata.Sources))
	}
	for _, source := range metadata.Sources {
		if source.Organization == "" || source.Title == "" || source.URL == "" || source.Purpose == "" || source.AccessedAt == "" {
			t.Fatalf("incomplete provenance source: %#v", source)
		}
	}
	if metadata.Sources[0].URL != "https://www.six-group.com/dam/download/financial-information/data-center/iso-currrency/lists/list-one.xml" || metadata.Sources[1].URL != metadata.Sources[0].URL || metadata.Sources[2].URL != metadata.Sources[0].URL {
		t.Fatalf("unexpected SIX XML sources: %#v", metadata.Sources[:3])
	}
	if metadata.Sources[0].AccessedAt != "2026-08-30T22:54:39Z" || metadata.Sources[1].AccessedAt != "2026-08-30T22:54:40Z" || metadata.Sources[2].AccessedAt != "2026-08-30T22:54:42Z" {
		t.Fatalf("unexpected XML retrieval timestamps: %#v", metadata.Sources[:3])
	}
	for _, snippet := range []string{
		"280 rows and 178 unique currency codes",
		"155 current monetary currencies remained",
		"TWD intentionally has an empty country_area_ids array",
		"Antarctica, State of Palestine and South Georgia and the South Sandwich Islands carry no reverse currency relationship",
		"Kosovo and the European Union are not added as SoftData country/area IDs",
		"SoftData's independent compilation, schema and metadata are licensed under CC BY 4.0",
		"source SHA-256",
		"deterministic name order",
	} {
		if !strings.Contains(metadata.Methodology, snippet) {
			t.Fatalf("metadata methodology missing %q: %q", snippet, metadata.Methodology)
		}
	}
	for _, forbidden := range []string{"ISO and SIX source publications are CC BY 4.0", "SIX source publications are CC BY 4.0"} {
		if strings.Contains(metadata.Methodology, forbidden) {
			t.Fatalf("metadata methodology incorrectly re-licenses source material: %q", forbidden)
		}
	}

	schema := loadCurrencySchema(t)
	if schema.Schema != "https://json-schema.org/draft/2020-12/schema" {
		t.Fatalf("unexpected schema dialect: %q", schema.Schema)
	}
	if schema.ID != "https://softdata-api.local/schemas/finance/currencies.schema.json" {
		t.Fatalf("unexpected schema id: %q", schema.ID)
	}
	if schema.Title != "World Currencies" {
		t.Fatalf("unexpected schema title: %q", schema.Title)
	}
	if schema.Type != "array" || schema.MinItems != 155 || schema.MaxItems != 155 || !schema.UniqueItems {
		t.Fatalf("unexpected schema constraints: %#v", schema)
	}
	if schema.Items.Type != "object" || schema.Items.AdditionalProperties {
		t.Fatalf("unexpected item constraints: %#v", schema.Items)
	}
	if !reflect.DeepEqual(sortedStrings(schema.Items.Required), []string{"alphabetic_code", "country_area_ids", "id", "minor_unit", "name", "numeric_code"}) {
		t.Fatalf("unexpected required fields: %#v", schema.Items.Required)
	}

	for _, field := range []string{"id", "name", "alphabetic_code", "numeric_code", "minor_unit", "country_area_ids"} {
		if _, ok := schema.Items.Properties[field]; !ok {
			t.Fatalf("schema missing property %q", field)
		}
	}
	if prop := schema.Items.Properties["id"]; prop.Pattern != "^[a-z]{3}$" {
		t.Fatalf("unexpected id pattern: %#v", prop)
	}
	if prop := schema.Items.Properties["alphabetic_code"]; prop.Pattern != "^[A-Z]{3}$" {
		t.Fatalf("unexpected alphabetic_code pattern: %#v", prop)
	}
	if prop := schema.Items.Properties["numeric_code"]; prop.Pattern != "^[0-9]{3}$" {
		t.Fatalf("unexpected numeric_code pattern: %#v", prop)
	}
	if prop := schema.Items.Properties["name"]; prop.MinLength != 1 {
		t.Fatalf("unexpected name constraints: %#v", prop)
	}
	if prop := schema.Items.Properties["minor_unit"]; prop.Type != "integer" || !reflect.DeepEqual(intEnum(prop.Enum), []int{0, 2, 3}) {
		t.Fatalf("unexpected minor unit constraints: %#v", prop)
	}
	if prop := schema.Items.Properties["country_area_ids"]; prop.Type != "array" || !prop.UniqueItems || prop.Items == nil || prop.Items.Type != "string" || len(prop.Items.Enum) != 248 {
		t.Fatalf("unexpected country_area_ids constraints: %#v", prop)
	} else {
		countryIDs := make([]string, 0, len(prop.Items.Enum))
		for _, id := range prop.Items.Enum {
			countryIDs = append(countryIDs, id)
		}
		sort.Strings(countryIDs)
		if !reflect.DeepEqual(countryIDs, sortedCountryIDs(loadCountryOrAreaDataset(t))) {
			t.Fatalf("country_area_ids enum does not match approved country ids")
		}
	}
}

func loadCurrenciesDataset(t *testing.T) []Currency {
	t.Helper()
	var currencies []Currency
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("finance/currencies.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&currencies); err != nil {
		t.Fatalf("decode currencies dataset: %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("currencies dataset contains trailing JSON")
	}
	return currencies
}

func loadCurrencyMetadata(t *testing.T) currencyMetadata {
	t.Helper()
	var metadata currencyMetadata
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("metadata/finance/currencies.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&metadata); err != nil {
		t.Fatalf("decode currency metadata: %v", err)
	}
	return metadata
}

func loadCurrencySchema(t *testing.T) currencySchema {
	t.Helper()
	var schema currencySchema
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("schemas/finance/currencies.schema.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&schema); err != nil {
		t.Fatalf("decode currency schema: %v", err)
	}
	return schema
}

func currencyCountryIDs(currencies []Currency, code string) []string {
	for _, currency := range currencies {
		if currency.AlphabeticCode == code {
			return append([]string{}, currency.CountryAreaIDs...)
		}
	}
	return nil
}

func containsCurrencyCode(currencies []Currency, code string) bool {
	for _, currency := range currencies {
		if currency.AlphabeticCode == code {
			return true
		}
	}
	return false
}

func requireCurrencyCountries(t *testing.T, countryToCodes map[string][]string, countryID string, want []string) {
	t.Helper()
	got := append([]string(nil), countryToCodes[countryID]...)
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected currency mapping for %s: got %#v want %#v", countryID, got, want)
	}
}

func containsCountryID(values []string) bool {
	return len(values) > 0
}

func sortedCountryIDs(countries []CountryOrArea) []string {
	ids := make([]string, 0, len(countries))
	for _, country := range countries {
		ids = append(ids, country.ID)
	}
	sort.Strings(ids)
	return ids
}

func intEnum(values []any) []int {
	result := make([]int, 0, len(values))
	for _, value := range values {
		if num, ok := value.(float64); ok {
			result = append(result, int(num))
		}
	}
	sort.Ints(result)
	return result
}

func matchString(pattern, value string) bool {
	matched, err := regexp.MatchString(pattern, value)
	return err == nil && matched
}
