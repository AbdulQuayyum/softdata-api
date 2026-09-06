package handlers

import (
	"os"
	"testing"
)

func TestOpenAPIDocumentsMicrofinanceBankPathsAndSchemas(t *testing.T) {
	doc, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(doc)
	listBlock := pathBlock(t, text, "/v1/finance/microfinance-banks")
	detailBlock := pathBlock(t, text, "/v1/finance/microfinance-banks/{bank_id}")

	requireContains(t, listBlock, "operationId: listMicrofinanceBanks")
	requireContains(t, listBlock, "\"200\":")
	requireContains(t, listBlock, "\"405\":")
	requireContains(t, listBlock, "\"500\":")
	requireNotContains(t, listBlock, "security:")
	requireNotContains(t, listBlock, "Page")
	requireNotContains(t, listBlock, "Limit")
	requireNotContains(t, listBlock, "Search")

	requireContains(t, detailBlock, "operationId: getMicrofinanceBank")
	requireContains(t, detailBlock, "- $ref: \"#/components/parameters/MicrofinanceBankID\"")
	requireContains(t, detailBlock, "\"404\":")
	requireContains(t, detailBlock, "\"405\":")
	requireContains(t, detailBlock, "\"422\":")
	requireContains(t, detailBlock, "\"500\":")
	requireNotContains(t, detailBlock, "security:")

	requireContains(t, text, "MicrofinanceBankID:")
	requireContains(t, text, "maxLength: 128\n        pattern: '^[a-z0-9]+(?:-[a-z0-9]+)*$'")
	requireContains(t, text, "MicrofinanceBank:\n      type: object\n      additionalProperties: false")
	requireContains(t, text, "required: [id, name, country_code]")
	requireContains(t, text, "MicrofinanceBankListResponse:")
	requireContains(t, text, "minItems: 790\n          maxItems: 790")
	requireContains(t, text, "MicrofinanceBankResponse:")
}
