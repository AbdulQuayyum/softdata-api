package models

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"
)

type imtoMetadata struct {
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

type imtoSchema struct {
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

func TestInternationalMoneyTransferOperatorsDatasetMatchesApprovedSnapshot(t *testing.T) {
	operators := loadInternationalMoneyTransferOperatorsDataset(t)
	if operators == nil {
		t.Fatal("decoded dataset is nil")
	}
	if got := len(operators); got != 108 {
		t.Fatalf("unexpected record count: got %d want 108", got)
	}

	idPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	seenIDs := make(map[string]struct{}, len(operators))
	seenNames := make(map[string]struct{}, len(operators))
	prevName := ""
	prevID := ""
	nameCounts := map[string]int{
		"NOUVEAU MOBILE LIMITED":                             0,
		"OLIVE MONIES EXPRESS LIMITED":                       0,
		"OLIVE MONIES EXPRESS LIMITEDNOUVEAU MOBILE LIMITED": 0,
	}
	formerNames := []string{
		"FLUTTERWAVE TECHNOLOGY SOLUTIONS LTD",
		"STERLING CURRENCY EXCHANGE LTD",
		"MULTIGATE GROUP HOLDING LIMITED",
		"PAGATECH LIMITED",
		"VTNETWORK LIMITED",
		"ALALAMIYA EXCHANGE LIMITED",
		"FIEM GROUP LLC DBA",
	}
	punctuatedNames := []string{
		"CENTREXCARD LIMITED T/A (TransferBoss)",
		"CHIME INC (SENDWAVE)",
		"GABTRANS UK LIMITED/MONEY TO LIMITED",
		"LEADREMIT INC (FIRST APPLE)",
		"RIGHTCARD PAYMENT SERVICES LTD (A LEMFI BRAND COMPANY)",
		"TERRA PAYMENT SERVICES (UK) LIMITED",
		"TRANSFER ZERO MONEY TRANSFER E.P., S.A.",
		"VOLOPA FINANCIAL SERVICES (SCOTLAND) LIMITED",
	}
	for i, operator := range operators {
		if operator.ID == "" || operator.Name == "" {
			t.Fatalf("record %d has empty required field: %#v", i, operator)
		}
		if operator.ID == "country_code" || operator.Name == "country_code" {
			t.Fatalf("record %d leaked a country_code field: %#v", i, operator)
		}
		if !idPattern.MatchString(operator.ID) {
			t.Fatalf("record %d has invalid id %q", i, operator.ID)
		}
		if expected := slugifyIMTOName(operator.Name); operator.ID != expected {
			t.Fatalf("record %d id does not match slugified name: got %q want %q", i, operator.ID, expected)
		}
		if _, ok := seenIDs[operator.ID]; ok {
			t.Fatalf("duplicate id found: %q", operator.ID)
		}
		seenIDs[operator.ID] = struct{}{}

		normalizedName := normalizeIMTOWords(operator.Name)
		if normalizedName == "" {
			t.Fatalf("record %d has empty normalized name: %#v", i, operator)
		}
		if _, ok := seenNames[normalizedName]; ok {
			t.Fatalf("duplicate name found after whitespace normalization: %q", normalizedName)
		}
		seenNames[normalizedName] = struct{}{}

		if strings.Contains(strings.ToLower(operator.Name), "formerly") {
			t.Fatalf("former-name annotation leaked into public name: %q", operator.Name)
		}

		if prevName != "" {
			prevLower := strings.ToLower(prevName)
			currLower := strings.ToLower(operator.Name)
			if prevLower > currLower || (prevLower == currLower && prevID > operator.ID) {
				t.Fatalf("dataset ordering is not deterministic at record %d: %#v before %#v", i, prevName, operator)
			}
		}
		prevName = operator.Name
		prevID = operator.ID

		if _, ok := nameCounts[operator.Name]; ok {
			nameCounts[operator.Name]++
		}
	}

	if nameCounts["NOUVEAU MOBILE LIMITED"] != 1 {
		t.Fatalf("unexpected NOUVEAU MOBILE LIMITED count: %d", nameCounts["NOUVEAU MOBILE LIMITED"])
	}
	if nameCounts["OLIVE MONIES EXPRESS LIMITED"] != 1 {
		t.Fatalf("unexpected OLIVE MONIES EXPRESS LIMITED count: %d", nameCounts["OLIVE MONIES EXPRESS LIMITED"])
	}
	if nameCounts["OLIVE MONIES EXPRESS LIMITEDNOUVEAU MOBILE LIMITED"] != 0 {
		t.Fatalf("malformed concatenated source row leaked into public dataset")
	}

	for _, forbidden := range formerNames {
		for _, operator := range operators {
			if operator.Name == forbidden {
				t.Fatalf("former name unexpectedly exposed as current name: %q", forbidden)
			}
		}
	}

	for _, want := range punctuatedNames {
		if !containsIMTOName(operators, want) {
			t.Fatalf("missing punctuated public name %q", want)
		}
	}

	serialized, err := json.Marshal(operators[0])
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(serialized, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if len(got) != 2 || got["id"] == nil || got["name"] == nil {
		t.Fatalf("unexpected serialized public fields: %#v", got)
	}
	if _, ok := got["country_code"]; ok {
		t.Fatalf("country_code should not be serialized: %#v", got)
	}
}

func TestInternationalMoneyTransferOperatorsMetadataSchemaAndVerifiedAt(t *testing.T) {
	metadata := loadInternationalMoneyTransferOperatorsMetadata(t)
	if metadata.DatasetKey != "ng-international-money-transfer-operators" {
		t.Fatalf("unexpected dataset key: %q", metadata.DatasetKey)
	}
	if metadata.Title != "Nigeria International Money Transfer Operators" {
		t.Fatalf("unexpected title: %q", metadata.Title)
	}
	if metadata.DatasetGroup != "finance" {
		t.Fatalf("unexpected dataset group: %q", metadata.DatasetGroup)
	}
	if metadata.CountryCode != "NG" {
		t.Fatalf("unexpected country code: %q", metadata.CountryCode)
	}
	if metadata.Format != "json" {
		t.Fatalf("unexpected format: %q", metadata.Format)
	}
	if metadata.RecordCount != 108 {
		t.Fatalf("unexpected record count: %d", metadata.RecordCount)
	}
	if metadata.RelativePath != "finance/international_money_transfer_operators.json" {
		t.Fatalf("unexpected relative path: %q", metadata.RelativePath)
	}
	if metadata.SchemaPath != "schemas/finance/international_money_transfer_operators.schema.json" {
		t.Fatalf("unexpected schema path: %q", metadata.SchemaPath)
	}
	if metadata.Version != "1.0.0" {
		t.Fatalf("unexpected version: %q", metadata.Version)
	}
	if metadata.LicenseID != "CC-BY-4.0" || metadata.LicenseURL != "https://creativecommons.org/licenses/by/4.0/" {
		t.Fatalf("unexpected license metadata: %#v", metadata)
	}
	if metadata.VerifiedAt != "2026-08-30" {
		t.Fatalf("unexpected verified_at: %q", metadata.VerifiedAt)
	}
	verifiedAt, err := time.Parse("2006-01-02", metadata.VerifiedAt)
	if err != nil {
		t.Fatalf("verified_at is not a valid date: %v", err)
	}
	if verifiedAt.After(time.Now().UTC()) {
		t.Fatalf("verified_at is in the future: %s", metadata.VerifiedAt)
	}
	if len(metadata.Sources) != 1 {
		t.Fatalf("unexpected source count: %d", len(metadata.Sources))
	}
	source := metadata.Sources[0]
	if source.Organization != "Central Bank of Nigeria" || source.Title != "International Money Transfer Operators" || source.URL != "https://www.cbn.gov.ng/PaymentsSystem/InternationalMoneyTransferOperators.html" {
		t.Fatalf("unexpected source metadata: %#v", source)
	}
	if source.AccessedAt != "2026-08-30" {
		t.Fatalf("unexpected source access date: %q", source.AccessedAt)
	}
	for _, snippet := range []string{
		"retrieved three times with consistent 108-row results",
		"concatenation defect at SN 63",
		"Addresses, incorporation country, licence numbers, status labels and corridor assumptions were not inferred",
		"No downloadable or alternate machine-readable register was observed",
		"no visible last-updated date was present",
		"does not guarantee current availability",
	} {
		if !strings.Contains(metadata.Methodology, snippet) {
			t.Fatalf("methodology missing %q: %q", snippet, metadata.Methodology)
		}
	}
	if strings.Contains(metadata.Methodology, "CC BY 4.0") {
		t.Fatalf("methodology should not claim the CBN source is CC BY 4.0: %q", metadata.Methodology)
	}

	schema := loadInternationalMoneyTransferOperatorsSchema(t)
	if schema.Schema != "https://json-schema.org/draft/2020-12/schema" {
		t.Fatalf("unexpected schema dialect: %q", schema.Schema)
	}
	if schema.ID != "https://softdata-api.local/schemas/finance/international_money_transfer_operators.schema.json" {
		t.Fatalf("unexpected schema id: %q", schema.ID)
	}
	if schema.Title != "Nigeria International Money Transfer Operators" {
		t.Fatalf("unexpected schema title: %q", schema.Title)
	}
	if schema.Type != "array" || schema.MinItems != 108 || schema.MaxItems != 108 || !schema.UniqueItems {
		t.Fatalf("unexpected schema constraints: %#v", schema)
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
	required, _ := items["required"].([]any)
	if len(required) != 2 {
		t.Fatalf("unexpected required field count: %#v", items["required"])
	}
}

func TestInternationalMoneyTransferOperatorsFormerNameNormalizationAudit(t *testing.T) {
	operators := loadInternationalMoneyTransferOperatorsDataset(t)
	nameSet := make(map[string]struct{}, len(operators))
	for _, operator := range operators {
		nameSet[normalizeIMTOWords(operator.Name)] = struct{}{}
	}

	for current, former := range map[string]string{
		"FLUTTERWAVE TECH PAYMENTS LIMITED": "FLUTTERWAVE TECHNOLOGY SOLUTIONS LTD",
		"LIGHTWAY FINANCE LIMITED":          "STERLING CURRENCY EXCHANGE LTD",
		"MULTIGATE NETWORK PAY INC":         "MULTIGATE GROUP HOLDING LIMITED",
		"PAGA REMIT LIMITED":                "PAGATECH LIMITED",
		"PAYMENT PROCESSING INC.":           "VTNETWORK LIMITED",
		"PEARL PAYMENTS LIMITED":            "ALALAMIYA EXCHANGE LIMITED",
		"PING EXPRESS":                      "FIEM GROUP LLC DBA",
	} {
		if _, ok := nameSet[current]; !ok {
			t.Fatalf("missing current name %q", current)
		}
		if _, ok := nameSet[former]; ok {
			t.Fatalf("former name leaked into public dataset: %q", former)
		}
	}
}

func loadInternationalMoneyTransferOperatorsDataset(t *testing.T) []InternationalMoneyTransferOperator {
	t.Helper()
	var operators []InternationalMoneyTransferOperator
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("finance/international_money_transfer_operators.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&operators); err != nil {
		t.Fatalf("decode IMTO dataset: %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("IMTO dataset contains trailing JSON")
	}
	return operators
}

func loadInternationalMoneyTransferOperatorsMetadata(t *testing.T) imtoMetadata {
	t.Helper()
	var metadata imtoMetadata
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("metadata/finance/international_money_transfer_operators.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&metadata); err != nil {
		t.Fatalf("decode IMTO metadata: %v", err)
	}
	return metadata
}

func loadInternationalMoneyTransferOperatorsSchema(t *testing.T) imtoSchema {
	t.Helper()
	var schema imtoSchema
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("schemas/finance/international_money_transfer_operators.schema.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&schema); err != nil {
		t.Fatalf("decode IMTO schema: %v", err)
	}
	return schema
}

func normalizeIMTOWords(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func slugifyIMTOName(value string) string {
	value = strings.ToLower(normalizeIMTOWords(value))
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	value = regexp.MustCompile(`-{2,}`).ReplaceAllString(value, "-")
	return value
}

func containsIMTOName(operators []InternationalMoneyTransferOperator, want string) bool {
	for _, operator := range operators {
		if operator.Name == want {
			return true
		}
	}
	return false
}
