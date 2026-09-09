package response

import "net/http"

// PageSizePaginationMeta describes pagination for contracts using page_size.
type PageSizePaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PaginatedPageSize writes a paginated envelope and normalizes empty data to [].
func PaginatedPageSize[T any](w http.ResponseWriter, status int, data []T, meta PageSizePaginationMeta) error {
	if err := validatePaginationMeta(PaginationMeta{Page: meta.Page, Limit: meta.PageSize, Total: meta.Total, TotalPages: meta.TotalPages}); err != nil {
		return err
	}
	if data == nil {
		data = make([]T, 0)
	}
	return JSON(w, status, struct {
		Success bool                   `json:"success"`
		Data    []T                    `json:"data"`
		Meta    PageSizePaginationMeta `json:"meta"`
	}{true, data, meta})
}
