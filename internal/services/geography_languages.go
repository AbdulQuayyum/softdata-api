package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

// CountryLanguageListInput carries optional relationship filters.
type CountryLanguageListInput struct {
	CountryAreaID string
	LanguageID    string
	Status        string
}

func (s *GeographyService) ListLanguages(ctx context.Context) ([]models.Language, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	languages, err := s.repository.ListLanguages(ctx, interfaces.LanguageFilter{})
	if err != nil {
		return nil, translateGeographyServiceError("list languages", err)
	}
	return cloneLanguageList(languages), nil
}

func (s *GeographyService) GetLanguage(ctx context.Context, languageID string) (models.Language, error) {
	if err := ctx.Err(); err != nil {
		return models.Language{}, err
	}

	languageID = strings.TrimSpace(languageID)
	if languageID == "" || !geographyLanguageIDPattern.MatchString(languageID) {
		return models.Language{}, ErrInvalidLanguageID
	}

	language, err := s.repository.GetLanguage(ctx, languageID)
	if err != nil {
		return models.Language{}, translateGeographyLanguageLookupError(err)
	}
	return language, nil
}

func (s *GeographyService) ListCountryLanguages(ctx context.Context, input CountryLanguageListInput) ([]models.CountryLanguage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	countryAreaID := strings.TrimSpace(input.CountryAreaID)
	if countryAreaID != "" && !geographyCountryOrAreaIDPattern.MatchString(countryAreaID) {
		return nil, ErrInvalidCountryLanguageCountryAreaID
	}
	languageID := strings.TrimSpace(input.LanguageID)
	if languageID != "" && !geographyLanguageIDPattern.MatchString(languageID) {
		return nil, ErrInvalidCountryLanguageLanguageID
	}
	status := strings.TrimSpace(input.Status)
	if status != "" {
		switch status {
		case "official", "de_facto_official", "official_regional", "used":
		default:
			return nil, ErrInvalidCountryLanguageStatus
		}
	}

	relations, err := s.repository.ListCountryLanguages(ctx, interfaces.CountryLanguageFilter{
		CountryAreaID: countryAreaID,
		LanguageID:    languageID,
		Status:        status,
	})
	if err != nil {
		return nil, translateGeographyCountryLanguageServiceError(err)
	}
	return cloneCountryLanguageList(relations), nil
}

func cloneLanguageList(languages []models.Language) []models.Language {
	if len(languages) == 0 {
		return make([]models.Language, 0)
	}
	cloned := make([]models.Language, len(languages))
	copy(cloned, languages)
	return cloned
}

func cloneCountryLanguageList(relations []models.CountryLanguage) []models.CountryLanguage {
	if len(relations) == 0 {
		return make([]models.CountryLanguage, 0)
	}
	cloned := make([]models.CountryLanguage, len(relations))
	copy(cloned, relations)
	return cloned
}

func translateGeographyLanguageLookupError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrLanguageNotFound):
		return ErrLanguageNotFound
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("get language: repository unavailable")
	default:
		return fmt.Errorf("get language: repository unavailable")
	}
}

func translateGeographyCountryLanguageServiceError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrInvalidCountryLanguageCountryAreaID):
		return ErrInvalidCountryLanguageCountryAreaID
	case errors.Is(err, interfaces.ErrInvalidCountryLanguageLanguageID):
		return ErrInvalidCountryLanguageLanguageID
	case errors.Is(err, interfaces.ErrInvalidCountryLanguageStatus):
		return ErrInvalidCountryLanguageStatus
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("list country languages: repository unavailable")
	default:
		return fmt.Errorf("list country languages: repository unavailable")
	}
}
