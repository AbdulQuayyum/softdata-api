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

var geographyStateIDPattern = regexp.MustCompile(`^[a-z]+(?:-[a-z]+)*$`)

// GeographyService provides state lookup operations over the geography repository.
type GeographyService struct {
	repository interfaces.GeographyRepository
}

func NewGeographyService(repository interfaces.GeographyRepository) (*GeographyService, error) {
	if repository == nil {
		return nil, fmt.Errorf("geography repository is required")
	}
	return &GeographyService{repository: repository}, nil
}

func (s *GeographyService) ListStates(ctx context.Context) ([]models.State, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	states, err := s.repository.ListStates(ctx)
	if err != nil {
		return nil, translateGeographyServiceError("list states", err)
	}
	return cloneStateList(states), nil
}

func (s *GeographyService) ListGeopoliticalZones(ctx context.Context) ([]models.GeopoliticalZone, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	zones, err := s.repository.ListGeopoliticalZones(ctx)
	if err != nil {
		return nil, translateGeographyServiceError("list geopolitical zones", err)
	}
	return cloneGeopoliticalZoneList(zones), nil
}

func (s *GeographyService) GetState(ctx context.Context, stateID string) (models.State, error) {
	if err := ctx.Err(); err != nil {
		return models.State{}, err
	}

	stateID = strings.TrimSpace(stateID)
	if stateID == "" || !geographyStateIDPattern.MatchString(stateID) {
		return models.State{}, ErrInvalidStateID
	}

	state, err := s.repository.GetStateByID(ctx, stateID)
	if err != nil {
		return models.State{}, translateGeographyServiceLookupError(err)
	}
	return state, nil
}

func (s *GeographyService) GetGeopoliticalZone(ctx context.Context, zoneID string) (models.GeopoliticalZone, error) {
	if err := ctx.Err(); err != nil {
		return models.GeopoliticalZone{}, err
	}

	zoneID = strings.TrimSpace(zoneID)
	if zoneID == "" || !geographyStateIDPattern.MatchString(zoneID) {
		return models.GeopoliticalZone{}, ErrInvalidGeopoliticalZoneID
	}

	zone, err := s.repository.GetGeopoliticalZone(ctx, zoneID)
	if err != nil {
		return models.GeopoliticalZone{}, translateGeographyZoneLookupError(err)
	}
	return zone, nil
}

func cloneStateList(states []models.State) []models.State {
	if len(states) == 0 {
		return make([]models.State, 0)
	}
	cloned := make([]models.State, len(states))
	copy(cloned, states)
	return cloned
}

func cloneGeopoliticalZoneList(zones []models.GeopoliticalZone) []models.GeopoliticalZone {
	if len(zones) == 0 {
		return make([]models.GeopoliticalZone, 0)
	}
	cloned := make([]models.GeopoliticalZone, len(zones))
	copy(cloned, zones)
	return cloned
}

func translateGeographyServiceLookupError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrStateNotFound):
		return ErrStateNotFound
	case errors.Is(err, interfaces.ErrInvalidDatasetFile):
		return fmt.Errorf("get state: repository unavailable")
	default:
		return fmt.Errorf("get state: repository unavailable")
	}
}

func translateGeographyZoneLookupError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrGeopoliticalZoneNotFound):
		return ErrGeopoliticalZoneNotFound
	case errors.Is(err, interfaces.ErrInvalidDatasetFile):
		return fmt.Errorf("get geopolitical zone: repository unavailable")
	default:
		return fmt.Errorf("get geopolitical zone: repository unavailable")
	}
}

func translateGeographyServiceError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrInvalidDatasetFile):
		return fmt.Errorf("%s: repository unavailable", op)
	default:
		return fmt.Errorf("%s: repository unavailable", op)
	}
}
