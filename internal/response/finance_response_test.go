package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

func TestErrorMapsPaymentServiceProviderServiceErrors(t *testing.T) {
	requestID := "req-psp"
	cases := []struct {
		name    string
		err     error
		status  int
		code    string
		message string
	}{
		{name: "not found", err: services.ErrPaymentServiceProviderNotFound, status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound},
		{name: "wrapped not found", err: fmt.Errorf("wrap: %w", services.ErrPaymentServiceProviderNotFound), status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound},
		{name: "invalid id", err: services.ErrInvalidPaymentServiceProviderID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest},
		{name: "wrapped invalid id", err: fmt.Errorf("wrap: %w", services.ErrInvalidPaymentServiceProviderID), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest},
		{name: "invalid type", err: services.ErrInvalidPaymentServiceProviderType, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest},
		{name: "wrapped invalid type", err: fmt.Errorf("wrap: %w", services.ErrInvalidPaymentServiceProviderType), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			if err := Error(rr, tc.err, requestID); err != nil {
				t.Fatalf("Error() error = %v", err)
			}
			if rr.Code != tc.status {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			var body ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if body.Success {
				t.Fatal("success should be false")
			}
			if body.Error.Code != tc.code {
				t.Fatalf("unexpected code: %q", body.Error.Code)
			}
			if body.Error.Message != tc.message {
				t.Fatalf("unexpected message: %q", body.Error.Message)
			}
			if body.Error.RequestID != requestID {
				t.Fatalf("unexpected request id: %q", body.Error.RequestID)
			}
		})
	}
}

func TestErrorMapsCurrencyServiceErrors(t *testing.T) {
	requestID := "req-currency"
	cases := []struct {
		name    string
		err     error
		status  int
		code    string
		message string
	}{
		{name: "not found", err: services.ErrCurrencyNotFound, status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound},
		{name: "wrapped not found", err: fmt.Errorf("wrap: %w", services.ErrCurrencyNotFound), status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound},
		{name: "invalid id", err: services.ErrInvalidCurrencyID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest},
		{name: "wrapped invalid id", err: fmt.Errorf("wrap: %w", services.ErrInvalidCurrencyID), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest},
		{name: "invalid country area id", err: services.ErrInvalidCurrencyCountryAreaID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest},
		{name: "wrapped invalid country area id", err: fmt.Errorf("wrap: %w", services.ErrInvalidCurrencyCountryAreaID), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			if err := Error(rr, tc.err, requestID); err != nil {
				t.Fatalf("Error() error = %v", err)
			}
			if rr.Code != tc.status {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			var body ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if body.Error.Code != tc.code || body.Error.Message != tc.message || body.Error.RequestID != requestID {
				t.Fatalf("unexpected response: %#v", body)
			}
		})
	}
}
