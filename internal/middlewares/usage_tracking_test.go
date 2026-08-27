package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type usageRecorderStub struct {
	mu       sync.Mutex
	input    services.RequestRecordInput
	calls    int
	err      error
	returned models.APIRequest
	lastCtx  context.Context
	recordFn func(context.Context, services.RequestRecordInput) (models.APIRequest, error)
}

func (s *usageRecorderStub) RecordRequest(ctx context.Context, input services.RequestRecordInput) (models.APIRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.input = input
	s.lastCtx = ctx
	if s.recordFn != nil {
		return s.recordFn(ctx, input)
	}
	return s.returned, s.err
}

type anonymousIdentifierUsageStub struct {
	mu    sync.Mutex
	value string
	err   error
	calls int
	last  *http.Request
}

func (s *anonymousIdentifierUsageStub) Identify(r *http.Request) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.last = r
	return s.value, s.err
}

func TestUsageTrackingConstructorValidation(t *testing.T) {
	recorder := &usageRecorderStub{}
	anon := &anonymousIdentifierUsageStub{value: "anon-opaque"}
	options := UsageTrackingOptions{Timeout: time.Second, AnonymousIdentifier: anon}

	if _, err := UsageTracking(nil, "/v1/datasets", "", options); err == nil {
		t.Fatal("UsageTracking() error = nil, want error")
	}
	if _, err := UsageTracking(recorder, "/v1/datasets", "", UsageTrackingOptions{Timeout: time.Second}); err == nil {
		t.Fatal("UsageTracking() error = nil, want error")
	}
	if _, err := UsageTracking(recorder, "", "", options); err == nil {
		t.Fatal("UsageTracking() error = nil, want error")
	}
	if _, err := UsageTracking(recorder, strings.Repeat("a", maxUsageTrackingEndpointLen+1), "", options); err == nil {
		t.Fatal("UsageTracking() error = nil, want error")
	}
	if _, err := UsageTracking(recorder, "/v1/datasets?search=kwara", "", options); err == nil {
		t.Fatal("UsageTracking() error = nil, want error")
	}
	if _, err := UsageTracking(recorder, "/v1/datasets", "bad-group", options); err == nil {
		t.Fatal("UsageTracking() error = nil, want error")
	}
	for _, group := range []string{"geography", "finance", "education", "healthcare", "emergency", "infrastructure", "statistics"} {
		if _, err := UsageTracking(recorder, "/v1/datasets", group, options); err != nil {
			t.Fatalf("UsageTracking() error = %v for group %q", err, group)
		}
	}
}

func TestUsageTrackingAllowsEmptyGroupForNonDatasetRoute(t *testing.T) {
	recorder := &usageRecorderStub{}
	anon := &anonymousIdentifierUsageStub{value: "anon-opaque"}
	mw, err := UsageTracking(recorder, "/v1/account/api-keys", "", UsageTrackingOptions{Timeout: time.Second, AnonymousIdentifier: anon})
	if err != nil {
		t.Fatalf("UsageTracking() error = %v", err)
	}

	handler := RequestID(mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})))

	req := httptest.NewRequest(http.MethodGet, "/v1/account/api-keys?search=finance", nil)
	req.Header.Set(requestIDHeader, "req_empty_group")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if recorder.input.DatasetGroup != nil {
		t.Fatalf("unexpected dataset group: %#v", recorder.input.DatasetGroup)
	}
}

func TestUsageTrackingRecordsAnonymousRequestOnce(t *testing.T) {
	fixedNow := time.Date(2026, 8, 27, 12, 30, 0, 0, time.FixedZone("WAT", 3600))
	recorder := &usageRecorderStub{
		returned: models.APIRequest{ID: 99},
	}
	anon := &anonymousIdentifierUsageStub{value: "anon-opaque"}
	mw, err := UsageTracking(recorder, "/v1/datasets", "geography", UsageTrackingOptions{
		Timeout:             time.Second,
		AnonymousIdentifier: anon,
		Now: func() time.Time {
			return fixedNow
		},
	})
	if err != nil {
		t.Fatalf("UsageTracking() error = %v", err)
	}

	nextCalls := 0
	handler := RequestID(mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalls++
		if _, ok := RequestIDFromContext(r.Context()); !ok {
			t.Fatal("request id missing from context")
		}
		_, _ = w.Write([]byte("payload"))
	})))

	req := httptest.NewRequest(http.MethodGet, "/v1/datasets?search=secret", strings.NewReader("secret body"))
	req.RemoteAddr = "203.0.113.10:1234"
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set(requestIDHeader, "req_abc123")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if nextCalls != 1 {
		t.Fatalf("unexpected next handler calls: %d", nextCalls)
	}
	if recorder.calls != 1 {
		t.Fatalf("unexpected recorder calls: %d", recorder.calls)
	}
	if anon.calls != 1 {
		t.Fatalf("unexpected anonymous identifier calls: %d", anon.calls)
	}
	if recorder.input.RequestID != "req_abc123" {
		t.Fatalf("unexpected request id: %q", recorder.input.RequestID)
	}
	if recorder.input.Route != "/v1/datasets" {
		t.Fatalf("unexpected route: %q", recorder.input.Route)
	}
	if recorder.input.Method != http.MethodGet {
		t.Fatalf("unexpected method: %q", recorder.input.Method)
	}
	if recorder.input.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", recorder.input.StatusCode)
	}
	if recorder.input.AnonymousID == nil || *recorder.input.AnonymousID != "anon-opaque" {
		t.Fatalf("unexpected anonymous id: %#v", recorder.input.AnonymousID)
	}
	if recorder.input.AccountID != nil || recorder.input.APIKeyID != nil {
		t.Fatalf("unexpected authenticated identities: %#v", recorder.input)
	}
	if recorder.input.DatasetGroup == nil || *recorder.input.DatasetGroup != "geography" {
		t.Fatalf("unexpected dataset group: %#v", recorder.input.DatasetGroup)
	}
	if recorder.input.RecordedAt.Location() != time.UTC || !recorder.input.RecordedAt.Equal(fixedNow.UTC()) {
		t.Fatalf("unexpected recorded time: %v", recorder.input.RecordedAt)
	}
	if recorder.input.ResponseTimeMS == nil || *recorder.input.ResponseTimeMS < 0 {
		t.Fatalf("unexpected response time: %#v", recorder.input.ResponseTimeMS)
	}
	if recorder.input.ResponseBytes == nil || *recorder.input.ResponseBytes != int64(len("payload")) {
		t.Fatalf("unexpected response bytes: %#v", recorder.input.ResponseBytes)
	}
	if recorder.input.RequestBytes != nil {
		t.Fatalf("unexpected request bytes: %#v", recorder.input.RequestBytes)
	}
}

func TestUsageTrackingUsesApiKeyIdentityWhenAvailable(t *testing.T) {
	recorder := &usageRecorderStub{}
	anon := &anonymousIdentifierUsageStub{value: "anon-opaque"}
	mw, err := UsageTracking(recorder, "geography.states.get", "", UsageTrackingOptions{Timeout: time.Second, AnonymousIdentifier: anon})
	if err != nil {
		t.Fatalf("UsageTracking() error = %v", err)
	}

	handler := RequestID(mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/states", nil)
	req = req.WithContext(WithAccountIdentity(req.Context(), AccountIdentity{AccountID: "acct-bearer"}))
	req = req.WithContext(WithAPIKeyIdentity(req.Context(), services.APIKeyIdentity{APIKeyID: "key-123", AccountID: "acct-key"}))
	req.Header.Set(requestIDHeader, "req_api")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if recorder.calls != 1 {
		t.Fatalf("unexpected recorder calls: %d", recorder.calls)
	}
	if recorder.input.AccountID == nil || *recorder.input.AccountID != "acct-key" {
		t.Fatalf("unexpected account id: %#v", recorder.input.AccountID)
	}
	if recorder.input.APIKeyID == nil || *recorder.input.APIKeyID != "key-123" {
		t.Fatalf("unexpected api key id: %#v", recorder.input.APIKeyID)
	}
	if recorder.input.AnonymousID != nil {
		t.Fatalf("unexpected anonymous id: %#v", recorder.input.AnonymousID)
	}
	if recorder.input.Route != "geography.states.get" {
		t.Fatalf("unexpected route: %q", recorder.input.Route)
	}
	if recorder.input.DatasetGroup != nil {
		t.Fatalf("unexpected dataset group: %#v", recorder.input.DatasetGroup)
	}
}

func TestUsageTrackingUsesBearerIdentityWhenItIsTheOnlyIdentity(t *testing.T) {
	recorder := &usageRecorderStub{}
	anon := &anonymousIdentifierUsageStub{value: "anon-opaque"}
	mw, err := UsageTracking(recorder, "/v1/account/api-keys", "", UsageTrackingOptions{Timeout: time.Second, AnonymousIdentifier: anon})
	if err != nil {
		t.Fatalf("UsageTracking() error = %v", err)
	}

	handler := RequestID(mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	req := httptest.NewRequest(http.MethodGet, "/v1/account/api-keys", nil)
	req = req.WithContext(WithAccountIdentity(req.Context(), AccountIdentity{AccountID: "acct-bearer"}))
	req.Header.Set(requestIDHeader, "req_bearer")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if recorder.input.AccountID == nil || *recorder.input.AccountID != "acct-bearer" {
		t.Fatalf("unexpected account id: %#v", recorder.input.AccountID)
	}
	if recorder.input.APIKeyID != nil || recorder.input.AnonymousID != nil {
		t.Fatalf("unexpected extra identities: %#v", recorder.input)
	}
	if recorder.input.DatasetGroup != nil {
		t.Fatalf("unexpected dataset group: %#v", recorder.input.DatasetGroup)
	}
}

func TestUsageTrackingFallsBackToAnonymousWhenNoIdentityIsPresent(t *testing.T) {
	recorder := &usageRecorderStub{}
	anon := &anonymousIdentifierUsageStub{value: "anon-opaque"}
	mw, err := UsageTracking(recorder, "/v1/datasets", "geography", UsageTrackingOptions{Timeout: time.Second, AnonymousIdentifier: anon})
	if err != nil {
		t.Fatalf("UsageTracking() error = %v", err)
	}

	handler := RequestID(mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	req := httptest.NewRequest(http.MethodGet, "/v1/datasets", nil)
	req.Header.Set(requestIDHeader, "req_anon")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if recorder.input.AnonymousID == nil || *recorder.input.AnonymousID != "anon-opaque" {
		t.Fatalf("unexpected anonymous id: %#v", recorder.input.AnonymousID)
	}
	if recorder.input.AccountID != nil || recorder.input.APIKeyID != nil {
		t.Fatalf("unexpected authenticated identities: %#v", recorder.input)
	}
	if recorder.input.DatasetGroup == nil || *recorder.input.DatasetGroup != "geography" {
		t.Fatalf("unexpected dataset group: %#v", recorder.input.DatasetGroup)
	}
	if anon.calls != 1 {
		t.Fatalf("unexpected anonymous identifier calls: %d", anon.calls)
	}
}

func TestUsageTrackingRecordsAllDownstreamStatuses(t *testing.T) {
	cases := []struct {
		name   string
		status int
	}{
		{name: "implicit 200", status: http.StatusOK},
		{name: "201", status: http.StatusCreated},
		{name: "204", status: http.StatusNoContent},
		{name: "400", status: http.StatusBadRequest},
		{name: "401", status: http.StatusUnauthorized},
		{name: "404", status: http.StatusNotFound},
		{name: "429", status: http.StatusTooManyRequests},
		{name: "500", status: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &usageRecorderStub{}
			anon := &anonymousIdentifierUsageStub{value: "anon-opaque"}
			mw, err := UsageTracking(recorder, "/v1/datasets", "geography", UsageTrackingOptions{Timeout: time.Second, AnonymousIdentifier: anon})
			if err != nil {
				t.Fatalf("UsageTracking() error = %v", err)
			}

			handler := RequestID(mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.status != http.StatusOK {
					w.WriteHeader(tc.status)
				}
				if tc.status != http.StatusNoContent {
					_, _ = w.Write([]byte("body"))
				}
			})))
			req := httptest.NewRequest(http.MethodGet, "/v1/datasets?group=finance", nil)
			req.Header.Set(requestIDHeader, "req_status")
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.status {
				t.Fatalf("unexpected downstream status: %d", rr.Code)
			}
			if recorder.input.StatusCode != tc.status {
				t.Fatalf("unexpected recorded status: %d", recorder.input.StatusCode)
			}
			if recorder.input.DatasetGroup == nil || *recorder.input.DatasetGroup != "geography" {
				t.Fatalf("unexpected dataset group: %#v", recorder.input.DatasetGroup)
			}
			if recorder.input.ResponseTimeMS == nil {
				t.Fatal("response time missing")
			}
		})
	}
}

func TestUsageTrackingRecorderFailureDoesNotChangeResponse(t *testing.T) {
	recorder := &usageRecorderStub{err: errors.New("database unavailable")}
	anon := &anonymousIdentifierUsageStub{value: "anon-opaque"}
	mw, err := UsageTracking(recorder, "/v1/datasets", "geography", UsageTrackingOptions{Timeout: time.Second, AnonymousIdentifier: anon})
	if err != nil {
		t.Fatalf("UsageTracking() error = %v", err)
	}

	handler := RequestID(mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})))

	req := httptest.NewRequest(http.MethodPost, "/v1/datasets", strings.NewReader("secret body"))
	req.Header.Set(requestIDHeader, "req_error")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if rr.Body.String() != "created" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
	if recorder.calls != 1 {
		t.Fatalf("unexpected recorder calls: %d", recorder.calls)
	}
	if recorder.input.DatasetGroup == nil || *recorder.input.DatasetGroup != "geography" {
		t.Fatalf("unexpected dataset group: %#v", recorder.input.DatasetGroup)
	}
}

func TestUsageTrackingRecordsEvenWhenRequestContextIsCanceled(t *testing.T) {
	recorder := &usageRecorderStub{}
	anon := &anonymousIdentifierUsageStub{value: "anon-opaque"}
	mw, err := UsageTracking(recorder, "/v1/datasets", "geography", UsageTrackingOptions{Timeout: time.Second, AnonymousIdentifier: anon})
	if err != nil {
		t.Fatalf("UsageTracking() error = %v", err)
	}

	handler := RequestID(mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})))

	req := httptest.NewRequest(http.MethodGet, "/v1/datasets", nil)
	req.Header.Set(requestIDHeader, "req_cancel")
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if recorder.calls != 1 {
		t.Fatalf("unexpected recorder calls: %d", recorder.calls)
	}
	if recorder.input.DatasetGroup == nil || *recorder.input.DatasetGroup != "geography" {
		t.Fatalf("unexpected dataset group: %#v", recorder.input.DatasetGroup)
	}
}

func TestUsageTrackingResponseWriterPreservesOptionalInterfaces(t *testing.T) {
	base := newRichResponseWriter()
	wrapped := newUsageTrackingResponseWriter(base)

	if _, ok := any(wrapped).(http.Flusher); !ok {
		t.Fatal("flusher not preserved")
	}
	if _, ok := any(wrapped).(http.Hijacker); !ok {
		t.Fatal("hijacker not preserved")
	}
	if _, ok := any(wrapped).(http.Pusher); !ok {
		t.Fatal("pusher not preserved")
	}
	if _, ok := any(wrapped).(io.ReaderFrom); !ok {
		t.Fatal("readerfrom not preserved")
	}

	wrapped.WriteHeader(http.StatusCreated)
	wrapped.WriteHeader(http.StatusTeapot)
	_, _ = wrapped.Write([]byte("payload"))

	if wrapped.Status() != http.StatusCreated {
		t.Fatalf("unexpected status: %d", wrapped.Status())
	}
	if wrapped.BytesWritten() != int64(len("payload")) {
		t.Fatalf("unexpected bytes written: %d", wrapped.BytesWritten())
	}
	if base.status != http.StatusCreated {
		t.Fatalf("unexpected base status: %d", base.status)
	}
}

func TestUsageTrackingRejectsUnsafeResponseWriterBehavior(t *testing.T) {
	base := newRichResponseWriter()
	wrapped := newUsageTrackingResponseWriter(base)

	if wrapped.Status() != http.StatusOK {
		t.Fatalf("unexpected default status: %d", wrapped.Status())
	}
	if wrapped.BytesWritten() != 0 {
		t.Fatalf("unexpected default bytes: %d", wrapped.BytesWritten())
	}
}

func TestUsageTrackingRecorderOutputIsJSONSafe(t *testing.T) {
	recorder := &usageRecorderStub{}
	anon := &anonymousIdentifierUsageStub{value: "anon-opaque"}
	mw, err := UsageTracking(recorder, "/v1/datasets", "geography", UsageTrackingOptions{Timeout: time.Second, AnonymousIdentifier: anon})
	if err != nil {
		t.Fatalf("UsageTracking() error = %v", err)
	}

	handler := RequestID(mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})))
	req := httptest.NewRequest(http.MethodGet, "/v1/datasets?search=kwara", nil)
	req.Header.Set(requestIDHeader, "req_json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	b, err := json.Marshal(recorder.input)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if strings.Contains(string(b), "search=kwara") || strings.Contains(string(b), "203.0.113") {
		t.Fatalf("unsafe value leaked into request record: %s", string(b))
	}
	if !strings.Contains(string(b), "geography") {
		t.Fatalf("dataset group missing from request record: %s", string(b))
	}
}
