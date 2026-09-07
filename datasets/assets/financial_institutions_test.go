package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFinancialInstitutionLogosAreEmbeddedAndAttributed(t *testing.T) {
	attribution, err := os.ReadFile(filepath.Join("financial-institutions", "ng", "ATTRIBUTION.md"))
	if err != nil {
		t.Fatal(err)
	}
	var count int
	err = filepath.Walk(filepath.Join("financial-institutions", "ng"), func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || filepath.Ext(path) != ".png" {
			return nil
		}
		count++
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")) {
			t.Errorf("%s is not PNG", path)
		}
		if config, err := png.DecodeConfig(bytes.NewReader(data)); err != nil || config.Width <= 0 || config.Height <= 0 {
			t.Errorf("%s has invalid dimensions: %v", path, err)
		}
		sum := sha256.Sum256(data)
		if !strings.Contains(string(attribution), hex.EncodeToString(sum[:])) {
			t.Errorf("%s hash is missing from attribution", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 422 {
		t.Fatalf("embedded logo count = %d, want 422", count)
	}
}

func TestMicrofinanceBankLogoURLsResolve(t *testing.T) {
	data, err := os.ReadFile("../finance/microfinance_banks.json")
	if err != nil {
		t.Fatal(err)
	}
	var records []struct {
		ID      string `json:"id"`
		LogoURL string `json:"logo_url"`
	}
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatal(err)
	}
	for _, record := range records {
		if record.LogoURL == "" {
			continue
		}
		logo, err := FinancialInstitutionLogo("microfinance-banks", record.ID, "png")
		if err != nil || len(logo) == 0 {
			t.Errorf("%s logo unavailable: %v", record.ID, err)
		}
	}
}

func TestFinancialInstitutionLogoURLsResolve(t *testing.T) {
	paths := []string{
		"../finance/non_interest_institutions.json",
		"../finance/merchant_banks.json",
		"../finance/payment_service_banks.json",
		"../finance/financial_holding_companies.json",
		"../finance/development_finance_institutions.json",
		"../finance/primary_mortgage_institutions.json",
		"../finance/payment_service_providers.json",
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var records []struct{ ID, LogoURL string }
		if err := json.Unmarshal(data, &records); err != nil {
			t.Fatal(err)
		}
		for _, record := range records {
			if record.LogoURL == "" {
				continue
			}
			parts := strings.Split(record.LogoURL, "/")
			if len(parts) < 2 || !strings.HasSuffix(record.LogoURL, ".png") {
				t.Errorf("%s has malformed logo URL %q", record.ID, record.LogoURL)
				continue
			}
			category := parts[len(parts)-2]
			logo, err := FinancialInstitutionLogo(category, record.ID, "png")
			if err != nil || len(logo) == 0 {
				t.Errorf("%s logo unavailable: %v", record.ID, err)
			}
		}
	}
}

func TestMissingLogoInventoryIsDocumented(t *testing.T) {
	audit, err := os.ReadFile(filepath.Join("financial-institutions", "ng", "ACQUISITION_AUDIT.md"))
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{
		"../finance/non_interest_institutions.json",
		"../finance/merchant_banks.json",
		"../finance/payment_service_banks.json",
		"../finance/financial_holding_companies.json",
		"../finance/development_finance_institutions.json",
		"../finance/primary_mortgage_institutions.json",
	}
	missingIDs := map[string]bool{}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var records []struct {
			ID      string `json:"id"`
			LogoURL string `json:"logo_url"`
		}
		if err := json.Unmarshal(data, &records); err != nil {
			t.Fatal(err)
		}
		for _, record := range records {
			if record.LogoURL == "" {
				missingIDs[record.ID] = true
				if !bytes.Contains(audit, []byte("| `"+record.ID+"` |")) {
					t.Errorf("missing logo %s is absent from acquisition audit", record.ID)
				}
			}
		}
	}
	wantMissing := map[string]bool{}
	if len(missingIDs) != len(wantMissing) {
		t.Fatalf("missing logo count = %d, want %d", len(missingIDs), len(wantMissing))
	}
	for id := range wantMissing {
		if !missingIDs[id] {
			t.Errorf("unexpected missing logo inventory: %s", id)
		}
	}
}

func TestMicrofinancePhase2CReconciliationIsComplete(t *testing.T) {
	data, err := os.ReadFile("../metadata/finance/microfinance_banks_phase2c_logo_reconciliation.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		SourceFileCount int            `json:"source_file_count"`
		Classifications map[string]int `json:"classifications"`
		Records         []struct {
			Classification string `json:"classification"`
		} `json:"records"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{
		"accepted_exact_name":                true,
		"accepted_alias":                     true,
		"accepted_nip_code_supported":        true,
		"accepted_cbn_code_supported":        true,
		"accepted_successor":                 true,
		"accepted_same_entity_cross_dataset": true,
		"preserved_existing_byte_identical":  true,
		"preserved_existing_stronger_source": true,
		"rejected_revoked":                   true,
		"rejected_obsolete":                  true,
		"rejected_distinct_identity":         true,
		"rejected_code_conflict":             true,
		"rejected_no_active_dataset_match":   true,
		"ambiguous_manual_review":            true,
	}
	if manifest.SourceFileCount != 315 || len(manifest.Records) != 315 {
		t.Fatalf("phase 2C records = %d/%d, want 315/315", manifest.SourceFileCount, len(manifest.Records))
	}
	var total int
	for classification, count := range manifest.Classifications {
		if !allowed[classification] {
			t.Fatalf("unexpected phase 2C classification %q", classification)
		}
		total += count
	}
	if total != 315 {
		t.Fatalf("phase 2C classification total = %d, want 315", total)
	}
}

func TestCommercialPhase2DReconciliationIsComplete(t *testing.T) {
	data, err := os.ReadFile("../metadata/finance/cross_dataset_finance_logo_phase2d_reconciliation.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		SourceFileCount int            `json:"source_file_count"`
		Classifications map[string]int `json:"classifications"`
		Records         []struct {
			Classification string `json:"classification"`
		} `json:"records"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{
		"preserved_existing_byte_identical":  true,
		"preserved_existing_stronger_source": true,
		"rejected_obsolete":                  true,
		"rejected_wrong_category":            true,
		"rejected_distinct_identity":         true,
		"rejected_no_active_dataset_match":   true,
		"ambiguous_manual_review":            true,
	}
	if manifest.SourceFileCount != 32 || len(manifest.Records) != 32 {
		t.Fatalf("phase 2D records = %d/%d, want 32/32", manifest.SourceFileCount, len(manifest.Records))
	}
	var total int
	for classification, count := range manifest.Classifications {
		if !allowed[classification] {
			t.Fatalf("unexpected phase 2D classification %q", classification)
		}
		total += count
	}
	if total != 32 {
		t.Fatalf("phase 2D classification total = %d, want 32", total)
	}
}

func TestCrossDatasetFinanceLogoArchiveReconciliationIsComplete(t *testing.T) {
	data, err := os.ReadFile("../metadata/finance/cross_dataset_finance_logo_reconciliation.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		SourceFileCount      int            `json:"source_file_count"`
		ClassificationCounts map[string]int `json:"classification_counts"`
		Records              []struct {
			SourcePath     string `json:"source_path"`
			Classification string `json:"classification"`
		} `json:"records"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.SourceFileCount != 610 || len(manifest.Records) != 610 {
		t.Fatalf("archive reconciliation records = %d/%d, want 610/610", manifest.SourceFileCount, len(manifest.Records))
	}
	seen := make(map[string]bool, len(manifest.Records))
	var total int
	for _, record := range manifest.Records {
		if record.SourcePath == "" || seen[record.SourcePath] {
			t.Fatalf("duplicate or empty archive source path %q", record.SourcePath)
		}
		seen[record.SourcePath] = true
		if record.Classification == "" {
			t.Fatal("archive record has no classification")
		}
	}
	for _, count := range manifest.ClassificationCounts {
		total += count
	}
	if total != 610 {
		t.Fatalf("archive classification total = %d, want 610", total)
	}
}
