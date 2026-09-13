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
	NHIAAccreditedHealthMaintenanceOrganisationDefaultPage     = 1
	NHIAAccreditedHealthMaintenanceOrganisationDefaultPageSize = 50
	NHIAAccreditedHealthMaintenanceOrganisationMaxPageSize     = 100
	NHIAAccreditedHealthMaintenanceOrganisationMaxSearchLength = 100
)

var nhiaAccreditedHMOServiceIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type NHIAAccreditedHealthMaintenanceOrganisationQuery = interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery
type NHIAAccreditedHealthMaintenanceOrganisationListResult = interfaces.NHIAAccreditedHealthMaintenanceOrganisationListResult

// NHIAAccreditedHealthMaintenanceOrganisationService provides validated use cases for the NHIA accredited HMO snapshot.
type NHIAAccreditedHealthMaintenanceOrganisationService struct {
	repository interfaces.NHIAAccreditedHealthMaintenanceOrganisationRepository
}

func NewNHIAAccreditedHealthMaintenanceOrganisationService(repository interfaces.NHIAAccreditedHealthMaintenanceOrganisationRepository) (*NHIAAccreditedHealthMaintenanceOrganisationService, error) {
	if repository == nil {
		return nil, fmt.Errorf("nhia accredited health maintenance organisation repository is required")
	}
	return &NHIAAccreditedHealthMaintenanceOrganisationService{repository: repository}, nil
}

func (s *NHIAAccreditedHealthMaintenanceOrganisationService) ListNHIAAccreditedHealthMaintenanceOrganisations(ctx context.Context, input NHIAAccreditedHealthMaintenanceOrganisationQuery) (NHIAAccreditedHealthMaintenanceOrganisationListResult, error) {
	if err := serviceContextError(ctx); err != nil {
		return NHIAAccreditedHealthMaintenanceOrganisationListResult{}, err
	}
	query, err := normalizeNHIAAccreditedHMOServiceQuery(input)
	if err != nil {
		return NHIAAccreditedHealthMaintenanceOrganisationListResult{}, err
	}
	result, err := s.repository.ListNHIAAccreditedHealthMaintenanceOrganisations(ctx, query)
	if err != nil {
		return NHIAAccreditedHealthMaintenanceOrganisationListResult{}, translateNHIAAccreditedHMOError("list NHIA accredited health maintenance organisations", err)
	}
	result.Records = cloneNHIAAccreditedHMOList(result.Records)
	return result, nil
}

func (s *NHIAAccreditedHealthMaintenanceOrganisationService) GetNHIAAccreditedHealthMaintenanceOrganisation(ctx context.Context, id string) (models.NHIAAccreditedHealthMaintenanceOrganisation, error) {
	if err := serviceContextError(ctx); err != nil {
		return models.NHIAAccreditedHealthMaintenanceOrganisation{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" || len(id) > models.NHIAAccreditedHealthMaintenanceOrganisationIDMaxLength || !nhiaAccreditedHMOServiceIDPattern.MatchString(id) {
		return models.NHIAAccreditedHealthMaintenanceOrganisation{}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationID
	}
	record, err := s.repository.GetNHIAAccreditedHealthMaintenanceOrganisation(ctx, id)
	if err != nil {
		return models.NHIAAccreditedHealthMaintenanceOrganisation{}, translateNHIAAccreditedHMOError("get NHIA accredited health maintenance organisation", err)
	}
	return record, nil
}

func normalizeNHIAAccreditedHMOServiceQuery(input NHIAAccreditedHealthMaintenanceOrganisationQuery) (NHIAAccreditedHealthMaintenanceOrganisationQuery, error) {
	query := input
	if query.Page == 0 {
		query.Page = NHIAAccreditedHealthMaintenanceOrganisationDefaultPage
	}
	if query.PageSize == 0 {
		query.PageSize = NHIAAccreditedHealthMaintenanceOrganisationDefaultPageSize
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > NHIAAccreditedHealthMaintenanceOrganisationMaxPageSize {
		return NHIAAccreditedHealthMaintenanceOrganisationQuery{}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationPagination
	}
	rawStatus := query.AccreditationStatus
	rawHMOID := query.HMOID
	rawSearch := query.Search
	query.AccreditationStatus = strings.TrimSpace(query.AccreditationStatus)
	query.HMOID = strings.TrimSpace(query.HMOID)
	query.Search = strings.TrimSpace(query.Search)
	if rawStatus != "" && query.AccreditationStatus == "" {
		return NHIAAccreditedHealthMaintenanceOrganisationQuery{}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationStatus
	}
	if rawHMOID != "" && query.HMOID == "" {
		return NHIAAccreditedHealthMaintenanceOrganisationQuery{}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationHMOID
	}
	if rawSearch != "" && query.Search == "" {
		return NHIAAccreditedHealthMaintenanceOrganisationQuery{}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationSearch
	}
	if query.AccreditationStatus != "" && query.AccreditationStatus != "accredited" {
		return NHIAAccreditedHealthMaintenanceOrganisationQuery{}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationStatus
	}
	if strings.ContainsAny(query.HMOID, "\r\n\x00") {
		return NHIAAccreditedHealthMaintenanceOrganisationQuery{}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationHMOID
	}
	if len([]rune(query.Search)) > NHIAAccreditedHealthMaintenanceOrganisationMaxSearchLength || strings.ContainsAny(query.Search, "\r\n\x00") {
		return NHIAAccreditedHealthMaintenanceOrganisationQuery{}, ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationSearch
	}
	return query, nil
}

func translateNHIAAccreditedHMOError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrNHIAAccreditedHealthMaintenanceOrganisationNotFound):
		return ErrNHIAAccreditedHealthMaintenanceOrganisationNotFound
	case errors.Is(err, interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationQuery):
		return ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationPagination
	case errors.Is(err, interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationStatusFilter):
		return ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationStatus
	case errors.Is(err, interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationHMOIDFilter):
		return ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationHMOID
	case errors.Is(err, interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationSearch):
		return ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationSearch
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationDataset
	default:
		return fmt.Errorf("%s: repository unavailable", op)
	}
}

func cloneNHIAAccreditedHMOList(items []models.NHIAAccreditedHealthMaintenanceOrganisation) []models.NHIAAccreditedHealthMaintenanceOrganisation {
	if len(items) == 0 {
		return make([]models.NHIAAccreditedHealthMaintenanceOrganisation, 0)
	}
	cloned := make([]models.NHIAAccreditedHealthMaintenanceOrganisation, len(items))
	copy(cloned, items)
	return cloned
}
