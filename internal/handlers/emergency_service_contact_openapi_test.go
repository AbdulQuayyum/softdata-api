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

func TestEmergencyServiceContactOpenAPIContract(t *testing.T) {
	raw, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(raw)
	prefix := "EmergencyServiceContact"
	list := pathBlock(t, doc, "/v1/emergency/emergency-service-contacts")
	detail := pathBlock(t, doc, "/v1/emergency/emergency-service-contacts/{contact_id}")
	for _, tc := range []struct {
		block, id string
		params    []string
	}{
		{list, "listEmergencyServiceContacts", []string{"Page", "PageSize", "ServiceType", "ContactType", "CoverageType", "ContactValue", "Search"}},
		{detail, "getEmergencyServiceContact", []string{"ID"}},
	} {
		requireContains(t, tc.block, "tags: [Emergency]")
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
	for _, phrase := range []string{
		"Dated source-verified institutional contact snapshot",
		"not a live operational-status guarantee",
		"complete national, state, or local emergency directory",
		"were not called or operationally tested",
		"Availability and call-cost fields appear only where explicitly published",
		"Duplicate contact values, such as 112, may represent different official service contexts",
	} {
		requireContains(t, list+detail, phrase)
	}
	for _, forbidden := range []string{"currently reachable", "every contact is toll-free", "every contact operates 24 hours", "replaces all other emergency channels"} {
		requireNotContains(t, strings.ToLower(list+detail), forbidden)
	}

	for _, tc := range []struct{ suffix, name, location, schema string }{
		{"ID", "contact_id", "path", "maxLength: 255"},
		{"Page", "page", "query", "type: integer, minimum: 1, default: 1"},
		{"PageSize", "page_size", "query", "type: integer, minimum: 1, maximum: 100, default: 50"},
		{"ServiceType", "service_type", "query", "enum: [general_emergency, disaster_management, road_emergency, police, fire, ambulance, other]"},
		{"ContactType", "contact_type", "query", "enum: [short_code, telephone]"},
		{"CoverageType", "coverage_type", "query", "enum: [national, state]"},
		{"ContactValue", "contact_value", "query", "type: string"},
		{"Search", "search", "query", "maxLength: 100"},
	} {
		block := nhiaHMODocBlock(t, doc, prefix+tc.suffix)
		for _, value := range []string{"name: " + tc.name, "in: " + tc.location, tc.schema, "description:"} {
			requireContains(t, block, value)
		}
	}
	requireContains(t, nhiaHMODocBlock(t, doc, prefix+"ContactValue"), "not parsed as numbers or rewritten")

	schema := nhiaHMODocBlock(t, doc, prefix)
	fields := regexp.MustCompile(`(?m)^        ([a-z_]+):`).FindAllStringSubmatch(schema, -1)
	expectedFields := []string{"id", "service_name", "agency_name", "service_type", "contact_type", "contact_value", "coverage_type", "country_code", "availability", "call_cost", "notes"}
	if len(fields) != len(expectedFields) {
		t.Fatalf("schema fields %d want %d", len(fields), len(expectedFields))
	}
	model := reflect.TypeOf(models.EmergencyServiceContact{})
	modelFields := map[string]struct{}{}
	required := []string{}
	for i := 0; i < model.NumField(); i++ {
		tag := strings.Split(model.Field(i).Tag.Get("json"), ",")[0]
		modelFields[tag] = struct{}{}
		if !strings.Contains(model.Field(i).Tag.Get("json"), "omitempty") {
			required = append(required, tag)
		}
	}
	for _, field := range expectedFields {
		if _, ok := modelFields[field]; !ok {
			t.Fatalf("schema field %q is not in model", field)
		}
		requireContains(t, schema, "        "+field+":")
	}
	requireContains(t, schema, "required: ["+strings.Join(required, ", ")+"]")
	requireContains(t, schema, "contact_value: {type: string")
	requireContains(t, schema, "service_type: {type: string, enum: [general_emergency, disaster_management, road_emergency, police, fire, ambulance, other]}")
	for _, forbidden := range []string{"state_id:", "lga_id:", "address:", "website:", "logo:", "phone:", "email:", "source:", "operational_status:"} {
		requireNotContains(t, schema, forbidden)
	}
	for _, suffix := range []string{"ListResponse", "DetailResponse"} {
		block := nhiaHMODocBlock(t, doc, prefix+suffix)
		requireContains(t, block, "#/components/schemas/"+prefix)
		requireContains(t, block, "success: {type: boolean}")
	}
	requireContains(t, nhiaHMODocBlock(t, doc, prefix+"Pagination"), "required: [page, page_size, total, total_pages]")
	requireContains(t, nhiaHMODocBlock(t, doc, "MethodNotAllowedGet"), "Allow:")
}

func TestEmergencyServiceContactPostmanContract(t *testing.T) {
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
	var emergency []struct {
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
	}
	for _, folder := range collection.Item {
		if folder.Name == "Emergency" {
			emergency = folder.Item
		}
	}
	got := map[string]string{}
	for _, item := range emergency {
		if item.Request.Method != http.MethodGet {
			t.Fatalf("method for %s = %s", item.Name, item.Request.Method)
		}
		for _, header := range item.Request.Header {
			if strings.EqualFold(header.Key, "X-API-Key") && (!header.Disabled || header.Value != "{{apiKey}}") {
				t.Fatalf("bad API key header: %#v", header)
			}
		}
		got[item.Name] = item.Request.URL.Raw
		if strings.Contains(strings.ToLower(item.Request.URL.Raw), "secret") {
			t.Fatal("secret marker in Postman emergency URL")
		}
	}
	if got["List emergency service contacts"] == "" || got["Get an emergency service contact"] == "" {
		t.Fatalf("expected emergency-contact requests in Emergency folder, got %#v", got)
	}
	if got["List emergency service contacts"] != "{{baseUrl}}/v1/emergency/emergency-service-contacts?page={{page}}&page_size={{page_size}}&service_type={{emergencyServiceType}}&contact_type={{emergencyContactType}}&coverage_type={{emergencyCoverageType}}&contact_value={{emergencyContactValue}}&search={{emergencySearch}}" {
		t.Fatalf("bad list URL: %q", got["List emergency service contacts"])
	}
	if got["Get an emergency service contact"] != "{{baseUrl}}/v1/emergency/emergency-service-contacts/{{contact_id}}" {
		t.Fatalf("bad detail URL: %q", got["Get an emergency service contact"])
	}
	values := map[string]string{}
	for _, variable := range collection.Variable {
		values[variable.Key] = variable.Value
	}
	if values["contact_id"] != "federal-fire-service-fire-112-national" || values["emergencyContactValue"] != "112" {
		t.Fatalf("bad emergency variables: %#v", values)
	}
	for _, key := range []string{"apiKey", "accessToken"} {
		if values[key] != "" {
			t.Fatalf("secret-like variable %s has value", key)
		}
	}
}

func TestEmergencyServiceContactOperationIDsRemainUnique(t *testing.T) {
	raw, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	matches := regexp.MustCompile(`(?m)^\s+operationId:\s+([A-Za-z0-9_]+)\s*$`).FindAllStringSubmatch(string(raw), -1)
	seen := map[string]int{}
	for _, match := range matches {
		seen[match[1]]++
	}
	for operationID, count := range seen {
		if count != 1 {
			t.Fatalf("duplicate operationId %s count=%d", operationID, count)
		}
	}
	if seen["listEmergencyServiceContacts"] != 1 || seen["getEmergencyServiceContact"] != 1 {
		t.Fatalf("missing emergency operation IDs: %#v", seen)
	}
}

func TestEmergencyServiceContactOpenAPIQueryParameters(t *testing.T) {
	raw, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	block := pathBlock(t, string(raw), "/v1/emergency/emergency-service-contacts")
	params := regexp.MustCompile(`#/components/parameters/EmergencyServiceContact([A-Za-z]+)`).FindAllStringSubmatch(block, -1)
	got := []string{}
	for _, param := range params {
		got = append(got, param[1])
	}
	sort.Strings(got)
	want := []string{"ContactType", "ContactValue", "CoverageType", "Page", "PageSize", "Search", "ServiceType"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("params=%v want %v", got, want)
	}
}
