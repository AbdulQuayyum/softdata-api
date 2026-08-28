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
	states []models.State
	err    error
}

func (s *geographyServiceStub) ListStates(context.Context) ([]models.State, error) {
	if s.err != nil {
		return nil, s.err
	}
	return append([]models.State(nil), s.states...), nil
}

func (s *geographyServiceStub) GetState(context.Context, string) (models.State, error) {
	return models.State{}, nil
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

func validGeographyStatesFixture() []models.State {
	states := make([]models.State, 0, 37)
	for i := 1; i <= 36; i++ {
		states = append(states, models.State{
			ID:                 strings.ToLower("state"),
			Name:               "State",
			OfficialName:       "State",
			AdministrativeType: "state",
			Capital:            "Capital",
			GeopoliticalZoneID: "north-central",
			CountryCode:        "NG",
		})
	}
	states = append(states, models.State{
		ID:                 "fct",
		Name:               "Federal Capital Territory",
		OfficialName:       "Federal Capital Territory",
		AdministrativeType: "federal_capital_territory",
		Capital:            "Abuja",
		GeopoliticalZoneID: "north-central",
		CountryCode:        "NG",
	})
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

	var gotRoot string
	var gotMaxBytes int64
	var gotStatesPath string

	handler, err := buildGeographyHandler(context.Background(), cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			gotRoot = root
			gotMaxBytes = maxBytes
			return &geographyJSONRepoStub{root: root, maxBytes: maxBytes}, nil
		},
		func(repository interfaces.JSONFileRepository, statesPath string) (interfaces.GeographyRepository, error) {
			gotStatesPath = statesPath
			return &geographyRepoStub{}, nil
		},
		func(repository interfaces.GeographyRepository) (geographyService, error) {
			return &geographyServiceStub{states: validGeographyStatesFixture()}, nil
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

	cfg := &config.Config{
		Datasets: config.DatasetConfig{
			Path:         root,
			JSONMaxBytes: int64(len(data)) + 1024,
		},
	}

	handler, err := buildGeographyHandler(context.Background(), cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			return fileRepo.NewJSONRepository(root, maxBytes)
		},
		func(repository interfaces.JSONFileRepository, statesPath string) (interfaces.GeographyRepository, error) {
			return fileRepo.NewGeographyRepository(repository, statesPath)
		},
		func(repository interfaces.GeographyRepository) (geographyService, error) {
			return services.NewGeographyService(repository)
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
				func(repository interfaces.JSONFileRepository, statesPath string) (interfaces.GeographyRepository, error) {
					return fileRepo.NewGeographyRepository(repository, statesPath)
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

	cfg := &config.Config{Datasets: config.DatasetConfig{Path: root, JSONMaxBytes: int64(len(data)) + 1024}}
	_, err = buildGeographyHandler(context.Background(), cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			return fileRepo.NewJSONRepository(root, maxBytes)
		},
		func(repository interfaces.JSONFileRepository, statesPath string) (interfaces.GeographyRepository, error) {
			return fileRepo.NewGeographyRepository(repository, statesPath)
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

	cfg := &config.Config{Datasets: config.DatasetConfig{Path: root, JSONMaxBytes: int64(len(data)) + 1024}}
	_, err = buildGeographyHandler(context.Background(), cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			return fileRepo.NewJSONRepository(root, maxBytes)
		},
		func(repository interfaces.JSONFileRepository, statesPath string) (interfaces.GeographyRepository, error) {
			return fileRepo.NewGeographyRepository(repository, statesPath)
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
		func(repository interfaces.JSONFileRepository, statesPath string) (interfaces.GeographyRepository, error) {
			return &geographyRepoStub{}, nil
		},
		func(repository interfaces.GeographyRepository) (geographyService, error) {
			return &geographyServiceStub{states: validGeographyStatesFixture()}, nil
		},
		func(service geographyService) (*handlers.GeographyHandler, error) {
			return handlers.NewGeographyHandler(service)
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("buildGeographyHandler() error = %v, want context.Canceled", err)
	}
}

func writeStatesFixture(path string, states []map[string]any) error {
	data, err := json.Marshal(states)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
