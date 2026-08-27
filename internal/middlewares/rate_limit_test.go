package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type rateLimitRepositoryStub struct {
	mu      sync.Mutex
	request interfaces.RateLimitRequest
	result  interfaces.RateLimitResult
	err     error
	calls   int
}

func (s *rateLimitRepositoryStub) Allow(ctx context.Context, request interfaces.RateLimitRequest) (interfaces.RateLimitResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.request = request
	return s.result, s.err
}

type anonymousIdentifierStub struct {
	mu    sync.Mutex
	value string
	err   error
	calls int
	last  *http.Request
}

func (s *anonymousIdentifierStub) Identify(r *http.Request) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.last = r
	return s.value, s.err
}

func TestRateLimitConstructorValidation(t *testing.T) {
	repo := &rateLimitRepositoryStub{}
	anon := &anonymousIdentifierStub{value: "anon"}
	policy := RateLimitPolicy{AnonymousLimit: 60, APIKeyLimit: 300, DownloadLimit: 10, Window: time.Minute}

	if _, err := RateLimit(nil, anon, policy); err == nil {
		t.Fatal("RateLimit() error = nil, want error")
	}
	if _, err := RateLimit(repo, nil, policy); err == nil {
		t.Fatal("RateLimit() error = nil, want error")
	}
	if _, err := RateLimit(repo, anon, RateLimitPolicy{}); err == nil {
		t.Fatal("RateLimit() error = nil, want error")
	}
	if _, err := DownloadRateLimit(repo, anon, RateLimitPolicy{AnonymousLimit: 60, APIKeyLimit: 300, DownloadLimit: 10}); err == nil {
		t.Fatal("DownloadRateLimit() error = nil, want error")
	}
}

func TestRateLimitAnonymousRequestAllowsAndSetsHeaders(t *testing.T) {
	repo := &rateLimitRepositoryStub{
		result: interfaces.RateLimitResult{
			Allowed:   true,
			Limit:     60,
			Remaining: 59,
			ResetAt:   time.Date(2026, 8, 27, 12, 1, 0, 0, time.UTC),
		},
	}
	anon := &anonymousIdentifierStub{value: "anon-opaque"}
	mw, err := RateLimit(repo, anon, RateLimitPolicy{AnonymousLimit: 60, APIKeyLimit: 300, DownloadLimit: 10, Window: time.Minute})
	if err != nil {
		t.Fatalf("RateLimit() error = %v", err)
	}

	nextCalls := 0
	handler := RequestID(mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalls++
		if got, ok := RequestIDFromContext(r.Context()); !ok || got == "" {
			t.Fatal("request id missing from context")
		}
		if _, ok := APIKeyIdentityFromContext(r.Context()); ok {
			t.Fatal("unexpected api key identity")
		}
	})))

	req := httptest.NewRequest(http.MethodGet, "/v1/datasets", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set("X-Forwarded-For", "198.51.100.77")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if nextCalls != 1 {
		t.Fatalf("unexpected next handler count: %d", nextCalls)
	}
	if repo.calls != 1 {
		t.Fatalf("unexpected repo call count: %d", repo.calls)
	}
	if repo.request.SubjectKind != interfaces.RateLimitSubjectAnonymous {
		t.Fatalf("unexpected subject kind: %#v", repo.request.SubjectKind)
	}
	if repo.request.Subject != "anon-opaque" {
		t.Fatalf("unexpected subject: %q", repo.request.Subject)
	}
	if repo.request.Limit != 60 {
		t.Fatalf("unexpected limit: %d", repo.request.Limit)
	}
	if got := rr.Header().Get(rateLimitHeaderLimit); got != "60" {
		t.Fatalf("unexpected limit header: %q", got)
	}
	if got := rr.Header().Get(rateLimitHeaderRemaining); got != "59" {
		t.Fatalf("unexpected remaining header: %q", got)
	}
	if got := rr.Header().Get(rateLimitHeaderReset); got != strconv.FormatInt(repo.result.ResetAt.Unix(), 10) {
		t.Fatalf("unexpected reset header: %q", got)
	}
	if got := rr.Header().Get(rateLimitHeaderRetryAfter); got != "" {
		t.Fatalf("unexpected retry-after header: %q", got)
	}
}

func TestRateLimitApiKeyRequestUsesApiKeyTier(t *testing.T) {
	repo := &rateLimitRepositoryStub{
		result: interfaces.RateLimitResult{
			Allowed:   true,
			Limit:     300,
			Remaining: 42,
			ResetAt:   time.Date(2026, 8, 27, 12, 1, 0, 0, time.UTC),
		},
	}
	anon := &anonymousIdentifierStub{value: "anon-opaque"}
	mw, err := RateLimit(repo, anon, RateLimitPolicy{AnonymousLimit: 60, APIKeyLimit: 300, DownloadLimit: 10, Window: time.Minute})
	if err != nil {
		t.Fatalf("RateLimit() error = %v", err)
	}

	handler := RequestID(mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/states", nil)
	req = req.WithContext(WithAPIKeyIdentity(req.Context(), services.APIKeyIdentity{APIKeyID: "key_123", AccountID: "acc_123"}))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if repo.request.SubjectKind != interfaces.RateLimitSubjectAPIKey {
		t.Fatalf("unexpected subject kind: %#v", repo.request.SubjectKind)
	}
	if repo.request.Subject != "key_123" {
		t.Fatalf("unexpected subject: %q", repo.request.Subject)
	}
	if repo.request.Limit != 300 {
		t.Fatalf("unexpected limit: %d", repo.request.Limit)
	}
}

func TestDownloadRateLimitUsesDownloadNamespace(t *testing.T) {
	repo := &rateLimitRepositoryStub{
		result: interfaces.RateLimitResult{
			Allowed:   true,
			Limit:     10,
			Remaining: 9,
			ResetAt:   time.Date(2026, 8, 27, 12, 1, 0, 0, time.UTC),
		},
	}
	anon := &anonymousIdentifierStub{value: "anon-opaque"}
	mw, err := DownloadRateLimit(repo, anon, RateLimitPolicy{AnonymousLimit: 60, APIKeyLimit: 300, DownloadLimit: 10, Window: time.Minute})
	if err != nil {
		t.Fatalf("DownloadRateLimit() error = %v", err)
	}

	handler := RequestID(mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	req := httptest.NewRequest(http.MethodGet, "/v1/datasets/example/download", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if repo.request.SubjectKind != interfaces.RateLimitSubjectDownload {
		t.Fatalf("unexpected subject kind: %#v", repo.request.SubjectKind)
	}
	if repo.request.Subject != "anon-opaque" {
		t.Fatalf("unexpected subject: %q", repo.request.Subject)
	}
	if repo.request.Limit != 10 {
		t.Fatalf("unexpected limit: %d", repo.request.Limit)
	}
}

func TestRateLimitRejectsAtLimitAndSets429Body(t *testing.T) {
	repo := &rateLimitRepositoryStub{
		result: interfaces.RateLimitResult{
			Allowed:   false,
			Limit:     60,
			Remaining: 0,
			ResetAt:   time.Date(2026, 8, 27, 12, 1, 30, 0, time.UTC),
		},
	}
	anon := &anonymousIdentifierStub{value: "anon-opaque"}
	mw, err := RateLimit(repo, anon, RateLimitPolicy{AnonymousLimit: 60, APIKeyLimit: 300, DownloadLimit: 10, Window: time.Minute})
	if err != nil {
		t.Fatalf("RateLimit() error = %v", err)
	}

	nextCalls := 0
	handler := RequestID(mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalls++
	})))

	req := httptest.NewRequest(http.MethodGet, "/v1/datasets", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if nextCalls != 0 {
		t.Fatalf("unexpected next handler calls: %d", nextCalls)
	}
	if got := rr.Header().Get(rateLimitHeaderRetryAfter); got == "" {
		t.Fatal("retry-after header missing")
	}
	if got := rr.Header().Get(rateLimitHeaderReset); got != strconv.FormatInt(repo.result.ResetAt.Unix(), 10) {
		t.Fatalf("unexpected reset header: %q", got)
	}

	var body rateLimitErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Error.Code != rateLimitErrorCode || body.Error.RequestID == "" || body.Error.RetryAfterSeconds <= 0 {
		t.Fatalf("unexpected error body: %#v", body)
	}
}

func TestRateLimitFailsOpenOnUnavailableRepository(t *testing.T) {
	repo := &rateLimitRepositoryStub{err: interfaces.ErrRateLimitUnavailable}
	anon := &anonymousIdentifierStub{value: "anon-opaque"}
	mw, err := RateLimit(repo, anon, RateLimitPolicy{AnonymousLimit: 60, APIKeyLimit: 300, DownloadLimit: 10, Window: time.Minute})
	if err != nil {
		t.Fatalf("RateLimit() error = %v", err)
	}

	nextCalls := 0
	handler := RequestID(mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalls++
	})))

	req := httptest.NewRequest(http.MethodGet, "/v1/datasets", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if nextCalls != 1 {
		t.Fatalf("unexpected next handler calls: %d", nextCalls)
	}
	if got := rr.Header().Get(rateLimitHeaderLimit); got != "" {
		t.Fatalf("unexpected limit header: %q", got)
	}
	if got := rr.Header().Get(rateLimitHeaderRemaining); got != "" {
		t.Fatalf("unexpected remaining header: %q", got)
	}
	if got := rr.Header().Get(rateLimitHeaderReset); got != "" {
		t.Fatalf("unexpected reset header: %q", got)
	}
	if got := rr.Header().Get(rateLimitHeaderRetryAfter); got != "" {
		t.Fatalf("unexpected retry-after header: %q", got)
	}
}

func TestRateLimitPropagatesContextErrorsAndSanitizesUnknownFailures(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{name: "canceled", err: context.Canceled, want: http.StatusServiceUnavailable},
		{name: "deadline", err: context.DeadlineExceeded, want: http.StatusServiceUnavailable},
		{name: "invalid input", err: interfaces.ErrInvalidRateLimitInput, want: http.StatusInternalServerError},
		{name: "unknown", err: errors.New("redis timeout at 127.0.0.1:6379"), want: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &rateLimitRepositoryStub{err: tc.err}
			anon := &anonymousIdentifierStub{value: "anon-opaque"}
			mw, err := RateLimit(repo, anon, RateLimitPolicy{AnonymousLimit: 60, APIKeyLimit: 300, DownloadLimit: 10, Window: time.Minute})
			if err != nil {
				t.Fatalf("RateLimit() error = %v", err)
			}

			nextCalls := 0
			handler := RequestID(mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				nextCalls++
			})))

			req := httptest.NewRequest(http.MethodGet, "/v1/datasets", nil)
			req.RemoteAddr = "203.0.113.10:1234"
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.want {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			if nextCalls != 0 {
				t.Fatalf("unexpected next handler calls: %d", nextCalls)
			}
			if tc.want == http.StatusServiceUnavailable {
				var body interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
					t.Fatalf("json.Unmarshal: %v", err)
				}
			}
		})
	}
}

func TestSecurityAnonymousIdentifierIgnoresForwardedHeaders(t *testing.T) {
	identifier, err := NewSecurityAnonymousIdentifier("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("NewSecurityAnonymousIdentifier() error = %v", err)
	}

	reqA := httptest.NewRequest(http.MethodGet, "/", nil)
	reqA.RemoteAddr = "203.0.113.10:1234"
	reqA.Header.Set("User-Agent", "TestAgent/1.0")
	reqA.Header.Set("X-Forwarded-For", "198.51.100.1")

	reqB := httptest.NewRequest(http.MethodGet, "/", nil)
	reqB.RemoteAddr = "203.0.113.10:1234"
	reqB.Header.Set("User-Agent", "TestAgent/1.0")
	reqB.Header.Set("X-Forwarded-For", "198.51.100.2")

	idA, err := identifier.Identify(reqA)
	if err != nil {
		t.Fatalf("Identify() error = %v", err)
	}
	idB, err := identifier.Identify(reqB)
	if err != nil {
		t.Fatalf("Identify() error = %v", err)
	}
	if idA != idB {
		t.Fatal("forwarded headers unexpectedly influenced anonymous identifier")
	}
}
