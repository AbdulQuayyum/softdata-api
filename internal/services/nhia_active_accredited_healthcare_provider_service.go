package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const (
	NHIAActiveAccreditedHealthcareProviderDefaultPage     = 1
	NHIAActiveAccreditedHealthcareProviderDefaultPageSize = 50
	NHIAActiveAccreditedHealthcareProviderMaxPageSize     = 100
	NHIAActiveAccreditedHealthcareProviderMaxSearchLength = 100
)

var (
	nhiaHCPServiceIDPattern           = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	nhiaHCPServiceProviderCodePattern = regexp.MustCompile(`^(?:[A-Z]{2,3})/[0-9]{4}/P$`)
)

type NHIAActiveAccreditedHealthcareProviderQuery = interfaces.NHIAActiveAccreditedHealthcareProviderQuery
type NHIAActiveAccreditedHealthcareProviderListResult = interfaces.NHIAActiveAccreditedHealthcareProviderListResult

type NHIAActiveAccreditedHealthcareProviderService struct {
	repository interfaces.NHIAActiveAccreditedHealthcareProviderRepository
}

func NewNHIAActiveAccreditedHealthcareProviderService(repository interfaces.NHIAActiveAccreditedHealthcareProviderRepository) (*NHIAActiveAccreditedHealthcareProviderService, error) {
	if repository == nil {
		return nil, fmt.Errorf("nhia active accredited healthcare provider repository is required")
	}
	return &NHIAActiveAccreditedHealthcareProviderService{repository: repository}, nil
}

func (s *NHIAActiveAccreditedHealthcareProviderService) ListNHIAActiveAccreditedHealthcareProviders(ctx context.Context, input NHIAActiveAccreditedHealthcareProviderQuery) (NHIAActiveAccreditedHealthcareProviderListResult, error) {
	if err := serviceContextError(ctx); err != nil {
		return NHIAActiveAccreditedHealthcareProviderListResult{}, err
	}
	query, err := normalizeNHIAHCPServiceQuery(input)
	if err != nil {
		return NHIAActiveAccreditedHealthcareProviderListResult{}, err
	}
	result, err := s.repository.ListNHIAActiveAccreditedHealthcareProviders(ctx, query)
	if err != nil {
		return NHIAActiveAccreditedHealthcareProviderListResult{}, translateNHIAHCPError("list NHIA active accredited healthcare providers", err)
	}
	result.Records = cloneNHIAHCPList(result.Records)
	return result, nil
}

func (s *NHIAActiveAccreditedHealthcareProviderService) GetNHIAActiveAccreditedHealthcareProvider(ctx context.Context, id string) (models.NHIAActiveAccreditedHealthcareProvider, error) {
	if err := serviceContextError(ctx); err != nil {
		return models.NHIAActiveAccreditedHealthcareProvider{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" || len(id) > models.NHIAActiveAccreditedHealthcareProviderIDMaxLength || !nhiaHCPServiceIDPattern.MatchString(id) {
		return models.NHIAActiveAccreditedHealthcareProvider{}, ErrInvalidNHIAActiveAccreditedHealthcareProviderID
	}
	record, err := s.repository.GetNHIAActiveAccreditedHealthcareProvider(ctx, id)
	if err != nil {
		return models.NHIAActiveAccreditedHealthcareProvider{}, translateNHIAHCPError("get NHIA active accredited healthcare provider", err)
	}
	return record, nil
}

func normalizeNHIAHCPServiceQuery(input NHIAActiveAccreditedHealthcareProviderQuery) (NHIAActiveAccreditedHealthcareProviderQuery, error) {
	query := input
	if query.Page == 0 {
		query.Page = NHIAActiveAccreditedHealthcareProviderDefaultPage
	}
	if query.PageSize == 0 {
		query.PageSize = NHIAActiveAccreditedHealthcareProviderDefaultPageSize
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > NHIAActiveAccreditedHealthcareProviderMaxPageSize {
		return NHIAActiveAccreditedHealthcareProviderQuery{}, ErrInvalidNHIAActiveAccreditedHealthcareProviderPagination
	}
	rawCode, rawType, rawStatus, rawSearch := query.ProviderCode, query.FacilityType, query.ListingStatus, query.Search
	query.ProviderCode = strings.TrimSpace(query.ProviderCode)
	query.FacilityType = strings.TrimSpace(query.FacilityType)
	query.ListingStatus = strings.TrimSpace(query.ListingStatus)
	query.Search = strings.TrimSpace(query.Search)
	if rawCode != "" && query.ProviderCode == "" {
		return NHIAActiveAccreditedHealthcareProviderQuery{}, ErrInvalidNHIAActiveAccreditedHealthcareProviderCode
	}
	if rawType != "" && query.FacilityType == "" {
		return NHIAActiveAccreditedHealthcareProviderQuery{}, ErrInvalidNHIAActiveAccreditedHealthcareProviderFacilityType
	}
	if rawStatus != "" && query.ListingStatus == "" {
		return NHIAActiveAccreditedHealthcareProviderQuery{}, ErrInvalidNHIAActiveAccreditedHealthcareProviderListingStatus
	}
	if rawSearch != "" && query.Search == "" {
		return NHIAActiveAccreditedHealthcareProviderQuery{}, ErrInvalidNHIAActiveAccreditedHealthcareProviderSearch
	}
	if query.ProviderCode != "" && !nhiaHCPServiceProviderCodePattern.MatchString(query.ProviderCode) {
		return NHIAActiveAccreditedHealthcareProviderQuery{}, ErrInvalidNHIAActiveAccreditedHealthcareProviderCode
	}
	if query.FacilityType != "" && query.FacilityType != "primary" && query.FacilityType != "primary_and_secondary" {
		return NHIAActiveAccreditedHealthcareProviderQuery{}, ErrInvalidNHIAActiveAccreditedHealthcareProviderFacilityType
	}
	if query.ListingStatus != "" && query.ListingStatus != "active_accredited" {
		return NHIAActiveAccreditedHealthcareProviderQuery{}, ErrInvalidNHIAActiveAccreditedHealthcareProviderListingStatus
	}
	if len([]rune(query.Search)) > NHIAActiveAccreditedHealthcareProviderMaxSearchLength || strings.ContainsAny(query.Search, "\r\n\x00") {
		return NHIAActiveAccreditedHealthcareProviderQuery{}, ErrInvalidNHIAActiveAccreditedHealthcareProviderSearch
	}
	return query, nil
}

func translateNHIAHCPError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrNHIAActiveAccreditedHealthcareProviderNotFound):
		return ErrNHIAActiveAccreditedHealthcareProviderNotFound
	case errors.Is(err, interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderQuery):
		return ErrInvalidNHIAActiveAccreditedHealthcareProviderPagination
	case errors.Is(err, interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderCodeFilter):
		return ErrInvalidNHIAActiveAccreditedHealthcareProviderCode
	case errors.Is(err, interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderFacilityTypeFilter):
		return ErrInvalidNHIAActiveAccreditedHealthcareProviderFacilityType
	case errors.Is(err, interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderListingStatusFilter):
		return ErrInvalidNHIAActiveAccreditedHealthcareProviderListingStatus
	case errors.Is(err, interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderSearch):
		return ErrInvalidNHIAActiveAccreditedHealthcareProviderSearch
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return ErrInvalidNHIAActiveAccreditedHealthcareProviderDataset
	default:
		return fmt.Errorf("%s: repository unavailable", op)
	}
}

func cloneNHIAHCPList(items []models.NHIAActiveAccreditedHealthcareProvider) []models.NHIAActiveAccreditedHealthcareProvider {
	if len(items) == 0 {
		return make([]models.NHIAActiveAccreditedHealthcareProvider, 0)
	}
	cloned := make([]models.NHIAActiveAccreditedHealthcareProvider, len(items))
	copy(cloned, items)
	return cloned
}
