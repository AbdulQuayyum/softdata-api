package models

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestHealthFacilityJSONOmitsUnavailableOptionalFields(t *testing.T) {
	record := HealthFacility{
		ID:           "example-facility",
		Name:         "Example Facility",
		FacilityType: "clinic",
		StateID:      "lagos",
		CountryCode:  "NG",
	}

	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal health facility: %v", err)
	}
	if string(encoded) != `{"id":"example-facility","name":"Example Facility","facility_type":"clinic","state_id":"lagos","country_code":"NG"}` {
		t.Fatalf("unexpected JSON: %s", encoded)
	}
}

func TestNigeriaHealthFacilitiesDatasetAndReconciliation(t *testing.T) {
	var facilities []HealthFacility
	decoder := json.NewDecoder(strings.NewReader(string(readTextBytes(t, datasetPath("healthcare/health_facilities.json")))))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&facilities); err != nil {
		t.Fatalf("decode health facilities: %v", err)
	}
	if len(facilities) != 50654 {
		t.Fatalf("unexpected health facility count: got %d want 50654", len(facilities))
	}

	states := loadStateDataset(t)
	stateIDs := make(map[string]struct{}, len(states))
	for _, state := range states {
		stateIDs[state.ID] = struct{}{}
	}
	var lgas []struct {
		ID      string `json:"id"`
		StateID string `json:"state_id"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("geography/lgas.json")), &lgas); err != nil {
		t.Fatalf("decode LGAs: %v", err)
	}
	lgaState := make(map[string]string, len(lgas))
	for _, lga := range lgas {
		lgaState[lga.ID] = lga.StateID
	}

	idPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	allowedTypes := map[string]struct{}{
		"health-post": {}, "primary-health-centre": {}, "clinic": {},
		"general-hospital": {}, "specialist-hospital": {}, "teaching-hospital": {}, "other": {},
	}
	allowedLevels := map[string]struct{}{"primary": {}, "secondary": {}, "tertiary": {}}
	allowedOwnership := map[string]struct{}{
		"federal": {}, "state": {}, "local-government": {}, "private": {}, "military": {}, "other-public": {},
	}
	seenIDs := make(map[string]struct{}, len(facilities))
	seenSources := make(map[string]struct{}, len(facilities))
	for i, facility := range facilities {
		if facility.ID == "" || facility.Name == "" || facility.FacilityType == "" || facility.StateID == "" || facility.CountryCode != "NG" {
			t.Fatalf("record %d has invalid required fields: %#v", i, facility)
		}
		if !idPattern.MatchString(facility.ID) {
			t.Fatalf("record %d has invalid id %q", i, facility.ID)
		}
		if _, ok := stateIDs[facility.StateID]; !ok {
			t.Fatalf("record %d has unknown state %q", i, facility.StateID)
		}
		if _, ok := allowedTypes[facility.FacilityType]; !ok {
			t.Fatalf("record %d has unsupported facility type %q", i, facility.FacilityType)
		}
		if facility.FacilityLevel != "" {
			if _, ok := allowedLevels[facility.FacilityLevel]; !ok {
				t.Fatalf("record %d has unsupported facility level %q", i, facility.FacilityLevel)
			}
		}
		if facility.OwnershipType != "" {
			if _, ok := allowedOwnership[facility.OwnershipType]; !ok {
				t.Fatalf("record %d has unsupported ownership %q", i, facility.OwnershipType)
			}
		}
		if facility.LGAID != "" && lgaState[facility.LGAID] != facility.StateID {
			t.Fatalf("record %d has LGA %q outside state %q", i, facility.LGAID, facility.StateID)
		}
		if facility.SourceFacilityID == "" {
			t.Fatalf("record %d has no source facility ID", i)
		}
		if _, ok := seenIDs[facility.ID]; ok {
			t.Fatalf("duplicate public ID %q", facility.ID)
		}
		if _, ok := seenSources[facility.SourceFacilityID]; ok {
			t.Fatalf("duplicate source facility ID %q", facility.SourceFacilityID)
		}
		seenIDs[facility.ID] = struct{}{}
		seenSources[facility.SourceFacilityID] = struct{}{}
		if facility.Latitude == nil || facility.Longitude == nil {
			t.Fatalf("record %d has incomplete coordinates", i)
		}
		if *facility.Latitude < 4.281710 || *facility.Latitude > 13.865239 || *facility.Longitude < 2.707790 || *facility.Longitude > 14.636383 {
			t.Fatalf("record %d has out-of-source-bounds coordinates", i)
		}
	}

	ordered := append([]HealthFacility(nil), facilities...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := ordered[i], ordered[j]
		leftKey := fmt.Sprintf("%s\x00%s\x00%s\x00%s", left.StateID, left.LGAID, strings.ToLower(left.Name), left.ID)
		rightKey := fmt.Sprintf("%s\x00%s\x00%s\x00%s", right.StateID, right.LGAID, strings.ToLower(right.Name), right.ID)
		return leftKey < rightKey
	})
	if !equalHealthFacilities(facilities, ordered) {
		t.Fatal("health facilities are not deterministically sorted")
	}

	var metadata struct {
		DatasetKey     string         `json:"dataset_key"`
		Status         string         `json:"status"`
		RecordCount    int            `json:"record_count"`
		SourceRows     int            `json:"source_rows"`
		DecisionCounts map[string]int `json:"decision_counts"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/healthcare/health_facilities.json")), &metadata); err != nil {
		t.Fatalf("decode health metadata: %v", err)
	}
	if metadata.DatasetKey != "ng-health-facilities" || metadata.Status != "active" || metadata.RecordCount != len(facilities) || metadata.SourceRows != 51022 {
		t.Fatalf("health metadata mismatch: %#v", metadata)
	}
	if metadata.DecisionCounts["retain"] != 50654 || metadata.DecisionCounts["exclude_invalid_geography"] != 363 || metadata.DecisionCounts["merge_exact_duplicate"] != 5 {
		t.Fatalf("health decision counts mismatch: %#v", metadata.DecisionCounts)
	}

	var schema struct {
		MinItems    int  `json:"minItems"`
		MaxItems    int  `json:"maxItems"`
		UniqueItems bool `json:"uniqueItems"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("schemas/healthcare/health_facilities.schema.json")), &schema); err != nil {
		t.Fatalf("decode health schema: %v", err)
	}
	if schema.MinItems != len(facilities) || schema.MaxItems != len(facilities) || !schema.UniqueItems {
		t.Fatalf("health schema bounds mismatch: %#v", schema)
	}

	assertHealthFacilityReconciliation(t, len(facilities), metadata.DecisionCounts)
}

func assertHealthFacilityReconciliation(t *testing.T, wantRetained int, wantDecisions map[string]int) {
	t.Helper()
	var index struct {
		SourceRows       int            `json:"source_rows"`
		FinalRecordCount int            `json:"final_record_count"`
		DecisionCounts   map[string]int `json:"decision_counts"`
		Partitions       []struct {
			Path       string `json:"path"`
			SourceRows int    `json:"source_rows"`
			SizeBytes  int64  `json:"size_bytes"`
			SHA256     string `json:"sha256"`
		} `json:"partitions"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/healthcare/health_facilities_reconciliation/index.json")), &index); err != nil {
		t.Fatalf("decode health reconciliation index: %v", err)
	}
	if index.SourceRows != 51022 || index.FinalRecordCount != wantRetained || len(index.Partitions) != 37 {
		t.Fatalf("health reconciliation index mismatch: %#v", index)
	}
	if !equalStringIntMap(index.DecisionCounts, wantDecisions) {
		t.Fatalf("health reconciliation decision mismatch: got %#v want %#v", index.DecisionCounts, wantDecisions)
	}
	for _, partition := range index.Partitions {
		path := datasetPath(partition.Path)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read reconciliation partition %s: %v", partition.Path, err)
		}
		if int64(len(data)) != partition.SizeBytes {
			t.Fatalf("partition size mismatch for %s", partition.Path)
		}
		hash := fmt.Sprintf("%x", sha256.Sum256(data))
		if hash != partition.SHA256 {
			t.Fatalf("partition hash mismatch for %s", partition.Path)
		}
		var value struct {
			SourceRows int `json:"source_rows"`
			Records    []struct {
				Decision string  `json:"decision"`
				FinalID  *string `json:"final_record_id"`
			} `json:"records"`
		}
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatalf("decode reconciliation partition %s: %v", partition.Path, err)
		}
		if value.SourceRows != len(value.Records) || value.SourceRows != partition.SourceRows {
			t.Fatalf("partition row count mismatch for %s", partition.Path)
		}
		for _, record := range value.Records {
			if record.Decision == "retain" && record.FinalID == nil {
				t.Fatalf("retained reconciliation row lacks final ID in %s", partition.Path)
			}
			if record.Decision != "retain" && record.FinalID != nil {
				t.Fatalf("excluded/merged reconciliation row has final ID in %s", partition.Path)
			}
		}
	}
}

func equalHealthFacilities(left, right []HealthFacility) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i].ID != right[i].ID {
			return false
		}
	}
	return true
}

func equalStringIntMap(left, right map[string]int) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}
