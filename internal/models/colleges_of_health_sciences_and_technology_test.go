package models

import (
	"bytes"
	"encoding/json"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type healthCollegeMetadata struct {
	DatasetKey       string         `json:"dataset_key"`
	OfficialCategory string         `json:"official_category_name"`
	RecordCount      int            `json:"record_count"`
	OwnershipCounts  map[string]int `json:"ownership_counts"`
	StateCoverage    []string       `json:"state_coverage"`
}

type healthCollegeReconciliation struct {
	RawRecordCount       int               `json:"raw_record_count"`
	FinalRecordCount     int               `json:"final_record_count"`
	ClassificationCounts map[string]int    `json:"classification_counts"`
	Records              []json.RawMessage `json:"records"`
}

func TestNigeriaHealthCollegeDatasetMatchesReconciledRoster(t *testing.T) {
	values := loadHealthCollegeDataset(t)
	states := loadStateDataset(t)
	polytechnics := loadPolytechnicDataset(t)
	monotechnics := loadMonotechnicDataset(t)
	agriculture := loadCollegeOfAgricultureDataset(t)
	if len(values) != 98 {
		t.Fatalf("unexpected record count: got %d want 98", len(values))
	}
	stateIDs := make(map[string]struct{}, len(states))
	for _, state := range states {
		stateIDs[state.ID] = struct{}{}
	}
	seenIDs, seenNames := map[string]struct{}{}, map[string]struct{}{}
	ownership := map[string]int{}
	ordered := append([]CollegeOfHealthSciencesAndTechnology(nil), values...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := strings.ToLower(ordered[i].Name), strings.ToLower(ordered[j].Name)
		if left == right {
			return ordered[i].ID < ordered[j].ID
		}
		return left < right
	})
	if !reflect.DeepEqual(values, ordered) {
		t.Fatal("dataset is not sorted by name then id")
	}
	for i, value := range values {
		if value.ID == "" || value.Name == "" || value.OwnershipType == "" || value.StateID == "" || value.CountryCode != "NG" {
			t.Fatalf("invalid record %d: %#v", i, value)
		}
		if !regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(value.ID) {
			t.Fatalf("invalid id %q", value.ID)
		}
		if _, ok := stateIDs[value.StateID]; !ok {
			t.Fatalf("unknown state %q", value.StateID)
		}
		if _, ok := seenIDs[value.ID]; ok {
			t.Fatalf("duplicate id %q", value.ID)
		}
		if _, ok := seenNames[value.Name]; ok {
			t.Fatalf("duplicate name %q", value.Name)
		}
		seenIDs[value.ID], seenNames[value.Name] = struct{}{}, struct{}{}
		ownership[value.OwnershipType]++
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(bytes.NewReader(encoded)).Decode(&raw); err != nil {
			t.Fatal(err)
		}
		if len(raw) != 5 {
			t.Fatalf("unexpected field count for %q", value.ID)
		}
	}
	if !reflect.DeepEqual(ownership, map[string]int{"federal": 4, "state": 31, "private": 63}) {
		t.Fatalf("unexpected ownership counts: %#v", ownership)
	}
	other := map[string]struct{}{}
	for _, value := range polytechnics {
		other[value.Name] = struct{}{}
	}
	for _, value := range monotechnics {
		other[value.Name] = struct{}{}
	}
	for _, value := range agriculture {
		other[value.Name] = struct{}{}
	}
	for _, value := range values {
		if _, ok := other[value.Name]; ok {
			t.Fatalf("improper cross-dataset overlap: %q", value.Name)
		}
	}
	if len(stateIDsForHealth(values)) != 33 {
		t.Fatalf("unexpected state coverage")
	}
}

func TestNigeriaHealthCollegeMetadataSchemaAndReconciliation(t *testing.T) {
	metadata := loadHealthCollegeMetadata(t)
	schema := loadHealthCollegeSchema(t)
	reconciliation := loadHealthCollegeReconciliation(t)
	if metadata.DatasetKey != "ng-colleges-of-health-sciences-and-technology" || metadata.OfficialCategory != "Colleges of Health Sciences and Technology" || metadata.RecordCount != 98 {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
	if !reflect.DeepEqual(metadata.OwnershipCounts, map[string]int{"federal": 4, "state": 31, "private": 63}) {
		t.Fatalf("unexpected ownership metadata: %#v", metadata.OwnershipCounts)
	}
	if schema.MinItems != 98 || schema.MaxItems != 98 || !schema.UniqueItems {
		t.Fatal("schema does not match dataset")
	}
	if reconciliation.RawRecordCount != 129 || reconciliation.FinalRecordCount != 98 || len(reconciliation.Records) != 129 || reconciliation.ClassificationCounts["exclude_teaching_hospital_only"] != 29 {
		t.Fatalf("reconciliation mismatch: %#v", reconciliation)
	}
	values := loadHealthCollegeDataset(t)
	for _, anchor := range []string{"Federal College of Orthopaedic Technology, Igbobi", "Kwara State College of Health Technology, Offa", "Adeshina College of Health Sciences and Technology, Share"} {
		if _, ok := findHealthCollege(values, anchor); !ok {
			t.Fatalf("missing anchor %q", anchor)
		}
	}
}

func loadHealthCollegeDataset(t *testing.T) []CollegeOfHealthSciencesAndTechnology {
	t.Helper()
	var values []CollegeOfHealthSciencesAndTechnology
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("education/colleges_of_health_sciences_and_technology.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&values); err != nil {
		t.Fatal(err)
	}
	return values
}
func loadHealthCollegeMetadata(t *testing.T) healthCollegeMetadata {
	t.Helper()
	var value healthCollegeMetadata
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/education/colleges_of_health_sciences_and_technology.json")), &value); err != nil {
		t.Fatal(err)
	}
	return value
}
func loadHealthCollegeSchema(t *testing.T) struct {
	MinItems    int  `json:"minItems"`
	MaxItems    int  `json:"maxItems"`
	UniqueItems bool `json:"uniqueItems"`
} {
	t.Helper()
	var value struct {
		MinItems    int  `json:"minItems"`
		MaxItems    int  `json:"maxItems"`
		UniqueItems bool `json:"uniqueItems"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("schemas/education/colleges_of_health_sciences_and_technology.schema.json")), &value); err != nil {
		t.Fatal(err)
	}
	return value
}
func loadHealthCollegeReconciliation(t *testing.T) healthCollegeReconciliation {
	t.Helper()
	var value healthCollegeReconciliation
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/education/colleges_of_health_sciences_and_technology_reconciliation.json")), &value); err != nil {
		t.Fatal(err)
	}
	return value
}
func findHealthCollege(values []CollegeOfHealthSciencesAndTechnology, name string) (CollegeOfHealthSciencesAndTechnology, bool) {
	for _, value := range values {
		if value.Name == name {
			return value, true
		}
	}
	return CollegeOfHealthSciencesAndTechnology{}, false
}
func stateIDsForHealth(values []CollegeOfHealthSciencesAndTechnology) map[string]struct{} {
	result := map[string]struct{}{}
	for _, value := range values {
		result[value.StateID] = struct{}{}
	}
	return result
}
