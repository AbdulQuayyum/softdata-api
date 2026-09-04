package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/config"
	"github.com/AbdulQuayyum/softdata-api/internal/handlers"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	fileRepo "github.com/AbdulQuayyum/softdata-api/internal/repository/file"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type geographyServiceStub struct {
	states    []models.State
	countries []models.CountryOrArea
	timeZones []models.TimeZone
	err       error

	stateListCalls    int
	countryListCalls  int
	timeZoneListCalls int
	lastCountryInput  services.CountryOrAreaListInput
	lastTimeZoneInput services.TimeZoneListInput
}

func (s *geographyServiceStub) ListStates(context.Context) ([]models.State, error) {
	s.stateListCalls++
	if s.err != nil {
		return nil, s.err
	}
	return append([]models.State(nil), s.states...), nil
}

func (s *geographyServiceStub) GetState(context.Context, string) (models.State, error) {
	return models.State{}, nil
}

func (s *geographyServiceStub) ListGeopoliticalZones(context.Context) ([]models.GeopoliticalZone, error) {
	return []models.GeopoliticalZone{}, nil
}

func (s *geographyServiceStub) GetGeopoliticalZone(context.Context, string) (models.GeopoliticalZone, error) {
	return models.GeopoliticalZone{}, nil
}

func (s *geographyServiceStub) ListLocalGovernmentUnits(context.Context) ([]models.LocalGovernmentUnit, error) {
	return []models.LocalGovernmentUnit{}, nil
}

func (s *geographyServiceStub) ListLocalGovernmentUnitsByState(context.Context, string) ([]models.LocalGovernmentUnit, error) {
	return []models.LocalGovernmentUnit{}, nil
}

func (s *geographyServiceStub) GetLocalGovernmentUnit(context.Context, string) (models.LocalGovernmentUnit, error) {
	return models.LocalGovernmentUnit{}, nil
}

func (s *geographyServiceStub) ListTimeZones(_ context.Context, input services.TimeZoneListInput) ([]models.TimeZone, error) {
	s.timeZoneListCalls++
	s.lastTimeZoneInput = input
	if s.err != nil {
		return nil, s.err
	}
	return append([]models.TimeZone(nil), s.timeZones...), nil
}

func (s *geographyServiceStub) GetTimeZone(context.Context, string) (models.TimeZone, error) {
	return models.TimeZone{}, nil
}

func (s *geographyServiceStub) ListCountriesAndAreas(_ context.Context, input services.CountryOrAreaListInput) ([]models.CountryOrArea, error) {
	s.countryListCalls++
	s.lastCountryInput = input
	if s.err != nil {
		return nil, s.err
	}
	return append([]models.CountryOrArea(nil), s.countries...), nil
}

func (s *geographyServiceStub) GetCountryOrArea(context.Context, string) (models.CountryOrArea, error) {
	return models.CountryOrArea{}, nil
}

func (s *geographyServiceStub) GetCountryProfile(context.Context, string) (models.CountryProfile, error) {
	return models.CountryProfile{}, nil
}

type geographyJSONRepoStub struct {
	root     string
	maxBytes int64
}

func (s *geographyJSONRepoStub) Decode(context.Context, string, any) error {
	return nil
}

type geographyRepoStub struct{}

func (s *geographyRepoStub) ListStates(context.Context) ([]models.State, error) {
	return nil, nil
}

func (s *geographyRepoStub) GetStateByID(context.Context, string) (models.State, error) {
	return models.State{}, nil
}

func (s *geographyRepoStub) ListGeopoliticalZones(context.Context) ([]models.GeopoliticalZone, error) {
	return nil, nil
}

func (s *geographyRepoStub) GetGeopoliticalZone(context.Context, string) (models.GeopoliticalZone, error) {
	return models.GeopoliticalZone{}, nil
}

func (s *geographyRepoStub) ListLocalGovernmentUnits(context.Context) ([]models.LocalGovernmentUnit, error) {
	return nil, nil
}

func (s *geographyRepoStub) ListLocalGovernmentUnitsByStateID(context.Context, string) ([]models.LocalGovernmentUnit, error) {
	return nil, nil
}

func (s *geographyRepoStub) GetLocalGovernmentUnit(context.Context, string) (models.LocalGovernmentUnit, error) {
	return models.LocalGovernmentUnit{}, nil
}

func (s *geographyRepoStub) ListLanguages(context.Context, interfaces.LanguageFilter) ([]models.Language, error) {
	return nil, nil
}

func (s *geographyRepoStub) GetLanguage(context.Context, string) (models.Language, error) {
	return models.Language{}, nil
}

func (s *geographyRepoStub) ListCountryLanguages(context.Context, interfaces.CountryLanguageFilter) ([]models.CountryLanguage, error) {
	return nil, nil
}

func (s *geographyRepoStub) ListTimeZones(context.Context, interfaces.TimeZoneFilter) ([]models.TimeZone, error) {
	return nil, nil
}

func (s *geographyRepoStub) GetTimeZone(context.Context, string) (models.TimeZone, error) {
	return models.TimeZone{}, nil
}

func (s *geographyRepoStub) ListCountriesAndAreas(context.Context, interfaces.CountryOrAreaFilter) ([]models.CountryOrArea, error) {
	return nil, nil
}

func (s *geographyRepoStub) GetCountryOrArea(context.Context, string) (models.CountryOrArea, error) {
	return models.CountryOrArea{}, nil
}

func validGeographyStatesFixture(t *testing.T) []models.State {
	t.Helper()

	path := filepath.Clean("../../datasets/geography/states.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()

	var states []models.State
	if err := dec.Decode(&states); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("state fixture contains trailing json")
	}
	return states
}

func TestBuildGeographyHandlerPassesConfiguredDatasetArgs(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Datasets: config.DatasetConfig{
			Path:         "/tmp/custom-datasets",
			JSONMaxBytes: 12345,
		},
	}
	countries := loadApprovedCountryFixture(t)
	timeZonesFixture := loadApprovedTimeZoneFixture(t)

	var gotRoot string
	var gotMaxBytes int64
	var gotStatesPath string
	var gotZonesPath string
	var gotLgasPath string
	var gotTimeZonesPath string
	var gotCountriesPath string

	handler, err := buildGeographyHandler(context.Background(), cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			gotRoot = root
			gotMaxBytes = maxBytes
			return &geographyJSONRepoStub{root: root, maxBytes: maxBytes}, nil
		},
		func(repository interfaces.JSONFileRepository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, languagesPath, countryLanguagesPath string) (interfaces.GeographyRepository, error) {
			gotStatesPath = statesPath
			gotZonesPath = zonesPath
			gotLgasPath = localGovernmentUnitsPath
			gotTimeZonesPath = timeZonesPath
			gotCountriesPath = countriesAndAreasPath
			return &geographyRepoStub{}, nil
		},
		func(repository interfaces.GeographyRepository) (geographyService, error) {
			return &geographyServiceStub{
				states:    validGeographyStatesFixture(t),
				countries: countries,
				timeZones: timeZonesFixture,
			}, nil
		},
		func(service geographyService) (*handlers.GeographyHandler, error) {
			return handlers.NewGeographyHandler(service)
		},
	)
	if err != nil {
		t.Fatalf("buildGeographyHandler() error = %v", err)
	}
	if handler == nil {
		t.Fatal("buildGeographyHandler() returned nil handler")
	}
	if gotRoot != cfg.Datasets.Path {
		t.Fatalf("unexpected dataset root: %q", gotRoot)
	}
	if gotMaxBytes != cfg.Datasets.JSONMaxBytes {
		t.Fatalf("unexpected JSON max bytes: %d", gotMaxBytes)
	}
	if gotStatesPath != geographyStatesRelativePath {
		t.Fatalf("unexpected states path: %q", gotStatesPath)
	}
	if gotZonesPath != geographyGeopoliticalZonesRelativePath {
		t.Fatalf("unexpected zones path: %q", gotZonesPath)
	}
	if gotLgasPath != geographyLocalGovernmentUnitsRelativePath {
		t.Fatalf("unexpected lgas path: %q", gotLgasPath)
	}
	if gotTimeZonesPath != geographyTimeZonesRelativePath {
		t.Fatalf("unexpected time zones path: %q", gotTimeZonesPath)
	}
	if gotCountriesPath != geographyCountriesAndAreasRelativePath {
		t.Fatalf("unexpected countries path: %q", gotCountriesPath)
	}
}

func TestBuildGeographyHandlerValidFixturePassesStartupVerification(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "geography"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Clean("../../datasets/geography/states.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "states.json"), data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	zones, err := os.ReadFile(filepath.Clean("../../datasets/geography/geopolitical_zones.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "geopolitical_zones.json"), zones, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	countries, err := os.ReadFile(filepath.Clean("../../datasets/geography/countries_and_areas.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "countries_and_areas.json"), countries, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	timeZones, err := os.ReadFile(filepath.Clean("../../datasets/geography/time_zones.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "time_zones.json"), timeZones, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg := &config.Config{
		Datasets: config.DatasetConfig{
			Path:         root,
			JSONMaxBytes: int64(len(data)+len(zones)+len(countries)+len(timeZones)) + 1024,
		},
	}

	countriesFixture := loadApprovedCountryFixture(t)
	timeZonesFixture := loadApprovedTimeZoneFixture(t)

	handler, err := buildGeographyHandler(context.Background(), cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			return fileRepo.NewJSONRepository(root, maxBytes)
		},
		func(repository interfaces.JSONFileRepository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, languagesPath, countryLanguagesPath string) (interfaces.GeographyRepository, error) {
			return fileRepo.NewGeographyRepository(repository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, "geography/languages.json", "geography/country_languages.json")
		},
		func(repository interfaces.GeographyRepository) (geographyService, error) {
			return &geographyServiceStub{
				states:    validGeographyStatesFixture(t),
				countries: countriesFixture,
				timeZones: timeZonesFixture,
			}, nil
		},
		func(service geographyService) (*handlers.GeographyHandler, error) {
			return handlers.NewGeographyHandler(service)
		},
	)
	if err != nil {
		t.Fatalf("buildGeographyHandler() error = %v", err)
	}
	if handler == nil {
		t.Fatal("buildGeographyHandler() returned nil handler")
	}
}

func TestBuildGeographyHandlerVerifiesCountriesAndAreasSnapshot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "geography"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	states, err := os.ReadFile(filepath.Clean("../../datasets/geography/states.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "states.json"), states, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	zones, err := os.ReadFile(filepath.Clean("../../datasets/geography/geopolitical_zones.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "geopolitical_zones.json"), zones, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	countries, err := os.ReadFile(filepath.Clean("../../datasets/geography/countries_and_areas.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "countries_and_areas.json"), countries, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	timeZones := loadApprovedTimeZoneFixture(t)
	timeZonesData := mustMarshalTimeZonesFixture(t, timeZones)
	if err := os.WriteFile(filepath.Join(root, "geography", "time_zones.json"), timeZonesData, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg := &config.Config{
		Datasets: config.DatasetConfig{
			Path:         root,
			JSONMaxBytes: int64(len(states)+len(zones)+len(countries)+len(timeZonesData)) + 1024,
		},
	}

	var stub geographyServiceStub
	stub.timeZones = loadApprovedTimeZoneFixture(t)
	handler, err := buildGeographyHandler(context.Background(), cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			return fileRepo.NewJSONRepository(root, maxBytes)
		},
		func(repository interfaces.JSONFileRepository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, languagesPath, countryLanguagesPath string) (interfaces.GeographyRepository, error) {
			return fileRepo.NewGeographyRepository(repository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, "geography/languages.json", "geography/country_languages.json")
		},
		func(repository interfaces.GeographyRepository) (geographyService, error) {
			stub.states = validGeographyStatesFixture(t)
			stub.countries = loadApprovedCountryFixture(t)
			return &stub, nil
		},
		func(service geographyService) (*handlers.GeographyHandler, error) {
			return handlers.NewGeographyHandler(service)
		},
	)
	if err != nil {
		t.Fatalf("buildGeographyHandler() error = %v", err)
	}
	if handler == nil {
		t.Fatal("buildGeographyHandler() returned nil handler")
	}
	if stub.stateListCalls != 1 {
		t.Fatalf("unexpected state verification calls: %d", stub.stateListCalls)
	}
	if stub.countryListCalls != 1 {
		t.Fatalf("unexpected country verification calls: %d", stub.countryListCalls)
	}
	if stub.lastCountryInput != (services.CountryOrAreaListInput{}) {
		t.Fatalf("unexpected country input: %#v", stub.lastCountryInput)
	}
	if stub.timeZoneListCalls != 1 {
		t.Fatalf("unexpected time zone verification calls: %d", stub.timeZoneListCalls)
	}
	if stub.lastTimeZoneInput != (services.TimeZoneListInput{}) {
		t.Fatalf("unexpected time zone input: %#v", stub.lastTimeZoneInput)
	}
}

func TestBuildGeographyHandlerFailsSafelyForInvalidDatasets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(string) (config.DatasetConfig, error)
	}{
		{
			name: "missing root",
			setup: func(root string) (config.DatasetConfig, error) {
				return config.DatasetConfig{Path: filepath.Join(root, "missing"), JSONMaxBytes: 1024}, nil
			},
		},
		{
			name: "missing states file",
			setup: func(root string) (config.DatasetConfig, error) {
				if err := os.MkdirAll(root, 0o755); err != nil {
					return config.DatasetConfig{}, err
				}
				return config.DatasetConfig{Path: root, JSONMaxBytes: 1024}, nil
			},
		},
		{
			name: "malformed json",
			setup: func(root string) (config.DatasetConfig, error) {
				if err := os.MkdirAll(filepath.Join(root, "geography"), 0o755); err != nil {
					return config.DatasetConfig{}, err
				}
				if err := os.WriteFile(filepath.Join(root, "geography", "states.json"), []byte("{bad"), 0o600); err != nil {
					return config.DatasetConfig{}, err
				}
				return config.DatasetConfig{Path: root, JSONMaxBytes: 1024}, nil
			},
		},
		{
			name: "oversized file",
			setup: func(root string) (config.DatasetConfig, error) {
				if err := os.MkdirAll(filepath.Join(root, "geography"), 0o755); err != nil {
					return config.DatasetConfig{}, err
				}
				if err := os.WriteFile(filepath.Join(root, "geography", "states.json"), []byte(strings.Repeat("a", 2048)), 0o600); err != nil {
					return config.DatasetConfig{}, err
				}
				return config.DatasetConfig{Path: root, JSONMaxBytes: 1}, nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			datasetCfg, err := tc.setup(root)
			if err != nil {
				t.Fatalf("setup() error = %v", err)
			}
			cfg := &config.Config{Datasets: datasetCfg}

			handler, err := buildGeographyHandler(context.Background(), cfg,
				func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
					return fileRepo.NewJSONRepository(root, maxBytes)
				},
				func(repository interfaces.JSONFileRepository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, languagesPath, countryLanguagesPath string) (interfaces.GeographyRepository, error) {
					return fileRepo.NewGeographyRepository(repository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, "geography/languages.json", "geography/country_languages.json")
				},
				func(repository interfaces.GeographyRepository) (geographyService, error) {
					return services.NewGeographyService(repository)
				},
				func(service geographyService) (*handlers.GeographyHandler, error) {
					return handlers.NewGeographyHandler(service)
				},
			)
			if err == nil {
				t.Fatalf("buildGeographyHandler() handler = %#v, want error", handler)
			}
			if strings.Contains(err.Error(), root) {
				t.Fatalf("error leaked dataset root: %v", err)
			}
		})
	}
}

func TestBuildGeographyHandlerRejectsWrongRecordCount(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "geography"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Clean("../../datasets/geography/states.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	zones, err := os.ReadFile(filepath.Clean("../../datasets/geography/geopolitical_zones.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	countries, err := os.ReadFile(filepath.Clean("../../datasets/geography/countries_and_areas.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var states []map[string]any
	if err := json.Unmarshal(data, &states); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if err := writeStatesFixture(filepath.Join(root, "geography", "states.json"), states[:36]); err != nil {
		t.Fatalf("writeStatesFixture() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "geopolitical_zones.json"), zones, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "countries_and_areas.json"), countries, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	timeZones, err := os.ReadFile(filepath.Clean("../../datasets/geography/time_zones.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "time_zones.json"), timeZones, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg := &config.Config{Datasets: config.DatasetConfig{Path: root, JSONMaxBytes: int64(len(data)+len(zones)+len(countries)+len(timeZones)) + 1024}}
	_, err = buildGeographyHandler(context.Background(), cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			return fileRepo.NewJSONRepository(root, maxBytes)
		},
		func(repository interfaces.JSONFileRepository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, languagesPath, countryLanguagesPath string) (interfaces.GeographyRepository, error) {
			return fileRepo.NewGeographyRepository(repository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, "geography/languages.json", "geography/country_languages.json")
		},
		func(repository interfaces.GeographyRepository) (geographyService, error) {
			return services.NewGeographyService(repository)
		},
		func(service geographyService) (*handlers.GeographyHandler, error) {
			return handlers.NewGeographyHandler(service)
		},
	)
	if err == nil {
		t.Fatal("buildGeographyHandler() error = nil, want record-count failure")
	}
}

func TestBuildGeographyHandlerRejectsWrongFCTComposition(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "geography"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Clean("../../datasets/geography/states.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	zones, err := os.ReadFile(filepath.Clean("../../datasets/geography/geopolitical_zones.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	countries, err := os.ReadFile(filepath.Clean("../../datasets/geography/countries_and_areas.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var states []map[string]any
	if err := json.Unmarshal(data, &states); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	for i := range states {
		if states[i]["id"] == "fct" {
			states[i]["id"] = "federal-capital-territory"
			break
		}
	}
	if err := writeStatesFixture(filepath.Join(root, "geography", "states.json"), states); err != nil {
		t.Fatalf("writeStatesFixture() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "geopolitical_zones.json"), zones, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "countries_and_areas.json"), countries, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	timeZones, err := os.ReadFile(filepath.Clean("../../datasets/geography/time_zones.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "time_zones.json"), timeZones, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg := &config.Config{Datasets: config.DatasetConfig{Path: root, JSONMaxBytes: int64(len(data)+len(zones)+len(countries)+len(timeZones)) + 1024}}
	_, err = buildGeographyHandler(context.Background(), cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			return fileRepo.NewJSONRepository(root, maxBytes)
		},
		func(repository interfaces.JSONFileRepository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, languagesPath, countryLanguagesPath string) (interfaces.GeographyRepository, error) {
			return fileRepo.NewGeographyRepository(repository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, "geography/languages.json", "geography/country_languages.json")
		},
		func(repository interfaces.GeographyRepository) (geographyService, error) {
			return services.NewGeographyService(repository)
		},
		func(service geographyService) (*handlers.GeographyHandler, error) {
			return handlers.NewGeographyHandler(service)
		},
	)
	if err == nil {
		t.Fatal("buildGeographyHandler() error = nil, want composition failure")
	}
}

func TestRunStartupCleanupClosesEachResourceOnce(t *testing.T) {
	t.Parallel()

	var postgresCalls atomic.Int32
	var redisCalls atomic.Int32
	err := errors.New("startup failure")
	cleanup := []func(){
		func() { postgresCalls.Add(1) },
		func() { redisCalls.Add(1) },
	}

	runStartupCleanup(&err, cleanup)

	if postgresCalls.Load() != 1 || redisCalls.Load() != 1 {
		t.Fatalf("unexpected cleanup calls: postgres=%d redis=%d", postgresCalls.Load(), redisCalls.Load())
	}
}

func TestBuildGeographyHandlerPropagatesContextCancellation(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Datasets: config.DatasetConfig{
			Path:         t.TempDir(),
			JSONMaxBytes: 1,
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := buildGeographyHandler(ctx, cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			return &geographyJSONRepoStub{root: root, maxBytes: maxBytes}, nil
		},
		func(repository interfaces.JSONFileRepository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, languagesPath, countryLanguagesPath string) (interfaces.GeographyRepository, error) {
			return &geographyRepoStub{}, nil
		},
		func(repository interfaces.GeographyRepository) (geographyService, error) {
			return &geographyServiceStub{states: validGeographyStatesFixture(t)}, nil
		},
		func(service geographyService) (*handlers.GeographyHandler, error) {
			return handlers.NewGeographyHandler(service)
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("buildGeographyHandler() error = %v, want context.Canceled", err)
	}
}

func TestBuildGeographyHandlerRejectsInvalidCountryOrAreaFixtures(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "geography"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	states, err := os.ReadFile(filepath.Clean("../../datasets/geography/states.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "states.json"), states, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	zones, err := os.ReadFile(filepath.Clean("../../datasets/geography/geopolitical_zones.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "geopolitical_zones.json"), zones, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	timeZones, err := os.ReadFile(filepath.Clean("../../datasets/geography/time_zones.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "time_zones.json"), timeZones, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	countries := loadApprovedCountryFixture(t)
	tests := []struct {
		name string
		mut  func([]models.CountryOrArea) []models.CountryOrArea
	}{
		{
			name: "wrong record count",
			mut: func(countries []models.CountryOrArea) []models.CountryOrArea {
				return append([]models.CountryOrArea(nil), countries[:247]...)
			},
		},
		{
			name: "missing nigeria",
			mut: func(countries []models.CountryOrArea) []models.CountryOrArea {
				out := append([]models.CountryOrArea(nil), countries...)
				for i := range out {
					if out[i].ID == "ng" {
						out[i].ID = "xx"
						out[i].Name = "Example"
						out[i].Alpha2Code = "XX"
						out[i].Alpha3Code = "XXX"
						out[i].NumericCode = "999"
						break
					}
				}
				return out
			},
		},
		{
			name: "algeria numeric loses leading zero",
			mut: func(countries []models.CountryOrArea) []models.CountryOrArea {
				out := append([]models.CountryOrArea(nil), countries...)
				for i := range out {
					if out[i].Name == "Algeria" {
						out[i].NumericCode = "12"
						break
					}
				}
				return out
			},
		},
		{
			name: "kosovo present",
			mut: func(countries []models.CountryOrArea) []models.CountryOrArea {
				out := append([]models.CountryOrArea(nil), countries...)
				return append(out, models.CountryOrArea{
					ID:          "xk",
					Name:        "Kosovo",
					Alpha2Code:  "XK",
					Alpha3Code:  "XKX",
					NumericCode: "926",
				})
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			writeCountries := tc.mut(countries)
			data, err := json.Marshal(writeCountries)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if err := os.WriteFile(filepath.Join(root, "geography", "countries_and_areas.json"), data, 0o600); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}

			cfg := &config.Config{Datasets: config.DatasetConfig{Path: root, JSONMaxBytes: int64(len(states)+len(zones)+len(data)+len(timeZones)) + 1024}}
			_, err = buildGeographyHandler(context.Background(), cfg,
				func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
					return fileRepo.NewJSONRepository(root, maxBytes)
				},
				func(repository interfaces.JSONFileRepository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, languagesPath, countryLanguagesPath string) (interfaces.GeographyRepository, error) {
					return fileRepo.NewGeographyRepository(repository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, "geography/languages.json", "geography/country_languages.json")
				},
				func(repository interfaces.GeographyRepository) (geographyService, error) {
					return services.NewGeographyService(repository)
				},
				func(service geographyService) (*handlers.GeographyHandler, error) {
					return handlers.NewGeographyHandler(service)
				},
			)
			if err == nil {
				t.Fatal("buildGeographyHandler() error = nil, want country manifest failure")
			}
		})
	}
}

func TestBuildGeographyHandlerRejectsInvalidTimeZoneFixtures(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "geography"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	states, err := os.ReadFile(filepath.Clean("../../datasets/geography/states.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "states.json"), states, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	zones, err := os.ReadFile(filepath.Clean("../../datasets/geography/geopolitical_zones.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "geopolitical_zones.json"), zones, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	countries := loadApprovedCountryFixture(t)
	countriesData, err := json.Marshal(countries)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "geography", "countries_and_areas.json"), countriesData, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	timeZones := loadApprovedTimeZoneFixture(t)
	badTimeZones := append([]models.TimeZone(nil), timeZones[:311]...)
	badTimeZonesData := mustMarshalTimeZonesFixture(t, badTimeZones)
	if err := os.WriteFile(filepath.Join(root, "geography", "time_zones.json"), badTimeZonesData, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg := &config.Config{Datasets: config.DatasetConfig{Path: root, JSONMaxBytes: int64(len(states)+len(zones)+len(countriesData)+len(badTimeZonesData)) + 1024}}
	_, err = buildGeographyHandler(context.Background(), cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			return fileRepo.NewJSONRepository(root, maxBytes)
		},
		func(repository interfaces.JSONFileRepository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, languagesPath, countryLanguagesPath string) (interfaces.GeographyRepository, error) {
			return fileRepo.NewGeographyRepository(repository, statesPath, zonesPath, localGovernmentUnitsPath, timeZonesPath, countriesAndAreasPath, "geography/languages.json", "geography/country_languages.json")
		},
		func(repository interfaces.GeographyRepository) (geographyService, error) {
			return services.NewGeographyService(repository)
		},
		func(service geographyService) (*handlers.GeographyHandler, error) {
			return handlers.NewGeographyHandler(service)
		},
	)
	if err == nil {
		t.Fatal("buildGeographyHandler() error = nil, want time zone failure")
	}
}

func writeStatesFixture(path string, states []map[string]any) error {
	data, err := json.Marshal(states)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func loadApprovedCountryFixture(t *testing.T) []models.CountryOrArea {
	t.Helper()

	path := filepath.Clean("../../datasets/geography/countries_and_areas.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()

	var countries []models.CountryOrArea
	if err := dec.Decode(&countries); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("fixture contains trailing json")
	}
	return countries
}

func loadApprovedTimeZoneFixture(t *testing.T) []models.TimeZone {
	t.Helper()

	path := filepath.Clean("../../datasets/geography/time_zones.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()

	var timeZones []models.TimeZone
	if err := dec.Decode(&timeZones); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("time zone fixture contains trailing json")
	}
	return timeZones
}

func mustMarshalTimeZonesFixture(t *testing.T, timeZones []models.TimeZone) []byte {
	t.Helper()

	data, err := json.Marshal(timeZones)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return data
}
