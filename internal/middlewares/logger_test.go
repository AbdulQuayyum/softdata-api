package middlewares

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.String()
}

type richResponseWriter struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func newRichResponseWriter() *richResponseWriter {
	return &richResponseWriter{header: make(http.Header)}
}

func (w *richResponseWriter) Header() http.Header        { return w.header }
func (w *richResponseWriter) WriteHeader(statusCode int) { w.status = statusCode }
func (w *richResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(p)
}
func (w *richResponseWriter) Flush() {}
func (w *richResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, http.ErrNotSupported
}
func (w *richResponseWriter) Push(string, *http.PushOptions) error { return http.ErrNotSupported }
func (w *richResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return io.Copy(&w.body, r)
}

func TestLoggerRecordsSafeRequestMetadata(t *testing.T) {
	var buf lockedBuffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	mw, err := NewLogger(logger)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	handler := RequestID(mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(http.Flusher); !ok {
			t.Fatal("flusher interface not preserved")
		}
		w.Header().Set("X-Handler", "ok")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("payload"))
		w.WriteHeader(http.StatusTeapot)
	})))

	req := httptest.NewRequest(http.MethodPost, "/v1/datasets?search=kwara", strings.NewReader("secret body"))
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Cookie", "session=secret")
	req.Header.Set("X-API-Key", "sd_live_secret")
	req.Header.Set(requestIDHeader, "req_abc123")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if rr.Header().Get("X-Handler") != "ok" {
		t.Fatal("downstream headers not preserved")
	}

	records := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(records) != 1 {
		t.Fatalf("expected one log record, got %d: %q", len(records), buf.String())
	}
	var entry map[string]any
	if err := json.Unmarshal([]byte(records[0]), &entry); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if entry["msg"] != "http request" {
		t.Fatalf("unexpected log message: %#v", entry["msg"])
	}
	if entry["request_id"] != "req_abc123" {
		t.Fatalf("unexpected request id: %#v", entry["request_id"])
	}
	if entry["method"] != "POST" {
		t.Fatalf("unexpected method: %#v", entry["method"])
	}
	if entry["path"] != "/v1/datasets" {
		t.Fatalf("unexpected path: %#v", entry["path"])
	}
	if _, ok := entry["duration"]; !ok {
		t.Fatal("duration missing from log entry")
	}
	if entry["status"] != float64(http.StatusCreated) {
		t.Fatalf("unexpected status: %#v", entry["status"])
	}
	if entry["bytes"] != float64(len("payload")) {
		t.Fatalf("unexpected bytes: %#v", entry["bytes"])
	}
	if strings.Contains(records[0], "secret body") || strings.Contains(records[0], "Bearer secret") || strings.Contains(records[0], "session=secret") || strings.Contains(records[0], "sd_live_secret") {
		t.Fatalf("sensitive value leaked into logs: %s", records[0])
	}
}

func TestLoggerPreservesOptionalResponseWriterInterfaces(t *testing.T) {
	var buf lockedBuffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	mw, err := NewLogger(logger)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(http.Flusher); !ok {
			t.Fatal("flusher not preserved")
		}
		if _, ok := w.(http.Hijacker); !ok {
			t.Fatal("hijacker not preserved")
		}
		if _, ok := w.(http.Pusher); !ok {
			t.Fatal("pusher not preserved")
		}
		if _, ok := w.(io.ReaderFrom); !ok {
			t.Fatal("readerfrom not preserved")
		}
		_, _ = w.Write([]byte("ok"))
	}))

	writer := newRichResponseWriter()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(writer, req)

	if writer.status != http.StatusOK {
		t.Fatalf("unexpected status: %d", writer.status)
	}
}

func TestLoggerRequiresLogger(t *testing.T) {
	if _, err := NewLogger(nil); err == nil {
		t.Fatal("NewLogger() error = nil, want error")
	}
}
