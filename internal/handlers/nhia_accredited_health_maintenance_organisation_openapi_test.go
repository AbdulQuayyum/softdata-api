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

func nhiaHMODocBlock(t *testing.T, doc, name string) string {
	t.Helper()
	start := strings.Index(doc, "    "+name+":\n")
	if start < 0 {
		t.Fatalf("missing component %s", name)
	}
	rest := doc[start+len("    "+name+":\n"):]
	end := regexp.MustCompile(`(?m)^    [A-Za-z][A-Za-z0-9]*:`).FindStringIndex(rest)
	if end != nil {
		return rest[:end[0]]
	}
	return rest
}

func TestNHIAAccreditedHMOOpenAPIContract(t *testing.T) {
	raw, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(raw)
	prefix := "NHIAAccreditedHealthMaintenanceOrganisation"
	list := pathBlock(t, doc, "/v1/healthcare/nhia-accredited-health-maintenance-organisations")
	detail := pathBlock(t, doc, "/v1/healthcare/nhia-accredited-health-maintenance-organisations/{organisation_id}")
	for _, tc := range []struct {
		block, id string
		params    []string
	}{{list, "listNHIAAccreditedHealthMaintenanceOrganisations", []string{"Page", "PageSize", "Status", "HMOID", "Search"}}, {detail, "getNHIAAccreditedHealthMaintenanceOrganisation", []string{"ID"}}} {
		requireContains(t, tc.block, "operationId: "+tc.id)
		if strings.Count(tc.block, "#/components/parameters/") != len(tc.params) {
			t.Fatal("unexpected parameters")
		}
		for _, p := range tc.params {
			requireContains(t, tc.block, "#/components/parameters/"+prefix+p)
		}
		for _, status := range []string{`"400":`, `"405": {$ref: "#/components/responses/MethodNotAllowedGet"}`, `"500":`, `"503":`} {
			requireContains(t, tc.block, status)
		}
		for _, phrase := range []string{"dated NHIA accreditation snapshot", "Health Maintenance Organisations", "94 organisation records", "all marked accredited", "not a complete live licensing, registration, or operational register", "does not prove that every Nigerian health-insurance organisation is licensed, registered, or operational"} {
			requireContains(t, tc.block, phrase)
		}
	}
	requireContains(t, detail, `"404":`)
	requireContains(t, detail, "malformed organisation_id returns 400")
	requireContains(t, detail, "unknown organisation_id returns 404")

	for _, tc := range []struct{ suffix, name, location, schema string }{
		{"ID", "organisation_id", "path", "maxLength: " + strconv.Itoa(models.NHIAAccreditedHealthMaintenanceOrganisationIDMaxLength)},
		{"Page", "page", "query", "type: integer, minimum: 1, default: 1"},
		{"PageSize", "page_size", "query", "type: integer, minimum: 1, maximum: 100, default: 50"},
		{"Status", "accreditation_status", "query", "enum: [accredited]"},
		{"HMOID", "hmo_id", "query", "type: string, maxLength: 64"},
		{"Search", "search", "query", "maxLength: 100"},
	} {
		block := nhiaHMODocBlock(t, doc, prefix+tc.suffix)
		for _, value := range []string{"name: " + tc.name, "in: " + tc.location, tc.schema, "description:", "example:"} {
			requireContains(t, block, value)
		}
	}
	requireContains(t, nhiaHMODocBlock(t, doc, prefix+"HMOID"), "leading zeros are preserved")
	requireContains(t, nhiaHMODocBlock(t, doc, prefix+"HMOID"), "not parsed as an integer")

	schema := nhiaHMODocBlock(t, doc, prefix)
	fields := regexp.MustCompile(`(?m)^        ([a-z_]+):`).FindAllStringSubmatch(schema, -1)
	typ := reflect.TypeOf(models.NHIAAccreditedHealthMaintenanceOrganisation{})
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
	requireContains(t, schema, "organisation_type: {type: string, enum: [health_maintenance_organisation]}")
	requireContains(t, schema, "accreditation_status: {type: string, enum: [accredited]}")
	requireContains(t, schema, "hmo_id: {type: string")

	pagination := nhiaHMODocBlock(t, doc, prefix+"Pagination")
	requireContains(t, pagination, "required: [page, page_size, total, total_pages]")
	requireNotContains(t, pagination, "limit:")
	for _, suffix := range []string{"ListResponse", "DetailResponse"} {
		block := nhiaHMODocBlock(t, doc, prefix+suffix)
		requireContains(t, block, "#/components/schemas/"+prefix)
		requireContains(t, block, "success: {type: boolean}")
	}
	requireContains(t, nhiaHMODocBlock(t, doc, prefix+"ListResponse"), "type: array")

	methodNotAllowed := nhiaHMODocBlock(t, doc, "MethodNotAllowedGet")
	for _, phrase := range []string{"description: Method not allowed", "Allow:", "Permitted HTTP method for this endpoint.", "enum: [GET]", "example: GET", "#/components/schemas/ErrorResponse"} {
		requireContains(t, methodNotAllowed, phrase)
	}
	for _, alias := range []string{"/v1/healthcare/hmos", "/v1/healthcare/health-maintenance-organizations", "/v1/healthcare/licensed-hmos", "/v1/healthcare/registered-hmos"} {
		requireNotContains(t, doc, alias)
	}
	for _, forbidden := range []string{"state_id", "lga_id", "address:", "website_url", "logo_url", "phone:", "email:", "contact_person", "operational_status", "licence_status", "registration_status"} {
		requireNotContains(t, list+detail+schema, forbidden)
	}
}

func TestNHIAAccreditedHMOOpenAPIOperationIDsUnique(t *testing.T) {
	raw, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	matches := regexp.MustCompile(`operationId: ([A-Za-z0-9_]+)`).FindAllStringSubmatch(string(raw), -1)
	seen := map[string]struct{}{}
	for _, match := range matches {
		if _, ok := seen[match[1]]; ok {
			t.Fatalf("duplicate operationId %q", match[1])
		}
		seen[match[1]] = struct{}{}
	}
}
