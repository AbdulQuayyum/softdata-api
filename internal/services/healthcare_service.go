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
	HealthFacilityDefaultPage     = 1
	HealthFacilityDefaultPageSize = 50
	HealthFacilityMaxPageSize     = 100
	HealthFacilityMaxSearchLength = 100
)

var healthFacilityServiceIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var healthFacilityServiceStates = map[string]struct{}{
	"abia": {}, "adamawa": {}, "akwa-ibom": {}, "anambra": {}, "bauchi": {}, "bayelsa": {},
	"benue": {}, "borno": {}, "cross-river": {}, "delta": {}, "ebonyi": {}, "edo": {},
	"ekiti": {}, "enugu": {}, "fct": {}, "gombe": {}, "imo": {}, "jigawa": {}, "kaduna": {},
	"kano": {}, "katsina": {}, "kebbi": {}, "kogi": {}, "kwara": {}, "lagos": {}, "nasarawa": {},
	"niger": {}, "ogun": {}, "ondo": {}, "osun": {}, "oyo": {}, "plateau": {}, "rivers": {},
	"sokoto": {}, "taraba": {}, "yobe": {}, "zamfara": {},
}

var healthFacilityServiceTypes = map[string]struct{}{
	"clinic": {}, "general-hospital": {}, "health-post": {}, "other": {},
	"primary-health-centre": {}, "specialist-hospital": {}, "teaching-hospital": {},
}

var healthFacilityServiceLevels = map[string]struct{}{"primary": {}, "secondary": {}, "tertiary": {}}

var healthFacilityServiceOwnerships = map[string]struct{}{
	"federal": {}, "local-government": {}, "military": {}, "other-public": {}, "private": {}, "state": {},
}

type HealthFacilityQuery = interfaces.HealthFacilityQuery
type HealthFacilityListResult = interfaces.HealthFacilityListResult

// HealthFacilityService provides validated use cases for the health snapshot.
type HealthFacilityService struct {
	repository interfaces.HealthFacilityRepository
}

func NewHealthFacilityService(repository interfaces.HealthFacilityRepository) (*HealthFacilityService, error) {
	if repository == nil {
		return nil, fmt.Errorf("health facility repository is required")
	}
	return &HealthFacilityService{repository: repository}, nil
}

func (s *HealthFacilityService) ListHealthFacilities(ctx context.Context, input HealthFacilityQuery) (HealthFacilityListResult, error) {
	if err := serviceContextError(ctx); err != nil {
		return HealthFacilityListResult{}, err
	}
	query, err := normalizeHealthFacilityServiceQuery(input)
	if err != nil {
		return HealthFacilityListResult{}, err
	}
	result, err := s.repository.ListHealthFacilities(ctx, query)
	if err != nil {
		return HealthFacilityListResult{}, translateHealthFacilityError("list health facilities", err)
	}
	result.Facilities = cloneHealthFacilityList(result.Facilities)
	return result, nil
}

func (s *HealthFacilityService) GetHealthFacility(ctx context.Context, id string) (models.HealthFacility, error) {
	if err := serviceContextError(ctx); err != nil {
		return models.HealthFacility{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" || !healthFacilityServiceIDPattern.MatchString(id) {
		return models.HealthFacility{}, ErrInvalidHealthFacilityID
	}
	facility, err := s.repository.GetHealthFacility(ctx, id)
	if err != nil {
		return models.HealthFacility{}, translateHealthFacilityError("get health facility", err)
	}
	return cloneHealthFacility(facility), nil
}

func normalizeHealthFacilityServiceQuery(input HealthFacilityQuery) (HealthFacilityQuery, error) {
	query := input
	if query.Page == 0 {
		query.Page = HealthFacilityDefaultPage
	}
	if query.PageSize == 0 {
		query.PageSize = HealthFacilityDefaultPageSize
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > HealthFacilityMaxPageSize {
		return HealthFacilityQuery{}, ErrInvalidHealthFacilityPagination
	}
	query.StateID = strings.TrimSpace(query.StateID)
	query.LGAID = strings.TrimSpace(query.LGAID)
	query.FacilityType = strings.TrimSpace(query.FacilityType)
	query.FacilityLevel = strings.TrimSpace(query.FacilityLevel)
	query.OwnershipType = strings.TrimSpace(query.OwnershipType)
	query.Search = strings.TrimSpace(query.Search)
	if query.StateID != "" && (!healthFacilityServiceIDPattern.MatchString(query.StateID) || !containsHealthFacility(healthFacilityServiceStates, query.StateID)) {
		return HealthFacilityQuery{}, ErrInvalidHealthFacilityStateID
	}
	if query.LGAID != "" && !healthFacilityServiceIDPattern.MatchString(query.LGAID) {
		return HealthFacilityQuery{}, ErrInvalidHealthFacilityLGAID
	}
	if query.FacilityType != "" && !containsHealthFacility(healthFacilityServiceTypes, query.FacilityType) {
		return HealthFacilityQuery{}, ErrInvalidHealthFacilityType
	}
	if query.FacilityLevel != "" && !containsHealthFacility(healthFacilityServiceLevels, query.FacilityLevel) {
		return HealthFacilityQuery{}, ErrInvalidHealthFacilityLevel
	}
	if query.OwnershipType != "" && !containsHealthFacility(healthFacilityServiceOwnerships, query.OwnershipType) {
		return HealthFacilityQuery{}, ErrInvalidHealthFacilityOwnership
	}
	if len([]rune(query.Search)) > HealthFacilityMaxSearchLength || strings.ContainsAny(query.Search, "\r\n\x00") {
		return HealthFacilityQuery{}, ErrInvalidHealthFacilitySearch
	}
	return query, nil
}

func translateHealthFacilityError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrHealthFacilityNotFound):
		return ErrHealthFacilityNotFound
	case errors.Is(err, interfaces.ErrInvalidHealthFacilityQuery):
		return ErrInvalidHealthFacilityPagination
	case errors.Is(err, interfaces.ErrInvalidHealthFacilityStateFilter):
		return ErrInvalidHealthFacilityStateID
	case errors.Is(err, interfaces.ErrInvalidHealthFacilityLGAFilter):
		return ErrInvalidHealthFacilityLGAID
	case errors.Is(err, interfaces.ErrInvalidHealthFacilityStateLGA):
		return ErrInvalidHealthFacilityStateLGA
	case errors.Is(err, interfaces.ErrInvalidHealthFacilityType):
		return ErrInvalidHealthFacilityType
	case errors.Is(err, interfaces.ErrInvalidHealthFacilityLevel):
		return ErrInvalidHealthFacilityLevel
	case errors.Is(err, interfaces.ErrInvalidHealthFacilityOwnership):
		return ErrInvalidHealthFacilityOwnership
	case errors.Is(err, interfaces.ErrInvalidHealthFacilitySearch):
		return ErrInvalidHealthFacilitySearch
	default:
		return fmt.Errorf("%s: repository unavailable", op)
	}
}

func containsHealthFacility(values map[string]struct{}, value string) bool {
	_, ok := values[value]
	return ok
}

func cloneHealthFacilityList(items []models.HealthFacility) []models.HealthFacility {
	if len(items) == 0 {
		return make([]models.HealthFacility, 0)
	}
	cloned := make([]models.HealthFacility, len(items))
	for i, item := range items {
		cloned[i] = cloneHealthFacility(item)
	}
	return cloned
}

func cloneHealthFacility(item models.HealthFacility) models.HealthFacility {
	if item.Latitude != nil {
		value := *item.Latitude
		item.Latitude = &value
	}
	if item.Longitude != nil {
		value := *item.Longitude
		item.Longitude = &value
	}
	return item
}

func serviceContextError(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}
