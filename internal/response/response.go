package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const jsonContentType = "application/json"

// successEnvelope is the documented success wrapper.
type successEnvelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

// paginatedEnvelope is the documented paginated success wrapper.
type paginatedEnvelope struct {
	Success bool           `json:"success"`
	Data    any            `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

// JSON writes a status code and a raw JSON-compatible payload.
func JSON(w http.ResponseWriter, status int, payload any) error {
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return fmt.Errorf("encode json response: %w", err)
	}
	return nil
}

// Success writes the documented single-resource success envelope.
func Success(w http.ResponseWriter, status int, data any) error {
	return JSON(w, status, successEnvelope{
		Success: true,
		Data:    data,
	})
}

// List writes the documented array success envelope and normalizes nil slices to [].
func List[T any](w http.ResponseWriter, status int, data []T) error {
	if data == nil {
		data = make([]T, 0)
	}
	return JSON(w, status, successEnvelope{
		Success: true,
		Data:    data,
	})
}

// Paginated writes the documented paginated success envelope.
func Paginated[T any](w http.ResponseWriter, status int, data []T, meta PaginationMeta) error {
	if err := validatePaginationMeta(meta); err != nil {
		return err
	}
	if data == nil {
		data = make([]T, 0)
	}
	return JSON(w, status, paginatedEnvelope{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// NoContent writes a status with no body.
func NoContent(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}
