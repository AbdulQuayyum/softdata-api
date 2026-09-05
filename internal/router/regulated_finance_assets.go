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

var regulatedFinanceAssetIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*\.png$`)
var regulatedFinanceAssetCategoryPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func serveFinancialInstitutionLogo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		_ = response.JSON(w, http.StatusMethodNotAllowed, response.ErrorResponse{Success: false, Error: response.ErrorBody{Code: "INVALID_REQUEST", Message: "The request was invalid."}})
		return
	}
	requestID := requestIDFromContext(r.Context())
	category := strings.TrimSpace(r.PathValue("category"))
	assetID := strings.TrimSpace(r.PathValue("institution_asset"))
	if !regulatedFinanceAssetCategoryPattern.MatchString(category) || !regulatedFinanceAssetIDPattern.MatchString(assetID) {
		_ = response.Validation(w, requestID, []response.ValidationError{{Field: "institution_asset", Message: "Institution logo ID must be a lowercase ID ending with .png."}})
		return
	}
	institutionID := strings.TrimSuffix(assetID, ".png")
	logo, err := assets.FinancialInstitutionLogo(category, institutionID, "png")
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
