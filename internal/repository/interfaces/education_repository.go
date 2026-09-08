package interfaces

import (
	"context"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

// UniversityFilter captures the canonical university filters supported by the repository.
type UniversityFilter struct {
	OwnershipType string
	StateID       string
}

// CollegeOfEducationFilter captures the canonical college filters supported by the repository.
type CollegeOfEducationFilter struct {
	OwnershipType string
	StateID       string
}

// PrimaryAndSecondarySchoolQuery captures the canonical school filters supported by the repository.
type PrimaryAndSecondarySchoolQuery struct {
	Page           int
	PageSize       int
	StateID        string
	LGAID          string
	EducationLevel string
	OwnershipType  string
	Search         string
}

// PrimaryAndSecondarySchoolListResult captures the canonical school pagination result.
type PrimaryAndSecondarySchoolListResult struct {
	Schools    []models.PrimaryAndSecondarySchool
	Page       int
	PageSize   int
	Total      int
	TotalPages int
}

// EducationRepository defines university lookup operations backed by the education dataset.
type EducationRepository interface {
	ListUniversities(ctx context.Context, filter UniversityFilter) ([]models.University, error)
	GetUniversityByID(ctx context.Context, universityID string) (models.University, error)
	ListCollegesOfEducation(ctx context.Context, filter CollegeOfEducationFilter) ([]models.CollegeOfEducation, error)
	GetCollegeOfEducation(ctx context.Context, collegeID string) (models.CollegeOfEducation, error)
	ListPolytechnics(ctx context.Context) ([]models.Polytechnic, error)
	GetPolytechnic(ctx context.Context, id string) (models.Polytechnic, error)
	ListMonotechnics(ctx context.Context) ([]models.Monotechnic, error)
	GetMonotechnic(ctx context.Context, id string) (models.Monotechnic, error)
	ListCollegesOfAgriculture(ctx context.Context) ([]models.CollegeOfAgriculture, error)
	GetCollegeOfAgriculture(ctx context.Context, id string) (models.CollegeOfAgriculture, error)
	ListCollegesOfHealthSciencesAndTechnology(ctx context.Context) ([]models.CollegeOfHealthSciencesAndTechnology, error)
	GetCollegeOfHealthSciencesAndTechnology(ctx context.Context, id string) (models.CollegeOfHealthSciencesAndTechnology, error)
	ListCollegesOfNursingAndMidwifery(ctx context.Context) ([]models.CollegeOfNursingAndMidwifery, error)
	GetCollegeOfNursingAndMidwifery(ctx context.Context, id string) (models.CollegeOfNursingAndMidwifery, error)
	ListTechnicalColleges(ctx context.Context) ([]models.TechnicalCollege, error)
	GetTechnicalCollege(ctx context.Context, id string) (models.TechnicalCollege, error)
	ListVocationalEnterpriseInstitutions(ctx context.Context) ([]models.VocationalEnterpriseInstitution, error)
	GetVocationalEnterpriseInstitution(ctx context.Context, id string) (models.VocationalEnterpriseInstitution, error)
	ListPrimaryAndSecondarySchools(ctx context.Context, query PrimaryAndSecondarySchoolQuery) (PrimaryAndSecondarySchoolListResult, error)
	GetPrimaryAndSecondarySchool(ctx context.Context, id string) (models.PrimaryAndSecondarySchool, error)
}
