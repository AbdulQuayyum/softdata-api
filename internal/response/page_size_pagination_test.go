package response

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPaginatedPageSize(t *testing.T) {
	for _, data := range [][]string{nil, {}} {
		w := httptest.NewRecorder()
		err := PaginatedPageSize(w, 200, data, PageSizePaginationMeta{Page: 1, PageSize: 50})
		if err != nil || w.Code != 200 || strings.TrimSpace(w.Body.String()) != `{"success":true,"data":[],"meta":{"page":1,"page_size":50,"total":0,"total_pages":0}}` {
			t.Fatalf("%v %s", err, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	if err := PaginatedPageSize(w, 200, []string{}, PageSizePaginationMeta{}); err == nil || w.Body.Len() != 0 {
		t.Fatal("invalid metadata wrote response")
	}
}
