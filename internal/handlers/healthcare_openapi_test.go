package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestOpenAPIDocumentsHealthFacilityPaths(t *testing.T) {
	doc, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	text := string(doc)
	list := pathBlock(t, text, "/v1/healthcare/health-facilities")
	detail := pathBlock(t, text, "/v1/healthcare/health-facilities/{facility_id}")
	for _, value := range []string{
		"listHealthFacilities", "getHealthFacility", "HealthFacility", "HealthFacilityPagination",
		"HealthFacilityListResponse", "HealthFacilityDetailResponse", "HealthFacilityType",
		"HealthFacilityLevel", "HealthFacilityOwnershipType", "HealthFacilitySearch",
	} {
		requireContains(t, text, value)
	}
	for _, value := range []string{"get:", "HealthFacilityPage", "HealthFacilityPageSize", "HealthFacilityStateID", "HealthFacilityLGAID", "HealthFacilityType", "HealthFacilityLevel", "HealthFacilityOwnershipType", "HealthFacilitySearch", `"400":`, `"500":`} {
		requireContains(t, list, value)
	}
	for _, value := range []string{"get:", "HealthFacilityID", `"400":`, `"404":`, `"500":`} {
		requireContains(t, detail, value)
	}
	if strings.Contains(list, "operational_status") || strings.Contains(detail, "operational_status") || strings.Contains(text, "HealthFacilityWebsite") || strings.Contains(text, "HealthFacilityLogo") {
		t.Fatal("health OpenAPI contract contains forbidden fields")
	}
	if !strings.Contains(list, "50,654") || !strings.Contains(list, "2024-11-11") || !strings.Contains(list, "does not establish current operational status") {
		t.Fatal("health snapshot limitation is not documented")
	}
}

func TestOpenAPIOperationIDsRemainUnique(t *testing.T) {
	doc, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]struct{}{}
	for _, line := range strings.Split(string(doc), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "operationId:") {
			continue
		}
		id := strings.TrimSpace(strings.TrimPrefix(line, "operationId:"))
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate operationId %q", id)
		}
		seen[id] = struct{}{}
	}
}
