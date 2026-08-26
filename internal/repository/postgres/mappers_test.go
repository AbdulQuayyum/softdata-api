package postgres

import (
	"errors"
	"net/netip"
	"reflect"
	"testing"
	"time"

	sqlc "github.com/AbdulQuayyum/softdata-api/internal/database/sqlc"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestDatasetFromRowCopiesSlicesAndKeepsIdentifiers(t *testing.T) {
	id, err := uuidFromString("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("uuidFromString: %v", err)
	}

	formats := []string{"json", "csv"}
	maintainers := []string{"Alice", "Bob"}
	now := time.Date(2026, time.August, 26, 10, 30, 0, 0, time.UTC)
	row := sqlc.Dataset{
		ID:          id,
		DatasetKey:  "ng-states",
		Slug:        "nigeria-states",
		Name:        "Nigerian States",
		Formats:     formats,
		Maintainers: maintainers,
		IsPublic:    true,
		CreatedAt:   timestamptzFromTimePtr(&now),
		UpdatedAt:   timestamptzFromTimePtr(&now),
	}

	model := datasetFromRow(row)
	if model.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("unexpected id: %s", model.ID)
	}
	if model.DatasetKey != "ng-states" {
		t.Fatalf("unexpected dataset key: %s", model.DatasetKey)
	}
	if !reflect.DeepEqual(model.Formats, []string{"json", "csv"}) {
		t.Fatalf("unexpected formats: %#v", model.Formats)
	}
	if !reflect.DeepEqual(model.Maintainers, []string{"Alice", "Bob"}) {
		t.Fatalf("unexpected maintainers: %#v", model.Maintainers)
	}

	formats[0] = "xml"
	maintainers[0] = "Carol"
	if model.Formats[0] != "json" {
		t.Fatalf("formats slice was not copied")
	}
	if model.Maintainers[0] != "Alice" {
		t.Fatalf("maintainers slice was not copied")
	}
}

func TestAPIRequestFromRowParsesQueryParams(t *testing.T) {
	requestID := "req_123"
	now := time.Date(2026, time.August, 26, 11, 0, 0, 0, time.UTC)
	addr := netip.MustParseAddr("127.0.0.1")
	accountID, err := uuidFromString("550e8400-e29b-41d4-a716-446655440001")
	if err != nil {
		t.Fatalf("uuidFromString account: %v", err)
	}
	apiKeyID, err := uuidFromString("550e8400-e29b-41d4-a716-446655440002")
	if err != nil {
		t.Fatalf("uuidFromString api key: %v", err)
	}
	anonymousID, err := uuidFromString("550e8400-e29b-41d4-a716-446655440003")
	if err != nil {
		t.Fatalf("uuidFromString anonymous: %v", err)
	}

	row := sqlc.ApiRequest{
		ID:             42,
		RequestID:      requestID,
		AccountID:      accountID,
		ApiKeyID:       apiKeyID,
		AnonymousID:    anonymousID,
		Method:         "GET",
		Path:           "/v1/datasets/ng-states",
		Route:          textFromString("/v1/datasets/{dataset_id}"),
		QueryParams:    []byte(`{"include":"sources"}`),
		StatusCode:     200,
		IpAddress:      &addr,
		UserAgent:      textFromString("test-agent"),
		ResponseTimeMs: pgtypeInt4(123),
		RequestBytes:   pgtypeInt8(456),
		ResponseBytes:  pgtypeInt8(789),
		CreatedAt:      timestamptzFromTimePtr(&now),
	}

	model, err := apiRequestFromRow(row)
	if err != nil {
		t.Fatalf("apiRequestFromRow: %v", err)
	}
	if model.RequestID != requestID {
		t.Fatalf("unexpected request id: %s", model.RequestID)
	}
	if got := model.QueryParams["include"]; got != "sources" {
		t.Fatalf("unexpected query params: %#v", model.QueryParams)
	}
	if model.IPAddress == nil || *model.IPAddress != "127.0.0.1" {
		t.Fatalf("unexpected ip address: %#v", model.IPAddress)
	}
	if model.AccountID == nil || *model.AccountID != accountID.String() {
		t.Fatalf("unexpected account id: %#v", model.AccountID)
	}
}

func TestTranslateError(t *testing.T) {
	if !errors.Is(translateError("missing", pgx.ErrNoRows), interfaces.ErrNotFound) {
		t.Fatalf("expected not found sentinel")
	}

	conflictErr := &pgconn.PgError{Code: "23505", ConstraintName: "accounts_username_unique"}
	if !errors.Is(translateError("conflict", conflictErr), interfaces.ErrConflict) {
		t.Fatalf("expected conflict sentinel")
	}
}

func pgtypeInt4(value int32) pgtype.Int4 {
	return pgtype.Int4{Int32: value, Valid: true}
}

func pgtypeInt8(value int64) pgtype.Int8 {
	return pgtype.Int8{Int64: value, Valid: true}
}
