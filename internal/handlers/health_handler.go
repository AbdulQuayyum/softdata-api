package handlers

import (
	"net/http"

	"github.com/AbdulQuayyum/softdata-api/internal/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_ = response.JSON(w, http.StatusOK, healthResponse{
		Success: true,
		Status:  "ok",
	})
}

type healthResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"`
}
