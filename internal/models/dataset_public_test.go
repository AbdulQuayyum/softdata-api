package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDatasetSourceResponseJSONOmitsInternalIdentifiers(t *testing.T) {
	now := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)
	resp := DatasetSourceResponse{
		ID:             "source-example",
		Name:           "Ministry of Examples",
		URL:            strPtr("https://example.test"),
		IsOfficial:     true,
		LastFetchedAt:  &now,
		LastVerifiedAt: &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if got["id"] != "source-example" {
		t.Fatalf("unexpected id: %#v", got["id"])
	}
	if _, ok := got["dataset_id"]; ok {
		t.Fatalf("dataset_id should not be exposed")
	}
	if _, ok := got["source_key"]; ok {
		t.Fatalf("source_key should not be exposed")
	}
}

func TestDatasetVersionResponseJSONOmitsInternalIdentifiers(t *testing.T) {
	now := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)
	resp := DatasetVersionResponse{
		Version:     "1.0.0",
		Format:      "json",
		Status:      DatasetVersionStatusPublished,
		RecordCount: 42,
		ReleasedAt:  &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if got["version"] != "1.0.0" {
		t.Fatalf("unexpected version: %#v", got["version"])
	}
	if _, ok := got["id"]; ok {
		t.Fatalf("id should not be exposed")
	}
	if _, ok := got["dataset_id"]; ok {
		t.Fatalf("dataset_id should not be exposed")
	}
}

func strPtr(value string) *string {
	return &value
}
