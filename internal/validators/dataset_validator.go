package validators

import (
	"strings"
)

// ValidateDatasetKey validates the public dataset identifier used in paths.
func ValidateDatasetKey(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ValidationErrors{Fields: []FieldError{{
			Field:   "dataset_id",
			Code:    codeRequired,
			Message: "Dataset ID is required.",
		}}}
	}
	return value, nil
}

// DatasetListQuery represents validated dataset-list query inputs.
type DatasetListQuery struct {
	Search string
	Pagination
}

// ValidateDatasetListQuery validates the documented dataset list query inputs.
func ValidateDatasetListQuery(searchValue, pageValue, limitValue string) (DatasetListQuery, error) {
	pagination, err := ValidatePagination(pageValue, limitValue)
	if err != nil {
		return DatasetListQuery{}, err
	}

	return DatasetListQuery{
		Search:     ValidateSearch(searchValue),
		Pagination: pagination,
	}, nil
}
