package models

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestNHIAAccreditedHealthMaintenanceOrganisationJSONContract(t *testing.T) {
	record := NHIAAccreditedHealthMaintenanceOrganisation{
		ID:                  "example-hmo-limited-7",
		Name:                "Example HMO Limited",
		CountryCode:         "NG",
		OrganisationType:    "health_maintenance_organisation",
		AccreditationStatus: "accredited",
		HMOID:               "7",
	}

	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal NHIA accredited HMO: %v", err)
	}
	if string(encoded) != `{"id":"example-hmo-limited-7","name":"Example HMO Limited","country_code":"NG","organisation_type":"health_maintenance_organisation","accreditation_status":"accredited","hmo_id":"7"}` {
		t.Fatalf("unexpected JSON: %s", encoded)
	}
}

func TestNHIAAccreditedHealthMaintenanceOrganisationsDatasetAndReconciliation(t *testing.T) {
	var records []NHIAAccreditedHealthMaintenanceOrganisation
	decoder := json.NewDecoder(strings.NewReader(string(readTextBytes(t, datasetPath("healthcare/nhia_accredited_health_maintenance_organisations.json")))))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&records); err != nil {
		t.Fatalf("decode NHIA accredited HMOs: %v", err)
	}
	if len(records) != 94 {
		t.Fatalf("unexpected NHIA HMO count: got %d want 94", len(records))
	}

	idPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	seenIDs := make(map[string]struct{}, len(records))
	seenHMOIDs := make(map[string]struct{}, len(records))
	finalIDs := make(map[string]struct{}, len(records))
	for i, record := range records {
		if record.ID == "" || record.Name == "" || record.CountryCode != "NG" || record.OrganisationType != "health_maintenance_organisation" || record.AccreditationStatus != "accredited" || record.HMOID == "" {
			t.Fatalf("record %d has invalid required fields: %#v", i, record)
		}
		if len(record.ID) > NHIAAccreditedHealthMaintenanceOrganisationIDMaxLength || NHIAAccreditedHealthMaintenanceOrganisationIDMaxLength != 255 || !idPattern.MatchString(record.ID) {
			t.Fatalf("record %d has invalid id %q", i, record.ID)
		}
		if !regexp.MustCompile(`^[0-9]+$`).MatchString(record.HMOID) {
			t.Fatalf("record %d has invalid HMO ID %q", i, record.HMOID)
		}
		for field, value := range map[string]string{"id": record.ID, "name": record.Name, "country_code": record.CountryCode, "organisation_type": record.OrganisationType, "accreditation_status": record.AccreditationStatus, "hmo_id": record.HMOID} {
			if value == "" || strings.TrimSpace(value) != value {
				t.Fatalf("record %d field %s is empty or untrimmed", i, field)
			}
		}
		publicJSON, _ := json.Marshal(record)
		lowered := strings.ToLower(string(publicJSON))
		for _, forbidden := range []string{"website", "email", "phone", "director", "practitioner", "patient", "enrollee", "policy_number", "bank", "logo", "address"} {
			if strings.Contains(lowered, forbidden) {
				t.Fatalf("record %d exposes forbidden marker %q", i, forbidden)
			}
		}
		if _, ok := seenIDs[record.ID]; ok {
			t.Fatalf("duplicate public ID %q", record.ID)
		}
		if _, ok := seenHMOIDs[record.HMOID]; ok {
			t.Fatalf("duplicate HMO ID %q", record.HMOID)
		}
		seenIDs[record.ID] = struct{}{}
		seenHMOIDs[record.HMOID] = struct{}{}
		finalIDs[record.ID] = struct{}{}
	}

	ordered := append([]NHIAAccreditedHealthMaintenanceOrganisation(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := ordered[i], ordered[j]
		leftID, err := strconv.Atoi(left.HMOID)
		if err != nil {
			t.Fatalf("parse left HMO ID: %v", err)
		}
		rightID, err := strconv.Atoi(right.HMOID)
		if err != nil {
			t.Fatalf("parse right HMO ID: %v", err)
		}
		leftKey := fmt.Sprintf("%s\x00%03d\x00%s", strings.ToLower(left.Name), leftID, left.ID)
		rightKey := fmt.Sprintf("%s\x00%03d\x00%s", strings.ToLower(right.Name), rightID, right.ID)
		return leftKey < rightKey
	})
	for i := range records {
		if records[i].ID != ordered[i].ID {
			t.Fatal("NHIA HMO records are not deterministically sorted")
		}
	}

	var metadata struct {
		DatasetKey                string         `json:"dataset_key"`
		Status                    string         `json:"status"`
		Title                     string         `json:"title"`
		Description               string         `json:"description"`
		RecordCount               int            `json:"record_count"`
		SourceRows                int            `json:"source_rows"`
		DecisionCounts            map[string]int `json:"decision_counts"`
		OrganisationTypeCounts    map[string]int `json:"organisation_type_counts"`
		AccreditationStatusCounts map[string]int `json:"accreditation_status_counts"`
		Source                    struct {
			Publisher             string `json:"publisher"`
			URL                   string `json:"url"`
			APIURL                string `json:"api_url"`
			ExtractedTableSHA256  string `json:"extracted_table_sha256"`
			RawRecordCount        int    `json:"raw_record_count"`
			CategorySemantics     string `json:"category_semantics"`
			StatusSemantics       string `json:"status_semantics"`
			PersonalFieldsPresent any    `json:"personal_fields_present"`
		} `json:"source"`
		RejectedDatasetNames map[string]string `json:"rejected_dataset_names"`
		SourceLimitations    []string          `json:"source_limitations"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/healthcare/nhia_accredited_health_maintenance_organisations.json")), &metadata); err != nil {
		t.Fatalf("decode NHIA HMO metadata: %v", err)
	}
	if metadata.DatasetKey != "ng-nhia-accredited-health-maintenance-organisations" || metadata.Status != "active" || metadata.RecordCount != len(records) || metadata.SourceRows != 94 {
		t.Fatalf("NHIA HMO metadata mismatch: %#v", metadata)
	}
	if metadata.Title != "Nigeria NHIA Accredited Health Maintenance Organisations Snapshot" {
		t.Fatalf("metadata title mismatch: %q", metadata.Title)
	}
	for _, snippet := range []string{"dated organisation-level snapshot", "not a complete live health-insurance licensing or registration register"} {
		if !strings.Contains(metadata.Description, snippet) {
			t.Fatalf("metadata description missing %q: %q", snippet, metadata.Description)
		}
	}
	if metadata.DecisionCounts["retain"] != 94 || len(metadata.DecisionCounts) != 1 {
		t.Fatalf("decision counts mismatch: %#v", metadata.DecisionCounts)
	}
	if metadata.OrganisationTypeCounts["health_maintenance_organisation"] != 94 || metadata.AccreditationStatusCounts["accredited"] != 94 {
		t.Fatalf("classification counts mismatch: %#v %#v", metadata.OrganisationTypeCounts, metadata.AccreditationStatusCounts)
	}
	if metadata.Source.Publisher != "National Health Insurance Authority" || metadata.Source.URL != "https://www.nhia.gov.ng/hmo/" || metadata.Source.APIURL == "" || metadata.Source.ExtractedTableSHA256 == "" || metadata.Source.RawRecordCount != 94 {
		t.Fatalf("source metadata mismatch: %#v", metadata.Source)
	}
	if len(metadata.RejectedDatasetNames) != 3 {
		t.Fatalf("dataset name audit missing: %#v", metadata.RejectedDatasetNames)
	}
	limitations := strings.ToLower(strings.Join(metadata.SourceLimitations, " "))
	for _, snippet := range []string{"not a complete live", "state social health insurance agencies", "no enrollee", "no websites"} {
		if !strings.Contains(limitations, snippet) {
			t.Fatalf("source limitations missing %q: %v", snippet, metadata.SourceLimitations)
		}
	}

	var schema struct {
		Title       string `json:"title"`
		MinItems    int    `json:"minItems"`
		MaxItems    int    `json:"maxItems"`
		UniqueItems bool   `json:"uniqueItems"`
		Defs        map[string]struct {
			Required   []string `json:"required"`
			Properties map[string]struct {
				MaxLength int      `json:"maxLength"`
				Enum      []string `json:"enum"`
			} `json:"properties"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("schemas/healthcare/nhia_accredited_health_maintenance_organisations.schema.json")), &schema); err != nil {
		t.Fatalf("decode NHIA HMO schema: %v", err)
	}
	if schema.Title != "Nigeria NHIA Accredited Health Maintenance Organisations Snapshot" || schema.MinItems != len(records) || schema.MaxItems != len(records) || !schema.UniqueItems {
		t.Fatalf("schema mismatch: %#v", schema)
	}
	definition := schema.Defs["nhiaAccreditedHealthMaintenanceOrganisation"]
	expectedFields := []string{"id", "name", "country_code", "organisation_type", "accreditation_status", "hmo_id"}
	if !reflect.DeepEqual(definition.Required, expectedFields) || len(definition.Properties) != len(expectedFields) {
		t.Fatalf("schema field mismatch: %#v", definition)
	}
	if definition.Properties["id"].MaxLength != NHIAAccreditedHealthMaintenanceOrganisationIDMaxLength {
		t.Fatal("ID bounds disagree")
	}
	if !reflect.DeepEqual(definition.Properties["organisation_type"].Enum, []string{"health_maintenance_organisation"}) || !reflect.DeepEqual(definition.Properties["accreditation_status"].Enum, []string{"accredited"}) {
		t.Fatalf("schema enum mismatch: %#v", definition.Properties)
	}

	assertNHIAAccreditedHMOReconciliation(t, finalIDs)
}

func TestNHIAAccreditedHealthMaintenanceOrganisationPublicFieldContract(t *testing.T) {
	expected := map[string]bool{"id": true, "name": true, "country_code": true, "organisation_type": true, "accreditation_status": true, "hmo_id": true}
	model := reflect.TypeOf(NHIAAccreditedHealthMaintenanceOrganisation{})
	if model.NumField() != len(expected) {
		t.Fatal("unexpected public model fields")
	}
	for i := 0; i < model.NumField(); i++ {
		if !expected[strings.Split(model.Field(i).Tag.Get("json"), ",")[0]] {
			t.Fatal("unsupported public field")
		}
	}
}

func assertNHIAAccreditedHMOReconciliation(t *testing.T, finalIDs map[string]struct{}) {
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
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/healthcare/nhia_accredited_health_maintenance_organisations_reconciliation/index.json")), &index); err != nil {
		t.Fatalf("decode NHIA HMO reconciliation index: %v", err)
	}
	if index.DatasetKey != "ng-nhia-accredited-health-maintenance-organisations" || index.SourceRows != 94 || index.FinalRecordCount != 94 || index.DecisionCounts["retain"] != 94 || len(index.DecisionCounts) != 1 || len(index.Partitions) != 1 {
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
				SourcePageSection         string         `json:"source_page_section"`
				SourceName                string         `json:"source_name"`
				SourceHMOID               string         `json:"source_hmo_id"`
				SourceCategory            string         `json:"source_category"`
				SourceAccreditationStatus string         `json:"source_accreditation_status"`
				RawOrganisationLevelLoc   string         `json:"raw_organisation_level_location"`
				NormalizedName            string         `json:"normalized_name"`
				NormalizedHMOID           string         `json:"normalized_hmo_id"`
				Decision                  string         `json:"decision"`
				FinalID                   *string        `json:"final_record_id"`
				MergeTargetID             *string        `json:"merge_target_id"`
				Reason                    string         `json:"reason"`
				Evidence                  map[string]any `json:"evidence"`
			} `json:"records"`
		}
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatalf("decode reconciliation partition %s: %v", partition.Path, err)
		}
		if value.DatasetKey != "ng-nhia-accredited-health-maintenance-organisations" || value.SourceRows != len(value.Records) || value.SourceRows != partition.SourceRows {
			t.Fatalf("partition row count mismatch for %s", partition.Path)
		}
		for _, record := range value.Records {
			if record.SourcePosition == 0 || record.SourcePageSection == "" || record.SourceName == "" || record.SourceHMOID == "" || record.SourceCategory == "" || record.SourceAccreditationStatus == "" || record.NormalizedName == "" || record.NormalizedHMOID == "" || record.Reason == "" || len(record.Evidence) == 0 {
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
	if len(seenPositions) != 94 || len(reconciledIDs) != len(finalIDs) {
		t.Fatalf("reconciliation coverage mismatch: positions=%d reconciled=%d final=%d", len(seenPositions), len(reconciledIDs), len(finalIDs))
	}
}
