package middlewares

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// NewLogger returns middleware that records safe request metadata with slog.
func NewLogger(logger *slog.Logger) (Middleware, error) {
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := newLoggingResponseWriter(w)
			next.ServeHTTP(recorder, r)

			path := r.URL.Path
			if r.Pattern != "" {
				path = r.Pattern
			}

			requestID, _ := RequestIDFromContext(r.Context())
			logger.LogAttrs(r.Context(), slog.LevelInfo, "http request",
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", path),
				slog.Int("status", recorder.Status()),
				slog.Int64("bytes", recorder.BytesWritten()),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}, nil
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w}
}

func (w *loggingResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	if w.status == 0 {
		w.status = statusCode
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func (w *loggingResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += int64(n)
	return n, err
}

func (w *loggingResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := io.Copy(w.ResponseWriter, r)
	w.bytes += n
	return n, err
}

func (w *loggingResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("underlying response writer does not support hijacking")
	}
	return hijacker.Hijack()
}

func (w *loggingResponseWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return pusher.Push(target, opts)
}

func (w *loggingResponseWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *loggingResponseWriter) BytesWritten() int64 {
	return w.bytes
}
