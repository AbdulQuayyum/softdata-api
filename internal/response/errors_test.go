package response

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

func TestErrorMapsKnownServiceErrors(t *testing.T) {
	requestID := "req-123"
	cases := []struct {
		name      string
		err       error
		status    int
		code      string
		message   string
		requestID string
	}{
		{name: "invalid credentials", err: services.ErrInvalidCredentials, status: http.StatusUnauthorized, code: codeInvalidCredentials, message: messageInvalidCredentials, requestID: requestID},
		{name: "invalid refresh token", err: services.ErrInvalidRefreshToken, status: http.StatusUnauthorized, code: codeInvalidCredentials, message: messageInvalidCredentials, requestID: requestID},
		{name: "username unavailable", err: services.ErrUsernameUnavailable, status: http.StatusConflict, code: codeResourceConflict, message: messageResourceConflict, requestID: requestID},
		{name: "email unavailable", err: services.ErrEmailUnavailable, status: http.StatusConflict, code: codeResourceConflict, message: messageResourceConflict, requestID: requestID},
		{name: "account not found", err: services.ErrAccountNotFound, status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "current password invalid", err: services.ErrCurrentPasswordInvalid, status: http.StatusUnauthorized, code: codeInvalidCredentials, message: messageInvalidCredentials, requestID: requestID},
		{name: "account inactive", err: services.ErrAccountInactive, status: http.StatusForbidden, code: codeInvalidRequest, message: messageOperationNotAllowed, requestID: requestID},
		{name: "api key not found", err: services.ErrAPIKeyNotFound, status: http.StatusUnauthorized, code: codeInvalidAPIKey, message: messageInvalidAPIKey, requestID: requestID},
		{name: "api key name required", err: services.ErrAPIKeyNameRequired, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "api key limit reached", err: services.ErrAPIKeyLimitReached, status: http.StatusConflict, code: codeResourceConflict, message: messageResourceConflict, requestID: requestID},
		{name: "invalid usage period", err: services.ErrInvalidUsagePeriod, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "invalid state id", err: services.ErrInvalidStateID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "invalid local government unit id", err: services.ErrInvalidLocalGovernmentUnitID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "invalid geopolitical zone id", err: services.ErrInvalidGeopoliticalZoneID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "payment service provider not found", err: services.ErrPaymentServiceProviderNotFound, status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "invalid payment service provider id", err: services.ErrInvalidPaymentServiceProviderID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "invalid payment service provider type", err: services.ErrInvalidPaymentServiceProviderType, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "university not found", err: services.ErrUniversityNotFound, status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "invalid university id", err: services.ErrInvalidUniversityID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "invalid university ownership type", err: services.ErrInvalidUniversityOwnershipType, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "invalid university state id", err: services.ErrInvalidUniversityStateID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "college of education not found", err: services.ErrCollegeOfEducationNotFound, status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "invalid college of education id", err: services.ErrInvalidCollegeOfEducationID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "invalid college of education ownership type", err: services.ErrInvalidCollegeOfEducationOwnershipType, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "invalid college of education state id", err: services.ErrInvalidCollegeOfEducationStateID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "international money transfer operator not found", err: services.ErrInternationalMoneyTransferOperatorNotFound, status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "invalid international money transfer operator id", err: services.ErrInvalidInternationalMoneyTransferOperatorID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "country or area not found", err: services.ErrCountryOrAreaNotFound, status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "invalid country or area id", err: services.ErrInvalidCountryOrAreaID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "invalid country or area region code", err: services.ErrInvalidCountryOrAreaRegionCode, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "invalid country or area subregion code", err: services.ErrInvalidCountryOrAreaSubregionCode, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "dataset not found", err: services.ErrDatasetNotFound, status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "state not found", err: services.ErrStateNotFound, status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "local government unit not found", err: services.ErrLocalGovernmentUnitNotFound, status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "time zone not found", err: services.ErrTimeZoneNotFound, status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "invalid time zone id", err: services.ErrInvalidTimeZoneID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "invalid time zone country area id", err: services.ErrInvalidTimeZoneCountryAreaID, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "geopolitical zone not found", err: services.ErrGeopoliticalZoneNotFound, status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "invalid dataset key", err: services.ErrInvalidDatasetKey, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "invalid pagination", err: services.ErrInvalidPagination, status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped dataset not found", err: fmt.Errorf("wrap: %w", services.ErrDatasetNotFound), status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "wrapped api key not found", err: fmt.Errorf("wrap: %w", services.ErrAPIKeyNotFound), status: http.StatusUnauthorized, code: codeInvalidAPIKey, message: messageInvalidAPIKey, requestID: requestID},
		{name: "wrapped invalid credentials", err: fmt.Errorf("wrap: %w", services.ErrInvalidCredentials), status: http.StatusUnauthorized, code: codeInvalidCredentials, message: messageInvalidCredentials, requestID: requestID},
		{name: "wrapped invalid geopolitical zone id", err: fmt.Errorf("wrap: %w", services.ErrInvalidGeopoliticalZoneID), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped payment service provider not found", err: fmt.Errorf("wrap: %w", services.ErrPaymentServiceProviderNotFound), status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "wrapped invalid payment service provider id", err: fmt.Errorf("wrap: %w", services.ErrInvalidPaymentServiceProviderID), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped invalid payment service provider type", err: fmt.Errorf("wrap: %w", services.ErrInvalidPaymentServiceProviderType), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped university not found", err: fmt.Errorf("wrap: %w", services.ErrUniversityNotFound), status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "wrapped invalid university id", err: fmt.Errorf("wrap: %w", services.ErrInvalidUniversityID), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped invalid university ownership type", err: fmt.Errorf("wrap: %w", services.ErrInvalidUniversityOwnershipType), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped invalid university state id", err: fmt.Errorf("wrap: %w", services.ErrInvalidUniversityStateID), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped college of education not found", err: fmt.Errorf("wrap: %w", services.ErrCollegeOfEducationNotFound), status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "wrapped invalid college of education id", err: fmt.Errorf("wrap: %w", services.ErrInvalidCollegeOfEducationID), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped invalid college of education ownership type", err: fmt.Errorf("wrap: %w", services.ErrInvalidCollegeOfEducationOwnershipType), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped invalid college of education state id", err: fmt.Errorf("wrap: %w", services.ErrInvalidCollegeOfEducationStateID), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped international money transfer operator not found", err: fmt.Errorf("wrap: %w", services.ErrInternationalMoneyTransferOperatorNotFound), status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "wrapped invalid international money transfer operator id", err: fmt.Errorf("wrap: %w", services.ErrInvalidInternationalMoneyTransferOperatorID), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped country or area not found", err: fmt.Errorf("wrap: %w", services.ErrCountryOrAreaNotFound), status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "wrapped invalid country or area id", err: fmt.Errorf("wrap: %w", services.ErrInvalidCountryOrAreaID), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped invalid country or area region code", err: fmt.Errorf("wrap: %w", services.ErrInvalidCountryOrAreaRegionCode), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped invalid country or area subregion code", err: fmt.Errorf("wrap: %w", services.ErrInvalidCountryOrAreaSubregionCode), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped invalid local government unit id", err: fmt.Errorf("wrap: %w", services.ErrInvalidLocalGovernmentUnitID), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped time zone not found", err: fmt.Errorf("wrap: %w", services.ErrTimeZoneNotFound), status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "wrapped invalid time zone id", err: fmt.Errorf("wrap: %w", services.ErrInvalidTimeZoneID), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped invalid time zone country area id", err: fmt.Errorf("wrap: %w", services.ErrInvalidTimeZoneCountryAreaID), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "wrapped geopolitical zone not found", err: fmt.Errorf("wrap: %w", services.ErrGeopoliticalZoneNotFound), status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "wrapped local government unit not found", err: fmt.Errorf("wrap: %w", services.ErrLocalGovernmentUnitNotFound), status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "wrapped invalid pagination", err: fmt.Errorf("wrap: %w", services.ErrInvalidPagination), status: http.StatusBadRequest, code: codeInvalidRequest, message: messageInvalidRequest, requestID: requestID},
		{name: "context deadline", err: context.DeadlineExceeded, status: http.StatusServiceUnavailable, code: codeServiceUnavailable, message: messageServiceUnavailable, requestID: requestID},
		{name: "repo not found fallback", err: interfaces.ErrNotFound, status: http.StatusNotFound, code: codeResourceNotFound, message: messageResourceNotFound, requestID: requestID},
		{name: "repo conflict fallback", err: interfaces.ErrConflict, status: http.StatusConflict, code: codeResourceConflict, message: messageResourceConflict, requestID: requestID},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			if err := Error(rr, tc.err, tc.requestID); err != nil {
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
			if body.Error.RequestID != tc.requestID {
				t.Fatalf("unexpected request id: %q", body.Error.RequestID)
			}
			if body.Error.Details != nil {
				t.Fatalf("unexpected details: %#v", body.Error.Details)
			}
		})
	}
}

func TestErrorValidationAndSafety(t *testing.T) {
	rr := httptest.NewRecorder()
	details := []ValidationError{{Field: "limit", Message: "Limit must not exceed 100."}}
	if err := Validation(rr, "req-456", details); err != nil {
		t.Fatalf("Validation() error = %v", err)
	}

	var body ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if body.Error.Code != codeValidationFailed {
		t.Fatalf("unexpected code: %q", body.Error.Code)
	}
	if len(body.Error.Details) != 1 || body.Error.Details[0].Field != "limit" {
		t.Fatalf("unexpected validation details: %#v", body.Error.Details)
	}
}

func TestErrorFallsBackToGenericInternalError(t *testing.T) {
	rr := httptest.NewRecorder()
	want := fmt.Errorf("repository unavailable: connect postgres://user:pass@host/db")
	if err := Error(rr, want, "req-789"); err != nil {
		t.Fatalf("Error() error = %v", err)
	}

	body := ErrorResponse{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if body.Error.Code != codeInternalError {
		t.Fatalf("unexpected code: %q", body.Error.Code)
	}
	if strings.Contains(body.Error.Message, "postgres") || strings.Contains(body.Error.Message, "host") {
		t.Fatalf("generic message exposed internal details: %q", body.Error.Message)
	}
}
