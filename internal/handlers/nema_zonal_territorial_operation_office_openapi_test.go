package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

func TestNEMAZonalTerritorialOperationOfficeOpenAPIContract(t *testing.T) {
	raw, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(raw)
	prefix := "NEMAZonalTerritorialOperationOffice"
	list := pathBlock(t, doc, "/v1/emergency/nema-zonal-territorial-operation-offices")
	detail := pathBlock(t, doc, "/v1/emergency/nema-zonal-territorial-operation-offices/{office_id}")
	for _, tc := range []struct {
		block, id string
		params    []string
	}{
		{list, "listNEMAZonalTerritorialOperationOffices", []string{"Page", "PageSize", "StateID", "Type", "Search"}},
		{detail, "getNEMAZonalTerritorialOperationOffice", []string{"ID"}},
	} {
		requireContains(t, tc.block, "tags: [Emergency]")
		requireContains(t, tc.block, "operationId: "+tc.id)
		if strings.Count(tc.block, "#/components/parameters/") != len(tc.params) {
			t.Fatalf("unexpected parameter count in %s", tc.id)
		}
		for _, p := range tc.params {
			requireContains(t, tc.block, "#/components/parameters/"+prefix+p)
		}
		for _, status := range []string{`"400":`, `"405": {$ref: "#/components/responses/MethodNotAllowedGet"}`, `"500":`, `"503":`} {
			requireContains(t, tc.block, status)
		}
	}
	requireContains(t, detail, `"404":`)
	for _, phrase := range []string{
		"dated 17-record NEMA Zonal, Territorial and Operation Offices Snapshot",
		"narrower source-described",
		"not a complete register of all NEMA offices",
		"does not guarantee current operational availability",
		"No contact details are published",
		"no NEMA contact was called or messaged",
		"Production-wired public GET route",
		"standard Emergency public middleware",
	} {
		requireContains(t, list+detail, phrase)
	}
	for _, stale := range []string{"production route registration and startup verification are not part of this phase", "production registration is deferred", "not production-wired", "http contract only"} {
		requireNotContains(t, strings.ToLower(list+detail), stale)
	}

	for _, tc := range []struct{ suffix, name, location, schema string }{
		{"ID", "office_id", "path", "maxLength: 255"},
		{"Page", "page", "query", "type: integer, minimum: 1, default: 1"},
		{"PageSize", "page_size", "query", "type: integer, minimum: 1, maximum: 100, default: 50"},
		{"StateID", "state_id", "query", "#/components/schemas/NEMAZonalTerritorialOperationOfficeStateID"},
		{"Type", "office_type", "query", "enum: [zonal_territorial_operation_office]"},
		{"Search", "search", "query", "maxLength: 100"},
	} {
		block := nhiaHMODocBlock(t, doc, prefix+tc.suffix)
		for _, value := range []string{"name: " + tc.name, "in: " + tc.location, tc.schema, "description:"} {
			requireContains(t, block, value)
		}
	}

	schema := nhiaHMODocBlock(t, doc, prefix)
	fields := regexp.MustCompile(`(?m)^        ([a-z_]+):`).FindAllStringSubmatch(schema, -1)
	expectedFields := []string{"id", "name", "office_type", "state_id", "country_code"}
	if len(fields) != len(expectedFields) {
		t.Fatalf("schema fields %d want %d", len(fields), len(expectedFields))
	}
	model := reflect.TypeOf(models.NEMAZonalTerritorialOperationOffice{})
	modelFields := map[string]struct{}{}
	required := []string{}
	for i := 0; i < model.NumField(); i++ {
		tag := strings.Split(model.Field(i).Tag.Get("json"), ",")[0]
		modelFields[tag] = struct{}{}
		required = append(required, tag)
	}
	for _, field := range expectedFields {
		if _, ok := modelFields[field]; !ok {
			t.Fatalf("schema field %q is not in model", field)
		}
		requireContains(t, schema, "        "+field+":")
	}
	requireContains(t, schema, "required: ["+strings.Join(required, ", ")+"]")
	for _, forbidden := range []string{"lga_id:", "address:", "telephone:", "phone:", "email:", "coordinates:", "coverage_area:", "operational_status:", "website:", "logo:", "contact_value:"} {
		requireNotContains(t, schema, forbidden)
	}
	requireContains(t, nhiaHMODocBlock(t, doc, prefix+"Pagination"), "required: [page, page_size, total, total_pages]")
	requireContains(t, nhiaHMODocBlock(t, doc, "MethodNotAllowedGet"), "Allow:")
}

func TestNEMAZonalTerritorialOperationOfficePostmanContract(t *testing.T) {
	raw, err := os.ReadFile("../../docs/softdata-api.postman_collection.json")
	if err != nil {
		t.Fatal(err)
	}
	var collection struct {
		Item []struct {
			Name string `json:"name"`
			Item []struct {
				Name    string `json:"name"`
				Request struct {
					Method string `json:"method"`
					Header []struct {
						Key      string `json:"key"`
						Value    string `json:"value"`
						Disabled bool   `json:"disabled"`
					} `json:"header"`
					URL struct {
						Raw   string   `json:"raw"`
						Path  []string `json:"path"`
						Query []struct {
							Key      string `json:"key"`
							Value    string `json:"value"`
							Disabled bool   `json:"disabled"`
						} `json:"query"`
					} `json:"url"`
				} `json:"request"`
			} `json:"item"`
		} `json:"item"`
		Variable []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"variable"`
	}
	if err := json.Unmarshal(raw, &collection); err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	seen := map[string]int{}
	for _, folder := range collection.Item {
		for _, item := range folder.Item {
			if item.Request.Method == "" {
				continue
			}
			key := item.Request.Method + " " + normalizePostmanURLForNEMATest(item.Request.URL.Raw)
			seen[key]++
			if strings.Contains(item.Request.URL.Raw, "nema-zonal-territorial-operation-offices") {
				if item.Request.Method != http.MethodGet {
					t.Fatalf("method for %s = %s", item.Name, item.Request.Method)
				}
				for _, header := range item.Request.Header {
					if strings.EqualFold(header.Key, "X-API-Key") && (!header.Disabled || header.Value != "{{apiKey}}") {
						t.Fatalf("bad API key header: %#v", header)
					}
				}
				got[item.Name] = item.Request.URL.Raw
			}
		}
	}
	if got["List NEMA Zonal, Territorial and Operation offices"] != "{{baseUrl}}/v1/emergency/nema-zonal-territorial-operation-offices?page={{page}}&page_size={{page_size}}&state_id={{nemaOfficeStateId}}&office_type={{nemaOfficeType}}&search={{nemaOfficeSearch}}" {
		t.Fatalf("bad list URL: %q", got["List NEMA Zonal, Territorial and Operation offices"])
	}
	if got["Get a NEMA Zonal, Territorial and Operation office"] != "{{baseUrl}}/v1/emergency/nema-zonal-territorial-operation-offices/{{office_id}}" {
		t.Fatalf("bad detail URL: %q", got["Get a NEMA Zonal, Territorial and Operation office"])
	}
	for op, count := range seen {
		if strings.Contains(op, "nema-zonal-territorial-operation-offices") && count != 1 {
			t.Fatalf("duplicate Postman NEMA operation %s count=%d", op, count)
		}
	}
	values := map[string]string{}
	for _, variable := range collection.Variable {
		values[variable.Key] = variable.Value
	}
	if values["office_id"] != "nema-lagos-zonal-territorial-operation-office" || values["nemaOfficeStateId"] != "lagos" || values["nemaOfficeType"] != "zonal_territorial_operation_office" {
		t.Fatalf("bad NEMA variables: %#v", values)
	}
	for _, key := range []string{"apiKey", "accessToken"} {
		if values[key] != "" {
			t.Fatalf("secret-like variable %s has value", key)
		}
	}
}

func TestNEMAZonalTerritorialOperationOfficeOpenAPIQueryParameters(t *testing.T) {
	raw, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	block := pathBlock(t, string(raw), "/v1/emergency/nema-zonal-territorial-operation-offices")
	params := regexp.MustCompile(`#/components/parameters/NEMAZonalTerritorialOperationOffice([A-Za-z]+)`).FindAllStringSubmatch(block, -1)
	got := []string{}
	for _, param := range params {
		got = append(got, param[1])
	}
	sort.Strings(got)
	want := []string{"Page", "PageSize", "Search", "StateID", "Type"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("params=%v want %v", got, want)
	}
}

func normalizePostmanURLForNEMATest(raw string) string {
	path := strings.TrimPrefix(strings.TrimSpace(raw), "{{baseUrl}}")
	if idx := strings.IndexByte(path, '?'); idx >= 0 {
		path = path[:idx]
	}
	return regexp.MustCompile(`\{\{([^}]+)\}\}`).ReplaceAllString(path, `{$1}`)
}
