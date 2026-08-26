package postgres

import (
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

func TestPublicDatasetListParamsUsesSearchLimitAndOffset(t *testing.T) {
	params := publicDatasetListParams(models.DatasetListFilter{
		Search: "states",
		Limit:  25,
		Offset: 50,
	})

	if params.Search != "states" {
		t.Fatalf("unexpected search: %q", params.Search)
	}
	if params.PageLimit != 25 {
		t.Fatalf("unexpected limit: %d", params.PageLimit)
	}
	if params.PageOffset != 50 {
		t.Fatalf("unexpected offset: %d", params.PageOffset)
	}
}
