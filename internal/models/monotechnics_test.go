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

type monotechnicMetadata struct {
	DatasetKey       string         `json:"dataset_key"`
	OfficialCategory string         `json:"official_category_name"`
	RecordCount      int            `json:"record_count"`
	OwnershipCounts  map[string]int `json:"ownership_counts"`
	StateCoverage    []string       `json:"state_coverage"`
}

func TestNigeriaMonotechnicsDatasetMatchesReconciledRoster(t *testing.T) {
	values := loadMonotechnicDataset(t)
	states := loadStateDataset(t)
	if len(values) != 86 {
		t.Fatalf("unexpected record count: got %d want 86", len(values))
	}
	stateIDs := make(map[string]struct{}, len(states))
	for _, state := range states {
		stateIDs[state.ID] = struct{}{}
	}
	seenIDs := make(map[string]struct{}, len(values))
	seenNames := make(map[string]struct{}, len(values))
	counts := map[string]int{}
	sorted := append([]Monotechnic(nil), values...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left, right := strings.ToLower(sorted[i].Name), strings.ToLower(sorted[j].Name)
		if left == right {
			return sorted[i].ID < sorted[j].ID
		}
		return left < right
	})
	if !reflect.DeepEqual(values, sorted) {
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
			t.Fatalf("unexpected field count for %q: %d", value.ID, len(raw))
		}
	}
	if !reflect.DeepEqual(counts, map[string]int{"federal": 32, "state": 4, "private": 50}) {
		t.Fatalf("unexpected ownership counts: %#v", counts)
	}
	if len(mapKeysForMonotechnics(values, "state_id")) != 27 {
		t.Fatal("unexpected state coverage")
	}
}

func TestNigeriaMonotechnicsMetadataSchemaAndReconciliation(t *testing.T) {
	metadata := loadMonotechnicMetadata(t)
	schema := loadMonotechnicSchema(t)
	reconciliation := loadMonotechnicReconciliation(t)
	if metadata.DatasetKey != "ng-monotechnics" || metadata.OfficialCategory != "Specialised Institutions (Monotechnics)" || metadata.RecordCount != 86 {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
	if !reflect.DeepEqual(metadata.OwnershipCounts, map[string]int{"federal": 32, "state": 4, "private": 50}) {
		t.Fatalf("unexpected ownership metadata: %#v", metadata.OwnershipCounts)
	}
	if schema.MinItems != 86 || schema.MaxItems != 86 || !schema.UniqueItems {
		t.Fatalf("schema does not match dataset: %#v", schema)
	}
	if reconciliation.RawRecordCount != 98 || reconciliation.FinalRecordCount != 86 || len(reconciliation.Records) != 98 {
		t.Fatalf("reconciliation arithmetic mismatch: %#v", reconciliation)
	}
	if reconciliation.ClassificationCounts["retain_active"] != 86 {
		t.Fatalf("unexpected retained count: %#v", reconciliation.ClassificationCounts)
	}
	if _, ok := findMonotechnic(valuesForMonotechnic(t), "Federal School of Surveying, Oyo"); !ok {
		t.Fatal("missing federal anchor")
	}
	if _, ok := findMonotechnic(valuesForMonotechnic(t), "International Institute of Tourism, Yenagoa"); !ok {
		t.Fatal("missing state anchor")
	}
	if _, ok := findMonotechnic(valuesForMonotechnic(t), "Charkin Maritime Academy, Port Harcourt"); !ok {
		t.Fatal("missing private anchor")
	}
}

type monotechnicReconciliation struct {
	RawRecordCount       int               `json:"raw_record_count"`
	FinalRecordCount     int               `json:"final_record_count"`
	ClassificationCounts map[string]int    `json:"classification_counts"`
	Records              []json.RawMessage `json:"records"`
}

func loadMonotechnicDataset(t *testing.T) []Monotechnic {
	t.Helper()
	var values []Monotechnic
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("education/monotechnics.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&values); err != nil {
		t.Fatal(err)
	}
	return values
}
func loadMonotechnicMetadata(t *testing.T) monotechnicMetadata {
	t.Helper()
	var value monotechnicMetadata
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("metadata/education/monotechnics.json"))))
	if err := dec.Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}
func loadMonotechnicSchema(t *testing.T) struct {
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
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("schemas/education/monotechnics.schema.json"))))
	if err := dec.Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}
func loadMonotechnicReconciliation(t *testing.T) monotechnicReconciliation {
	t.Helper()
	var value monotechnicReconciliation
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("metadata/education/monotechnics_reconciliation.json"))))
	if err := dec.Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}
func valuesForMonotechnic(t *testing.T) []Monotechnic { return loadMonotechnicDataset(t) }
func findMonotechnic(values []Monotechnic, name string) (Monotechnic, bool) {
	for _, value := range values {
		if value.Name == name {
			return value, true
		}
	}
	return Monotechnic{}, false
}
func mapKeysForMonotechnics(values []Monotechnic, field string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, value := range values {
		if field == "state_id" {
			result[value.StateID] = struct{}{}
		}
	}
	return result
}
