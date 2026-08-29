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
var localGovernmentUnitIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)

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

var validLocalGovernmentUnitTypes = map[string]struct{}{
	"local_government_area": {},
	"area_council":          {},
}

var expectedGeopoliticalZoneCounts = map[string]int{
	"north-central": 7,
	"north-east":    6,
	"north-west":    7,
	"south-east":    5,
	"south-south":   6,
	"south-west":    6,
}

var expectedLocalGovernmentUnitCounts = map[string]int{
	"abia": 17, "adamawa": 21, "akwa-ibom": 31, "anambra": 21, "bauchi": 20,
	"bayelsa": 8, "benue": 23, "borno": 27, "cross-river": 18, "delta": 25,
	"ebonyi": 13, "edo": 18, "ekiti": 16, "enugu": 17, "fct": 6, "gombe": 11,
	"imo": 27, "jigawa": 27, "kaduna": 23, "kano": 44, "katsina": 34,
	"kebbi": 21, "kogi": 21, "kwara": 16, "lagos": 20, "nasarawa": 13,
	"niger": 25, "ogun": 20, "ondo": 18, "osun": 30, "oyo": 33, "plateau": 17,
	"rivers": 23, "sokoto": 23, "taraba": 16, "yobe": 17, "zamfara": 14,
}

var expectedLocalGovernmentUnitZoneCounts = map[string]int{
	"north-central": 121,
	"north-east":    112,
	"north-west":    186,
	"south-east":    95,
	"south-south":   123,
	"south-west":    137,
}

// GeographyFileRepository reads Nigerian geography records from JSON dataset files.
type GeographyFileRepository struct {
	jsonRepository           interfaces.JSONFileRepository
	statesPath               string
	zonesPath                string
	localGovernmentUnitsPath string
}

var _ interfaces.GeographyRepository = (*GeographyFileRepository)(nil)

// NewGeographyRepository constructs a file-backed geography repository.
func NewGeographyRepository(jsonRepository interfaces.JSONFileRepository, statesPath, zonesPath, localGovernmentUnitsPath string) (*GeographyFileRepository, error) {
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	cleanedStatesPath, err := validateGeographyDatasetPath("states", statesPath)
	if err != nil {
		return nil, err
	}
	cleanedZonesPath, err := validateGeographyDatasetPath("geopolitical zones", zonesPath)
	if err != nil {
		return nil, err
	}
	cleanedLocalGovernmentUnitsPath, err := validateGeographyDatasetPath("local government units", localGovernmentUnitsPath)
	if err != nil {
		return nil, err
	}

	return &GeographyFileRepository{
		jsonRepository:           jsonRepository,
		statesPath:               cleanedStatesPath,
		zonesPath:                cleanedZonesPath,
		localGovernmentUnitsPath: cleanedLocalGovernmentUnitsPath,
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

// ListLocalGovernmentUnits returns the ordered list of local-government units from the dataset files.
func (r *GeographyFileRepository) ListLocalGovernmentUnits(ctx context.Context) ([]models.LocalGovernmentUnit, error) {
	states, zones, units, err := r.loadLocalGovernmentUnitData(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateLocalGovernmentUnits(units, states, zones); err != nil {
		return nil, err
	}
	return cloneLocalGovernmentUnitList(units), nil
}

// ListLocalGovernmentUnitsByStateID returns the ordered list of local-government units for one state.
func (r *GeographyFileRepository) ListLocalGovernmentUnitsByStateID(ctx context.Context, stateID string) ([]models.LocalGovernmentUnit, error) {
	stateID = strings.TrimSpace(stateID)
	if stateID == "" || !stateIDPattern.MatchString(stateID) {
		return nil, fmt.Errorf("%w", interfaces.ErrStateNotFound)
	}

	states, zones, units, err := r.loadLocalGovernmentUnitData(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateLocalGovernmentUnits(units, states, zones); err != nil {
		return nil, err
	}
	if _, ok := findStateByID(states, stateID); !ok {
		return nil, fmt.Errorf("%w", interfaces.ErrStateNotFound)
	}

	filtered := make([]models.LocalGovernmentUnit, 0, expectedLocalGovernmentUnitCounts[stateID])
	for _, unit := range units {
		if unit.StateID == stateID {
			filtered = append(filtered, unit)
		}
	}
	return cloneLocalGovernmentUnitList(filtered), nil
}

// GetLocalGovernmentUnit returns a single local-government unit using its public slug identifier.
func (r *GeographyFileRepository) GetLocalGovernmentUnit(ctx context.Context, unitID string) (models.LocalGovernmentUnit, error) {
	unitID = strings.TrimSpace(unitID)
	if unitID == "" || !localGovernmentUnitIDPattern.MatchString(unitID) {
		return models.LocalGovernmentUnit{}, fmt.Errorf("%w", interfaces.ErrLocalGovernmentUnitNotFound)
	}

	states, zones, units, err := r.loadLocalGovernmentUnitData(ctx)
	if err != nil {
		return models.LocalGovernmentUnit{}, err
	}
	if err := validateLocalGovernmentUnits(units, states, zones); err != nil {
		return models.LocalGovernmentUnit{}, err
	}

	for _, unit := range units {
		if unit.ID == unitID {
			return cloneLocalGovernmentUnit(unit), nil
		}
	}

	return models.LocalGovernmentUnit{}, fmt.Errorf("%w", interfaces.ErrLocalGovernmentUnitNotFound)
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

func (r *GeographyFileRepository) loadLocalGovernmentUnitData(ctx context.Context) ([]models.State, []models.GeopoliticalZone, []models.LocalGovernmentUnit, error) {
	states, zones, err := r.loadGeographyData(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	units, err := r.loadLocalGovernmentUnitsOnly(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	return states, zones, units, nil
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

func (r *GeographyFileRepository) loadLocalGovernmentUnitsOnly(ctx context.Context) ([]models.LocalGovernmentUnit, error) {
	if r == nil || r.jsonRepository == nil {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var units []models.LocalGovernmentUnit
	if err := r.jsonRepository.Decode(ctx, r.localGovernmentUnitsPath, &units); err != nil {
		return nil, translateGeographyLoadError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if units == nil || len(units) == 0 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	return units, nil
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

func validateLocalGovernmentUnits(units []models.LocalGovernmentUnit, states []models.State, zones []models.GeopoliticalZone) error {
	if len(units) != 774 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	stateByID := make(map[string]models.State, len(states))
	stateOrder := make(map[string]int, len(states))
	for i, state := range states {
		stateByID[state.ID] = state
		stateOrder[state.ID] = i
	}

	zoneByID := make(map[string]models.GeopoliticalZone, len(zones))
	for _, zone := range zones {
		zoneByID[zone.ID] = zone
	}

	seenIDs := make(map[string]struct{}, len(units))
	seenStateNames := make(map[string]map[string]struct{}, len(states))
	stateCounts := make(map[string]int, len(states))
	zoneCounts := map[string]int{
		"north-central": 0,
		"north-east":    0,
		"north-west":    0,
		"south-east":    0,
		"south-south":   0,
		"south-west":    0,
	}
	seenStates := make(map[string]struct{}, len(states))
	fctSeen := make(map[string]string, 6)
	currentState := ""
	currentStateOrder := -1
	lastName := ""
	lastID := ""

	for _, unit := range units {
		if unit.ID == "" || unit.Name == "" || unit.StateID == "" || unit.CountryCode == "" || unit.AdministrativeType == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if unit.CountryCode != "NG" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		state, ok := stateByID[unit.StateID]
		if !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		zone, ok := zoneByID[state.GeopoliticalZoneID]
		if !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := validLocalGovernmentUnitTypes[unit.AdministrativeType]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if unit.StateID == "fct" {
			if unit.AdministrativeType != "area_council" {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		} else if unit.AdministrativeType != "local_government_area" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !localGovernmentUnitIDPattern.MatchString(unit.ID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !strings.HasPrefix(unit.ID, unit.StateID+"-") {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if wantID := unit.StateID + "-" + slugifyLocalGovernmentUnitName(unit.Name); unit.ID != wantID {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenIDs[unit.ID]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenStateNames[unit.StateID]; !ok {
			seenStateNames[unit.StateID] = make(map[string]struct{})
		}
		if _, ok := seenStateNames[unit.StateID][unit.Name]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}

		order := stateOrder[unit.StateID]
		if currentState == "" {
			currentState = unit.StateID
			currentStateOrder = order
			lastName = ""
			lastID = ""
		} else if unit.StateID != currentState {
			if order < currentStateOrder {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			if _, ok := seenStates[unit.StateID]; ok {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			seenStates[currentState] = struct{}{}
			currentState = unit.StateID
			currentStateOrder = order
			lastName = ""
			lastID = ""
		}
		if lastName != "" {
			cmp := strings.Compare(strings.ToLower(lastName), strings.ToLower(unit.Name))
			if cmp > 0 || (cmp == 0 && strings.Compare(lastID, unit.ID) > 0) {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}

		seenIDs[unit.ID] = struct{}{}
		seenStateNames[unit.StateID][unit.Name] = struct{}{}
		stateCounts[unit.StateID]++
		zoneCounts[zone.ID]++
		if unit.StateID == "fct" {
			fctSeen[unit.ID] = unit.Name
		}
		lastName = unit.Name
		lastID = unit.ID
	}
	if currentState != "" {
		seenStates[currentState] = struct{}{}
	}
	if len(seenStates) != len(states) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if !reflect.DeepEqual(stateCounts, expectedLocalGovernmentUnitCounts) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if !reflect.DeepEqual(zoneCounts, expectedLocalGovernmentUnitZoneCounts) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if !reflect.DeepEqual(fctSeen, map[string]string{
		"fct-abaji":           "Abaji",
		"fct-abuja-municipal": "Abuja Municipal",
		"fct-bwari":           "Bwari",
		"fct-gwagwalada":      "Gwagwalada",
		"fct-kuje":            "Kuje",
		"fct-kwali":           "Kwali",
	}) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	for stateID, want := range expectedLocalGovernmentUnitCounts {
		if stateCounts[stateID] != want {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	for zoneID, want := range expectedLocalGovernmentUnitZoneCounts {
		if zoneCounts[zoneID] != want {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
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

func cloneLocalGovernmentUnitList(units []models.LocalGovernmentUnit) []models.LocalGovernmentUnit {
	if len(units) == 0 {
		return make([]models.LocalGovernmentUnit, 0)
	}
	cloned := make([]models.LocalGovernmentUnit, len(units))
	copy(cloned, units)
	return cloned
}

func cloneLocalGovernmentUnit(unit models.LocalGovernmentUnit) models.LocalGovernmentUnit {
	return unit
}

func findStateByID(states []models.State, stateID string) (models.State, bool) {
	for _, state := range states {
		if state.ID == stateID {
			return state, true
		}
	}
	return models.State{}, false
}

func validateGeographyDatasetPath(label, datasetPath string) (string, error) {
	datasetPath = strings.TrimSpace(datasetPath)
	if datasetPath == "" {
		return "", fmt.Errorf("%s path is required", label)
	}
	if filepath.IsAbs(datasetPath) || filepath.VolumeName(datasetPath) != "" {
		return "", fmt.Errorf("%s path must remain relative", label)
	}
	cleanedPath := filepath.Clean(datasetPath)
	if cleanedPath == "." || cleanedPath == ".." || strings.HasPrefix(cleanedPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%s path must remain relative", label)
	}
	return cleanedPath, nil
}

func slugifyLocalGovernmentUnitName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(name, "-")
	name = regexp.MustCompile(`-+`).ReplaceAllString(name, "-")
	return strings.Trim(name, "-")
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
