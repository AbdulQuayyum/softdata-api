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

type nursingCollegeMetadata struct {
	DatasetKey      string         `json:"dataset_key"`
	RecordCount     int            `json:"record_count"`
	OwnershipCounts map[string]int `json:"ownership_counts"`
	StateCoverage   int            `json:"state_coverage_count"`
}

func TestNigeriaNursingCollegeDataset(t *testing.T) {
	values := loadNursingCollegeDataset(t)
	if len(values) != 176 {
		t.Fatalf("unexpected record count: got %d want 176", len(values))
	}
	states := loadStateDataset(t)
	stateIDs := map[string]struct{}{}
	for _, state := range states {
		stateIDs[state.ID] = struct{}{}
	}
	ordered := append([]CollegeOfNursingAndMidwifery(nil), values...)
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
	ids, names := map[string]struct{}{}, map[string]struct{}{}
	counts := map[string]int{}
	stateCount := map[string]struct{}{}
	for _, value := range values {
		if value.ID == "" || value.Name == "" || value.OwnershipType == "" || value.StateID == "" || value.CountryCode != "NG" {
			t.Fatalf("invalid record: %#v", value)
		}
		if !regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(value.ID) {
			t.Fatalf("invalid id %q", value.ID)
		}
		if _, ok := stateIDs[value.StateID]; !ok {
			t.Fatalf("unknown state %q", value.StateID)
		}
		if _, ok := ids[value.ID]; ok {
			t.Fatalf("duplicate id %q", value.ID)
		}
		if _, ok := names[value.Name]; ok {
			t.Fatalf("duplicate name %q", value.Name)
		}
		ids[value.ID], names[value.Name], stateCount[value.StateID] = struct{}{}, struct{}{}, struct{}{}
		counts[value.OwnershipType]++
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(encoded, &raw); err != nil {
			t.Fatal(err)
		}
		if len(raw) != 5 {
			t.Fatalf("unexpected public fields for %q", value.ID)
		}
	}
	if !reflect.DeepEqual(counts, map[string]int{"federal": 13, "state": 57, "private": 106}) {
		t.Fatalf("ownership mismatch: %#v", counts)
	}
	if len(stateCount) != 32 {
		t.Fatalf("state coverage mismatch: %d", len(stateCount))
	}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value.Name), "university") && !strings.Contains(strings.ToLower(value.Name), "teaching hospital") {
			t.Fatalf("university department leaked into roster: %q", value.Name)
		}
	}
}

func TestNigeriaNursingCollegeMetadataAndReconciliation(t *testing.T) {
	var metadata nursingCollegeMetadata
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/education/colleges_of_nursing_and_midwifery.json")), &metadata); err != nil {
		t.Fatal(err)
	}
	if metadata.DatasetKey != "ng-colleges-of-nursing-and-midwifery" || metadata.RecordCount != 176 || metadata.StateCoverage != 32 {
		t.Fatalf("metadata mismatch: %#v", metadata)
	}
	if !reflect.DeepEqual(metadata.OwnershipCounts, map[string]int{"federal": 13, "state": 57, "private": 106}) {
		t.Fatalf("metadata ownership mismatch: %#v", metadata.OwnershipCounts)
	}
	var reconciliation struct {
		SourceArithmetic struct {
			Raw    int `json:"final_active_institutions"`
			NMCN   int `json:"nmcn_raw_entries"`
			Merged int `json:"nmcn_programme_duplicates_or_subrows_merged"`
		} `json:"source_arithmetic"`
		Records []json.RawMessage `json:"records"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/education/colleges_of_nursing_and_midwifery_reconciliation.json")), &reconciliation); err != nil {
		t.Fatal(err)
	}
	if reconciliation.SourceArithmetic.Raw != 176 || len(reconciliation.Records) != 176 || reconciliation.SourceArithmetic.NMCN != 290 || reconciliation.SourceArithmetic.Merged != 114 {
		t.Fatalf("reconciliation mismatch: %#v", reconciliation)
	}
}

func loadNursingCollegeDataset(t *testing.T) []CollegeOfNursingAndMidwifery {
	t.Helper()
	var values []CollegeOfNursingAndMidwifery
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("education/colleges_of_nursing_and_midwifery.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&values); err != nil {
		t.Fatal(err)
	}
	return values
}
