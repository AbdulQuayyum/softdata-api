package models

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type primaryAndSecondaryMetadata struct {
	DatasetKey      string         `json:"dataset_key"`
	RecordCount     int            `json:"record_count"`
	OwnershipCounts map[string]int `json:"ownership_counts"`
	StateCoverage   []string       `json:"state_coverage"`
	LGAcoverage     int            `json:"lga_coverage"`
}

type primaryAndSecondaryPartition struct {
	StateID   string                        `json:"state_id"`
	Decisions []primaryAndSecondaryDecision `json:"decisions"`
}

type primaryAndSecondaryDecision struct {
	Source       string `json:"source"`
	SourceRow    int    `json:"source_row"`
	StateID      string `json:"state_id"`
	CanonicalID  string `json:"canonical_id"`
	PublicRecord bool   `json:"public_record"`
	Decision     string `json:"decision"`
}

type primaryAndSecondaryIndex struct {
	Partitions []struct {
		StateID string `json:"state_id"`
		Path    string `json:"path"`
		SHA256  string `json:"sha256"`
	} `json:"partitions"`
}

type primaryAndSecondarySchema struct {
	MinItems    int  `json:"minItems"`
	MaxItems    int  `json:"maxItems"`
	UniqueItems bool `json:"uniqueItems"`
}

func TestNigeriaPrimaryAndSecondarySchoolsDataset(t *testing.T) {
	var values []PrimaryAndSecondarySchool
	decoder := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("education/primary_and_secondary_schools.json"))))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&values); err != nil {
		t.Fatal(err)
	}
	if len(values) != 166604 {
		t.Fatalf("unexpected record count: got %d want 166604", len(values))
	}
	states := loadStateDataset(t)
	lgas := loadLocalGovernmentUnitDataset(t)
	stateIDs := map[string]bool{}
	lgaByID := map[string]LocalGovernmentUnit{}
	for _, state := range states {
		stateIDs[state.ID] = true
	}
	for _, lga := range lgas {
		lgaByID[lga.ID] = lga
	}
	seenIDs := map[string]bool{}
	seenIdentity := map[string]bool{}
	counts := map[string]int{}
	levelCounts := map[string]int{}
	ordered := append([]PrimaryAndSecondarySchool(nil), values...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := strings.ToLower(ordered[i].Name), strings.ToLower(ordered[j].Name)
		if left == right {
			return ordered[i].ID < ordered[j].ID
		}
		return left < right
	})
	if !equalPrimaryAndSecondary(ordered, values) {
		t.Fatal("dataset is not sorted by name then id")
	}
	validID := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	validLevel := map[string]bool{"pre-primary": true, "primary": true, "junior-secondary": true, "senior-secondary": true}
	for i, value := range values {
		if value.ID == "" || value.Name == "" || value.StateID == "" || value.CountryCode != "NG" || len(value.EducationLevels) == 0 {
			t.Fatalf("invalid record %d: %#v", i, value)
		}
		if !validID.MatchString(value.ID) {
			t.Fatalf("invalid id %q", value.ID)
		}
		if !stateIDs[value.StateID] {
			t.Fatalf("unknown state %q", value.StateID)
		}
		if value.LGAID == "" || lgaByID[value.LGAID].StateID != value.StateID {
			t.Fatalf("invalid LGA for %q: %q", value.ID, value.LGAID)
		}
		if seenIDs[value.ID] {
			t.Fatalf("duplicate id %q", value.ID)
		}
		seenIDs[value.ID] = true
		identity := strings.ToLower(value.Name) + "|" + value.StateID + "|" + value.LGAID + "|" + value.ID
		if seenIdentity[identity] {
			t.Fatalf("duplicate identity %q", identity)
		}
		seenIdentity[identity] = true
		if value.OwnershipType != "public" && value.OwnershipType != "private" {
			t.Fatalf("invalid ownership %q", value.OwnershipType)
		}
		counts[value.OwnershipType]++
		for _, level := range value.EducationLevels {
			if !validLevel[level] {
				t.Fatalf("invalid education level %q", level)
			}
			levelCounts[level]++
		}
	}
	if counts["public"] != 81160 || counts["private"] != 85444 {
		t.Fatalf("unexpected ownership counts: %#v", counts)
	}
	if len(mapStateIDs(values)) != 37 || len(mapLGAIDs(values)) != 772 {
		t.Fatalf("unexpected geography coverage")
	}
	if levelCounts["pre-primary"] != 92483 || levelCounts["primary"] != 131078 || levelCounts["junior-secondary"] != 38112 || levelCounts["senior-secondary"] != 20261 {
		t.Fatalf("unexpected level counts: %#v", levelCounts)
	}
}

func TestNigeriaPrimaryAndSecondarySchoolsReconciliation(t *testing.T) {
	directory := datasetPath("metadata/education/primary_and_secondary_schools_reconciliation")
	var index primaryAndSecondaryIndex
	if err := json.Unmarshal(readTextBytes(t, filepath.Join(directory, "index.json")), &index); err != nil {
		t.Fatal(err)
	}
	if len(index.Partitions) != 37 {
		t.Fatalf("unexpected partition count: %d", len(index.Partitions))
	}
	seenRows := map[string]bool{}
	seenIDs := map[string]bool{}
	counts := map[string]int{}
	rows := 0
	for _, entry := range index.Partitions {
		path := filepath.Join(directory, entry.Path)
		contents := readTextBytes(t, path)
		if fmt.Sprintf("%x", sha256.Sum256(contents)) != entry.SHA256 {
			t.Fatalf("partition hash mismatch: %s", entry.Path)
		}
		var partition primaryAndSecondaryPartition
		if err := json.Unmarshal(contents, &partition); err != nil {
			t.Fatal(err)
		}
		if partition.StateID != entry.StateID {
			t.Fatalf("partition state mismatch: %s", entry.Path)
		}
		for _, decision := range partition.Decisions {
			key := decision.Source + ":" + fmt.Sprint(decision.SourceRow)
			if seenRows[key] {
				t.Fatalf("duplicated source decision: %s", key)
			}
			seenRows[key] = true
			rows++
			counts[decision.Decision]++
			if decision.PublicRecord {
				seenIDs[decision.CanonicalID] = true
			}
		}
	}
	if rows != 171027 || len(seenRows) != 171027 || len(seenIDs) != 166604 {
		t.Fatalf("unexpected partition totals: rows=%d ids=%d", rows, len(seenIDs))
	}
	if counts["excluded_unresolved_geography"] != 69 || counts["retained_source_identity"] != 162331 || counts["merged_cross_workbook_or_duplicate_row"] != 1910 || counts["merged_identity_code_conflict"] != 6717 {
		t.Fatalf("unexpected decision arithmetic: %#v", counts)
	}
	var dataset []PrimaryAndSecondarySchool
	decoder := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("education/primary_and_secondary_schools.json"))))
	if err := decoder.Decode(&dataset); err != nil {
		t.Fatal(err)
	}
	if len(dataset) != len(seenIDs) {
		t.Fatalf("dataset/reconciliation mismatch")
	}
}

func TestNigeriaPrimaryAndSecondarySchoolsContractsAgree(t *testing.T) {
	var metadata primaryAndSecondaryMetadata
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/education/primary_and_secondary_schools.json")), &metadata); err != nil {
		t.Fatal(err)
	}
	var schema primaryAndSecondarySchema
	if err := json.Unmarshal(readTextBytes(t, datasetPath("schemas/education/primary_and_secondary_schools.schema.json")), &schema); err != nil {
		t.Fatal(err)
	}
	if metadata.DatasetKey != "ng-primary-and-secondary-schools" || metadata.RecordCount != 166604 {
		t.Fatalf("unexpected metadata identity/count: %#v", metadata)
	}
	if schema.MinItems != 166604 || schema.MaxItems != 166604 || !schema.UniqueItems {
		t.Fatalf("unexpected schema bounds: %#v", schema)
	}
	if metadata.OwnershipCounts["public"] != 81160 || metadata.OwnershipCounts["private"] != 85444 || len(metadata.StateCoverage) != 37 || metadata.LGAcoverage != 772 {
		t.Fatalf("unexpected metadata coverage: %#v", metadata)
	}
}

func equalPrimaryAndSecondary(left, right []PrimaryAndSecondarySchool) bool {
	return strings.EqualFold(string(mustJSON(left)), string(mustJSON(right)))
}
func mustJSON(value any) []byte { encoded, _ := json.Marshal(value); return encoded }
func mapStateIDs(values []PrimaryAndSecondarySchool) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		result[value.StateID] = true
	}
	return result
}
func mapLGAIDs(values []PrimaryAndSecondarySchool) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		result[value.LGAID] = true
	}
	return result
}
