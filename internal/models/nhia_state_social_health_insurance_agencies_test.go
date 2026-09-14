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
)

func TestNHIAStateSocialHealthInsuranceAgencyJSONContract(t *testing.T) {
	record := NHIAStateSocialHealthInsuranceAgency{
		ID:               "abia-state-health-insurance-agency",
		Name:             "Abia State Health Insurance Agency",
		StateID:          "abia",
		CountryCode:      "NG",
		OrganisationType: "state_social_health_insurance_agency",
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal NHIA SSHIA: %v", err)
	}
	if string(encoded) != `{"id":"abia-state-health-insurance-agency","name":"Abia State Health Insurance Agency","state_id":"abia","country_code":"NG","organisation_type":"state_social_health_insurance_agency"}` {
		t.Fatalf("unexpected JSON: %s", encoded)
	}
}

func TestNHIAStateSocialHealthInsuranceAgenciesDatasetAndReconciliation(t *testing.T) {
	var records []NHIAStateSocialHealthInsuranceAgency
	decoder := json.NewDecoder(strings.NewReader(string(readTextBytes(t, datasetPath("healthcare/nhia_state_social_health_insurance_agencies.json")))))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&records); err != nil {
		t.Fatalf("decode NHIA SSHIAs: %v", err)
	}
	if len(records) != 37 {
		t.Fatalf("unexpected NHIA SSHIA count: got %d want 37", len(records))
	}

	stateIDs := canonicalStateIDs(t)
	idPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	seenIDs := make(map[string]struct{}, len(records))
	seenStates := make(map[string]struct{}, len(records))
	finalIDs := make(map[string]struct{}, len(records))
	for i, record := range records {
		if record.ID == "" || record.Name == "" || record.StateID == "" || record.CountryCode != "NG" || record.OrganisationType != "state_social_health_insurance_agency" {
			t.Fatalf("record %d has invalid required fields: %#v", i, record)
		}
		if len(record.ID) > NHIAStateSocialHealthInsuranceAgencyIDMaxLength || NHIAStateSocialHealthInsuranceAgencyIDMaxLength != 255 || !idPattern.MatchString(record.ID) {
			t.Fatalf("record %d has invalid id %q", i, record.ID)
		}
		if _, ok := stateIDs[record.StateID]; !ok {
			t.Fatalf("record %d has unknown state %q", i, record.StateID)
		}
		for field, value := range map[string]string{"id": record.ID, "name": record.Name, "state_id": record.StateID, "country_code": record.CountryCode, "organisation_type": record.OrganisationType} {
			if value == "" || strings.TrimSpace(value) != value {
				t.Fatalf("record %d field %s is empty or untrimmed", i, field)
			}
		}
		if _, ok := seenIDs[record.ID]; ok {
			t.Fatalf("duplicate public ID %q", record.ID)
		}
		if _, ok := seenStates[record.StateID]; ok {
			t.Fatalf("duplicate state ID %q", record.StateID)
		}
		seenIDs[record.ID] = struct{}{}
		seenStates[record.StateID] = struct{}{}
		finalIDs[record.ID] = struct{}{}
	}
	if len(seenStates) != len(stateIDs) {
		t.Fatalf("state coverage mismatch: got %d want %d", len(seenStates), len(stateIDs))
	}
	ordered := append([]NHIAStateSocialHealthInsuranceAgency(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].StateID < ordered[j].StateID })
	for i := range records {
		if records[i].ID != ordered[i].ID {
			t.Fatal("NHIA SSHIA records are not deterministically sorted")
		}
	}

	assertNHIAStateSocialHealthInsuranceAgencyMetadataAndSchema(t, len(records))
	assertNHIAStateSocialHealthInsuranceAgencyReconciliation(t, finalIDs, stateIDs)
	assertNHIAStateSocialHealthInsuranceAgencyPrivacy(t)
}

func TestNHIAStateSocialHealthInsuranceAgencyPublicFieldContract(t *testing.T) {
	expected := map[string]bool{"id": true, "name": true, "state_id": true, "country_code": true, "organisation_type": true}
	model := reflect.TypeOf(NHIAStateSocialHealthInsuranceAgency{})
	if model.NumField() != len(expected) {
		t.Fatal("unexpected public model fields")
	}
	for i := 0; i < model.NumField(); i++ {
		if !expected[strings.Split(model.Field(i).Tag.Get("json"), ",")[0]] {
			t.Fatal("unsupported public field")
		}
	}
}

func assertNHIAStateSocialHealthInsuranceAgencyMetadataAndSchema(t *testing.T, recordCount int) {
	t.Helper()
	var metadata struct {
		DatasetKey             string         `json:"dataset_key"`
		Title                  string         `json:"title"`
		Group                  string         `json:"group"`
		CountryCode            string         `json:"country_code"`
		RecordCount            int            `json:"record_count"`
		SourceRows             int            `json:"source_rows"`
		DecisionCounts         map[string]int `json:"decision_counts"`
		OrganisationTypeCounts map[string]int `json:"organisation_type_counts"`
		Source                 struct {
			URL                       string   `json:"url"`
			APIURL                    string   `json:"api_url"`
			RawRecordCount            int      `json:"raw_record_count"`
			NormalizedSafeTableSHA256 string   `json:"normalized_safe_table_sha256"`
			DiscardedSourceColumns    []string `json:"discarded_source_columns"`
			RawHTMLHashNote           string   `json:"raw_html_hash_note"`
		} `json:"source"`
		SourceLimitations []string `json:"source_limitations"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/healthcare/nhia_state_social_health_insurance_agencies.json")), &metadata); err != nil {
		t.Fatalf("decode NHIA SSHIA metadata: %v", err)
	}
	if metadata.DatasetKey != "ng-nhia-state-social-health-insurance-agencies" || metadata.Title != "Nigeria NHIA-Listed State Social Health Insurance Agencies Snapshot" || metadata.Group != "healthcare" || metadata.CountryCode != "NG" || metadata.RecordCount != recordCount || metadata.SourceRows != 37 {
		t.Fatalf("metadata mismatch: %#v", metadata)
	}
	if metadata.DecisionCounts["retain"] != 37 || len(metadata.DecisionCounts) != 1 || metadata.OrganisationTypeCounts["state_social_health_insurance_agency"] != 37 {
		t.Fatalf("metadata counts mismatch: %#v", metadata)
	}
	if metadata.Source.URL != "https://www.nhia.gov.ng/sshias/" || metadata.Source.APIURL != "https://www.nhia.gov.ng/wp-json/wp/v2/pages/921" || metadata.Source.RawRecordCount != 37 || metadata.Source.NormalizedSafeTableSHA256 == "" {
		t.Fatalf("source metadata mismatch: %#v", metadata.Source)
	}
	for _, discarded := range []string{"DIRECTOR", "PHONE NO.", "E-MAIL ADDRESS", "WEBSITES", "ADDRESS"} {
		if !contains(metadata.Source.DiscardedSourceColumns, discarded) {
			t.Fatalf("discarded source column %q missing", discarded)
		}
	}
	if !strings.Contains(metadata.Source.RawHTMLHashNote, "different byte sizes") || !strings.Contains(strings.Join(metadata.SourceLimitations, " "), "not necessarily currently accredited") {
		t.Fatal("metadata limitations are not explicit")
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
	if err := json.Unmarshal(readTextBytes(t, datasetPath("schemas/healthcare/nhia_state_social_health_insurance_agencies.schema.json")), &schema); err != nil {
		t.Fatalf("decode NHIA SSHIA schema: %v", err)
	}
	definition := schema.Defs["nhiaStateSocialHealthInsuranceAgency"]
	expectedFields := []string{"id", "name", "state_id", "country_code", "organisation_type"}
	if schema.Title != metadata.Title || schema.MinItems != recordCount || schema.MaxItems != recordCount || !schema.UniqueItems || !reflect.DeepEqual(definition.Required, expectedFields) || len(definition.Properties) != len(expectedFields) {
		t.Fatalf("schema mismatch: %#v", schema)
	}
	if definition.Properties["id"].MaxLength != NHIAStateSocialHealthInsuranceAgencyIDMaxLength || !reflect.DeepEqual(definition.Properties["organisation_type"].Enum, []string{"state_social_health_insurance_agency"}) {
		t.Fatal("schema bounds or enum mismatch")
	}
}

func assertNHIAStateSocialHealthInsuranceAgencyReconciliation(t *testing.T, finalIDs, stateIDs map[string]struct{}) {
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
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/healthcare/nhia_state_social_health_insurance_agencies_reconciliation/index.json")), &index); err != nil {
		t.Fatalf("decode NHIA SSHIA reconciliation index: %v", err)
	}
	if index.DatasetKey != "ng-nhia-state-social-health-insurance-agencies" || index.SourceRows != 37 || index.FinalRecordCount != 37 || index.DecisionCounts["retain"] != 37 || len(index.DecisionCounts) != 1 || len(index.Partitions) != 1 {
		t.Fatalf("reconciliation index mismatch: %#v", index)
	}
	seenPositions := map[int]struct{}{}
	seenStates := map[string]struct{}{}
	seenFinalIDs := map[string]struct{}{}
	for _, partition := range index.Partitions {
		data, err := os.ReadFile(datasetPath(partition.Path))
		if err != nil {
			t.Fatalf("read reconciliation partition %s: %v", partition.Path, err)
		}
		if int64(len(data)) != partition.SizeBytes || fmt.Sprintf("%x", sha256.Sum256(data)) != partition.SHA256 {
			t.Fatalf("partition size or hash mismatch for %s", partition.Path)
		}
		var value struct {
			Records []struct {
				SourcePosition     int            `json:"source_position"`
				RawOrganisation    string         `json:"raw_organisation_name"`
				RawState           string         `json:"raw_state"`
				NormalizedName     string         `json:"normalized_name"`
				StateID            string         `json:"state_id"`
				FinalID            *string        `json:"final_record_id"`
				Decision           string         `json:"decision"`
				Reason             string         `json:"reason"`
				SafeEvidence       map[string]any `json:"safe_evidence"`
				UnsupportedContact string         `json:"contact"`
			} `json:"records"`
		}
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatalf("decode reconciliation partition %s: %v", partition.Path, err)
		}
		for _, record := range value.Records {
			if record.SourcePosition == 0 || record.RawOrganisation == "" || record.RawState == "" || record.NormalizedName == "" || record.StateID == "" || record.FinalID == nil || record.Decision != "retain" || record.Reason == "" || len(record.SafeEvidence) == 0 || record.UnsupportedContact != "" {
				t.Fatalf("invalid reconciliation row: %#v", record)
			}
			if _, ok := seenPositions[record.SourcePosition]; ok {
				t.Fatalf("duplicate source position %d", record.SourcePosition)
			}
			if _, ok := stateIDs[record.StateID]; !ok {
				t.Fatalf("unknown state in reconciliation %q", record.StateID)
			}
			if _, ok := finalIDs[*record.FinalID]; !ok {
				t.Fatalf("reconciliation target missing %q", *record.FinalID)
			}
			seenPositions[record.SourcePosition] = struct{}{}
			seenStates[record.StateID] = struct{}{}
			seenFinalIDs[*record.FinalID] = struct{}{}
		}
	}
	if len(seenPositions) != 37 || len(seenStates) != 37 || len(seenFinalIDs) != len(finalIDs) {
		t.Fatalf("reconciliation coverage mismatch positions=%d states=%d ids=%d", len(seenPositions), len(seenStates), len(seenFinalIDs))
	}
}

func assertNHIAStateSocialHealthInsuranceAgencyPrivacy(t *testing.T) {
	t.Helper()
	for _, path := range []string{
		"healthcare/nhia_state_social_health_insurance_agencies.json",
		"metadata/healthcare/nhia_state_social_health_insurance_agencies_reconciliation/index.json",
		"metadata/healthcare/nhia_state_social_health_insurance_agencies_reconciliation/state_social_health_insurance_agencies.json",
	} {
		text := string(readTextBytes(t, datasetPath(path)))
		lower := strings.ToLower(text)
		for _, marker := range []string{"director", "phone", "phone_number", "email", "email_address", "contact", "contact_person", "personal", "enrollee", "practitioner", "claim", "policy_number", "bank", "account", "credential", "token", "<html", "<table"} {
			if strings.Contains(lower, marker) {
				t.Fatalf("%s contains prohibited marker %q", path, marker)
			}
		}
		if regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`).MatchString(text) || regexp.MustCompile(`\b(?:\+?234|0)[789][01]\d{8}\b`).MatchString(text) {
			t.Fatalf("%s contains contact-looking value", path)
		}
	}
}

func canonicalStateIDs(t *testing.T) map[string]struct{} {
	t.Helper()
	var states []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("geography/states.json")), &states); err != nil {
		t.Fatalf("decode states: %v", err)
	}
	out := make(map[string]struct{}, len(states))
	for _, state := range states {
		out[state.ID] = struct{}{}
	}
	return out
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
