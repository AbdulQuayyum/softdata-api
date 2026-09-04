package models

import (
	"bytes"
	"encoding/json"
	"regexp"
	"testing"
)

func TestCommercialBanksIdentifierContract(t *testing.T) {
	var banks []CommercialBank
	if err := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("finance/commercial_banks.json")))).Decode(&banks); err != nil {
		t.Fatalf("decode commercial banks: %v", err)
	}
	if len(banks) != 28 {
		t.Fatalf("commercial bank count = %d, want 28", len(banks))
	}

	expected := map[string][2]string{
		"access-bank": {"044", "000014"}, "alpha-morgan-bank": {"", "000041"},
		"citibank-nigeria": {"023", "000009"}, "ecobank-nigeria": {"050", "000010"},
		"fidelity-bank": {"070", "000007"}, "first-bank-of-nigeria": {"011", "000016"},
		"first-city-monument-bank": {"214", "000003"}, "globus-bank": {"103", "000027"},
		"guaranty-trust-bank": {"058", "000013"}, "keystone-bank": {"082", "000002"},
		"nova-bank": {"", ""}, "optimus-bank": {"107", "000036"},
		"parallex-bank": {"104", "000030"}, "polaris-bank": {"076", "000008"},
		"premium-trust-bank": {"105", "000031"}, "providus-bank": {"101", "000021"},
		"signature-bank": {"", "000034"}, "stanbic-ibtc-bank": {"221", "000012"},
		"standard-chartered-bank": {"068", ""}, "sterling-bank": {"232", "000001"},
		"suntrust-bank": {"100", ""}, "tatum-bank": {"109", "000042"},
		"titan-trust-bank": {"102", "000025"}, "union-bank": {"032", "000018"},
		"united-bank-for-africa": {"033", "000004"}, "unity-bank": {"215", "000011"},
		"wema-bank": {"035", "000017"}, "zenith-bank": {"057", "000015"},
	}
	cbnPattern := regexp.MustCompile(`^[0-9]{3}$`)
	nipPattern := regexp.MustCompile(`^[0-9]{6}$`)
	cbnSeen := map[string]string{}
	nipSeen := map[string]string{}
	for _, bank := range banks {
		want, ok := expected[bank.ID]
		if !ok {
			t.Fatalf("unexpected bank ID %q", bank.ID)
		}
		if bank.CBNCode != want[0] || bank.NIPCode != want[1] {
			t.Errorf("%s identifiers = (%q, %q), want (%q, %q)", bank.ID, bank.CBNCode, bank.NIPCode, want[0], want[1])
		}
		if bank.CBNCode != "" {
			if !cbnPattern.MatchString(bank.CBNCode) {
				t.Errorf("%s cbn_code is not three digits: %q", bank.ID, bank.CBNCode)
			}
			if previous, exists := cbnSeen[bank.CBNCode]; exists {
				t.Errorf("duplicate cbn_code %q for %s and %s", bank.CBNCode, previous, bank.ID)
			}
			cbnSeen[bank.CBNCode] = bank.ID
		}
		if bank.NIPCode != "" {
			if !nipPattern.MatchString(bank.NIPCode) {
				t.Errorf("%s nip_code is not six digits: %q", bank.ID, bank.NIPCode)
			}
			if previous, exists := nipSeen[bank.NIPCode]; exists {
				t.Errorf("duplicate nip_code %q for %s and %s", bank.NIPCode, previous, bank.ID)
			}
			nipSeen[bank.NIPCode] = bank.ID
		}
		if bank.OfficialWebsiteURL == "" || bank.LogoURL == "" || bank.CountryCode != "NG" {
			t.Errorf("%s lost an existing core/logo field", bank.ID)
		}
	}
	if len(expected) != len(banks) {
		t.Fatalf("expected mapping count = %d, dataset count = %d", len(expected), len(banks))
	}
	if len(cbnSeen) != 25 || len(nipSeen) != 25 {
		t.Fatalf("identifier coverage = cbn %d, nip %d; want 25 and 25", len(cbnSeen), len(nipSeen))
	}
}

func TestCommercialBankIdentifierLeadingZerosSurviveRoundTrip(t *testing.T) {
	var banks []CommercialBank
	if err := json.Unmarshal(readTextBytes(t, datasetPath("finance/commercial_banks.json")), &banks); err != nil {
		t.Fatalf("decode commercial banks: %v", err)
	}
	raw, err := json.Marshal(banks)
	if err != nil {
		t.Fatalf("encode commercial banks: %v", err)
	}
	if !bytes.Contains(raw, []byte(`"nip_code":"000014"`)) || !bytes.Contains(raw, []byte(`"nip_code":"000002"`)) {
		t.Fatal("leading-zero NIP codes were not preserved as strings")
	}
}
