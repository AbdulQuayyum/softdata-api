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

type polytechnicMetadata struct {
	DatasetKey      string            `json:"dataset_key"`
	RecordCount     int               `json:"record_count"`
	OwnershipCounts map[string]int    `json:"ownership_counts"`
	StateCoverage   []string          `json:"state_coverage"`
	Sources         []json.RawMessage `json:"sources"`
}

func TestNigeriaPolytechnicsDatasetMatchesReconciledRoster(t *testing.T) {
	polytechnics := loadPolytechnicDataset(t)
	states := loadStateDataset(t)
	if len(polytechnics) != 168 {
		t.Fatalf("unexpected record count: got %d want 168", len(polytechnics))
	}
	stateIDs := make(map[string]struct{}, len(states))
	for _, state := range states {
		stateIDs[state.ID] = struct{}{}
	}
	seenIDs := make(map[string]struct{}, len(polytechnics))
	seenNames := make(map[string]struct{}, len(polytechnics))
	counts := map[string]int{}
	ordered := append([]Polytechnic(nil), polytechnics...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := strings.ToLower(ordered[i].Name), strings.ToLower(ordered[j].Name)
		if left == right {
			return ordered[i].ID < ordered[j].ID
		}
		return left < right
	})
	if !reflect.DeepEqual(polytechnics, ordered) {
		t.Fatal("dataset is not sorted by name then id")
	}
	for i, poly := range polytechnics {
		if poly.ID == "" || poly.Name == "" || poly.OwnershipType == "" || poly.StateID == "" || poly.CountryCode != "NG" {
			t.Fatalf("invalid record %d: %#v", i, poly)
		}
		if !regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(poly.ID) {
			t.Fatalf("invalid id %q", poly.ID)
		}
		if _, ok := stateIDs[poly.StateID]; !ok {
			t.Fatalf("unknown state %q", poly.StateID)
		}
		if _, ok := seenIDs[poly.ID]; ok {
			t.Fatalf("duplicate id %q", poly.ID)
		}
		if _, ok := seenNames[poly.Name]; ok {
			t.Fatalf("duplicate name %q", poly.Name)
		}
		seenIDs[poly.ID], seenNames[poly.Name] = struct{}{}, struct{}{}
		counts[poly.OwnershipType]++
		var raw map[string]json.RawMessage
		encoded, err := json.Marshal(poly)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.NewDecoder(bytes.NewReader(encoded)).Decode(&raw); err != nil {
			t.Fatal(err)
		}
		if len(raw) != 5 {
			t.Fatalf("unexpected field count for %q: %d", poly.ID, len(raw))
		}
	}
	if !reflect.DeepEqual(counts, map[string]int{"federal": 35, "state": 43, "private": 90}) {
		t.Fatalf("unexpected ownership counts: %#v", counts)
	}
	if len(stateIDs) != 37 {
		t.Fatalf("unexpected state count: %d", len(stateIDs))
	}
}

func TestNigeriaPolytechnicsMetadataAndSchema(t *testing.T) {
	metadata := loadPolytechnicMetadata(t)
	schema := loadPolytechnicSchema(t)
	if metadata.DatasetKey != "ng-polytechnics" || metadata.RecordCount != 168 {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
	if !reflect.DeepEqual(metadata.OwnershipCounts, map[string]int{"federal": 35, "state": 43, "private": 90}) {
		t.Fatalf("unexpected metadata ownership counts: %#v", metadata.OwnershipCounts)
	}
	if len(metadata.StateCoverage) != 37 || len(metadata.Sources) == 0 {
		t.Fatalf("incomplete metadata coverage")
	}
	if schema.MinItems != 168 || schema.MaxItems != 168 || !schema.UniqueItems {
		t.Fatalf("schema count/uniqueness mismatch: %#v", schema)
	}
}

func loadPolytechnicDataset(t *testing.T) []Polytechnic {
	t.Helper()
	var values []Polytechnic
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("education/polytechnics.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&values); err != nil {
		t.Fatal(err)
	}
	return values
}
func loadPolytechnicMetadata(t *testing.T) polytechnicMetadata {
	t.Helper()
	var value polytechnicMetadata
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("metadata/education/polytechnics.json"))))
	if err := dec.Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}
func loadPolytechnicSchema(t *testing.T) struct {
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
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("schemas/education/polytechnics.schema.json"))))
	if err := dec.Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}
