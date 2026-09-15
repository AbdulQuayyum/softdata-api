package handlers

import (
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

func TestNHIAActiveAccreditedHealthcareProviderOpenAPIContract(t *testing.T) {
	raw, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(raw)
	prefix := "NHIAActiveAccreditedHealthcareProvider"
	list := pathBlock(t, doc, "/v1/healthcare/nhia-active-accredited-healthcare-providers")
	detail := pathBlock(t, doc, "/v1/healthcare/nhia-active-accredited-healthcare-providers/{provider_id}")
	for _, tc := range []struct {
		block, id string
		params    []string
	}{
		{list, "listNHIAActiveAccreditedHealthcareProviders", []string{"Page", "PageSize", "Code", "FacilityType", "ListingStatus", "Search"}},
		{detail, "getNHIAActiveAccreditedHealthcareProvider", []string{"ID"}},
	} {
		requireContains(t, tc.block, "operationId: "+tc.id)
		if strings.Count(tc.block, "#/components/parameters/") != len(tc.params) {
			t.Fatal("unexpected parameter count")
		}
		for _, p := range tc.params {
			requireContains(t, tc.block, "#/components/parameters/"+prefix+p)
		}
		for _, status := range []string{`"400":`, `"405": {$ref: "#/components/responses/MethodNotAllowedGet"}`, `"500":`, `"503":`} {
			requireContains(t, tc.block, status)
		}
	}
	requireContains(t, detail, `"404":`)
	for _, phrase := range []string{"Production-wired public GET route", "standard healthcare public middleware", "ACTIVEACCREDITED NHIA HEALTHCARE PROVIDER.csv", "6,536 retained public provider records", "four unresolved source observations were excluded", "not a complete live licensing, registration, or operational-status register", "separate from the GRID3 ng-health-facilities snapshot"} {
		requireContains(t, list+detail, phrase)
	}
	for _, stale := range []string{"future public GET route", "Production route registration and startup verification are intentionally deferred"} {
		requireNotContains(t, list+detail, stale)
	}

	for _, tc := range []struct{ suffix, name, location, schema string }{
		{"ID", "provider_id", "path", "maxLength: " + strconv.Itoa(models.NHIAActiveAccreditedHealthcareProviderIDMaxLength)},
		{"Page", "page", "query", "type: integer, minimum: 1, default: 1"},
		{"PageSize", "page_size", "query", "type: integer, minimum: 1, maximum: 100, default: 50"},
		{"Code", "provider_code", "query", "type: string"},
		{"FacilityType", "facility_type", "query", "enum: [primary, primary_and_secondary]"},
		{"ListingStatus", "listing_status", "query", "enum: [active_accredited]"},
		{"Search", "search", "query", "maxLength: 100"},
	} {
		block := nhiaHMODocBlock(t, doc, prefix+tc.suffix)
		for _, value := range []string{"name: " + tc.name, "in: " + tc.location, tc.schema, "description:"} {
			requireContains(t, block, value)
		}
	}
	requireContains(t, nhiaHMODocBlock(t, doc, prefix+"Code"), "Slash characters are part of the code")

	schema := nhiaHMODocBlock(t, doc, prefix)
	fields := regexp.MustCompile(`(?m)^        ([a-z_]+):`).FindAllStringSubmatch(schema, -1)
	typ := reflect.TypeOf(models.NHIAActiveAccreditedHealthcareProvider{})
	if len(fields) != typ.NumField() {
		t.Fatalf("schema fields %d model %d", len(fields), typ.NumField())
	}
	required := []string{}
	for i := 0; i < typ.NumField(); i++ {
		tag := strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]
		requireContains(t, schema, "        "+tag+":")
		required = append(required, tag)
	}
	requireContains(t, schema, "required: ["+strings.Join(required, ", ")+"]")
	requireContains(t, schema, "facility_type: {type: string, enum: [primary, primary_and_secondary]}")
	requireContains(t, schema, "listing_status: {type: string, enum: [active_accredited]}")
	requireContains(t, schema, "provider_code: {type: string")
	for _, forbidden := range []string{"state_id", "lga_id", "address", "website", "logo", "phone", "email", "director", "contact", "ownership", "coordinate", "operational_status", "licence_status", "license_status", "registration_status"} {
		requireNotContains(t, schema, forbidden+":")
	}
	for _, suffix := range []string{"ListResponse", "DetailResponse"} {
		block := nhiaHMODocBlock(t, doc, prefix+suffix)
		requireContains(t, block, "#/components/schemas/"+prefix)
		requireContains(t, block, "success: {type: boolean}")
	}
	requireContains(t, nhiaHMODocBlock(t, doc, prefix+"Pagination"), "required: [page, page_size, total, total_pages]")
	requireContains(t, nhiaHMODocBlock(t, doc, "MethodNotAllowedGet"), "Allow:")
}
