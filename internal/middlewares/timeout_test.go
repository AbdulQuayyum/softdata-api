package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeoutAddsDeadlineAndPropagatesCancellation(t *testing.T) {
	mw, err := NewTimeout(25 * time.Millisecond)
	if err != nil {
		t.Fatalf("NewTimeout() error = %v", err)
	}

	done := make(chan error, 1)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Context().Deadline()
		if !ok {
			t.Fatal("deadline missing from context")
		}
		<-r.Context().Done()
		done <- r.Context().Err()
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	select {
	case err := <-done:
		if err != context.DeadlineExceeded {
			t.Fatalf("unexpected context error: %v", err)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("timeout context did not expire")
	}
}

func TestTimeoutPreservesShorterExistingDeadline(t *testing.T) {
	mw, err := NewTimeout(time.Hour)
	if err != nil {
		t.Fatalf("NewTimeout() error = %v", err)
	}

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deadline, ok := r.Context().Deadline()
		if !ok {
			t.Fatal("deadline missing from context")
		}
		if time.Until(deadline) > 10*time.Minute {
			t.Fatalf("existing shorter deadline was not preserved: %v", deadline)
		}
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestTimeoutRejectsInvalidDuration(t *testing.T) {
	if _, err := NewTimeout(0); err == nil {
		t.Fatal("NewTimeout() error = nil, want error")
	}
}
