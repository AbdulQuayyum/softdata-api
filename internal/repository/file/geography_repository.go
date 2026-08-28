package file

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
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

var expectedGeopoliticalZoneCounts = map[string]int{
	"north-central": 7,
	"north-east":    6,
	"north-west":    7,
	"south-east":    5,
	"south-south":   6,
	"south-west":    6,
}

// GeographyFileRepository reads Nigerian geography records from JSON dataset files.
type GeographyFileRepository struct {
	jsonRepository interfaces.JSONFileRepository
	statesPath     string
	zonesPath      string
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
		zonesPath:      filepath.Join(filepath.Dir(cleanedPath), "geopolitical_zones.json"),
	}, nil
}

// ListStates returns the ordered list of states from the dataset file.
func (r *GeographyFileRepository) ListStates(ctx context.Context) ([]models.State, error) {
	states, _, err := r.loadGeographyData(ctx)
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

	states, _, err := r.loadGeographyData(ctx)
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

// ListGeopoliticalZones returns the ordered list of geopolitical zones from the dataset file.
func (r *GeographyFileRepository) ListGeopoliticalZones(ctx context.Context) ([]models.GeopoliticalZone, error) {
	_, zones, err := r.loadGeographyData(ctx)
	if err != nil {
		return nil, err
	}
	return cloneGeopoliticalZones(zones), nil
}

// GetGeopoliticalZone returns a single geopolitical zone using its public slug identifier.
func (r *GeographyFileRepository) GetGeopoliticalZone(ctx context.Context, zoneID string) (models.GeopoliticalZone, error) {
	zoneID = strings.TrimSpace(zoneID)
	if zoneID == "" || !stateIDPattern.MatchString(zoneID) {
		return models.GeopoliticalZone{}, fmt.Errorf("%w", interfaces.ErrGeopoliticalZoneNotFound)
	}

	_, zones, err := r.loadGeographyData(ctx)
	if err != nil {
		return models.GeopoliticalZone{}, err
	}

	for _, zone := range zones {
		if zone.ID == zoneID {
			return cloneGeopoliticalZone(zone), nil
		}
	}

	return models.GeopoliticalZone{}, fmt.Errorf("%w", interfaces.ErrGeopoliticalZoneNotFound)
}

func (r *GeographyFileRepository) loadGeographyData(ctx context.Context) ([]models.State, []models.GeopoliticalZone, error) {
	states, err := r.loadStatesOnly(ctx)
	if err != nil {
		return nil, nil, err
	}
	zones, err := r.loadGeopoliticalZonesOnly(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := validateGeographyData(states, zones); err != nil {
		return nil, nil, err
	}
	return states, zones, nil
}

func (r *GeographyFileRepository) loadStatesOnly(ctx context.Context) ([]models.State, error) {
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
	if states == nil || len(states) == 0 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if err := validateStates(states); err != nil {
		return nil, err
	}

	return states, nil
}

func (r *GeographyFileRepository) loadGeopoliticalZonesOnly(ctx context.Context) ([]models.GeopoliticalZone, error) {
	if r == nil || r.jsonRepository == nil {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var zones []models.GeopoliticalZone
	if err := r.jsonRepository.Decode(ctx, r.zonesPath, &zones); err != nil {
		return nil, translateGeographyLoadError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if zones == nil || len(zones) == 0 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if err := validateGeopoliticalZones(zones); err != nil {
		return nil, err
	}

	return zones, nil
}

func validateStates(states []models.State) error {
	seenIDs := make(map[string]struct{}, len(states))
	seenNames := make(map[string]struct{}, len(states))
	fctCount := 0
	stateCount := 0

	for i, state := range states {
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
		if i > 0 && strings.Compare(states[i-1].Name, state.Name) > 0 {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if state.AdministrativeType == "federal_capital_territory" {
			fctCount++
			if state.ID != "fct" || state.GeopoliticalZoneID != "north-central" {
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

func validateGeopoliticalZones(zones []models.GeopoliticalZone) error {
	seenIDs := make(map[string]struct{}, len(zones))
	seenNames := make(map[string]struct{}, len(zones))

	for i, zone := range zones {
		if zone.ID == "" || zone.Name == "" || zone.CountryCode == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !stateIDPattern.MatchString(zone.ID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenIDs[zone.ID]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenNames[zone.Name]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := validGeopoliticalZones[zone.ID]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if zone.CountryCode != "NG" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if i > 0 && strings.Compare(zones[i-1].Name, zone.Name) > 0 {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}

		seenIDs[zone.ID] = struct{}{}
		seenNames[zone.Name] = struct{}{}
	}

	if len(seenIDs) != 6 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	return nil
}

func validateGeographyData(states []models.State, zones []models.GeopoliticalZone) error {
	if len(states) != 37 || len(zones) != 6 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	zonesByID := make(map[string]struct{}, len(zones))
	for _, zone := range zones {
		zonesByID[zone.ID] = struct{}{}
	}

	zoneCounts := make(map[string]int, len(zones))
	for _, zone := range zones {
		zoneCounts[zone.ID] = 0
	}

	stateCount := 0
	fctCount := 0
	for _, state := range states {
		if _, ok := zonesByID[state.GeopoliticalZoneID]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		zoneCounts[state.GeopoliticalZoneID]++
		switch state.AdministrativeType {
		case "state":
			stateCount++
		case "federal_capital_territory":
			fctCount++
			if state.ID != "fct" || state.GeopoliticalZoneID != "north-central" {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		default:
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}

	if stateCount != 36 || fctCount != 1 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if !reflect.DeepEqual(zoneCounts, expectedGeopoliticalZoneCounts) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	for zoneID, count := range zoneCounts {
		if count == 0 {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := zonesByID[zoneID]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
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

func cloneGeopoliticalZones(zones []models.GeopoliticalZone) []models.GeopoliticalZone {
	if len(zones) == 0 {
		return make([]models.GeopoliticalZone, 0)
	}
	cloned := make([]models.GeopoliticalZone, len(zones))
	copy(cloned, zones)
	return cloned
}

func cloneGeopoliticalZone(zone models.GeopoliticalZone) models.GeopoliticalZone {
	return zone
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
