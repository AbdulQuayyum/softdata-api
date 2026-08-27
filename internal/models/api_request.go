package models

import "time"

// APIRequest records a request log entry without unsafe request payloads.
type APIRequest struct {
	ID             int64
	RequestID      string
	AccountID      *string
	APIKeyID       *string
	AnonymousID    *string
	DatasetGroup   *string
	Method         string
	Path           string
	Route          *string
	QueryParams    map[string]any
	StatusCode     int
	IPAddress      *string
	UserAgent      *string
	ResponseTimeMS *int64
	RequestBytes   *int64
	ResponseBytes  *int64
	CreatedAt      time.Time
}
