package postgres

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestEndpointUsageResponseTrimsRoute(t *testing.T) {
	row := endpointUsageResponse(pgtype.Text{String: " /v1/datasets/{dataset_id} ", Valid: true}, 7)
	if row.Endpoint != "/v1/datasets/{dataset_id}" {
		t.Fatalf("unexpected endpoint: %#v", row)
	}
	if row.RequestCount != 7 {
		t.Fatalf("unexpected request count: %#v", row)
	}
}

