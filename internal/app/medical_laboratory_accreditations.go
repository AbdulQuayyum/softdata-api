package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/handlers"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	fileRepo "github.com/AbdulQuayyum/softdata-api/internal/repository/file"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

const (
	medicalLaboratoryAccreditationsPath = "healthcare/medical_laboratory_accreditations.json"
	medicalLaboratoryFirstAnchor        = "university-of-uyo-teaching-hospital-polymerase-chain-reaction-mega-laboratory-akwa-ibom-olusugun-obasanjo-way-ml0018"
	medicalLaboratoryLastAnchor         = "rivers-state-university-teaching-hospital-polymerase-chain-reaction-laboratory-rivers-port-harcourt-ml0016"
)

type medicalLaboratoryAccreditationService interface {
	ListMedicalLaboratoryAccreditations(context.Context, interfaces.MedicalLaboratoryAccreditationQuery) (interfaces.MedicalLaboratoryAccreditationListResult, error)
	GetMedicalLaboratoryAccreditation(context.Context, string) (models.MedicalLaboratoryAccreditation, error)
}

func buildMedicalLaboratoryAccreditationHandler(ctx context.Context, jsonRepository interfaces.JSONFileRepository) (*handlers.MedicalLaboratoryAccreditationHandler, error) {
	repository, err := fileRepo.NewMedicalLaboratoryAccreditationRepository(jsonRepository, medicalLaboratoryAccreditationsPath, geographyStatesRelativePath)
	if err != nil {
		return nil, wrapMedicalLaboratoryVerificationError("initialize repository", err)
	}
	service, err := services.NewMedicalLaboratoryAccreditationService(repository)
	if err != nil {
		return nil, wrapMedicalLaboratoryVerificationError("initialize service", err)
	}
	if err := verifyMedicalLaboratoryAccreditations(ctx, service); err != nil {
		return nil, err
	}
	return handlers.NewMedicalLaboratoryAccreditationHandler(service)
}

// verifyMedicalLaboratoryAccreditations uses bounded queries on the same cached service served by HTTP.
func verifyMedicalLaboratoryAccreditations(ctx context.Context, service medicalLaboratoryAccreditationService) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if service == nil {
		return invalidMedicalLaboratoryVerification("service is required")
	}
	totals := make([]int, 0, 3)
	for _, status := range []string{"", "accredited", "expired"} {
		page, err := service.ListMedicalLaboratoryAccreditations(ctx, interfaces.MedicalLaboratoryAccreditationQuery{Page: 1, PageSize: 1, AccreditationStatus: status})
		if err != nil {
			return wrapMedicalLaboratoryVerificationError("load verification page", err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if page.Records == nil || len(page.Records) != 1 || page.Page != 1 || page.PageSize != 1 || page.TotalPages != page.Total || page.Total < 1 {
			return invalidMedicalLaboratoryVerification("invalid verification page")
		}
		if err := validateStartupMedicalLaboratoryAccreditation(page.Records[0]); err != nil {
			return err
		}
		if status != "" && page.Records[0].AccreditationStatus != status {
			return invalidMedicalLaboratoryVerification("status filter mismatch")
		}
		totals = append(totals, page.Total)
	}
	if totals[1]+totals[2] != totals[0] {
		return invalidMedicalLaboratoryVerification("invalid status arithmetic")
	}
	if totals[0] != 30 || totals[1] != 26 || totals[2] != 4 {
		return invalidMedicalLaboratoryVerification("unexpected snapshot counts")
	}
	for _, id := range []string{medicalLaboratoryFirstAnchor, medicalLaboratoryLastAnchor} {
		record, err := service.GetMedicalLaboratoryAccreditation(ctx, id)
		if err != nil {
			return wrapMedicalLaboratoryVerificationError("retrieve anchor", err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if record.ID != id {
			return invalidMedicalLaboratoryVerification("anchor mismatch")
		}
		if err := validateStartupMedicalLaboratoryAccreditation(record); err != nil {
			return err
		}
	}
	beyond, err := service.ListMedicalLaboratoryAccreditations(ctx, interfaces.MedicalLaboratoryAccreditationQuery{Page: 31, PageSize: 1})
	if err != nil {
		return wrapMedicalLaboratoryVerificationError("load beyond-final page", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if beyond.Records == nil || len(beyond.Records) != 0 || beyond.Page != 31 || beyond.PageSize != 1 || beyond.Total != 30 || beyond.TotalPages != 30 {
		return invalidMedicalLaboratoryVerification("invalid beyond-final page")
	}
	return nil
}

func validateStartupMedicalLaboratoryAccreditation(record models.MedicalLaboratoryAccreditation) error {
	if len(record.ID) > models.MedicalLaboratoryAccreditationIDMaxLength || !startupHealthFacilityIDPattern.MatchString(record.ID) || strings.TrimSpace(record.Name) == "" || record.CountryCode != "NG" {
		return invalidMedicalLaboratoryVerification("invalid record identity")
	}
	if _, ok := approvedUniversityStateIDs[record.StateID]; !ok {
		return invalidMedicalLaboratoryVerification("invalid record state")
	}
	if record.AccreditationStatus != "accredited" && record.AccreditationStatus != "expired" {
		return invalidMedicalLaboratoryVerification("invalid record status")
	}
	for _, value := range []string{record.AccreditationNumber, record.Address} {
		if value != "" && strings.TrimSpace(value) == "" {
			return invalidMedicalLaboratoryVerification("blank optional field")
		}
	}
	for _, value := range []string{record.ApprovalDate, record.ExpiryDate} {
		if value != "" {
			if _, err := time.Parse(time.DateOnly, value); err != nil {
				return invalidMedicalLaboratoryVerification("invalid optional date")
			}
		}
	}
	return nil
}
func invalidMedicalLaboratoryVerification(detail string) error {
	return fmt.Errorf("verify medical laboratory accreditations: %w (%s)", interfaces.ErrInvalidDatasetFile, detail)
}
func wrapMedicalLaboratoryVerificationError(operation string, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return invalidMedicalLaboratoryVerification(operation + " failed")
}
