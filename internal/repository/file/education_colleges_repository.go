package file

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var expectedCollegeOfEducationOwnershipCounts = map[string]int{
	"federal": 28,
	"state":   48,
	"private": 168,
}

var expectedCollegeOfEducationStateCounts = map[string]int{
	"abia": 5, "adamawa": 2, "akwa-ibom": 2, "anambra": 6, "bauchi": 16,
	"bayelsa": 1, "benue": 15, "borno": 4, "cross-river": 3, "delta": 5,
	"ebonyi": 3, "edo": 3, "ekiti": 3, "enugu": 8, "fct": 5, "gombe": 8,
	"imo": 4, "jigawa": 3, "kaduna": 6, "kano": 16, "katsina": 4,
	"kebbi": 4, "kogi": 9, "kwara": 21, "lagos": 12, "nasarawa": 6,
	"niger": 2, "ogun": 10, "ondo": 9, "osun": 13, "oyo": 12, "plateau": 9,
	"rivers": 2, "sokoto": 3, "taraba": 4, "yobe": 4, "zamfara": 2,
}

var expectedCollegeOfEducationExcludedNames = map[string]struct{}{
	"Abubakar Tatari Polytechnic":                                  {},
	"Aminu Kano College of Islamic and Legal Studies":              {},
	"Bauchi Institute for Arabic and Islamic Studies":              {},
	"Hassan Usman Katsina Polytechnic":                             {},
	"Institute of Ecumenical Education (Thinkers Corner)":          {},
	"Jigawa State Polytechnic":                                     {},
	"Kaduna Polytechnics":                                          {},
	"Kano State Polytechnic":                                       {},
	"Kebbi State Polytechnic":                                      {},
	"Muhammad Goni College of Legal and Islamic Studies (MOGOLIS)": {},
	"National Institute for Nigerian Languages":                    {},
	"National Teachers Institute (NTI)":                            {},
	"Nigerian Army College of Education (NACOE), Ilorin":           {},
	"Nuhu Bamalli Polytechnic":                                     {},
	"Plateau State Polytechnic":                                    {},
	"Ramat Polytechnic":                                            {},
	"The Polytechnic Iree, Osun State":                             {},
	"Waziri Umaru Federal Polytechnic":                             {},
	"Zaria Institute of Information Technology":                    {},
	"Cross River State Coll. of Education, Akampa":                 {},
}

var _ interfaces.EducationRepository = (*EducationFileRepository)(nil)

// ListCollegesOfEducation returns the ordered list of colleges matching the supplied filter.
func (r *EducationFileRepository) ListCollegesOfEducation(ctx context.Context, filter interfaces.CollegeOfEducationFilter) ([]models.CollegeOfEducation, error) {
	colleges, err := r.loadCollegesOfEducation(ctx)
	if err != nil {
		return nil, err
	}

	filtered := make([]models.CollegeOfEducation, 0, len(colleges))
	for _, college := range colleges {
		if filter.OwnershipType != "" && college.OwnershipType != filter.OwnershipType {
			continue
		}
		if filter.StateID != "" && college.StateID != filter.StateID {
			continue
		}
		filtered = append(filtered, college)
	}
	return cloneCollegeOfEducationList(filtered), nil
}

// GetCollegeOfEducation returns a single college using its public slug identifier.
func (r *EducationFileRepository) GetCollegeOfEducation(ctx context.Context, collegeID string) (models.CollegeOfEducation, error) {
	colleges, err := r.loadCollegesOfEducation(ctx)
	if err != nil {
		return models.CollegeOfEducation{}, err
	}

	for _, college := range colleges {
		if college.ID == collegeID {
			return cloneCollegeOfEducation(college), nil
		}
	}

	return models.CollegeOfEducation{}, fmt.Errorf("%w", interfaces.ErrCollegeOfEducationNotFound)
}

func (r *EducationFileRepository) loadCollegesOfEducation(ctx context.Context) ([]models.CollegeOfEducation, error) {
	if r == nil || r.jsonRepository == nil {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var colleges []models.CollegeOfEducation
	if err := r.jsonRepository.Decode(ctx, r.collegesOfEducationPath, &colleges); err != nil {
		return nil, translateEducationLoadError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if colleges == nil || len(colleges) == 0 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if err := validateCollegesOfEducation(colleges); err != nil {
		return nil, err
	}
	return colleges, nil
}

func validateCollegesOfEducation(colleges []models.CollegeOfEducation) error {
	if len(colleges) != 244 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	seenIDs := make(map[string]struct{}, len(colleges))
	seenPairs := make(map[string]struct{}, len(colleges))
	ownershipCounts := make(map[string]int, len(expectedCollegeOfEducationOwnershipCounts))
	stateCounts := make(map[string]int, len(expectedCollegeOfEducationStateCounts))
	for ownershipType := range expectedCollegeOfEducationOwnershipCounts {
		ownershipCounts[ownershipType] = 0
	}
	for stateID := range expectedCollegeOfEducationStateCounts {
		stateCounts[stateID] = 0
	}

	lastName := ""
	lastID := ""
	for _, college := range colleges {
		if college.ID == "" || college.Name == "" || college.OwnershipType == "" || college.StateID == "" || college.CountryCode == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if college.CountryCode != "NG" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := expectedCollegeOfEducationOwnershipCounts[college.OwnershipType]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := expectedCollegeOfEducationStateCounts[college.StateID]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !universityIDPattern.MatchString(college.ID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if wantID := slugifyUniversityName(college.Name); college.ID != wantID {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenIDs[college.ID]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		pairKey := college.StateID + "\x00" + college.Name
		if _, ok := seenPairs[pairKey]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := expectedCollegeOfEducationExcludedNames[college.Name]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if lastName != "" {
			cmp := strings.Compare(strings.ToLower(lastName), strings.ToLower(college.Name))
			if cmp > 0 || (cmp == 0 && strings.Compare(lastID, college.ID) > 0) {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}

		seenIDs[college.ID] = struct{}{}
		seenPairs[pairKey] = struct{}{}
		ownershipCounts[college.OwnershipType]++
		stateCounts[college.StateID]++
		lastName = college.Name
		lastID = college.ID
	}

	if !reflect.DeepEqual(ownershipCounts, expectedCollegeOfEducationOwnershipCounts) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if !reflect.DeepEqual(stateCounts, expectedCollegeOfEducationStateCounts) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	return nil
}

func cloneCollegeOfEducationList(colleges []models.CollegeOfEducation) []models.CollegeOfEducation {
	if len(colleges) == 0 {
		return make([]models.CollegeOfEducation, 0)
	}
	cloned := make([]models.CollegeOfEducation, len(colleges))
	copy(cloned, colleges)
	return cloned
}

func cloneCollegeOfEducation(college models.CollegeOfEducation) models.CollegeOfEducation {
	return college
}
