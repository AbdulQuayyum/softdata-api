package router

import (
	"errors"
	"io/fs"
	"net/http"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/datasets/assets"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
)

var commercialBankAssetIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+\.png$`)

func serveCommercialBankLogo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		_ = response.JSON(w, http.StatusMethodNotAllowed, response.ErrorResponse{
			Success: false,
			Error:   response.ErrorBody{Code: "INVALID_REQUEST", Message: "The request was invalid."},
		})
		return
	}

	requestID := requestIDFromContext(r.Context())
	assetID := strings.TrimSpace(r.PathValue("bank_asset"))
	if !commercialBankAssetIDPattern.MatchString(assetID) {
		_ = response.Validation(w, requestID, []response.ValidationError{{
			Field:   "bank_asset",
			Message: "Bank logo ID must be a lowercase bank ID ending with .png.",
		}})
		return
	}

	bankID := strings.TrimSuffix(assetID, ".png")
	logo, err := assets.BankLogo(bankID, "png")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			_ = response.Error(w, interfaces.ErrNotFound, requestID)
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(logo)
}
