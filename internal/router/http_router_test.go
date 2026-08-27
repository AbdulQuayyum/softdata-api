package router

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestHTTPRouterStreamsSuccessfulResponsesWithoutBuffering(t *testing.T) {
	t.Parallel()

	writer := newStreamingTestWriter()
	mux := http.NewServeMux()
	catalog := &routeCatalog{}
	errCh := make(chan error, 1)

	mux.Handle("GET /stream", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(*streamingTestWriter); !ok {
			errCh <- errors.New("handler received unexpected writer type")
			return
		}
		_, _ = io.WriteString(w, "chunk-1")
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		} else {
			errCh <- errors.New("handler did not receive flusher")
			return
		}
		<-writer.resume
		_, _ = io.WriteString(w, "chunk-2")
	}))
	if err := catalog.add("GET /stream"); err != nil {
		t.Fatalf("catalog.add() error = %v", err)
	}

	router := newHTTPRouter(mux, catalog)
	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	done := make(chan struct{})
	go func() {
		router.ServeHTTP(writer, req)
		close(done)
	}()

	writer.waitForFlush(t)
	if got := writer.bodyString(); got != "chunk-1" {
		t.Fatalf("unexpected streamed prefix before handler completion: %q", got)
	}
	select {
	case <-done:
		t.Fatal("router finished before the handler was released")
	default:
	}

	writer.release()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("router did not finish")
	}
	select {
	case err := <-errCh:
		t.Fatalf("handler error: %v", err)
	default:
	}

	if got := writer.bodyString(); got != "chunk-1chunk-2" {
		t.Fatalf("unexpected streamed body: %q", got)
	}
	if writer.flushCount() == 0 {
		t.Fatal("expected flush to reach the underlying writer")
	}
}

func TestHTTPRouterPreservesOptionalInterfacesAndReaderFrom(t *testing.T) {
	t.Parallel()

	writer := newOptionalInterfaceWriter()
	mux := http.NewServeMux()
	catalog := &routeCatalog{}
	payload := strings.Repeat("x", 128*1024)

	mux.Handle("GET /large", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(*optionalInterfaceWriter); !ok {
			t.Fatalf("handler received unexpected writer type %T", w)
		}
		if _, ok := w.(http.Flusher); !ok {
			t.Fatal("expected flusher support")
		}
		if _, ok := w.(http.Hijacker); !ok {
			t.Fatal("expected hijacker support")
		}
		if _, ok := w.(http.Pusher); !ok {
			t.Fatal("expected pusher support")
		}

		controller := http.NewResponseController(w)
		if err := controller.Flush(); err != nil {
			t.Fatalf("ResponseController.Flush() error = %v", err)
		}
		if _, err := io.Copy(w, io.LimitReader(strings.NewReader(payload), int64(len(payload)))); err != nil {
			t.Fatalf("io.Copy() error = %v", err)
		}
	}))
	if err := catalog.add("GET /large"); err != nil {
		t.Fatalf("catalog.add() error = %v", err)
	}

	router := newHTTPRouter(mux, catalog)
	req := httptest.NewRequest(http.MethodGet, "/large", nil)
	router.ServeHTTP(writer, req)

	if got := writer.body.String(); got != payload {
		t.Fatalf("unexpected response body size: got %d want %d", len(got), len(payload))
	}
	if writer.readerFromCount() == 0 {
		t.Fatal("expected ReaderFrom to be used on the underlying writer")
	}
	if writer.flushCount() == 0 {
		t.Fatal("expected flush to reach the underlying writer")
	}
}

func TestHTTPRouterLeavesHandlerGeneratedErrorsUnchanged(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		path   string
		status int
		body   string
	}{
		{
			name:   "404",
			path:   "/generated-404",
			status: http.StatusNotFound,
			body:   "handler not found",
		},
		{
			name:   "405",
			path:   "/generated-405",
			status: http.StatusMethodNotAllowed,
			body:   "handler method not allowed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			catalog := &routeCatalog{}
			mux.Handle("GET "+tc.path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			if err := catalog.add("GET " + tc.path); err != nil {
				t.Fatalf("catalog.add() error = %v", err)
			}

			router := newHTTPRouter(mux, catalog)
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			router.ServeHTTP(rr, req)

			if rr.Code != tc.status {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			if got := rr.Body.String(); got != tc.body {
				t.Fatalf("unexpected body: %q", got)
			}
		})
	}
}

type streamingTestWriter struct {
	mu         sync.Mutex
	header     http.Header
	body       bytes.Buffer
	status     int
	flushes    int
	ready      chan struct{}
	resume     chan struct{}
	readyOnce  sync.Once
	resumeOnce sync.Once
}

func newStreamingTestWriter() *streamingTestWriter {
	return &streamingTestWriter{
		header: make(http.Header),
		ready:  make(chan struct{}),
		resume: make(chan struct{}),
	}
}

func (w *streamingTestWriter) Header() http.Header {
	return w.header
}

func (w *streamingTestWriter) WriteHeader(statusCode int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.status != 0 {
		return
	}
	w.status = statusCode
}

func (w *streamingTestWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(p)
}

func (w *streamingTestWriter) Flush() {
	w.mu.Lock()
	w.flushes++
	w.mu.Unlock()
	w.readyOnce.Do(func() {
		close(w.ready)
	})
}

func (w *streamingTestWriter) release() {
	w.resumeOnce.Do(func() {
		close(w.resume)
	})
}

func (w *streamingTestWriter) waitForFlush(t *testing.T) {
	t.Helper()
	select {
	case <-w.ready:
	case <-time.After(time.Second):
		t.Fatal("handler did not flush")
	}
}

func (w *streamingTestWriter) bodyString() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.body.String()
}

func (w *streamingTestWriter) flushCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.flushes
}

type optionalInterfaceWriter struct {
	mu              sync.Mutex
	header          http.Header
	body            bytes.Buffer
	status          int
	flushes         int
	readerFromCalls int
}

func newOptionalInterfaceWriter() *optionalInterfaceWriter {
	return &optionalInterfaceWriter{header: make(http.Header)}
}

func (w *optionalInterfaceWriter) Header() http.Header {
	return w.header
}

func (w *optionalInterfaceWriter) WriteHeader(statusCode int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.status != 0 {
		return
	}
	w.status = statusCode
}

func (w *optionalInterfaceWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(p)
}

func (w *optionalInterfaceWriter) Flush() {
	w.mu.Lock()
	w.flushes++
	w.mu.Unlock()
}

func (w *optionalInterfaceWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("hijack not supported")
}

func (w *optionalInterfaceWriter) Push(string, *http.PushOptions) error {
	return errors.New("push not supported")
}

func (w *optionalInterfaceWriter) ReadFrom(r io.Reader) (int64, error) {
	w.mu.Lock()
	w.readerFromCalls++
	w.mu.Unlock()
	return io.Copy(&w.body, r)
}

func (w *optionalInterfaceWriter) flushCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.flushes
}

func (w *optionalInterfaceWriter) readerFromCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.readerFromCalls
}
