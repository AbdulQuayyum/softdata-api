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

// EducationRepository defines university lookup operations backed by the education dataset.
type EducationRepository interface {
	ListUniversities(ctx context.Context, filter UniversityFilter) ([]models.University, error)
	GetUniversityByID(ctx context.Context, universityID string) (models.University, error)
	ListCollegesOfEducation(ctx context.Context, filter CollegeOfEducationFilter) ([]models.CollegeOfEducation, error)
	GetCollegeOfEducation(ctx context.Context, collegeID string) (models.CollegeOfEducation, error)
}
