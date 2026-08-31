package interfaces

import (
	"context"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

// GeographyRepository defines state lookup operations backed by a geography dataset.
type GeographyRepository interface {
	ListStates(ctx context.Context) ([]models.State, error)
	GetStateByID(ctx context.Context, stateID string) (models.State, error)
	ListGeopoliticalZones(ctx context.Context) ([]models.GeopoliticalZone, error)
	GetGeopoliticalZone(ctx context.Context, zoneID string) (models.GeopoliticalZone, error)
	ListLocalGovernmentUnits(ctx context.Context) ([]models.LocalGovernmentUnit, error)
	ListLocalGovernmentUnitsByStateID(ctx context.Context, stateID string) ([]models.LocalGovernmentUnit, error)
	GetLocalGovernmentUnit(ctx context.Context, unitID string) (models.LocalGovernmentUnit, error)
	ListCountriesAndAreas(ctx context.Context, filter CountryOrAreaFilter) ([]models.CountryOrArea, error)
	GetCountryOrArea(ctx context.Context, countryOrAreaID string) (models.CountryOrArea, error)
}

// CountryOrAreaFilter narrows the countries-and-areas dataset by geography codes.
type CountryOrAreaFilter struct {
	RegionCode    string
	SubregionCode string
}
