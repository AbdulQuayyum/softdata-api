package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuccessWritesDocumentedEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()

	if err := Success(rr, http.StatusOK, map[string]string{"id": "ng-states"}); err != nil {
		t.Fatalf("Success() error = %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != jsonContentType {
		t.Fatalf("unexpected content type: %q", got)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body["success"] != true {
		t.Fatalf("unexpected success flag: %#v", body["success"])
	}
	data := body["data"].(map[string]any)
	if data["id"] != "ng-states" {
		t.Fatalf("unexpected payload: %#v", data)
	}
}

func TestListNormalizesNilSliceToEmptyArray(t *testing.T) {
	rr := httptest.NewRecorder()

	if err := List[string](rr, http.StatusOK, nil); err != nil {
		t.Fatalf("List() error = %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	items, ok := body["data"].([]any)
	if !ok {
		t.Fatalf("data was not an array: %#v", body["data"])
	}
	if len(items) != 0 {
		t.Fatalf("expected empty array, got %#v", items)
	}
}

func TestPaginatedWritesMetaWithoutRecomputing(t *testing.T) {
	rr := httptest.NewRecorder()
	meta := PaginationMeta{Page: 2, Limit: 20, Total: 41, TotalPages: 3}

	if err := Paginated[string](rr, http.StatusOK, nil, meta); err != nil {
		t.Fatalf("Paginated() error = %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	metaBody := body["meta"].(map[string]any)
	if metaBody["page"].(float64) != 2 || metaBody["limit"].(float64) != 20 || metaBody["total_pages"].(float64) != 3 {
		t.Fatalf("unexpected meta: %#v", metaBody)
	}
	items := body["data"].([]any)
	if len(items) != 0 {
		t.Fatalf("expected empty data array, got %#v", items)
	}
}

func TestPaginatedRejectsInvalidMetadata(t *testing.T) {
	rr := httptest.NewRecorder()
	if err := Paginated[string](rr, http.StatusOK, nil, PaginationMeta{Page: 0, Limit: 20, Total: 1, TotalPages: 1}); err == nil {
		t.Fatal("Paginated() error = nil, want error")
	}
}

func TestNoContentWritesNoBody(t *testing.T) {
	rr := httptest.NewRecorder()

	NoContent(rr, http.StatusNoContent)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("expected no body, got %q", rr.Body.String())
	}
}
