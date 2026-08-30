package models

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"
)

type paymentServiceProviderMetadata struct {
	DatasetKey   string          `json:"dataset_key"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	CountryCode  string          `json:"country_code"`
	DatasetGroup string          `json:"dataset_group"`
	Format       string          `json:"format"`
	RelativePath string          `json:"relative_path"`
	SchemaPath   string          `json:"schema_path"`
	RecordCount  int             `json:"record_count"`
	Version      string          `json:"version"`
	LicenseID    string          `json:"license_id"`
	LicenseURL   string          `json:"license_url"`
	Methodology  string          `json:"methodology"`
	Sources      []datasetSource `json:"sources"`
	VerifiedAt   string          `json:"verified_at"`
}

type paymentServiceProviderSchema struct {
	Schema      string          `json:"$schema"`
	ID          string          `json:"$id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Type        string          `json:"type"`
	MinItems    int             `json:"minItems"`
	MaxItems    int             `json:"maxItems"`
	UniqueItems bool            `json:"uniqueItems"`
	Items       json.RawMessage `json:"items"`
}

func TestPaymentServiceProvidersDatasetMatchesApprovedSnapshot(t *testing.T) {
	providers := loadPaymentServiceProvidersDataset(t)
	if providers == nil {
		t.Fatal("decoded dataset is nil")
	}
	if got := len(providers); got != 255 {
		t.Fatalf("unexpected record count: got %d want 255", got)
	}

	wantCounts := map[string]int{
		"mobile_money_operator":               17,
		"switching_and_processing_company":    19,
		"payment_solution_service_provider":   108,
		"payment_terminal_service_provider":   47,
		"super_agent":                         61,
		"payment_service_holding_company":     1,
		"payment_terminal_service_aggregator": 2,
	}
	categoryCounts := make(map[string]int, len(wantCounts))
	seenIDs := make(map[string]struct{}, len(providers))
	seenPairs := make(map[string]struct{}, len(providers))
	idPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)
	order := map[string]int{
		"mobile_money_operator":               0,
		"switching_and_processing_company":    1,
		"payment_solution_service_provider":   2,
		"payment_terminal_service_provider":   3,
		"super_agent":                         4,
		"payment_service_holding_company":     5,
		"payment_terminal_service_aggregator": 6,
	}
	prevOrder := -1
	prevName := ""
	prevID := ""

	for i, provider := range providers {
		if provider.ID == "" || provider.Name == "" || provider.InstitutionType == "" || provider.CountryCode == "" {
			t.Fatalf("record %d has empty required field: %#v", i, provider)
		}
		if provider.CountryCode != "NG" {
			t.Fatalf("record %d has unexpected country code %q", i, provider.CountryCode)
		}
		if !idPattern.MatchString(provider.ID) {
			t.Fatalf("record %d has invalid id %q", i, provider.ID)
		}
		typePrefix := strings.ReplaceAll(provider.InstitutionType, "_", "-")
		if !strings.HasPrefix(provider.ID, typePrefix+"-") {
			t.Fatalf("record %d id %q does not start with type prefix %q", i, provider.ID, typePrefix)
		}
		if _, ok := wantCounts[provider.InstitutionType]; !ok {
			t.Fatalf("record %d has unsupported institution type %q", i, provider.InstitutionType)
		}
		categoryCounts[provider.InstitutionType]++
		if _, ok := seenIDs[provider.ID]; ok {
			t.Fatalf("duplicate id found: %q", provider.ID)
		}
		seenIDs[provider.ID] = struct{}{}
		pairKey := provider.InstitutionType + "\x00" + provider.Name
		if _, ok := seenPairs[pairKey]; ok {
			t.Fatalf("duplicate institution type/name pair found: %q / %q", provider.InstitutionType, provider.Name)
		}
		seenPairs[pairKey] = struct{}{}
		if strings.Contains(strings.ToLower(provider.Name), "formerly") || strings.Contains(strings.ToLower(provider.Name), "fromerly") {
			t.Fatalf("former-name annotation leaked into public name: %q", provider.Name)
		}
		currentOrder := order[provider.InstitutionType]
		if currentOrder < prevOrder || (currentOrder == prevOrder && (strings.ToLower(provider.Name) < strings.ToLower(prevName) || (strings.ToLower(provider.Name) == strings.ToLower(prevName) && provider.ID < prevID))) {
			t.Fatalf("dataset ordering is not deterministic at record %d: %#v", i, provider)
		}
		prevOrder = currentOrder
		prevName = provider.Name
		prevID = provider.ID
	}

	for k, want := range wantCounts {
		if got := categoryCounts[k]; got != want {
			t.Fatalf("unexpected count for %s: got %d want %d", k, got, want)
		}
	}

	mustContain := map[string]string{
		"mobile-money-operator-kongapay-technologies-limited":                   "KongaPay Technologies Limited",
		"mobile-money-operator-nownow-digital-systems-limited":                  "NowNow Digital Systems Limited",
		"mobile-money-operator-opay-digital-services-limited":                   "Opay Digital Services Limited",
		"switching-and-processing-company-zone-payment-network-limited":         "Zone Payment Network Limited",
		"payment-solution-service-provider-cyberpay-limited":                    "Cyberpay Limited",
		"payment-solution-service-provider-montra-technology-solutions-limited": "Montra Technology Solutions Limited",
		"payment-solution-service-provider-nomba-financial-services-limited":    "Nomba Financial Services Limited",
		"payment-solution-service-provider-payfixy-nigeria-limited":             "Payfixy Nigeria Limited",
		"payment-solution-service-provider-swwipe-financial-services-limited":   "SWWIPE Financial Services Limited",
		"payment-terminal-service-provider-funds-konnect-limited":               "Funds Konnect Limited",
		"super-agent-crowd-force-limited":                                       "Crowd Force Limited",
	}
	byID := make(map[string]PaymentServiceProvider, len(providers))
	for _, provider := range providers {
		byID[provider.ID] = provider
	}
	for id, wantName := range mustContain {
		provider, ok := byID[id]
		if !ok {
			t.Fatalf("missing expected provider id %q", id)
		}
		if provider.Name != wantName {
			t.Fatalf("unexpected name for %q: got %q want %q", id, provider.Name, wantName)
		}
	}
}

func TestPaymentServiceProvidersMetadataSchemaAndVerifiedAt(t *testing.T) {
	metadata := loadPaymentServiceProvidersMetadata(t)
	if metadata.DatasetKey != "ng-payment-service-providers" {
		t.Fatalf("unexpected dataset key: %q", metadata.DatasetKey)
	}
	if metadata.Title != "Nigeria Payment Service Providers" {
		t.Fatalf("unexpected title: %q", metadata.Title)
	}
	if metadata.DatasetGroup != "finance" {
		t.Fatalf("unexpected dataset group: %q", metadata.DatasetGroup)
	}
	if metadata.Format != "json" {
		t.Fatalf("unexpected format: %q", metadata.Format)
	}
	if metadata.CountryCode != "NG" {
		t.Fatalf("unexpected country code: %q", metadata.CountryCode)
	}
	if metadata.RecordCount != 255 {
		t.Fatalf("unexpected record count: %d", metadata.RecordCount)
	}
	if metadata.RelativePath != "finance/payment_service_providers.json" {
		t.Fatalf("unexpected relative path: %q", metadata.RelativePath)
	}
	if metadata.SchemaPath != "schemas/finance/payment_service_providers.schema.json" {
		t.Fatalf("unexpected schema path: %q", metadata.SchemaPath)
	}
	if metadata.LicenseID != "CC-BY-4.0" || metadata.LicenseURL != "https://creativecommons.org/licenses/by/4.0/" {
		t.Fatalf("unexpected license metadata: %#v", metadata)
	}
	if metadata.Version != "1.0.0" {
		t.Fatalf("unexpected version: %q", metadata.Version)
	}
	if metadata.VerifiedAt == "" {
		t.Fatal("verified_at is empty")
	}
	verifiedAt, err := time.Parse("2006-01-02", metadata.VerifiedAt)
	if err != nil {
		t.Fatalf("verified_at is not a valid date: %v", err)
	}
	nowUTC := time.Now().UTC()
	if verifiedAt.After(nowUTC) {
		t.Fatalf("verified_at is in the future: %s > %s", verifiedAt.Format("2006-01-02"), nowUTC.Format("2006-01-02"))
	}
	if metadata.VerifiedAt != "2026-08-29" {
		t.Fatalf("unexpected verified_at: %q", metadata.VerifiedAt)
	}
	if len(metadata.Sources) != 1 {
		t.Fatalf("unexpected source count: %d", len(metadata.Sources))
	}
	if metadata.Sources[0].Organization != "Central Bank of Nigeria" || metadata.Sources[0].Title != "Payment Service Providers" || metadata.Sources[0].URL != "https://www.cbn.gov.ng/PaymentsSystem/PSPs.html" {
		t.Fatalf("unexpected source metadata: %#v", metadata.Sources[0])
	}
	if metadata.Sources[0].AccessedAt != "2026-08-29" {
		t.Fatalf("unexpected source access date: %q", metadata.Sources[0].AccessedAt)
	}
	for _, snippet := range []string{
		"three times with consistent results",
		"mutable",
		"do not establish a universal institution identifier",
		"excluded card schemes, payment schemes, clearing-house content, IMTOs and adjacent page material",
	} {
		if !strings.Contains(metadata.Methodology, snippet) {
			t.Fatalf("methodology missing %q: %q", snippet, metadata.Methodology)
		}
	}

	schema := loadPaymentServiceProvidersSchema(t)
	if schema.Schema != "https://json-schema.org/draft/2020-12/schema" {
		t.Fatalf("unexpected schema dialect: %q", schema.Schema)
	}
	if schema.ID != "https://softdata-api.local/schemas/finance/payment_service_providers.schema.json" {
		t.Fatalf("unexpected schema id: %q", schema.ID)
	}
	if schema.Type != "array" || schema.MinItems != 255 || schema.MaxItems != 255 || !schema.UniqueItems {
		t.Fatalf("unexpected schema constraints: %#v", schema)
	}
	if schema.Title != "Nigeria Payment Service Providers" {
		t.Fatalf("unexpected schema title: %q", schema.Title)
	}
	if schema.Description == "" {
		t.Fatal("schema description is empty")
	}

	var items map[string]any
	if err := json.Unmarshal(schema.Items, &items); err != nil {
		t.Fatalf("decode schema items: %v", err)
	}
	if got := items["type"]; got != "object" {
		t.Fatalf("unexpected schema item type: %#v", got)
	}
	if got := items["additionalProperties"]; got != false {
		t.Fatalf("unexpected additionalProperties: %#v", got)
	}
}

func TestPaymentServiceProvidersPunctuationAndFormerNameAudit(t *testing.T) {
	providers := loadPaymentServiceProvidersDataset(t)
	punctuated := []string{
		"Funds And Electronic Transfer (FETS) Limited",
		"Terra Switching & Processing Company Limited",
		"Clane Company Nig. Ltd.",
		"Onepipe.Io Services Ltd",
		"Saanapay Corporate Investments Management Limited (SAANACORP)",
		"Nigerian Postal Service (NIPOST)",
		"Y'ello Digital Financial Services..",
		"Swift Link-NZ Global Services Ltd.",
	}
	providerNames := make(map[string]struct{}, len(providers))
	for _, provider := range providers {
		providerNames[provider.Name] = struct{}{}
	}
	for _, want := range punctuated {
		if _, ok := providerNames[want]; !ok {
			t.Fatalf("missing punctuated public name %q", want)
		}
	}
	for _, forbidden := range []string{"Formerly", "Fromerly"} {
		for name := range providerNames {
			if strings.Contains(name, forbidden) {
				t.Fatalf("former-name annotation leaked into public name %q", name)
			}
		}
	}
}

func loadPaymentServiceProvidersDataset(t *testing.T) []PaymentServiceProvider {
	t.Helper()
	var providers []PaymentServiceProvider
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("finance/payment_service_providers.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&providers); err != nil {
		t.Fatalf("decode payment service providers dataset: %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("payment service providers dataset contains trailing JSON")
	}
	return providers
}

func loadPaymentServiceProvidersMetadata(t *testing.T) paymentServiceProviderMetadata {
	t.Helper()
	var metadata paymentServiceProviderMetadata
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("metadata/finance/payment_service_providers.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&metadata); err != nil {
		t.Fatalf("decode payment service providers metadata: %v", err)
	}
	return metadata
}

func loadPaymentServiceProvidersSchema(t *testing.T) paymentServiceProviderSchema {
	t.Helper()
	var schema paymentServiceProviderSchema
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("schemas/finance/payment_service_providers.schema.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&schema); err != nil {
		t.Fatalf("decode payment service providers schema: %v", err)
	}
	return schema
}
