package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

type lgaSchema struct {
	Schema      string         `json:"$schema"`
	ID          string         `json:"$id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Type        string         `json:"type"`
	MinItems    int            `json:"minItems"`
	MaxItems    int            `json:"maxItems"`
	UniqueItems bool           `json:"uniqueItems"`
	Items       map[string]any `json:"items"`
}

func TestNigeriaLocalGovernmentUnitsDatasetAndCrossReferences(t *testing.T) {
	states := loadStateDataset(t)
	zones := loadGeopoliticalZoneDataset(t)
	units := loadLocalGovernmentUnitDataset(t)

	if len(units) != 774 {
		t.Fatalf("unexpected record count: got %d want 774", len(units))
	}

	stateByID := make(map[string]State, len(states))
	stateOrder := make(map[string]int, len(states))
	for i, state := range states {
		stateByID[state.ID] = state
		stateOrder[state.ID] = i
	}

	zoneByID := make(map[string]GeopoliticalZone, len(zones))
	for _, zone := range zones {
		zoneByID[zone.ID] = zone
	}

	wantCounts := map[string]int{
		"abia": 17, "adamawa": 21, "akwa-ibom": 31, "anambra": 21, "bauchi": 20,
		"bayelsa": 8, "benue": 23, "borno": 27, "cross-river": 18, "delta": 25,
		"ebonyi": 13, "edo": 18, "ekiti": 16, "enugu": 17, "fct": 6, "gombe": 11,
		"imo": 27, "jigawa": 27, "kaduna": 23, "kano": 44, "katsina": 34,
		"kebbi": 21, "kogi": 21, "kwara": 16, "lagos": 20, "nasarawa": 13,
		"niger": 25, "ogun": 20, "ondo": 18, "osun": 30, "oyo": 33, "plateau": 17,
		"rivers": 23, "sokoto": 23, "taraba": 16, "yobe": 17, "zamfara": 14,
	}
	wantZoneCounts := map[string]int{
		"north-central": 121,
		"north-east":    112,
		"north-west":    186,
		"south-east":    95,
		"south-south":   123,
		"south-west":    137,
	}
	wantFCT := map[string]string{
		"fct-abaji":           "Abaji",
		"fct-abuja-municipal": "Abuja Municipal",
		"fct-bwari":           "Bwari",
		"fct-gwagwalada":      "Gwagwalada",
		"fct-kuje":            "Kuje",
		"fct-kwali":           "Kwali",
	}

	seenIDs := make(map[string]struct{}, len(units))
	seenNamesByState := make(map[string]map[string]struct{}, len(states))
	stateCounts := make(map[string]int, len(states))
	zoneCounts := map[string]int{
		"north-central": 0,
		"north-east":    0,
		"north-west":    0,
		"south-east":    0,
		"south-south":   0,
		"south-west":    0,
	}
	seenStates := make(map[string]struct{}, len(states))
	fctSeen := make(map[string]string, len(wantFCT))
	currentState := ""
	currentOrder := -1
	lastName := ""

	for i, unit := range units {
		if unit.ID == "" || unit.Name == "" || unit.StateID == "" || unit.CountryCode == "" || unit.AdministrativeType == "" {
			t.Fatalf("record %d has empty required field: %#v", i, unit)
		}
		if unit.CountryCode != "NG" {
			t.Fatalf("record %d has unexpected country code %q", i, unit.CountryCode)
		}
		if _, ok := stateByID[unit.StateID]; !ok {
			t.Fatalf("record %d references unknown state %q", i, unit.StateID)
		}
		if !regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`).MatchString(unit.ID) {
			t.Fatalf("record %d has invalid id %q", i, unit.ID)
		}
		if !strings.HasPrefix(unit.ID, unit.StateID+"-") {
			t.Fatalf("record %d id %q does not start with state prefix %q", i, unit.ID, unit.StateID+"-")
		}
		if wantID := unit.StateID + "-" + slugifyRecordName(unit.Name); unit.ID != wantID {
			t.Fatalf("record %d id mismatch: got %q want %q", i, unit.ID, wantID)
		}
		if unit.StateID == "fct" {
			if unit.AdministrativeType != "area_council" {
				t.Fatalf("record %d expected area_council for fct: %#v", i, unit)
			}
			fctSeen[unit.ID] = unit.Name
		} else if unit.AdministrativeType != "local_government_area" {
			t.Fatalf("record %d expected local_government_area: %#v", i, unit)
		}
		if _, ok := seenIDs[unit.ID]; ok {
			t.Fatalf("duplicate id found: %q", unit.ID)
		}
		seenIDs[unit.ID] = struct{}{}

		nameSet, ok := seenNamesByState[unit.StateID]
		if !ok {
			nameSet = make(map[string]struct{})
			seenNamesByState[unit.StateID] = nameSet
		}
		if _, ok := nameSet[unit.Name]; ok {
			t.Fatalf("duplicate name within state %q: %q", unit.StateID, unit.Name)
		}
		nameSet[unit.Name] = struct{}{}

		stateCounts[unit.StateID]++
		seenStates[unit.StateID] = struct{}{}
		zoneID := stateByID[unit.StateID].GeopoliticalZoneID
		if _, ok := zoneByID[zoneID]; !ok {
			t.Fatalf("record %d references unknown zone %q", i, zoneID)
		}
		zoneCounts[zoneID]++

		order := stateOrder[unit.StateID]
		if currentState != unit.StateID {
			if order < currentOrder {
				t.Fatalf("state groups are out of order: %q before %q", currentState, unit.StateID)
			}
			currentState = unit.StateID
			currentOrder = order
			lastName = ""
		}
		if lastName != "" && strings.Compare(strings.ToLower(lastName), strings.ToLower(unit.Name)) > 0 {
			t.Fatalf("records are not sorted by name within state %q: %q before %q", unit.StateID, lastName, unit.Name)
		}
		lastName = unit.Name

		marshaled, err := json.Marshal(unit)
		if err != nil {
			t.Fatalf("marshal unit %d: %v", i, err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(marshaled, &raw); err != nil {
			t.Fatalf("unmarshal marshaled unit %d: %v", i, err)
		}
		if len(raw) != 5 {
			t.Fatalf("record %d marshaled with unexpected field count: got %d want 5", i, len(raw))
		}
		for _, field := range []string{"id", "name", "state_id", "country_code", "administrative_type"} {
			if _, ok := raw[field]; !ok {
				t.Fatalf("record %d missing serialized field %q", i, field)
			}
		}
	}

	if len(seenStates) != len(states) {
		t.Fatalf("unexpected parent coverage: got %d states with children want %d", len(seenStates), len(states))
	}
	for _, state := range states {
		if got := stateCounts[state.ID]; got != wantCounts[state.ID] {
			t.Fatalf("unexpected count for %s: got %d want %d", state.ID, got, wantCounts[state.ID])
		}
	}
	if !reflect.DeepEqual(zoneCounts, wantZoneCounts) {
		t.Fatalf("unexpected zone totals: got %#v want %#v", zoneCounts, wantZoneCounts)
	}
	if !reflect.DeepEqual(fctSeen, wantFCT) {
		t.Fatalf("unexpected FCT units: got %#v want %#v", fctSeen, wantFCT)
	}
	if len(fctSeen) != 6 {
		t.Fatalf("unexpected FCT count: got %d want 6", len(fctSeen))
	}
}

func TestNigeriaLocalGovernmentUnitsMetadataSchemaAndNotice(t *testing.T) {
	metadata := loadLocalGovernmentUnitMetadata(t)
	schema := loadLocalGovernmentUnitSchema(t)
	licenseText := readTextFile(t, datasetPath("LICENSE.md"))
	readmeText := readTextFile(t, datasetPath("README.md"))

	if metadata.DatasetKey != "ng-lgas" {
		t.Fatalf("unexpected dataset key: %q", metadata.DatasetKey)
	}
	if metadata.Title != "Nigeria Local Government Areas and FCT Area Councils" {
		t.Fatalf("unexpected title: %q", metadata.Title)
	}
	if metadata.Description == "" || !strings.Contains(metadata.Description, "768 Local Government Areas") || !strings.Contains(metadata.Description, "six Area Councils") {
		t.Fatalf("unexpected description: %q", metadata.Description)
	}
	if metadata.CountryCode != "NG" || metadata.DatasetGroup != "geography" || metadata.Format != "json" {
		t.Fatalf("unexpected metadata classification: %#v", metadata)
	}
	if metadata.RelativePath != "geography/lgas.json" {
		t.Fatalf("unexpected relative path: %q", metadata.RelativePath)
	}
	if metadata.SchemaPath != "schemas/geography/lgas.schema.json" {
		t.Fatalf("unexpected schema path: %q", metadata.SchemaPath)
	}
	if metadata.RecordCount != 774 {
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
	verifiedAt, err := parseISODate(metadata.VerifiedAt)
	if err != nil {
		t.Fatalf("verified_at is not a valid date: %v", err)
	}
	if verifiedAt.IsZero() {
		t.Fatal("verified_at is empty")
	}
	nowUTC := time.Now().UTC().Truncate(24 * time.Hour)
	if verifiedAt.After(nowUTC) {
		t.Fatalf("verified_at is in the future: %s > %s", verifiedAt.Format("2006-01-02"), nowUTC.Format("2006-01-02"))
	}
	if verifiedAt.Format("2006-01-02") != "2026-08-28" {
		t.Fatalf("unexpected verified_at: %q", metadata.VerifiedAt)
	}
	futureVerifiedAt := nowUTC.Add(24 * time.Hour).Format("2006-01-02")
	for _, invalid := range []string{"", "2026-13-01", futureVerifiedAt, "not-a-date"} {
		if _, err := parseVerifiedAt(invalid, nowUTC); err == nil {
			t.Fatalf("expected invalid verified_at to fail: %q", invalid)
		}
	}
	if len(metadata.Sources) != 4 {
		t.Fatalf("unexpected source count: %d", len(metadata.Sources))
	}
	for _, source := range metadata.Sources {
		if source.Organization == "" || source.Title == "" || source.URL == "" || source.Purpose == "" || source.AccessedAt == "" {
			t.Fatalf("incomplete provenance source: %#v", source)
		}
		if !strings.HasPrefix(source.URL, "https://") {
			t.Fatalf("non-https provenance url: %q", source.URL)
		}
	}
	for _, snippet := range []string{"independently compiled", "title case", "Area Councils", "state dataset", "Abuja Municipal"} {
		if !strings.Contains(strings.ToLower(metadata.Methodology), strings.ToLower(snippet)) {
			t.Fatalf("methodology missing snippet %q: %q", snippet, metadata.Methodology)
		}
	}
	for _, forbidden := range []string{"licensed under CC BY 4.0", "government publication is CC BY 4.0", "source publications are CC BY 4.0"} {
		if strings.Contains(strings.ToLower(metadata.Methodology), strings.ToLower(forbidden)) {
			t.Fatalf("metadata contains forbidden licence claim: %q", forbidden)
		}
	}
	if !strings.Contains(licenseText+readmeText, "SoftData's independently compiled dataset package is available under CC BY 4.0.") {
		t.Fatal("missing CC BY 4.0 boundary notice")
	}
	if !strings.Contains(licenseText+readmeText, "The original source organizations retain any rights they hold in their own publications.") {
		t.Fatal("missing source-rights boundary notice")
	}

	if schema.Schema != "https://json-schema.org/draft/2020-12/schema" {
		t.Fatalf("unexpected schema draft: %q", schema.Schema)
	}
	if schema.Type != "array" || schema.MinItems != 774 || schema.MaxItems != 774 || !schema.UniqueItems {
		t.Fatalf("unexpected schema top-level constraints: %#v", schema)
	}
	if schema.Title != "Nigeria Local Government Areas and FCT Area Councils" {
		t.Fatalf("unexpected schema title: %q", schema.Title)
	}
	if !strings.Contains(schema.Description, "ng-lgas") {
		t.Fatalf("unexpected schema description: %q", schema.Description)
	}

	items := schema.Items
	if got, ok := items["type"].(string); !ok || got != "object" {
		t.Fatalf("unexpected item type: %#v", items["type"])
	}
	if got, ok := items["additionalProperties"].(bool); !ok || got {
		t.Fatalf("unexpected additionalProperties: %#v", items["additionalProperties"])
	}
	required := asStringSlice(t, items["required"])
	if !reflect.DeepEqual(required, []string{"id", "name", "state_id", "country_code", "administrative_type"}) {
		t.Fatalf("unexpected required fields: %#v", required)
	}
	props := asMap(t, items["properties"])
	idProp := asMap(t, props["id"])
	if got, ok := idProp["pattern"].(string); !ok || got != "^[a-z0-9]+(?:-[a-z0-9]+)+$" {
		t.Fatalf("unexpected id pattern: %#v", idProp["pattern"])
	}
	stateProp := asMap(t, props["state_id"])
	stateEnum := asStringSlice(t, stateProp["enum"])
	if len(stateEnum) != 37 || stateEnum[14] != "fct" || stateEnum[0] != "abia" || stateEnum[len(stateEnum)-1] != "zamfara" {
		t.Fatalf("unexpected state_id enum: %#v", stateEnum)
	}
	adminProp := asMap(t, props["administrative_type"])
	adminEnum := asStringSlice(t, adminProp["enum"])
	if !reflect.DeepEqual(adminEnum, []string{"local_government_area", "area_council"}) {
		t.Fatalf("unexpected administrative_type enum: %#v", adminEnum)
	}
	countryProp := asMap(t, props["country_code"])
	if got, ok := countryProp["const"].(string); !ok || got != "NG" {
		t.Fatalf("unexpected country code const: %#v", countryProp["const"])
	}

	allOf := asSlice(t, items["allOf"])
	if len(allOf) != 2 {
		t.Fatalf("unexpected conditional count: %d", len(allOf))
	}
	first := asMap(t, allOf[0])
	firstIf := asMap(t, first["if"])
	firstThen := asMap(t, first["then"])
	if got, ok := asMap(t, asMap(t, firstIf["properties"])["state_id"])["const"].(string); !ok || got != "fct" {
		t.Fatalf("unexpected fct conditional: %#v", first)
	}
	if got, ok := asMap(t, asMap(t, firstThen["properties"])["administrative_type"])["const"].(string); !ok || got != "area_council" {
		t.Fatalf("unexpected fct then clause: %#v", firstThen)
	}
	second := asMap(t, allOf[1])
	secondIf := asMap(t, second["if"])
	secondThen := asMap(t, second["then"])
	if _, ok := asMap(t, asMap(t, secondIf["properties"])["state_id"])["not"]; !ok {
		t.Fatalf("missing non-fct conditional: %#v", secondIf)
	}
	if got, ok := asMap(t, asMap(t, secondThen["properties"])["administrative_type"])["const"].(string); !ok || got != "local_government_area" {
		t.Fatalf("unexpected non-fct then clause: %#v", secondThen)
	}
}

func loadLocalGovernmentUnitDataset(t *testing.T) []LocalGovernmentUnit {
	t.Helper()

	var units []LocalGovernmentUnit
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("geography/lgas.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&units); err != nil {
		t.Fatalf("decode lgas dataset: %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("lgas dataset contains trailing JSON")
	}
	return units
}

func loadLocalGovernmentUnitMetadata(t *testing.T) datasetMetadata {
	t.Helper()

	var metadata datasetMetadata
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("metadata/geography/lgas.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&metadata); err != nil {
		t.Fatalf("decode lgas metadata: %v", err)
	}
	return metadata
}

func loadLocalGovernmentUnitSchema(t *testing.T) lgaSchema {
	t.Helper()

	var schema lgaSchema
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("schemas/geography/lgas.schema.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&schema); err != nil {
		t.Fatalf("decode lgas schema: %v", err)
	}
	return schema
}

func slugifyRecordName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, ".", "")
	name = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(name, "-")
	name = regexp.MustCompile(`-+`).ReplaceAllString(name, "-")
	return strings.Trim(name, "-")
}

func asMap(t *testing.T, value any) map[string]any {
	t.Helper()

	result, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", value)
	}
	return result
}

func asSlice(t *testing.T, value any) []any {
	t.Helper()

	result, ok := value.([]any)
	if !ok {
		t.Fatalf("expected slice, got %T", value)
	}
	return result
}

func asStringSlice(t *testing.T, value any) []string {
	t.Helper()

	raw := asSlice(t, value)
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		str, ok := item.(string)
		if !ok {
			t.Fatalf("expected string slice element, got %T", item)
		}
		result = append(result, str)
	}
	return result
}

var (
	errEmptyVerifiedAt  = errors.New("verified_at is empty")
	errFutureVerifiedAt = errors.New("verified_at is in the future")
)

func parseVerifiedAt(value string, nowUTC time.Time) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, errEmptyVerifiedAt
	}
	parsed, err := parseISODate(value)
	if err != nil {
		return time.Time{}, err
	}
	if parsed.After(nowUTC) {
		return time.Time{}, errFutureVerifiedAt
	}
	return parsed, nil
}

func parseISODate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}
