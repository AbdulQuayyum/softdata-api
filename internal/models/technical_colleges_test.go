package models

import (
	"bytes"
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestNigeriaTechnicalCollegeSnapshot(t *testing.T) {
	var values []TechnicalCollege
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("education/technical_colleges.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&values); err != nil {
		t.Fatal(err)
	}
	if len(values) != 115 {
		t.Fatalf("unexpected record count: %d", len(values))
	}
	ordered := append([]TechnicalCollege(nil), values...)
	sort.Slice(ordered, func(i, j int) bool {
		a, b := strings.ToLower(ordered[i].Name), strings.ToLower(ordered[j].Name)
		if a == b {
			return ordered[i].ID < ordered[j].ID
		}
		return a < b
	})
	if !equalTechnicalCollegeSlices(values, ordered) {
		t.Fatal("dataset is not deterministically ordered")
	}
	ids, names := map[string]bool{}, map[string]bool{}
	stateIDs := map[string]bool{}
	for _, state := range loadStateDataset(t) {
		stateIDs[state.ID] = true
	}
	for _, value := range values {
		if value.ID == "" || value.Name == "" || value.CountryCode != "NG" || !regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(value.ID) {
			t.Fatalf("invalid record: %#v", value)
		}
		if value.OwnershipType != "federal" && value.OwnershipType != "state" && value.OwnershipType != "private" {
			t.Fatalf("invalid ownership: %#v", value)
		}
		if !stateIDs[value.StateID] {
			t.Fatalf("unknown state: %q", value.StateID)
		}
		if ids[value.ID] || names[value.Name] {
			t.Fatalf("duplicate identity: %#v", value)
		}
		ids[value.ID], names[value.Name] = true, true
	}
}

func equalTechnicalCollegeSlices(a, b []TechnicalCollege) bool {
	return len(a) == len(b) && func() bool {
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}()
}
