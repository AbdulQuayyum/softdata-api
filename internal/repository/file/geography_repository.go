package file

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var stateIDPattern = regexp.MustCompile(`^[a-z]+(?:-[a-z]+)*$`)

var validGeopoliticalZones = map[string]struct{}{
	"north-central": {},
	"north-east":    {},
	"north-west":    {},
	"south-east":    {},
	"south-south":   {},
	"south-west":    {},
}

var validAdministrativeTypes = map[string]struct{}{
	"state":                     {},
	"federal_capital_territory": {},
}

// GeographyFileRepository reads Nigerian state records from a JSON dataset file.
type GeographyFileRepository struct {
	jsonRepository interfaces.JSONFileRepository
	statesPath     string
}

var _ interfaces.GeographyRepository = (*GeographyFileRepository)(nil)

// NewGeographyRepository constructs a file-backed geography repository.
func NewGeographyRepository(jsonRepository interfaces.JSONFileRepository, statesPath string) (*GeographyFileRepository, error) {
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	statesPath = strings.TrimSpace(statesPath)
	if statesPath == "" {
		return nil, fmt.Errorf("states path is required")
	}
	if filepath.IsAbs(statesPath) || filepath.VolumeName(statesPath) != "" {
		return nil, fmt.Errorf("states path must remain relative")
	}
	cleanedPath := filepath.Clean(statesPath)
	if cleanedPath == "." || strings.HasPrefix(cleanedPath, ".."+string(filepath.Separator)) || cleanedPath == ".." {
		return nil, fmt.Errorf("states path must remain relative")
	}

	return &GeographyFileRepository{
		jsonRepository: jsonRepository,
		statesPath:     cleanedPath,
	}, nil
}

// ListStates returns the ordered list of states from the dataset file.
func (r *GeographyFileRepository) ListStates(ctx context.Context) ([]models.State, error) {
	states, err := r.loadStates(ctx)
	if err != nil {
		return nil, err
	}
	return cloneStates(states), nil
}

// GetStateByID returns a single state using its public slug identifier.
func (r *GeographyFileRepository) GetStateByID(ctx context.Context, stateID string) (models.State, error) {
	stateID = strings.TrimSpace(stateID)
	if stateID == "" || !stateIDPattern.MatchString(stateID) {
		return models.State{}, fmt.Errorf("%w", interfaces.ErrStateNotFound)
	}

	states, err := r.loadStates(ctx)
	if err != nil {
		return models.State{}, err
	}

	for _, state := range states {
		if state.ID == stateID {
			return cloneState(state), nil
		}
	}

	return models.State{}, fmt.Errorf("%w", interfaces.ErrStateNotFound)
}

func (r *GeographyFileRepository) loadStates(ctx context.Context) ([]models.State, error) {
	if r == nil || r.jsonRepository == nil {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var states []models.State
	if err := r.jsonRepository.Decode(ctx, r.statesPath, &states); err != nil {
		return nil, translateGeographyLoadError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if states == nil {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if len(states) == 0 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if len(states) != 37 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if err := validateStates(states); err != nil {
		return nil, err
	}

	return states, nil
}

func validateStates(states []models.State) error {
	seenIDs := make(map[string]struct{}, len(states))
	seenNames := make(map[string]struct{}, len(states))
	fctCount := 0
	stateCount := 0

	for _, state := range states {
		if state.ID == "" || state.Name == "" || state.OfficialName == "" || state.AdministrativeType == "" || state.Capital == "" || state.GeopoliticalZoneID == "" || state.CountryCode == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !stateIDPattern.MatchString(state.ID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenIDs[state.ID]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenNames[state.Name]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := validAdministrativeTypes[state.AdministrativeType]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := validGeopoliticalZones[state.GeopoliticalZoneID]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if state.CountryCode != "NG" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if state.AdministrativeType == "federal_capital_territory" {
			fctCount++
			if state.ID != "fct" {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		} else {
			stateCount++
		}

		seenIDs[state.ID] = struct{}{}
		seenNames[state.Name] = struct{}{}
	}

	if stateCount != 36 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if fctCount != 1 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	return nil
}

func cloneStates(states []models.State) []models.State {
	if len(states) == 0 {
		return make([]models.State, 0)
	}
	cloned := make([]models.State, len(states))
	copy(cloned, states)
	return cloned
}

func cloneState(state models.State) models.State {
	return state
}

func translateGeographyLoadError(err error) error {
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
