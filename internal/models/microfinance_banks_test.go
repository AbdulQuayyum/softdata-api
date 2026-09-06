package models

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestMicrofinanceBanksDatasetContract(t *testing.T) {
	var records []MicrofinanceBank
	if err := json.Unmarshal(readTextBytes(t, datasetPath("finance/microfinance_banks.json")), &records); err != nil {
		t.Fatalf("decode microfinance banks: %v", err)
	}
	if len(records) != 790 {
		t.Fatalf("record count = %d, want 790", len(records))
	}

	idPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	seenNames := make(map[string]struct{}, len(records))
	seenIDs := make(map[string]struct{}, len(records))
	seenCBN := make(map[string]struct{}, len(records))
	seenNIP := make(map[string]struct{}, len(records))
	cbnPattern := regexp.MustCompile(`^[0-9]{3}$`)
	nipPattern := regexp.MustCompile(`^[0-9]{6}$`)
	for i, record := range records {
		if record.ID == "" || record.Name == "" {
			t.Fatalf("record %d has an empty ID or name: %#v", i, record)
		}
		if record.CountryCode != "NG" {
			t.Fatalf("record %d has country code %q, want NG", i, record.CountryCode)
		}
		if !idPattern.MatchString(record.ID) {
			t.Fatalf("record %d has invalid ID %q", i, record.ID)
		}
		if _, exists := seenNames[record.Name]; exists {
			t.Fatalf("duplicate name: %q", record.Name)
		}
		if _, exists := seenIDs[record.ID]; exists {
			t.Fatalf("duplicate ID: %q", record.ID)
		}
		seenNames[record.Name] = struct{}{}
		seenIDs[record.ID] = struct{}{}
		if record.CBNCode != "" {
			if !cbnPattern.MatchString(record.CBNCode) {
				t.Fatalf("record %q has invalid cbn_code %q", record.ID, record.CBNCode)
			}
			if _, exists := seenCBN[record.CBNCode]; exists {
				t.Fatalf("duplicate cbn_code %q", record.CBNCode)
			}
			seenCBN[record.CBNCode] = struct{}{}
		}
		if record.NIPCode != "" {
			if !nipPattern.MatchString(record.NIPCode) {
				t.Fatalf("record %q has invalid nip_code %q", record.ID, record.NIPCode)
			}
			if _, exists := seenNIP[record.NIPCode]; exists {
				t.Fatalf("duplicate nip_code %q", record.NIPCode)
			}
			seenNIP[record.NIPCode] = struct{}{}
		}
		if strings.TrimSpace(record.Name) != record.Name {
			t.Fatalf("record %d has surrounding whitespace: %q", i, record.Name)
		}
	}

	for i := 1; i < len(records); i++ {
		previous, current := records[i-1], records[i]
		if strings.ToLower(previous.Name) > strings.ToLower(current.Name) ||
			(strings.EqualFold(previous.Name, current.Name) && previous.ID > current.ID) {
			t.Fatalf("records are not ordered by name then ID at %d: %#v, %#v", i, previous, current)
		}
	}

	for _, name := range []string{
		"BWAY Microfinance bank Limited",
		"Katsu Microfinance Bank Limited",
		"Paragon Microfinance Bank Limited",
		"Teerus Microfinance Bank Limited",
		"Oganiru Microfinance Bank Limited",
	} {
		if _, exists := seenNames[name]; !exists {
			t.Errorf("retained institution missing: %q", name)
		}
	}
	for _, name := range []string{
		"AKPO MICROFINANCE BANK LIMITED",
		"Verdant-Capital Microfinance Bank Limited",
	} {
		if _, exists := seenNames[name]; exists {
			t.Errorf("excluded institution present: %q", name)
		}
	}

	for _, field := range []string{"institution_type", "category", "state", "website", "logo"} {
		if strings.Contains(string(readTextBytes(t, datasetPath("finance/microfinance_banks.json"))), `"`+field+`"`) {
			t.Errorf("deferred field %q leaked into public dataset", field)
		}
	}
	for _, record := range records {
		if record.WebsiteURL != "" && !strings.HasPrefix(record.WebsiteURL, "https://") {
			t.Fatalf("record %q has a non-HTTPS website: %q", record.ID, record.WebsiteURL)
		}
		if record.LogoURL != "" && !strings.HasPrefix(record.LogoURL, "/v1/assets/financial-institutions/ng/microfinance-banks/") {
			t.Fatalf("record %q has an invalid logo URL: %q", record.ID, record.LogoURL)
		}
	}
}

func TestMicrofinanceBanksReconciliationManifest(t *testing.T) {
	var manifest struct {
		ExcludedRecords []struct {
			CBNID int `json:"cbn_id"`
		} `json:"excluded_records"`
		ExcludedRecordCount int `json:"excluded_record_count"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/finance/microfinance_banks_reconciliation.json")), &manifest); err != nil {
		t.Fatalf("decode reconciliation manifest: %v", err)
	}
	if manifest.ExcludedRecordCount != 39 || len(manifest.ExcludedRecords) != 39 {
		t.Fatalf("excluded count = %d/%d, want 39/39", manifest.ExcludedRecordCount, len(manifest.ExcludedRecords))
	}
	seen := make(map[int]struct{}, len(manifest.ExcludedRecords))
	for _, record := range manifest.ExcludedRecords {
		if record.CBNID == 0 {
			t.Fatal("reconciliation record has empty CBN ID")
		}
		if _, exists := seen[record.CBNID]; exists {
			t.Fatalf("duplicate excluded CBN ID: %d", record.CBNID)
		}
		seen[record.CBNID] = struct{}{}
	}
}

func TestMicrofinanceBanksMetadataAndSchema(t *testing.T) {
	var metadata struct {
		DatasetKey       string `json:"dataset_key"`
		DatasetGroup     string `json:"dataset_group"`
		RecordCount      int    `json:"record_count"`
		Version          string `json:"version"`
		RawRowExclusions int    `json:"raw_row_exclusions"`
		PrimarySource    struct {
			SHA256 string `json:"sha256"`
		} `json:"primary_source"`
		ExclusionArithmetic []string `json:"exclusion_arithmetic"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("metadata/finance/microfinance_banks.json")), &metadata); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	if metadata.DatasetKey != "ng-microfinance-banks" || metadata.DatasetGroup != "finance" || metadata.RecordCount != 790 || metadata.Version != "1.0.0" {
		t.Fatalf("unexpected metadata identity: %#v", metadata)
	}
	if metadata.RawRowExclusions != 39 || len(metadata.ExclusionArithmetic) != 7 {
		t.Fatalf("unexpected exclusion metadata: %#v", metadata)
	}
	if metadata.PrimarySource.SHA256 != "89d9791403589a98d82e4c9c2cd3457d26e226d194dd9fb8e5d83db2940b4725" {
		t.Fatalf("unexpected CBN snapshot hash: %q", metadata.PrimarySource.SHA256)
	}

	var schema struct {
		MinItems    int  `json:"minItems"`
		MaxItems    int  `json:"maxItems"`
		UniqueItems bool `json:"uniqueItems"`
		Items       struct {
			AdditionalProperties bool     `json:"additionalProperties"`
			Required             []string `json:"required"`
		} `json:"items"`
	}
	if err := json.Unmarshal(readTextBytes(t, datasetPath("schemas/finance/microfinance_banks.schema.json")), &schema); err != nil {
		t.Fatalf("decode schema: %v", err)
	}
	if schema.MinItems != 790 || schema.MaxItems != 790 || !schema.UniqueItems || schema.Items.AdditionalProperties {
		t.Fatalf("unexpected schema constraints: %#v", schema)
	}
	wantRequired := []string{"id", "name", "country_code"}
	if len(schema.Items.Required) != len(wantRequired) {
		t.Fatalf("unexpected schema required fields: %#v", schema.Items.Required)
	}
	for _, field := range wantRequired {
		found := false
		for _, required := range schema.Items.Required {
			if required == field {
				found = true
			}
		}
		if !found {
			t.Errorf("schema missing required field %q", field)
		}
	}
	if _, err := os.Stat(datasetPath("finance/microfinance_banks.json")); err != nil {
		t.Fatalf("dataset missing: %v", err)
	}
}
