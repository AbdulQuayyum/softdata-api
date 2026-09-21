package app

import (
	"context"
	"errors"
	"io/fs"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/datasets"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	fileRepo "github.com/AbdulQuayyum/softdata-api/internal/repository/file"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type nemaOfficeAppJSONRepository struct{}

func (nemaOfficeAppJSONRepository) Decode(context.Context, string, any) error { return nil }

type nemaOfficeAppRepository struct{}

func (nemaOfficeAppRepository) ListNEMAZonalTerritorialOperationOffices(context.Context, interfaces.NEMAZonalTerritorialOperationOfficeQuery) (interfaces.NEMAZonalTerritorialOperationOfficeListResult, error) {
	return interfaces.NEMAZonalTerritorialOperationOfficeListResult{}, nil
}

func (nemaOfficeAppRepository) GetNEMAZonalTerritorialOperationOffice(context.Context, string) (models.NEMAZonalTerritorialOperationOffice, error) {
	return models.NEMAZonalTerritorialOperationOffice{}, nil
}

func TestNEMAZonalTerritorialOperationOfficeDependencyConstructionUsesDatasetPath(t *testing.T) {
	var gotPath string
	service, err := buildNEMAZonalTerritorialOperationOfficeServiceFromJSONRepository(context.Background(), nemaOfficeAppJSONRepository{},
		func(repository interfaces.JSONFileRepository, recordsPath string) (interfaces.NEMAZonalTerritorialOperationOfficeRepository, error) {
			if repository == nil {
				t.Fatal("nil JSON repository")
			}
			gotPath = recordsPath
			return nemaOfficeAppRepository{}, nil
		},
		func(repository interfaces.NEMAZonalTerritorialOperationOfficeRepository) (nemaZonalTerritorialOperationOfficeService, error) {
			return services.NewNEMAZonalTerritorialOperationOfficeService(repository)
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if service == nil || gotPath != emergencyNEMAZonalTerritorialOperationOfficesRelativePath {
		t.Fatalf("service=%v path=%q", service, gotPath)
	}
}

func TestNEMAZonalTerritorialOperationOfficeDependencyConstructionPreservesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := buildNEMAZonalTerritorialOperationOfficeServiceFromJSONRepository(ctx, nemaOfficeAppJSONRepository{},
		func(interfaces.JSONFileRepository, string) (interfaces.NEMAZonalTerritorialOperationOfficeRepository, error) {
			t.Fatal("repository constructor should not be called")
			return nil, nil
		},
		func(interfaces.NEMAZonalTerritorialOperationOfficeRepository) (nemaZonalTerritorialOperationOfficeService, error) {
			return nil, nil
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestNEMAZonalTerritorialOperationOfficeEmbeddedDatasetFallback(t *testing.T) {
	jsonRepository, err := fileRepo.NewEmbeddedJSONRepository(datasets.Files(), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := fileRepo.NewNEMAZonalTerritorialOperationOfficeRepository(jsonRepository, emergencyNEMAZonalTerritorialOperationOfficesRelativePath)
	if err != nil {
		t.Fatal(err)
	}
	service, err := services.NewNEMAZonalTerritorialOperationOfficeService(repository)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.ListNEMAZonalTerritorialOperationOffices(context.Background(), services.NEMAZonalTerritorialOperationOfficeQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 17 || len(result.Records) != 17 {
		t.Fatalf("embedded result mismatch: %#v", result)
	}
	if _, err := fs.ReadFile(datasets.Files(), "metadata/emergency/nema_zonal_territorial_operation_offices_reconciliation/index.json"); err == nil {
		t.Fatal("reconciliation metadata must not be embedded for runtime")
	}
}
