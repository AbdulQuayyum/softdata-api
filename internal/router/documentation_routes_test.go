package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type postmanCollection struct {
	Info     postmanInfo       `json:"info"`
	Variable []postmanVariable `json:"variable"`
	Item     []postmanItem     `json:"item"`
}

type postmanInfo struct {
	Description string `json:"description"`
}

type postmanVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type postmanItem struct {
	Name    string          `json:"name"`
	Item    []postmanItem   `json:"item"`
	Request *postmanRequest `json:"request"`
}

type postmanRequest struct {
	Method string          `json:"method"`
	Header []postmanHeader `json:"header"`
	URL    postmanURL      `json:"url"`
	Auth   json.RawMessage `json:"auth"`
}

type postmanHeader struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Disabled bool   `json:"disabled"`
}

type postmanURL struct {
	Raw string `json:"raw"`
}

func TestDocumentationRoutesServePublicArtifacts(t *testing.T) {
	rec := &routerRecorder{}
	r := newTestRouter(t, rec)

	for _, tc := range []struct {
		path        string
		contentType string
		contains    string
	}{
		{path: "/", contentType: "text/html", contains: "/docs"},
		{path: "/docs", contentType: "text/html", contains: `data-url="/openapi.yaml"`},
		{path: "/openapi.yaml", contentType: "application/yaml", contains: "openapi: 3.1.0"},
		{path: "/postman.json", contentType: "application/json", contains: "SoftData API - Current Routes"},
	} {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d body = %s", tc.path, rr.Code, rr.Body.String())
		}
		if got := rr.Header().Get("Content-Type"); !strings.Contains(got, tc.contentType) {
			t.Fatalf("GET %s content type = %q, want %q", tc.path, got, tc.contentType)
		}
		if !strings.Contains(rr.Body.String(), tc.contains) {
			t.Fatalf("GET %s body missing %q", tc.path, tc.contains)
		}
		if containsSecretLikeValue(rr.Body.String()) {
			t.Fatalf("GET %s contains a populated credential-looking value", tc.path)
		}
	}

	joined := strings.Join(rec.snapshot(), ",")
	if strings.Contains(joined, "optional_api_key") || strings.Contains(joined, "usage:") {
		t.Fatalf("documentation routes should not use API-key identification or usage tracking: %v", rec.snapshot())
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/docs", nil))
	if rr.Code != http.StatusMethodNotAllowed || rr.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("POST /docs status=%d allow=%q body=%s", rr.Code, rr.Header().Get("Allow"), rr.Body.String())
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("POST /docs content type = %q", rr.Header().Get("Content-Type"))
	}
}

func TestDocumentationArtifactsMatchCommittedFiles(t *testing.T) {
	r := newTestRouter(t, &routerRecorder{})
	for _, tc := range []struct {
		route string
		file  string
	}{
		{route: "/openapi.yaml", file: "../../docs/openapi.yaml"},
		{route: "/postman.json", file: "../../docs/softdata-api.postman_collection.json"},
	} {
		want, err := os.ReadFile(tc.file)
		if err != nil {
			t.Fatalf("read %s: %v", tc.file, err)
		}
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, tc.route, nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d", tc.route, rr.Code)
		}
		if got := rr.Body.Bytes(); string(got) != string(want) {
			t.Fatalf("GET %s did not match committed %s", tc.route, tc.file)
		}
	}
}

func TestOpenAPIPostmanParity(t *testing.T) {
	openAPIRaw := readTestFile(t, "../../docs/openapi.yaml")
	postmanRaw := readTestFile(t, "../../docs/softdata-api.postman_collection.json")

	openAPIOps := extractOpenAPIOperations(t, string(openAPIRaw))
	postmanOps, collection := extractPostmanOperations(t, postmanRaw)

	if !strings.Contains(collection.Info.Description, "docs/openapi.yaml is the canonical API contract") {
		t.Fatalf("Postman description does not name canonical OpenAPI contract")
	}

	if missing := subtractOperations(openAPIOps, postmanOps); len(missing) > 0 {
		t.Fatalf("OpenAPI operations missing from Postman: %v", missing)
	}
	if extra := subtractOperations(postmanOps, openAPIOps); len(extra) > 0 {
		t.Fatalf("Postman operations absent from OpenAPI: %v", extra)
	}

	for _, op := range expectedHealthcareOperations() {
		if _, ok := postmanOps[op]; !ok {
			t.Fatalf("healthcare operation missing from Postman: %s", op)
		}
		if !strings.HasPrefix(op.path, "/v1/healthcare/") || op.method != http.MethodGet {
			t.Fatalf("bad healthcare operation expectation: %s", op)
		}
	}

	for op, count := range postmanOps {
		if count != 1 {
			t.Fatalf("duplicate Postman operation %s count=%d", op, count)
		}
	}

	combined := strings.ToLower(string(openAPIRaw) + "\n" + string(postmanRaw))
	for _, deferred := range []string{"pcn", "nafdac", "greenbook", "mdcn", "housemanship"} {
		if strings.Contains(combined, "/v1/healthcare/"+deferred) || strings.Contains(combined, "/v1/"+deferred) {
			t.Fatalf("deferred route %q appears in public docs", deferred)
		}
	}

	assertNoPopulatedCollectionSecrets(t, collection)
	assertOperationIDsUnique(t, string(openAPIRaw))
}

func readTestFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

type operation struct {
	method string
	path   string
}

func (o operation) String() string {
	return o.method + " " + o.path
}

func extractOpenAPIOperations(t *testing.T, doc string) map[operation]int {
	t.Helper()
	ops := map[operation]int{}
	pathPattern := regexp.MustCompile(`^  (/[^:]*):\s*$`)
	methodPattern := regexp.MustCompile(`^    (get|post|put|patch|delete):\s*$`)
	currentPath := ""
	for _, line := range strings.Split(doc, "\n") {
		if match := pathPattern.FindStringSubmatch(line); match != nil {
			currentPath = match[1]
			continue
		}
		match := methodPattern.FindStringSubmatch(line)
		if match == nil || currentPath == "" {
			continue
		}
		ops[operation{method: strings.ToUpper(match[1]), path: currentPath}]++
	}
	if len(ops) == 0 {
		t.Fatal("no OpenAPI operations found")
	}
	return ops
}

func extractPostmanOperations(t *testing.T, raw []byte) (map[operation]int, postmanCollection) {
	t.Helper()
	var collection postmanCollection
	if err := json.Unmarshal(raw, &collection); err != nil {
		t.Fatalf("parse Postman collection: %v", err)
	}
	ops := map[operation]int{}
	var walk func([]postmanItem)
	walk = func(items []postmanItem) {
		for _, item := range items {
			if len(item.Item) > 0 {
				walk(item.Item)
				continue
			}
			if item.Request == nil {
				continue
			}
			ops[operation{method: strings.ToUpper(item.Request.Method), path: normalizePostmanPath(item.Request.URL.Raw)}]++
		}
	}
	walk(collection.Item)
	if len(ops) == 0 {
		t.Fatal("no Postman operations found")
	}
	return ops, collection
}

func normalizePostmanPath(raw string) string {
	path := strings.TrimPrefix(strings.TrimSpace(raw), "{{baseUrl}}")
	if idx := strings.IndexByte(path, '?'); idx >= 0 {
		path = path[:idx]
	}
	path = regexp.MustCompile(`\{\{([^}]+)\}\}`).ReplaceAllString(path, `{$1}`)
	if path == "" {
		return "/"
	}
	return path
}

func subtractOperations(left, right map[operation]int) []operation {
	missing := make([]operation, 0)
	for op := range left {
		if _, ok := right[op]; !ok {
			missing = append(missing, op)
		}
	}
	sort.Slice(missing, func(i, j int) bool { return missing[i].String() < missing[j].String() })
	return missing
}

func expectedHealthcareOperations() []operation {
	return []operation{
		{http.MethodGet, "/v1/healthcare/health-facilities"},
		{http.MethodGet, "/v1/healthcare/health-facilities/{facility_id}"},
		{http.MethodGet, "/v1/healthcare/medical-laboratory-accreditations"},
		{http.MethodGet, "/v1/healthcare/medical-laboratory-accreditations/{accreditation_id}"},
		{http.MethodGet, "/v1/healthcare/nhia-accredited-health-maintenance-organisations"},
		{http.MethodGet, "/v1/healthcare/nhia-accredited-health-maintenance-organisations/{organisation_id}"},
		{http.MethodGet, "/v1/healthcare/nhia-state-social-health-insurance-agencies"},
		{http.MethodGet, "/v1/healthcare/nhia-state-social-health-insurance-agencies/{agency_id}"},
		{http.MethodGet, "/v1/healthcare/nhia-active-accredited-healthcare-providers"},
		{http.MethodGet, "/v1/healthcare/nhia-active-accredited-healthcare-providers/{provider_id}"},
	}
}

func assertOperationIDsUnique(t *testing.T, doc string) {
	t.Helper()
	matches := regexp.MustCompile(`(?m)^\s+operationId:\s*([A-Za-z0-9_]+)\s*$`).FindAllStringSubmatch(doc, -1)
	seen := map[string]struct{}{}
	for _, match := range matches {
		if _, ok := seen[match[1]]; ok {
			t.Fatalf("duplicate OpenAPI operationId %q", match[1])
		}
		seen[match[1]] = struct{}{}
	}
}

func assertNoPopulatedCollectionSecrets(t *testing.T, collection postmanCollection) {
	t.Helper()
	for _, variable := range collection.Variable {
		switch strings.ToLower(variable.Key) {
		case "apikey", "api_key", "accesstoken", "access_token", "token", "cookie", "password":
			if strings.TrimSpace(variable.Value) != "" {
				t.Fatalf("secret-like Postman variable %q has a populated value", variable.Key)
			}
		}
	}
	var walk func([]postmanItem)
	walk = func(items []postmanItem) {
		for _, item := range items {
			if len(item.Item) > 0 {
				walk(item.Item)
				continue
			}
			if item.Request == nil {
				continue
			}
			for _, header := range item.Request.Header {
				key := strings.ToLower(header.Key)
				value := strings.TrimSpace(header.Value)
				if key == "x-api-key" && value != "{{apiKey}}" {
					t.Fatalf("X-API-Key header has non-placeholder value")
				}
				if (key == "authorization" || key == "cookie") && value != "" && !strings.Contains(value, "{{") {
					t.Fatalf("credential header %q has non-placeholder value", header.Key)
				}
			}
			if len(item.Request.Auth) > 0 {
				auth := string(item.Request.Auth)
				if !strings.Contains(auth, "{{accessToken}}") || containsSecretLikeValue(auth) {
					t.Fatalf("request %q defines populated request-level auth", item.Name)
				}
			}
		}
	}
	walk(collection.Item)
}

func containsSecretLikeValue(body string) bool {
	for _, pattern := range []string{"sd_live_", "Bearer eyJ", "AKIA", "BEGIN PRIVATE KEY", "Set-Cookie:"} {
		if strings.Contains(body, pattern) {
			return true
		}
	}
	return false
}
