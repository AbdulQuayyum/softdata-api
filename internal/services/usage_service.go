package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

// RequestRecordInput is the safe internal input accepted by UsageService.
type RequestRecordInput struct {
	RequestID      string
	AccountID      *string
	APIKeyID       *string
	AnonymousID    *string
	DatasetGroup   *string
	Method         string
	Route          string
	StatusCode     int
	ResponseTimeMS *int64
	RequestBytes   *int64
	ResponseBytes  *int64
	RecordedAt     time.Time
}

type UsageService struct {
	usages           interfaces.UsageRepository
	apiKeys          interfaces.APIKeyRepository
	clock            Clock
	monthlyAllowance int64
}

func NewUsageService(usages interfaces.UsageRepository, apiKeys interfaces.APIKeyRepository, clock Clock, monthlyAllowance int64) (*UsageService, error) {
	switch {
	case usages == nil:
		return nil, fmt.Errorf("usage repository is required")
	case apiKeys == nil:
		return nil, fmt.Errorf("api key repository is required")
	case clock == nil:
		return nil, fmt.Errorf("clock is required")
	case monthlyAllowance <= 0:
		return nil, fmt.Errorf("monthly allowance must be positive")
	}

	return &UsageService{
		usages:           usages,
		apiKeys:          apiKeys,
		clock:            clock,
		monthlyAllowance: monthlyAllowance,
	}, nil
}

func (s *UsageService) RecordRequest(ctx context.Context, input RequestRecordInput) (models.APIRequest, error) {
	datasetGroup, err := normalizeDatasetGroup(input.DatasetGroup)
	if err != nil {
		return models.APIRequest{}, err
	}
	if err := s.validateRequestRecordInput(input); err != nil {
		return models.APIRequest{}, err
	}

	recordedAt := input.RecordedAt
	if recordedAt.IsZero() {
		recordedAt = s.clock.Now()
	}
	recordedAt = recordedAt.UTC()

	request := models.APIRequest{
		RequestID:      strings.TrimSpace(input.RequestID),
		AccountID:      cloneStringPtr(input.AccountID),
		APIKeyID:       cloneStringPtr(input.APIKeyID),
		AnonymousID:    cloneStringPtr(input.AnonymousID),
		DatasetGroup:   cloneStringPtr(datasetGroup),
		Method:         strings.TrimSpace(input.Method),
		Path:           normalizeUsageRoute(input.Route),
		Route:          cloneStringPtr(stringPtr(normalizeUsageRoute(input.Route))),
		QueryParams:    nil,
		StatusCode:     input.StatusCode,
		IPAddress:      nil,
		UserAgent:      nil,
		ResponseTimeMS: cloneInt64Ptr(input.ResponseTimeMS),
		RequestBytes:   cloneInt64Ptr(input.RequestBytes),
		ResponseBytes:  cloneInt64Ptr(input.ResponseBytes),
		CreatedAt:      recordedAt,
	}

	recorded, err := s.usages.RecordRequest(ctx, request)
	if err != nil {
		return models.APIRequest{}, fmt.Errorf("record request: %w", err)
	}

	scopeID, scopeType, err := requestScope(input)
	if err != nil {
		return models.APIRequest{}, err
	}
	requestCount := int64(1)
	successfulCount, errorCount := countBuckets(input.StatusCode)
	datasetDownloadCount := int64(0)
	if successfulCount > 0 && isDatasetDownloadRoute(request.Path) {
		datasetDownloadCount = 1
	}
	responseBytes := int64(0)
	if input.ResponseBytes != nil {
		responseBytes = *input.ResponseBytes
	}

	usageDate := utcDay(recordedAt)
	switch scopeType {
	case models.UsageScopeAnonymous:
		if err := s.usages.UpsertAnonymousDaily(ctx, usageDate, scopeID, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes); err != nil {
			return models.APIRequest{}, fmt.Errorf("record anonymous usage: %w", err)
		}
	case models.UsageScopeAccount:
		if err := s.usages.UpsertAccountDaily(ctx, usageDate, scopeID, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes); err != nil {
			return models.APIRequest{}, fmt.Errorf("record account usage: %w", err)
		}
	case models.UsageScopeAPIKey:
		if err := s.usages.UpsertAPIKeyDaily(ctx, usageDate, scopeID, requestCount, successfulCount, errorCount, datasetDownloadCount, responseBytes); err != nil {
			return models.APIRequest{}, fmt.Errorf("record api key usage: %w", err)
		}
	default:
		return models.APIRequest{}, fmt.Errorf("record request: unsupported usage scope")
	}

	return recorded, nil
}

func (s *UsageService) GetUsageHistory(ctx context.Context, accountID string, start, end time.Time) ([]models.UsageDailyResponse, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return nil, fmt.Errorf("get usage history: account id is required")
	}

	startUTC, endUTC, err := normalizeUsageRange(start, end)
	if err != nil {
		return nil, err
	}

	rows, err := s.collectAccountUsage(ctx, accountID, startUTC, endUTC)
	if err != nil {
		return nil, err
	}

	return dailyUsageResponsesFromSummaries(rows), nil
}

func (s *UsageService) GetAPIKeyUsageHistory(ctx context.Context, accountID, apiKeyID string, start, end time.Time) ([]models.UsageDailyResponse, error) {
	accountID = strings.TrimSpace(accountID)
	apiKeyID = strings.TrimSpace(apiKeyID)
	if accountID == "" || apiKeyID == "" {
		return nil, ErrAPIKeyNotFound
	}

	key, err := s.apiKeys.GetByID(ctx, apiKeyID)
	if err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, fmt.Errorf("get api key usage history: %w", err)
	}
	if strings.TrimSpace(key.AccountID) != accountID {
		return nil, ErrAPIKeyNotFound
	}

	startUTC, endUTC, err := normalizeUsageRange(start, end)
	if err != nil {
		return nil, err
	}

	rows, err := s.usages.GetSummaryByAPIKeyID(ctx, apiKeyID, int32(math.MaxInt32), 0)
	if err != nil {
		return nil, fmt.Errorf("get api key usage history: %w", err)
	}

	return dailyUsageResponsesFromSummaries(filterUsageSummaries(rows, startUTC, endUTC)), nil
}

func (s *UsageService) GetUsageSummary(ctx context.Context, accountID string, apiKeyID *string, start, end time.Time) (models.UsageSummaryReportResponse, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return models.UsageSummaryReportResponse{}, fmt.Errorf("get usage summary: account id is required")
	}

	startUTC, endUTC, err := normalizeUsageRange(start, end)
	if err != nil {
		return models.UsageSummaryReportResponse{}, err
	}

	var history []models.UsageDailyResponse
	if apiKeyID == nil || strings.TrimSpace(*apiKeyID) == "" {
		history, err = s.GetUsageHistory(ctx, accountID, startUTC, endUTC)
	} else {
		history, err = s.GetAPIKeyUsageHistory(ctx, accountID, *apiKeyID, startUTC, endUTC)
	}
	if err != nil {
		return models.UsageSummaryReportResponse{}, err
	}

	remainingAllowance, err := s.GetRemainingAllowance(ctx, accountID, endUTC.Add(-time.Nanosecond))
	if err != nil {
		return models.UsageSummaryReportResponse{}, err
	}

	summary := models.UsageSummaryReportResponse{
		TotalRequests:      totalUsageRequests(history),
		SuccessfulRequests: totalUsageSuccessfulRequests(history),
		ErrorCount:         totalUsageErrors(history),
		CurrentAllowance:   s.monthlyAllowance,
		RemainingAllowance: remainingAllowance,
		PeriodStart:        startUTC,
		PeriodEnd:          endUTC,
	}
	return summary, nil
}

func (s *UsageService) GetEndpointUsage(ctx context.Context, route string, start, end time.Time) (int64, error) {
	route = normalizeUsageRoute(route)
	if route == "" {
		return 0, fmt.Errorf("get endpoint usage: route is required")
	}

	startUTC, endUTC, err := normalizeUsageRange(start, end)
	if err != nil {
		return 0, err
	}

	count, err := s.usages.CountRequestsByRoute(ctx, route, startUTC, endUTC)
	if err != nil {
		return 0, fmt.Errorf("get endpoint usage: %w", err)
	}
	return count, nil
}

func (s *UsageService) GetDatasetGroupUsage(ctx context.Context, accountID string, start, end time.Time) ([]models.DatasetGroupUsageResponse, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return nil, fmt.Errorf("get dataset group usage: account id is required")
	}

	startUTC, endUTC, err := normalizeUsageRange(start, end)
	if err != nil {
		return nil, err
	}

	rows, err := s.collectAccountDatasetGroupUsage(ctx, accountID, startUTC, endUTC)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (s *UsageService) GetAPIKeyDatasetGroupUsage(ctx context.Context, accountID, apiKeyID string, start, end time.Time) ([]models.DatasetGroupUsageResponse, error) {
	accountID = strings.TrimSpace(accountID)
	apiKeyID = strings.TrimSpace(apiKeyID)
	if accountID == "" || apiKeyID == "" {
		return nil, ErrAPIKeyNotFound
	}

	key, err := s.apiKeys.GetByID(ctx, apiKeyID)
	if err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, fmt.Errorf("get api key dataset group usage: %w", err)
	}
	if strings.TrimSpace(key.AccountID) != accountID {
		return nil, ErrAPIKeyNotFound
	}

	startUTC, endUTC, err := normalizeUsageRange(start, end)
	if err != nil {
		return nil, err
	}

	rows, err := s.usages.GetDatasetGroupUsageByAPIKeyID(ctx, apiKeyID, startUTC, endUTC)
	if err != nil {
		return nil, fmt.Errorf("get api key dataset group usage: %w", err)
	}
	return rows, nil
}

func (s *UsageService) GetErrorCounts(ctx context.Context, accountID string, start, end time.Time) (int64, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return 0, fmt.Errorf("get error counts: account id is required")
	}

	startUTC, endUTC, err := normalizeUsageRange(start, end)
	if err != nil {
		return 0, err
	}

	rows, err := s.collectAccountUsage(ctx, accountID, startUTC, endUTC)
	if err != nil {
		return 0, err
	}

	var total int64
	for _, row := range rows {
		total += row.ErrorCount
	}
	return total, nil
}

func (s *UsageService) GetRemainingAllowance(ctx context.Context, accountID string, asOf time.Time) (int64, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return 0, fmt.Errorf("get remaining allowance: account id is required")
	}

	start, end := monthRangeUTC(asOf)
	rows, err := s.collectAccountUsage(ctx, accountID, start, end)
	if err != nil {
		return 0, err
	}

	var used int64
	for _, row := range rows {
		used += row.RequestCount
	}
	remaining := s.monthlyAllowance - used
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}

func (s *UsageService) collectAccountUsage(ctx context.Context, accountID string, start, end time.Time) ([]models.UsageSummaryResponse, error) {
	accountRows, err := s.usages.GetSummaryByAccountID(ctx, accountID, int32(math.MaxInt32), 0)
	if err != nil {
		return nil, fmt.Errorf("get usage history: %w", err)
	}

	ownedKeys, err := s.apiKeys.ListByAccountID(ctx, accountID, int32(math.MaxInt32), 0)
	if err != nil {
		return nil, fmt.Errorf("list api keys for usage: %w", err)
	}

	aggregated := make(map[string]*summaryAccumulator)
	addRows := func(rows []models.UsageSummaryResponse) {
		for _, row := range rows {
			date, ok := parseUsageDate(row.UsageDate)
			if !ok || !withinHalfOpenRange(date, start, end) {
				continue
			}
			acc := aggregated[row.UsageDate]
			if acc == nil {
				acc = &summaryAccumulator{}
				aggregated[row.UsageDate] = acc
			}
			acc.add(row)
		}
	}

	addRows(accountRows)
	for _, key := range ownedKeys {
		keyRows, err := s.usages.GetSummaryByAPIKeyID(ctx, key.ID, int32(math.MaxInt32), 0)
		if err != nil {
			return nil, fmt.Errorf("get api key usage history: %w", err)
		}
		addRows(keyRows)
	}

	return summariesFromAccumulator(aggregated), nil
}

func (s *UsageService) collectAccountDatasetGroupUsage(ctx context.Context, accountID string, start, end time.Time) ([]models.DatasetGroupUsageResponse, error) {
	accountRows, err := s.usages.GetDatasetGroupUsageByAccountID(ctx, accountID, start, end)
	if err != nil {
		return nil, fmt.Errorf("get dataset group usage: %w", err)
	}

	ownedKeys, err := s.apiKeys.ListByAccountID(ctx, accountID, int32(math.MaxInt32), 0)
	if err != nil {
		return nil, fmt.Errorf("list api keys for dataset group usage: %w", err)
	}

	aggregated := make(map[string]int64)
	addRows := func(rows []models.DatasetGroupUsageResponse) {
		for _, row := range rows {
			group := strings.TrimSpace(row.DatasetGroup)
			if group == "" {
				continue
			}
			aggregated[group] += row.RequestCount
		}
	}

	addRows(accountRows)
	for _, key := range ownedKeys {
		keyRows, err := s.usages.GetDatasetGroupUsageByAPIKeyID(ctx, key.ID, start, end)
		if err != nil {
			return nil, fmt.Errorf("get api key dataset group usage: %w", err)
		}
		addRows(keyRows)
	}

	return datasetGroupUsageFromCounts(aggregated), nil
}

func normalizeUsageRange(start, end time.Time) (time.Time, time.Time, error) {
	startUTC := start.UTC()
	endUTC := end.UTC()
	if startUTC.IsZero() || endUTC.IsZero() || !endUTC.After(startUTC) {
		return time.Time{}, time.Time{}, ErrInvalidUsagePeriod
	}
	return startUTC, endUTC, nil
}

func monthRangeUTC(value time.Time) (time.Time, time.Time) {
	t := value.UTC()
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 1, 0)
}

func utcDay(value time.Time) time.Time {
	t := value.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func withinHalfOpenRange(value, start, end time.Time) bool {
	return !value.Before(start) && value.Before(end)
}

func parseUsageDate(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	parsed, err := time.ParseInLocation("2006-01-02", value, time.UTC)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func filterUsageSummaries(rows []models.UsageSummaryResponse, start, end time.Time) []models.UsageSummaryResponse {
	aggregated := make(map[string]*summaryAccumulator)
	for _, row := range rows {
		date, ok := parseUsageDate(row.UsageDate)
		if !ok || !withinHalfOpenRange(date, start, end) {
			continue
		}
		acc := aggregated[row.UsageDate]
		if acc == nil {
			acc = &summaryAccumulator{}
			aggregated[row.UsageDate] = acc
		}
		acc.add(row)
	}
	return summariesFromAccumulator(aggregated)
}

func summariesFromAccumulator(aggregated map[string]*summaryAccumulator) []models.UsageSummaryResponse {
	items := make([]models.UsageSummaryResponse, 0, len(aggregated))
	for date, acc := range aggregated {
		items = append(items, models.UsageSummaryResponse{
			UsageDate:            date,
			ScopeType:            models.UsageScopeAccount,
			RequestCount:         acc.requestCount,
			SuccessfulCount:      acc.successfulCount,
			ErrorCount:           acc.errorCount,
			DatasetDownloadCount: acc.datasetDownloadCount,
			ResponseBytes:        acc.responseBytes,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].UsageDate > items[j].UsageDate
	})
	return items
}

func dailyUsageResponsesFromSummaries(rows []models.UsageSummaryResponse) []models.UsageDailyResponse {
	items := make([]models.UsageDailyResponse, 0, len(rows))
	for _, row := range rows {
		date, ok := parseUsageDate(row.UsageDate)
		if !ok {
			continue
		}
		items = append(items, models.UsageDailyResponse{
			Date:               date,
			TotalRequests:      row.RequestCount,
			SuccessfulRequests: row.SuccessfulCount,
			ErrorCount:         row.ErrorCount,
		})
	}
	return items
}

func totalUsageRequests(rows []models.UsageDailyResponse) int64 {
	var total int64
	for _, row := range rows {
		total += row.TotalRequests
	}
	return total
}

func totalUsageSuccessfulRequests(rows []models.UsageDailyResponse) int64 {
	var total int64
	for _, row := range rows {
		total += row.SuccessfulRequests
	}
	return total
}

func totalUsageErrors(rows []models.UsageDailyResponse) int64 {
	var total int64
	for _, row := range rows {
		total += row.ErrorCount
	}
	return total
}

func requestScope(input RequestRecordInput) (string, models.UsageScopeType, error) {
	accountID := normalizeOptionalString(input.AccountID)
	apiKeyID := normalizeOptionalString(input.APIKeyID)
	anonymousID := normalizeOptionalString(input.AnonymousID)

	switch {
	case accountID != "" && apiKeyID == "" && anonymousID == "":
		return accountID, models.UsageScopeAccount, nil
	case accountID == "" && apiKeyID != "" && anonymousID == "":
		return apiKeyID, models.UsageScopeAPIKey, nil
	case accountID == "" && apiKeyID == "" && anonymousID != "":
		return anonymousID, models.UsageScopeAnonymous, nil
	default:
		return "", "", fmt.Errorf("record request: exactly one usage identity must be provided")
	}
}

func (s *UsageService) validateRequestRecordInput(input RequestRecordInput) error {
	if strings.TrimSpace(input.RequestID) == "" {
		return fmt.Errorf("record request: request id is required")
	}
	if strings.TrimSpace(input.Method) == "" {
		return fmt.Errorf("record request: method is required")
	}
	route := normalizeUsageRoute(input.Route)
	if route == "" {
		return fmt.Errorf("record request: route is required")
	}
	if strings.Contains(route, "?") || strings.Contains(route, "://") {
		return fmt.Errorf("record request: route must be normalized")
	}
	if input.StatusCode < 100 || input.StatusCode > 599 {
		return fmt.Errorf("record request: status code is invalid")
	}
	if input.ResponseTimeMS != nil && *input.ResponseTimeMS < 0 {
		return fmt.Errorf("record request: response time must be non-negative")
	}
	if input.RequestBytes != nil && *input.RequestBytes < 0 {
		return fmt.Errorf("record request: request bytes must be non-negative")
	}
	if input.ResponseBytes != nil && *input.ResponseBytes < 0 {
		return fmt.Errorf("record request: response bytes must be non-negative")
	}
	_, _, err := requestScope(input)
	return err
}

func normalizeUsageRoute(route string) string {
	return strings.TrimSpace(route)
}

func isDatasetDownloadRoute(route string) bool {
	return strings.HasSuffix(route, "/download")
}

func countBuckets(statusCode int) (successfulCount, errorCount int64) {
	if statusCode >= 400 {
		return 0, 1
	}
	return 1, 0
}

func normalizeOptionalString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func normalizeDatasetGroup(value *string) (*string, error) {
	normalized := normalizeOptionalString(value)
	if normalized == "" {
		return nil, nil
	}

	switch strings.ToLower(normalized) {
	case "geography", "finance", "education", "healthcare", "emergency", "infrastructure", "statistics":
		v := strings.ToLower(normalized)
		return &v, nil
	default:
		return nil, ErrInvalidDatasetGroup
	}
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := strings.TrimSpace(*value)
	return &cloned
}

func cloneInt64Ptr(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

type summaryAccumulator struct {
	requestCount         int64
	successfulCount      int64
	errorCount           int64
	datasetDownloadCount int64
	responseBytes        int64
}

func (a *summaryAccumulator) add(row models.UsageSummaryResponse) {
	a.requestCount += row.RequestCount
	a.successfulCount += row.SuccessfulCount
	a.errorCount += row.ErrorCount
	a.datasetDownloadCount += row.DatasetDownloadCount
	a.responseBytes += row.ResponseBytes
}

func datasetGroupUsageFromCounts(counts map[string]int64) []models.DatasetGroupUsageResponse {
	items := make([]models.DatasetGroupUsageResponse, 0, len(counts))
	for group, count := range counts {
		items = append(items, models.DatasetGroupUsageResponse{
			DatasetGroup: group,
			RequestCount: count,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].RequestCount == items[j].RequestCount {
			return items[i].DatasetGroup < items[j].DatasetGroup
		}
		return items[i].RequestCount > items[j].RequestCount
	})
	return items
}
