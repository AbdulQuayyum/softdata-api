package file

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const (
	healthFacilityRecordCount  = 50654
	healthFacilityDefaultPage  = 1
	healthFacilityDefaultSize  = 50
	healthFacilityMaxPageSize  = 100
	healthFacilityMaxSearch    = 100
	healthFacilityMinLatitude  = 4.281710
	healthFacilityMaxLatitude  = 13.865239
	healthFacilityMinLongitude = 2.707790
	healthFacilityMaxLongitude = 14.636383
)

var healthFacilityIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var healthFacilityTypes = map[string]struct{}{
	"clinic": {}, "general-hospital": {}, "health-post": {},
	"other": {}, "primary-health-centre": {}, "specialist-hospital": {},
	"teaching-hospital": {},
}

var healthFacilityLevels = map[string]struct{}{
	"primary": {}, "secondary": {}, "tertiary": {},
}

var healthFacilityOwnerships = map[string]struct{}{
	"federal": {}, "local-government": {}, "military": {}, "other-public": {},
	"private": {}, "state": {},
}

// HealthFacilityFileRepository provides lazy, indexed access to the health snapshot.
type HealthFacilityFileRepository struct {
	jsonRepository interfaces.JSONFileRepository
	healthPath     string
	statesPath     string
	lgasPath       string
	cache          lazyDatasetCache[healthFacilityDataset]
}

type healthFacilityDataset struct {
	records     []models.HealthFacility
	all         []int
	searchNames []string
	byID        map[string]int
	byState     map[string][]int
	byLGA       map[string][]int
	byType      map[string][]int
	byLevel     map[string][]int
	byOwnership map[string][]int
	stateIDs    map[string]struct{}
	lgaStates   map[string]string
}

var _ interfaces.HealthFacilityRepository = (*HealthFacilityFileRepository)(nil)

// NewHealthFacilityRepository constructs a file-backed health repository.
func NewHealthFacilityRepository(jsonRepository interfaces.JSONFileRepository, healthPath, statesPath, lgasPath string) (*HealthFacilityFileRepository, error) {
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	cleanHealthPath, err := validateGeographyDatasetPath("health facilities", healthPath)
	if err != nil {
		return nil, err
	}
	cleanStatesPath, err := validateGeographyDatasetPath("states", statesPath)
	if err != nil {
		return nil, err
	}
	cleanLGAsPath, err := validateGeographyDatasetPath("local government areas", lgasPath)
	if err != nil {
		return nil, err
	}
	return &HealthFacilityFileRepository{
		jsonRepository: jsonRepository,
		healthPath:     cleanHealthPath,
		statesPath:     cleanStatesPath,
		lgasPath:       cleanLGAsPath,
	}, nil
}

func (r *HealthFacilityFileRepository) ListHealthFacilities(ctx context.Context, query interfaces.HealthFacilityQuery) (interfaces.HealthFacilityListResult, error) {
	dataset, err := r.load(ctx)
	if err != nil {
		return interfaces.HealthFacilityListResult{}, err
	}
	query, err = normalizeHealthFacilityQuery(query, dataset)
	if err != nil {
		return interfaces.HealthFacilityListResult{}, err
	}
	if err := contextError(ctx); err != nil {
		return interfaces.HealthFacilityListResult{}, err
	}

	candidates := healthFacilityCandidates(dataset, query)
	matching := make([]int, 0, len(candidates))
	search := strings.ToLower(query.Search)
	for i, index := range candidates {
		if i%256 == 0 {
			if err := contextError(ctx); err != nil {
				return interfaces.HealthFacilityListResult{}, err
			}
		}
		facility := dataset.records[index]
		if query.StateID != "" && facility.StateID != query.StateID {
			continue
		}
		if query.LGAID != "" && facility.LGAID != query.LGAID {
			continue
		}
		if query.FacilityType != "" && facility.FacilityType != query.FacilityType {
			continue
		}
		if query.FacilityLevel != "" && facility.FacilityLevel != query.FacilityLevel {
			continue
		}
		if query.OwnershipType != "" && facility.OwnershipType != query.OwnershipType {
			continue
		}
		if search != "" && !strings.Contains(dataset.searchNames[index], search) {
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
		return interfaces.HealthFacilityListResult{
			Facilities: make([]models.HealthFacility, 0), Page: query.Page, PageSize: query.PageSize,
			Total: total, TotalPages: totalPages,
		}, nil
	}
	end := start + query.PageSize
	if end > total {
		end = total
	}
	page := make([]models.HealthFacility, end-start)
	for i, index := range matching[start:end] {
		page[i] = cloneHealthFacility(dataset.records[index])
	}
	return interfaces.HealthFacilityListResult{
		Facilities: page, Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages,
	}, nil
}

func (r *HealthFacilityFileRepository) GetHealthFacility(ctx context.Context, id string) (models.HealthFacility, error) {
	dataset, err := r.load(ctx)
	if err != nil {
		return models.HealthFacility{}, err
	}
	if err := contextError(ctx); err != nil {
		return models.HealthFacility{}, err
	}
	index, ok := dataset.byID[id]
	if !ok {
		return models.HealthFacility{}, fmt.Errorf("%w", interfaces.ErrHealthFacilityNotFound)
	}
	return cloneHealthFacility(dataset.records[index]), nil
}

func (r *HealthFacilityFileRepository) load(ctx context.Context) (healthFacilityDataset, error) {
	values, err := r.cache.get(ctx, func(ctx context.Context) ([]healthFacilityDataset, error) {
		dataset, err := r.decodeAndIndex(ctx)
		if err != nil {
			return nil, err
		}
		return []healthFacilityDataset{dataset}, nil
	}, func(value []healthFacilityDataset) []healthFacilityDataset {
		return value
	})
	if err != nil {
		return healthFacilityDataset{}, err
	}
	return values[0], nil
}

func (r *HealthFacilityFileRepository) decodeAndIndex(ctx context.Context) (healthFacilityDataset, error) {
	var records []models.HealthFacility
	if err := r.jsonRepository.Decode(ctx, r.healthPath, &records); err != nil {
		return healthFacilityDataset{}, sanitizeHealthFacilityLoadError(err)
	}
	var states []models.State
	if err := r.jsonRepository.Decode(ctx, r.statesPath, &states); err != nil {
		return healthFacilityDataset{}, sanitizeHealthFacilityLoadError(err)
	}
	var lgas []models.LocalGovernmentUnit
	if err := r.jsonRepository.Decode(ctx, r.lgasPath, &lgas); err != nil {
		return healthFacilityDataset{}, sanitizeHealthFacilityLoadError(err)
	}

	stateIDs := make(map[string]struct{}, len(states))
	for i, state := range states {
		if i%128 == 0 {
			if err := contextError(ctx); err != nil {
				return healthFacilityDataset{}, err
			}
		}
		if state.ID == "" || state.CountryCode != "NG" || !healthFacilityIDPattern.MatchString(state.ID) {
			return healthFacilityDataset{}, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		stateIDs[state.ID] = struct{}{}
	}
	lgaStates := make(map[string]string, len(lgas))
	for i, lga := range lgas {
		if i%128 == 0 {
			if err := contextError(ctx); err != nil {
				return healthFacilityDataset{}, err
			}
		}
		if lga.ID == "" || lga.StateID == "" || lga.CountryCode != "NG" || !healthFacilityIDPattern.MatchString(lga.ID) {
			return healthFacilityDataset{}, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := stateIDs[lga.StateID]; !ok {
			return healthFacilityDataset{}, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, duplicate := lgaStates[lga.ID]; duplicate {
			return healthFacilityDataset{}, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		lgaStates[lga.ID] = lga.StateID
	}
	if err := validateHealthFacilityRecords(ctx, records, stateIDs, lgaStates); err != nil {
		return healthFacilityDataset{}, err
	}

	dataset := healthFacilityDataset{
		records: records, all: make([]int, len(records)), searchNames: make([]string, len(records)), byID: make(map[string]int, len(records)), byState: map[string][]int{}, byLGA: map[string][]int{},
		byType: map[string][]int{}, byLevel: map[string][]int{}, byOwnership: map[string][]int{}, stateIDs: stateIDs, lgaStates: lgaStates,
	}
	for i, facility := range records {
		if i%256 == 0 {
			if err := contextError(ctx); err != nil {
				return healthFacilityDataset{}, err
			}
		}
		dataset.byID[facility.ID] = i
		dataset.all[i] = i
		dataset.searchNames[i] = strings.ToLower(facility.Name)
		dataset.byState[facility.StateID] = append(dataset.byState[facility.StateID], i)
		if facility.LGAID != "" {
			dataset.byLGA[facility.LGAID] = append(dataset.byLGA[facility.LGAID], i)
		}
		dataset.byType[facility.FacilityType] = append(dataset.byType[facility.FacilityType], i)
		if facility.FacilityLevel != "" {
			dataset.byLevel[facility.FacilityLevel] = append(dataset.byLevel[facility.FacilityLevel], i)
		}
		if facility.OwnershipType != "" {
			dataset.byOwnership[facility.OwnershipType] = append(dataset.byOwnership[facility.OwnershipType], i)
		}
	}
	return dataset, nil
}

func validateHealthFacilityRecords(ctx context.Context, records []models.HealthFacility, stateIDs map[string]struct{}, lgaStates map[string]string) error {
	if len(records) != healthFacilityRecordCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	seenIDs := make(map[string]struct{}, len(records))
	seenSources := make(map[string]struct{}, len(records))
	for i, facility := range records {
		if i%256 == 0 {
			if err := contextError(ctx); err != nil {
				return err
			}
		}
		if facility.ID == "" || facility.Name == "" || facility.FacilityType == "" || facility.StateID == "" || facility.CountryCode != "NG" || facility.SourceFacilityID == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !healthFacilityIDPattern.MatchString(facility.ID) || strings.TrimSpace(facility.SourceFacilityID) != facility.SourceFacilityID {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := healthFacilityTypes[facility.FacilityType]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if facility.FacilityLevel != "" {
			if _, ok := healthFacilityLevels[facility.FacilityLevel]; !ok {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		if facility.OwnershipType != "" {
			if _, ok := healthFacilityOwnerships[facility.OwnershipType]; !ok {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		if facility.OperationalStatus != "" || strings.TrimSpace(facility.Name) != facility.Name || strings.TrimSpace(facility.SourceFacilityID) != facility.SourceFacilityID {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := stateIDs[facility.StateID]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if facility.LGAID != "" {
			if !healthFacilityIDPattern.MatchString(facility.LGAID) || lgaStates[facility.LGAID] != facility.StateID {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		if facility.Latitude == nil || facility.Longitude == nil || *facility.Latitude < healthFacilityMinLatitude || *facility.Latitude > healthFacilityMaxLatitude || *facility.Longitude < healthFacilityMinLongitude || *facility.Longitude > healthFacilityMaxLongitude {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, duplicate := seenIDs[facility.ID]; duplicate {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, duplicate := seenSources[facility.SourceFacilityID]; duplicate {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		seenIDs[facility.ID] = struct{}{}
		seenSources[facility.SourceFacilityID] = struct{}{}
		if i > 0 && healthFacilitySortKey(records[i-1]) > healthFacilitySortKey(facility) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	return nil
}

func normalizeHealthFacilityQuery(query interfaces.HealthFacilityQuery, dataset healthFacilityDataset) (interfaces.HealthFacilityQuery, error) {
	if query.Page == 0 {
		query.Page = healthFacilityDefaultPage
	}
	if query.PageSize == 0 {
		query.PageSize = healthFacilityDefaultSize
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > healthFacilityMaxPageSize {
		return interfaces.HealthFacilityQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidHealthFacilityQuery)
	}
	query.StateID = strings.TrimSpace(query.StateID)
	query.LGAID = strings.TrimSpace(query.LGAID)
	query.FacilityType = strings.TrimSpace(query.FacilityType)
	query.FacilityLevel = strings.TrimSpace(query.FacilityLevel)
	query.OwnershipType = strings.TrimSpace(query.OwnershipType)
	query.Search = strings.TrimSpace(query.Search)
	if query.StateID != "" {
		if _, ok := dataset.stateIDs[query.StateID]; !ok || !healthFacilityIDPattern.MatchString(query.StateID) {
			return interfaces.HealthFacilityQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidHealthFacilityStateFilter)
		}
	}
	if query.LGAID != "" {
		stateID, ok := dataset.lgaStates[query.LGAID]
		if !ok || !healthFacilityIDPattern.MatchString(query.LGAID) {
			return interfaces.HealthFacilityQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidHealthFacilityLGAFilter)
		}
		if query.StateID != "" && stateID != query.StateID {
			return interfaces.HealthFacilityQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidHealthFacilityStateLGA)
		}
	}
	if query.FacilityType != "" {
		if _, ok := healthFacilityTypes[query.FacilityType]; !ok {
			return interfaces.HealthFacilityQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidHealthFacilityType)
		}
	}
	if query.FacilityLevel != "" {
		if _, ok := healthFacilityLevels[query.FacilityLevel]; !ok {
			return interfaces.HealthFacilityQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidHealthFacilityLevel)
		}
	}
	if query.OwnershipType != "" {
		if _, ok := healthFacilityOwnerships[query.OwnershipType]; !ok {
			return interfaces.HealthFacilityQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidHealthFacilityOwnership)
		}
	}
	if len([]rune(query.Search)) > healthFacilityMaxSearch || strings.ContainsAny(query.Search, "\r\n\x00") {
		return interfaces.HealthFacilityQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidHealthFacilitySearch)
	}
	return query, nil
}

func healthFacilityCandidates(dataset healthFacilityDataset, query interfaces.HealthFacilityQuery) []int {
	sets := make([][]int, 0, 6)
	if query.StateID != "" {
		sets = append(sets, dataset.byState[query.StateID])
	}
	if query.LGAID != "" {
		sets = append(sets, dataset.byLGA[query.LGAID])
	}
	if query.FacilityType != "" {
		sets = append(sets, dataset.byType[query.FacilityType])
	}
	if query.FacilityLevel != "" {
		sets = append(sets, dataset.byLevel[query.FacilityLevel])
	}
	if query.OwnershipType != "" {
		sets = append(sets, dataset.byOwnership[query.OwnershipType])
	}
	if len(sets) == 0 {
		return dataset.all
	}
	sort.SliceStable(sets, func(i, j int) bool { return len(sets[i]) < len(sets[j]) })
	return sets[0]
}

func healthFacilitySortKey(facility models.HealthFacility) string {
	return facility.StateID + "\x00" + facility.LGAID + "\x00" + strings.ToLower(facility.Name) + "\x00" + facility.ID
}

func cloneHealthFacility(facility models.HealthFacility) models.HealthFacility {
	if facility.Latitude != nil {
		value := *facility.Latitude
		facility.Latitude = &value
	}
	if facility.Longitude != nil {
		value := *facility.Longitude
		facility.Longitude = &value
	}
	return facility
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func sanitizeHealthFacilityLoadError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, interfaces.ErrDatasetFileNotFound) || errors.Is(err, interfaces.ErrDatasetFileUnavailable) {
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
}
