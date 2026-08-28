package models

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

type datasetMetadata struct {
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

type datasetSource struct {
	Organization string `json:"organization"`
	Title        string `json:"title"`
	URL          string `json:"url"`
	Purpose      string `json:"purpose"`
	AccessedAt   string `json:"accessed_at"`
}

type stateSchema struct {
	Schema      string          `json:"$schema"`
	ID          string          `json:"$id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Type        string          `json:"type"`
	MinItems    int             `json:"minItems"`
	MaxItems    int             `json:"maxItems"`
	UniqueItems bool            `json:"uniqueItems"`
	Items       json.RawMessage `json:"items"`
}

func TestNigeriaStatesDatasetMatchesApprovedManifest(t *testing.T) {
	states := loadStateDataset(t)

	if got := len(states); got != 37 {
		t.Fatalf("unexpected record count: got %d want 37", got)
	}

	if !reflect.DeepEqual(states, approvedStates()) {
		t.Fatalf("dataset records do not match approved manifest:\n got: %#v\nwant: %#v", states, approvedStates())
	}

	seenIDs := make(map[string]struct{}, len(states))
	seenNames := make(map[string]struct{}, len(states))
	zoneCounts := map[string]int{
		"north-central": 0,
		"north-east":    0,
		"north-west":    0,
		"south-east":    0,
		"south-south":   0,
		"south-west":    0,
	}
	stateCount := 0
	fctCount := 0

	for i, state := range states {
		if state.ID == "" || state.Name == "" || state.OfficialName == "" || state.AdministrativeType == "" || state.Capital == "" || state.GeopoliticalZoneID == "" || state.CountryCode == "" {
			t.Fatalf("record %d has empty required field: %#v", i, state)
		}
		if !regexp.MustCompile(`^[a-z]+(?:-[a-z]+)*$`).MatchString(state.ID) {
			t.Fatalf("record %d has invalid id %q", i, state.ID)
		}
		if state.CountryCode != "NG" {
			t.Fatalf("record %d has unexpected country code %q", i, state.CountryCode)
		}
		if _, ok := seenIDs[state.ID]; ok {
			t.Fatalf("duplicate id found: %q", state.ID)
		}
		seenIDs[state.ID] = struct{}{}
		if _, ok := seenNames[state.Name]; ok {
			t.Fatalf("duplicate name found: %q", state.Name)
		}
		seenNames[state.Name] = struct{}{}
		if i > 0 && strings.Compare(states[i-1].Name, state.Name) > 0 {
			t.Fatalf("records are not sorted by name: %q before %q", states[i-1].Name, state.Name)
		}
		switch state.AdministrativeType {
		case "state":
			stateCount++
		case "federal_capital_territory":
			fctCount++
		default:
			t.Fatalf("record %d has invalid administrative type %q", i, state.AdministrativeType)
		}
		if _, ok := zoneCounts[state.GeopoliticalZoneID]; !ok {
			t.Fatalf("record %d has invalid zone %q", i, state.GeopoliticalZoneID)
		}
		zoneCounts[state.GeopoliticalZoneID]++
	}

	if stateCount != 36 {
		t.Fatalf("unexpected state count: got %d want 36", stateCount)
	}
	if fctCount != 1 {
		t.Fatalf("unexpected FCT count: got %d want 1", fctCount)
	}

	fct := states[14]
	if fct.ID != "fct" || fct.Capital != "Abuja" || fct.GeopoliticalZoneID != "north-central" || fct.AdministrativeType != "federal_capital_territory" {
		t.Fatalf("unexpected FCT record: %#v", fct)
	}

	wantZones := map[string]int{
		"north-central": 7,
		"north-east":    6,
		"north-west":    7,
		"south-east":    5,
		"south-south":   6,
		"south-west":    6,
	}
	if !reflect.DeepEqual(zoneCounts, wantZones) {
		t.Fatalf("unexpected zone counts: got %#v want %#v", zoneCounts, wantZones)
	}
}

func TestNigeriaStatesMetadataSchemaAndNotices(t *testing.T) {
	metadata := loadDatasetMetadata(t)
	if metadata.DatasetKey != "ng-states" {
		t.Fatalf("unexpected dataset key: %q", metadata.DatasetKey)
	}
	if metadata.Title != "Nigeria States and Federal Capital Territory" {
		t.Fatalf("unexpected title: %q", metadata.Title)
	}
	if metadata.RecordCount != 37 {
		t.Fatalf("unexpected record count: %d", metadata.RecordCount)
	}
	if metadata.Version != "1.0.0" {
		t.Fatalf("unexpected version: %q", metadata.Version)
	}
	if metadata.LicenseID != "CC-BY-4.0" {
		t.Fatalf("unexpected license id: %q", metadata.LicenseID)
	}
	if metadata.LicenseURL != "https://creativecommons.org/licenses/by/4.0/" {
		t.Fatalf("unexpected license url: %q", metadata.LicenseURL)
	}
	if metadata.RelativePath != "geography/states.json" {
		t.Fatalf("unexpected relative path: %q", metadata.RelativePath)
	}
	if metadata.SchemaPath != "schemas/geography/states.schema.json" {
		t.Fatalf("unexpected schema path: %q", metadata.SchemaPath)
	}
	if metadata.VerifiedAt != "2026-08-28" {
		t.Fatalf("unexpected verified_at: %q", metadata.VerifiedAt)
	}

	if _, err := os.Stat(datasetPath(metadata.RelativePath)); err != nil {
		t.Fatalf("data file missing: %v", err)
	}
	if _, err := os.Stat(datasetPath(metadata.SchemaPath)); err != nil {
		t.Fatalf("schema file missing: %v", err)
	}

	if !strings.Contains(metadata.Methodology, "independently compiled") {
		t.Fatalf("unexpected methodology: %q", metadata.Methodology)
	}
	for _, source := range metadata.Sources {
		if source.Organization == "" || source.Title == "" || source.URL == "" || source.Purpose == "" || source.AccessedAt == "" {
			t.Fatalf("incomplete provenance source: %#v", source)
		}
		if !strings.HasPrefix(source.URL, "https://") {
			t.Fatalf("non-https provenance url: %q", source.URL)
		}
	}

	schema := loadDatasetSchema(t)
	if schema.Schema == "" || schema.ID == "" || schema.Title == "" || schema.Description == "" {
		t.Fatalf("schema missing required fields: %#v", schema)
	}
	if schema.Type != "array" || schema.MinItems != 37 || schema.MaxItems != 37 || !schema.UniqueItems {
		t.Fatalf("unexpected schema constraints: %#v", schema)
	}

	licenseText := readTextFile(t, datasetPath("LICENSE.md"))
	readmeText := readTextFile(t, datasetPath("README.md"))
	requiredSnippets := []string{
		"SoftData's independently compiled dataset package is available under CC BY 4.0.",
		"SoftData API contributors, “Nigeria States and Federal Capital Territory”, version 1.0.0.",
		"https://creativecommons.org/licenses/by/4.0/",
		"original source organizations retain any rights",
		"does not imply endorsement",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(licenseText+readmeText, snippet) {
			t.Fatalf("missing required notice snippet: %q", snippet)
		}
	}

	prohibitedClaims := []string{
		"INEC licensed under CC BY 4.0",
		"NBS licensed under CC BY 4.0",
		"NYSC licensed under CC BY 4.0",
		"NPC licensed under CC BY 4.0",
		"source agencies licensed their publications under CC BY 4.0",
	}
	for _, claim := range prohibitedClaims {
		if strings.Contains(licenseText, claim) || strings.Contains(readmeText, claim) {
			t.Fatalf("found prohibited licensing claim: %q", claim)
		}
	}
}

func TestNigeriaGeopoliticalZonesDatasetAndCrossReferences(t *testing.T) {
	zones := loadGeopoliticalZoneDataset(t)
	metadata := loadGeopoliticalZoneMetadata(t)
	schema := loadGeopoliticalZoneSchema(t)
	states := loadStateDataset(t)

	if got := len(zones); got != 6 {
		t.Fatalf("unexpected zone record count: got %d want 6", got)
	}

	if !reflect.DeepEqual(zones, approvedGeopoliticalZones()) {
		t.Fatalf("dataset records do not match approved manifest:\n got: %#v\nwant: %#v", zones, approvedGeopoliticalZones())
	}

	if metadata.DatasetKey != "ng-geopolitical-zones" {
		t.Fatalf("unexpected dataset key: %q", metadata.DatasetKey)
	}
	if metadata.Title != "Nigeria Geopolitical Zones" {
		t.Fatalf("unexpected title: %q", metadata.Title)
	}
	if metadata.RecordCount != 6 {
		t.Fatalf("unexpected record count: %d", metadata.RecordCount)
	}
	if metadata.Version != "1.0.0" {
		t.Fatalf("unexpected version: %q", metadata.Version)
	}
	if metadata.RelativePath != "geography/geopolitical_zones.json" {
		t.Fatalf("unexpected relative path: %q", metadata.RelativePath)
	}
	if metadata.SchemaPath != "schemas/geography/geopolitical_zones.schema.json" {
		t.Fatalf("unexpected schema path: %q", metadata.SchemaPath)
	}
	if metadata.VerifiedAt != "2026-08-28" {
		t.Fatalf("unexpected verified_at: %q", metadata.VerifiedAt)
	}
	if len(metadata.Sources) != 4 {
		t.Fatalf("unexpected source count: %d", len(metadata.Sources))
	}
	if !strings.Contains(metadata.Methodology, "National Bureau of Statistics") || !strings.Contains(metadata.Methodology, "UBEC") || !strings.Contains(metadata.Methodology, "Niger State Government") || !strings.Contains(metadata.Methodology, "RMAFC") {
		t.Fatalf("methodology does not document the conflict sources: %q", metadata.Methodology)
	}
	for _, want := range []string{
		"https://microdata.nigerianstat.gov.ng/index.php/catalog/55/study-description",
		"https://ubec.gov.ng/zonal-and-state-offices/",
		"https://nigerstate.gov.ng/about/",
		"https://rmafc.gov.ng/structure/",
	} {
		if !strings.Contains(metadata.Methodology, want) {
			t.Fatalf("methodology missing url %q: %q", want, metadata.Methodology)
		}
	}
	for _, source := range metadata.Sources {
		if source.Organization == "" || source.Title == "" || source.URL == "" || source.Purpose == "" || source.AccessedAt == "" {
			t.Fatalf("incomplete provenance source: %#v", source)
		}
		if !strings.HasPrefix(source.URL, "https://") {
			t.Fatalf("non-https provenance url: %q", source.URL)
		}
	}
	if schema.Type != "array" || schema.MinItems != 6 || schema.MaxItems != 6 || !schema.UniqueItems {
		t.Fatalf("unexpected schema constraints: %#v", schema)
	}
	if !strings.Contains(schema.Description, "ng-geopolitical-zones") {
		t.Fatalf("unexpected schema description: %q", schema.Description)
	}

	seenIDs := make(map[string]struct{}, len(zones))
	seenNames := make(map[string]struct{}, len(zones))
	for i, zone := range zones {
		if zone.ID == "" || zone.Name == "" || zone.CountryCode == "" {
			t.Fatalf("zone record %d has empty required field: %#v", i, zone)
		}
		if !regexp.MustCompile(`^[a-z]+(?:-[a-z]+)*$`).MatchString(zone.ID) {
			t.Fatalf("zone record %d has invalid id %q", i, zone.ID)
		}
		if zone.CountryCode != "NG" {
			t.Fatalf("zone record %d has unexpected country code %q", i, zone.CountryCode)
		}
		if _, ok := seenIDs[zone.ID]; ok {
			t.Fatalf("duplicate zone id found: %q", zone.ID)
		}
		seenIDs[zone.ID] = struct{}{}
		if _, ok := seenNames[zone.Name]; ok {
			t.Fatalf("duplicate zone name found: %q", zone.Name)
		}
		seenNames[zone.Name] = struct{}{}
		if i > 0 && strings.Compare(zones[i-1].Name, zone.Name) > 0 {
			t.Fatalf("zone records are not sorted by name: %q before %q", zones[i-1].Name, zone.Name)
		}
	}

	if len(states) != 37 {
		t.Fatalf("unexpected state record count: got %d want 37", len(states))
	}

	zonesByID := make(map[string]GeopoliticalZone, len(zones))
	for _, zone := range zones {
		zonesByID[zone.ID] = zone
	}

	zoneCounts := make(map[string]int, len(zones))
	for _, zone := range zones {
		zoneCounts[zone.ID] = 0
	}

	stateCount := 0
	fctCount := 0
	fctSeen := false
	for i, state := range states {
		if state.CountryCode != "NG" {
			t.Fatalf("state record %d has unexpected country code %q", i, state.CountryCode)
		}
		if _, ok := zonesByID[state.GeopoliticalZoneID]; !ok {
			t.Fatalf("state record %d references unknown zone %q", i, state.GeopoliticalZoneID)
		}
		zoneCounts[state.GeopoliticalZoneID]++
		switch state.AdministrativeType {
		case "state":
			stateCount++
		case "federal_capital_territory":
			fctCount++
			if state.ID == "fct" {
				fctSeen = true
			}
		default:
			t.Fatalf("state record %d has invalid administrative type %q", i, state.AdministrativeType)
		}
	}

	if stateCount != 36 {
		t.Fatalf("unexpected state count: got %d want 36", stateCount)
	}
	if fctCount != 1 {
		t.Fatalf("unexpected FCT count: got %d want 1", fctCount)
	}
	if !fctSeen {
		t.Fatal("fct record not found in states dataset")
	}

	wantCounts := map[string]int{
		"north-central": 7,
		"north-east":    6,
		"north-west":    7,
		"south-east":    5,
		"south-south":   6,
		"south-west":    6,
	}
	if !reflect.DeepEqual(zoneCounts, wantCounts) {
		t.Fatalf("unexpected zone counts: got %#v want %#v", zoneCounts, wantCounts)
	}

	if metadata.LicenseID != "CC-BY-4.0" || metadata.LicenseURL != "https://creativecommons.org/licenses/by/4.0/" {
		t.Fatalf("unexpected license metadata: %#v", metadata)
	}

	if state := findStateByID(states, "fct"); state == nil || state.GeopoliticalZoneID != "north-central" {
		t.Fatalf("unexpected FCT zone assignment: %#v", state)
	}
}

func loadStateDataset(t *testing.T) []State {
	t.Helper()

	var states []State
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("geography/states.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&states); err != nil {
		t.Fatalf("decode states dataset: %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("dataset contains trailing JSON")
	}
	return states
}

func loadDatasetMetadata(t *testing.T) datasetMetadata {
	t.Helper()

	var metadata datasetMetadata
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("metadata/geography/states.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&metadata); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	return metadata
}

func loadDatasetSchema(t *testing.T) stateSchema {
	t.Helper()

	var schema stateSchema
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("schemas/geography/states.schema.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&schema); err != nil {
		t.Fatalf("decode schema: %v", err)
	}
	return schema
}

func loadGeopoliticalZoneDataset(t *testing.T) []GeopoliticalZone {
	t.Helper()

	var zones []GeopoliticalZone
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("geography/geopolitical_zones.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&zones); err != nil {
		t.Fatalf("decode geopolitical zones dataset: %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("geopolitical zones dataset contains trailing JSON")
	}
	return zones
}

func loadGeopoliticalZoneMetadata(t *testing.T) datasetMetadata {
	t.Helper()

	var metadata datasetMetadata
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("metadata/geography/geopolitical_zones.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&metadata); err != nil {
		t.Fatalf("decode geopolitical zones metadata: %v", err)
	}
	return metadata
}

func loadGeopoliticalZoneSchema(t *testing.T) stateSchema {
	t.Helper()

	var schema stateSchema
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("schemas/geography/geopolitical_zones.schema.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&schema); err != nil {
		t.Fatalf("decode geopolitical zones schema: %v", err)
	}
	return schema
}

func readTextFile(t *testing.T, path string) string {
	t.Helper()

	return string(readTextBytes(t, path))
}

func readTextBytes(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func datasetPath(parts ...string) string {
	elems := append([]string{"..", "..", "datasets"}, parts...)
	return filepath.Join(elems...)
}

func approvedStates() []State {
	return []State{
		{ID: "abia", Name: "Abia", OfficialName: "Abia State", AdministrativeType: "state", Capital: "Umuahia", GeopoliticalZoneID: "south-east", CountryCode: "NG"},
		{ID: "adamawa", Name: "Adamawa", OfficialName: "Adamawa State", AdministrativeType: "state", Capital: "Yola", GeopoliticalZoneID: "north-east", CountryCode: "NG"},
		{ID: "akwa-ibom", Name: "Akwa Ibom", OfficialName: "Akwa Ibom State", AdministrativeType: "state", Capital: "Uyo", GeopoliticalZoneID: "south-south", CountryCode: "NG"},
		{ID: "anambra", Name: "Anambra", OfficialName: "Anambra State", AdministrativeType: "state", Capital: "Awka", GeopoliticalZoneID: "south-east", CountryCode: "NG"},
		{ID: "bauchi", Name: "Bauchi", OfficialName: "Bauchi State", AdministrativeType: "state", Capital: "Bauchi", GeopoliticalZoneID: "north-east", CountryCode: "NG"},
		{ID: "bayelsa", Name: "Bayelsa", OfficialName: "Bayelsa State", AdministrativeType: "state", Capital: "Yenagoa", GeopoliticalZoneID: "south-south", CountryCode: "NG"},
		{ID: "benue", Name: "Benue", OfficialName: "Benue State", AdministrativeType: "state", Capital: "Makurdi", GeopoliticalZoneID: "north-central", CountryCode: "NG"},
		{ID: "borno", Name: "Borno", OfficialName: "Borno State", AdministrativeType: "state", Capital: "Maiduguri", GeopoliticalZoneID: "north-east", CountryCode: "NG"},
		{ID: "cross-river", Name: "Cross River", OfficialName: "Cross River State", AdministrativeType: "state", Capital: "Calabar", GeopoliticalZoneID: "south-south", CountryCode: "NG"},
		{ID: "delta", Name: "Delta", OfficialName: "Delta State", AdministrativeType: "state", Capital: "Asaba", GeopoliticalZoneID: "south-south", CountryCode: "NG"},
		{ID: "ebonyi", Name: "Ebonyi", OfficialName: "Ebonyi State", AdministrativeType: "state", Capital: "Abakaliki", GeopoliticalZoneID: "south-east", CountryCode: "NG"},
		{ID: "edo", Name: "Edo", OfficialName: "Edo State", AdministrativeType: "state", Capital: "Benin City", GeopoliticalZoneID: "south-south", CountryCode: "NG"},
		{ID: "ekiti", Name: "Ekiti", OfficialName: "Ekiti State", AdministrativeType: "state", Capital: "Ado-Ekiti", GeopoliticalZoneID: "south-west", CountryCode: "NG"},
		{ID: "enugu", Name: "Enugu", OfficialName: "Enugu State", AdministrativeType: "state", Capital: "Enugu", GeopoliticalZoneID: "south-east", CountryCode: "NG"},
		{ID: "fct", Name: "Federal Capital Territory", OfficialName: "Federal Capital Territory", AdministrativeType: "federal_capital_territory", Capital: "Abuja", GeopoliticalZoneID: "north-central", CountryCode: "NG"},
		{ID: "gombe", Name: "Gombe", OfficialName: "Gombe State", AdministrativeType: "state", Capital: "Gombe", GeopoliticalZoneID: "north-east", CountryCode: "NG"},
		{ID: "imo", Name: "Imo", OfficialName: "Imo State", AdministrativeType: "state", Capital: "Owerri", GeopoliticalZoneID: "south-east", CountryCode: "NG"},
		{ID: "jigawa", Name: "Jigawa", OfficialName: "Jigawa State", AdministrativeType: "state", Capital: "Dutse", GeopoliticalZoneID: "north-west", CountryCode: "NG"},
		{ID: "kaduna", Name: "Kaduna", OfficialName: "Kaduna State", AdministrativeType: "state", Capital: "Kaduna", GeopoliticalZoneID: "north-west", CountryCode: "NG"},
		{ID: "kano", Name: "Kano", OfficialName: "Kano State", AdministrativeType: "state", Capital: "Kano", GeopoliticalZoneID: "north-west", CountryCode: "NG"},
		{ID: "katsina", Name: "Katsina", OfficialName: "Katsina State", AdministrativeType: "state", Capital: "Katsina", GeopoliticalZoneID: "north-west", CountryCode: "NG"},
		{ID: "kebbi", Name: "Kebbi", OfficialName: "Kebbi State", AdministrativeType: "state", Capital: "Birnin Kebbi", GeopoliticalZoneID: "north-west", CountryCode: "NG"},
		{ID: "kogi", Name: "Kogi", OfficialName: "Kogi State", AdministrativeType: "state", Capital: "Lokoja", GeopoliticalZoneID: "north-central", CountryCode: "NG"},
		{ID: "kwara", Name: "Kwara", OfficialName: "Kwara State", AdministrativeType: "state", Capital: "Ilorin", GeopoliticalZoneID: "north-central", CountryCode: "NG"},
		{ID: "lagos", Name: "Lagos", OfficialName: "Lagos State", AdministrativeType: "state", Capital: "Ikeja", GeopoliticalZoneID: "south-west", CountryCode: "NG"},
		{ID: "nasarawa", Name: "Nasarawa", OfficialName: "Nasarawa State", AdministrativeType: "state", Capital: "Lafia", GeopoliticalZoneID: "north-central", CountryCode: "NG"},
		{ID: "niger", Name: "Niger", OfficialName: "Niger State", AdministrativeType: "state", Capital: "Minna", GeopoliticalZoneID: "north-central", CountryCode: "NG"},
		{ID: "ogun", Name: "Ogun", OfficialName: "Ogun State", AdministrativeType: "state", Capital: "Abeokuta", GeopoliticalZoneID: "south-west", CountryCode: "NG"},
		{ID: "ondo", Name: "Ondo", OfficialName: "Ondo State", AdministrativeType: "state", Capital: "Akure", GeopoliticalZoneID: "south-west", CountryCode: "NG"},
		{ID: "osun", Name: "Osun", OfficialName: "Osun State", AdministrativeType: "state", Capital: "Osogbo", GeopoliticalZoneID: "south-west", CountryCode: "NG"},
		{ID: "oyo", Name: "Oyo", OfficialName: "Oyo State", AdministrativeType: "state", Capital: "Ibadan", GeopoliticalZoneID: "south-west", CountryCode: "NG"},
		{ID: "plateau", Name: "Plateau", OfficialName: "Plateau State", AdministrativeType: "state", Capital: "Jos", GeopoliticalZoneID: "north-central", CountryCode: "NG"},
		{ID: "rivers", Name: "Rivers", OfficialName: "Rivers State", AdministrativeType: "state", Capital: "Port Harcourt", GeopoliticalZoneID: "south-south", CountryCode: "NG"},
		{ID: "sokoto", Name: "Sokoto", OfficialName: "Sokoto State", AdministrativeType: "state", Capital: "Sokoto", GeopoliticalZoneID: "north-west", CountryCode: "NG"},
		{ID: "taraba", Name: "Taraba", OfficialName: "Taraba State", AdministrativeType: "state", Capital: "Jalingo", GeopoliticalZoneID: "north-east", CountryCode: "NG"},
		{ID: "yobe", Name: "Yobe", OfficialName: "Yobe State", AdministrativeType: "state", Capital: "Damaturu", GeopoliticalZoneID: "north-east", CountryCode: "NG"},
		{ID: "zamfara", Name: "Zamfara", OfficialName: "Zamfara State", AdministrativeType: "state", Capital: "Gusau", GeopoliticalZoneID: "north-west", CountryCode: "NG"},
	}
}

func approvedGeopoliticalZones() []GeopoliticalZone {
	return []GeopoliticalZone{
		{ID: "north-central", Name: "North Central", CountryCode: "NG"},
		{ID: "north-east", Name: "North East", CountryCode: "NG"},
		{ID: "north-west", Name: "North West", CountryCode: "NG"},
		{ID: "south-east", Name: "South East", CountryCode: "NG"},
		{ID: "south-south", Name: "South South", CountryCode: "NG"},
		{ID: "south-west", Name: "South West", CountryCode: "NG"},
	}
}

func findStateByID(states []State, id string) *State {
	for i := range states {
		if states[i].ID == id {
			return &states[i]
		}
	}
	return nil
}
