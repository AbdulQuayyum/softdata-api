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

var collegeOfEducationIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)

// CollegeOfEducationListInput captures the canonical college filters supported by the service.
type CollegeOfEducationListInput struct {
	OwnershipType string
	StateID       string
}

func (s *EducationService) ListCollegesOfEducation(ctx context.Context, input CollegeOfEducationListInput) ([]models.CollegeOfEducation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	filter, err := normalizeCollegeOfEducationListInput(input)
	if err != nil {
		return nil, err
	}

	colleges, err := s.repository.ListCollegesOfEducation(ctx, filter)
	if err != nil {
		return nil, translateCollegeOfEducationServiceError("list colleges of education", err)
	}
	return cloneCollegeOfEducationList(colleges), nil
}

func (s *EducationService) GetCollegeOfEducation(ctx context.Context, collegeID string) (models.CollegeOfEducation, error) {
	if err := ctx.Err(); err != nil {
		return models.CollegeOfEducation{}, err
	}

	collegeID = strings.TrimSpace(collegeID)
	if collegeID == "" || !collegeOfEducationIDPattern.MatchString(collegeID) {
		return models.CollegeOfEducation{}, ErrInvalidCollegeOfEducationID
	}

	college, err := s.repository.GetCollegeOfEducation(ctx, collegeID)
	if err != nil {
		return models.CollegeOfEducation{}, translateCollegeOfEducationLookupError(err)
	}
	return college, nil
}

func normalizeCollegeOfEducationListInput(input CollegeOfEducationListInput) (interfaces.CollegeOfEducationFilter, error) {
	filter := interfaces.CollegeOfEducationFilter{}

	ownershipType := strings.TrimSpace(input.OwnershipType)
	if ownershipType != "" {
		if _, ok := allowedUniversityOwnershipTypes[ownershipType]; !ok {
			return interfaces.CollegeOfEducationFilter{}, ErrInvalidCollegeOfEducationOwnershipType
		}
		filter.OwnershipType = ownershipType
	}

	stateID := strings.TrimSpace(input.StateID)
	if stateID != "" {
		if _, ok := allowedUniversityStateIDs[stateID]; !ok {
			return interfaces.CollegeOfEducationFilter{}, ErrInvalidCollegeOfEducationStateID
		}
		filter.StateID = stateID
	}

	return filter, nil
}

func cloneCollegeOfEducationList(colleges []models.CollegeOfEducation) []models.CollegeOfEducation {
	if len(colleges) == 0 {
		return make([]models.CollegeOfEducation, 0)
	}
	cloned := make([]models.CollegeOfEducation, len(colleges))
	copy(cloned, colleges)
	return cloned
}

func translateCollegeOfEducationLookupError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrCollegeOfEducationNotFound):
		return ErrCollegeOfEducationNotFound
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("get college of education: repository unavailable")
	default:
		return fmt.Errorf("get college of education: repository unavailable")
	}
}

func translateCollegeOfEducationServiceError(op string, err error) error {
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
