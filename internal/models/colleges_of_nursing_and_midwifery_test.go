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
	DatasetKey        string         `json:"dataset_key"`
	Status            string         `json:"status"`
	PubliclyPublished bool           `json:"publicly_published"`
	RecordCount       int            `json:"record_count"`
	OwnershipCounts   map[string]int `json:"ownership_counts"`
	StateCoverage     int            `json:"state_coverage_count"`
}

func TestNigeriaNursingCollegeDataset(t *testing.T) {
	values := loadNursingCollegeDataset(t)
	if len(values) != 156 {
		t.Fatalf("unexpected record count: got %d want 156", len(values))
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
	if !reflect.DeepEqual(counts, map[string]int{"federal": 13, "state": 52, "private": 91}) {
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
	if metadata.DatasetKey != "ng-colleges-of-nursing-and-midwifery" || metadata.Status != "active" || !metadata.PubliclyPublished || metadata.RecordCount != 156 || metadata.StateCoverage != 32 {
		t.Fatalf("metadata mismatch: %#v", metadata)
	}
	if !reflect.DeepEqual(metadata.OwnershipCounts, map[string]int{"federal": 13, "state": 52, "private": 91}) {
		t.Fatalf("metadata ownership mismatch: %#v", metadata.OwnershipCounts)
	}
	var reconciliation struct {
		Status     string `json:"status"`
		SourceRows int    `json:"source_rows"`
		Published  int    `json:"published_records"`
		Unresolved int    `json:"unresolved_exclusions"`
		Records    []struct {
			Decision string  `json:"decision"`
			FinalID  *string `json:"final_record_id"`
		} `json:"records"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/education/colleges_of_nursing_and_midwifery_reconciliation.json")), &reconciliation); err != nil {
		t.Fatal(err)
	}
	if reconciliation.Status != "active" || reconciliation.SourceRows != 376 || reconciliation.Published != 156 || reconciliation.Unresolved != 220 || len(reconciliation.Records) != 376 {
		t.Fatalf("reconciliation mismatch: %#v", reconciliation)
	}
	retained := 0
	for _, record := range reconciliation.Records {
		if record.Decision != "retain" && record.Decision != "exclude_unresolved" {
			t.Fatalf("unsupported decision %q", record.Decision)
		}
		if record.Decision == "retain" {
			if record.FinalID == nil {
				t.Fatal("retained row has no final record")
			}
			retained++
		} else if record.FinalID != nil {
			t.Fatal("excluded row has a final record")
		}
	}
	if retained != 156 {
		t.Fatalf("retained rows: %d", retained)
	}
}

func TestNigeriaNursingCollegePublicRosterFingerprint(t *testing.T) {
	if len(loadNursingCollegeDataset(t)) != 156 {
		t.Fatal("snapshot fingerprint/count changed")
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
