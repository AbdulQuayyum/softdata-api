package interfaces

import (
	"context"
	"time"
)

type RateLimitSubjectKind string

const (
	RateLimitSubjectAnonymous RateLimitSubjectKind = "anonymous"
	RateLimitSubjectAPIKey    RateLimitSubjectKind = "api_key"
	RateLimitSubjectDownload  RateLimitSubjectKind = "download"
)

type RateLimitRequest struct {
	SubjectKind RateLimitSubjectKind
	Subject     string
	Limit       int64
	Window      time.Duration
}

type RateLimitResult struct {
	Allowed   bool
	Limit     int64
	Remaining int64
	ResetAt   time.Time
}

type RateLimitRepository interface {
	Allow(ctx context.Context, request RateLimitRequest) (RateLimitResult, error)
}
