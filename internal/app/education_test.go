package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/config"
	"github.com/AbdulQuayyum/softdata-api/internal/handlers"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	fileRepo "github.com/AbdulQuayyum/softdata-api/internal/repository/file"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type educationServiceStub struct {
	universities []models.University
	err          error
	calls        int
	lastInput    services.UniversityListInput
}

func (s *educationServiceStub) ListUniversities(ctx context.Context, input services.UniversityListInput) ([]models.University, error) {
	s.calls++
	s.lastInput = input
	if s.err != nil {
		return nil, s.err
	}
	return append([]models.University(nil), s.universities...), nil
}

func (s *educationServiceStub) GetUniversity(context.Context, string) (models.University, error) {
	return models.University{}, nil
}

func (s *educationServiceStub) ListCollegesOfEducation(context.Context, services.CollegeOfEducationListInput) ([]models.CollegeOfEducation, error) {
	return nil, nil
}

func (s *educationServiceStub) GetCollegeOfEducation(context.Context, string) (models.CollegeOfEducation, error) {
	return models.CollegeOfEducation{}, nil
}

type educationRepositoryStub struct{}

func (s *educationRepositoryStub) ListUniversities(context.Context, interfaces.UniversityFilter) ([]models.University, error) {
	return nil, nil
}

func (s *educationRepositoryStub) GetUniversityByID(context.Context, string) (models.University, error) {
	return models.University{}, nil
}

func (s *educationRepositoryStub) ListCollegesOfEducation(context.Context, interfaces.CollegeOfEducationFilter) ([]models.CollegeOfEducation, error) {
	return nil, nil
}

func (s *educationRepositoryStub) GetCollegeOfEducation(context.Context, string) (models.CollegeOfEducation, error) {
	return models.CollegeOfEducation{}, nil
}

type educationJSONRepoStub struct{}

func (s *educationJSONRepoStub) Decode(context.Context, string, any) error {
	return nil
}

func loadApprovedUniversities(t *testing.T) []models.University {
	t.Helper()

	data, err := os.ReadFile(filepath.Clean("../../datasets/education/universities.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var universities []models.University
	if err := json.Unmarshal(data, &universities); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return universities
}

func writeEducationFixture(path string, universities []models.University) error {
	data, err := json.Marshal(universities)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func TestBuildEducationHandlerPassesConfiguredDatasetArgs(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Datasets: config.DatasetConfig{
			Path:         "/tmp/custom-datasets",
			JSONMaxBytes: 54321,
		},
	}

	var gotRoot string
	var gotMaxBytes int64
	var gotUniversityPath string
	var gotCollegePath string
	handler, err := buildEducationHandler(context.Background(), cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			gotRoot = root
			gotMaxBytes = maxBytes
			return &educationJSONRepoStub{}, nil
		},
		func(repository interfaces.JSONFileRepository, universitiesPath, collegesOfEducationPath string) (interfaces.EducationRepository, error) {
			gotUniversityPath = universitiesPath
			gotCollegePath = collegesOfEducationPath
			return &educationRepositoryStub{}, nil
		},
		func(repository interfaces.EducationRepository) (educationService, error) {
			return &educationServiceStub{universities: loadApprovedUniversities(t)}, nil
		},
		func(service educationService) (*handlers.EducationHandler, error) {
			return handlers.NewEducationHandler(service)
		},
	)
	if err != nil {
		t.Fatalf("buildEducationHandler() error = %v", err)
	}
	if handler == nil {
		t.Fatal("buildEducationHandler() returned nil handler")
	}
	if gotRoot != cfg.Datasets.Path {
		t.Fatalf("unexpected dataset root: %q", gotRoot)
	}
	if gotMaxBytes != cfg.Datasets.JSONMaxBytes {
		t.Fatalf("unexpected JSON max bytes: %d", gotMaxBytes)
	}
	if gotUniversityPath != educationUniversitiesRelativePath {
		t.Fatalf("unexpected universities path: %q", gotUniversityPath)
	}
	if gotCollegePath != educationCollegesOfEducationRelativePath {
		t.Fatalf("unexpected colleges path: %q", gotCollegePath)
	}
}

func TestBuildEducationHandlerValidFixturePassesStartupVerification(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "education"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Clean("../../datasets/education/universities.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "education", "universities.json"), data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg := &config.Config{
		Datasets: config.DatasetConfig{
			Path:         root,
			JSONMaxBytes: int64(len(data)) + 1024,
		},
	}

	handler, err := buildEducationHandler(context.Background(), cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			return fileRepo.NewJSONRepository(root, maxBytes)
		},
		func(repository interfaces.JSONFileRepository, universitiesPath, collegesOfEducationPath string) (interfaces.EducationRepository, error) {
			return fileRepo.NewEducationRepository(repository, universitiesPath, collegesOfEducationPath)
		},
		func(repository interfaces.EducationRepository) (educationService, error) {
			return services.NewEducationService(repository)
		},
		func(service educationService) (*handlers.EducationHandler, error) {
			return handlers.NewEducationHandler(service)
		},
	)
	if err != nil {
		t.Fatalf("buildEducationHandler() error = %v", err)
	}
	if handler == nil {
		t.Fatal("buildEducationHandler() returned nil handler")
	}
}

func TestBuildEducationHandlerFailsSafelyForInvalidDatasets(t *testing.T) {
	t.Parallel()

	fixture := loadApprovedUniversities(t)
	tests := []struct {
		name string
		set  func(root string) error
	}{
		{
			name: "missing file",
			set: func(root string) error {
				return os.MkdirAll(root, 0o755)
			},
		},
		{
			name: "malformed json",
			set: func(root string) error {
				if err := os.MkdirAll(filepath.Join(root, "education"), 0o755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(root, "education", "universities.json"), []byte("{bad"), 0o600)
			},
		},
		{
			name: "wrong record count",
			set: func(root string) error {
				if err := os.MkdirAll(filepath.Join(root, "education"), 0o755); err != nil {
					return err
				}
				return writeEducationFixture(filepath.Join(root, "education", "universities.json"), fixture[:327])
			},
		},
		{
			name: "wrong composition",
			set: func(root string) error {
				if err := os.MkdirAll(filepath.Join(root, "education"), 0o755); err != nil {
					return err
				}
				mutated := append([]models.University(nil), fixture...)
				if mutated[0].OwnershipType == "federal" {
					mutated[0].OwnershipType = "private"
				} else {
					mutated[0].OwnershipType = "federal"
				}
				return writeEducationFixture(filepath.Join(root, "education", "universities.json"), mutated)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := tc.set(root); err != nil {
				t.Fatalf("setup() error = %v", err)
			}

			cfg := &config.Config{Datasets: config.DatasetConfig{Path: root, JSONMaxBytes: 4096}}
			_, err := buildEducationHandler(context.Background(), cfg,
				func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
					return fileRepo.NewJSONRepository(root, maxBytes)
				},
				func(repository interfaces.JSONFileRepository, universitiesPath, collegesOfEducationPath string) (interfaces.EducationRepository, error) {
					return fileRepo.NewEducationRepository(repository, universitiesPath, collegesOfEducationPath)
				},
				func(repository interfaces.EducationRepository) (educationService, error) {
					return services.NewEducationService(repository)
				},
				func(service educationService) (*handlers.EducationHandler, error) {
					return handlers.NewEducationHandler(service)
				},
			)
			if err == nil {
				t.Fatal("buildEducationHandler() error = nil, want failure")
			}
			if strings.Contains(err.Error(), root) {
				t.Fatalf("error leaked dataset root: %v", err)
			}
		})
	}
}

func TestBuildEducationHandlerPropagatesContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := &config.Config{
		Datasets: config.DatasetConfig{
			Path:         t.TempDir(),
			JSONMaxBytes: 1,
		},
	}

	_, err := buildEducationHandler(ctx, cfg,
		func(root string, maxBytes int64) (interfaces.JSONFileRepository, error) {
			return &educationJSONRepoStub{}, nil
		},
		func(repository interfaces.JSONFileRepository, universitiesPath, collegesOfEducationPath string) (interfaces.EducationRepository, error) {
			return &educationRepositoryStub{}, nil
		},
		func(repository interfaces.EducationRepository) (educationService, error) {
			return &educationServiceStub{universities: loadApprovedUniversities(t)}, nil
		},
		func(service educationService) (*handlers.EducationHandler, error) {
			return handlers.NewEducationHandler(service)
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("buildEducationHandler() error = %v, want context.Canceled", err)
	}
}

func TestBuildEducationHandlerVerifiesThroughServiceAbstraction(t *testing.T) {
	t.Parallel()

	service := &educationServiceStub{universities: loadApprovedUniversities(t)}
	handler, err := buildEducationHandlerFromJSONRepository(context.Background(), &educationJSONRepoStub{},
		func(repository interfaces.JSONFileRepository, universitiesPath, collegesOfEducationPath string) (interfaces.EducationRepository, error) {
			return &educationRepositoryStub{}, nil
		},
		func(repository interfaces.EducationRepository) (educationService, error) {
			return service, nil
		},
		func(service educationService) (*handlers.EducationHandler, error) {
			return handlers.NewEducationHandler(service)
		},
	)
	if err != nil {
		t.Fatalf("buildEducationHandlerFromJSONRepository() error = %v", err)
	}
	if handler == nil {
		t.Fatal("buildEducationHandlerFromJSONRepository() returned nil handler")
	}
	if service.calls != 1 {
		t.Fatalf("expected startup verification to call list-all once, got %d", service.calls)
	}
}
