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
	emergencyServiceContactExpectedCount = 5
	emergencyServiceContactFirstID       = "federal-fire-service-fire-112-national"
	emergencyServiceContactFirstService  = "Federal Fire Service Emergency Response"
	emergencyServiceContactFirstAgency   = "Federal Fire Service"
	emergencyServiceContactFirstType     = "fire"
	emergencyServiceContactFirstContact  = "short_code"
	emergencyServiceContactFirstValue    = "112"
	emergencyServiceContactLastID        = "nigerian-communications-commission-general-emergency-112-national"
	emergencyServiceContactLastService   = "112 Emergency Number"
	emergencyServiceContactLastAgency    = "Nigerian Communications Commission"
	emergencyServiceContactLastType      = "general_emergency"
	emergencyServiceContactLastContact   = "short_code"
	emergencyServiceContactLastValue     = "112"
)

var startupEmergencyServiceContactIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func buildEmergencyServiceContactServiceFromJSONRepository(
	ctx context.Context,
	jsonRepository interfaces.JSONFileRepository,
	newRepository func(interfaces.JSONFileRepository, string) (interfaces.EmergencyServiceContactRepository, error),
	newService func(interfaces.EmergencyServiceContactRepository) (emergencyServiceContactService, error),
) (emergencyServiceContactService, error) {
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
		return nil, fmt.Errorf("emergency service contact repository constructor is required")
	}
	if newService == nil {
		return nil, fmt.Errorf("emergency service contact service constructor is required")
	}
	repository, err := newRepository(jsonRepository, emergencyServiceContactsRelativePath)
	if err != nil {
		return nil, fmt.Errorf("initialize emergency service contact repository: %w", err)
	}
	service, err := newService(repository)
	if err != nil {
		return nil, fmt.Errorf("initialize emergency service contact service: %w", err)
	}
	return service, nil
}

func buildEmergencyServiceContactHandler(
	ctx context.Context,
	jsonRepository interfaces.JSONFileRepository,
	newRepository func(interfaces.JSONFileRepository, string) (interfaces.EmergencyServiceContactRepository, error),
	newService func(interfaces.EmergencyServiceContactRepository) (emergencyServiceContactService, error),
	newHandler func(emergencyServiceContactService) (*handlers.EmergencyServiceContactHandler, error),
) (emergencyServiceContactService, *handlers.EmergencyServiceContactHandler, error) {
	service, err := buildEmergencyServiceContactServiceFromJSONRepository(ctx, jsonRepository, newRepository, newService)
	if err != nil {
		return nil, nil, err
	}
	if err := verifyEmergencyServiceContacts(ctx, service); err != nil {
		return nil, nil, err
	}
	handler, err := newHandler(service)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize emergency service contact handler: %w", err)
	}
	return service, handler, nil
}

// verifyEmergencyServiceContacts uses bounded queries on the same cached service served by HTTP.
func verifyEmergencyServiceContacts(ctx context.Context, service emergencyServiceContactService) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if service == nil {
		return invalidEmergencyServiceContactVerification("service is required")
	}

	page, err := service.ListEmergencyServiceContacts(ctx, interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 1})
	if err != nil {
		return wrapEmergencyServiceContactVerificationError("load verification page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if page.Records == nil || len(page.Records) != 1 || page.Page != 1 || page.PageSize != 1 || page.Total != emergencyServiceContactExpectedCount || page.TotalPages != emergencyServiceContactExpectedCount {
		return invalidEmergencyServiceContactVerification("invalid verification page")
	}
	if err := validateStartupEmergencyServiceContactAnchor(page.Records[0], emergencyServiceContactAnchor{
		id: emergencyServiceContactFirstID, serviceName: emergencyServiceContactFirstService, agencyName: emergencyServiceContactFirstAgency,
		serviceType: emergencyServiceContactFirstType, contactType: emergencyServiceContactFirstContact, contactValue: emergencyServiceContactFirstValue,
	}); err != nil {
		return err
	}

	anchors := []emergencyServiceContactAnchor{
		{id: emergencyServiceContactFirstID, serviceName: emergencyServiceContactFirstService, agencyName: emergencyServiceContactFirstAgency, serviceType: emergencyServiceContactFirstType, contactType: emergencyServiceContactFirstContact, contactValue: emergencyServiceContactFirstValue},
		{id: emergencyServiceContactLastID, serviceName: emergencyServiceContactLastService, agencyName: emergencyServiceContactLastAgency, serviceType: emergencyServiceContactLastType, contactType: emergencyServiceContactLastContact, contactValue: emergencyServiceContactLastValue},
	}
	for _, anchor := range anchors {
		record, err := service.GetEmergencyServiceContact(ctx, anchor.id)
		if err != nil {
			return wrapEmergencyServiceContactVerificationError("retrieve anchor", err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := validateStartupEmergencyServiceContactAnchor(record, anchor); err != nil {
			return err
		}
	}

	repeated, err := service.ListEmergencyServiceContacts(ctx, interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 2, ContactValue: "112"})
	if err != nil {
		return wrapEmergencyServiceContactVerificationError("load repeated contact-value page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if repeated.Records == nil || len(repeated.Records) != 2 || repeated.Total != 2 || repeated.TotalPages != 1 {
		return invalidEmergencyServiceContactVerification("invalid repeated contact-value page")
	}
	if repeated.Records[0].ID == repeated.Records[1].ID || repeated.Records[0].ServiceName == repeated.Records[1].ServiceName || repeated.Records[0].AgencyName == repeated.Records[1].AgencyName {
		return invalidEmergencyServiceContactVerification("repeated contact records collapsed")
	}
	for _, record := range repeated.Records {
		if record.ContactValue != "112" {
			return invalidEmergencyServiceContactVerification("unexpected repeated contact value")
		}
		if err := validateStartupEmergencyServiceContact(record); err != nil {
			return err
		}
	}

	if err := verifyEmergencyServiceContactFilteredTotals(ctx, service, "service_type", map[string]int{
		"general_emergency":   1,
		"road_emergency":      1,
		"disaster_management": 1,
		"fire":                2,
	}); err != nil {
		return err
	}

	national, err := service.ListEmergencyServiceContacts(ctx, interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 1, CoverageType: "national"})
	if err != nil {
		return wrapEmergencyServiceContactVerificationError("load coverage verification page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if national.Records == nil || len(national.Records) != 1 || national.Total != emergencyServiceContactExpectedCount || national.TotalPages != emergencyServiceContactExpectedCount || national.Records[0].CoverageType != "national" {
		return invalidEmergencyServiceContactVerification("invalid coverage verification page")
	}
	if err := validateStartupEmergencyServiceContact(national.Records[0]); err != nil {
		return err
	}

	if err := verifyEmergencyServiceContactFilteredTotals(ctx, service, "contact_type", map[string]int{
		"short_code": 3,
		"telephone":  2,
	}); err != nil {
		return err
	}

	beyond, err := service.ListEmergencyServiceContacts(ctx, interfaces.EmergencyServiceContactQuery{Page: emergencyServiceContactExpectedCount + 1, PageSize: 1})
	if err != nil {
		return wrapEmergencyServiceContactVerificationError("load beyond-final page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if beyond.Records == nil || len(beyond.Records) != 0 || beyond.Page != emergencyServiceContactExpectedCount+1 || beyond.PageSize != 1 || beyond.Total != emergencyServiceContactExpectedCount || beyond.TotalPages != emergencyServiceContactExpectedCount {
		return invalidEmergencyServiceContactVerification("invalid beyond-final page")
	}
	return nil
}

type emergencyServiceContactAnchor struct {
	id           string
	serviceName  string
	agencyName   string
	serviceType  string
	contactType  string
	contactValue string
}

func verifyEmergencyServiceContactFilteredTotals(ctx context.Context, service emergencyServiceContactService, filter string, totals map[string]int) error {
	sum := 0
	for value, total := range totals {
		query := interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 1}
		switch filter {
		case "service_type":
			query.ServiceType = value
		case "contact_type":
			query.ContactType = value
		default:
			return invalidEmergencyServiceContactVerification("unsupported filter verification")
		}
		result, err := service.ListEmergencyServiceContacts(ctx, query)
		if err != nil {
			return wrapEmergencyServiceContactVerificationError("load "+filter+" verification page", err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		sum += result.Total
		if result.Records == nil || len(result.Records) != 1 || result.Page != 1 || result.PageSize != 1 || result.Total != total || result.TotalPages != total {
			return invalidEmergencyServiceContactVerification("invalid " + filter + " verification page")
		}
		record := result.Records[0]
		if filter == "service_type" && record.ServiceType != value {
			return invalidEmergencyServiceContactVerification("invalid service-type verification record")
		}
		if filter == "contact_type" && record.ContactType != value {
			return invalidEmergencyServiceContactVerification("invalid contact-type verification record")
		}
		if err := validateStartupEmergencyServiceContact(record); err != nil {
			return err
		}
	}
	if sum != emergencyServiceContactExpectedCount {
		return invalidEmergencyServiceContactVerification(filter + " totals mismatch")
	}
	return nil
}

func validateStartupEmergencyServiceContactAnchor(record models.EmergencyServiceContact, anchor emergencyServiceContactAnchor) error {
	if record.ID != anchor.id || record.ServiceName != anchor.serviceName || record.AgencyName != anchor.agencyName || record.ServiceType != anchor.serviceType || record.ContactType != anchor.contactType || record.ContactValue != anchor.contactValue {
		return invalidEmergencyServiceContactVerification("anchor mismatch")
	}
	return validateStartupEmergencyServiceContact(record)
}

func validateStartupEmergencyServiceContact(record models.EmergencyServiceContact) error {
	if record.ID == "" || len(record.ID) > models.EmergencyServiceContactIDMaxLength || !startupEmergencyServiceContactIDPattern.MatchString(record.ID) {
		return invalidEmergencyServiceContactVerification("invalid record identity")
	}
	if strings.TrimSpace(record.ServiceName) == "" || strings.TrimSpace(record.AgencyName) == "" || strings.TrimSpace(record.ContactValue) == "" || record.CountryCode != "NG" {
		return invalidEmergencyServiceContactVerification("invalid record contract")
	}
	switch record.ServiceType {
	case "general_emergency", "disaster_management", "road_emergency", "police", "fire", "ambulance", "other":
	default:
		return invalidEmergencyServiceContactVerification("invalid record service type")
	}
	switch record.ContactType {
	case "short_code", "telephone":
	default:
		return invalidEmergencyServiceContactVerification("invalid record contact type")
	}
	switch record.CoverageType {
	case "national", "state":
	default:
		return invalidEmergencyServiceContactVerification("invalid record coverage type")
	}
	if record.CoverageType == "national" && record.StateID != "" {
		return invalidEmergencyServiceContactVerification("national record has state")
	}
	return nil
}

func invalidEmergencyServiceContactVerification(detail string) error {
	return fmt.Errorf("verify emergency service contacts: %w (%s)", interfaces.ErrInvalidDatasetFile, detail)
}

func wrapEmergencyServiceContactVerificationError(operation string, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return invalidEmergencyServiceContactVerification(operation + " failed")
}
