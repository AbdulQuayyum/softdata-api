package file

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var languageIDPattern = regexp.MustCompile(`^[a-z]{2,3}$`)
var languageCountryAreaIDPattern = regexp.MustCompile(`^[a-z]{2}$`)

var expectedLanguageCount = 633
var expectedCountryLanguageCount = 1289
var expectedCountryLanguageStatusCounts = map[string]int{
	"used":              833,
	"official":          319,
	"official_regional": 117,
	"de_facto_official": 20,
}

var allowedCountryLanguageStatuses = map[string]struct{}{
	"official":          {},
	"de_facto_official": {},
	"official_regional": {},
	"used":              {},
}

var countryLanguageSpecialRemaps = map[string]string{
	"fat": "ak",
	"sh":  "sr-Latn",
	"tl":  "fil",
	"tw":  "ak",
}

type countryLanguageSourceRow struct {
	CountryAreaID string
	LanguageID    string
	Status        string
	BaseLanguage  bool
}

// normalizeCountryLanguageRows collapses source rows using the exact base-row-wins rule.
func normalizeCountryLanguageRows(rows []countryLanguageSourceRow) []models.CountryLanguage {
	type selectedRow struct {
		relation models.CountryLanguage
		base     bool
	}
	selected := make(map[string]selectedRow, len(rows))
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		languageID := normalizeCountryLanguageID(row.LanguageID)
		if languageID == "" || row.CountryAreaID == "" {
			continue
		}
		if row.Status == "" {
			row.Status = "used"
		}
		key := row.CountryAreaID + "\x00" + languageID
		candidate := selectedRow{relation: models.CountryLanguage{CountryAreaID: row.CountryAreaID, LanguageID: languageID, Status: row.Status}, base: row.BaseLanguage}
		current, exists := selected[key]
		if !exists {
			order = append(order, key)
			selected[key] = candidate
			continue
		}
		if candidate.base || !current.base {
			selected[key] = candidate
		}
	}

	result := make([]models.CountryLanguage, 0, len(selected))
	for _, key := range order {
		result = append(result, selected[key].relation)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].CountryAreaID == result[j].CountryAreaID {
			return result[i].LanguageID < result[j].LanguageID
		}
		return result[i].CountryAreaID < result[j].CountryAreaID
	})
	return result
}

func normalizeCountryLanguageID(languageID string) string {
	if remapped, ok := countryLanguageSpecialRemaps[languageID]; ok {
		languageID = remapped
	}
	if separator := strings.IndexByte(languageID, '-'); separator >= 0 {
		languageID = languageID[:separator]
	}
	return languageID
}

// ListLanguages returns the ordered list of current CLDR base language identifiers.
func (r *GeographyFileRepository) ListLanguages(ctx context.Context, filter interfaces.LanguageFilter) ([]models.Language, error) {
	languages, err := r.loadLanguagesOnly(ctx)
	if err != nil {
		return nil, err
	}
	if len(languages) == 0 {
		return make([]models.Language, 0), nil
	}
	return cloneLanguageList(languages), nil
}

// GetLanguage returns a single CLDR base language using its public identifier.
func (r *GeographyFileRepository) GetLanguage(ctx context.Context, languageID string) (models.Language, error) {
	languageID = strings.TrimSpace(languageID)
	if languageID == "" || !languageIDPattern.MatchString(languageID) {
		return models.Language{}, fmt.Errorf("%w", interfaces.ErrLanguageNotFound)
	}

	languages, err := r.loadLanguagesOnly(ctx)
	if err != nil {
		return models.Language{}, err
	}
	for _, language := range languages {
		if language.ID == languageID {
			return cloneLanguage(language), nil
		}
	}

	return models.Language{}, fmt.Errorf("%w", interfaces.ErrLanguageNotFound)
}

// ListCountryLanguages returns the ordered list of country/area language relationships.
func (r *GeographyFileRepository) ListCountryLanguages(ctx context.Context, filter interfaces.CountryLanguageFilter) ([]models.CountryLanguage, error) {
	countries, languages, relations, err := r.loadCountryLanguageData(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateCountryLanguages(relations, countries, languages); err != nil {
		return nil, err
	}

	filter.CountryAreaID = strings.TrimSpace(filter.CountryAreaID)
	filter.LanguageID = strings.TrimSpace(filter.LanguageID)
	filter.Status = strings.TrimSpace(filter.Status)
	if filter.CountryAreaID != "" && !languageCountryAreaIDPattern.MatchString(filter.CountryAreaID) {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidCountryLanguageCountryAreaID)
	}
	if filter.LanguageID != "" && !languageIDPattern.MatchString(filter.LanguageID) {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidCountryLanguageLanguageID)
	}
	if filter.Status != "" {
		if _, ok := allowedCountryLanguageStatuses[filter.Status]; !ok {
			return nil, fmt.Errorf("%w", interfaces.ErrInvalidCountryLanguageStatus)
		}
	}

	filtered := make([]models.CountryLanguage, 0, len(relations))
	for _, relation := range relations {
		if filter.CountryAreaID != "" && relation.CountryAreaID != filter.CountryAreaID {
			continue
		}
		if filter.LanguageID != "" && relation.LanguageID != filter.LanguageID {
			continue
		}
		if filter.Status != "" && relation.Status != filter.Status {
			continue
		}
		filtered = append(filtered, relation)
	}
	if len(filtered) == 0 {
		return make([]models.CountryLanguage, 0), nil
	}
	return cloneCountryLanguageList(filtered), nil
}

func (r *GeographyFileRepository) loadLanguagesOnly(ctx context.Context) ([]models.Language, error) {
	if r == nil || r.jsonRepository == nil {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var languages []models.Language
	if err := r.jsonRepository.Decode(ctx, r.languagesPath, &languages); err != nil {
		return nil, translateGeographyLoadError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if languages == nil || len(languages) == 0 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if err := validateLanguages(languages); err != nil {
		return nil, err
	}
	return languages, nil
}

func (r *GeographyFileRepository) loadCountryLanguageData(ctx context.Context) ([]models.CountryOrArea, []models.Language, []models.CountryLanguage, error) {
	countries, err := r.loadCountriesAndAreasOnly(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := validateCountryOrAreas(countries); err != nil {
		return nil, nil, nil, err
	}

	languages, err := r.loadLanguagesOnly(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	var relations []models.CountryLanguage
	if r == nil || r.jsonRepository == nil {
		return nil, nil, nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, nil, err
	}
	if err := r.jsonRepository.Decode(ctx, r.countryLanguagesPath, &relations); err != nil {
		return nil, nil, nil, translateGeographyLoadError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, nil, err
	}
	if relations == nil || len(relations) == 0 {
		return nil, nil, nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	// The checked-in dataset is already normalized; applying the same collapse keeps
	// repository behavior correct if source-shaped rows are supplied later.
	sourceRows := make([]countryLanguageSourceRow, len(relations))
	for i, relation := range relations {
		sourceRows[i] = countryLanguageSourceRow{CountryAreaID: relation.CountryAreaID, LanguageID: relation.LanguageID, Status: relation.Status, BaseLanguage: true}
	}
	relations = normalizeCountryLanguageRows(sourceRows)

	return countries, languages, relations, nil
}

func validateLanguages(languages []models.Language) error {
	if len(languages) != expectedLanguageCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	seenIDs := make(map[string]struct{}, len(languages))
	seenNames := make(map[string]struct{}, len(languages))
	var prevID string

	for i, language := range languages {
		if language.ID == "" || language.Name == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !languageIDPattern.MatchString(language.ID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenIDs[language.ID]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenNames[language.Name]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if i > 0 && strings.Compare(prevID, language.ID) > 0 {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		seenIDs[language.ID] = struct{}{}
		seenNames[language.Name] = struct{}{}
		prevID = language.ID
	}
	return nil
}

func validateCountryLanguages(relations []models.CountryLanguage, countries []models.CountryOrArea, languages []models.Language) error {
	if len(relations) != expectedCountryLanguageCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	countryIDs := make(map[string]struct{}, len(countries))
	for _, country := range countries {
		countryIDs[country.ID] = struct{}{}
	}

	languageIDs := make(map[string]struct{}, len(languages))
	for _, language := range languages {
		languageIDs[language.ID] = struct{}{}
	}

	seenPairs := make(map[string]struct{}, len(relations))
	countryCounts := make(map[string]int, len(countries))
	statusCounts := map[string]int{
		"used":              0,
		"official":          0,
		"official_regional": 0,
		"de_facto_official": 0,
	}

	var prevCountry string
	var prevLanguage string
	for i, relation := range relations {
		if relation.CountryAreaID == "" || relation.LanguageID == "" || relation.Status == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := countryIDs[relation.CountryAreaID]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := languageIDs[relation.LanguageID]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := allowedCountryLanguageStatuses[relation.Status]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		pairKey := relation.CountryAreaID + "\x00" + relation.LanguageID
		if _, ok := seenPairs[pairKey]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if i > 0 {
			if strings.Compare(prevCountry, relation.CountryAreaID) > 0 || (prevCountry == relation.CountryAreaID && strings.Compare(prevLanguage, relation.LanguageID) > 0) {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
		}
		countryCounts[relation.CountryAreaID]++
		statusCounts[relation.Status]++
		seenPairs[pairKey] = struct{}{}
		prevCountry = relation.CountryAreaID
		prevLanguage = relation.LanguageID
	}

	if len(countryCounts) != len(countryIDs) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	for countryID := range countryIDs {
		if countryCounts[countryID] == 0 {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	if !reflect.DeepEqual(statusCounts, expectedCountryLanguageStatusCounts) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	return nil
}

func cloneLanguageList(languages []models.Language) []models.Language {
	if len(languages) == 0 {
		return make([]models.Language, 0)
	}
	cloned := make([]models.Language, len(languages))
	copy(cloned, languages)
	return cloned
}

func cloneLanguage(language models.Language) models.Language {
	return language
}

func cloneCountryLanguageList(relations []models.CountryLanguage) []models.CountryLanguage {
	if len(relations) == 0 {
		return make([]models.CountryLanguage, 0)
	}
	cloned := make([]models.CountryLanguage, len(relations))
	copy(cloned, relations)
	return cloned
}
