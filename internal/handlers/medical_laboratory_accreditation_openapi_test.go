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

func accreditationDocBlock(t *testing.T, doc, name string) string {
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
func TestAccreditationOpenAPIContract(t *testing.T) {
	raw, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(raw)
	prefix := "MedicalLaboratoryAccreditation"
	list := pathBlock(t, doc, "/v1/healthcare/medical-laboratory-accreditations")
	detail := pathBlock(t, doc, "/v1/healthcare/medical-laboratory-accreditations/{accreditation_id}")
	for _, tc := range []struct {
		block, id string
		params    []string
	}{{list, "listMedicalLaboratoryAccreditations", []string{"Page", "PageSize", "StateID", "Status", "Search"}}, {detail, "getMedicalLaboratoryAccreditation", []string{"ID"}}} {
		requireContains(t, tc.block, "operationId: "+tc.id)
		if strings.Count(tc.block, "#/components/parameters/") != len(tc.params) {
			t.Fatal("unexpected parameters")
		}
		for _, p := range tc.params {
			requireContains(t, tc.block, "#/components/parameters/"+prefix+p)
		}
		for _, status := range []string{`"400":`, `"500":`, `"503":`} {
			requireContains(t, tc.block, status)
		}
		for _, phrase := range []string{"A dated snapshot of medical laboratory facility accreditation records published by the MLSCN Accreditation Service.", "30 records", "26 were marked accredited", "4 were marked expired", "not a complete register of licensed medical-laboratory premises", "An expired entry is not currently accredited", "no personal practitioner data"} {
			requireContains(t, tc.block, phrase)
		}
	}
	requireContains(t, detail, `"404":`)
	for _, tc := range []struct{ suffix, name, location, schema string }{
		{"ID", "accreditation_id", "path", "maxLength: " + strconv.Itoa(models.MedicalLaboratoryAccreditationIDMaxLength)},
		{"Page", "page", "query", "type: integer, minimum: 1, default: 1"},
		{"PageSize", "page_size", "query", "type: integer, minimum: 1, maximum: 100, default: 50"},
		{"StateID", "state_id", "query", "pattern: '^[a-z0-9]+(?:-[a-z0-9]+)*$'"},
		{"Status", "accreditation_status", "query", "enum: [accredited, expired]"},
		{"Search", "search", "query", "maxLength: 100"},
	} {
		b := accreditationDocBlock(t, doc, prefix+tc.suffix)
		for _, v := range []string{"name: " + tc.name, "in: " + tc.location, tc.schema, "description:", "example:"} {
			requireContains(t, b, v)
		}
	}
	schema := accreditationDocBlock(t, doc, prefix)
	fields := regexp.MustCompile(`(?m)^        ([a-z_]+):`).FindAllStringSubmatch(schema, -1)
	typ := reflect.TypeOf(models.MedicalLaboratoryAccreditation{})
	if len(fields) != typ.NumField() {
		t.Fatalf("schema fields %d model %d", len(fields), typ.NumField())
	}
	required := []string{}
	for i := 0; i < typ.NumField(); i++ {
		tag := strings.Split(typ.Field(i).Tag.Get("json"), ",")
		requireContains(t, schema, "        "+tag[0]+":")
		if len(tag) == 1 {
			required = append(required, tag[0])
		}
	}
	requireContains(t, schema, "required: ["+strings.Join(required, ", ")+"]")
	requireContains(t, schema, "enum: [accredited, expired]")
	for _, field := range []string{"approval_date", "expiry_date"} {
		requireContains(t, schema, field+": {type: string, format: date")
	}
	pagination := accreditationDocBlock(t, doc, prefix+"Pagination")
	requireContains(t, pagination, "required: [page, page_size, total, total_pages]")
	requireNotContains(t, pagination, "limit:")
	for _, suffix := range []string{"ListResponse", "DetailResponse"} {
		b := accreditationDocBlock(t, doc, prefix+suffix)
		requireContains(t, b, "#/components/schemas/"+prefix)
		requireContains(t, b, "success: {type: boolean}")
	}
	requireContains(t, accreditationDocBlock(t, doc, prefix+"ListResponse"), "type: array")
	for _, bad := range []string{"registration_status", "mlscn_premises_number", "licence_year", "laboratory_type", "ownership_type", "lga_id", "operational_status", "website", "logo", "practitioner_name", "phone", "email", "licensed laboratories"} {
		requireNotContains(t, list+detail+schema, bad)
	}
	requireNotContains(t, doc, "/v1/healthcare/licensed-medical-laboratories")
}
