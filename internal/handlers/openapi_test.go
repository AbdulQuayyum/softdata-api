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
