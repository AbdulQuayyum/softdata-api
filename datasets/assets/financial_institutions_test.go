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
	if count != 393 {
		t.Fatalf("embedded logo count = %d, want 393", count)
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
