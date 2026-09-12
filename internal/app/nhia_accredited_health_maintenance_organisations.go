package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/handlers"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const (
	nhiaAccreditedHMOExpectedCount = 94
	nhiaAccreditedHMOFirstAnchor   = "a-and-m-healthcare-trust-limited-102"
	nhiaAccreditedHMOLastAnchor    = "zuma-health-trust-28"
)

func buildNHIAAccreditedHMOHandler(
	ctx context.Context,
	jsonRepository interfaces.JSONFileRepository,
	newRepository func(interfaces.JSONFileRepository, string) (interfaces.NHIAAccreditedHealthMaintenanceOrganisationRepository, error),
	newService func(interfaces.NHIAAccreditedHealthMaintenanceOrganisationRepository) (nhiaAccreditedHMOService, error),
	newHandler func(nhiaAccreditedHMOService) (*handlers.NHIAAccreditedHealthMaintenanceOrganisationHandler, error),
) (nhiaAccreditedHMOService, *handlers.NHIAAccreditedHealthMaintenanceOrganisationHandler, error) {
	service, err := buildNHIAAccreditedHMOServiceFromJSONRepository(ctx, jsonRepository, newRepository, newService)
	if err != nil {
		return nil, nil, err
	}
	if err := verifyNHIAAccreditedHMOs(ctx, service); err != nil {
		return nil, nil, err
	}
	handler, err := newHandler(service)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize nhia accredited health maintenance organisation handler: %w", err)
	}
	return service, handler, nil
}

// verifyNHIAAccreditedHMOs uses bounded queries on the same cached service served by HTTP.
func verifyNHIAAccreditedHMOs(ctx context.Context, service nhiaAccreditedHMOService) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if service == nil {
		return invalidNHIAAccreditedHMOVerification("service is required")
	}
	for _, status := range []string{"", "accredited"} {
		page, err := service.ListNHIAAccreditedHealthMaintenanceOrganisations(ctx, interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Page: 1, PageSize: 1, AccreditationStatus: status})
		if err != nil {
			return wrapNHIAAccreditedHMOVerificationError("load verification page", err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if page.Records == nil || len(page.Records) != 1 || page.Page != 1 || page.PageSize != 1 || page.Total != nhiaAccreditedHMOExpectedCount || page.TotalPages != nhiaAccreditedHMOExpectedCount {
			return invalidNHIAAccreditedHMOVerification("invalid verification page")
		}
		if err := validateStartupNHIAAccreditedHMO(page.Records[0]); err != nil {
			return err
		}
		if status != "" && page.Records[0].AccreditationStatus != status {
			return invalidNHIAAccreditedHMOVerification("status filter mismatch")
		}
	}
	for _, id := range []string{nhiaAccreditedHMOFirstAnchor, nhiaAccreditedHMOLastAnchor} {
		record, err := service.GetNHIAAccreditedHealthMaintenanceOrganisation(ctx, id)
		if err != nil {
			return wrapNHIAAccreditedHMOVerificationError("retrieve anchor", err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if record.ID != id {
			return invalidNHIAAccreditedHMOVerification("anchor mismatch")
		}
		if err := validateStartupNHIAAccreditedHMO(record); err != nil {
			return err
		}
	}
	beyond, err := service.ListNHIAAccreditedHealthMaintenanceOrganisations(ctx, interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Page: nhiaAccreditedHMOExpectedCount + 1, PageSize: 1})
	if err != nil {
		return wrapNHIAAccreditedHMOVerificationError("load beyond-final page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if beyond.Records == nil || len(beyond.Records) != 0 || beyond.Page != nhiaAccreditedHMOExpectedCount+1 || beyond.PageSize != 1 || beyond.Total != nhiaAccreditedHMOExpectedCount || beyond.TotalPages != nhiaAccreditedHMOExpectedCount {
		return invalidNHIAAccreditedHMOVerification("invalid beyond-final page")
	}
	return nil
}

func validateStartupNHIAAccreditedHMO(record models.NHIAAccreditedHealthMaintenanceOrganisation) error {
	if record.ID == "" || len(record.ID) > models.NHIAAccreditedHealthMaintenanceOrganisationIDMaxLength || !startupHealthFacilityIDPattern.MatchString(record.ID) || strings.TrimSpace(record.Name) == "" {
		return invalidNHIAAccreditedHMOVerification("invalid record identity")
	}
	if record.CountryCode != "NG" || record.OrganisationType != "health_maintenance_organisation" || record.AccreditationStatus != "accredited" || strings.TrimSpace(record.HMOID) == "" || strings.TrimSpace(record.HMOID) != record.HMOID {
		return invalidNHIAAccreditedHMOVerification("invalid record contract")
	}
	return nil
}

func invalidNHIAAccreditedHMOVerification(detail string) error {
	return fmt.Errorf("verify nhia accredited health maintenance organisations: %w (%s)", interfaces.ErrInvalidDatasetFile, detail)
}

func wrapNHIAAccreditedHMOVerificationError(operation string, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return invalidNHIAAccreditedHMOVerification(operation + " failed")
}
