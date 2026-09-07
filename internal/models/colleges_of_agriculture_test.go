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

type collegeOfAgricultureMetadata struct {
	DatasetKey       string         `json:"dataset_key"`
	OfficialCategory string         `json:"official_category_name"`
	RecordCount      int            `json:"record_count"`
	OwnershipCounts  map[string]int `json:"ownership_counts"`
	StateCoverage    []string       `json:"state_coverage"`
}

type collegeOfAgricultureReconciliation struct {
	RawRecordCount       int               `json:"raw_record_count"`
	FinalRecordCount     int               `json:"final_record_count"`
	ClassificationCounts map[string]int    `json:"classification_counts"`
	Records              []json.RawMessage `json:"records"`
}

func TestNigeriaCollegesOfAgricultureDatasetMatchesReconciledRoster(t *testing.T) {
	values := loadCollegeOfAgricultureDataset(t)
	states := loadStateDataset(t)
	polytechnics := loadPolytechnicDataset(t)
	monotechnics := loadMonotechnicDataset(t)
	if len(values) != 31 {
		t.Fatalf("unexpected record count: got %d want 31", len(values))
	}
	stateIDs := make(map[string]struct{}, len(states))
	for _, state := range states {
		stateIDs[state.ID] = struct{}{}
	}
	seenIDs, seenNames := map[string]struct{}{}, map[string]struct{}{}
	counts := map[string]int{}
	ordered := append([]CollegeOfAgriculture(nil), values...)
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
		counts[value.OwnershipType]++
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
	if !reflect.DeepEqual(counts, map[string]int{"federal": 23, "state": 7, "private": 1}) {
		t.Fatalf("unexpected ownership counts: %#v", counts)
	}
	polyNames, monoNames := map[string]struct{}{}, map[string]struct{}{}
	for _, value := range polytechnics {
		polyNames[value.Name] = struct{}{}
	}
	for _, value := range monotechnics {
		monoNames[value.Name] = struct{}{}
	}
	for _, value := range values {
		if _, ok := polyNames[value.Name]; ok {
			t.Fatalf("unexpected polytechnic overlap: %q", value.Name)
		}
		if _, ok := monoNames[value.Name]; ok {
			t.Fatalf("unexpected monotechnic overlap: %q", value.Name)
		}
	}
	if len(newStringSet(values)) != 16 {
		t.Fatalf("unexpected state coverage")
	}
}

func TestNigeriaCollegesOfAgricultureMetadataSchemaAndReconciliation(t *testing.T) {
	metadata := loadCollegeOfAgricultureMetadata(t)
	schema := loadCollegeOfAgricultureSchema(t)
	reconciliation := loadCollegeOfAgricultureReconciliation(t)
	if metadata.DatasetKey != "ng-colleges-of-agriculture" || metadata.OfficialCategory != "Colleges of Agriculture and Related Disciplines" || metadata.RecordCount != 31 {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
	if !reflect.DeepEqual(metadata.OwnershipCounts, map[string]int{"federal": 23, "state": 7, "private": 1}) {
		t.Fatalf("unexpected ownership metadata: %#v", metadata.OwnershipCounts)
	}
	if schema.MinItems != 31 || schema.MaxItems != 31 || !schema.UniqueItems {
		t.Fatalf("schema does not match dataset")
	}
	if reconciliation.RawRecordCount != 32 || reconciliation.FinalRecordCount != 31 || len(reconciliation.Records) != 32 || reconciliation.ClassificationCounts["exclude_converted"] != 1 {
		t.Fatalf("reconciliation mismatch: %#v", reconciliation)
	}
	values := loadCollegeOfAgricultureDataset(t)
	for _, anchor := range []string{"Federal College of Agriculture, Akure", "Bauchi State College of Agriculture, Bauchi", "Usteem College of Agriculture, Osogbo"} {
		if _, ok := findCollegeOfAgriculture(values, anchor); !ok {
			t.Fatalf("missing anchor %q", anchor)
		}
	}
}

func loadCollegeOfAgricultureDataset(t *testing.T) []CollegeOfAgriculture {
	t.Helper()
	var values []CollegeOfAgriculture
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("education/colleges_of_agriculture.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&values); err != nil {
		t.Fatal(err)
	}
	return values
}
func loadCollegeOfAgricultureMetadata(t *testing.T) collegeOfAgricultureMetadata {
	t.Helper()
	var value collegeOfAgricultureMetadata
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/education/colleges_of_agriculture.json")), &value); err != nil {
		t.Fatal(err)
	}
	return value
}
func loadCollegeOfAgricultureSchema(t *testing.T) struct {
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
	if err := json.Unmarshal(readTextBytes(t, datasetPath("schemas/education/colleges_of_agriculture.schema.json")), &value); err != nil {
		t.Fatal(err)
	}
	return value
}
func loadCollegeOfAgricultureReconciliation(t *testing.T) collegeOfAgricultureReconciliation {
	t.Helper()
	var value collegeOfAgricultureReconciliation
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/education/colleges_of_agriculture_reconciliation.json")), &value); err != nil {
		t.Fatal(err)
	}
	return value
}
func findCollegeOfAgriculture(values []CollegeOfAgriculture, name string) (CollegeOfAgriculture, bool) {
	for _, value := range values {
		if value.Name == name {
			return value, true
		}
	}
	return CollegeOfAgriculture{}, false
}
func newStringSet(values []CollegeOfAgriculture) map[string]struct{} {
	result := map[string]struct{}{}
	for _, value := range values {
		result[value.StateID] = struct{}{}
	}
	return result
}
