package handlers

import (
	"os"
	"testing"
)

func TestOpenAPIDocumentsLanguagePathsAndSchemas(t *testing.T) {
	doc, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(doc)
	for _, path := range []string{"/v1/geography/languages", "/v1/geography/languages/{language_id}", "/v1/geography/country-languages"} {
		requireContains(t, text, path+":")
		block := pathBlock(t, text, path)
		requireContains(t, block, "get:")
		requireNotContains(t, block, "security:")
	}
	requireContains(t, text, "LanguageID:")
	requireContains(t, text, "CountryLanguageStatus:")
	requireContains(t, text, "LanguageListResponse:")
	requireContains(t, text, "LanguageResponse:")
	requireContains(t, text, "CountryLanguageListResponse:")
	requireContains(t, text, "name: language_id\n      in: path")
	requireContains(t, text, "name: country_area_id\n      in: query")
	requireContains(t, text, "name: status\n      in: query")
	requireContains(t, text, "enum: [official, de_facto_official, official_regional, used]")
}
