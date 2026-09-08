package models

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type nursingCollegeMetadata struct {
	DatasetKey      string         `json:"dataset_key"`
	Status          string         `json:"status"`
	PubliclyPublished bool         `json:"publicly_published"`
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
	if metadata.DatasetKey != "ng-colleges-of-nursing-and-midwifery" || metadata.Status != "draft" || metadata.PubliclyPublished || metadata.RecordCount != 176 || metadata.StateCoverage != 32 {
		t.Fatalf("metadata mismatch: %#v", metadata)
	}
	if !reflect.DeepEqual(metadata.OwnershipCounts, map[string]int{"federal": 13, "state": 57, "private": 106}) {
		t.Fatalf("metadata ownership mismatch: %#v", metadata.OwnershipCounts)
	}
	var reconciliation struct {
		Status              string `json:"status"`
		ReconciliationStatus string `json:"reconciliation_status"`
		SourceLedgerIsExact bool   `json:"source_ledger_is_exact"`
		PDFStructure struct {
			NumberedRows int `json:"numbered_table_rows"`
			Sections     int `json:"section_headings"`
			Markers      int `json:"programme_markers"`
			Labels       int `json:"unique_printed_row_labels"`
		} `json:"pdf_structure"`
		SourceArithmetic struct {
			Raw    int `json:"current_provisional_public_records"`
			NMCN   int `json:"nmcn_webpage_reported_training_institutions"`
			Merged int `json:"programme_merges"`
		} `json:"source_arithmetic"`
		FinalClassifications map[string]int `json:"final_classification_counts"`
		Records              []struct {
			SourcePosition           int    `json:"source_position"`
			CanonicalInstitutionID   string `json:"canonical_institution_id"`
			Decision                 string `json:"decision"`
			CreatesPublicRecord      bool   `json:"creates_public_record"`
			MergedIntoSourcePosition *int   `json:"merged_into_source_position"`
		} `json:"records"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/education/colleges_of_nursing_and_midwifery_reconciliation.json")), &reconciliation); err != nil {
		t.Fatal(err)
	}
	if reconciliation.Status != "draft" || reconciliation.ReconciliationStatus != "provisional" || reconciliation.SourceLedgerIsExact || reconciliation.PDFStructure.NumberedRows != 340 || reconciliation.PDFStructure.Sections != 36 || reconciliation.PDFStructure.Markers != 494 || reconciliation.PDFStructure.Labels != 329 || reconciliation.SourceArithmetic.Raw != 176 || len(reconciliation.Records) != 290 || reconciliation.SourceArithmetic.NMCN != 290 || reconciliation.SourceArithmetic.Merged != 114 {
		t.Fatalf("reconciliation mismatch: raw=%d records=%d nmcn=%d merges=%d", reconciliation.SourceArithmetic.Raw, len(reconciliation.Records), reconciliation.SourceArithmetic.NMCN, reconciliation.SourceArithmetic.Merged)
	}
	if !reflect.DeepEqual(reconciliation.FinalClassifications, map[string]int{"retained_as_institution": 176, "merged_programme_under_institution": 114}) {
		t.Fatalf("unexpected reconciliation classifications: %#v", reconciliation.FinalClassifications)
	}
	targets := map[string]struct{}{}
	for _, value := range loadNursingCollegeDataset(t) {
		targets[value.ID] = struct{}{}
	}
	positions := map[int]struct{}{}
	retained := map[int]struct{}{}
	for _, record := range reconciliation.Records {
		if record.SourcePosition < 1 || record.SourcePosition > 290 {
			t.Fatalf("invalid source position: %d", record.SourcePosition)
		}
		if _, ok := positions[record.SourcePosition]; ok {
			t.Fatalf("duplicate source position: %d", record.SourcePosition)
		}
		positions[record.SourcePosition] = struct{}{}
		if _, ok := targets[record.CanonicalInstitutionID]; !ok {
			t.Fatalf("unknown target id: %q", record.CanonicalInstitutionID)
		}
		if record.Decision == "retained_as_institution" {
			if !record.CreatesPublicRecord {
				t.Fatalf("retained source entry does not create a public record: %d", record.SourcePosition)
			}
			retained[record.SourcePosition] = struct{}{}
		} else if record.Decision == "merged_programme_under_institution" {
			if record.CreatesPublicRecord || record.MergedIntoSourcePosition == nil {
				t.Fatalf("invalid programme merge: %d", record.SourcePosition)
			}
			if _, ok := retained[*record.MergedIntoSourcePosition]; !ok {
				t.Fatalf("merge target is not retained: %d -> %d", record.SourcePosition, *record.MergedIntoSourcePosition)
			}
		} else {
			t.Fatalf("unsupported decision %q", record.Decision)
		}
	}
	if len(positions) != 290 || len(retained) != 176 {
		t.Fatalf("incomplete source ledger: positions=%d retained=%d", len(positions), len(retained))
	}
	if len(targets) != 176 {
		t.Fatalf("unexpected public target count: %d", len(targets))
	}
	for position := 1; position <= 290; position++ {
		if _, ok := positions[position]; !ok {
			t.Fatalf("missing source position: %d", position)
		}
	}
	for target := range targets {
		found := false
		for _, record := range reconciliation.Records {
			if record.CanonicalInstitutionID == target {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("public target is not referenced: %q", target)
		}
	}
}

func TestNigeriaNursingCollegePublicRosterFingerprint(t *testing.T) {
	data := readTextBytes(t, datasetPath("education/colleges_of_nursing_and_midwifery.json"))
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != "1f6e6567998ff864f3bfa0a6a5a1992a4a15dafd916518fc4a9522828512985b" {
		t.Fatalf("public roster fingerprint changed: %s", got)
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
