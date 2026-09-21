package file

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const (
	emergencyServiceContactRecordCount = 5
	emergencyServiceContactDefaultPage = 1
	emergencyServiceContactDefaultSize = 50
	emergencyServiceContactMaxPageSize = 100
	emergencyServiceContactMaxSearch   = 100
)

var (
	emergencyServiceContactIDPattern        = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	emergencyServiceContactShortCodePattern = regexp.MustCompile(`^[0-9]{3}$`)
	emergencyServiceContactPhonePattern     = regexp.MustCompile(`^(?:0[0-9]{10}|0800[0-9]{8}|\+234[0-9]{10})$`)
)

type EmergencyServiceContactFileRepository struct {
	jsonRepository interfaces.JSONFileRepository
	recordsPath    string
	cache          lazyDatasetCache[emergencyServiceContactDataset]
}

type emergencyServiceContactDataset struct {
	records        []models.EmergencyServiceContact
	all            []int
	searchText     []string
	byID           map[string]int
	byServiceType  map[string][]int
	byContactType  map[string][]int
	byCoverageType map[string][]int
	byContactValue map[string][]int
}

var _ interfaces.EmergencyServiceContactRepository = (*EmergencyServiceContactFileRepository)(nil)

func NewEmergencyServiceContactRepository(jsonRepository interfaces.JSONFileRepository, recordsPath string) (*EmergencyServiceContactFileRepository, error) {
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	cleanRecordsPath, err := validateGeographyDatasetPath("emergency service contacts", recordsPath)
	if err != nil {
		return nil, err
	}
	return &EmergencyServiceContactFileRepository{jsonRepository: jsonRepository, recordsPath: cleanRecordsPath}, nil
}

func (r *EmergencyServiceContactFileRepository) ListEmergencyServiceContacts(ctx context.Context, query interfaces.EmergencyServiceContactQuery) (interfaces.EmergencyServiceContactListResult, error) {
	dataset, err := r.load(ctx)
	if err != nil {
		return interfaces.EmergencyServiceContactListResult{}, err
	}
	query, err = normalizeEmergencyServiceContactQuery(query)
	if err != nil {
		return interfaces.EmergencyServiceContactListResult{}, err
	}
	if err := contextError(ctx); err != nil {
		return interfaces.EmergencyServiceContactListResult{}, err
	}

	candidates := dataset.all
	if query.ContactValue != "" {
		candidates = dataset.byContactValue[query.ContactValue]
	} else if query.ServiceType != "" {
		candidates = dataset.byServiceType[query.ServiceType]
	} else if query.ContactType != "" {
		candidates = dataset.byContactType[query.ContactType]
	} else if query.CoverageType != "" {
		candidates = dataset.byCoverageType[query.CoverageType]
	}

	matching := make([]int, 0, len(candidates))
	search := strings.ToLower(query.Search)
	for _, index := range candidates {
		if err := contextError(ctx); err != nil {
			return interfaces.EmergencyServiceContactListResult{}, err
		}
		record := dataset.records[index]
		if query.ServiceType != "" && record.ServiceType != query.ServiceType {
			continue
		}
		if query.ContactType != "" && record.ContactType != query.ContactType {
			continue
		}
		if query.CoverageType != "" && record.CoverageType != query.CoverageType {
			continue
		}
		if query.ContactValue != "" && record.ContactValue != query.ContactValue {
			continue
		}
		if search != "" && !strings.Contains(dataset.searchText[index], search) {
			continue
		}
		matching = append(matching, index)
	}

	total := len(matching)
	totalPages := 0
	if total > 0 {
		totalPages = (total + query.PageSize - 1) / query.PageSize
	}
	start := 0
	if query.Page > 1 && query.Page-1 > math.MaxInt/query.PageSize {
		start = total
	} else {
		start = (query.Page - 1) * query.PageSize
	}
	if start >= total {
		return interfaces.EmergencyServiceContactListResult{Records: make([]models.EmergencyServiceContact, 0), Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages}, nil
	}
	end := start + query.PageSize
	if end > total {
		end = total
	}
	page := make([]models.EmergencyServiceContact, end-start)
	for i, index := range matching[start:end] {
		page[i] = dataset.records[index]
	}
	return interfaces.EmergencyServiceContactListResult{Records: page, Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages}, nil
}

func (r *EmergencyServiceContactFileRepository) GetEmergencyServiceContact(ctx context.Context, id string) (models.EmergencyServiceContact, error) {
	dataset, err := r.load(ctx)
	if err != nil {
		return models.EmergencyServiceContact{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" || len(id) > models.EmergencyServiceContactIDMaxLength || !emergencyServiceContactIDPattern.MatchString(id) {
		return models.EmergencyServiceContact{}, fmt.Errorf("%w", interfaces.ErrInvalidEmergencyServiceContactQuery)
	}
	if err := contextError(ctx); err != nil {
		return models.EmergencyServiceContact{}, err
	}
	index, ok := dataset.byID[id]
	if !ok {
		return models.EmergencyServiceContact{}, fmt.Errorf("%w", interfaces.ErrEmergencyServiceContactNotFound)
	}
	return dataset.records[index], nil
}

func (r *EmergencyServiceContactFileRepository) load(ctx context.Context) (emergencyServiceContactDataset, error) {
	values, err := r.cache.get(ctx, func(ctx context.Context) ([]emergencyServiceContactDataset, error) {
		dataset, err := r.decodeAndIndex(ctx)
		if err != nil {
			return nil, err
		}
		return []emergencyServiceContactDataset{dataset}, nil
	}, func(value []emergencyServiceContactDataset) []emergencyServiceContactDataset { return value })
	if err != nil {
		return emergencyServiceContactDataset{}, err
	}
	return values[0], nil
}

func (r *EmergencyServiceContactFileRepository) decodeAndIndex(ctx context.Context) (emergencyServiceContactDataset, error) {
	var records []models.EmergencyServiceContact
	if err := r.jsonRepository.Decode(ctx, r.recordsPath, &records); err != nil {
		return emergencyServiceContactDataset{}, sanitizeEmergencyServiceContactLoadError(err)
	}
	if err := validateEmergencyServiceContactRecords(ctx, records); err != nil {
		return emergencyServiceContactDataset{}, err
	}
	dataset := emergencyServiceContactDataset{
		records:        append([]models.EmergencyServiceContact(nil), records...),
		all:            make([]int, len(records)),
		searchText:     make([]string, len(records)),
		byID:           make(map[string]int, len(records)),
		byServiceType:  map[string][]int{},
		byContactType:  map[string][]int{},
		byCoverageType: map[string][]int{},
		byContactValue: map[string][]int{},
	}
	for i, record := range records {
		if err := contextError(ctx); err != nil {
			return emergencyServiceContactDataset{}, err
		}
		dataset.byID[record.ID] = i
		dataset.all[i] = i
		dataset.searchText[i] = strings.ToLower(record.ServiceName + " " + record.AgencyName)
		dataset.byServiceType[record.ServiceType] = append(dataset.byServiceType[record.ServiceType], i)
		dataset.byContactType[record.ContactType] = append(dataset.byContactType[record.ContactType], i)
		dataset.byCoverageType[record.CoverageType] = append(dataset.byCoverageType[record.CoverageType], i)
		dataset.byContactValue[record.ContactValue] = append(dataset.byContactValue[record.ContactValue], i)
	}
	return dataset, nil
}

func validateEmergencyServiceContactRecords(ctx context.Context, records []models.EmergencyServiceContact) error {
	if len(records) != emergencyServiceContactRecordCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	seenIDs := make(map[string]struct{}, len(records))
	for i, record := range records {
		if err := contextError(ctx); err != nil {
			return err
		}
		if record.ID == "" || record.ServiceName == "" || record.AgencyName == "" || record.ServiceType == "" || record.ContactType == "" || record.ContactValue == "" || record.CoverageType == "" || record.CountryCode != "NG" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		for _, value := range []string{record.ID, record.ServiceName, record.AgencyName, record.ServiceType, record.ContactType, record.ContactValue, record.CoverageType, record.CountryCode, record.StateID, record.Availability, record.CallCost, record.Notes} {
			if strings.TrimSpace(value) != value {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		if len(record.ID) > models.EmergencyServiceContactIDMaxLength || !emergencyServiceContactIDPattern.MatchString(record.ID) || !validEmergencyServiceType(record.ServiceType) || !validEmergencyContactType(record.ContactType) || !validEmergencyCoverageType(record.CoverageType) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if record.ContactType == "short_code" && !emergencyServiceContactShortCodePattern.MatchString(record.ContactValue) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if record.ContactType == "telephone" && !emergencyServiceContactPhonePattern.MatchString(record.ContactValue) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if record.CoverageType == "national" && record.StateID != "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if record.CoverageType == "state" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if record.Availability != "" && record.Availability != "24_hours" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if record.CallCost != "" && record.CallCost != "toll_free" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, duplicate := seenIDs[record.ID]; duplicate {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		seenIDs[record.ID] = struct{}{}
		if i > 0 && records[i-1].ID > record.ID {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	return nil
}

func normalizeEmergencyServiceContactQuery(query interfaces.EmergencyServiceContactQuery) (interfaces.EmergencyServiceContactQuery, error) {
	if query.Page == 0 {
		query.Page = emergencyServiceContactDefaultPage
	}
	if query.PageSize == 0 {
		query.PageSize = emergencyServiceContactDefaultSize
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > emergencyServiceContactMaxPageSize {
		return interfaces.EmergencyServiceContactQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidEmergencyServiceContactQuery)
	}
	rawServiceType, rawContactType, rawCoverageType, rawContactValue, rawSearch := query.ServiceType, query.ContactType, query.CoverageType, query.ContactValue, query.Search
	query.ServiceType = strings.TrimSpace(query.ServiceType)
	query.ContactType = strings.TrimSpace(query.ContactType)
	query.CoverageType = strings.TrimSpace(query.CoverageType)
	query.ContactValue = strings.TrimSpace(query.ContactValue)
	query.Search = strings.TrimSpace(query.Search)
	if rawServiceType != "" && query.ServiceType == "" {
		return interfaces.EmergencyServiceContactQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidEmergencyServiceContactServiceTypeFilter)
	}
	if rawContactType != "" && query.ContactType == "" {
		return interfaces.EmergencyServiceContactQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidEmergencyServiceContactContactTypeFilter)
	}
	if rawCoverageType != "" && query.CoverageType == "" {
		return interfaces.EmergencyServiceContactQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidEmergencyServiceContactCoverageTypeFilter)
	}
	if rawContactValue != "" && query.ContactValue == "" {
		return interfaces.EmergencyServiceContactQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidEmergencyServiceContactContactValueFilter)
	}
	if rawSearch != "" && query.Search == "" {
		return interfaces.EmergencyServiceContactQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidEmergencyServiceContactSearch)
	}
	if query.ServiceType != "" && !validEmergencyServiceType(query.ServiceType) {
		return interfaces.EmergencyServiceContactQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidEmergencyServiceContactServiceTypeFilter)
	}
	if query.ContactType != "" && !validEmergencyContactType(query.ContactType) {
		return interfaces.EmergencyServiceContactQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidEmergencyServiceContactContactTypeFilter)
	}
	if query.CoverageType != "" && !validEmergencyCoverageType(query.CoverageType) {
		return interfaces.EmergencyServiceContactQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidEmergencyServiceContactCoverageTypeFilter)
	}
	if query.ContactValue != "" && !(emergencyServiceContactShortCodePattern.MatchString(query.ContactValue) || emergencyServiceContactPhonePattern.MatchString(query.ContactValue)) {
		return interfaces.EmergencyServiceContactQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidEmergencyServiceContactContactValueFilter)
	}
	if len([]rune(query.Search)) > emergencyServiceContactMaxSearch || strings.ContainsAny(query.Search, "\r\n\x00") {
		return interfaces.EmergencyServiceContactQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidEmergencyServiceContactSearch)
	}
	return query, nil
}

func validEmergencyServiceType(value string) bool {
	switch value {
	case "general_emergency", "disaster_management", "road_emergency", "police", "fire", "ambulance", "other":
		return true
	default:
		return false
	}
}

func validEmergencyContactType(value string) bool {
	return value == "short_code" || value == "telephone"
}

func validEmergencyCoverageType(value string) bool {
	return value == "national" || value == "state"
}

func sanitizeEmergencyServiceContactLoadError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, interfaces.ErrDatasetFileNotFound) || errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
}
