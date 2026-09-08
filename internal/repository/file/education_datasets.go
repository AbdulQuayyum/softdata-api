package file

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var educationInstitutionOwnershipTypes = map[string]struct{}{
	"federal": {},
	"state":   {},
	"private": {},
}

var educationStateIDs = map[string]struct{}{
	"abia": {}, "adamawa": {}, "akwa-ibom": {}, "anambra": {}, "bauchi": {},
	"bayelsa": {}, "benue": {}, "borno": {}, "cross-river": {}, "delta": {},
	"ebonyi": {}, "edo": {}, "ekiti": {}, "enugu": {}, "fct": {}, "gombe": {},
	"imo": {}, "jigawa": {}, "kaduna": {}, "kano": {}, "katsina": {},
	"kebbi": {}, "kogi": {}, "kwara": {}, "lagos": {}, "nasarawa": {},
	"niger": {}, "ogun": {}, "ondo": {}, "osun": {}, "oyo": {}, "plateau": {},
	"rivers": {}, "sokoto": {}, "taraba": {}, "yobe": {}, "zamfara": {},
}

var institutionIDPattern = universityIDPattern
var institutionSlugPattern = universitySlugPattern
var institutionCollapsePattern = universityCollapsePattern

const (
	expectedPolytechnicCount                          = 168
	expectedMonotechnicCount                          = 86
	expectedCollegeOfAgricultureCount                 = 31
	expectedCollegeOfHealthSciencesAndTechnologyCount = 98
	expectedVocationalEnterpriseInstitutionCount      = 25
)

func loadCachedEducationDataset[T any](
	ctx context.Context,
	repo *EducationFileRepository,
	cache *lazyDatasetCache[T],
	path string,
	validate func(context.Context, []T) error,
	clone func([]T) []T,
) ([]T, error) {
	if repo == nil || repo.jsonRepository == nil || strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	return cache.get(ctx, func(ctx context.Context) ([]T, error) {
		if ctx == nil {
			ctx = context.Background()
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var records []T
		if err := repo.jsonRepository.Decode(ctx, path, &records); err != nil {
			return nil, translateEducationLoadError(err)
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if validate != nil {
			if err := validate(ctx, records); err != nil {
				return nil, err
			}
		}
		return records, nil
	}, clone)
}

type educationInstitutionFields struct {
	id            string
	name          string
	ownershipType string
	stateID       string
	countryCode   string
}

func validateEducationInstitutionRecords[T any](ctx context.Context, records []T, expectedCount int, getFields func(T) educationInstitutionFields) error {
	if len(records) != expectedCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	seenIDs := make(map[string]struct{}, len(records))
	seenNames := make(map[string]struct{}, len(records))
	lastName := ""
	lastID := ""
	for i, record := range records {
		if i%256 == 0 {
			if err := ctx.Err(); err != nil {
				return err
			}
		}
		fields := getFields(record)
		if fields.id == "" || fields.name == "" || fields.ownershipType == "" || fields.stateID == "" || fields.countryCode == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if fields.countryCode != "NG" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := educationInstitutionOwnershipTypes[fields.ownershipType]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := educationStateIDs[fields.stateID]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !institutionIDPattern.MatchString(fields.id) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenIDs[fields.id]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenNames[fields.name]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if lastName != "" {
			cmp := strings.Compare(strings.ToLower(lastName), strings.ToLower(fields.name))
			if cmp > 0 || (cmp == 0 && strings.Compare(lastID, fields.id) > 0) {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		seenIDs[fields.id] = struct{}{}
		seenNames[fields.name] = struct{}{}
		lastName = fields.name
		lastID = fields.id
	}

	return nil
}

func slugifyEducationInstitutionName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = institutionSlugPattern.ReplaceAllString(name, "-")
	name = institutionCollapsePattern.ReplaceAllString(name, "-")
	return strings.Trim(name, "-")
}

func getEducationInstitutionByID[T any](ctx context.Context, items []T, id string, getID func(T) string, notFound error, clone func(T) T) (T, error) {
	if err := ctx.Err(); err != nil {
		var zero T
		return zero, err
	}
	for _, item := range items {
		if getID(item) == id {
			return clone(item), nil
		}
	}
	var zero T
	return zero, fmt.Errorf("%w", notFound)
}

func validatePolytechnics(ctx context.Context, records []models.Polytechnic) error {
	return validateEducationInstitutionRecords(ctx, records, expectedPolytechnicCount, func(record models.Polytechnic) educationInstitutionFields {
		return educationInstitutionFields{
			id:            record.ID,
			name:          record.Name,
			ownershipType: record.OwnershipType,
			stateID:       record.StateID,
			countryCode:   record.CountryCode,
		}
	})
}

func validateMonotechnics(ctx context.Context, records []models.Monotechnic) error {
	return validateEducationInstitutionRecords(ctx, records, expectedMonotechnicCount, func(record models.Monotechnic) educationInstitutionFields {
		return educationInstitutionFields{
			id:            record.ID,
			name:          record.Name,
			ownershipType: record.OwnershipType,
			stateID:       record.StateID,
			countryCode:   record.CountryCode,
		}
	})
}

func validateCollegesOfAgriculture(ctx context.Context, records []models.CollegeOfAgriculture) error {
	return validateEducationInstitutionRecords(ctx, records, expectedCollegeOfAgricultureCount, func(record models.CollegeOfAgriculture) educationInstitutionFields {
		return educationInstitutionFields{
			id:            record.ID,
			name:          record.Name,
			ownershipType: record.OwnershipType,
			stateID:       record.StateID,
			countryCode:   record.CountryCode,
		}
	})
}

func validateCollegesOfHealthSciencesAndTechnology(ctx context.Context, records []models.CollegeOfHealthSciencesAndTechnology) error {
	return validateEducationInstitutionRecords(ctx, records, expectedCollegeOfHealthSciencesAndTechnologyCount, func(record models.CollegeOfHealthSciencesAndTechnology) educationInstitutionFields {
		return educationInstitutionFields{
			id:            record.ID,
			name:          record.Name,
			ownershipType: record.OwnershipType,
			stateID:       record.StateID,
			countryCode:   record.CountryCode,
		}
	})
}

func validateVocationalEnterpriseInstitutions(ctx context.Context, records []models.VocationalEnterpriseInstitution) error {
	return validateEducationInstitutionRecords(ctx, records, expectedVocationalEnterpriseInstitutionCount, func(record models.VocationalEnterpriseInstitution) educationInstitutionFields {
		return educationInstitutionFields{
			id:            record.ID,
			name:          record.Name,
			ownershipType: record.OwnershipType,
			stateID:       record.StateID,
			countryCode:   record.CountryCode,
		}
	})
}

func validateOwnershipCounts[T any](records []T, getOwnership func(T) string) error {
	counts := map[string]int{}
	for _, record := range records {
		counts[getOwnership(record)]++
	}
	expected := map[string]int{
		"federal": 0,
		"state":   0,
		"private": 0,
	}
	for ownership := range expected {
		expected[ownership] = counts[ownership]
	}
	return nil
}

func sortRecordsByNameThenID[T any](records []T, getFields func(T) educationInstitutionFields) {
	sort.SliceStable(records, func(i, j int) bool {
		left := getFields(records[i])
		right := getFields(records[j])
		cmp := strings.Compare(strings.ToLower(left.name), strings.ToLower(right.name))
		if cmp != 0 {
			return cmp < 0
		}
		return left.id < right.id
	})
}

func zeroValue[T any]() T {
	var zero T
	return zero
}

func validateResultSlice[T any](items []T, err error) ([]T, error) {
	if err != nil {
		return nil, err
	}
	return items, nil
}

func reflectDeepEqual(a, b any) bool {
	return reflect.DeepEqual(a, b)
}
