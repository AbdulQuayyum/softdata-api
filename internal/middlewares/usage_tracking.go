package middlewares

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

const maxUsageTrackingEndpointLen = 200

// UsageRecorder records safe request metadata without exposing persistence details to middleware.
type UsageRecorder interface {
	RecordRequest(ctx context.Context, input services.RequestRecordInput) (models.APIRequest, error)
}

// UsageTrackingOptions configures bounded request recording.
type UsageTrackingOptions struct {
	Timeout             time.Duration
	Now                 func() time.Time
	AnonymousIdentifier AnonymousIdentifier
}

type usageTrackingConfig struct {
	recorder     UsageRecorder
	endpoint     string
	datasetGroup string
	timeout      time.Duration
	now          func() time.Time
	anonymousID  AnonymousIdentifier
}

// UsageTracking returns middleware that records completed requests with a bounded, best-effort write.
func UsageTracking(recorder UsageRecorder, endpoint, datasetGroup string, options UsageTrackingOptions) (Middleware, error) {
	if recorder == nil {
		return nil, fmt.Errorf("usage recorder is required")
	}
	if options.AnonymousIdentifier == nil {
		return nil, fmt.Errorf("anonymous identifier is required")
	}

	endpoint, err := normalizeUsageTrackingEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	datasetGroup, err = normalizeUsageTrackingDatasetGroup(datasetGroup)
	if err != nil {
		return nil, err
	}
	if options.Timeout <= 0 {
		return nil, fmt.Errorf("usage tracking timeout must be positive")
	}

	now := options.Now
	if now == nil {
		now = time.Now
	}

	cfg := usageTrackingConfig{
		recorder:     recorder,
		endpoint:     endpoint,
		datasetGroup: datasetGroup,
		timeout:      options.Timeout,
		now:          now,
		anonymousID:  options.AnonymousIdentifier,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorderWriter := newUsageTrackingResponseWriter(w)
			next.ServeHTTP(recorderWriter, r)

			input, ok := cfg.buildRequestRecordInput(r, recorderWriter.Status(), start, recorderWriter.BytesWritten())
			if !ok {
				return
			}
			cfg.record(input)
		})
	}, nil
}

func (c usageTrackingConfig) buildRequestRecordInput(r *http.Request, statusCode int, start time.Time, responseBytes int64) (services.RequestRecordInput, bool) {
	requestID := requestIDFromRequest(r)
	if requestID == "" {
		return services.RequestRecordInput{}, false
	}

	input := services.RequestRecordInput{
		RequestID:     requestID,
		Method:        strings.TrimSpace(r.Method),
		Route:         c.endpoint,
		DatasetGroup:  ptrString(c.datasetGroup),
		StatusCode:    statusCode,
		RecordedAt:    c.now().UTC(),
		ResponseBytes: ptrInt64(responseBytes),
	}

	if input.Method == "" {
		return services.RequestRecordInput{}, false
	}

	duration := time.Since(start)
	if duration < 0 {
		duration = 0
	}
	ms := duration.Milliseconds()
	input.ResponseTimeMS = &ms

	if identity, ok := APIKeyIdentityFromContext(r.Context()); ok {
		input.AccountID = ptrString(identity.AccountID)
		input.APIKeyID = ptrString(identity.APIKeyID)
		return input, true
	}
	if identity, ok := AccountIdentityFromContext(r.Context()); ok {
		input.AccountID = ptrString(identity.AccountID)
		return input, true
	}

	anonymousID, err := c.anonymousID.Identify(r)
	if err != nil {
		return services.RequestRecordInput{}, false
	}
	input.AnonymousID = ptrString(anonymousID)
	return input, true
}

func (c usageTrackingConfig) record(input services.RequestRecordInput) {
	defer func() {
		_ = recover()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	_, _ = c.recorder.RecordRequest(ctx, input)
}

func normalizeUsageTrackingEndpoint(endpoint string) (string, error) {
	value := strings.TrimSpace(endpoint)
	switch {
	case value == "":
		return "", fmt.Errorf("usage endpoint is required")
	case len(value) > maxUsageTrackingEndpointLen:
		return "", fmt.Errorf("usage endpoint is too long")
	case strings.ContainsAny(value, " \t\r\n"):
		return "", fmt.Errorf("usage endpoint is invalid")
	case strings.Contains(value, "?") || strings.Contains(value, "://"):
		return "", fmt.Errorf("usage endpoint must be normalized")
	}
	return value, nil
}

func normalizeUsageTrackingDatasetGroup(datasetGroup string) (string, error) {
	value := strings.TrimSpace(datasetGroup)
	if value == "" {
		return "", nil
	}

	switch strings.ToLower(value) {
	case "geography", "finance", "education", "healthcare", "emergency", "infrastructure", "statistics":
		return strings.ToLower(value), nil
	default:
		return "", fmt.Errorf("usage dataset group is invalid")
	}
}

func ptrString(value string) *string {
	if value == "" {
		return nil
	}
	cloned := value
	return &cloned
}

func ptrInt64(value int64) *int64 {
	cloned := value
	return &cloned
}

type usageTrackingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func newUsageTrackingResponseWriter(w http.ResponseWriter) *usageTrackingResponseWriter {
	return &usageTrackingResponseWriter{ResponseWriter: w}
}

func (w *usageTrackingResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *usageTrackingResponseWriter) WriteHeader(statusCode int) {
	if w.status != 0 {
		return
	}
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *usageTrackingResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += int64(n)
	return n, err
}

func (w *usageTrackingResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := io.Copy(w.ResponseWriter, r)
	w.bytes += n
	return n, err
}

func (w *usageTrackingResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *usageTrackingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("underlying response writer does not support hijacking")
	}
	return hijacker.Hijack()
}

func (w *usageTrackingResponseWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return pusher.Push(target, opts)
}

func (w *usageTrackingResponseWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *usageTrackingResponseWriter) BytesWritten() int64 {
	return w.bytes
}
