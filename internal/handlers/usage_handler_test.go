package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/middlewares"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
)

type usageHandlerStub struct {
	getSummaryFn       func(context.Context, string, *string, time.Time, time.Time) (models.UsageSummaryReportResponse, error)
	getHistoryFn       func(context.Context, string, time.Time, time.Time) ([]models.UsageDailyResponse, error)
	getAPIKeyHistoryFn func(context.Context, string, string, time.Time, time.Time) ([]models.UsageDailyResponse, error)
	listEndpointFn     func(context.Context, string, time.Time, time.Time) ([]models.EndpointUsageResponse, error)
	listAPIKeyFn       func(context.Context, string, string, time.Time, time.Time) ([]models.EndpointUsageResponse, error)
	getGroupFn         func(context.Context, string, time.Time, time.Time) ([]models.DatasetGroupUsageResponse, error)
	getAPIKeyGroupFn   func(context.Context, string, string, time.Time, time.Time) ([]models.DatasetGroupUsageResponse, error)

	getSummaryCalls       int
	getHistoryCalls       int
	getAPIKeyHistoryCalls int
	listEndpointCalls     int
	listAPIKeyCalls       int
	getGroupCalls         int
	getAPIKeyGroupCalls   int

	lastAccountID string
	lastAPIKeyID  string
	lastStart     time.Time
	lastEnd       time.Time
}

func (s *usageHandlerStub) GetUsageSummary(ctx context.Context, accountID string, apiKeyID *string, start, end time.Time) (models.UsageSummaryReportResponse, error) {
	s.getSummaryCalls++
	s.lastAccountID = accountID
	s.lastStart = start
	s.lastEnd = end
	if apiKeyID != nil {
		s.lastAPIKeyID = *apiKeyID
	} else {
		s.lastAPIKeyID = ""
	}
	if s.getSummaryFn != nil {
		return s.getSummaryFn(ctx, accountID, apiKeyID, start, end)
	}
	return models.UsageSummaryReportResponse{}, nil
}

func (s *usageHandlerStub) GetUsageHistory(ctx context.Context, accountID string, start, end time.Time) ([]models.UsageDailyResponse, error) {
	s.getHistoryCalls++
	s.lastAccountID = accountID
	s.lastStart = start
	s.lastEnd = end
	if s.getHistoryFn != nil {
		return s.getHistoryFn(ctx, accountID, start, end)
	}
	return nil, nil
}

func (s *usageHandlerStub) GetAPIKeyUsageHistory(ctx context.Context, accountID, apiKeyID string, start, end time.Time) ([]models.UsageDailyResponse, error) {
	s.getAPIKeyHistoryCalls++
	s.lastAccountID = accountID
	s.lastAPIKeyID = apiKeyID
	s.lastStart = start
	s.lastEnd = end
	if s.getAPIKeyHistoryFn != nil {
		return s.getAPIKeyHistoryFn(ctx, accountID, apiKeyID, start, end)
	}
	return nil, nil
}

func (s *usageHandlerStub) ListEndpointUsage(ctx context.Context, accountID string, start, end time.Time) ([]models.EndpointUsageResponse, error) {
	s.listEndpointCalls++
	s.lastAccountID = accountID
	s.lastStart = start
	s.lastEnd = end
	if s.listEndpointFn != nil {
		return s.listEndpointFn(ctx, accountID, start, end)
	}
	return nil, nil
}

func (s *usageHandlerStub) ListAPIKeyEndpointUsage(ctx context.Context, accountID, apiKeyID string, start, end time.Time) ([]models.EndpointUsageResponse, error) {
	s.listAPIKeyCalls++
	s.lastAccountID = accountID
	s.lastAPIKeyID = apiKeyID
	s.lastStart = start
	s.lastEnd = end
	if s.listAPIKeyFn != nil {
		return s.listAPIKeyFn(ctx, accountID, apiKeyID, start, end)
	}
	return nil, nil
}

func (s *usageHandlerStub) GetDatasetGroupUsage(ctx context.Context, accountID string, start, end time.Time) ([]models.DatasetGroupUsageResponse, error) {
	s.getGroupCalls++
	s.lastAccountID = accountID
	s.lastStart = start
	s.lastEnd = end
	if s.getGroupFn != nil {
		return s.getGroupFn(ctx, accountID, start, end)
	}
	return nil, nil
}

func (s *usageHandlerStub) GetAPIKeyDatasetGroupUsage(ctx context.Context, accountID, apiKeyID string, start, end time.Time) ([]models.DatasetGroupUsageResponse, error) {
	s.getAPIKeyGroupCalls++
	s.lastAccountID = accountID
	s.lastAPIKeyID = apiKeyID
	s.lastStart = start
	s.lastEnd = end
	if s.getAPIKeyGroupFn != nil {
		return s.getAPIKeyGroupFn(ctx, accountID, apiKeyID, start, end)
	}
	return nil, nil
}

func TestNewUsageHandlerRejectsNilService(t *testing.T) {
	if _, err := NewUsageHandler(nil); err == nil {
		t.Fatal("NewUsageHandler(nil) error = nil, want error")
	}
}

func TestUsageHandlerUsageSummaryUsesBearerAccountAndApiKeyFilter(t *testing.T) {
	stub := &usageHandlerStub{
		getSummaryFn: func(ctx context.Context, accountID string, apiKeyID *string, start, end time.Time) (models.UsageSummaryReportResponse, error) {
			if accountID != "acc_123" {
				t.Fatalf("unexpected account id: %q", accountID)
			}
			if apiKeyID == nil || *apiKeyID != "key_123" {
				t.Fatalf("unexpected api key id: %#v", apiKeyID)
			}
			wantStart := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
			wantEnd := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
			if !start.Equal(wantStart) || !end.Equal(wantEnd) {
				t.Fatalf("unexpected range: %s - %s", start, end)
			}
			return models.UsageSummaryReportResponse{
				TotalRequests:      10,
				SuccessfulRequests: 9,
				ErrorCount:         1,
				CurrentAllowance:   50000,
				RemainingAllowance: 49990,
				PeriodStart:        wantStart,
				PeriodEnd:          wantEnd,
			}, nil
		},
	}
	h, err := NewUsageHandler(stub)
	if err != nil {
		t.Fatalf("NewUsageHandler() error = %v", err)
	}
	h.now = func() time.Time {
		return time.Date(2026, time.August, 27, 16, 30, 0, 0, time.UTC)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/account/usage?"+url.Values{"api_key_id": []string{" key_123 "}}.Encode(), nil)
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	rr := httptest.NewRecorder()

	h.UsageSummary(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type: %q", got)
	}
	if stub.getSummaryCalls != 1 {
		t.Fatalf("unexpected summary call count: %d", stub.getSummaryCalls)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body["success"] != true {
		t.Fatalf("unexpected success flag: %#v", body["success"])
	}
	data := body["data"].(map[string]any)
	if data["total_requests"] != float64(10) || data["remaining_allowance"] != float64(49990) {
		t.Fatalf("unexpected summary payload: %#v", data)
	}
}

func TestUsageHandlerUsageHistorySwitchesBetweenAccountAndAPIKeyHistory(t *testing.T) {
	stub := &usageHandlerStub{
		getHistoryFn: func(ctx context.Context, accountID string, start, end time.Time) ([]models.UsageDailyResponse, error) {
			if accountID != "acc_123" {
				t.Fatalf("unexpected account id: %q", accountID)
			}
			return nil, nil
		},
		getAPIKeyHistoryFn: func(ctx context.Context, accountID, apiKeyID string, start, end time.Time) ([]models.UsageDailyResponse, error) {
			if accountID != "acc_123" || apiKeyID != "key_123" {
				t.Fatalf("unexpected identity: %q %q", accountID, apiKeyID)
			}
			return []models.UsageDailyResponse{{
				Date:               time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
				TotalRequests:      2,
				SuccessfulRequests: 2,
				ErrorCount:         0,
			}}, nil
		},
	}
	h, err := NewUsageHandler(stub)
	if err != nil {
		t.Fatalf("NewUsageHandler() error = %v", err)
	}
	h.now = func() time.Time {
		return time.Date(2026, time.August, 27, 16, 30, 0, 0, time.UTC)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/account/usage/history?api_key_id=key_123", nil)
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	rr := httptest.NewRecorder()

	h.UsageHistory(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.getHistoryCalls != 0 {
		t.Fatalf("unexpected account history call count: %d", stub.getHistoryCalls)
	}
	if stub.getAPIKeyHistoryCalls != 1 {
		t.Fatalf("unexpected api key history call count: %d", stub.getAPIKeyHistoryCalls)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	data := body["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("unexpected history payload: %#v", data)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/account/usage/history", nil)
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	h.UsageHistory(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.getHistoryCalls != 1 {
		t.Fatalf("unexpected account history call count: %d", stub.getHistoryCalls)
	}
	if stub.getAPIKeyHistoryCalls != 1 {
		t.Fatalf("unexpected api key history call count: %d", stub.getAPIKeyHistoryCalls)
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	data = body["data"].([]any)
	if len(data) != 0 {
		t.Fatalf("expected empty slice, got %#v", data)
	}
}

func TestUsageHandlerEndpointAndDatasetGroupUsageReturnEmptySlices(t *testing.T) {
	stub := &usageHandlerStub{}
	h, err := NewUsageHandler(stub)
	if err != nil {
		t.Fatalf("NewUsageHandler() error = %v", err)
	}
	h.now = func() time.Time {
		return time.Date(2026, time.August, 27, 16, 30, 0, 0, time.UTC)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/account/usage/endpoints", nil)
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	rr := httptest.NewRecorder()
	h.EndpointUsage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.listEndpointCalls != 1 {
		t.Fatalf("unexpected endpoint call count: %d", stub.listEndpointCalls)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if got := body["data"].([]any); len(got) != 0 {
		t.Fatalf("expected empty endpoint list, got %#v", got)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/account/usage/dataset-groups?api_key_id=key_123", nil)
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	h.DatasetGroupUsage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.getAPIKeyGroupCalls != 1 {
		t.Fatalf("unexpected api key group call count: %d", stub.getAPIKeyGroupCalls)
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if got := body["data"].([]any); len(got) != 0 {
		t.Fatalf("expected empty dataset group list, got %#v", got)
	}
}

func TestUsageHandlerRejectsMissingIdentityInvalidQueryAndUnsupportedMethods(t *testing.T) {
	stub := &usageHandlerStub{}
	h, err := NewUsageHandler(stub)
	if err != nil {
		t.Fatalf("NewUsageHandler() error = %v", err)
	}
	h.now = func() time.Time {
		return time.Date(2026, time.August, 27, 16, 30, 0, 0, time.UTC)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/account/usage", nil)
	h.UsageSummary(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status for missing identity: %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/account/usage?start=2026-08-01", nil)
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	h.UsageSummary(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status for invalid query: %d", rr.Code)
	}
	var body response.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Error.Code != "VALIDATION_FAILED" {
		t.Fatalf("unexpected error code: %q", body.Error.Code)
	}

	methods := []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
		path string
	}{
		{name: "summary", fn: h.UsageSummary, path: "/v1/account/usage"},
		{name: "history", fn: h.UsageHistory, path: "/v1/account/usage/history"},
		{name: "endpoints", fn: h.EndpointUsage, path: "/v1/account/usage/endpoints"},
		{name: "groups", fn: h.DatasetGroupUsage, path: "/v1/account/usage/dataset-groups"},
	}

	for _, tc := range methods {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, tc.path, nil)
			req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
			tc.fn(rr, req)
			if rr.Code != http.StatusMethodNotAllowed {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			if got := rr.Header().Get("Allow"); got != http.MethodGet {
				t.Fatalf("unexpected allow header: %q", got)
			}
		})
	}
}

func TestUsageHandlerDefaultRangeUsesCurrentUTCMonth(t *testing.T) {
	stub := &usageHandlerStub{
		getGroupFn: func(ctx context.Context, accountID string, start, end time.Time) ([]models.DatasetGroupUsageResponse, error) {
			wantStart := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
			wantEnd := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
			if !start.Equal(wantStart) || !end.Equal(wantEnd) {
				t.Fatalf("unexpected default range: %s - %s", start, end)
			}
			return nil, nil
		},
	}
	h, err := NewUsageHandler(stub)
	if err != nil {
		t.Fatalf("NewUsageHandler() error = %v", err)
	}
	h.now = func() time.Time {
		return time.Date(2026, time.August, 27, 16, 30, 0, 0, time.UTC)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/account/usage/dataset-groups", nil)
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))
	rr := httptest.NewRecorder()

	h.DatasetGroupUsage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestUsageHandlerPreservesQueryValidationErrors(t *testing.T) {
	stub := &usageHandlerStub{}
	h, err := NewUsageHandler(stub)
	if err != nil {
		t.Fatalf("NewUsageHandler() error = %v", err)
	}
	h.now = func() time.Time {
		return time.Date(2026, time.August, 27, 16, 30, 0, 0, time.UTC)
	}

	values := url.Values{}
	values.Add("start", "2026-08-01")
	values.Add("start", "2026-08-02")
	values.Add("end", "2026-09-01")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/account/usage?"+values.Encode(), nil)
	req = req.WithContext(middlewares.WithAccountIdentity(req.Context(), middlewares.AccountIdentity{AccountID: "acc_123"}))

	h.UsageSummary(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.getSummaryCalls != 0 {
		t.Fatalf("service should not be called for invalid query")
	}
}
