package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

type usageRepoStub struct {
	recordFn          func(context.Context, models.APIRequest) (models.APIRequest, error)
	upsertAnonFn      func(context.Context, time.Time, string, int64, int64, int64, int64, int64) error
	upsertAccountFn   func(context.Context, time.Time, string, int64, int64, int64, int64, int64) error
	upsertAPIKeyFn    func(context.Context, time.Time, string, int64, int64, int64, int64, int64) error
	getSummaryAcctFn  func(context.Context, string, int32, int32) ([]models.UsageSummaryResponse, error)
	getSummaryKeyFn   func(context.Context, string, int32, int32) ([]models.UsageSummaryResponse, error)
	getGroupAcctFn    func(context.Context, string, time.Time, time.Time) ([]models.DatasetGroupUsageResponse, error)
	getGroupKeyFn     func(context.Context, string, time.Time, time.Time) ([]models.DatasetGroupUsageResponse, error)
	countRouteFn      func(context.Context, string, time.Time, time.Time) (int64, error)
	lastRecorded      models.APIRequest
	lastUsageDate     time.Time
	lastUsageID       string
	lastRoute         string
	lastFrom          time.Time
	lastTo            time.Time
	recordCalls       int
	accountUpserts    int
	apiKeyUpserts     int
	anonymousUpserts  int
	accountSummaryErr error
}

func (s *usageRepoStub) RecordRequest(ctx context.Context, request models.APIRequest) (models.APIRequest, error) {
	s.recordCalls++
	s.lastRecorded = request
	if s.recordFn != nil {
		return s.recordFn(ctx, request)
	}
	return request, nil
}

func (s *usageRepoStub) UpsertAnonymousDaily(ctx context.Context, usageDate time.Time, anonymousID string, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes int64) error {
	s.anonymousUpserts++
	s.lastUsageDate = usageDate
	s.lastUsageID = anonymousID
	if s.upsertAnonFn != nil {
		return s.upsertAnonFn(ctx, usageDate, anonymousID, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes)
	}
	return nil
}

func (s *usageRepoStub) UpsertAccountDaily(ctx context.Context, usageDate time.Time, accountID string, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes int64) error {
	s.accountUpserts++
	s.lastUsageDate = usageDate
	s.lastUsageID = accountID
	if s.upsertAccountFn != nil {
		return s.upsertAccountFn(ctx, usageDate, accountID, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes)
	}
	return nil
}

func (s *usageRepoStub) UpsertAPIKeyDaily(ctx context.Context, usageDate time.Time, apiKeyID string, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes int64) error {
	s.apiKeyUpserts++
	s.lastUsageDate = usageDate
	s.lastUsageID = apiKeyID
	if s.upsertAPIKeyFn != nil {
		return s.upsertAPIKeyFn(ctx, usageDate, apiKeyID, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes)
	}
	return nil
}

func (s *usageRepoStub) GetDailyByDate(context.Context, time.Time) ([]models.UsageSummary, error) {
	return nil, nil
}

func (s *usageRepoStub) GetSummaryByAccountID(ctx context.Context, accountID string, limit, offset int32) ([]models.UsageSummaryResponse, error) {
	if s.getSummaryAcctFn != nil {
		return s.getSummaryAcctFn(ctx, accountID, limit, offset)
	}
	return nil, s.accountSummaryErr
}

func (s *usageRepoStub) GetSummaryByAPIKeyID(ctx context.Context, apiKeyID string, limit, offset int32) ([]models.UsageSummaryResponse, error) {
	if s.getSummaryKeyFn != nil {
		return s.getSummaryKeyFn(ctx, apiKeyID, limit, offset)
	}
	return nil, nil
}

func (s *usageRepoStub) GetSummaryByAnonymousID(context.Context, string, int32, int32) ([]models.UsageSummaryResponse, error) {
	return nil, nil
}

func (s *usageRepoStub) GetDatasetGroupUsageByAccountID(ctx context.Context, accountID string, createdFrom, createdTo time.Time) ([]models.DatasetGroupUsageResponse, error) {
	if s.getGroupAcctFn != nil {
		return s.getGroupAcctFn(ctx, accountID, createdFrom, createdTo)
	}
	return nil, nil
}

func (s *usageRepoStub) GetDatasetGroupUsageByAPIKeyID(ctx context.Context, apiKeyID string, createdFrom, createdTo time.Time) ([]models.DatasetGroupUsageResponse, error) {
	if s.getGroupKeyFn != nil {
		return s.getGroupKeyFn(ctx, apiKeyID, createdFrom, createdTo)
	}
	return nil, nil
}

func (s *usageRepoStub) CountRequestsByRoute(ctx context.Context, route string, createdFrom, createdTo time.Time) (int64, error) {
	s.lastRoute = route
	s.lastFrom = createdFrom
	s.lastTo = createdTo
	if s.countRouteFn != nil {
		return s.countRouteFn(ctx, route, createdFrom, createdTo)
	}
	return 0, nil
}

func (s *usageRepoStub) DeleteExpired(context.Context, time.Time) error {
	return nil
}

func TestUsageServiceRecordRequest(t *testing.T) {
	repo := &usageRepoStub{
		recordFn: func(_ context.Context, request models.APIRequest) (models.APIRequest, error) {
			if request.IPAddress != nil || request.UserAgent != nil || request.QueryParams != nil {
				t.Fatalf("unsafe values were propagated into persistence request: %#v", request)
			}
			if request.DatasetGroup == nil || *request.DatasetGroup != "geography" {
				t.Fatalf("dataset group was not normalized or preserved: %#v", request.DatasetGroup)
			}
			if request.Path != "/v1/datasets/{dataset_id}/download" || request.Route == nil || *request.Route != "/v1/datasets/{dataset_id}/download" {
				t.Fatalf("route was not normalized: %#v", request)
			}
			if request.CreatedAt.Location() != time.UTC {
				t.Fatalf("recorded time was not normalized to UTC: %v", request.CreatedAt)
			}
			return request, nil
		},
		upsertAccountFn: func(_ context.Context, usageDate time.Time, accountID string, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes int64) error {
			if accountID != "acct-1" {
				t.Fatalf("unexpected account id: %q", accountID)
			}
			if !usageDate.Equal(time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC)) {
				t.Fatalf("unexpected usage date: %v", usageDate)
			}
			if requestCount != 1 || successfulCount != 1 || errorCount != 0 || datasetDownloadCount != 1 || responseBytes != 128 {
				t.Fatalf("unexpected aggregate counts: %d %d %d %d %d", requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes)
			}
			return nil
		},
	}
	svc, err := NewUsageService(repo, &apiKeyRepoStub{}, clockStub{now: time.Date(2026, 8, 26, 15, 4, 5, 0, time.FixedZone("WAT", 3600))}, 50000)
	if err != nil {
		t.Fatalf("NewUsageService() error = %v", err)
	}

	responseBytes := int64(128)
	recorded, err := svc.RecordRequest(context.Background(), RequestRecordInput{
		RequestID:     "req-1",
		AccountID:     ptrString("acct-1"),
		DatasetGroup:  ptrString(" Geography "),
		Method:        "GET",
		Route:         " /v1/datasets/{dataset_id}/download ",
		StatusCode:    200,
		ResponseBytes: &responseBytes,
		RecordedAt:    time.Date(2026, 8, 26, 15, 4, 5, 0, time.FixedZone("WAT", 3600)),
	})
	if err != nil {
		t.Fatalf("RecordRequest() error = %v", err)
	}
	if recorded.RequestID != "req-1" || recorded.Path != "/v1/datasets/{dataset_id}/download" {
		t.Fatalf("unexpected returned record: %#v", recorded)
	}
	if repo.accountUpserts != 1 || repo.apiKeyUpserts != 0 || repo.anonymousUpserts != 0 {
		t.Fatalf("unexpected upsert counts: account=%d apiKey=%d anonymous=%d", repo.accountUpserts, repo.apiKeyUpserts, repo.anonymousUpserts)
	}
}

func TestUsageServiceRecordRequestRejectsRawRoute(t *testing.T) {
	svc, err := NewUsageService(&usageRepoStub{}, &apiKeyRepoStub{}, clockStub{now: time.Now().UTC()}, 50000)
	if err != nil {
		t.Fatalf("NewUsageService() error = %v", err)
	}

	_, err = svc.RecordRequest(context.Background(), RequestRecordInput{
		RequestID:  "req-1",
		AccountID:  ptrString("acct-1"),
		Method:     "GET",
		Route:      "/v1/datasets/{dataset_id}?raw=1",
		StatusCode: 200,
		RecordedAt: time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("RecordRequest() error = nil, want error")
	}
}

func TestUsageServiceRecordRequestRejectsInvalidDatasetGroup(t *testing.T) {
	svc, err := NewUsageService(&usageRepoStub{}, &apiKeyRepoStub{}, clockStub{now: time.Now().UTC()}, 50000)
	if err != nil {
		t.Fatalf("NewUsageService() error = %v", err)
	}

	_, err = svc.RecordRequest(context.Background(), RequestRecordInput{
		RequestID:    "req-1",
		AccountID:    ptrString("acct-1"),
		DatasetGroup: ptrString("unknown"),
		Method:       "GET",
		Route:        "/v1/datasets/{dataset_id}",
		StatusCode:   200,
		RecordedAt:   time.Now().UTC(),
	})
	if !errors.Is(err, ErrInvalidDatasetGroup) {
		t.Fatalf("RecordRequest() error = %v, want ErrInvalidDatasetGroup", err)
	}
}

func TestUsageServiceHistoryAndAllowance(t *testing.T) {
	repo := &usageRepoStub{
		getSummaryAcctFn: func(context.Context, string, int32, int32) ([]models.UsageSummaryResponse, error) {
			return []models.UsageSummaryResponse{
				{UsageDate: "2026-08-26", RequestCount: 2, SuccessfulCount: 1, ErrorCount: 1, DatasetDownloadCount: 0, ResponseBytes: 10},
				{UsageDate: "2026-08-25", RequestCount: 1, SuccessfulCount: 1, ErrorCount: 0, DatasetDownloadCount: 1, ResponseBytes: 5},
				{UsageDate: "2026-07-31", RequestCount: 9, SuccessfulCount: 8, ErrorCount: 1, DatasetDownloadCount: 1, ResponseBytes: 99},
			}, nil
		},
		getSummaryKeyFn: func(context.Context, string, int32, int32) ([]models.UsageSummaryResponse, error) {
			return []models.UsageSummaryResponse{
				{UsageDate: "2026-08-26", RequestCount: 3, SuccessfulCount: 2, ErrorCount: 1, DatasetDownloadCount: 1, ResponseBytes: 7},
				{UsageDate: "2026-08-24", RequestCount: 4, SuccessfulCount: 4, ErrorCount: 0, DatasetDownloadCount: 0, ResponseBytes: 9},
			}, nil
		},
		countRouteFn: func(context.Context, string, time.Time, time.Time) (int64, error) {
			return 12, nil
		},
	}
	apiKeyRepo := &apiKeyRepoStub{
		listFn: func(context.Context, string, int32, int32) ([]models.APIKey, error) {
			return []models.APIKey{
				{ID: "key-1", AccountID: "acct-1"},
			}, nil
		},
		getByIDFn: func(context.Context, string) (models.APIKey, error) {
			return models.APIKey{ID: "key-1", AccountID: "acct-1"}, nil
		},
	}
	svc, err := NewUsageService(repo, apiKeyRepo, clockStub{now: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)}, 50000)
	if err != nil {
		t.Fatalf("NewUsageService() error = %v", err)
	}

	history, err := svc.GetUsageHistory(context.Background(), "acct-1", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetUsageHistory() error = %v", err)
	}
	if len(history) != 3 {
		t.Fatalf("expected 3 aggregated days, got %d", len(history))
	}
	if !history[0].Date.Equal(time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC)) || history[0].TotalRequests != 5 || history[0].ErrorCount != 2 {
		t.Fatalf("unexpected first history row: %#v", history[0])
	}
	if !history[1].Date.Equal(time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)) || history[1].TotalRequests != 1 || history[1].SuccessfulRequests != 1 {
		t.Fatalf("unexpected second history row: %#v", history[1])
	}

	summary, err := svc.GetUsageSummary(context.Background(), "acct-1", nil, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetUsageSummary() error = %v", err)
	}
	if summary.TotalRequests != 10 || summary.SuccessfulRequests != 8 || summary.ErrorCount != 2 {
		t.Fatalf("unexpected usage summary totals: %#v", summary)
	}
	if summary.CurrentAllowance != 50000 || summary.RemainingAllowance != 49990 {
		t.Fatalf("unexpected usage allowance values: %#v", summary)
	}
	if !summary.PeriodStart.Equal(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)) || !summary.PeriodEnd.Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected usage period: %#v", summary)
	}

	keyHistory, err := svc.GetAPIKeyUsageHistory(context.Background(), "acct-1", "key-1", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetAPIKeyUsageHistory() error = %v", err)
	}
	if len(keyHistory) != 2 || !keyHistory[0].Date.Equal(time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC)) || keyHistory[0].TotalRequests != 3 {
		t.Fatalf("unexpected api key history: %#v", keyHistory)
	}

	keyID := "key-1"
	keySummary, err := svc.GetUsageSummary(context.Background(), "acct-1", &keyID, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetUsageSummary() with api key error = %v", err)
	}
	if keySummary.TotalRequests != 7 || keySummary.SuccessfulRequests != 6 || keySummary.ErrorCount != 1 {
		t.Fatalf("unexpected api key usage summary totals: %#v", keySummary)
	}

	errorsCount, err := svc.GetErrorCounts(context.Background(), "acct-1", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetErrorCounts() error = %v", err)
	}
	if errorsCount != 2 {
		t.Fatalf("unexpected error count: %d", errorsCount)
	}

	remaining, err := svc.GetRemainingAllowance(context.Background(), "acct-1", time.Date(2026, 8, 26, 10, 0, 0, 0, time.FixedZone("PDT", -7*3600)))
	if err != nil {
		t.Fatalf("GetRemainingAllowance() error = %v", err)
	}
	if remaining != 49990 {
		t.Fatalf("unexpected remaining allowance: %d", remaining)
	}

	usage, err := svc.GetEndpointUsage(context.Background(), "/v1/datasets/{dataset_id}/download", time.Date(2026, 8, 1, 0, 0, 0, 0, time.FixedZone("PDT", -7*3600)), time.Date(2026, 9, 1, 0, 0, 0, 0, time.FixedZone("PDT", -7*3600)))
	if err != nil {
		t.Fatalf("GetEndpointUsage() error = %v", err)
	}
	if usage != 12 {
		t.Fatalf("unexpected endpoint usage: %d", usage)
	}
	if !repo.lastFrom.Equal(time.Date(2026, 8, 1, 7, 0, 0, 0, time.UTC)) || !repo.lastTo.Equal(time.Date(2026, 9, 1, 7, 0, 0, 0, time.UTC)) {
		t.Fatalf("endpoint range was not normalized to UTC: %v %v", repo.lastFrom, repo.lastTo)
	}
}

func TestUsageServiceAPIKeyOwnership(t *testing.T) {
	repo := &usageRepoStub{
		getSummaryKeyFn: func(context.Context, string, int32, int32) ([]models.UsageSummaryResponse, error) {
			return []models.UsageSummaryResponse{{UsageDate: "2026-08-26", RequestCount: 1}}, nil
		},
	}
	apiKeyRepo := &apiKeyRepoStub{
		getByIDFn: func(context.Context, string) (models.APIKey, error) {
			return models.APIKey{ID: "key-1", AccountID: "acct-1"}, nil
		},
	}
	svc, err := NewUsageService(repo, apiKeyRepo, clockStub{now: time.Now().UTC()}, 50000)
	if err != nil {
		t.Fatalf("NewUsageService() error = %v", err)
	}

	if _, err := svc.GetAPIKeyUsageHistory(context.Background(), "acct-2", "key-1", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)); !errors.Is(err, ErrAPIKeyNotFound) {
		t.Fatalf("GetAPIKeyUsageHistory() error = %v, want ErrAPIKeyNotFound", err)
	}
}

func TestUsageServiceDatasetGroupUsageAggregatesAccountAndKeys(t *testing.T) {
	repo := &usageRepoStub{
		getGroupAcctFn: func(context.Context, string, time.Time, time.Time) ([]models.DatasetGroupUsageResponse, error) {
			return []models.DatasetGroupUsageResponse{
				{DatasetGroup: "geography", RequestCount: 2},
				{DatasetGroup: "education", RequestCount: 1},
			}, nil
		},
		getGroupKeyFn: func(ctx context.Context, apiKeyID string, start, end time.Time) ([]models.DatasetGroupUsageResponse, error) {
			if apiKeyID != "key-1" {
				t.Fatalf("unexpected api key id: %q", apiKeyID)
			}
			return []models.DatasetGroupUsageResponse{
				{DatasetGroup: "geography", RequestCount: 3},
				{DatasetGroup: "finance", RequestCount: 4},
			}, nil
		},
	}
	apiKeyRepo := &apiKeyRepoStub{
		listFn: func(context.Context, string, int32, int32) ([]models.APIKey, error) {
			return []models.APIKey{{ID: "key-1", AccountID: "acct-1"}}, nil
		},
		getByIDFn: func(context.Context, string) (models.APIKey, error) {
			return models.APIKey{ID: "key-1", AccountID: "acct-1"}, nil
		},
	}
	svc, err := NewUsageService(repo, apiKeyRepo, clockStub{now: time.Now().UTC()}, 50000)
	if err != nil {
		t.Fatalf("NewUsageService() error = %v", err)
	}

	rows, err := svc.GetDatasetGroupUsage(context.Background(), "acct-1", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetDatasetGroupUsage() error = %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("unexpected dataset group rows: %#v", rows)
	}
	if rows[0].DatasetGroup != "geography" || rows[0].RequestCount != 5 {
		t.Fatalf("unexpected first row: %#v", rows[0])
	}
	if rows[1].DatasetGroup != "finance" || rows[1].RequestCount != 4 {
		t.Fatalf("unexpected second row: %#v", rows[1])
	}
	if rows[2].DatasetGroup != "education" || rows[2].RequestCount != 1 {
		t.Fatalf("unexpected third row: %#v", rows[2])
	}
}

func TestUsageServiceAPIKeyDatasetGroupUsageRequiresOwnership(t *testing.T) {
	repo := &usageRepoStub{}
	apiKeyRepo := &apiKeyRepoStub{
		getByIDFn: func(context.Context, string) (models.APIKey, error) {
			return models.APIKey{ID: "key-1", AccountID: "acct-1"}, nil
		},
	}
	svc, err := NewUsageService(repo, apiKeyRepo, clockStub{now: time.Now().UTC()}, 50000)
	if err != nil {
		t.Fatalf("NewUsageService() error = %v", err)
	}

	if _, err := svc.GetAPIKeyDatasetGroupUsage(context.Background(), "acct-2", "key-1", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)); !errors.Is(err, ErrAPIKeyNotFound) {
		t.Fatalf("GetAPIKeyDatasetGroupUsage() error = %v, want ErrAPIKeyNotFound", err)
	}
}

func TestUsageServiceBoundaryHelpers(t *testing.T) {
	start, end, err := normalizeUsageRange(time.Date(2026, 8, 1, 15, 0, 0, 0, time.FixedZone("PDT", -7*3600)), time.Date(2026, 8, 2, 15, 0, 0, 0, time.FixedZone("PDT", -7*3600)))
	if err != nil {
		t.Fatalf("normalizeUsageRange() error = %v", err)
	}
	if !start.Equal(time.Date(2026, 8, 1, 22, 0, 0, 0, time.UTC)) || !end.Equal(time.Date(2026, 8, 2, 22, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected normalized range: %v %v", start, end)
	}

	monthStart, monthEnd := monthRangeUTC(time.Date(2026, 8, 26, 15, 0, 0, 0, time.FixedZone("WAT", 3600)))
	if !monthStart.Equal(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)) || !monthEnd.Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected month range: %v %v", monthStart, monthEnd)
	}
}

func ptrString(value string) *string {
	return &value
}
