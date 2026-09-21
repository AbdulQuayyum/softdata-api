package app

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/handlers"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

const (
	nemaZonalTerritorialOperationOfficeExpectedCount = 17
	nemaZonalTerritorialOperationOfficeType          = "zonal_territorial_operation_office"
	nemaZonalTerritorialOperationOfficeFirstID       = "nema-abuja-zonal-territorial-operation-office"
	nemaZonalTerritorialOperationOfficeFirstName     = "NEMA Abuja Office"
	nemaZonalTerritorialOperationOfficeFirstState    = "fct"
	nemaZonalTerritorialOperationOfficeLastID        = "nema-yola-zonal-territorial-operation-office"
	nemaZonalTerritorialOperationOfficeLastName      = "NEMA Yola Office"
	nemaZonalTerritorialOperationOfficeLastState     = "adamawa"
)

var startupNEMAZonalTerritorialOperationOfficeIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

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

func buildNEMAZonalTerritorialOperationOfficeHandler(
	ctx context.Context,
	jsonRepository interfaces.JSONFileRepository,
	newRepository func(interfaces.JSONFileRepository, string) (interfaces.NEMAZonalTerritorialOperationOfficeRepository, error),
	newService func(interfaces.NEMAZonalTerritorialOperationOfficeRepository) (nemaZonalTerritorialOperationOfficeService, error),
	newHandler func(nemaZonalTerritorialOperationOfficeService) (*handlers.NEMAZonalTerritorialOperationOfficeHandler, error),
) (nemaZonalTerritorialOperationOfficeService, *handlers.NEMAZonalTerritorialOperationOfficeHandler, error) {
	service, err := buildNEMAZonalTerritorialOperationOfficeServiceFromJSONRepository(ctx, jsonRepository, newRepository, newService)
	if err != nil {
		return nil, nil, err
	}
	if err := verifyNEMAZonalTerritorialOperationOffices(ctx, service); err != nil {
		return nil, nil, err
	}
	handler, err := newHandler(service)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize nema zonal territorial operation office handler: %w", err)
	}
	return service, handler, nil
}

// verifyNEMAZonalTerritorialOperationOffices uses bounded queries on the same cached service served by HTTP.
func verifyNEMAZonalTerritorialOperationOffices(ctx context.Context, service nemaZonalTerritorialOperationOfficeService) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if service == nil {
		return invalidNEMAZonalTerritorialOperationOfficeVerification("service is required")
	}

	page, err := service.ListNEMAZonalTerritorialOperationOffices(ctx, interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: 1, PageSize: 1})
	if err != nil {
		return wrapNEMAZonalTerritorialOperationOfficeVerificationError("load verification page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if page.Records == nil || len(page.Records) != 1 || page.Page != 1 || page.PageSize != 1 || page.Total != nemaZonalTerritorialOperationOfficeExpectedCount || page.TotalPages != nemaZonalTerritorialOperationOfficeExpectedCount {
		return invalidNEMAZonalTerritorialOperationOfficeVerification("invalid verification page")
	}

	first := nemaZonalTerritorialOperationOfficeAnchor{id: nemaZonalTerritorialOperationOfficeFirstID, name: nemaZonalTerritorialOperationOfficeFirstName, stateID: nemaZonalTerritorialOperationOfficeFirstState}
	last := nemaZonalTerritorialOperationOfficeAnchor{id: nemaZonalTerritorialOperationOfficeLastID, name: nemaZonalTerritorialOperationOfficeLastName, stateID: nemaZonalTerritorialOperationOfficeLastState}
	if err := validateStartupNEMAZonalTerritorialOperationOfficeAnchor(page.Records[0], first); err != nil {
		return err
	}
	for _, anchor := range []nemaZonalTerritorialOperationOfficeAnchor{first, last} {
		record, err := service.GetNEMAZonalTerritorialOperationOffice(ctx, anchor.id)
		if err != nil {
			return wrapNEMAZonalTerritorialOperationOfficeVerificationError("retrieve anchor", err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := validateStartupNEMAZonalTerritorialOperationOfficeAnchor(record, anchor); err != nil {
			return err
		}
	}

	officeType, err := service.ListNEMAZonalTerritorialOperationOffices(ctx, interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: 1, PageSize: 1, OfficeType: nemaZonalTerritorialOperationOfficeType})
	if err != nil {
		return wrapNEMAZonalTerritorialOperationOfficeVerificationError("load office-type verification page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if officeType.Records == nil || len(officeType.Records) != 1 || officeType.Total != nemaZonalTerritorialOperationOfficeExpectedCount || officeType.TotalPages != nemaZonalTerritorialOperationOfficeExpectedCount || officeType.Records[0].OfficeType != nemaZonalTerritorialOperationOfficeType {
		return invalidNEMAZonalTerritorialOperationOfficeVerification("invalid office-type verification page")
	}
	if err := validateStartupNEMAZonalTerritorialOperationOffice(officeType.Records[0]); err != nil {
		return err
	}

	state, err := service.ListNEMAZonalTerritorialOperationOffices(ctx, interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: 1, PageSize: 1, StateID: first.stateID})
	if err != nil {
		return wrapNEMAZonalTerritorialOperationOfficeVerificationError("load state verification page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if state.Records == nil || len(state.Records) != 1 || state.Total != 1 || state.TotalPages != 1 {
		return invalidNEMAZonalTerritorialOperationOfficeVerification("invalid state verification page")
	}
	if err := validateStartupNEMAZonalTerritorialOperationOfficeAnchor(state.Records[0], first); err != nil {
		return err
	}

	beyond, err := service.ListNEMAZonalTerritorialOperationOffices(ctx, interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: nemaZonalTerritorialOperationOfficeExpectedCount + 1, PageSize: 1})
	if err != nil {
		return wrapNEMAZonalTerritorialOperationOfficeVerificationError("load beyond-final page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if beyond.Records == nil || len(beyond.Records) != 0 || beyond.Page != nemaZonalTerritorialOperationOfficeExpectedCount+1 || beyond.PageSize != 1 || beyond.Total != nemaZonalTerritorialOperationOfficeExpectedCount || beyond.TotalPages != nemaZonalTerritorialOperationOfficeExpectedCount {
		return invalidNEMAZonalTerritorialOperationOfficeVerification("invalid beyond-final page")
	}
	return nil
}

type nemaZonalTerritorialOperationOfficeAnchor struct {
	id      string
	name    string
	stateID string
}

func validateStartupNEMAZonalTerritorialOperationOfficeAnchor(record models.NEMAZonalTerritorialOperationOffice, anchor nemaZonalTerritorialOperationOfficeAnchor) error {
	if record.ID != anchor.id || record.Name != anchor.name || record.StateID != anchor.stateID {
		return invalidNEMAZonalTerritorialOperationOfficeVerification("anchor mismatch")
	}
	return validateStartupNEMAZonalTerritorialOperationOffice(record)
}

func validateStartupNEMAZonalTerritorialOperationOffice(record models.NEMAZonalTerritorialOperationOffice) error {
	if record.ID == "" || len(record.ID) > models.NEMAZonalTerritorialOperationOfficeIDMaxLength || !startupNEMAZonalTerritorialOperationOfficeIDPattern.MatchString(record.ID) {
		return invalidNEMAZonalTerritorialOperationOfficeVerification("invalid record identity")
	}
	if strings.TrimSpace(record.Name) == "" || strings.TrimSpace(record.StateID) == "" || record.CountryCode != "NG" {
		return invalidNEMAZonalTerritorialOperationOfficeVerification("invalid record contract")
	}
	if record.OfficeType != nemaZonalTerritorialOperationOfficeType {
		return invalidNEMAZonalTerritorialOperationOfficeVerification("invalid record office type")
	}
	return nil
}

func invalidNEMAZonalTerritorialOperationOfficeVerification(detail string) error {
	return fmt.Errorf("verify nema zonal territorial operation offices: %w (%s)", interfaces.ErrInvalidDatasetFile, detail)
}

func wrapNEMAZonalTerritorialOperationOfficeVerificationError(operation string, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return invalidNEMAZonalTerritorialOperationOfficeVerification(operation + " failed")
}

type nemaZonalTerritorialOperationOfficeService interface {
	ListNEMAZonalTerritorialOperationOffices(context.Context, services.NEMAZonalTerritorialOperationOfficeQuery) (services.NEMAZonalTerritorialOperationOfficeListResult, error)
	GetNEMAZonalTerritorialOperationOffice(context.Context, string) (models.NEMAZonalTerritorialOperationOffice, error)
}
