package app

import (
	"context"
	"fmt"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

func buildNEMAZonalTerritorialOperationOfficeServiceFromJSONRepository(
	ctx context.Context,
	jsonRepository interfaces.JSONFileRepository,
	newRepository func(interfaces.JSONFileRepository, string) (interfaces.NEMAZonalTerritorialOperationOfficeRepository, error),
	newService func(interfaces.NEMAZonalTerritorialOperationOfficeRepository) (nemaZonalTerritorialOperationOfficeService, error),
) (nemaZonalTerritorialOperationOfficeService, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	if newRepository == nil {
		return nil, fmt.Errorf("nema zonal territorial operation office repository constructor is required")
	}
	if newService == nil {
		return nil, fmt.Errorf("nema zonal territorial operation office service constructor is required")
	}
	repository, err := newRepository(jsonRepository, emergencyNEMAZonalTerritorialOperationOfficesRelativePath)
	if err != nil {
		return nil, fmt.Errorf("initialize nema zonal territorial operation office repository: %w", err)
	}
	service, err := newService(repository)
	if err != nil {
		return nil, fmt.Errorf("initialize nema zonal territorial operation office service: %w", err)
	}
	return service, nil
}

type nemaZonalTerritorialOperationOfficeService interface {
	ListNEMAZonalTerritorialOperationOffices(context.Context, services.NEMAZonalTerritorialOperationOfficeQuery) (services.NEMAZonalTerritorialOperationOfficeListResult, error)
	GetNEMAZonalTerritorialOperationOffice(context.Context, string) (models.NEMAZonalTerritorialOperationOffice, error)
}
