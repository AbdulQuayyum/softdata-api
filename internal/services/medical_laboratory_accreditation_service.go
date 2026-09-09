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
	MedicalLaboratoryAccreditationDefaultPage     = 1
	MedicalLaboratoryAccreditationDefaultPageSize = 50
	MedicalLaboratoryAccreditationMaxPageSize     = 100
	MedicalLaboratoryAccreditationMaxSearchLength = 100
)

var medicalLaboratoryAccreditationServiceIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var medicalLaboratoryAccreditationServiceStates = map[string]struct{}{
	"abia": {}, "adamawa": {}, "akwa-ibom": {}, "anambra": {}, "bauchi": {}, "bayelsa": {},
	"benue": {}, "borno": {}, "cross-river": {}, "delta": {}, "ebonyi": {}, "edo": {},
	"ekiti": {}, "enugu": {}, "fct": {}, "gombe": {}, "imo": {}, "jigawa": {}, "kaduna": {},
	"kano": {}, "katsina": {}, "kebbi": {}, "kogi": {}, "kwara": {}, "lagos": {}, "nasarawa": {},
	"niger": {}, "ogun": {}, "ondo": {}, "osun": {}, "oyo": {}, "plateau": {}, "rivers": {},
	"sokoto": {}, "taraba": {}, "yobe": {}, "zamfara": {},
}

var medicalLaboratoryAccreditationServiceStatuses = map[string]struct{}{"accredited": {}, "expired": {}}

type MedicalLaboratoryAccreditationQuery = interfaces.MedicalLaboratoryAccreditationQuery
type MedicalLaboratoryAccreditationListResult = interfaces.MedicalLaboratoryAccreditationListResult

// MedicalLaboratoryAccreditationService provides validated use cases for the MLSCN accreditation snapshot.
type MedicalLaboratoryAccreditationService struct {
	repository interfaces.MedicalLaboratoryAccreditationRepository
}

func NewMedicalLaboratoryAccreditationService(repository interfaces.MedicalLaboratoryAccreditationRepository) (*MedicalLaboratoryAccreditationService, error) {
	if repository == nil {
		return nil, fmt.Errorf("medical laboratory accreditation repository is required")
	}
	return &MedicalLaboratoryAccreditationService{repository: repository}, nil
}

func (s *MedicalLaboratoryAccreditationService) ListMedicalLaboratoryAccreditations(ctx context.Context, input MedicalLaboratoryAccreditationQuery) (MedicalLaboratoryAccreditationListResult, error) {
	if err := serviceContextError(ctx); err != nil {
		return MedicalLaboratoryAccreditationListResult{}, err
	}
	query, err := normalizeMedicalLaboratoryAccreditationServiceQuery(input)
	if err != nil {
		return MedicalLaboratoryAccreditationListResult{}, err
	}
	result, err := s.repository.ListMedicalLaboratoryAccreditations(ctx, query)
	if err != nil {
		return MedicalLaboratoryAccreditationListResult{}, translateMedicalLaboratoryAccreditationError("list medical laboratory accreditations", err)
	}
	result.Records = cloneMedicalLaboratoryAccreditationList(result.Records)
	return result, nil
}

func (s *MedicalLaboratoryAccreditationService) GetMedicalLaboratoryAccreditation(ctx context.Context, id string) (models.MedicalLaboratoryAccreditation, error) {
	if err := serviceContextError(ctx); err != nil {
		return models.MedicalLaboratoryAccreditation{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" || len(id) > models.MedicalLaboratoryAccreditationIDMaxLength || !medicalLaboratoryAccreditationServiceIDPattern.MatchString(id) {
		return models.MedicalLaboratoryAccreditation{}, ErrInvalidMedicalLaboratoryAccreditationID
	}
	record, err := s.repository.GetMedicalLaboratoryAccreditation(ctx, id)
	if err != nil {
		return models.MedicalLaboratoryAccreditation{}, translateMedicalLaboratoryAccreditationError("get medical laboratory accreditation", err)
	}
	return record, nil
}

func normalizeMedicalLaboratoryAccreditationServiceQuery(input MedicalLaboratoryAccreditationQuery) (MedicalLaboratoryAccreditationQuery, error) {
	query := input
	if query.Page == 0 {
		query.Page = MedicalLaboratoryAccreditationDefaultPage
	}
	if query.PageSize == 0 {
		query.PageSize = MedicalLaboratoryAccreditationDefaultPageSize
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > MedicalLaboratoryAccreditationMaxPageSize {
		return MedicalLaboratoryAccreditationQuery{}, ErrInvalidMedicalLaboratoryAccreditationPagination
	}
	query.StateID = strings.TrimSpace(query.StateID)
	query.AccreditationStatus = strings.TrimSpace(query.AccreditationStatus)
	query.Search = strings.TrimSpace(query.Search)
	if query.StateID != "" && (!medicalLaboratoryAccreditationServiceIDPattern.MatchString(query.StateID) || !containsMedicalLaboratoryAccreditation(medicalLaboratoryAccreditationServiceStates, query.StateID)) {
		return MedicalLaboratoryAccreditationQuery{}, ErrInvalidMedicalLaboratoryAccreditationStateID
	}
	if query.AccreditationStatus != "" && !containsMedicalLaboratoryAccreditation(medicalLaboratoryAccreditationServiceStatuses, query.AccreditationStatus) {
		return MedicalLaboratoryAccreditationQuery{}, ErrInvalidMedicalLaboratoryAccreditationStatus
	}
	if len([]rune(query.Search)) > MedicalLaboratoryAccreditationMaxSearchLength || strings.ContainsAny(query.Search, "\r\n\x00") {
		return MedicalLaboratoryAccreditationQuery{}, ErrInvalidMedicalLaboratoryAccreditationSearch
	}
	return query, nil
}

func translateMedicalLaboratoryAccreditationError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrMedicalLaboratoryAccreditationNotFound):
		return ErrMedicalLaboratoryAccreditationNotFound
	case errors.Is(err, interfaces.ErrInvalidMedicalLaboratoryAccreditationQuery):
		return ErrInvalidMedicalLaboratoryAccreditationPagination
	case errors.Is(err, interfaces.ErrInvalidMedicalLaboratoryAccreditationStateFilter):
		return ErrInvalidMedicalLaboratoryAccreditationStateID
	case errors.Is(err, interfaces.ErrInvalidMedicalLaboratoryAccreditationStatusFilter):
		return ErrInvalidMedicalLaboratoryAccreditationStatus
	case errors.Is(err, interfaces.ErrInvalidMedicalLaboratoryAccreditationSearch):
		return ErrInvalidMedicalLaboratoryAccreditationSearch
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return ErrInvalidMedicalLaboratoryAccreditationDataset
	default:
		return fmt.Errorf("%s: repository unavailable", op)
	}
}

func containsMedicalLaboratoryAccreditation(values map[string]struct{}, value string) bool {
	_, ok := values[value]
	return ok
}

func cloneMedicalLaboratoryAccreditationList(items []models.MedicalLaboratoryAccreditation) []models.MedicalLaboratoryAccreditation {
	if len(items) == 0 {
		return make([]models.MedicalLaboratoryAccreditation, 0)
	}
	cloned := make([]models.MedicalLaboratoryAccreditation, len(items))
	copy(cloned, items)
	return cloned
}
