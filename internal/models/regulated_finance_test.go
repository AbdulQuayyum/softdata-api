package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type regulatedFinancialInstitutionRecord struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	CBNCode            string `json:"cbn_code,omitempty"`
	NIPCode            string `json:"nip_code,omitempty"`
	OfficialWebsiteURL string `json:"official_website_url,omitempty"`
	LogoURL            string `json:"logo_url,omitempty"`
	CountryCode        string `json:"country_code"`
}

func TestRegulatedFinanceDatasetContracts(t *testing.T) {
	cases := []struct {
		name  string
		path  string
		count int
	}{
		{"non-interest institutions", "finance/non_interest_institutions.json", 6},
		{"merchant banks", "finance/merchant_banks.json", 6},
		{"payment service banks", "finance/payment_service_banks.json", 5},
		{"financial holding companies", "finance/financial_holding_companies.json", 7},
		{"development finance institutions", "finance/development_finance_institutions.json", 8},
		{"primary mortgage institutions", "finance/primary_mortgage_institutions.json", 33},
	}
	idPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)
	cbnPattern := regexp.MustCompile(`^[0-9]{3}$`)
	nipPattern := regexp.MustCompile(`^[0-9]{6}$`)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := readTextBytes(t, datasetPath(tc.path))
			var records []regulatedFinancialInstitutionRecord
			if err := json.Unmarshal(raw, &records); err != nil {
				t.Fatalf("decode dataset: %v", err)
			}
			if len(records) != tc.count {
				t.Fatalf("record count = %d, want %d", len(records), tc.count)
			}
			ids := map[string]struct{}{}
			names := map[string]struct{}{}
			cbn := map[string]struct{}{}
			nip := map[string]struct{}{}
			previousName, previousID := "", ""
			for _, record := range records {
				if record.ID == "" || record.Name == "" || record.CountryCode != "NG" {
					t.Fatalf("invalid required fields: %#v", record)
				}
				if !idPattern.MatchString(record.ID) {
					t.Fatalf("invalid ID %q", record.ID)
				}
				if _, ok := ids[record.ID]; ok {
					t.Fatalf("duplicate ID %q", record.ID)
				}
				if _, ok := names[record.Name]; ok {
					t.Fatalf("duplicate name %q", record.Name)
				}
				ids[record.ID], names[record.Name] = struct{}{}, struct{}{}
				if strings.ToLower(record.Name) < strings.ToLower(previousName) || (strings.EqualFold(record.Name, previousName) && record.ID < previousID) {
					t.Fatalf("records are not sorted at %q", record.ID)
				}
				previousName, previousID = record.Name, record.ID
				if record.CBNCode != "" {
					if !cbnPattern.MatchString(record.CBNCode) {
						t.Fatalf("invalid CBN code %q", record.CBNCode)
					}
					if _, ok := cbn[record.CBNCode]; ok {
						t.Fatalf("duplicate CBN code %q", record.CBNCode)
					}
					cbn[record.CBNCode] = struct{}{}
				}
				if record.NIPCode != "" {
					if !nipPattern.MatchString(record.NIPCode) {
						t.Fatalf("invalid NIP code %q", record.NIPCode)
					}
					if _, ok := nip[record.NIPCode]; ok {
						t.Fatalf("duplicate NIP code %q", record.NIPCode)
					}
					nip[record.NIPCode] = struct{}{}
				}
				if record.OfficialWebsiteURL != "" {
					u, err := url.Parse(record.OfficialWebsiteURL)
					if err != nil || u.Scheme != "https" || u.Host == "" {
						t.Fatalf("invalid website URL %q", record.OfficialWebsiteURL)
					}
				}
			}
			var roundTrip []regulatedFinancialInstitutionRecord
			encoded, err := json.Marshal(records)
			if err != nil {
				t.Fatalf("encode dataset: %v", err)
			}
			if err := json.NewDecoder(bytes.NewReader(encoded)).Decode(&roundTrip); err != nil {
				t.Fatalf("round trip: %v", err)
			}
			if len(roundTrip) != tc.count {
				t.Fatalf("round-trip count = %d, want %d", len(roundTrip), tc.count)
			}
			ordered := append([]regulatedFinancialInstitutionRecord(nil), records...)
			sort.Slice(ordered, func(i, j int) bool {
				return strings.ToLower(ordered[i].Name)+"\x00"+ordered[i].ID < strings.ToLower(ordered[j].Name)+"\x00"+ordered[j].ID
			})
			if fmt.Sprint(records) != fmt.Sprint(ordered) {
				t.Fatal("dataset ordering is not deterministic")
			}
		})
	}
}

func TestFinancialHoldingCompanyDoesNotExposeBankCodes(t *testing.T) {
	var company FinancialHoldingCompany
	if err := json.Unmarshal([]byte(`{"id":"gtb-hold-co","name":"GTB Hold Co","country_code":"NG"}`), &company); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(company)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("cbn_code")) || bytes.Contains(raw, []byte("nip_code")) {
		t.Fatalf("holding company exposed bank codes: %s", raw)
	}
}
