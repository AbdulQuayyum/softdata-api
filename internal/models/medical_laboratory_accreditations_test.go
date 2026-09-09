package models

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestMedicalLaboratoryAccreditationJSONOmitsUnavailableOptionalFields(t *testing.T) {
	record := MedicalLaboratoryAccreditation{
		ID:                  "example-laboratory",
		Name:                "Example Laboratory",
		StateID:             "lagos",
		CountryCode:         "NG",
		AccreditationStatus: "accredited",
	}

	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal medical laboratory accreditation: %v", err)
	}
	if string(encoded) != `{"id":"example-laboratory","name":"Example Laboratory","state_id":"lagos","country_code":"NG","accreditation_status":"accredited"}` {
		t.Fatalf("unexpected JSON: %s", encoded)
	}
}

func TestNigeriaMedicalLaboratoryAccreditationsDatasetAndReconciliation(t *testing.T) {
	var records []MedicalLaboratoryAccreditation
	decoder := json.NewDecoder(strings.NewReader(string(readTextBytes(t, datasetPath("healthcare/medical_laboratory_accreditations.json")))))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&records); err != nil {
		t.Fatalf("decode medical laboratory accreditations: %v", err)
	}
	if len(records) != 30 {
		t.Fatalf("unexpected medical laboratory accreditation count: got %d want 30", len(records))
	}

	states := loadStateDataset(t)
	stateIDs := make(map[string]struct{}, len(states))
	for _, state := range states {
		stateIDs[state.ID] = struct{}{}
	}

	idPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	allowedStatus := map[string]struct{}{"accredited": {}, "expired": {}}
	seenIDs := make(map[string]struct{}, len(records))
	seenAccreditationNumbers := make(map[string]struct{}, len(records))
	statusCounts := make(map[string]int)
	finalIDs := make(map[string]struct{}, len(records))
	longestID := ""
	for i, record := range records {
		if record.ID == "" || record.Name == "" || record.StateID == "" || record.CountryCode != "NG" || record.AccreditationStatus == "" {
			t.Fatalf("record %d has invalid required fields: %#v", i, record)
		}
		if len(record.ID) > MedicalLaboratoryAccreditationIDMaxLength || !idPattern.MatchString(record.ID) {
			t.Fatalf("record %d has invalid id %q", i, record.ID)
		}
		if len(record.ID) > len(longestID) {
			longestID = record.ID
		}
		if _, ok := stateIDs[record.StateID]; !ok {
			t.Fatalf("record %d has unknown state %q", i, record.StateID)
		}
		if _, ok := allowedStatus[record.AccreditationStatus]; !ok {
			t.Fatalf("record %d has unsupported accreditation status %q", i, record.AccreditationStatus)
		}
		statusCounts[record.AccreditationStatus]++
		if record.AccreditationNumber == "" || !regexp.MustCompile(`^ML[0-9]{4}$`).MatchString(record.AccreditationNumber) {
			t.Fatalf("record %d has invalid accreditation number %q", i, record.AccreditationNumber)
		}
		for _, dateValue := range []string{record.ApprovalDate, record.ExpiryDate} {
			if _, err := time.Parse("2006-01-02", dateValue); err != nil {
				t.Fatalf("record %d has invalid date %q: %v", i, dateValue, err)
			}
		}
		if record.Address == "" {
			t.Fatalf("record %d has no public business address", i)
		}
		for field, value := range map[string]string{"name": record.Name, "address": record.Address, "accreditation_status": record.AccreditationStatus, "approval_date": record.ApprovalDate, "expiry_date": record.ExpiryDate} {
			if value == "" || strings.TrimSpace(value) != value {
				t.Fatalf("record %d field %s is empty or untrimmed", i, field)
			}
		}
		publicJSON, _ := json.Marshal(record)
		lowered := strings.ToLower(string(publicJSON))
		for _, forbidden := range []string{"scientist", "superintendent", "phone", "email", "personal", "practitioner", "lic" + "ence", "lic" + "ensed"} {
			if strings.Contains(lowered, forbidden) {
				t.Fatalf("record %d exposes forbidden active-contract marker %q", i, forbidden)
			}
		}
		if record.AccreditationStatus == "expired" && !strings.HasPrefix(record.ExpiryDate, "202") {
			t.Fatalf("expired record has unexpected expiry date: %#v", record)
		}
		if _, ok := seenIDs[record.ID]; ok {
			t.Fatalf("duplicate public ID %q", record.ID)
		}
		if _, ok := seenAccreditationNumbers[record.AccreditationNumber]; ok {
			t.Fatalf("duplicate accreditation number %q", record.AccreditationNumber)
		}
		seenIDs[record.ID] = struct{}{}
		seenAccreditationNumbers[record.AccreditationNumber] = struct{}{}
		finalIDs[record.ID] = struct{}{}
	}
	if statusCounts["accredited"] != 26 || statusCounts["expired"] != 4 || len(statusCounts) != 2 {
		t.Fatalf("accreditation status counts mismatch: %#v", statusCounts)
	}
	if len(longestID) != 163 {
		t.Fatalf("longest ID length changed: got %d id %q", len(longestID), longestID)
	}

	ordered := append([]MedicalLaboratoryAccreditation(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := ordered[i], ordered[j]
		leftKey := fmt.Sprintf("%s\x00%s\x00%s", left.StateID, strings.ToLower(left.Name), left.ID)
		rightKey := fmt.Sprintf("%s\x00%s\x00%s", right.StateID, strings.ToLower(right.Name), right.ID)
		return leftKey < rightKey
	})
	if !equalMedicalLaboratoryAccreditations(records, ordered) {
		t.Fatal("medical laboratory accreditations are not deterministically sorted")
	}

	var metadata struct {
		DatasetKey                string         `json:"dataset_key"`
		Status                    string         `json:"status"`
		Title                     string         `json:"title"`
		Description               string         `json:"description"`
		RecordCount               int            `json:"record_count"`
		SourceRows                int            `json:"source_rows"`
		DecisionCounts            map[string]int `json:"decision_counts"`
		AccreditationStatusCounts map[string]int `json:"accreditation_status_counts"`
		StateCoverageCount        int            `json:"state_coverage_count"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/healthcare/medical_laboratory_accreditations.json")), &metadata); err != nil {
		t.Fatalf("decode medical laboratory accreditation metadata: %v", err)
	}
	if metadata.DatasetKey != "ng-medical-laboratory-accreditations" || metadata.Status != "active" || metadata.RecordCount != len(records) || metadata.SourceRows != 30 {
		t.Fatalf("medical laboratory accreditation metadata mismatch: %#v", metadata)
	}
	if metadata.Title != "Nigeria Medical Laboratory Accreditation Register Snapshot" {
		t.Fatalf("metadata title mismatch: %q", metadata.Title)
	}
	if !strings.Contains(metadata.Description, "not a complete register") {
		t.Fatalf("metadata description does not document scope: %q", metadata.Description)
	}
	if metadata.DecisionCounts["retain"] != 30 || len(metadata.DecisionCounts) != 1 {
		t.Fatalf("decision counts mismatch: %#v", metadata.DecisionCounts)
	}
	if metadata.AccreditationStatusCounts["accredited"] != 26 || metadata.AccreditationStatusCounts["expired"] != 4 || len(metadata.AccreditationStatusCounts) != 2 {
		t.Fatalf("metadata status counts mismatch: %#v", metadata.AccreditationStatusCounts)
	}
	if metadata.StateCoverageCount != 11 {
		t.Fatalf("coverage metadata mismatch: %#v", metadata)
	}

	var schema struct {
		Title       string `json:"title"`
		MinItems    int    `json:"minItems"`
		MaxItems    int    `json:"maxItems"`
		UniqueItems bool   `json:"uniqueItems"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("schemas/healthcare/medical_laboratory_accreditations.schema.json")), &schema); err != nil {
		t.Fatalf("decode medical laboratory accreditation schema: %v", err)
	}
	if schema.Title != "Nigeria Medical Laboratory Accreditation Register Snapshot" || schema.MinItems != len(records) || schema.MaxItems != len(records) || !schema.UniqueItems {
		t.Fatalf("schema mismatch: %#v", schema)
	}

	assertMedicalLaboratoryAccreditationReconciliation(t, finalIDs)
	assertMedicalLaboratoryAccreditationBranchesRemainDistinct(t, records)
}

func assertMedicalLaboratoryAccreditationBranchesRemainDistinct(t *testing.T, records []MedicalLaboratoryAccreditation) {
	t.Helper()
	branches := map[string]MedicalLaboratoryAccreditation{}
	for _, record := range records {
		if record.AccreditationNumber == "ML0014" || record.AccreditationNumber == "ML0030" {
			branches[record.AccreditationNumber] = record
		}
	}
	if len(branches) != 2 {
		t.Fatalf("Everight branches missing: %#v", branches)
	}
	if branches["ML0014"].ID == branches["ML0030"].ID || branches["ML0014"].StateID == branches["ML0030"].StateID {
		t.Fatalf("Everight branches were not preserved as separate accreditation records: %#v", branches)
	}
}

func assertMedicalLaboratoryAccreditationReconciliation(t *testing.T, finalIDs map[string]struct{}) {
	t.Helper()
	var index struct {
		DatasetKey       string         `json:"dataset_key"`
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
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/healthcare/medical_laboratory_accreditations_reconciliation/index.json")), &index); err != nil {
		t.Fatalf("decode medical laboratory accreditation reconciliation index: %v", err)
	}
	if index.DatasetKey != "ng-medical-laboratory-accreditations" || index.SourceRows != 30 || index.FinalRecordCount != 30 || index.DecisionCounts["retain"] != 30 || len(index.DecisionCounts) != 1 || len(index.Partitions) != 11 {
		t.Fatalf("reconciliation index mismatch: %#v", index)
	}
	seenPositions := make(map[int]struct{}, index.SourceRows)
	reconciledIDs := make(map[string]struct{}, len(finalIDs))
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
			DatasetKey string `json:"dataset_key"`
			SourceRows int    `json:"source_rows"`
			Records    []struct {
				SourcePosition            int            `json:"source_position"`
				SourceName                string         `json:"source_name"`
				SourceAccreditationStatus string         `json:"source_accreditation_status"`
				SourceAccreditationNumber string         `json:"source_accreditation_number"`
				SourceApprovalDate        string         `json:"source_approval_date"`
				SourceExpiryDate          string         `json:"source_expiry_date"`
				RawState                  string         `json:"raw_state"`
				RawAddress                string         `json:"raw_address"`
				NormalizedName            string         `json:"normalized_name"`
				NormalizedStateID         string         `json:"normalized_state_id"`
				Decision                  string         `json:"decision"`
				FinalID                   *string        `json:"final_record_id"`
				MergeTargetID             *string        `json:"merge_target_id"`
				Evidence                  map[string]any `json:"evidence"`
			} `json:"records"`
		}
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatalf("decode reconciliation partition %s: %v", partition.Path, err)
		}
		if value.DatasetKey != "ng-medical-laboratory-accreditations" || value.SourceRows != len(value.Records) || value.SourceRows != partition.SourceRows {
			t.Fatalf("partition row count mismatch for %s", partition.Path)
		}
		for _, record := range value.Records {
			if record.SourceName == "" || record.SourceAccreditationStatus == "" || record.SourceAccreditationNumber == "" || record.SourceApprovalDate == "" || record.SourceExpiryDate == "" || record.RawState == "" || record.RawAddress == "" || record.NormalizedName == "" || record.NormalizedStateID == "" || len(record.Evidence) == 0 {
				t.Fatalf("incomplete reconciliation row in %s: %#v", partition.Path, record)
			}
			if _, ok := seenPositions[record.SourcePosition]; ok {
				t.Fatalf("duplicate source position %d", record.SourcePosition)
			}
			seenPositions[record.SourcePosition] = struct{}{}
			if record.Decision != "retain" || record.FinalID == nil || record.MergeTargetID != nil {
				t.Fatalf("unexpected reconciliation decision: %#v", record)
			}
			if _, ok := finalIDs[*record.FinalID]; !ok {
				t.Fatalf("retained reconciliation row points to missing public record %q", *record.FinalID)
			}
			reconciledIDs[*record.FinalID] = struct{}{}
		}
	}
	if len(seenPositions) != 30 || len(reconciledIDs) != len(finalIDs) {
		t.Fatalf("reconciliation coverage mismatch: positions=%d reconciled=%d final=%d", len(seenPositions), len(reconciledIDs), len(finalIDs))
	}
}

func equalMedicalLaboratoryAccreditations(left, right []MedicalLaboratoryAccreditation) bool {
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

func TestMedicalLaboratoryAccreditationPublicFieldContract(t *testing.T) {
	expected := map[string]bool{"id": true, "name": true, "state_id": true, "country_code": true, "accreditation_status": true, "accreditation_number": true, "approval_date": true, "expiry_date": true, "address": true}
	model := reflect.TypeOf(MedicalLaboratoryAccreditation{})
	if model.NumField() != len(expected) {
		t.Fatal("unexpected public model fields")
	}
	for i := 0; i < model.NumField(); i++ {
		if !expected[strings.Split(model.Field(i).Tag.Get("json"), ",")[0]] {
			t.Fatal("unsupported public field")
		}
	}
	var schema struct {
		Defs map[string]struct {
			Properties map[string]struct {
				MaxLength int `json:"maxLength"`
			} `json:"properties"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("schemas/healthcare/medical_laboratory_accreditations.schema.json")), &schema); err != nil {
		t.Fatal(err)
	}
	properties := schema.Defs["medicalLaboratoryAccreditation"].Properties
	if len(properties) != len(expected) {
		t.Fatal("unexpected schema fields")
	}
	for key := range properties {
		if !expected[key] {
			t.Fatal("unsupported schema field")
		}
	}
	if properties["id"].MaxLength != MedicalLaboratoryAccreditationIDMaxLength || MedicalLaboratoryAccreditationIDMaxLength != 255 {
		t.Fatal("ID bounds disagree")
	}
}
