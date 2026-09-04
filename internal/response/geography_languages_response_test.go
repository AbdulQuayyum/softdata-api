package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

func TestErrorMapsLanguageServiceErrors(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{
		{services.ErrLanguageNotFound, http.StatusNotFound},
		{fmt.Errorf("wrapped: %w", services.ErrLanguageNotFound), http.StatusNotFound},
		{services.ErrInvalidLanguageID, http.StatusBadRequest},
		{services.ErrInvalidCountryLanguageStatus, http.StatusBadRequest},
	} {
		rr := httptest.NewRecorder()
		if err := Error(rr, tc.err, "req-language"); err != nil {
			t.Fatalf("Error() error = %v", err)
		}
		if rr.Code != tc.status {
			t.Fatalf("Error(%v) status = %d, want %d", tc.err, rr.Code, tc.status)
		}
		var body ErrorResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil || body.Success {
			t.Fatalf("unexpected error body: %s", rr.Body.String())
		}
	}
}
