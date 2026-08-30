package file

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var universityIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)
var universitySlugPattern = regexp.MustCompile(`[^a-z0-9]+`)
var universityCollapsePattern = regexp.MustCompile(`-+`)

var expectedUniversityOwnershipOrder = map[string]int{
	"federal": 0,
	"state":   1,
	"private": 2,
}

var expectedUniversityOwnershipCounts = map[string]int{
	"federal": 77,
	"state":   69,
	"private": 182,
}

var expectedUniversityStateCounts = map[string]int{
	"abia": 8, "adamawa": 4, "akwa-ibom": 9, "anambra": 10, "bauchi": 3,
	"bayelsa": 6, "benue": 5, "borno": 4, "cross-river": 8, "delta": 15,
	"ebonyi": 7, "edo": 11, "ekiti": 8, "enugu": 11, "fct": 21, "gombe": 4,
	"imo": 11, "jigawa": 4, "kaduna": 11, "kano": 14, "katsina": 5,
	"kebbi": 4, "kogi": 6, "kwara": 14, "lagos": 16, "nasarawa": 6,
	"niger": 10, "ogun": 25, "ondo": 11, "osun": 15, "oyo": 13,
	"plateau": 5, "rivers": 7, "sokoto": 6, "taraba": 5, "yobe": 2,
	"zamfara": 4,
}

// EducationFileRepository reads university records from a JSON dataset file.
type EducationFileRepository struct {
	jsonRepository   interfaces.JSONFileRepository
	universitiesPath string
}

var _ interfaces.EducationRepository = (*EducationFileRepository)(nil)

// NewEducationRepository constructs a file-backed education repository.
func NewEducationRepository(jsonRepository interfaces.JSONFileRepository, universitiesPath string) (*EducationFileRepository, error) {
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	cleanedPath, err := validateGeographyDatasetPath("universities", universitiesPath)
	if err != nil {
		return nil, err
	}

	return &EducationFileRepository{
		jsonRepository:   jsonRepository,
		universitiesPath: cleanedPath,
	}, nil
}

// ListUniversities returns the ordered list of universities matching the supplied filter.
func (r *EducationFileRepository) ListUniversities(ctx context.Context, filter interfaces.UniversityFilter) ([]models.University, error) {
	universities, err := r.loadUniversities(ctx)
	if err != nil {
		return nil, err
	}

	filtered := make([]models.University, 0, len(universities))
	for _, university := range universities {
		if filter.OwnershipType != "" && university.OwnershipType != filter.OwnershipType {
			continue
		}
		if filter.StateID != "" && university.StateID != filter.StateID {
			continue
		}
		filtered = append(filtered, university)
	}
	return cloneUniversityList(filtered), nil
}

// GetUniversityByID returns a single university using its public slug identifier.
func (r *EducationFileRepository) GetUniversityByID(ctx context.Context, universityID string) (models.University, error) {
	universities, err := r.loadUniversities(ctx)
	if err != nil {
		return models.University{}, err
	}

	for _, university := range universities {
		if university.ID == universityID {
			return cloneUniversity(university), nil
		}
	}

	return models.University{}, fmt.Errorf("%w", interfaces.ErrUniversityNotFound)
}

func (r *EducationFileRepository) loadUniversities(ctx context.Context) ([]models.University, error) {
	if r == nil || r.jsonRepository == nil {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var universities []models.University
	if err := r.jsonRepository.Decode(ctx, r.universitiesPath, &universities); err != nil {
		return nil, translateEducationLoadError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if universities == nil || len(universities) == 0 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if err := validateUniversities(universities); err != nil {
		return nil, err
	}

	return universities, nil
}

func validateUniversities(universities []models.University) error {
	if len(universities) != 328 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	seenIDs := make(map[string]struct{}, len(universities))
	seenNames := make(map[string]struct{}, len(universities))
	seenOwnerships := make(map[string]struct{}, len(expectedUniversityOwnershipCounts))
	ownershipCounts := make(map[string]int, len(expectedUniversityOwnershipCounts))
	stateCounts := make(map[string]int, len(expectedUniversityStateCounts))
	currentOwnership := ""
	currentOwnershipOrder := -1
	lastName := ""
	lastID := ""

	for stateID := range expectedUniversityStateCounts {
		stateCounts[stateID] = 0
	}

	for _, university := range universities {
		if university.ID == "" || university.Name == "" || university.OwnershipType == "" || university.StateID == "" || university.CountryCode == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if university.CountryCode != "NG" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		ownershipOrder, ok := expectedUniversityOwnershipOrder[university.OwnershipType]
		if !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := expectedUniversityStateCounts[university.StateID]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !universityIDPattern.MatchString(university.ID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if wantID := slugifyUniversityName(university.Name); university.ID != wantID {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenIDs[university.ID]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenNames[university.Name]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}

		if currentOwnership == "" {
			currentOwnership = university.OwnershipType
			currentOwnershipOrder = ownershipOrder
			lastName = ""
			lastID = ""
		} else if university.OwnershipType != currentOwnership {
			if ownershipOrder < currentOwnershipOrder {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			if _, ok := seenOwnerships[university.OwnershipType]; ok {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			seenOwnerships[currentOwnership] = struct{}{}
			currentOwnership = university.OwnershipType
			currentOwnershipOrder = ownershipOrder
			lastName = ""
			lastID = ""
		}
		if lastName != "" {
			cmp := strings.Compare(strings.ToLower(lastName), strings.ToLower(university.Name))
			if cmp > 0 || (cmp == 0 && strings.Compare(lastID, university.ID) > 0) {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}

		seenIDs[university.ID] = struct{}{}
		seenNames[university.Name] = struct{}{}
		ownershipCounts[university.OwnershipType]++
		stateCounts[university.StateID]++
		lastName = university.Name
		lastID = university.ID
	}

	if currentOwnership != "" {
		seenOwnerships[currentOwnership] = struct{}{}
	}
	if !reflect.DeepEqual(ownershipCounts, expectedUniversityOwnershipCounts) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if !reflect.DeepEqual(stateCounts, expectedUniversityStateCounts) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if len(seenOwnerships) != len(expectedUniversityOwnershipCounts) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	return nil
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

func slugifyUniversityName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = universitySlugPattern.ReplaceAllString(name, "-")
	name = universityCollapsePattern.ReplaceAllString(name, "-")
	return strings.Trim(name, "-")
}

func translateEducationLoadError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrDatasetFileNotFound):
		return fmt.Errorf("%w", interfaces.ErrDatasetFileNotFound)
	case errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	case errors.Is(err, interfaces.ErrInvalidDatasetFile):
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	default:
		return fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
}
