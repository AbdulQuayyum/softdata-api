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

func TestNHIAActiveAccreditedHealthcareProviderJSONContract(t *testing.T) {
	record := NHIAActiveAccreditedHealthcareProvider{
		ID:            "ab-0001-p",
		Name:          "Example Health Centre",
		CountryCode:   "NG",
		ProviderCode:  "AB/0001/P",
		FacilityType:  "primary",
		ListingStatus: "active_accredited",
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal NHIA HCP: %v", err)
	}
	if string(encoded) != `{"id":"ab-0001-p","name":"Example Health Centre","country_code":"NG","provider_code":"AB/0001/P","facility_type":"primary","listing_status":"active_accredited"}` {
		t.Fatalf("unexpected JSON: %s", encoded)
	}
}

func TestNHIAActiveAccreditedHealthcareProvidersDatasetAndReconciliation(t *testing.T) {
	var records []NHIAActiveAccreditedHealthcareProvider
	decoder := json.NewDecoder(strings.NewReader(string(readTextBytes(t, datasetPath("healthcare/nhia_active_accredited_healthcare_providers.json")))))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&records); err != nil {
		t.Fatalf("decode NHIA HCPs: %v", err)
	}
	if len(records) != 6536 {
		t.Fatalf("unexpected NHIA HCP count: got %d want 6536", len(records))
	}

	idPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	codePattern := regexp.MustCompile(`^(?:[A-Z]{2,3})/[0-9]{4}/P$`)
	seenIDs := make(map[string]struct{}, len(records))
	seenCodes := make(map[string]struct{}, len(records))
	recordsByID := make(map[string]NHIAActiveAccreditedHealthcareProvider, len(records))
	facilityTypes := map[string]int{}
	for i, record := range records {
		if record.ID == "" || record.Name == "" || record.CountryCode != "NG" || record.ProviderCode == "" || record.FacilityType == "" || record.ListingStatus != "active_accredited" {
			t.Fatalf("record %d has invalid required fields: %#v", i, record)
		}
		if len(record.ID) > NHIAActiveAccreditedHealthcareProviderIDMaxLength || NHIAActiveAccreditedHealthcareProviderIDMaxLength != 255 || !idPattern.MatchString(record.ID) {
			t.Fatalf("record %d has invalid id %q", i, record.ID)
		}
		if !codePattern.MatchString(record.ProviderCode) {
			t.Fatalf("record %d has invalid provider code %q", i, record.ProviderCode)
		}
		if record.FacilityType != "primary" && record.FacilityType != "primary_and_secondary" {
			t.Fatalf("record %d has invalid facility type %q", i, record.FacilityType)
		}
		if _, ok := seenIDs[record.ID]; ok {
			t.Fatalf("duplicate public ID %q", record.ID)
		}
		if _, ok := seenCodes[record.ProviderCode]; ok {
			t.Fatalf("duplicate provider code %q", record.ProviderCode)
		}
		seenIDs[record.ID] = struct{}{}
		seenCodes[record.ProviderCode] = struct{}{}
		recordsByID[record.ID] = record
		facilityTypes[record.FacilityType]++
	}
	if facilityTypes["primary"] != 4001 || facilityTypes["primary_and_secondary"] != 2535 {
		t.Fatalf("unexpected retained facility type counts: %#v", facilityTypes)
	}
	ordered := append([]NHIAActiveAccreditedHealthcareProvider(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	for i := range records {
		if records[i].ID != ordered[i].ID {
			t.Fatal("NHIA HCP records are not deterministically sorted")
		}
	}

	assertNHIAActiveAccreditedHealthcareProviderMetadataAndSchema(t, len(records))
	assertNHIAActiveAccreditedHealthcareProviderReconciliation(t, recordsByID)
	assertNHIAActiveAccreditedHealthcareProviderPrivacy(t)
}

func TestNHIAActiveAccreditedHealthcareProviderPublicFieldContract(t *testing.T) {
	expected := map[string]bool{"id": true, "name": true, "country_code": true, "provider_code": true, "facility_type": true, "listing_status": true}
	model := reflect.TypeOf(NHIAActiveAccreditedHealthcareProvider{})
	if model.NumField() != len(expected) {
		t.Fatal("unexpected public model fields")
	}
	for i := 0; i < model.NumField(); i++ {
		if !expected[strings.Split(model.Field(i).Tag.Get("json"), ",")[0]] {
			t.Fatal("unsupported public field")
		}
	}
}

func TestNHIAActiveAccreditedHealthcareProvidersDeterministicRegeneration(t *testing.T) {
	tmp := t.TempDir()
	first := filepath.Join(tmp, "first")
	second := filepath.Join(tmp, "second")
	copyProviderArtifacts(t, first)
	copyProviderArtifacts(t, second)
	runProviderGeneratorReplay(t, first)
	runProviderGeneratorReplay(t, second)
	firstFiles := readJSONTree(t, first)
	secondFiles := readJSONTree(t, second)
	if !reflect.DeepEqual(firstFiles, secondFiles) {
		t.Fatal("NHIA HCP regeneration is not byte-identical")
	}
	committed := readJSONTree(t, filepath.Join("..", ".."))
	for path, data := range firstFiles {
		if !reflect.DeepEqual(committed[path], data) {
			t.Fatalf("regenerated artifact differs from committed %s", path)
		}
	}
}

func assertNHIAActiveAccreditedHealthcareProviderMetadataAndSchema(t *testing.T, recordCount int) {
	t.Helper()
	var metadata struct {
		DatasetKey                 string         `json:"dataset_key"`
		Title                      string         `json:"title"`
		Group                      string         `json:"group"`
		CountryCode                string         `json:"country_code"`
		RecordCount                int            `json:"record_count"`
		SourceRows                 int            `json:"source_rows"`
		NormalizedSafeTableSHA256  string         `json:"normalized_safe_table_sha256"`
		DecisionCounts             map[string]int `json:"decision_counts"`
		FacilityTypeCounts         map[string]int `json:"facility_type_counts"`
		ListingStatusCounts        map[string]int `json:"listing_status_counts"`
		SourceRetrieval            map[string]any `json:"source_retrieval"`
		PrivacyPolicy              map[string]any `json:"privacy_policy"`
		Limitations                []string       `json:"limitations"`
		UnsupportedAccreditation   string         `json:"accreditation_status"`
		UnsupportedLicenceStatus   string         `json:"licence_status"`
		UnsupportedOperationalStat string         `json:"operational_status"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/healthcare/nhia_active_accredited_healthcare_providers.json")), &metadata); err != nil {
		t.Fatalf("decode NHIA HCP metadata: %v", err)
	}
	if metadata.DatasetKey != "ng-nhia-active-accredited-healthcare-providers" || metadata.Title != "Nigeria NHIA Active Accredited Healthcare Providers Snapshot" || metadata.Group != "healthcare" || metadata.CountryCode != "NG" || metadata.RecordCount != recordCount || metadata.SourceRows != 6540 {
		t.Fatalf("metadata mismatch: %#v", metadata)
	}
	if metadata.DecisionCounts["retain"] != 6536 || metadata.DecisionCounts["exclude_unresolved_code_conflict"] != 2 || metadata.DecisionCounts["exclude_unresolved_safe_field_conflict"] != 2 || len(metadata.DecisionCounts) != 3 {
		t.Fatalf("metadata decision counts mismatch: %#v", metadata.DecisionCounts)
	}
	if metadata.FacilityTypeCounts["primary"] != 4004 || metadata.FacilityTypeCounts["primary_and_secondary"] != 2536 || metadata.ListingStatusCounts["active_accredited"] != 6536 || metadata.NormalizedSafeTableSHA256 == "" {
		t.Fatalf("metadata source counts mismatch: %#v", metadata)
	}
	if metadata.UnsupportedAccreditation != "" || metadata.UnsupportedLicenceStatus != "" || metadata.UnsupportedOperationalStat != "" {
		t.Fatal("metadata exposes unsupported status fields")
	}
	limitations := strings.Join(metadata.Limitations, " ")
	if !strings.Contains(limitations, "not a complete live licensing, registration or operational-status register") {
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
	if err := json.Unmarshal(readTextBytes(t, datasetPath("schemas/healthcare/nhia_active_accredited_healthcare_providers.schema.json")), &schema); err != nil {
		t.Fatalf("decode NHIA HCP schema: %v", err)
	}
	definition := schema.Defs["nhiaActiveAccreditedHealthcareProvider"]
	expectedFields := []string{"id", "name", "country_code", "provider_code", "facility_type", "listing_status"}
	if schema.Title != metadata.Title || schema.MinItems != recordCount || schema.MaxItems != recordCount || !schema.UniqueItems || !reflect.DeepEqual(definition.Required, expectedFields) || len(definition.Properties) != len(expectedFields) {
		t.Fatalf("schema mismatch: %#v", schema)
	}
	if definition.Properties["id"].MaxLength != NHIAActiveAccreditedHealthcareProviderIDMaxLength || !reflect.DeepEqual(definition.Properties["listing_status"].Enum, []string{"active_accredited"}) {
		t.Fatal("schema bounds or enum mismatch")
	}
}

func assertNHIAActiveAccreditedHealthcareProviderReconciliation(t *testing.T, recordsByID map[string]NHIAActiveAccreditedHealthcareProvider) {
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
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/healthcare/nhia_active_accredited_healthcare_providers_reconciliation/index.json")), &index); err != nil {
		t.Fatalf("decode NHIA HCP reconciliation index: %v", err)
	}
	if index.DatasetKey != "ng-nhia-active-accredited-healthcare-providers" || index.SourceRows != 6540 || index.FinalRecordCount != 6536 || index.DecisionCounts["retain"] != 6536 || index.DecisionCounts["exclude_unresolved_code_conflict"] != 2 || index.DecisionCounts["exclude_unresolved_safe_field_conflict"] != 2 || len(index.Partitions) != 1 {
		t.Fatalf("reconciliation index mismatch: %#v", index)
	}
	seenPositions := map[int]struct{}{}
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
				SourcePosition     int      `json:"source_position"`
				SourceProviderCode string   `json:"source_provider_code"`
				SourceProviderName string   `json:"source_provider_name"`
				SourceFacilityType string   `json:"source_facility_type"`
				ProviderCode       string   `json:"provider_code"`
				NormalizedName     string   `json:"normalized_name"`
				FacilityType       string   `json:"facility_type"`
				Decision           string   `json:"decision"`
				TargetID           *string  `json:"target_id"`
				Reason             string   `json:"reason"`
				Address            string   `json:"address"`
				Email              string   `json:"email"`
				Phone              string   `json:"phone"`
				Normalizations     []string `json:"normalization_rules"`
			} `json:"records"`
		}
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatalf("decode reconciliation partition %s: %v", partition.Path, err)
		}
		for _, record := range value.Records {
			if record.SourcePosition == 0 || record.SourceProviderCode == "" || record.SourceProviderName == "" || record.SourceFacilityType == "" || record.ProviderCode == "" || record.NormalizedName == "" || record.FacilityType == "" || record.Decision == "" || record.Reason == "" {
				t.Fatalf("invalid reconciliation row at position %d", record.SourcePosition)
			}
			if record.Address != "" || record.Email != "" || record.Phone != "" {
				t.Fatalf("prohibited reconciliation field at position %d", record.SourcePosition)
			}
			if _, ok := seenPositions[record.SourcePosition]; ok {
				t.Fatalf("duplicate source position %d", record.SourcePosition)
			}
			seenPositions[record.SourcePosition] = struct{}{}
			if record.Decision == "retain" {
				retained++
				if record.TargetID == nil {
					t.Fatalf("retained row %d has no target", record.SourcePosition)
				}
				public, ok := recordsByID[*record.TargetID]
				if !ok || public.ProviderCode != record.ProviderCode || public.Name != record.NormalizedName || public.FacilityType != record.FacilityType {
					t.Fatalf("retained row %d target mismatch", record.SourcePosition)
				}
			}
		}
	}
	if len(seenPositions) != 6540 || retained != len(recordsByID) {
		t.Fatalf("reconciliation coverage mismatch: positions=%d retained=%d", len(seenPositions), retained)
	}
}

func assertNHIAActiveAccreditedHealthcareProviderPrivacy(t *testing.T) {
	t.Helper()
	paths := []string{
		datasetPath("healthcare/nhia_active_accredited_healthcare_providers.json"),
		datasetPath("schemas/healthcare/nhia_active_accredited_healthcare_providers.schema.json"),
		datasetPath("metadata/healthcare/nhia_active_accredited_healthcare_providers_reconciliation/index.json"),
		datasetPath("metadata/healthcare/nhia_active_accredited_healthcare_providers_reconciliation/active_accredited_healthcare_providers.json"),
	}
	prohibitedKeys := regexp.MustCompile(`"(?i:address|website|website_url|logo|logo_url|phone|phone_number|email|email_address|director|contact|contact_person|personal|enrollee|practitioner|claim|policy_number|bank|account|credential|token)"\s*:`)
	emailPattern := regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	phonePattern := regexp.MustCompile(`(?:\+?234|0)[789][01][0-9][\s-]?[0-9]{3}[\s-]?[0-9]{4}`)
	for _, path := range paths {
		data := readTextFile(t, path)
		if prohibitedKeys.MatchString(data) || emailPattern.MatchString(data) || phonePattern.MatchString(data) || strings.Contains(data, "<table") || strings.Contains(data, "<html") {
			t.Fatalf("privacy scan failed for %s", path)
		}
	}
}

func copyProviderArtifacts(t *testing.T, target string) {
	t.Helper()
	for _, rel := range []string{
		"datasets/healthcare/nhia_active_accredited_healthcare_providers.json",
		"datasets/schemas/healthcare/nhia_active_accredited_healthcare_providers.schema.json",
		"datasets/metadata/healthcare/nhia_active_accredited_healthcare_providers.json",
		"datasets/metadata/healthcare/nhia_active_accredited_healthcare_providers_reconciliation/index.json",
		"datasets/metadata/healthcare/nhia_active_accredited_healthcare_providers_reconciliation/active_accredited_healthcare_providers.json",
	} {
		data := readTextBytes(t, filepath.Join("..", "..", rel))
		path := filepath.Join(target, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
}

func runProviderGeneratorReplay(t *testing.T, repo string) {
	t.Helper()
	cmd := exec.Command("python3", "tools/generate_nhia_active_accredited_healthcare_providers.py", "--repo", repo, "--from-reconciliation")
	cmd.Dir = filepath.Join("..", "..")
	cmd.Env = append(os.Environ(), "PYTHONPYCACHEPREFIX="+filepath.Join(t.TempDir(), "pycache"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("regenerate NHIA HCP artifacts: %v\n%s", err, output)
	}
}

func readJSONTree(t *testing.T, repo string) map[string][]byte {
	t.Helper()
	result := map[string][]byte{}
	for _, rel := range []string{
		"datasets/healthcare/nhia_active_accredited_healthcare_providers.json",
		"datasets/schemas/healthcare/nhia_active_accredited_healthcare_providers.schema.json",
		"datasets/metadata/healthcare/nhia_active_accredited_healthcare_providers.json",
		"datasets/metadata/healthcare/nhia_active_accredited_healthcare_providers_reconciliation/index.json",
		"datasets/metadata/healthcare/nhia_active_accredited_healthcare_providers_reconciliation/active_accredited_healthcare_providers.json",
	} {
		data, err := os.ReadFile(filepath.Join(repo, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		result[rel] = data
	}
	return result
}
