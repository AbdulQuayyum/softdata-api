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

var universityIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)

var allowedUniversityOwnershipTypes = map[string]struct{}{
	"federal": {},
	"state":   {},
	"private": {},
}

var allowedUniversityStateIDs = map[string]struct{}{
	"abia": {}, "adamawa": {}, "akwa-ibom": {}, "anambra": {}, "bauchi": {},
	"bayelsa": {}, "benue": {}, "borno": {}, "cross-river": {}, "delta": {},
	"ebonyi": {}, "edo": {}, "ekiti": {}, "enugu": {}, "fct": {}, "gombe": {},
	"imo": {}, "jigawa": {}, "kaduna": {}, "kano": {}, "katsina": {},
	"kebbi": {}, "kogi": {}, "kwara": {}, "lagos": {}, "nasarawa": {},
	"niger": {}, "ogun": {}, "ondo": {}, "osun": {}, "oyo": {}, "plateau": {},
	"rivers": {}, "sokoto": {}, "taraba": {}, "yobe": {}, "zamfara": {},
}

// UniversityListInput captures the canonical university filters supported by the service.
type UniversityListInput struct {
	OwnershipType string
	StateID       string
}

// EducationService provides university lookups over the education repository.
type EducationService struct {
	repository interfaces.EducationRepository
}

func NewEducationService(repository interfaces.EducationRepository) (*EducationService, error) {
	if repository == nil {
		return nil, fmt.Errorf("education repository is required")
	}
	return &EducationService{repository: repository}, nil
}

func (s *EducationService) ListUniversities(ctx context.Context, input UniversityListInput) ([]models.University, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	filter, err := normalizeUniversityListInput(input)
	if err != nil {
		return nil, err
	}

	universities, err := s.repository.ListUniversities(ctx, filter)
	if err != nil {
		return nil, translateEducationServiceError("list universities", err)
	}
	return cloneUniversityList(universities), nil
}

func (s *EducationService) GetUniversity(ctx context.Context, universityID string) (models.University, error) {
	if err := ctx.Err(); err != nil {
		return models.University{}, err
	}

	universityID = strings.TrimSpace(universityID)
	if universityID == "" || !universityIDPattern.MatchString(universityID) {
		return models.University{}, ErrInvalidUniversityID
	}

	university, err := s.repository.GetUniversityByID(ctx, universityID)
	if err != nil {
		return models.University{}, translateEducationLookupError(err)
	}
	return university, nil
}

func normalizeUniversityListInput(input UniversityListInput) (interfaces.UniversityFilter, error) {
	filter := interfaces.UniversityFilter{}

	ownershipType := strings.TrimSpace(input.OwnershipType)
	if ownershipType != "" {
		if _, ok := allowedUniversityOwnershipTypes[ownershipType]; !ok {
			return interfaces.UniversityFilter{}, ErrInvalidUniversityOwnershipType
		}
		filter.OwnershipType = ownershipType
	}

	stateID := strings.TrimSpace(input.StateID)
	if stateID != "" {
		if _, ok := allowedUniversityStateIDs[stateID]; !ok {
			return interfaces.UniversityFilter{}, ErrInvalidUniversityStateID
		}
		filter.StateID = stateID
	}

	return filter, nil
}

func cloneUniversityList(universities []models.University) []models.University {
	if len(universities) == 0 {
		return make([]models.University, 0)
	}
	cloned := make([]models.University, len(universities))
	copy(cloned, universities)
	return cloned
}

func cloneUniversity(university models.University) models.University {
	return university
}

func translateEducationLookupError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrUniversityNotFound):
		return ErrUniversityNotFound
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("get university: repository unavailable")
	default:
		return fmt.Errorf("get university: repository unavailable")
	}
}

func translateEducationServiceError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("%s: repository unavailable", op)
	default:
		return fmt.Errorf("%s: repository unavailable", op)
	}
}
