package models

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestNEMAZonalTerritorialOperationOfficeJSONContract(t *testing.T) {
	record := NEMAZonalTerritorialOperationOffice{
		ID:          "nema-abuja-zonal-territorial-operation-office",
		Name:        "NEMA Abuja Office",
		OfficeType:  "zonal_territorial_operation_office",
		StateID:     "fct",
		CountryCode: "NG",
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal NEMA office: %v", err)
	}
	want := `{"id":"nema-abuja-zonal-territorial-operation-office","name":"NEMA Abuja Office","office_type":"zonal_territorial_operation_office","state_id":"fct","country_code":"NG"}`
	if string(encoded) != want {
		t.Fatalf("unexpected JSON: %s", encoded)
	}
}

func TestNEMAZonalTerritorialOperationOfficesDatasetAndReconciliation(t *testing.T) {
	var records []NEMAZonalTerritorialOperationOffice
	decoder := json.NewDecoder(strings.NewReader(string(readTextBytes(t, datasetPath("emergency/nema_zonal_territorial_operation_offices.json")))))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&records); err != nil {
		t.Fatalf("decode NEMA offices: %v", err)
	}
	if len(records) != 17 {
		t.Fatalf("unexpected NEMA office count: got %d want 17", len(records))
	}

	validStates := map[string]struct{}{}
	var states []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("geography/states.json")), &states); err != nil {
		t.Fatalf("decode states: %v", err)
	}
	for _, state := range states {
		validStates[state.ID] = struct{}{}
	}

	idPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	recordsByID := map[string]NEMAZonalTerritorialOperationOffice{}
	stateCounts := map[string]int{}
	for i, record := range records {
		if record.ID == "" || record.Name == "" || record.OfficeType != "zonal_territorial_operation_office" || record.StateID == "" || record.CountryCode != "NG" {
			t.Fatalf("record %d has invalid required fields: %#v", i, record)
		}
		if len(record.ID) > NEMAZonalTerritorialOperationOfficeIDMaxLength || NEMAZonalTerritorialOperationOfficeIDMaxLength != 255 || !idPattern.MatchString(record.ID) {
			t.Fatalf("record %d has invalid id %q", i, record.ID)
		}
		if _, ok := validStates[record.StateID]; !ok {
			t.Fatalf("record %d has invalid state_id %q", i, record.StateID)
		}
		if _, ok := recordsByID[record.ID]; ok {
			t.Fatalf("duplicate public ID %q", record.ID)
		}
		recordsByID[record.ID] = record
		stateCounts[record.StateID]++
	}
	if len(stateCounts) != 17 {
		t.Fatalf("unexpected state coverage: %#v", stateCounts)
	}

	ordered := append([]NEMAZonalTerritorialOperationOffice(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	if !reflect.DeepEqual(records, ordered) {
		t.Fatal("NEMA offices are not deterministically sorted by ID")
	}

	assertNEMAZonalTerritorialOperationOfficeMetadataAndSchema(t, len(records))
	assertNEMAZonalTerritorialOperationOfficeReconciliation(t, recordsByID)
	assertNEMAZonalTerritorialOperationOfficePrivacy(t)
}

func TestNEMAZonalTerritorialOperationOfficePublicFieldContract(t *testing.T) {
	expected := map[string]bool{"id": true, "name": true, "office_type": true, "state_id": true, "country_code": true}
	model := reflect.TypeOf(NEMAZonalTerritorialOperationOffice{})
	if model.NumField() != len(expected) {
		t.Fatal("unexpected public model fields")
	}
	for i := 0; i < model.NumField(); i++ {
		if !expected[strings.Split(model.Field(i).Tag.Get("json"), ",")[0]] {
			t.Fatal("unsupported public field")
		}
	}
}

func TestNEMAZonalTerritorialOperationOfficesDeterministicRegeneration(t *testing.T) {
	cmd := exec.Command("python3", "../../tools/generate_nema_zonal_territorial_operation_offices.py", "--repo-root", "../..", "--check")
	cmd.Env = append(os.Environ(), "PYTHONPYCACHEPREFIX="+filepath.Join(t.TempDir(), "pycache"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("NEMA offices generator check failed: %v\n%s", err, output)
	}
}

func assertNEMAZonalTerritorialOperationOfficeMetadataAndSchema(t *testing.T, recordCount int) {
	t.Helper()
	var metadata struct {
		DatasetKey      string         `json:"dataset_key"`
		Title           string         `json:"title"`
		Group           string         `json:"group"`
		CountryCode     string         `json:"country_code"`
		RecordCount     int            `json:"record_count"`
		SourceRows      int            `json:"source_rows"`
		DecisionCounts  map[string]int `json:"decision_counts"`
		OfficeTypeCount map[string]int `json:"office_type_counts"`
		StateCounts     map[string]int `json:"state_counts"`
		Limitations     []string       `json:"limitations"`
		Sources         map[string]any `json:"sources"`
		Privacy         string         `json:"privacy"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/emergency/nema_zonal_territorial_operation_offices.json")), &metadata); err != nil {
		t.Fatalf("decode NEMA metadata: %v", err)
	}
	if metadata.DatasetKey != "ng-nema-zonal-territorial-operation-offices" || metadata.Title != "Nigeria NEMA Zonal, Territorial and Operation Offices Snapshot" || metadata.Group != "emergency" || metadata.CountryCode != "NG" || metadata.RecordCount != recordCount || metadata.SourceRows != 17 {
		t.Fatalf("metadata mismatch: %#v", metadata)
	}
	if !reflect.DeepEqual(metadata.DecisionCounts, map[string]int{"retain": 17}) || !reflect.DeepEqual(metadata.OfficeTypeCount, map[string]int{"zonal_territorial_operation_office": 17}) || len(metadata.StateCounts) != 17 {
		t.Fatalf("metadata count mismatch: %#v", metadata)
	}
	if len(metadata.Sources) != 3 || !strings.Contains(strings.Join(metadata.Limitations, " "), "not a live operational-status guarantee") || !strings.Contains(metadata.Privacy, "Staff") {
		t.Fatal("metadata source, limitation or privacy statement mismatch")
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
				Const     string   `json:"const"`
			} `json:"properties"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("schemas/emergency/nema_zonal_territorial_operation_offices.schema.json")), &schema); err != nil {
		t.Fatalf("decode NEMA schema: %v", err)
	}
	definition := schema.Defs["nemaOffice"]
	if schema.Title != metadata.Title || schema.MinItems != recordCount || schema.MaxItems != recordCount || !schema.UniqueItems || !reflect.DeepEqual(definition.Required, []string{"id", "name", "office_type", "state_id", "country_code"}) {
		t.Fatalf("schema mismatch: %#v", schema)
	}
	if definition.Properties["id"].MaxLength != NEMAZonalTerritorialOperationOfficeIDMaxLength || definition.Properties["country_code"].Const != "NG" || !reflect.DeepEqual(definition.Properties["office_type"].Enum, []string{"zonal_territorial_operation_office"}) {
		t.Fatal("schema bounds, country code or office type mismatch")
	}
}

func assertNEMAZonalTerritorialOperationOfficeReconciliation(t *testing.T, recordsByID map[string]NEMAZonalTerritorialOperationOffice) {
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
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/emergency/nema_zonal_territorial_operation_offices_reconciliation/index.json")), &index); err != nil {
		t.Fatalf("decode NEMA reconciliation index: %v", err)
	}
	if index.DatasetKey != "ng-nema-zonal-territorial-operation-offices" || index.SourceRows != 17 || index.FinalRecordCount != len(recordsByID) || !reflect.DeepEqual(index.DecisionCounts, map[string]int{"retain": 17}) || len(index.Partitions) != 17 {
		t.Fatalf("reconciliation index mismatch: %#v", index)
	}
	seenSourceRows := map[string]struct{}{}
	retained := 0
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
				SourceKey         string          `json:"source_key"`
				SourcePosition    int             `json:"source_position"`
				PublishedLocation string          `json:"published_location"`
				Normalized        json.RawMessage `json:"normalized"`
				Decision          string          `json:"decision"`
				TargetID          *string         `json:"target_id"`
				Reason            string          `json:"reason"`
			} `json:"records"`
		}
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatalf("decode reconciliation partition %s: %v", partition.Path, err)
		}
		if len(value.Records) != 1 {
			t.Fatalf("unexpected partition row count for %s", partition.Path)
		}
		for _, row := range value.Records {
			key := fmt.Sprintf("%s/%d", row.SourceKey, row.SourcePosition)
			if _, ok := seenSourceRows[key]; ok {
				t.Fatalf("duplicate source row %s", key)
			}
			seenSourceRows[key] = struct{}{}
			if row.PublishedLocation == "" || len(row.Normalized) == 0 || row.Reason == "" {
				t.Fatalf("unsafe or incomplete reconciliation row: %#v", row)
			}
			if row.Decision != "retain" || row.TargetID == nil {
				t.Fatalf("unexpected decision row: %#v", row)
			}
			retained++
			if _, ok := recordsByID[*row.TargetID]; !ok {
				t.Fatalf("target ID does not resolve: %q", *row.TargetID)
			}
		}
	}
	if len(seenSourceRows) != index.SourceRows || retained != index.FinalRecordCount {
		t.Fatalf("reconciliation arithmetic mismatch: source=%d retained=%d final=%d", len(seenSourceRows), retained, index.FinalRecordCount)
	}
}

func assertNEMAZonalTerritorialOperationOfficePrivacy(t *testing.T) {
	t.Helper()
	paths := []string{
		"emergency/nema_zonal_territorial_operation_offices.json",
		"metadata/emergency/nema_zonal_territorial_operation_offices.json",
		"metadata/emergency/nema_zonal_territorial_operation_offices_reconciliation/index.json",
		"schemas/emergency/nema_zonal_territorial_operation_offices.schema.json",
	}
	var index struct {
		Partitions []struct {
			Path string `json:"path"`
		} `json:"partitions"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/emergency/nema_zonal_territorial_operation_offices_reconciliation/index.json")), &index); err != nil {
		t.Fatalf("decode NEMA reconciliation index: %v", err)
	}
	for _, partition := range index.Partitions {
		paths = append(paths, partition.Path)
	}
	for _, path := range paths {
		text := strings.ToLower(string(readTextBytes(t, datasetPath(path))))
		for _, forbidden := range []string{"@", "password", "api_key", "secret", "token", "victim", "patient", "whatsapp", "director general"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("privacy-sensitive marker %q found in %s", forbidden, path)
			}
		}
	}
}
