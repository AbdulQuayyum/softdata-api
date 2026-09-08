package file

import (
	"context"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func (r *EducationFileRepository) ListPolytechnics(ctx context.Context) ([]models.Polytechnic, error) {
	return loadCachedEducationDataset(ctx, r, &r.polytechnicsCache, r.polytechnicsPath, validatePolytechnics, cloneSlice[models.Polytechnic])
}

func (r *EducationFileRepository) GetPolytechnic(ctx context.Context, id string) (models.Polytechnic, error) {
	items, err := r.ListPolytechnics(ctx)
	if err != nil {
		return models.Polytechnic{}, err
	}
	return getEducationInstitutionByID(ctx, items, id, func(item models.Polytechnic) string { return item.ID }, interfaces.ErrPolytechnicNotFound, clonePolytechnic)
}

func (r *EducationFileRepository) ListMonotechnics(ctx context.Context) ([]models.Monotechnic, error) {
	return loadCachedEducationDataset(ctx, r, &r.monotechnicsCache, r.monotechnicsPath, validateMonotechnics, cloneSlice[models.Monotechnic])
}

func (r *EducationFileRepository) GetMonotechnic(ctx context.Context, id string) (models.Monotechnic, error) {
	items, err := r.ListMonotechnics(ctx)
	if err != nil {
		return models.Monotechnic{}, err
	}
	return getEducationInstitutionByID(ctx, items, id, func(item models.Monotechnic) string { return item.ID }, interfaces.ErrMonotechnicNotFound, cloneMonotechnic)
}

func (r *EducationFileRepository) ListCollegesOfAgriculture(ctx context.Context) ([]models.CollegeOfAgriculture, error) {
	return loadCachedEducationDataset(ctx, r, &r.collegesOfAgricultureCache, r.collegesOfAgriculturePath, validateCollegesOfAgriculture, cloneSlice[models.CollegeOfAgriculture])
}

func (r *EducationFileRepository) GetCollegeOfAgriculture(ctx context.Context, id string) (models.CollegeOfAgriculture, error) {
	items, err := r.ListCollegesOfAgriculture(ctx)
	if err != nil {
		return models.CollegeOfAgriculture{}, err
	}
	return getEducationInstitutionByID(ctx, items, id, func(item models.CollegeOfAgriculture) string { return item.ID }, interfaces.ErrCollegeOfAgricultureNotFound, cloneCollegeOfAgriculture)
}

func (r *EducationFileRepository) ListCollegesOfHealthSciencesAndTechnology(ctx context.Context) ([]models.CollegeOfHealthSciencesAndTechnology, error) {
	return loadCachedEducationDataset(ctx, r, &r.collegesOfHealthSciencesAndTechnologyCache, r.collegesOfHealthSciencesAndTechnologyPath, validateCollegesOfHealthSciencesAndTechnology, cloneSlice[models.CollegeOfHealthSciencesAndTechnology])
}

func (r *EducationFileRepository) GetCollegeOfHealthSciencesAndTechnology(ctx context.Context, id string) (models.CollegeOfHealthSciencesAndTechnology, error) {
	items, err := r.ListCollegesOfHealthSciencesAndTechnology(ctx)
	if err != nil {
		return models.CollegeOfHealthSciencesAndTechnology{}, err
	}
	return getEducationInstitutionByID(ctx, items, id, func(item models.CollegeOfHealthSciencesAndTechnology) string { return item.ID }, interfaces.ErrCollegeOfHealthSciencesAndTechnologyNotFound, cloneCollegeOfHealthSciencesAndTechnology)
}

func (r *EducationFileRepository) ListVocationalEnterpriseInstitutions(ctx context.Context) ([]models.VocationalEnterpriseInstitution, error) {
	return loadCachedEducationDataset(ctx, r, &r.vocationalEnterpriseInstitutionsCache, r.vocationalEnterpriseInstitutionsPath, validateVocationalEnterpriseInstitutions, cloneSlice[models.VocationalEnterpriseInstitution])
}

func (r *EducationFileRepository) GetVocationalEnterpriseInstitution(ctx context.Context, id string) (models.VocationalEnterpriseInstitution, error) {
	items, err := r.ListVocationalEnterpriseInstitutions(ctx)
	if err != nil {
		return models.VocationalEnterpriseInstitution{}, err
	}
	return getEducationInstitutionByID(ctx, items, id, func(item models.VocationalEnterpriseInstitution) string { return item.ID }, interfaces.ErrVocationalEnterpriseInstitutionNotFound, cloneVocationalEnterpriseInstitution)
}

func clonePolytechnic(item models.Polytechnic) models.Polytechnic {
	return item
}

func cloneMonotechnic(item models.Monotechnic) models.Monotechnic {
	return item
}

func cloneCollegeOfAgriculture(item models.CollegeOfAgriculture) models.CollegeOfAgriculture {
	return item
}

func cloneCollegeOfHealthSciencesAndTechnology(item models.CollegeOfHealthSciencesAndTechnology) models.CollegeOfHealthSciencesAndTechnology {
	return item
}

func cloneVocationalEnterpriseInstitution(item models.VocationalEnterpriseInstitution) models.VocationalEnterpriseInstitution {
	return item
}
