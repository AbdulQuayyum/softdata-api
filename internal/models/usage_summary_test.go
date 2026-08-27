package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUsageSummaryReportResponseJSON(t *testing.T) {
	value := UsageSummaryReportResponse{
		TotalRequests:      12,
		SuccessfulRequests: 10,
		ErrorCount:         2,
		CurrentAllowance:   50000,
		RemainingAllowance: 49988,
		PeriodStart:        time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:          time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
	}

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	for _, key := range []string{"total_requests", "successful_requests", "error_count", "current_allowance", "remaining_allowance", "period_start", "period_end"} {
		if _, ok := got[key]; !ok {
			t.Fatalf("expected key %q to be present in JSON: %s", key, string(raw))
		}
	}
	for _, key := range []string{"id", "account_id", "api_key_id", "request_count"} {
		if _, ok := got[key]; ok {
			t.Fatalf("unexpected key %q in JSON: %s", key, string(raw))
		}
	}
}

func TestUsageDailyResponseJSON(t *testing.T) {
	value := UsageDailyResponse{
		Date:               time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC),
		TotalRequests:      5,
		SuccessfulRequests: 4,
		ErrorCount:         1,
	}

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	for _, key := range []string{"date", "total_requests", "successful_requests", "error_count"} {
		if _, ok := got[key]; !ok {
			t.Fatalf("expected key %q to be present in JSON: %s", key, string(raw))
		}
	}
	for _, key := range []string{"scope_type", "dataset_download_count", "response_bytes"} {
		if _, ok := got[key]; ok {
			t.Fatalf("unexpected key %q in JSON: %s", key, string(raw))
		}
	}
}
