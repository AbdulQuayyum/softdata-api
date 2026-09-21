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

func TestFRSCZonalCommandJSONContract(t *testing.T) {
	record := FRSCZonalCommand{
		ID:          "frsc-rs7hq-fct-zonal-command",
		Name:        "FRSC RS7HQ FCT Zonal Command",
		CommandType: "zonal_command",
		CommandCode: "RS7HQ",
		StateID:     "fct",
		CountryCode: "NG",
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal FRSC zonal command: %v", err)
	}
	want := `{"id":"frsc-rs7hq-fct-zonal-command","name":"FRSC RS7HQ FCT Zonal Command","command_type":"zonal_command","command_code":"RS7HQ","state_id":"fct","country_code":"NG"}`
	if string(encoded) != want {
		t.Fatalf("unexpected JSON: %s", encoded)
	}
}

func TestFRSCZonalCommandsDatasetAndReconciliation(t *testing.T) {
	var records []FRSCZonalCommand
	decoder := json.NewDecoder(strings.NewReader(string(readTextBytes(t, datasetPath("emergency/frsc_zonal_commands.json")))))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&records); err != nil {
		t.Fatalf("decode FRSC zonal commands: %v", err)
	}
	if len(records) != 12 {
		t.Fatalf("unexpected FRSC zonal command count: got %d want 12", len(records))
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
	codePattern := regexp.MustCompile(`^RS(?:[1-9]|1[0-2])HQ$`)
	recordsByID := map[string]FRSCZonalCommand{}
	seenCodes := map[string]struct{}{}
	for i, record := range records {
		if record.ID == "" || record.Name == "" || record.CommandType != "zonal_command" || record.CommandCode == "" || record.StateID == "" || record.CountryCode != "NG" {
			t.Fatalf("record %d has invalid required fields: %#v", i, record)
		}
		if len(record.ID) > FRSCZonalCommandIDMaxLength || FRSCZonalCommandIDMaxLength != 255 || !idPattern.MatchString(record.ID) {
			t.Fatalf("record %d has invalid id %q", i, record.ID)
		}
		if !codePattern.MatchString(record.CommandCode) {
			t.Fatalf("record %d has invalid command_code %q", i, record.CommandCode)
		}
		if _, ok := validStates[record.StateID]; !ok {
			t.Fatalf("record %d has invalid state_id %q", i, record.StateID)
		}
		if _, ok := recordsByID[record.ID]; ok {
			t.Fatalf("duplicate public ID %q", record.ID)
		}
		recordsByID[record.ID] = record
		if _, ok := seenCodes[record.CommandCode]; ok {
			t.Fatalf("duplicate command_code %q", record.CommandCode)
		}
		seenCodes[record.CommandCode] = struct{}{}
	}

	ordered := append([]FRSCZonalCommand(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	if !reflect.DeepEqual(records, ordered) {
		t.Fatal("FRSC zonal commands are not deterministically sorted by ID")
	}

	assertFRSCZonalCommandMetadataAndSchema(t, len(records))
	assertFRSCZonalCommandReconciliation(t, recordsByID)
	assertFRSCZonalCommandPrivacy(t)
}

func TestFRSCZonalCommandPublicFieldContract(t *testing.T) {
	expected := map[string]bool{"id": true, "name": true, "command_type": true, "command_code": true, "state_id": true, "country_code": true}
	model := reflect.TypeOf(FRSCZonalCommand{})
	if model.NumField() != len(expected) {
		t.Fatal("unexpected public model fields")
	}
	for i := 0; i < model.NumField(); i++ {
		if !expected[strings.Split(model.Field(i).Tag.Get("json"), ",")[0]] {
			t.Fatal("unsupported public field")
		}
	}
}

func TestFRSCZonalCommandsDeterministicRegeneration(t *testing.T) {
	cmd := exec.Command("python3", "../../tools/generate_frsc_zonal_commands.py", "--repo-root", "../..", "--check")
	cmd.Env = append(os.Environ(), "PYTHONPYCACHEPREFIX="+filepath.Join(t.TempDir(), "pycache"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("FRSC zonal commands generator check failed: %v\n%s", err, output)
	}
}

func assertFRSCZonalCommandMetadataAndSchema(t *testing.T, recordCount int) {
	t.Helper()
	var metadata struct {
		DatasetKey        string         `json:"dataset_key"`
		Title             string         `json:"title"`
		Group             string         `json:"group"`
		CountryCode       string         `json:"country_code"`
		RecordCount       int            `json:"record_count"`
		SourceRows        int            `json:"source_rows"`
		DecisionCounts    map[string]int `json:"decision_counts"`
		CommandTypeCounts map[string]int `json:"command_type_counts"`
		StateCounts       map[string]int `json:"state_counts"`
		Limitations       []string       `json:"limitations"`
		Sources           map[string]any `json:"sources"`
		Privacy           string         `json:"privacy"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/emergency/frsc_zonal_commands.json")), &metadata); err != nil {
		t.Fatalf("decode FRSC metadata: %v", err)
	}
	if metadata.DatasetKey != "ng-frsc-zonal-commands" || metadata.Title != "Nigeria FRSC Zonal Commands Snapshot" || metadata.Group != "emergency" || metadata.CountryCode != "NG" || metadata.RecordCount != recordCount || metadata.SourceRows != 12 {
		t.Fatalf("metadata mismatch: %#v", metadata)
	}
	if !reflect.DeepEqual(metadata.DecisionCounts, map[string]int{"retain": 12}) || !reflect.DeepEqual(metadata.CommandTypeCounts, map[string]int{"zonal_command": 12}) || len(metadata.StateCounts) != 12 {
		t.Fatalf("metadata count mismatch: %#v", metadata)
	}
	limitationText := strings.Join(metadata.Limitations, " ")
	if len(metadata.Sources) != 1 || !strings.Contains(limitationText, "not a live operational-status guarantee") || !strings.Contains(limitationText, "sector commands") || !strings.Contains(metadata.Privacy, "commander") {
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
				Pattern   string   `json:"pattern"`
			} `json:"properties"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("schemas/emergency/frsc_zonal_commands.schema.json")), &schema); err != nil {
		t.Fatalf("decode FRSC schema: %v", err)
	}
	definition := schema.Defs["frscZonalCommand"]
	wantRequired := []string{"id", "name", "command_type", "command_code", "state_id", "country_code"}
	if schema.Title != metadata.Title || schema.MinItems != recordCount || schema.MaxItems != recordCount || !schema.UniqueItems || !reflect.DeepEqual(definition.Required, wantRequired) {
		t.Fatalf("schema mismatch: %#v", schema)
	}
	if definition.Properties["id"].MaxLength != FRSCZonalCommandIDMaxLength || definition.Properties["country_code"].Const != "NG" || !reflect.DeepEqual(definition.Properties["command_type"].Enum, []string{"zonal_command"}) || definition.Properties["command_code"].Pattern == "" {
		t.Fatal("schema bounds, country code, command code or command type mismatch")
	}
}

func assertFRSCZonalCommandReconciliation(t *testing.T, recordsByID map[string]FRSCZonalCommand) {
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
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/emergency/frsc_zonal_commands_reconciliation/index.json")), &index); err != nil {
		t.Fatalf("decode FRSC reconciliation index: %v", err)
	}
	if index.DatasetKey != "ng-frsc-zonal-commands" || index.SourceRows != 12 || index.FinalRecordCount != len(recordsByID) || !reflect.DeepEqual(index.DecisionCounts, map[string]int{"retain": 12}) || len(index.Partitions) != 12 {
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
		if len(data) > 25*1024*1024 {
			t.Fatalf("partition %s exceeds file-size policy", partition.Path)
		}
		var value struct {
			Records []struct {
				SourceKey             string          `json:"source_key"`
				SourcePosition        int             `json:"source_position"`
				SourceCommandCode     string          `json:"source_command_code"`
				NormalizedCommandType string          `json:"normalized_command_type"`
				Normalized            json.RawMessage `json:"normalized"`
				Decision              string          `json:"decision"`
				TargetID              *string         `json:"target_id"`
				Reason                string          `json:"reason"`
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
			if row.Decision != "retain" || row.TargetID == nil || row.Reason == "" || row.NormalizedCommandType != "zonal_command" || row.SourceCommandCode == "" {
				t.Fatalf("invalid reconciliation row in %s: %#v", partition.Path, row)
			}
			if _, ok := recordsByID[*row.TargetID]; !ok {
				t.Fatalf("target %q missing from public records", *row.TargetID)
			}
			retained++
		}
	}
	if retained != len(recordsByID) || len(seenSourceRows) != 12 {
		t.Fatal("reconciliation arithmetic does not close")
	}
}

func assertFRSCZonalCommandPrivacy(t *testing.T) {
	t.Helper()
	paths := []string{
		"emergency/frsc_zonal_commands.json",
		"schemas/emergency/frsc_zonal_commands.schema.json",
		"metadata/emergency/frsc_zonal_commands.json",
		"metadata/emergency/frsc_zonal_commands_reconciliation/index.json",
	}
	for _, path := range paths {
		data := strings.ToLower(string(readTextBytes(t, datasetPath(path))))
		for _, marker := range []string{"@", "password", "api_key", "secret", "token", "victim", "patient", "offender", "whatsapp"} {
			if strings.Contains(data, marker) {
				t.Fatalf("privacy marker %q found in %s", marker, path)
			}
		}
	}
}
