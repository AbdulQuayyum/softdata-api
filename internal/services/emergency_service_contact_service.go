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
	EmergencyServiceContactDefaultPage     = 1
	EmergencyServiceContactDefaultPageSize = 50
	EmergencyServiceContactMaxPageSize     = 100
	EmergencyServiceContactMaxSearchLength = 100
)

var (
	emergencyServiceContactServiceIDPattern        = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	emergencyServiceContactServiceShortCodePattern = regexp.MustCompile(`^[0-9]{3}$`)
	emergencyServiceContactServicePhonePattern     = regexp.MustCompile(`^(?:0[0-9]{10}|0800[0-9]{8}|\+234[0-9]{10})$`)
)

type EmergencyServiceContactQuery = interfaces.EmergencyServiceContactQuery
type EmergencyServiceContactListResult = interfaces.EmergencyServiceContactListResult

type EmergencyServiceContactService struct {
	repository interfaces.EmergencyServiceContactRepository
}

func NewEmergencyServiceContactService(repository interfaces.EmergencyServiceContactRepository) (*EmergencyServiceContactService, error) {
	if repository == nil {
		return nil, fmt.Errorf("emergency service contact repository is required")
	}
	return &EmergencyServiceContactService{repository: repository}, nil
}

func (s *EmergencyServiceContactService) ListEmergencyServiceContacts(ctx context.Context, input EmergencyServiceContactQuery) (EmergencyServiceContactListResult, error) {
	if err := serviceContextError(ctx); err != nil {
		return EmergencyServiceContactListResult{}, err
	}
	query, err := normalizeEmergencyServiceContactServiceQuery(input)
	if err != nil {
		return EmergencyServiceContactListResult{}, err
	}
	result, err := s.repository.ListEmergencyServiceContacts(ctx, query)
	if err != nil {
		return EmergencyServiceContactListResult{}, translateEmergencyServiceContactError("list emergency service contacts", err)
	}
	result.Records = cloneEmergencyServiceContactList(result.Records)
	return result, nil
}

func (s *EmergencyServiceContactService) GetEmergencyServiceContact(ctx context.Context, id string) (models.EmergencyServiceContact, error) {
	if err := serviceContextError(ctx); err != nil {
		return models.EmergencyServiceContact{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" || len(id) > models.EmergencyServiceContactIDMaxLength || !emergencyServiceContactServiceIDPattern.MatchString(id) {
		return models.EmergencyServiceContact{}, ErrInvalidEmergencyServiceContactID
	}
	record, err := s.repository.GetEmergencyServiceContact(ctx, id)
	if err != nil {
		return models.EmergencyServiceContact{}, translateEmergencyServiceContactError("get emergency service contact", err)
	}
	return record, nil
}

func normalizeEmergencyServiceContactServiceQuery(input EmergencyServiceContactQuery) (EmergencyServiceContactQuery, error) {
	query := input
	if query.Page == 0 {
		query.Page = EmergencyServiceContactDefaultPage
	}
	if query.PageSize == 0 {
		query.PageSize = EmergencyServiceContactDefaultPageSize
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > EmergencyServiceContactMaxPageSize {
		return EmergencyServiceContactQuery{}, ErrInvalidEmergencyServiceContactPagination
	}
	rawServiceType, rawContactType, rawCoverageType, rawContactValue, rawSearch := query.ServiceType, query.ContactType, query.CoverageType, query.ContactValue, query.Search
	query.ServiceType = strings.TrimSpace(query.ServiceType)
	query.ContactType = strings.TrimSpace(query.ContactType)
	query.CoverageType = strings.TrimSpace(query.CoverageType)
	query.ContactValue = strings.TrimSpace(query.ContactValue)
	query.Search = strings.TrimSpace(query.Search)
	if rawServiceType != "" && query.ServiceType == "" {
		return EmergencyServiceContactQuery{}, ErrInvalidEmergencyServiceContactServiceType
	}
	if rawContactType != "" && query.ContactType == "" {
		return EmergencyServiceContactQuery{}, ErrInvalidEmergencyServiceContactContactType
	}
	if rawCoverageType != "" && query.CoverageType == "" {
		return EmergencyServiceContactQuery{}, ErrInvalidEmergencyServiceContactCoverageType
	}
	if rawContactValue != "" && query.ContactValue == "" {
		return EmergencyServiceContactQuery{}, ErrInvalidEmergencyServiceContactContactValue
	}
	if rawSearch != "" && query.Search == "" {
		return EmergencyServiceContactQuery{}, ErrInvalidEmergencyServiceContactSearch
	}
	if query.ServiceType != "" && !validEmergencyServiceContactServiceType(query.ServiceType) {
		return EmergencyServiceContactQuery{}, ErrInvalidEmergencyServiceContactServiceType
	}
	if query.ContactType != "" && !validEmergencyServiceContactType(query.ContactType) {
		return EmergencyServiceContactQuery{}, ErrInvalidEmergencyServiceContactContactType
	}
	if query.CoverageType != "" && !validEmergencyServiceCoverageType(query.CoverageType) {
		return EmergencyServiceContactQuery{}, ErrInvalidEmergencyServiceContactCoverageType
	}
	if query.ContactValue != "" && !(emergencyServiceContactServiceShortCodePattern.MatchString(query.ContactValue) || emergencyServiceContactServicePhonePattern.MatchString(query.ContactValue)) {
		return EmergencyServiceContactQuery{}, ErrInvalidEmergencyServiceContactContactValue
	}
	if len([]rune(query.Search)) > EmergencyServiceContactMaxSearchLength || strings.ContainsAny(query.Search, "\r\n\x00") {
		return EmergencyServiceContactQuery{}, ErrInvalidEmergencyServiceContactSearch
	}
	return query, nil
}

func translateEmergencyServiceContactError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrEmergencyServiceContactNotFound):
		return ErrEmergencyServiceContactNotFound
	case errors.Is(err, interfaces.ErrInvalidEmergencyServiceContactQuery):
		return ErrInvalidEmergencyServiceContactPagination
	case errors.Is(err, interfaces.ErrInvalidEmergencyServiceContactServiceTypeFilter):
		return ErrInvalidEmergencyServiceContactServiceType
	case errors.Is(err, interfaces.ErrInvalidEmergencyServiceContactContactTypeFilter):
		return ErrInvalidEmergencyServiceContactContactType
	case errors.Is(err, interfaces.ErrInvalidEmergencyServiceContactCoverageTypeFilter):
		return ErrInvalidEmergencyServiceContactCoverageType
	case errors.Is(err, interfaces.ErrInvalidEmergencyServiceContactContactValueFilter):
		return ErrInvalidEmergencyServiceContactContactValue
	case errors.Is(err, interfaces.ErrInvalidEmergencyServiceContactSearch):
		return ErrInvalidEmergencyServiceContactSearch
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return ErrInvalidEmergencyServiceContactDataset
	default:
		return fmt.Errorf("%s: repository unavailable", op)
	}
}

func cloneEmergencyServiceContactList(items []models.EmergencyServiceContact) []models.EmergencyServiceContact {
	if len(items) == 0 {
		return make([]models.EmergencyServiceContact, 0)
	}
	cloned := make([]models.EmergencyServiceContact, len(items))
	copy(cloned, items)
	return cloned
}

func validEmergencyServiceContactServiceType(value string) bool {
	switch value {
	case "general_emergency", "disaster_management", "road_emergency", "police", "fire", "ambulance", "other":
		return true
	default:
		return false
	}
}

func validEmergencyServiceContactType(value string) bool {
	return value == "short_code" || value == "telephone"
}

func validEmergencyServiceCoverageType(value string) bool {
	return value == "national" || value == "state"
}
