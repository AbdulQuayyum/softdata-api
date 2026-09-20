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

func TestEmergencyServiceContactJSONContract(t *testing.T) {
	record := EmergencyServiceContact{
		ID:           "federal-road-safety-corps-road-emergency-122-national",
		ServiceName:  "Emergency Ambulance Service Scheme (EASS) Zebra",
		AgencyName:   "Federal Road Safety Corps",
		ServiceType:  "road_emergency",
		ContactType:  "short_code",
		ContactValue: "122",
		CoverageType: "national",
		CountryCode:  "NG",
		CallCost:     "toll_free",
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal emergency service contact: %v", err)
	}
	want := `{"id":"federal-road-safety-corps-road-emergency-122-national","service_name":"Emergency Ambulance Service Scheme (EASS) Zebra","agency_name":"Federal Road Safety Corps","service_type":"road_emergency","contact_type":"short_code","contact_value":"122","coverage_type":"national","country_code":"NG","call_cost":"toll_free"}`
	if string(encoded) != want {
		t.Fatalf("unexpected JSON: %s", encoded)
	}
}

func TestEmergencyServiceContactsDatasetAndReconciliation(t *testing.T) {
	var records []EmergencyServiceContact
	decoder := json.NewDecoder(strings.NewReader(string(readTextBytes(t, datasetPath("emergency/emergency_service_contacts.json")))))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&records); err != nil {
		t.Fatalf("decode emergency contacts: %v", err)
	}
	if len(records) != 5 {
		t.Fatalf("unexpected emergency contact count: got %d want 5", len(records))
	}

	idPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	telephonePattern := regexp.MustCompile(`^(?:0[0-9]{10}|0800[0-9]{8}|\+234[0-9]{10})$`)
	seenIDs := map[string]struct{}{}
	recordsByID := map[string]EmergencyServiceContact{}
	serviceTypes := map[string]int{}
	coverageTypes := map[string]int{}
	values := map[string]struct{}{}
	for i, record := range records {
		if record.ID == "" || record.ServiceName == "" || record.AgencyName == "" || record.ServiceType == "" || record.ContactType == "" || record.ContactValue == "" || record.CoverageType == "" || record.CountryCode != "NG" {
			t.Fatalf("record %d has invalid required fields: %#v", i, record)
		}
		if len(record.ID) > EmergencyServiceContactIDMaxLength || EmergencyServiceContactIDMaxLength != 255 || !idPattern.MatchString(record.ID) {
			t.Fatalf("record %d has invalid id %q", i, record.ID)
		}
		switch record.ServiceType {
		case "general_emergency", "disaster_management", "road_emergency", "fire":
		default:
			t.Fatalf("record %d has unsupported service type %q", i, record.ServiceType)
		}
		switch record.ContactType {
		case "short_code":
			if !regexp.MustCompile(`^[0-9]{3}$`).MatchString(record.ContactValue) {
				t.Fatalf("record %d has invalid short code %q", i, record.ContactValue)
			}
		case "telephone":
			if !telephonePattern.MatchString(record.ContactValue) {
				t.Fatalf("record %d has invalid telephone %q", i, record.ContactValue)
			}
		default:
			t.Fatalf("record %d has unsupported contact type %q", i, record.ContactType)
		}
		if record.CoverageType != "national" {
			t.Fatalf("record %d has unsupported coverage %q", i, record.CoverageType)
		}
		if record.StateID != "" {
			t.Fatalf("record %d unexpectedly has state_id %q", i, record.StateID)
		}
		if record.Availability != "" && record.Availability != "24_hours" {
			t.Fatalf("record %d has invalid availability %q", i, record.Availability)
		}
		if record.CallCost != "" && record.CallCost != "toll_free" {
			t.Fatalf("record %d has invalid call cost %q", i, record.CallCost)
		}
		if _, ok := seenIDs[record.ID]; ok {
			t.Fatalf("duplicate public ID %q", record.ID)
		}
		seenIDs[record.ID] = struct{}{}
		recordsByID[record.ID] = record
		serviceTypes[record.ServiceType]++
		coverageTypes[record.CoverageType]++
		values[record.ContactValue] = struct{}{}
	}
	if !reflect.DeepEqual(serviceTypes, map[string]int{"disaster_management": 1, "fire": 2, "general_emergency": 1, "road_emergency": 1}) {
		t.Fatalf("unexpected service type counts: %#v", serviceTypes)
	}
	if !reflect.DeepEqual(coverageTypes, map[string]int{"national": 5}) {
		t.Fatalf("unexpected coverage counts: %#v", coverageTypes)
	}
	for _, value := range []string{"112", "122", "080022556362", "+2348032003557"} {
		if _, ok := values[value]; !ok {
			t.Fatalf("published contact %q was not preserved", value)
		}
	}

	ordered := append([]EmergencyServiceContact(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	if !reflect.DeepEqual(records, ordered) {
		t.Fatal("emergency contacts are not deterministically sorted by ID")
	}

	assertEmergencyServiceContactMetadataAndSchema(t, len(records))
	assertEmergencyServiceContactReconciliation(t, recordsByID)
	assertEmergencyServiceContactPrivacy(t)
}

func TestEmergencyServiceContactPublicFieldContract(t *testing.T) {
	expected := map[string]bool{
		"id": true, "service_name": true, "agency_name": true, "service_type": true,
		"contact_type": true, "contact_value": true, "coverage_type": true, "country_code": true,
		"state_id": true, "availability": true, "call_cost": true, "notes": true,
	}
	model := reflect.TypeOf(EmergencyServiceContact{})
	if model.NumField() != len(expected) {
		t.Fatal("unexpected public model fields")
	}
	for i := 0; i < model.NumField(); i++ {
		if !expected[strings.Split(model.Field(i).Tag.Get("json"), ",")[0]] {
			t.Fatal("unsupported public field")
		}
	}
}

func TestEmergencyServiceContactsDeterministicRegeneration(t *testing.T) {
	cmd := exec.Command("python3", "../../tools/generate_emergency_service_contacts.py", "--repo-root", "../..", "--check")
	cmd.Env = append(os.Environ(), "PYTHONPYCACHEPREFIX="+filepath.Join(t.TempDir(), "pycache"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("emergency generator check failed: %v\n%s", err, output)
	}
}

func assertEmergencyServiceContactMetadataAndSchema(t *testing.T, recordCount int) {
	t.Helper()
	var metadata struct {
		DatasetKey         string         `json:"dataset_key"`
		Title              string         `json:"title"`
		Group              string         `json:"group"`
		CountryCode        string         `json:"country_code"`
		RecordCount        int            `json:"record_count"`
		SourceRows         int            `json:"source_rows"`
		DecisionCounts     map[string]int `json:"decision_counts"`
		ServiceTypeCounts  map[string]int `json:"service_type_counts"`
		CoverageTypeCounts map[string]int `json:"coverage_type_counts"`
		ContactTypeCounts  map[string]int `json:"contact_type_counts"`
		Limitations        []string       `json:"limitations"`
		Sources            map[string]any `json:"sources"`
		Privacy            string         `json:"privacy"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/emergency/emergency_service_contacts.json")), &metadata); err != nil {
		t.Fatalf("decode emergency metadata: %v", err)
	}
	if metadata.DatasetKey != "ng-emergency-service-contacts" || metadata.Title != "Nigeria Official Emergency Service Contacts Snapshot" || metadata.Group != "emergency" || metadata.CountryCode != "NG" || metadata.RecordCount != recordCount || metadata.SourceRows != 5 {
		t.Fatalf("metadata mismatch: %#v", metadata)
	}
	if !reflect.DeepEqual(metadata.DecisionCounts, map[string]int{"retain": 5}) {
		t.Fatalf("decision counts mismatch: %#v", metadata.DecisionCounts)
	}
	if !reflect.DeepEqual(metadata.ServiceTypeCounts, map[string]int{"disaster_management": 1, "fire": 2, "general_emergency": 1, "road_emergency": 1}) || !reflect.DeepEqual(metadata.CoverageTypeCounts, map[string]int{"national": 5}) || !reflect.DeepEqual(metadata.ContactTypeCounts, map[string]int{"short_code": 3, "telephone": 2}) {
		t.Fatalf("metadata count mismatch: %#v", metadata)
	}
	if len(metadata.Sources) != 4 || !strings.Contains(strings.Join(metadata.Limitations, " "), "not a complete nationwide register") || !strings.Contains(metadata.Privacy, "Personal") {
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
	if err := json.Unmarshal(readTextBytes(t, datasetPath("schemas/emergency/emergency_service_contacts.schema.json")), &schema); err != nil {
		t.Fatalf("decode emergency schema: %v", err)
	}
	definition := schema.Defs["emergencyServiceContact"]
	expectedFields := []string{"id", "service_name", "agency_name", "service_type", "contact_type", "contact_value", "coverage_type", "country_code"}
	if schema.Title != metadata.Title || schema.MinItems != recordCount || schema.MaxItems != recordCount || !schema.UniqueItems || !reflect.DeepEqual(definition.Required, expectedFields) {
		t.Fatalf("schema mismatch: %#v", schema)
	}
	if definition.Properties["id"].MaxLength != EmergencyServiceContactIDMaxLength || definition.Properties["country_code"].Const != "NG" {
		t.Fatal("schema bounds or country code mismatch")
	}
}

func assertEmergencyServiceContactReconciliation(t *testing.T, recordsByID map[string]EmergencyServiceContact) {
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
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/emergency/emergency_service_contacts_reconciliation/index.json")), &index); err != nil {
		t.Fatalf("decode emergency reconciliation index: %v", err)
	}
	if index.DatasetKey != "ng-emergency-service-contacts" || index.SourceRows != 5 || index.FinalRecordCount != len(recordsByID) || !reflect.DeepEqual(index.DecisionCounts, map[string]int{"retain": 5}) || len(index.Partitions) != 1 {
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
				PublishedContact  string          `json:"published_contact"`
				PublishedCoverage string          `json:"published_coverage"`
				Normalized        json.RawMessage `json:"normalized"`
				Decision          string          `json:"decision"`
				TargetID          *string         `json:"target_id"`
				Reason            string          `json:"reason"`
			} `json:"records"`
		}
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatalf("decode reconciliation partition %s: %v", partition.Path, err)
		}
		for _, row := range value.Records {
			key := fmt.Sprintf("%s/%d", row.SourceKey, row.SourcePosition)
			if _, ok := seenSourceRows[key]; ok {
				t.Fatalf("duplicate source row %s", key)
			}
			seenSourceRows[key] = struct{}{}
			if row.PublishedContact == "" || row.PublishedCoverage == "" || len(row.Normalized) == 0 || row.Reason == "" {
				t.Fatalf("unsafe or incomplete reconciliation row: %#v", row)
			}
			if row.Decision == "retain" {
				retained++
				if row.TargetID == nil {
					t.Fatal("retained row is missing target ID")
				}
				if _, ok := recordsByID[*row.TargetID]; !ok {
					t.Fatalf("target ID does not resolve: %q", *row.TargetID)
				}
			} else if row.TargetID != nil {
				t.Fatalf("excluded row has target ID %q", *row.TargetID)
			}
		}
	}
	if len(seenSourceRows) != index.SourceRows || retained != index.FinalRecordCount {
		t.Fatalf("reconciliation arithmetic mismatch: source=%d retained=%d final=%d", len(seenSourceRows), retained, index.FinalRecordCount)
	}
}

func assertEmergencyServiceContactPrivacy(t *testing.T) {
	t.Helper()
	for _, path := range []string{
		"emergency/emergency_service_contacts.json",
		"metadata/emergency/emergency_service_contacts.json",
		"metadata/emergency/emergency_service_contacts_reconciliation/index.json",
		"metadata/emergency/emergency_service_contacts_reconciliation/national.json",
		"schemas/emergency/emergency_service_contacts.schema.json",
	} {
		text := strings.ToLower(string(readTextBytes(t, datasetPath(path))))
		for _, forbidden := range []string{"@", "password", "api_key", "secret", "token", "victim", "patient", "whatsapp"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("privacy-sensitive marker %q found in %s", forbidden, path)
			}
		}
	}
}
