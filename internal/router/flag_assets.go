package router

import (
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/datasets/assets"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
)

func serveCountryFlagSVG(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		_ = response.JSON(w, http.StatusMethodNotAllowed, response.ErrorResponse{
			Success: false,
			Error: response.ErrorBody{
				Code:    "INVALID_REQUEST",
				Message: "The request was invalid.",
			},
		})
		return
	}

	requestID := requestIDFromContext(r.Context())
	rawCountryID := strings.TrimSpace(r.PathValue("country_id"))
	if rawCountryID == "" || !strings.HasSuffix(rawCountryID, ".svg") {
		_ = response.Validation(w, requestID, []response.ValidationError{{
			Field:   "country_id",
			Message: "Country or area flag ID must end with .svg.",
		}})
		return
	}

	countryID := strings.TrimSuffix(rawCountryID, ".svg")
	if len(countryID) != 2 || countryID != strings.ToLower(countryID) || !isLowercaseASCII(countryID) {
		_ = response.Validation(w, requestID, []response.ValidationError{{
			Field:   "country_id",
			Message: "Country or area flag ID must be a lowercase alpha-2 code ending with .svg.",
		}})
		return
	}

	svg, err := assets.FlagSVG(countryID)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			_ = response.Error(w, interfaces.ErrNotFound, requestID)
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(svg)
}

func isLowercaseASCII(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] < 'a' || value[i] > 'z' {
			return false
		}
	}
	return true
}
