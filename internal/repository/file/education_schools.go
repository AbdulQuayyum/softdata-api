package file

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const (
	defaultPrimaryAndSecondarySchoolPage     = 1
	defaultPrimaryAndSecondarySchoolPageSize = 50
	maxPrimaryAndSecondarySchoolPageSize     = 100
	maxPrimaryAndSecondarySchoolSearchLength = 256
	maxEducationInstitutionIDLength          = 128
)

var schoolOwnershipTypes = map[string]struct{}{
	"public":  {},
	"private": {},
}

var schoolEducationLevels = map[string]struct{}{
	"pre-primary":      {},
	"primary":          {},
	"junior-secondary": {},
	"senior-secondary": {},
}

var schoolGovernmentOwnerTypes = map[string]struct{}{
	"federal": {},
	"state":   {},
	"local":   {},
}

var schoolCodePattern = regexp.MustCompile(`^[0-9]+$`)

const (
	firstSchoolAnchorID = "na-ube-junior-secondary-school-ajaka-kogi-igalamela-odolu-ajaka"
	lastSchoolAnchorID  = "chediya-primary-school-chediya-village-katsina-danja-chediya-village"
)

type primaryAndSecondarySchoolDataset struct {
	records          []models.PrimaryAndSecondarySchool
	byID             map[string]int
	byStateID        map[string][]int
	byLGAID          map[string][]int
	lgaToStateID     map[string]string
	byEducationLevel map[string][]int
	byOwnershipType  map[string][]int
	allIndices       []int
}

type schoolDatasetCache struct {
	mu      sync.Mutex
	loaded  bool
	dataset *primaryAndSecondarySchoolDataset
	attempt *datasetLoadAttempt
}

func (c *schoolDatasetCache) get(ctx context.Context, load func(context.Context) (*primaryAndSecondarySchoolDataset, error)) (*primaryAndSecondarySchoolDataset, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		c.mu.Lock()
		if c.loaded {
			dataset := c.dataset
			c.mu.Unlock()
			return dataset, nil
		}
		if c.attempt != nil {
			attempt := c.attempt
			c.mu.Unlock()
			select {
			case <-attempt.done:
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				if attempt.err != nil {
					return nil, attempt.err
				}
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		attempt := &datasetLoadAttempt{done: make(chan struct{})}
		c.attempt = attempt
		c.mu.Unlock()

		dataset, err := load(ctx)

		c.mu.Lock()
		if err == nil {
			c.dataset = dataset
			c.loaded = true
		}
		c.attempt = nil
		attempt.err = err
		close(attempt.done)
		c.mu.Unlock()

		if err != nil {
			return nil, err
		}
		return dataset, nil
	}
}

func (r *EducationFileRepository) ListPrimaryAndSecondarySchools(ctx context.Context, query interfaces.PrimaryAndSecondarySchoolQuery) (interfaces.PrimaryAndSecondarySchoolListResult, error) {
	normalized, err := normalizePrimaryAndSecondarySchoolQuery(query)
	if err != nil {
		return interfaces.PrimaryAndSecondarySchoolListResult{}, err
	}

	dataset, err := r.loadPrimaryAndSecondarySchoolDataset(ctx)
	if err != nil {
		return interfaces.PrimaryAndSecondarySchoolListResult{}, err
	}
	if err := validatePrimaryAndSecondarySchoolQueryAgainstDataset(normalized, dataset); err != nil {
		return interfaces.PrimaryAndSecondarySchoolListResult{}, err
	}

	result, err := listPrimaryAndSecondarySchoolsFromDataset(ctx, dataset, normalized)
	if err != nil {
		return interfaces.PrimaryAndSecondarySchoolListResult{}, err
	}
	return result, nil
}

func (r *EducationFileRepository) GetPrimaryAndSecondarySchool(ctx context.Context, id string) (models.PrimaryAndSecondarySchool, error) {
	id = strings.TrimSpace(id)
	if id == "" || len(id) > maxEducationInstitutionIDLength || !institutionIDPattern.MatchString(id) {
		return models.PrimaryAndSecondarySchool{}, fmt.Errorf("%w", interfaces.ErrPrimaryAndSecondarySchoolNotFound)
	}

	dataset, err := r.loadPrimaryAndSecondarySchoolDataset(ctx)
	if err != nil {
		return models.PrimaryAndSecondarySchool{}, err
	}
	index, ok := dataset.byID[id]
	if !ok {
		return models.PrimaryAndSecondarySchool{}, fmt.Errorf("%w", interfaces.ErrPrimaryAndSecondarySchoolNotFound)
	}
	return clonePrimaryAndSecondarySchool(dataset.records[index]), nil
}

func (r *EducationFileRepository) loadPrimaryAndSecondarySchoolDataset(ctx context.Context) (*primaryAndSecondarySchoolDataset, error) {
	if r == nil || r.jsonRepository == nil || strings.TrimSpace(r.primaryAndSecondarySchoolsPath) == "" {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	return r.primaryAndSecondarySchoolsCache.get(ctx, func(ctx context.Context) (*primaryAndSecondarySchoolDataset, error) {
		if ctx == nil {
			ctx = context.Background()
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var records []models.PrimaryAndSecondarySchool
		if err := r.jsonRepository.Decode(ctx, r.primaryAndSecondarySchoolsPath, &records); err != nil {
			return nil, translateEducationLoadError(err)
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}

		dataset, err := buildPrimaryAndSecondarySchoolDataset(ctx, records)
		if err != nil {
			return nil, err
		}
		return dataset, nil
	})
}

func normalizePrimaryAndSecondarySchoolQuery(query interfaces.PrimaryAndSecondarySchoolQuery) (interfaces.PrimaryAndSecondarySchoolQuery, error) {
	normalized := interfaces.PrimaryAndSecondarySchoolQuery{
		Page:           query.Page,
		PageSize:       query.PageSize,
		StateID:        strings.TrimSpace(query.StateID),
		LGAID:          strings.TrimSpace(query.LGAID),
		EducationLevel: strings.TrimSpace(query.EducationLevel),
		OwnershipType:  strings.TrimSpace(query.OwnershipType),
		Search:         strings.TrimSpace(query.Search),
	}

	if normalized.Page == 0 {
		normalized.Page = defaultPrimaryAndSecondarySchoolPage
	}
	if normalized.PageSize == 0 {
		normalized.PageSize = defaultPrimaryAndSecondarySchoolPageSize
	}
	if normalized.Page < 1 || normalized.PageSize < 1 || normalized.PageSize > maxPrimaryAndSecondarySchoolPageSize {
		return interfaces.PrimaryAndSecondarySchoolQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidPrimaryAndSecondarySchoolQuery)
	}
	if normalized.Search != "" && len(normalized.Search) > maxPrimaryAndSecondarySchoolSearchLength {
		return interfaces.PrimaryAndSecondarySchoolQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidPrimaryAndSecondarySchoolQuery)
	}
	if normalized.StateID != "" {
		if _, ok := educationStateIDs[normalized.StateID]; !ok {
			return interfaces.PrimaryAndSecondarySchoolQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidPrimaryAndSecondarySchoolQuery)
		}
	}
	if normalized.LGAID != "" {
		if len(normalized.LGAID) > maxEducationInstitutionIDLength || !institutionIDPattern.MatchString(normalized.LGAID) {
			return interfaces.PrimaryAndSecondarySchoolQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidPrimaryAndSecondarySchoolQuery)
		}
	}
	if normalized.EducationLevel != "" {
		if _, ok := schoolEducationLevels[normalized.EducationLevel]; !ok {
			return interfaces.PrimaryAndSecondarySchoolQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidPrimaryAndSecondarySchoolQuery)
		}
	}
	if normalized.OwnershipType != "" {
		if _, ok := schoolOwnershipTypes[normalized.OwnershipType]; !ok {
			return interfaces.PrimaryAndSecondarySchoolQuery{}, fmt.Errorf("%w", interfaces.ErrInvalidPrimaryAndSecondarySchoolQuery)
		}
	}
	return normalized, nil
}

func validatePrimaryAndSecondarySchoolQueryAgainstDataset(query interfaces.PrimaryAndSecondarySchoolQuery, dataset *primaryAndSecondarySchoolDataset) error {
	if dataset == nil {
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if query.LGAID != "" {
		stateID, ok := dataset.lgaToStateID[query.LGAID]
		if !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidPrimaryAndSecondarySchoolQuery)
		}
		if query.StateID != "" && stateID != query.StateID {
			return fmt.Errorf("%w", interfaces.ErrInvalidPrimaryAndSecondarySchoolQuery)
		}
	}
	return nil
}

func buildPrimaryAndSecondarySchoolDataset(ctx context.Context, records []models.PrimaryAndSecondarySchool) (*primaryAndSecondarySchoolDataset, error) {
	cloned := clonePrimaryAndSecondarySchoolList(records)
	if len(cloned) != 166604 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	dataset := &primaryAndSecondarySchoolDataset{
		records:          cloned,
		byID:             make(map[string]int, len(cloned)),
		byStateID:        make(map[string][]int),
		byLGAID:          make(map[string][]int),
		lgaToStateID:     make(map[string]string),
		byEducationLevel: make(map[string][]int),
		byOwnershipType:  make(map[string][]int),
		allIndices:       make([]int, 0, len(cloned)),
	}

	first := cloned[0]
	last := cloned[len(cloned)-1]
	if first.ID != firstSchoolAnchorID || last.ID != lastSchoolAnchorID {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	lastName := ""
	lastID := ""
	for i, school := range cloned {
		if i%1024 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
		if school.ID == "" || school.Name == "" || school.StateID == "" || school.CountryCode == "" || len(school.EducationLevels) == 0 {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if school.CountryCode != "NG" {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if school.OwnershipType == "" {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := schoolOwnershipTypes[school.OwnershipType]; !ok {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := educationStateIDs[school.StateID]; !ok {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if school.LGAID == "" || len(school.LGAID) > maxEducationInstitutionIDLength || !institutionIDPattern.MatchString(school.LGAID) {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !strings.HasPrefix(school.LGAID, school.StateID+"-") {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if school.UBECSchoolCode != "" && !schoolCodePattern.MatchString(school.UBECSchoolCode) {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if school.GovernmentOwner != "" {
			if _, ok := schoolGovernmentOwnerTypes[school.GovernmentOwner]; !ok {
				return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		if !institutionIDPattern.MatchString(school.ID) {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := dataset.byID[school.ID]; ok {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if lastName != "" {
			cmp := strings.Compare(strings.ToLower(lastName), strings.ToLower(school.Name))
			if cmp > 0 || (cmp == 0 && strings.Compare(lastID, school.ID) > 0) {
				return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}

		levelSeen := make(map[string]struct{}, len(school.EducationLevels))
		for _, level := range school.EducationLevels {
			if level == "" {
				return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			if _, ok := schoolEducationLevels[level]; !ok {
				return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			if _, ok := levelSeen[level]; ok {
				return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			levelSeen[level] = struct{}{}
			dataset.byEducationLevel[level] = append(dataset.byEducationLevel[level], i)
		}

		dataset.byID[school.ID] = i
		dataset.byStateID[school.StateID] = append(dataset.byStateID[school.StateID], i)
		dataset.byLGAID[school.LGAID] = append(dataset.byLGAID[school.LGAID], i)
		if existingState, ok := dataset.lgaToStateID[school.LGAID]; ok && existingState != school.StateID {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		dataset.lgaToStateID[school.LGAID] = school.StateID
		dataset.byOwnershipType[school.OwnershipType] = append(dataset.byOwnershipType[school.OwnershipType], i)
		dataset.allIndices = append(dataset.allIndices, i)

		lastName = school.Name
		lastID = school.ID
	}

	return dataset, nil
}

func listPrimaryAndSecondarySchoolsFromDataset(ctx context.Context, dataset *primaryAndSecondarySchoolDataset, query interfaces.PrimaryAndSecondarySchoolQuery) (interfaces.PrimaryAndSecondarySchoolListResult, error) {
	candidates := dataset.allIndices
	if indices := dataset.byStateID[query.StateID]; len(indices) > 0 && len(indices) < len(candidates) {
		candidates = indices
	}
	if indices := dataset.byLGAID[query.LGAID]; len(indices) > 0 && len(indices) < len(candidates) {
		candidates = indices
	}
	if indices := dataset.byEducationLevel[query.EducationLevel]; len(indices) > 0 && len(indices) < len(candidates) {
		candidates = indices
	}
	if indices := dataset.byOwnershipType[query.OwnershipType]; len(indices) > 0 && len(indices) < len(candidates) {
		candidates = indices
	}

	pageStart := int64(query.Page-1) * int64(query.PageSize)
	pageEnd := pageStart + int64(query.PageSize)
	result := make([]models.PrimaryAndSecondarySchool, 0, query.PageSize)
	var total int

	searchLower := strings.ToLower(query.Search)
	for i, index := range candidates {
		if i%1024 == 0 {
			if err := ctx.Err(); err != nil {
				return interfaces.PrimaryAndSecondarySchoolListResult{}, err
			}
		}
		school := dataset.records[index]
		if query.StateID != "" && school.StateID != query.StateID {
			continue
		}
		if query.LGAID != "" && school.LGAID != query.LGAID {
			continue
		}
		if query.EducationLevel != "" && !schoolHasEducationLevel(school.EducationLevels, query.EducationLevel) {
			continue
		}
		if query.OwnershipType != "" && school.OwnershipType != query.OwnershipType {
			continue
		}
		if searchLower != "" && !schoolMatchesSearch(school, searchLower) {
			continue
		}

		if int64(total) >= pageStart && int64(total) < pageEnd {
			result = append(result, clonePrimaryAndSecondarySchool(school))
		}
		total++
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + query.PageSize - 1) / query.PageSize
	}
	if int64(total) <= pageStart {
		result = make([]models.PrimaryAndSecondarySchool, 0)
	}
	return interfaces.PrimaryAndSecondarySchoolListResult{
		Schools:    result,
		Page:       query.Page,
		PageSize:   query.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func schoolHasEducationLevel(levels []string, want string) bool {
	for _, level := range levels {
		if level == want {
			return true
		}
	}
	return false
}

func schoolMatchesSearch(school models.PrimaryAndSecondarySchool, searchLower string) bool {
	if searchLower == "" {
		return true
	}
	if strings.Contains(strings.ToLower(school.Name), searchLower) {
		return true
	}
	if school.UBECSchoolCode != "" && strings.Contains(strings.ToLower(school.UBECSchoolCode), searchLower) {
		return true
	}
	return false
}
