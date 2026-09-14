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
	nhiaSSHIAExpectedCount    = 37
	nhiaSSHIAFirstAnchor      = "abia-state-health-insurance-agency-abshia"
	nhiaSSHIAFirstAnchorName  = "Abia State Health Insurance Agency (ABSHIA)"
	nhiaSSHIACheckState       = "lagos"
	nhiaSSHIACheckStateAnchor = "lashma"
	nhiaSSHIACheckStateName   = "LASHMA"
	nhiaSSHIAFCTAnchor        = "fct-health-insurance-scheme-fhis"
	nhiaSSHIAFCTAnchorName    = "FCT Health Insurance Scheme (FHIS)"
	nhiaSSHIALastAnchor       = "zamchema"
	nhiaSSHIALastAnchorName   = "ZAMCHEMA"
	nhiaSSHIAOrganisationType = "state_social_health_insurance_agency"
)

func buildNHIASSHIAHandler(
	ctx context.Context,
	jsonRepository interfaces.JSONFileRepository,
	newRepository func(interfaces.JSONFileRepository, string) (interfaces.NHIAStateSocialHealthInsuranceAgencyRepository, error),
	newService func(interfaces.NHIAStateSocialHealthInsuranceAgencyRepository) (nhiaSSHIAService, error),
	newHandler func(nhiaSSHIAService) (*handlers.NHIAStateSocialHealthInsuranceAgencyHandler, error),
) (nhiaSSHIAService, *handlers.NHIAStateSocialHealthInsuranceAgencyHandler, error) {
	service, err := buildNHIASSHIAServiceFromJSONRepository(ctx, jsonRepository, newRepository, newService)
	if err != nil {
		return nil, nil, err
	}
	if err := verifyNHIASSHIAs(ctx, service); err != nil {
		return nil, nil, err
	}
	handler, err := newHandler(service)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize nhia state social health insurance agency handler: %w", err)
	}
	return service, handler, nil
}

// verifyNHIASSHIAs uses bounded queries on the same cached service served by HTTP.
func verifyNHIASSHIAs(ctx context.Context, service nhiaSSHIAService) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if service == nil {
		return invalidNHIASSHIAVerification("service is required")
	}
	page, err := service.ListNHIAStateSocialHealthInsuranceAgencies(ctx, interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 1, PageSize: 1})
	if err != nil {
		return wrapNHIASSHIAVerificationError("load verification page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if page.Records == nil || len(page.Records) != 1 || page.Page != 1 || page.PageSize != 1 || page.Total != nhiaSSHIAExpectedCount || page.TotalPages != nhiaSSHIAExpectedCount {
		return invalidNHIASSHIAVerification("invalid verification page")
	}
	if err := validateStartupNHIASSHIA(page.Records[0]); err != nil {
		return err
	}
	for _, anchor := range []struct {
		id   string
		name string
	}{
		{nhiaSSHIAFirstAnchor, nhiaSSHIAFirstAnchorName},
		{nhiaSSHIALastAnchor, nhiaSSHIALastAnchorName},
	} {
		record, err := service.GetNHIAStateSocialHealthInsuranceAgency(ctx, anchor.id)
		if err != nil {
			return wrapNHIASSHIAVerificationError("retrieve anchor", err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if record.ID != anchor.id || record.Name != anchor.name {
			return invalidNHIASSHIAVerification("anchor mismatch")
		}
		if err := validateStartupNHIASSHIA(record); err != nil {
			return err
		}
	}
	for _, filter := range []struct {
		stateID string
		id      string
		name    string
	}{
		{nhiaSSHIACheckState, nhiaSSHIACheckStateAnchor, nhiaSSHIACheckStateName},
		{"fct", nhiaSSHIAFCTAnchor, nhiaSSHIAFCTAnchorName},
	} {
		result, err := service.ListNHIAStateSocialHealthInsuranceAgencies(ctx, interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 1, PageSize: 1, StateID: filter.stateID})
		if err != nil {
			return wrapNHIASSHIAVerificationError("load state verification page", err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if result.Records == nil || len(result.Records) != 1 || result.Page != 1 || result.PageSize != 1 || result.Total != 1 || result.TotalPages != 1 {
			return invalidNHIASSHIAVerification("invalid state verification page")
		}
		record := result.Records[0]
		if record.ID != filter.id || record.Name != filter.name || record.StateID != filter.stateID {
			return invalidNHIASSHIAVerification("state verification mismatch")
		}
		if err := validateStartupNHIASSHIA(record); err != nil {
			return err
		}
	}
	beyond, err := service.ListNHIAStateSocialHealthInsuranceAgencies(ctx, interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: nhiaSSHIAExpectedCount + 1, PageSize: 1})
	if err != nil {
		return wrapNHIASSHIAVerificationError("load beyond-final page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if beyond.Records == nil || len(beyond.Records) != 0 || beyond.Page != nhiaSSHIAExpectedCount+1 || beyond.PageSize != 1 || beyond.Total != nhiaSSHIAExpectedCount || beyond.TotalPages != nhiaSSHIAExpectedCount {
		return invalidNHIASSHIAVerification("invalid beyond-final page")
	}
	return nil
}

func validateStartupNHIASSHIA(record models.NHIAStateSocialHealthInsuranceAgency) error {
	if record.ID == "" || len(record.ID) > models.NHIAStateSocialHealthInsuranceAgencyIDMaxLength || !startupHealthFacilityIDPattern.MatchString(record.ID) || strings.TrimSpace(record.Name) == "" || strings.TrimSpace(record.StateID) == "" {
		return invalidNHIASSHIAVerification("invalid record identity")
	}
	if record.CountryCode != "NG" || record.OrganisationType != nhiaSSHIAOrganisationType {
		return invalidNHIASSHIAVerification("invalid record contract")
	}
	return nil
}

func invalidNHIASSHIAVerification(detail string) error {
	return fmt.Errorf("verify nhia state social health insurance agencies: %w (%s)", interfaces.ErrInvalidDatasetFile, detail)
}

func wrapNHIASSHIAVerificationError(operation string, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return invalidNHIASSHIAVerification(operation + " failed")
}
