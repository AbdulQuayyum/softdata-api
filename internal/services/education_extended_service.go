package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const maxEducationInstitutionIDLength = 128
const maxSchoolSearchLength = 256

var educationInstitutionIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)

var schoolServiceEducationLevels = map[string]struct{}{
	"pre-primary":      {},
	"primary":          {},
	"junior-secondary": {},
	"senior-secondary": {},
}

var schoolServiceOwnershipTypes = map[string]struct{}{
	"public":  {},
	"private": {},
}

func (s *EducationService) ListPolytechnics(ctx context.Context) ([]models.Polytechnic, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := s.repository.ListPolytechnics(ctx)
	if err != nil {
		return nil, translateEducationInstitutionListError("list polytechnics", err)
	}
	return clonePolytechnicList(rows), nil
}

func (s *EducationService) GetPolytechnic(ctx context.Context, id string) (models.Polytechnic, error) {
	if err := ctx.Err(); err != nil {
		return models.Polytechnic{}, err
	}
	id, err := normalizeEducationInstitutionID(id)
	if err != nil {
		return models.Polytechnic{}, err
	}
	rows, err := s.repository.GetPolytechnic(ctx, id)
	if err != nil {
		return models.Polytechnic{}, translateEducationInstitutionLookupError("get polytechnic", err, ErrPolytechnicNotFound)
	}
	return rows, nil
}

func (s *EducationService) ListMonotechnics(ctx context.Context) ([]models.Monotechnic, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := s.repository.ListMonotechnics(ctx)
	if err != nil {
		return nil, translateEducationInstitutionListError("list monotechnics", err)
	}
	return cloneMonotechnicList(rows), nil
}

func (s *EducationService) GetMonotechnic(ctx context.Context, id string) (models.Monotechnic, error) {
	if err := ctx.Err(); err != nil {
		return models.Monotechnic{}, err
	}
	id, err := normalizeEducationInstitutionID(id)
	if err != nil {
		return models.Monotechnic{}, err
	}
	rows, err := s.repository.GetMonotechnic(ctx, id)
	if err != nil {
		return models.Monotechnic{}, translateEducationInstitutionLookupError("get monotechnic", err, ErrMonotechnicNotFound)
	}
	return rows, nil
}

func (s *EducationService) ListCollegesOfAgriculture(ctx context.Context) ([]models.CollegeOfAgriculture, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := s.repository.ListCollegesOfAgriculture(ctx)
	if err != nil {
		return nil, translateEducationInstitutionListError("list colleges of agriculture", err)
	}
	return cloneCollegeOfAgricultureList(rows), nil
}

func (s *EducationService) GetCollegeOfAgriculture(ctx context.Context, id string) (models.CollegeOfAgriculture, error) {
	if err := ctx.Err(); err != nil {
		return models.CollegeOfAgriculture{}, err
	}
	id, err := normalizeEducationInstitutionID(id)
	if err != nil {
		return models.CollegeOfAgriculture{}, err
	}
	rows, err := s.repository.GetCollegeOfAgriculture(ctx, id)
	if err != nil {
		return models.CollegeOfAgriculture{}, translateEducationInstitutionLookupError("get college of agriculture", err, ErrCollegeOfAgricultureNotFound)
	}
	return rows, nil
}

func (s *EducationService) ListCollegesOfHealthSciencesAndTechnology(ctx context.Context) ([]models.CollegeOfHealthSciencesAndTechnology, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := s.repository.ListCollegesOfHealthSciencesAndTechnology(ctx)
	if err != nil {
		return nil, translateEducationInstitutionListError("list colleges of health sciences and technology", err)
	}
	return cloneCollegeOfHealthSciencesAndTechnologyList(rows), nil
}

func (s *EducationService) GetCollegeOfHealthSciencesAndTechnology(ctx context.Context, id string) (models.CollegeOfHealthSciencesAndTechnology, error) {
	if err := ctx.Err(); err != nil {
		return models.CollegeOfHealthSciencesAndTechnology{}, err
	}
	id, err := normalizeEducationInstitutionID(id)
	if err != nil {
		return models.CollegeOfHealthSciencesAndTechnology{}, err
	}
	rows, err := s.repository.GetCollegeOfHealthSciencesAndTechnology(ctx, id)
	if err != nil {
		return models.CollegeOfHealthSciencesAndTechnology{}, translateEducationInstitutionLookupError("get college of health sciences and technology", err, ErrCollegeOfHealthSciencesAndTechnologyNotFound)
	}
	return rows, nil
}

func (s *EducationService) ListCollegesOfNursingAndMidwifery(ctx context.Context) ([]models.CollegeOfNursingAndMidwifery, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := s.repository.ListCollegesOfNursingAndMidwifery(ctx)
	if err != nil {
		return nil, translateEducationInstitutionListError("list colleges of nursing and midwifery", err)
	}
	return cloneCollegeOfNursingAndMidwiferyList(rows), nil
}

func (s *EducationService) GetCollegeOfNursingAndMidwifery(ctx context.Context, id string) (models.CollegeOfNursingAndMidwifery, error) {
	if err := ctx.Err(); err != nil {
		return models.CollegeOfNursingAndMidwifery{}, err
	}
	id, err := normalizeEducationInstitutionID(id)
	if err != nil {
		return models.CollegeOfNursingAndMidwifery{}, err
	}
	rows, err := s.repository.GetCollegeOfNursingAndMidwifery(ctx, id)
	if err != nil {
		return models.CollegeOfNursingAndMidwifery{}, translateEducationInstitutionLookupError("get college of nursing and midwifery", err, ErrCollegeOfNursingAndMidwiferyNotFound)
	}
	return rows, nil
}

func (s *EducationService) ListTechnicalColleges(ctx context.Context) ([]models.TechnicalCollege, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := s.repository.ListTechnicalColleges(ctx)
	if err != nil {
		return nil, translateEducationInstitutionListError("list technical colleges", err)
	}
	return cloneTechnicalCollegeList(rows), nil
}

func (s *EducationService) GetTechnicalCollege(ctx context.Context, id string) (models.TechnicalCollege, error) {
	if err := ctx.Err(); err != nil {
		return models.TechnicalCollege{}, err
	}
	id, err := normalizeEducationInstitutionID(id)
	if err != nil {
		return models.TechnicalCollege{}, err
	}
	rows, err := s.repository.GetTechnicalCollege(ctx, id)
	if err != nil {
		return models.TechnicalCollege{}, translateEducationInstitutionLookupError("get technical college", err, ErrTechnicalCollegeNotFound)
	}
	return rows, nil
}

func (s *EducationService) ListVocationalEnterpriseInstitutions(ctx context.Context) ([]models.VocationalEnterpriseInstitution, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := s.repository.ListVocationalEnterpriseInstitutions(ctx)
	if err != nil {
		return nil, translateEducationInstitutionListError("list vocational enterprise institutions", err)
	}
	return cloneVocationalEnterpriseInstitutionList(rows), nil
}

func (s *EducationService) GetVocationalEnterpriseInstitution(ctx context.Context, id string) (models.VocationalEnterpriseInstitution, error) {
	if err := ctx.Err(); err != nil {
		return models.VocationalEnterpriseInstitution{}, err
	}
	id, err := normalizeEducationInstitutionID(id)
	if err != nil {
		return models.VocationalEnterpriseInstitution{}, err
	}
	rows, err := s.repository.GetVocationalEnterpriseInstitution(ctx, id)
	if err != nil {
		return models.VocationalEnterpriseInstitution{}, translateEducationInstitutionLookupError("get vocational enterprise institution", err, ErrVocationalEnterpriseInstitutionNotFound)
	}
	return rows, nil
}

func (s *EducationService) ListPrimaryAndSecondarySchools(ctx context.Context, query interfaces.PrimaryAndSecondarySchoolQuery) (interfaces.PrimaryAndSecondarySchoolListResult, error) {
	if err := ctx.Err(); err != nil {
		return interfaces.PrimaryAndSecondarySchoolListResult{}, err
	}

	normalized, err := normalizePrimaryAndSecondarySchoolQuery(query)
	if err != nil {
		return interfaces.PrimaryAndSecondarySchoolListResult{}, err
	}

	result, err := s.repository.ListPrimaryAndSecondarySchools(ctx, normalized)
	if err != nil {
		return interfaces.PrimaryAndSecondarySchoolListResult{}, translatePrimaryAndSecondarySchoolListError(err, normalized)
	}
	result.Schools = clonePrimaryAndSecondarySchoolList(result.Schools)
	return result, nil
}

func (s *EducationService) GetPrimaryAndSecondarySchool(ctx context.Context, id string) (models.PrimaryAndSecondarySchool, error) {
	if err := ctx.Err(); err != nil {
		return models.PrimaryAndSecondarySchool{}, err
	}
	id, err := normalizeEducationInstitutionID(id)
	if err != nil {
		return models.PrimaryAndSecondarySchool{}, err
	}

	school, err := s.repository.GetPrimaryAndSecondarySchool(ctx, id)
	if err != nil {
		return models.PrimaryAndSecondarySchool{}, translatePrimaryAndSecondarySchoolLookupError(err)
	}
	return clonePrimaryAndSecondarySchool(school), nil
}

func normalizeEducationInstitutionID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxEducationInstitutionIDLength || !educationInstitutionIDPattern.MatchString(value) {
		return "", ErrInvalidEducationInstitutionID
	}
	return value, nil
}

func normalizePrimaryAndSecondarySchoolQuery(query interfaces.PrimaryAndSecondarySchoolQuery) (interfaces.PrimaryAndSecondarySchoolQuery, error) {
	normalized := interfaces.PrimaryAndSecondarySchoolQuery{
		Page:           query.Page,
		PageSize:       query.PageSize,
		StateID:        strings.TrimSpace(query.StateID),
		LGAID:          strings.TrimSpace(query.LGAID),
		EducationLevel: strings.TrimSpace(query.EducationLevel),
		OwnershipType:  strings.TrimSpace(query.OwnershipType),
		Search:         strings.TrimSpace(query.Search),
	}

	if normalized.Page == 0 {
		normalized.Page = 1
	}
	if normalized.PageSize == 0 {
		normalized.PageSize = 50
	}
	if normalized.Page < 1 || normalized.PageSize < 1 || normalized.PageSize > 100 {
		return interfaces.PrimaryAndSecondarySchoolQuery{}, ErrInvalidSchoolPagination
	}
	if normalized.Search != "" && len(normalized.Search) > maxSchoolSearchLength {
		return interfaces.PrimaryAndSecondarySchoolQuery{}, ErrInvalidSchoolSearch
	}
	if normalized.StateID != "" {
		if _, ok := allowedUniversityStateIDs[normalized.StateID]; !ok {
			return interfaces.PrimaryAndSecondarySchoolQuery{}, ErrInvalidSchoolStateFilter
		}
	}
	if normalized.LGAID != "" && (len(normalized.LGAID) > maxEducationInstitutionIDLength || !educationInstitutionIDPattern.MatchString(normalized.LGAID)) {
		return interfaces.PrimaryAndSecondarySchoolQuery{}, ErrInvalidSchoolLGAFilter
	}
	if normalized.EducationLevel != "" {
		if _, ok := schoolServiceEducationLevels[normalized.EducationLevel]; !ok {
			return interfaces.PrimaryAndSecondarySchoolQuery{}, ErrInvalidSchoolEducationLevel
		}
	}
	if normalized.OwnershipType != "" {
		if _, ok := schoolServiceOwnershipTypes[normalized.OwnershipType]; !ok {
			return interfaces.PrimaryAndSecondarySchoolQuery{}, ErrInvalidSchoolOwnershipType
		}
	}
	return normalized, nil
}

func translateEducationInstitutionListError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable), errors.Is(err, interfaces.ErrInvalidDatasetFile):
		return fmt.Errorf("%s: repository unavailable", op)
	default:
		return fmt.Errorf("%s: repository unavailable", op)
	}
}

func translateEducationInstitutionLookupError(op string, err error, notFound error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrPolytechnicNotFound), errors.Is(err, interfaces.ErrMonotechnicNotFound), errors.Is(err, interfaces.ErrCollegeOfAgricultureNotFound), errors.Is(err, interfaces.ErrCollegeOfHealthSciencesAndTechnologyNotFound), errors.Is(err, interfaces.ErrCollegeOfNursingAndMidwiferyNotFound), errors.Is(err, interfaces.ErrTechnicalCollegeNotFound), errors.Is(err, interfaces.ErrVocationalEnterpriseInstitutionNotFound):
		switch {
		case errors.Is(err, interfaces.ErrPolytechnicNotFound):
			return ErrPolytechnicNotFound
		case errors.Is(err, interfaces.ErrMonotechnicNotFound):
			return ErrMonotechnicNotFound
		case errors.Is(err, interfaces.ErrCollegeOfAgricultureNotFound):
			return ErrCollegeOfAgricultureNotFound
		case errors.Is(err, interfaces.ErrCollegeOfHealthSciencesAndTechnologyNotFound):
			return ErrCollegeOfHealthSciencesAndTechnologyNotFound
		case errors.Is(err, interfaces.ErrCollegeOfNursingAndMidwiferyNotFound):
			return ErrCollegeOfNursingAndMidwiferyNotFound
		case errors.Is(err, interfaces.ErrTechnicalCollegeNotFound):
			return ErrTechnicalCollegeNotFound
		case errors.Is(err, interfaces.ErrVocationalEnterpriseInstitutionNotFound):
			return ErrVocationalEnterpriseInstitutionNotFound
		}
		return fmt.Errorf("%s: repository unavailable", op)
	case errors.Is(err, notFound):
		return notFound
	case errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable), errors.Is(err, interfaces.ErrInvalidDatasetFile):
		return fmt.Errorf("%s: repository unavailable", op)
	default:
		return fmt.Errorf("%s: repository unavailable", op)
	}
}

func translatePrimaryAndSecondarySchoolListError(err error, query interfaces.PrimaryAndSecondarySchoolQuery) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrInvalidPrimaryAndSecondarySchoolQuery):
		switch {
		case query.LGAID != "":
			return ErrInvalidSchoolLGAFilter
		case query.StateID != "":
			return ErrInvalidSchoolStateFilter
		case query.EducationLevel != "":
			return ErrInvalidSchoolEducationLevel
		case query.OwnershipType != "":
			return ErrInvalidSchoolOwnershipType
		case query.Search != "":
			return ErrInvalidSchoolSearch
		default:
			return ErrInvalidSchoolPagination
		}
	case errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable), errors.Is(err, interfaces.ErrInvalidDatasetFile):
		return fmt.Errorf("list primary and secondary schools: repository unavailable")
	default:
		return fmt.Errorf("list primary and secondary schools: repository unavailable")
	}
}

func translatePrimaryAndSecondarySchoolLookupError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrPrimaryAndSecondarySchoolNotFound):
		return ErrPrimaryAndSecondarySchoolNotFound
	case errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable), errors.Is(err, interfaces.ErrInvalidDatasetFile):
		return fmt.Errorf("get primary and secondary school: repository unavailable")
	default:
		return fmt.Errorf("get primary and secondary school: repository unavailable")
	}
}

func clonePolytechnicList(rows []models.Polytechnic) []models.Polytechnic {
	return cloneEducationSlice(rows)
}

func cloneMonotechnicList(rows []models.Monotechnic) []models.Monotechnic {
	return cloneEducationSlice(rows)
}

func cloneCollegeOfAgricultureList(rows []models.CollegeOfAgriculture) []models.CollegeOfAgriculture {
	return cloneEducationSlice(rows)
}

func cloneCollegeOfHealthSciencesAndTechnologyList(rows []models.CollegeOfHealthSciencesAndTechnology) []models.CollegeOfHealthSciencesAndTechnology {
	return cloneEducationSlice(rows)
}

func cloneCollegeOfNursingAndMidwiferyList(rows []models.CollegeOfNursingAndMidwifery) []models.CollegeOfNursingAndMidwifery {
	return cloneEducationSlice(rows)
}

func cloneTechnicalCollegeList(rows []models.TechnicalCollege) []models.TechnicalCollege {
	return cloneEducationSlice(rows)
}

func cloneVocationalEnterpriseInstitutionList(rows []models.VocationalEnterpriseInstitution) []models.VocationalEnterpriseInstitution {
	return cloneEducationSlice(rows)
}

func cloneEducationSlice[T any](rows []T) []T {
	if len(rows) == 0 {
		return make([]T, 0)
	}
	cloned := make([]T, len(rows))
	copy(cloned, rows)
	return cloned
}

func clonePrimaryAndSecondarySchoolList(rows []models.PrimaryAndSecondarySchool) []models.PrimaryAndSecondarySchool {
	if len(rows) == 0 {
		return make([]models.PrimaryAndSecondarySchool, 0)
	}
	cloned := make([]models.PrimaryAndSecondarySchool, len(rows))
	for i, row := range rows {
		cloned[i] = clonePrimaryAndSecondarySchool(row)
	}
	return cloned
}

func clonePrimaryAndSecondarySchool(row models.PrimaryAndSecondarySchool) models.PrimaryAndSecondarySchool {
	if len(row.EducationLevels) > 0 {
		row.EducationLevels = append([]string(nil), row.EducationLevels...)
	}
	return row
}
