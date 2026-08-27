package handlers

import (
	"net/http"

	"github.com/AbdulQuayyum/softdata-api/internal/response"
)

type DiscoveryHandler struct{}

func NewDiscoveryHandler() *DiscoveryHandler {
	return &DiscoveryHandler{}
}

func (h *DiscoveryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_ = response.JSON(w, http.StatusOK, apiInfoResponse{
		Success: true,
		Data:    apiInfoData{},
	})
}

type apiInfoResponse struct {
	Success bool        `json:"success"`
	Data    apiInfoData `json:"data"`
}

type apiInfoData struct{}
