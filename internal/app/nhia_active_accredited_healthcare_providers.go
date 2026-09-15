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
)

const (
	nhiaHCPExpectedCount                   = 6536
	nhiaHCPFirstAnchor                     = "ab-0001-p"
	nhiaHCPFirstAnchorName                 = "SANCTA MARIA SPECIALIST AND MAT."
	nhiaHCPFirstAnchorCode                 = "AB/0001/P"
	nhiaHCPFirstAnchorFacilityType         = "primary_and_secondary"
	nhiaHCPLastAnchor                      = "zf-0204-p"
	nhiaHCPLastAnchorName                  = "AMSAF SPECIALIST HOSPITA"
	nhiaHCPLastAnchorCode                  = "ZF/0204/P"
	nhiaHCPLastAnchorFacilityType          = "primary_and_secondary"
	nhiaHCPProviderCodeCheckID             = "fct-0001-p"
	nhiaHCPProviderCodeCheckName           = "WILDOT CLINIC"
	nhiaHCPProviderCodeCheckCode           = "FCT/0001/P"
	nhiaHCPProviderCodeCheckFacilityType   = "primary"
	nhiaHCPPrimaryCount                    = 4001
	nhiaHCPPrimaryAndSecondaryCount        = 2535
	nhiaHCPListingStatus                   = "active_accredited"
	nhiaHCPPrimaryFacilityType             = "primary"
	nhiaHCPPrimaryAndSecondaryFacilityType = "primary_and_secondary"
)

var startupNHIAHCPProviderCodePattern = regexp.MustCompile(`^[A-Z]{2,3}/[0-9]{4}/P$`)

func buildNHIAActiveAccreditedHealthcareProviderHandler(
	ctx context.Context,
	jsonRepository interfaces.JSONFileRepository,
	newRepository func(interfaces.JSONFileRepository, string) (interfaces.NHIAActiveAccreditedHealthcareProviderRepository, error),
	newService func(interfaces.NHIAActiveAccreditedHealthcareProviderRepository) (nhiaActiveAccreditedHealthcareProviderService, error),
	newHandler func(nhiaActiveAccreditedHealthcareProviderService) (*handlers.NHIAActiveAccreditedHealthcareProviderHandler, error),
) (nhiaActiveAccreditedHealthcareProviderService, *handlers.NHIAActiveAccreditedHealthcareProviderHandler, error) {
	service, err := buildNHIAActiveAccreditedHealthcareProviderServiceFromJSONRepository(ctx, jsonRepository, newRepository, newService)
	if err != nil {
		return nil, nil, err
	}
	if err := verifyNHIAActiveAccreditedHealthcareProviders(ctx, service); err != nil {
		return nil, nil, err
	}
	handler, err := newHandler(service)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize nhia active accredited healthcare provider handler: %w", err)
	}
	return service, handler, nil
}

// verifyNHIAActiveAccreditedHealthcareProviders uses bounded queries on the same cached service served by HTTP.
func verifyNHIAActiveAccreditedHealthcareProviders(ctx context.Context, service nhiaActiveAccreditedHealthcareProviderService) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if service == nil {
		return invalidNHIAHCPVerification("service is required")
	}

	page, err := service.ListNHIAActiveAccreditedHealthcareProviders(ctx, interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 1})
	if err != nil {
		return wrapNHIAHCPVerificationError("load verification page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if page.Records == nil || len(page.Records) != 1 || page.Page != 1 || page.PageSize != 1 || page.Total != nhiaHCPExpectedCount || page.TotalPages != nhiaHCPExpectedCount {
		return invalidNHIAHCPVerification("invalid verification page")
	}
	if err := validateStartupNHIAHCPAnchor(page.Records[0], nhiaHCPFirstAnchor, nhiaHCPFirstAnchorName, nhiaHCPFirstAnchorCode, nhiaHCPFirstAnchorFacilityType); err != nil {
		return err
	}

	for _, anchor := range []struct {
		id           string
		name         string
		providerCode string
		facilityType string
	}{
		{nhiaHCPFirstAnchor, nhiaHCPFirstAnchorName, nhiaHCPFirstAnchorCode, nhiaHCPFirstAnchorFacilityType},
		{nhiaHCPLastAnchor, nhiaHCPLastAnchorName, nhiaHCPLastAnchorCode, nhiaHCPLastAnchorFacilityType},
	} {
		record, err := service.GetNHIAActiveAccreditedHealthcareProvider(ctx, anchor.id)
		if err != nil {
			return wrapNHIAHCPVerificationError("retrieve anchor", err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := validateStartupNHIAHCPAnchor(record, anchor.id, anchor.name, anchor.providerCode, anchor.facilityType); err != nil {
			return err
		}
	}

	providerCodePage, err := service.ListNHIAActiveAccreditedHealthcareProviders(ctx, interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 1, ProviderCode: nhiaHCPProviderCodeCheckCode})
	if err != nil {
		return wrapNHIAHCPVerificationError("load provider code verification page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if providerCodePage.Records == nil || len(providerCodePage.Records) != 1 || providerCodePage.Page != 1 || providerCodePage.PageSize != 1 || providerCodePage.Total != 1 || providerCodePage.TotalPages != 1 {
		return invalidNHIAHCPVerification("invalid provider code verification page")
	}
	if err := validateStartupNHIAHCPAnchor(providerCodePage.Records[0], nhiaHCPProviderCodeCheckID, nhiaHCPProviderCodeCheckName, nhiaHCPProviderCodeCheckCode, nhiaHCPProviderCodeCheckFacilityType); err != nil {
		return err
	}

	facilityTotals := 0
	for _, filter := range []struct {
		facilityType string
		total        int
	}{
		{nhiaHCPPrimaryFacilityType, nhiaHCPPrimaryCount},
		{nhiaHCPPrimaryAndSecondaryFacilityType, nhiaHCPPrimaryAndSecondaryCount},
	} {
		result, err := service.ListNHIAActiveAccreditedHealthcareProviders(ctx, interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 1, FacilityType: filter.facilityType})
		if err != nil {
			return wrapNHIAHCPVerificationError("load facility type verification page", err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		facilityTotals += result.Total
		if result.Records == nil || len(result.Records) != 1 || result.Page != 1 || result.PageSize != 1 || result.Total != filter.total || result.TotalPages != filter.total || result.Records[0].FacilityType != filter.facilityType {
			return invalidNHIAHCPVerification("invalid facility type verification page")
		}
		if err := validateStartupNHIAHCP(result.Records[0]); err != nil {
			return err
		}
	}
	if facilityTotals != nhiaHCPExpectedCount {
		return invalidNHIAHCPVerification("facility type totals mismatch")
	}

	statusPage, err := service.ListNHIAActiveAccreditedHealthcareProviders(ctx, interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 1, ListingStatus: nhiaHCPListingStatus})
	if err != nil {
		return wrapNHIAHCPVerificationError("load listing status verification page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if statusPage.Records == nil || len(statusPage.Records) != 1 || statusPage.Page != 1 || statusPage.PageSize != 1 || statusPage.Total != nhiaHCPExpectedCount || statusPage.TotalPages != nhiaHCPExpectedCount || statusPage.Records[0].ListingStatus != nhiaHCPListingStatus {
		return invalidNHIAHCPVerification("invalid listing status verification page")
	}
	if err := validateStartupNHIAHCP(statusPage.Records[0]); err != nil {
		return err
	}

	beyond, err := service.ListNHIAActiveAccreditedHealthcareProviders(ctx, interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: nhiaHCPExpectedCount + 1, PageSize: 1})
	if err != nil {
		return wrapNHIAHCPVerificationError("load beyond-final page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if beyond.Records == nil || len(beyond.Records) != 0 || beyond.Page != nhiaHCPExpectedCount+1 || beyond.PageSize != 1 || beyond.Total != nhiaHCPExpectedCount || beyond.TotalPages != nhiaHCPExpectedCount {
		return invalidNHIAHCPVerification("invalid beyond-final page")
	}
	return nil
}

func validateStartupNHIAHCPAnchor(record models.NHIAActiveAccreditedHealthcareProvider, id, name, providerCode, facilityType string) error {
	if record.ID != id || record.Name != name || record.ProviderCode != providerCode || record.FacilityType != facilityType {
		return invalidNHIAHCPVerification("anchor mismatch")
	}
	return validateStartupNHIAHCP(record)
}

func validateStartupNHIAHCP(record models.NHIAActiveAccreditedHealthcareProvider) error {
	if record.ID == "" || len(record.ID) > models.NHIAActiveAccreditedHealthcareProviderIDMaxLength || !startupHealthFacilityIDPattern.MatchString(record.ID) || strings.TrimSpace(record.Name) == "" {
		return invalidNHIAHCPVerification("invalid record identity")
	}
	if record.CountryCode != "NG" || !startupNHIAHCPProviderCodePattern.MatchString(record.ProviderCode) || strings.TrimSpace(record.ProviderCode) != record.ProviderCode || record.ListingStatus != nhiaHCPListingStatus {
		return invalidNHIAHCPVerification("invalid record contract")
	}
	switch record.FacilityType {
	case nhiaHCPPrimaryFacilityType, nhiaHCPPrimaryAndSecondaryFacilityType:
		return nil
	default:
		return invalidNHIAHCPVerification("invalid record contract")
	}
}

func invalidNHIAHCPVerification(detail string) error {
	return fmt.Errorf("verify nhia active accredited healthcare providers: %w (%s)", interfaces.ErrInvalidDatasetFile, detail)
}

func wrapNHIAHCPVerificationError(operation string, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return invalidNHIAHCPVerification(operation + " failed")
}
