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

func TestNHIAStateSocialHealthInsuranceAgencyOpenAPIContract(t *testing.T) {
	raw, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(raw)
	prefix := "NHIAStateSocialHealthInsuranceAgency"
	list := pathBlock(t, doc, "/v1/healthcare/nhia-state-social-health-insurance-agencies")
	detail := pathBlock(t, doc, "/v1/healthcare/nhia-state-social-health-insurance-agencies/{agency_id}")
	for _, tc := range []struct {
		block, id string
		params    []string
	}{{list, "listNHIAStateSocialHealthInsuranceAgencies", []string{"Page", "PageSize", "StateID", "Search"}}, {detail, "getNHIAStateSocialHealthInsuranceAgency", []string{"ID"}}} {
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
		for _, phrase := range []string{"dated NHIA-listed snapshot", "37 organisation records", "all 36 states and the Federal Capital Territory", "does not prove accreditation, licensing, registration, or operational status"} {
			requireContains(t, tc.block, phrase)
		}
		requireContains(t, tc.block, "not described here as production-wired")
	}
	for _, phrase := range []string{"No personal director or contact information", "addresses, websites, and logos are deferred", "no personal director, contact, address, website, or logo fields"} {
		requireContains(t, list+detail, phrase)
	}
	requireContains(t, detail, `"404":`)
	requireContains(t, detail, "malformed agency_id returns 400")
	requireContains(t, detail, "unknown agency_id returns 404")

	for _, tc := range []struct{ suffix, name, location, schema string }{
		{"ID", "agency_id", "path", "maxLength: " + strconv.Itoa(models.NHIAStateSocialHealthInsuranceAgencyIDMaxLength)},
		{"Page", "page", "query", "type: integer, minimum: 1, default: 1"},
		{"PageSize", "page_size", "query", "type: integer, minimum: 1, maximum: 100, default: 50"},
		{"StateID", "state_id", "query", "enum: [abia, adamawa"},
		{"Search", "search", "query", "maxLength: 100"},
	} {
		block := nhiaHMODocBlock(t, doc, prefix+tc.suffix)
		for _, value := range []string{"name: " + tc.name, "in: " + tc.location, tc.schema, "description:", "example:"} {
			requireContains(t, block, value)
		}
	}
	stateBlock := nhiaHMODocBlock(t, doc, prefix+"StateID")
	for _, value := range []string{"AKS", "CRS", "River State", "akwa-ibom", "cross-river", "fct", "rivers"} {
		requireContains(t, stateBlock, value)
	}

	schema := nhiaHMODocBlock(t, doc, prefix)
	fields := regexp.MustCompile(`(?m)^        ([a-z_]+):`).FindAllStringSubmatch(schema, -1)
	typ := reflect.TypeOf(models.NHIAStateSocialHealthInsuranceAgency{})
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
	requireContains(t, schema, "organisation_type: {type: string, enum: [state_social_health_insurance_agency]}")
	for _, forbidden := range []string{"director:", "staff:", "phone:", "phone_number:", "email:", "email_address:", "contact:", "contact_person:", "address:", "website:", "website_url:", "logo:", "logo_url:", "accreditation_status:", "licence_status:", "registration_status:", "operational_status:"} {
		requireNotContains(t, schema, forbidden)
	}

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
	for _, phrase := range []string{"description: Method not allowed", "Allow:", "enum: [GET]", "example: GET", "#/components/schemas/ErrorResponse"} {
		requireContains(t, methodNotAllowed, phrase)
	}
	for _, alias := range []string{"/v1/healthcare/sshias", "/v1/healthcare/state-health-insurance", "/v1/healthcare/state-health-insurance-agencies", "/v1/healthcare/accredited-sshias", "/v1/healthcare/licensed-sshias"} {
		requireNotContains(t, doc, alias)
	}
}
