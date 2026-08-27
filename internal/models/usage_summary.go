package models

import "time"

type UsageScopeType string

const (
	UsageScopeAnonymous UsageScopeType = "anonymous"
	UsageScopeAccount   UsageScopeType = "account"
	UsageScopeAPIKey    UsageScopeType = "api_key"
)

// UsageSummary matches the daily usage schema and query results.
type UsageSummary struct {
	ID                   int64
	UsageDate            time.Time
	ScopeType            UsageScopeType
	AccountID            *string
	APIKeyID             *string
	AnonymousID          *string
	RequestCount         int64
	SuccessfulCount      int64
	ErrorCount           int64
	DatasetDownloadCount int64
	ResponseBytes        int64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// UsageSummaryResponse is the public aggregate usage view.
type UsageSummaryResponse struct {
	UsageDate            string         `json:"usage_date"`
	ScopeType            UsageScopeType `json:"scope_type"`
	RequestCount         int64          `json:"request_count"`
	SuccessfulCount      int64          `json:"successful_count"`
	ErrorCount           int64          `json:"error_count"`
	DatasetDownloadCount int64          `json:"dataset_download_count"`
	ResponseBytes        int64          `json:"response_bytes"`
}

// UsageSummaryReportResponse is the safe public usage summary view.
type UsageSummaryReportResponse struct {
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	ErrorCount         int64     `json:"error_count"`
	CurrentAllowance   int64     `json:"current_allowance"`
	RemainingAllowance int64     `json:"remaining_allowance"`
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
}

// UsageDailyResponse is the safe public aggregated daily usage view.
type UsageDailyResponse struct {
	Date               time.Time `json:"date"`
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	ErrorCount         int64     `json:"error_count"`
}

// EndpointUsageResponse is the safe public endpoint breakdown view.
type EndpointUsageResponse struct {
	Endpoint     string `json:"endpoint"`
	RequestCount int64  `json:"request_count"`
}

// DatasetGroupUsageResponse is the safe public dataset-group breakdown view.
type DatasetGroupUsageResponse struct {
	DatasetGroup string `json:"dataset_group"`
	RequestCount int64  `json:"request_count"`
}
