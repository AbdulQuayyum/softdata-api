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

func TestNigeriaVocationalEnterpriseInstitutionsDataset(t *testing.T) {
	var values []VocationalEnterpriseInstitution
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("education/vocational_enterprise_institutions.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&values); err != nil {
		t.Fatal(err)
	}
	if len(values) != 25 {
		t.Fatalf("record count: got %d want 25", len(values))
	}
	states := map[string]struct{}{}
	for _, state := range loadStateDataset(t) {
		states[state.ID] = struct{}{}
	}
	ids, names, stateSet := map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}
	ordered := append([]VocationalEnterpriseInstitution(nil), values...)
	sort.SliceStable(ordered, func(i, j int) bool {
		a, b := strings.ToLower(ordered[i].Name), strings.ToLower(ordered[j].Name)
		if a == b {
			return ordered[i].ID < ordered[j].ID
		}
		return a < b
	})
	if !reflect.DeepEqual(values, ordered) {
		t.Fatal("dataset is not sorted by name then id")
	}
	for _, value := range values {
		if value.ID == "" || value.Name == "" || value.OwnershipType == "" || value.StateID == "" || value.CountryCode != "NG" {
			t.Fatalf("invalid record: %#v", value)
		}
		if !regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(value.ID) {
			t.Fatalf("invalid id: %q", value.ID)
		}
		if _, ok := states[value.StateID]; !ok {
			t.Fatalf("unknown state: %q", value.StateID)
		}
		if _, ok := ids[value.ID]; ok {
			t.Fatalf("duplicate id: %q", value.ID)
		}
		ids[value.ID] = struct{}{}
		if _, ok := names[value.Name]; ok {
			t.Fatalf("duplicate name: %q", value.Name)
		}
		names[value.Name] = struct{}{}
		stateSet[value.StateID] = struct{}{}
		if value.OwnershipType != "federal" && value.OwnershipType != "state" && value.OwnershipType != "private" {
			t.Fatalf("invalid ownership: %q", value.OwnershipType)
		}
	}
	if len(stateSet) != 11 {
		t.Fatalf("unexpected state coverage: %d", len(stateSet))
	}
	counts := map[string]int{"federal": 0, "state": 0, "private": 0}
	for _, value := range values {
		counts[value.OwnershipType]++
	}
	if !reflect.DeepEqual(counts, map[string]int{"federal": 0, "state": 4, "private": 21}) {
		t.Fatalf("ownership counts: %#v", counts)
	}
}

func TestNigeriaVocationalEnterpriseInstitutionsMetadataAndReconciliation(t *testing.T) {
	var metadata struct {
		DatasetKey      string         `json:"dataset_key"`
		RecordCount     int            `json:"record_count"`
		OwnershipCounts map[string]int `json:"ownership_counts"`
		StateCoverage   []string       `json:"state_coverage"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/education/vocational_enterprise_institutions.json")), &metadata); err != nil {
		t.Fatal(err)
	}
	if metadata.DatasetKey != "ng-vocational-enterprise-institutions" || metadata.RecordCount != 25 || len(metadata.StateCoverage) != 11 {
		t.Fatalf("metadata mismatch: %#v", metadata)
	}
	if !reflect.DeepEqual(metadata.OwnershipCounts, map[string]int{"federal": 0, "state": 4, "private": 21}) {
		t.Fatalf("ownership metadata: %#v", metadata.OwnershipCounts)
	}
	var schema struct {
		MinItems    int  `json:"minItems"`
		MaxItems    int  `json:"maxItems"`
		UniqueItems bool `json:"uniqueItems"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("schemas/education/vocational_enterprise_institutions.schema.json")), &schema); err != nil {
		t.Fatal(err)
	}
	if schema.MinItems != 25 || schema.MaxItems != 25 || !schema.UniqueItems {
		t.Fatalf("schema mismatch: %#v", schema)
	}
	var reconciliation struct {
		Arithmetic struct {
			Raw   int `json:"raw_vei_entries"`
			Final int `json:"final_active_veis"`
		} `json:"arithmetic"`
		Counts  map[string]int    `json:"classification_counts"`
		Records []json.RawMessage `json:"records"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/education/vocational_enterprise_institutions_reconciliation.json")), &reconciliation); err != nil {
		t.Fatal(err)
	}
	if reconciliation.Arithmetic.Raw != 25 || reconciliation.Arithmetic.Final != 25 || len(reconciliation.Records) != 25 || !reflect.DeepEqual(reconciliation.Counts, map[string]int{"retain_active_vei": 25}) {
		t.Fatalf("reconciliation mismatch: %#v", reconciliation)
	}
}
