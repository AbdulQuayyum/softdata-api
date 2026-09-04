package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestOpenAPIDocumentsLocalGovernmentUnitPaths(t *testing.T) {
	doc, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(doc)

	listBlock := pathBlock(t, text, "/v1/geography/lgas")
	detailBlock := pathBlock(t, text, "/v1/geography/lgas/{lga_id}")

	requireContains(t, text, "    StateIDQuery:")
	requireContains(t, text, "    LGAID:")
	requireContains(t, text, "    LocalGovernmentUnit:")
	requireContains(t, text, "    LocalGovernmentUnitListResponse:")
	requireContains(t, text, "    LocalGovernmentUnitResponse:")

	requireContains(t, listBlock, "get:")
	requireContains(t, listBlock, "- $ref: \"#/components/parameters/StateIDQuery\"")
	requireContains(t, listBlock, "\"422\":")
	requireContains(t, listBlock, "\"500\":")
	requireNotContains(t, listBlock, "security:")

	requireContains(t, detailBlock, "get:")
	requireContains(t, detailBlock, "- $ref: \"#/components/parameters/LGAID\"")
	requireContains(t, detailBlock, "\"404\":")
	requireContains(t, detailBlock, "\"422\":")
	requireContains(t, detailBlock, "\"500\":")
	requireNotContains(t, detailBlock, "security:")

	requireContains(t, text, "StateIDQuery:\n      name: state_id\n      in: query\n      required: false\n      description: Optional public state or FCT ID used to filter local-government-unit results.\n      schema:\n        type: string")
	requireContains(t, text, "LGAID:\n      name: lga_id\n      in: path\n      required: true\n      description: Stable public identifier for a local government unit or FCT Area Council.\n      schema:\n        type: string")

	requireContains(t, text, "LocalGovernmentUnit:\n      type: object\n      additionalProperties: false\n      properties:\n        id:\n          type: string\n        name:\n          type: string\n        state_id:\n          type: string\n        country_code:\n          type: string\n          enum: [NG]\n        administrative_type:\n          type: string\n          enum: [local_government_area, area_council]\n      required: [id, name, state_id, country_code, administrative_type]")
	requireContains(t, text, "LocalGovernmentUnitListResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          type: array\n          items:\n            $ref: \"#/components/schemas/LocalGovernmentUnit\"\n      required: [success, data]")
	requireContains(t, text, "LocalGovernmentUnitResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          $ref: \"#/components/schemas/LocalGovernmentUnit\"\n      required: [success, data]")
}

func TestOpenAPIDocumentsCountryOrAreaPaths(t *testing.T) {
	doc, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(doc)

	listBlock := pathBlock(t, text, "/v1/geography/countries")
	detailBlock := pathBlock(t, text, "/v1/geography/countries/{country_id}")
	profileBlock := pathBlock(t, text, "/v1/geography/countries/{country_id}/profile")
	flagBlock := pathBlock(t, text, "/v1/assets/flags/{country_id}.svg")

	requireContains(t, text, "    CountryOrAreaID:")
	requireContains(t, text, "    CountryOrAreaRegionCode:")
	requireContains(t, text, "    CountryOrAreaSubregionCode:")
	requireContains(t, text, "    CountryOrArea:")
	requireContains(t, text, "    CountryOrAreaListResponse:")
	requireContains(t, text, "    CountryOrAreaResponse:")
	requireContains(t, text, "    CountryProfile:")
	requireContains(t, text, "    CountryProfileResponse:")
	requireContains(t, text, "    CountryOrAreaFlagID:")

	requireContains(t, listBlock, "get:")
	requireContains(t, listBlock, "- $ref: \"#/components/parameters/CountryOrAreaRegionCode\"")
	requireContains(t, listBlock, "- $ref: \"#/components/parameters/CountryOrAreaSubregionCode\"")
	requireContains(t, listBlock, "\"422\":")
	requireContains(t, listBlock, "\"500\":")
	requireNotContains(t, listBlock, "security:")
	requireNotContains(t, listBlock, "Page")
	requireNotContains(t, listBlock, "Limit")
	requireNotContains(t, listBlock, "Search")

	requireContains(t, detailBlock, "get:")
	requireContains(t, detailBlock, "- $ref: \"#/components/parameters/CountryOrAreaID\"")
	requireContains(t, detailBlock, "\"404\":")
	requireContains(t, detailBlock, "\"422\":")
	requireContains(t, detailBlock, "\"500\":")
	requireNotContains(t, detailBlock, "security:")

	requireContains(t, text, "CountryOrAreaID:\n      name: country_id\n      in: path\n      required: true\n      description: Stable lowercase alpha-2 identifier for a UN M49 country or area.\n      schema:\n        type: string\n        pattern: '^[a-z]{2}$'")
	requireContains(t, text, "CountryOrAreaFlagID:\n      name: country_id\n      in: path\n      required: true\n      description: Stable lowercase alpha-2 identifier for a UN M49 country or area SVG flag asset.\n      schema:\n        type: string\n        pattern: '^[a-z]{2}\\.svg$'")
	requireContains(t, text, "CountryOrAreaRegionCode:\n      name: region_code\n      in: query\n      required: false\n      description: Optional UN M49 region code used to filter country or area records.\n      schema:\n        type: string\n        pattern: '^[0-9]{3}$'")
	requireContains(t, text, "CountryOrAreaSubregionCode:\n      name: subregion_code\n      in: query\n      required: false\n      description: Optional UN M49 subregion code used to filter country or area records.\n      schema:\n        type: string\n        pattern: '^[0-9]{3}$'")
	requireContains(t, text, "CountryOrArea:\n      type: object\n      additionalProperties: false\n      properties:\n        id:\n          type: string\n          pattern: '^[a-z]{2}$'\n        name:\n          type: string\n        alpha_2_code:\n          type: string\n          pattern: '^[A-Z]{2}$'\n        alpha_3_code:\n          type: string\n          pattern: '^[A-Z]{3}$'\n        numeric_code:\n          type: string\n          pattern: '^[0-9]{3}$'\n        calling_codes:\n          type: array\n          uniqueItems: true\n          minItems: 1\n          items:\n            type: string\n            pattern: '^\\+[1-9][0-9]{0,2}(?:-[0-9]{1,4})*$'\n        flag_emoji:\n          type: string\n        flag_svg_url:\n          type: string\n          pattern: '^/v1/assets/flags/[a-z]{2}\\.svg$'\n        region_code:\n          type: string\n          pattern: '^[0-9]{3}$'\n        region_name:\n          type: string\n        subregion_code:\n          type: string\n          pattern: '^[0-9]{3}$'\n        subregion_name:\n          type: string\n        intermediate_region_code:\n          type: string\n          pattern: '^[0-9]{3}$'\n        intermediate_region_name:\n          type: string\n      required: [id, name, alpha_2_code, alpha_3_code, numeric_code]")
	requireContains(t, text, "CountryOrAreaListResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          type: array\n          items:\n            $ref: \"#/components/schemas/CountryOrArea\"\n      required: [success, data]")
	requireContains(t, text, "CountryOrAreaResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          $ref: \"#/components/schemas/CountryOrArea\"\n      required: [success, data]")
	requireContains(t, text, "CountryProfile:\n      type: object\n      additionalProperties: false\n      properties:\n        id:\n          type: string\n          pattern: '^[a-z]{2}$'\n        name:\n          type: string\n        alpha_2_code:\n          type: string\n          pattern: '^[A-Z]{2}$'\n        alpha_3_code:\n          type: string\n          pattern: '^[A-Z]{3}$'\n        numeric_code:\n          type: string\n          pattern: '^[0-9]{3}$'\n        calling_codes:\n          type: array\n          uniqueItems: true\n          minItems: 1\n          items:\n            type: string\n            pattern: '^\\+[1-9][0-9]{0,2}(?:-[0-9]{1,4})*$'\n        flag_emoji:\n          type: string\n        flag_svg_url:\n          type: string\n          pattern: '^/v1/assets/flags/[a-z]{2}\\.svg$'\n        region_code:\n          type: string\n          pattern: '^[0-9]{3}$'\n        region_name:\n          type: string\n        subregion_code:\n          type: string\n          pattern: '^[0-9]{3}$'\n        subregion_name:\n          type: string\n        intermediate_region_code:\n          type: string\n          pattern: '^[0-9]{3}$'\n        intermediate_region_name:\n          type: string\n        currency_ids:\n          type: array\n          uniqueItems: true\n          items:\n            type: string\n            pattern: '^[a-z]{3}$'\n        time_zone_ids:\n          type: array\n          uniqueItems: true\n          items:\n            type: string\n        language_ids:\n          type: array\n          description: Sorted unique IDs from world-languages. Includes every CLDR-derived country-language relationship regardless of status; presence does not mean the language is nationally official. Use /v1/geography/country-languages for relationship status details. Empty arrays are valid.\n          uniqueItems: true\n          items:\n            type: string\n            pattern: '^[a-z]{2,3}$'\n      required: [id, name, alpha_2_code, alpha_3_code, numeric_code, currency_ids, time_zone_ids, language_ids]")
	requireContains(t, text, "CountryProfileResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          $ref: \"#/components/schemas/CountryProfile\"\n      required: [success, data]")
	requireContains(t, text, "Statistical designations are used for reference only and do not imply political recognition or legal status.")

	requireContains(t, profileBlock, "get:")
	requireContains(t, profileBlock, "- $ref: \"#/components/parameters/CountryOrAreaID\"")
	requireContains(t, profileBlock, "\"404\":")
	requireContains(t, profileBlock, "\"422\":")
	requireContains(t, profileBlock, "\"500\":")
	requireNotContains(t, profileBlock, "security:")

	requireContains(t, flagBlock, "get:")
	requireContains(t, flagBlock, "\"200\":")
	requireContains(t, flagBlock, "\"404\":")
	requireContains(t, flagBlock, "\"422\":")
	requireContains(t, flagBlock, "\"500\":")
	requireContains(t, flagBlock, "image/svg+xml")
	requireContains(t, flagBlock, "- $ref: \"#/components/parameters/CountryOrAreaFlagID\"")
}

func TestOpenAPIDocumentsPaymentServiceProviderPaths(t *testing.T) {
	doc, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(doc)

	listBlock := pathBlock(t, text, "/v1/finance/payment-service-providers")
	detailBlock := pathBlock(t, text, "/v1/finance/payment-service-providers/{provider_id}")

	requireContains(t, text, "    PaymentServiceProviderID:")
	requireContains(t, text, "    PaymentServiceProviderType:")
	requireContains(t, text, "    PaymentServiceProvider:")
	requireContains(t, text, "    PaymentServiceProviderListResponse:")
	requireContains(t, text, "    PaymentServiceProviderResponse:")
	requireContains(t, text, "PaymentServiceProviderType:\n      type: string\n      enum:\n        - mobile_money_operator\n        - switching_and_processing_company\n        - payment_solution_service_provider\n        - payment_terminal_service_provider\n        - super_agent\n        - payment_service_holding_company\n        - payment_terminal_service_aggregator")

	requireContains(t, listBlock, "get:")
	requireContains(t, listBlock, "- $ref: \"#/components/parameters/PaymentServiceProviderTypeQuery\"")
	requireContains(t, listBlock, "\"422\":")
	requireContains(t, listBlock, "\"500\":")
	requireNotContains(t, listBlock, "security:")
	requireNotContains(t, listBlock, "#/components/parameters/Page")
	requireNotContains(t, listBlock, "#/components/parameters/Limit")
	requireNotContains(t, listBlock, "#/components/parameters/Search")

	requireContains(t, detailBlock, "get:")
	requireContains(t, detailBlock, "- $ref: \"#/components/parameters/PaymentServiceProviderID\"")
	requireContains(t, detailBlock, "\"404\":")
	requireContains(t, detailBlock, "\"422\":")
	requireContains(t, detailBlock, "\"500\":")
	requireNotContains(t, detailBlock, "security:")
	requireNotContains(t, text, "institution_code")

	requireContains(t, text, "PaymentServiceProviderTypeQuery:\n      name: institution_type\n      in: query\n      required: false\n      description: Optional payment-service-provider category used to filter provider-category memberships.\n      schema:\n        $ref: \"#/components/schemas/PaymentServiceProviderType\"")
	requireContains(t, text, "PaymentServiceProviderID:\n      name: provider_id\n      in: path\n      required: true\n      description: Stable public identifier for a payment-service-provider membership.\n      schema:\n        type: string\n        pattern: '^[a-z0-9]+(?:-[a-z0-9]+)+$'")
	requireContains(t, text, "PaymentServiceProvider:\n      type: object\n      additionalProperties: false\n      properties:\n        id:\n          type: string\n        name:\n          type: string\n        institution_type:\n          $ref: \"#/components/schemas/PaymentServiceProviderType\"\n        country_code:\n          type: string\n          enum: [NG]\n      required: [id, name, institution_type, country_code]")
	requireContains(t, text, "PaymentServiceProviderListResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          type: array\n          items:\n            $ref: \"#/components/schemas/PaymentServiceProvider\"\n      required: [success, data]")
	requireContains(t, text, "PaymentServiceProviderResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          $ref: \"#/components/schemas/PaymentServiceProvider\"\n      required: [success, data]")
}

func TestOpenAPIDocumentsInternationalMoneyTransferOperatorPaths(t *testing.T) {
	doc, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(doc)

	listBlock := pathBlock(t, text, "/v1/finance/international-money-transfer-operators")
	detailBlock := pathBlock(t, text, "/v1/finance/international-money-transfer-operators/{operator_id}")

	requireContains(t, text, "    InternationalMoneyTransferOperatorID:")
	requireContains(t, text, "    InternationalMoneyTransferOperator:")
	requireContains(t, text, "    InternationalMoneyTransferOperatorListResponse:")
	requireContains(t, text, "    InternationalMoneyTransferOperatorResponse:")

	requireContains(t, listBlock, "get:")
	requireNotContains(t, listBlock, "parameters:")
	requireContains(t, listBlock, "\"500\":")
	requireNotContains(t, listBlock, "security:")

	requireContains(t, detailBlock, "get:")
	requireContains(t, detailBlock, "- $ref: \"#/components/parameters/InternationalMoneyTransferOperatorID\"")
	requireContains(t, detailBlock, "\"404\":")
	requireContains(t, detailBlock, "\"422\":")
	requireContains(t, detailBlock, "\"500\":")
	requireNotContains(t, detailBlock, "security:")

	requireContains(t, text, "InternationalMoneyTransferOperatorID:\n      name: operator_id\n      in: path\n      required: true\n      description: Stable public identifier for an international money transfer operator.\n      schema:\n        type: string\n        pattern: '^[a-z0-9]+(?:-[a-z0-9]+)*$'")
	requireContains(t, text, "InternationalMoneyTransferOperator:\n      type: object\n      additionalProperties: false\n      properties:\n        id:\n          type: string\n        name:\n          type: string\n      required: [id, name]")
	requireContains(t, text, "InternationalMoneyTransferOperatorListResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          type: array\n          items:\n            $ref: \"#/components/schemas/InternationalMoneyTransferOperator\"\n      required: [success, data]")
	requireContains(t, text, "InternationalMoneyTransferOperatorResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          $ref: \"#/components/schemas/InternationalMoneyTransferOperator\"\n      required: [success, data]")
}

func TestOpenAPIDocumentsCurrencyPaths(t *testing.T) {
	doc, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(doc)

	listBlock := pathBlock(t, text, "/v1/finance/currencies")
	detailBlock := pathBlock(t, text, "/v1/finance/currencies/{currency_id}")

	requireContains(t, text, "    CurrencyID:")
	requireContains(t, text, "    CurrencyCountryAreaIDQuery:")
	requireContains(t, text, "    Currency:")
	requireContains(t, text, "    CurrencyListResponse:")
	requireContains(t, text, "    CurrencyResponse:")

	requireContains(t, listBlock, "get:")
	requireContains(t, listBlock, "- $ref: \"#/components/parameters/CurrencyCountryAreaIDQuery\"")
	requireContains(t, listBlock, "\"422\":")
	requireContains(t, listBlock, "\"500\":")
	requireNotContains(t, listBlock, "security:")

	requireContains(t, detailBlock, "get:")
	requireContains(t, detailBlock, "- $ref: \"#/components/parameters/CurrencyID\"")
	requireContains(t, detailBlock, "\"404\":")
	requireContains(t, detailBlock, "\"422\":")
	requireContains(t, detailBlock, "\"500\":")
	requireNotContains(t, detailBlock, "security:")

	requireContains(t, text, "CurrencyID:\n      name: currency_id\n      in: path\n      required: true\n      description: Stable public identifier for a currency.\n      schema:\n        type: string\n        pattern: '^[a-z]{3}$'")
	requireContains(t, text, "CurrencyCountryAreaIDQuery:\n      name: country_area_id\n      in: query\n      required: false\n      description: Optional UN M49 country or area ID used to filter currency records.\n      schema:\n        type: string\n        pattern: '^[a-z]{2}$'")
	requireContains(t, text, "Currency:\n      type: object\n      additionalProperties: false\n      properties:\n        id:\n          type: string\n          pattern: '^[a-z]{3}$'\n        name:\n          type: string\n        alphabetic_code:\n          type: string\n          pattern: '^[A-Z]{3}$'\n        numeric_code:\n          type: string\n          pattern: '^[0-9]{3}$'\n        minor_unit:\n          type: integer\n          enum: [0, 2, 3]\n        country_area_ids:\n          type: array\n          uniqueItems: true\n          items:\n            type: string\n            pattern: '^[a-z]{2}$'\n      required: [id, name, alphabetic_code, numeric_code, minor_unit, country_area_ids]")
	requireContains(t, text, "CurrencyListResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          type: array\n          items:\n            $ref: \"#/components/schemas/Currency\"\n      required: [success, data]")
	requireContains(t, text, "CurrencyResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          $ref: \"#/components/schemas/Currency\"\n      required: [success, data]")
}

func TestOpenAPIDocumentsUniversityPaths(t *testing.T) {
	doc, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(doc)

	listBlock := pathBlock(t, text, "/v1/education/universities")
	detailBlock := pathBlock(t, text, "/v1/education/universities/{university_id}")

	requireContains(t, text, "    UniversityID:")
	requireContains(t, text, "    UniversityOwnershipType:")
	requireContains(t, text, "    UniversityStateID:")
	requireContains(t, text, "    University:")
	requireContains(t, text, "    UniversityListResponse:")
	requireContains(t, text, "    UniversityResponse:")

	requireContains(t, listBlock, "get:")
	requireContains(t, listBlock, "- $ref: \"#/components/parameters/UniversityOwnershipType\"")
	requireContains(t, listBlock, "- $ref: \"#/components/parameters/UniversityStateID\"")
	requireContains(t, listBlock, "\"422\":")
	requireContains(t, listBlock, "\"500\":")
	requireNotContains(t, listBlock, "security:")

	requireContains(t, detailBlock, "get:")
	requireContains(t, detailBlock, "- $ref: \"#/components/parameters/UniversityID\"")
	requireContains(t, detailBlock, "\"404\":")
	requireContains(t, detailBlock, "\"422\":")
	requireContains(t, detailBlock, "\"500\":")
	requireNotContains(t, detailBlock, "security:")

	requireContains(t, text, "UniversityID:\n      name: university_id\n      in: path\n      required: true\n      description: Stable public identifier for a current NUC-listed university.\n      schema:\n        type: string\n        pattern: '^[a-z0-9]+(?:-[a-z0-9]+)+$'")
	requireContains(t, text, "UniversityOwnershipType:\n      name: ownership_type\n      in: query\n      required: false\n      description: Optional university ownership category used to filter universities.\n      schema:\n        $ref: \"#/components/schemas/UniversityOwnershipType\"")
	requireContains(t, text, "UniversityStateID:\n      name: state_id\n      in: query\n      required: false\n      description: Optional public state or FCT ID used to filter universities.\n      schema:\n        $ref: \"#/components/schemas/UniversityStateID\"")
	requireContains(t, text, "UniversityOwnershipType:\n      type: string\n      enum: [federal, state, private]")
	requireContains(t, text, "UniversityStateID:\n      type: string\n      enum:\n        - abia")
	requireContains(t, text, "University:\n      type: object\n      additionalProperties: false\n      properties:\n        id:\n          type: string\n        name:\n          type: string\n        ownership_type:\n          $ref: \"#/components/schemas/UniversityOwnershipType\"\n        state_id:\n          $ref: \"#/components/schemas/UniversityStateID\"\n        country_code:\n          type: string\n          enum: [NG]\n      required: [id, name, ownership_type, state_id, country_code]")
	requireContains(t, text, "UniversityListResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          type: array\n          items:\n            $ref: \"#/components/schemas/University\"\n      required: [success, data]")
	requireContains(t, text, "UniversityResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          $ref: \"#/components/schemas/University\"\n      required: [success, data]")
}

func TestOpenAPIDocumentsCollegeOfEducationPaths(t *testing.T) {
	doc, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(doc)

	listBlock := pathBlock(t, text, "/v1/education/colleges-of-education")
	detailBlock := pathBlock(t, text, "/v1/education/colleges-of-education/{college_id}")

	requireContains(t, text, "    CollegeOfEducationID:")
	requireContains(t, text, "    CollegeOfEducationOwnershipType:")
	requireContains(t, text, "    CollegeOfEducationStateID:")
	requireContains(t, text, "    CollegeOfEducation:")
	requireContains(t, text, "    CollegeOfEducationListResponse:")
	requireContains(t, text, "    CollegeOfEducationResponse:")

	requireContains(t, listBlock, "get:")
	requireContains(t, listBlock, "- $ref: \"#/components/parameters/CollegeOfEducationOwnershipType\"")
	requireContains(t, listBlock, "- $ref: \"#/components/parameters/CollegeOfEducationStateID\"")
	requireContains(t, listBlock, "\"422\":")
	requireContains(t, listBlock, "\"500\":")
	requireNotContains(t, listBlock, "security:")

	requireContains(t, detailBlock, "get:")
	requireContains(t, detailBlock, "- $ref: \"#/components/parameters/CollegeOfEducationID\"")
	requireContains(t, detailBlock, "\"404\":")
	requireContains(t, detailBlock, "\"422\":")
	requireContains(t, detailBlock, "\"500\":")
	requireNotContains(t, detailBlock, "security:")

	requireContains(t, text, "CollegeOfEducationID:\n      name: college_id\n      in: path\n      required: true\n      description: Stable public identifier for a college of education.\n      schema:\n        type: string\n        pattern: '^[a-z0-9]+(?:-[a-z0-9]+)+$'")
	requireContains(t, text, "CollegeOfEducationOwnershipType:\n      name: ownership_type\n      in: query\n      required: false\n      description: Optional college-of-education ownership category used to filter colleges.\n      schema:\n        $ref: \"#/components/schemas/CollegeOfEducationOwnershipType\"")
	requireContains(t, text, "CollegeOfEducationStateID:\n      name: state_id\n      in: query\n      required: false\n      description: Optional public state or FCT ID used to filter college-of-education records.\n      schema:\n        $ref: \"#/components/schemas/CollegeOfEducationStateID\"")
	requireContains(t, text, "CollegeOfEducationOwnershipType:\n      type: string\n      enum: [federal, state, private]")
	requireContains(t, text, "CollegeOfEducationStateID:\n      type: string\n      enum:\n        - abia")
	requireContains(t, text, "CollegeOfEducation:\n      type: object\n      additionalProperties: false\n      properties:\n        id:\n          type: string\n        name:\n          type: string\n        ownership_type:\n          $ref: \"#/components/schemas/CollegeOfEducationOwnershipType\"\n        state_id:\n          $ref: \"#/components/schemas/CollegeOfEducationStateID\"\n        country_code:\n          type: string\n          enum: [NG]\n      required: [id, name, ownership_type, state_id, country_code]")
	requireContains(t, text, "CollegeOfEducationListResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          type: array\n          items:\n            $ref: \"#/components/schemas/CollegeOfEducation\"\n      required: [success, data]")
	requireContains(t, text, "CollegeOfEducationResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          $ref: \"#/components/schemas/CollegeOfEducation\"\n      required: [success, data]")
}

func TestOpenAPIDocumentsTimeZonePaths(t *testing.T) {
	doc, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(doc)

	listBlock := pathBlock(t, text, "/v1/geography/time-zones")
	detailBlock := pathBlock(t, text, "/v1/geography/time-zones/{time_zone_id}")

	requireContains(t, text, "    TimeZoneID:")
	requireContains(t, text, "    TimeZoneCountryAreaID:")
	requireContains(t, text, "    TimeZone:")
	requireContains(t, text, "    TimeZoneListResponse:")
	requireContains(t, text, "    TimeZoneResponse:")

	requireContains(t, listBlock, "get:")
	requireContains(t, listBlock, "- $ref: \"#/components/parameters/TimeZoneCountryAreaID\"")
	requireContains(t, listBlock, "\"422\":")
	requireContains(t, listBlock, "\"500\":")
	requireNotContains(t, listBlock, "security:")

	requireContains(t, detailBlock, "get:")
	requireContains(t, detailBlock, "- $ref: \"#/components/parameters/TimeZoneID\"")
	requireContains(t, detailBlock, "\"404\":")
	requireContains(t, detailBlock, "\"422\":")
	requireContains(t, detailBlock, "\"500\":")
	requireNotContains(t, detailBlock, "security:")

	requireContains(t, text, "TimeZoneID:\n      name: time_zone_id\n      in: path\n      required: true\n      allowReserved: true\n      description: Exact canonical IANA time-zone identifier. Percent-encoded slashes are accepted by the HTTP layer and decoded before validation. Aliases are excluded from v1.\n      schema:\n        type: string")
	requireContains(t, text, "TimeZoneCountryAreaID:\n      name: country_area_id\n      in: query\n      required: false\n      description: Optional public country or area ID used to filter time-zone records.\n      schema:\n        type: string\n        pattern: '^[a-z]{2}$'")
	requireContains(t, text, "TimeZone:\n      type: object\n      additionalProperties: false\n      properties:\n        id:\n          type: string\n        country_area_ids:\n          type: array\n          uniqueItems: true\n          items:\n            type: string\n            pattern: '^[a-z]{2}$'\n      required: [id, country_area_ids]")
	requireContains(t, text, "TimeZoneListResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          type: array\n          items:\n            $ref: \"#/components/schemas/TimeZone\"\n      required: [success, data]")
	requireContains(t, text, "TimeZoneResponse:\n      type: object\n      additionalProperties: false\n      properties:\n        success:\n          type: boolean\n        data:\n          $ref: \"#/components/schemas/TimeZone\"\n      required: [success, data]")
	requireContains(t, text, "canonical IANA time-zone manifest")
	requireContains(t, text, "zone1970.tab")
	requireContains(t, text, "Asia/Taipei")
	requireContains(t, text, "bv")
	requireContains(t, text, "hm")
}

func pathBlock(t *testing.T, doc, path string) string {
	t.Helper()

	marker := "\n  " + path + ":\n"
	start := strings.Index(doc, marker)
	if start < 0 {
		t.Fatalf("missing path: %s", path)
	}

	rest := doc[start+len(marker):]
	next := strings.Index(rest, "\n  /")
	if next < 0 {
		return rest
	}
	return rest[:next]
}

func requireContains(t *testing.T, text, substring string) {
	t.Helper()
	if !strings.Contains(text, substring) {
		t.Fatalf("missing substring: %q", substring)
	}
}

func requireNotContains(t *testing.T, text, substring string) {
	t.Helper()
	if strings.Contains(text, substring) {
		t.Fatalf("unexpected substring: %q", substring)
	}
}
