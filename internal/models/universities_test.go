package models

import (
	"bytes"
	"encoding/json"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

type universityMetadata struct {
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

type universitySchema struct {
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

func TestNigeriaUniversitiesDatasetMatchesVerifiedRegister(t *testing.T) {
	states := loadStateDataset(t)
	universities := loadUniversityDataset(t)

	if universities == nil {
		t.Fatal("decoded dataset is nil")
	}
	if got := len(universities); got != 328 {
		t.Fatalf("unexpected record count: got %d want 328", got)
	}

	stateByID := make(map[string]State, len(states))
	for _, state := range states {
		stateByID[state.ID] = state
	}

	wantCounts := map[string]int{
		"abia": 8, "adamawa": 4, "akwa-ibom": 9, "anambra": 10, "bauchi": 3,
		"bayelsa": 6, "benue": 5, "borno": 4, "cross-river": 8, "delta": 15,
		"ebonyi": 7, "edo": 11, "ekiti": 8, "enugu": 11, "fct": 21, "gombe": 4,
		"imo": 11, "jigawa": 4, "kaduna": 11, "kano": 14, "katsina": 5,
		"kebbi": 4, "kogi": 6, "kwara": 14, "lagos": 16, "nasarawa": 6,
		"niger": 10, "ogun": 25, "ondo": 11, "osun": 15, "oyo": 13,
		"plateau": 5, "rivers": 7, "sokoto": 6, "taraba": 5, "yobe": 2,
		"zamfara": 4,
	}
	wantOwnershipCounts := map[string]int{
		"federal": 77,
		"state":   69,
		"private": 182,
	}
	wantSpecific := map[string]struct {
		stateID string
		own     string
	}{
		"abdulkadir-kure-university-minna":          {stateID: "niger", own: "state"},
		"adeleke-university-ede":                    {stateID: "osun", own: "private"},
		"african-aviation-and-aerospace-university": {stateID: "fct", own: "federal"},
		"bayelsa-medical-university":                {stateID: "bayelsa", own: "state"},
		"baze-university":                           {stateID: "fct", own: "private"},
		"crawford-university-igbesa":                {stateID: "ogun", own: "private"},
		"cross-river-university-of-education-and-entrepreneurship-akamkpa-cross-river-state": {stateID: "cross-river", own: "state"},
		"european-university-of-nigeria-duboyi-fct":                                          {stateID: "fct", own: "private"},
		"federal-university-of-science-and-technology-epe":                                   {stateID: "lagos", own: "federal"},
		"fountain-university-osogbo":                                                         {stateID: "osun", own: "private"},
		"gombe-state-university-gombe":                                                       {stateID: "gombe", own: "state"},
		"leadership-university-abuja-fct":                                                    {stateID: "fct", own: "private"},
		"niger-delta-university-yenagoa":                                                     {stateID: "bayelsa", own: "state"},
		"nigerian-army-university-biu":                                                       {stateID: "borno", own: "federal"},
		"olusegun-agagu-university-of-science-and-technology-okitipupa-ondo":                 {stateID: "ondo", own: "state"},
		"redeemer-s-university-ede":                                                          {stateID: "osun", own: "private"},
		"rev-fr-moses-orshio-adasu-makurdi":                                                  {stateID: "benue", own: "state"},
		"summit-university-offa":                                                             {stateID: "kwara", own: "private"},
		"tai-solarin-federal-university-of-education-ijagun-ijebu-ode":                       {stateID: "ogun", own: "federal"},
		"tansian-university-umunya":                                                          {stateID: "anambra", own: "private"},
		"university-of-port-harcourt":                                                        {stateID: "rivers", own: "federal"},
	}

	seenIDs := make(map[string]struct{}, len(universities))
	seenNamesByState := make(map[string]map[string]struct{}, len(states))
	stateCounts := make(map[string]int, len(states))
	ownershipCounts := make(map[string]int, len(wantOwnershipCounts))
	currentOwnership := ""
	currentOrder := -1
	lastName := ""

	for i, university := range universities {
		if university.ID == "" || university.Name == "" || university.OwnershipType == "" || university.StateID == "" || university.CountryCode == "" {
			t.Fatalf("record %d has empty required field: %#v", i, university)
		}
		if university.CountryCode != "NG" {
			t.Fatalf("record %d has unexpected country code %q", i, university.CountryCode)
		}
		if _, ok := stateByID[university.StateID]; !ok {
			t.Fatalf("record %d references unknown state %q", i, university.StateID)
		}
		if !regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(university.ID) {
			t.Fatalf("record %d has invalid id %q", i, university.ID)
		}
		if wantID := slugifyRecordName(university.Name); university.ID != wantID {
			t.Fatalf("record %d id mismatch: got %q want %q", i, university.ID, wantID)
		}
		if _, ok := seenIDs[university.ID]; ok {
			t.Fatalf("duplicate id found: %q", university.ID)
		}
		seenIDs[university.ID] = struct{}{}

		if _, ok := wantCounts[university.StateID]; !ok {
			t.Fatalf("record %d has unsupported state %q", i, university.StateID)
		}
		stateCounts[university.StateID]++
		ownershipCounts[university.OwnershipType]++

		stateNames, ok := seenNamesByState[university.StateID]
		if !ok {
			stateNames = make(map[string]struct{})
			seenNamesByState[university.StateID] = stateNames
		}
		if _, ok := stateNames[university.Name]; ok {
			t.Fatalf("duplicate name within state %q: %q", university.StateID, university.Name)
		}
		stateNames[university.Name] = struct{}{}

		lowerName := strings.ToLower(university.Name)
		if strings.Contains(lowerName, "formerly") {
			t.Fatalf("former-name annotation leaked into public name: %q", university.Name)
		}

		if currentOwnership != university.OwnershipType {
			order := map[string]int{"federal": 0, "state": 1, "private": 2}[university.OwnershipType]
			if order < currentOrder {
				t.Fatalf("ownership groups are out of order: %q before %q", currentOwnership, university.OwnershipType)
			}
			currentOwnership = university.OwnershipType
			currentOrder = order
			lastName = ""
		}
		if lastName != "" && strings.Compare(strings.ToLower(lastName), strings.ToLower(university.Name)) > 0 {
			t.Fatalf("records are not sorted by name within ownership %q: %q before %q", university.OwnershipType, lastName, university.Name)
		}
		lastName = university.Name

		marshaled, err := json.Marshal(university)
		if err != nil {
			t.Fatalf("marshal university %d: %v", i, err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(marshaled, &raw); err != nil {
			t.Fatalf("unmarshal marshaled university %d: %v", i, err)
		}
		if len(raw) != 5 {
			t.Fatalf("record %d marshaled with unexpected field count: got %d want 5", i, len(raw))
		}
		for _, field := range []string{"id", "name", "ownership_type", "state_id", "country_code"} {
			if _, ok := raw[field]; !ok {
				t.Fatalf("record %d missing serialized field %q", i, field)
			}
		}
	}

	if len(seenNamesByState) != len(states) {
		t.Fatalf("unexpected parent coverage: got %d states with children want %d", len(seenNamesByState), len(states))
	}
	for _, state := range states {
		if got := stateCounts[state.ID]; got != wantCounts[state.ID] {
			t.Fatalf("unexpected count for %s: got %d want %d", state.ID, got, wantCounts[state.ID])
		}
	}
	if !reflect.DeepEqual(ownershipCounts, wantOwnershipCounts) {
		t.Fatalf("unexpected ownership totals: got %#v want %#v", ownershipCounts, wantOwnershipCounts)
	}

	for wantID, want := range wantSpecific {
		var found *University
		for i := range universities {
			if universities[i].ID == wantID {
				found = &universities[i]
				break
			}
		}
		if found == nil {
			t.Fatalf("missing expected university id %q", wantID)
		}
		if found.StateID != want.stateID {
			t.Fatalf("unexpected state for %q: got %q want %q", wantID, found.StateID, want.stateID)
		}
		if found.OwnershipType != want.own {
			t.Fatalf("unexpected ownership type for %q: got %q want %q", wantID, found.OwnershipType, want.own)
		}
	}
}

func TestNigeriaUniversitiesMetadataSchemaAndNotice(t *testing.T) {
	metadata := loadUniversityMetadata(t)
	schema := loadUniversitySchema(t)

	if metadata.DatasetKey != "ng-universities" {
		t.Fatalf("unexpected dataset key: %q", metadata.DatasetKey)
	}
	if metadata.Title != "Nigeria Universities" {
		t.Fatalf("unexpected title: %q", metadata.Title)
	}
	if metadata.Description == "" || !strings.Contains(metadata.Description, "current National Universities Commission register") {
		t.Fatalf("unexpected description: %q", metadata.Description)
	}
	if metadata.CountryCode != "NG" || metadata.DatasetGroup != "education" || metadata.Format != "json" {
		t.Fatalf("unexpected metadata classification: %#v", metadata)
	}
	if metadata.RelativePath != "education/universities.json" {
		t.Fatalf("unexpected relative path: %q", metadata.RelativePath)
	}
	if metadata.SchemaPath != "schemas/education/universities.schema.json" {
		t.Fatalf("unexpected schema path: %q", metadata.SchemaPath)
	}
	if metadata.RecordCount != 328 {
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
	for _, source := range metadata.Sources {
		if source.Organization == "" || source.Title == "" || source.URL == "" || source.Purpose == "" || source.AccessedAt == "" {
			t.Fatalf("incomplete provenance source: %#v", source)
		}
		if !strings.HasPrefix(source.URL, "https://") {
			t.Fatalf("non-https provenance url: %q", source.URL)
		}
	}
	for _, want := range []string{
		"retrieved three times with consistent results",
		"77",
		"69",
		"182",
		"former-name annotations",
		"do not establish a universal institution identifier",
	} {
		if !strings.Contains(metadata.Methodology, want) {
			t.Fatalf("methodology missing %q: %q", want, metadata.Methodology)
		}
	}

	if schema.Schema != "https://json-schema.org/draft/2020-12/schema" {
		t.Fatalf("unexpected schema dialect: %q", schema.Schema)
	}
	if schema.ID != "https://softdata-api.local/schemas/education/universities.schema.json" {
		t.Fatalf("unexpected schema id: %q", schema.ID)
	}
	if schema.Title != "Nigeria Universities" {
		t.Fatalf("unexpected schema title: %q", schema.Title)
	}
	if schema.Type != "array" || schema.MinItems != 328 || schema.MaxItems != 328 || !schema.UniqueItems {
		t.Fatalf("unexpected schema constraints: %#v", schema)
	}

	var items map[string]any
	if err := json.Unmarshal(schema.Items, &items); err != nil {
		t.Fatalf("decode schema items: %v", err)
	}
	if got := items["type"]; got != "object" {
		t.Fatalf("unexpected schema item type: %#v", got)
	}
	if got := items["additionalProperties"]; got != false {
		t.Fatalf("unexpected additionalProperties: %#v", got)
	}
	required := asStringSlice(t, items["required"])
	if !reflect.DeepEqual(required, []string{"id", "name", "ownership_type", "state_id", "country_code"}) {
		t.Fatalf("unexpected required fields: %#v", required)
	}
	props := asMap(t, items["properties"])
	ownershipProp := asMap(t, props["ownership_type"])
	ownershipEnum := asStringSlice(t, ownershipProp["enum"])
	if !reflect.DeepEqual(ownershipEnum, []string{"federal", "state", "private"}) {
		t.Fatalf("unexpected ownership_type enum: %#v", ownershipEnum)
	}
	stateProp := asMap(t, props["state_id"])
	stateEnum := asStringSlice(t, stateProp["enum"])
	if len(stateEnum) != 37 || !containsString(stateEnum, "fct") || !containsString(stateEnum, "taraba") || !containsString(stateEnum, "abia") {
		t.Fatalf("unexpected state_id enum: %#v", stateEnum)
	}
	countryProp := asMap(t, props["country_code"])
	if got := countryProp["const"]; got != "NG" {
		t.Fatalf("unexpected country_code const: %#v", got)
	}
	if schema.Description == "" || !strings.Contains(schema.Description, "ng-universities") {
		t.Fatalf("unexpected schema description: %q", schema.Description)
	}
}

func loadUniversityDataset(t *testing.T) []University {
	t.Helper()

	var universities []University
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("education/universities.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&universities); err != nil {
		t.Fatalf("decode universities dataset: %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("universities dataset contains trailing JSON")
	}
	return universities
}

func loadUniversityMetadata(t *testing.T) universityMetadata {
	t.Helper()

	var metadata universityMetadata
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("metadata/education/universities.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&metadata); err != nil {
		t.Fatalf("decode universities metadata: %v", err)
	}
	return metadata
}

func loadUniversitySchema(t *testing.T) universitySchema {
	t.Helper()

	var schema universitySchema
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("schemas/education/universities.schema.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&schema); err != nil {
		t.Fatalf("decode universities schema: %v", err)
	}
	return schema
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
