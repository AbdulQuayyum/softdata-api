package app

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const (
	healthFacilityExpectedCount = 50654
	healthFacilityFirstAnchor   = "222-cliford-medical-center-abia-abia-aba-north-cea7a6ea-db7e-4845-8057-5caf45dc26c1"
	healthFacilityLastAnchor    = "zurmi-town-health-post-zamfara-zamfara-zurmi-662c2adb-b082-407c-b7dd-e6830b9f3b46"
)

var startupHealthFacilityIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func verifyHealthFacilityDataset(ctx context.Context, service healthFacilityService) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if service == nil {
		return fmt.Errorf("verify health facility dataset: service is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	firstPage, err := service.ListHealthFacilities(ctx, interfaces.HealthFacilityQuery{Page: 1, PageSize: 1})
	if err != nil {
		return wrapHealthFacilityVerificationError("load first page", err)
	}
	if firstPage.Facilities == nil || len(firstPage.Facilities) != 1 || firstPage.Page != 1 || firstPage.PageSize != 1 || firstPage.Total != healthFacilityExpectedCount || firstPage.TotalPages != healthFacilityExpectedCount {
		return invalidHealthFacilityVerification("invalid first-page pagination or total")
	}
	first := firstPage.Facilities[0]
	if err := validateStartupHealthFacility(first); err != nil {
		return err
	}

	firstByID, err := service.GetHealthFacility(ctx, first.ID)
	if err != nil {
		return wrapHealthFacilityVerificationError("retrieve first facility", err)
	}
	if firstByID.ID != first.ID {
		return invalidHealthFacilityVerification("first facility lookup mismatch")
	}

	anchor, err := service.GetHealthFacility(ctx, healthFacilityLastAnchor)
	if err != nil {
		return wrapHealthFacilityVerificationError("retrieve anchor facility", err)
	}
	if anchor.ID != healthFacilityLastAnchor {
		return invalidHealthFacilityVerification("anchor facility lookup mismatch")
	}
	if _, err := service.GetHealthFacility(ctx, healthFacilityFirstAnchor); err != nil {
		return wrapHealthFacilityVerificationError("retrieve first retained anchor", err)
	}

	if first.LGAID != "" {
		filtered, err := service.ListHealthFacilities(ctx, interfaces.HealthFacilityQuery{
			Page: 1, PageSize: 1, StateID: first.StateID, LGAID: first.LGAID,
		})
		if err != nil {
			return wrapHealthFacilityVerificationError("verify first facility geography", err)
		}
		if filtered.Facilities == nil || len(filtered.Facilities) == 0 || filtered.Facilities[0].ID != first.ID {
			return invalidHealthFacilityVerification("first facility state/LGA relationship is invalid")
		}
	}

	beyond, err := service.ListHealthFacilities(ctx, interfaces.HealthFacilityQuery{Page: healthFacilityExpectedCount + 1, PageSize: 1})
	if err != nil {
		return wrapHealthFacilityVerificationError("load beyond-final page", err)
	}
	if beyond.Facilities == nil || len(beyond.Facilities) != 0 || beyond.Page != healthFacilityExpectedCount+1 || beyond.PageSize != 1 || beyond.Total != healthFacilityExpectedCount || beyond.TotalPages != healthFacilityExpectedCount {
		return invalidHealthFacilityVerification("invalid beyond-final page")
	}
	return nil
}

func validateStartupHealthFacility(facility models.HealthFacility) error {
	if facility.ID == "" || !startupHealthFacilityIDPattern.MatchString(facility.ID) || strings.TrimSpace(facility.Name) == "" || facility.CountryCode != "NG" || facility.StateID == "" {
		return invalidHealthFacilityVerification("invalid first facility identity")
	}
	if _, ok := approvedUniversityStateIDs[facility.StateID]; !ok {
		return invalidHealthFacilityVerification("invalid first facility state")
	}
	switch facility.FacilityType {
	case "primary-health-centre", "clinic", "health-post", "general-hospital", "other", "teaching-hospital", "specialist-hospital":
	default:
		return invalidHealthFacilityVerification("invalid first facility type")
	}
	if facility.FacilityLevel != "" && facility.FacilityLevel != "primary" && facility.FacilityLevel != "secondary" && facility.FacilityLevel != "tertiary" {
		return invalidHealthFacilityVerification("invalid first facility level")
	}
	if facility.OwnershipType != "" {
		switch facility.OwnershipType {
		case "federal", "state", "local-government", "private", "military", "other-public":
		default:
			return invalidHealthFacilityVerification("invalid first facility ownership")
		}
	}
	if facility.Latitude == nil || facility.Longitude == nil || *facility.Latitude < 4.281710 || *facility.Latitude > 13.865239 || *facility.Longitude < 2.707790 || *facility.Longitude > 14.636383 {
		return invalidHealthFacilityVerification("invalid first facility coordinates")
	}
	return nil
}

func invalidHealthFacilityVerification(detail string) error {
	return fmt.Errorf("verify health facility dataset: %w (%s)", interfaces.ErrInvalidDatasetFile, detail)
}

func wrapHealthFacilityVerificationError(operation string, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return fmt.Errorf("verify health facility dataset: %s: %w", operation, err)
}
