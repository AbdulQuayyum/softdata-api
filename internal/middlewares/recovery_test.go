package middlewares

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/response"
)

func TestRecoveryPassesThroughNormalRequests(t *testing.T) {
	called := 0
	handler := Recovery()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusNoContent)
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if called != 1 {
		t.Fatalf("handler called %d times", called)
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestRecoveryWritesGenericErrorOnPanic(t *testing.T) {
	handler := RequestID(Recovery()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("secret panic text")
	})))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", rr.Code)
	}

	var body response.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.Error.Code != "INTERNAL_ERROR" {
		t.Fatalf("unexpected error code: %q", body.Error.Code)
	}
	if strings.Contains(rr.Body.String(), "secret panic text") {
		t.Fatalf("panic text leaked in response: %s", rr.Body.String())
	}
	if body.Error.RequestID == "" {
		t.Fatal("request id missing from panic response")
	}
}

func TestRecoveryPreservesErrAbortHandler(t *testing.T) {
	handler := Recovery()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	defer func() {
		if rec := recover(); !errors.Is(asError(rec), http.ErrAbortHandler) && rec != http.ErrAbortHandler {
			t.Fatalf("panic recovery changed ErrAbortHandler: %#v", rec)
		}
	}()

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	t.Fatal("expected panic")
}
