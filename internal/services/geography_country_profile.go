package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type countryProfileCountryService interface {
	GetCountryOrArea(context.Context, string) (models.CountryOrArea, error)
}

type countryProfileCurrencyService interface {
	ListCurrencies(context.Context, CurrencyListInput) ([]models.Currency, error)
}

type countryProfileTimeZoneService interface {
	ListTimeZones(context.Context, TimeZoneListInput) ([]models.TimeZone, error)
}

// CountryProfileService derives country profile views from the verified geography, finance and time-zone services.
type CountryProfileService struct {
	countryService  countryProfileCountryService
	currencyService countryProfileCurrencyService
	timeZoneService countryProfileTimeZoneService
}

func NewCountryProfileService(countryService countryProfileCountryService, currencyService countryProfileCurrencyService, timeZoneService countryProfileTimeZoneService) (*CountryProfileService, error) {
	if countryService == nil {
		return nil, fmt.Errorf("country service is required")
	}
	if currencyService == nil {
		return nil, fmt.Errorf("currency service is required")
	}
	if timeZoneService == nil {
		return nil, fmt.Errorf("time zone service is required")
	}
	return &CountryProfileService{
		countryService:  countryService,
		currencyService: currencyService,
		timeZoneService: timeZoneService,
	}, nil
}

func (s *CountryProfileService) GetCountryProfile(ctx context.Context, countryID string) (models.CountryProfile, error) {
	if err := ctx.Err(); err != nil {
		return models.CountryProfile{}, err
	}

	normalizedID, err := normalizeCountryOrAreaID(countryID)
	if err != nil {
		return models.CountryProfile{}, err
	}

	country, err := s.countryService.GetCountryOrArea(ctx, normalizedID)
	if err != nil {
		return models.CountryProfile{}, translateCountryProfileLookupError(err)
	}
	if err := ctx.Err(); err != nil {
		return models.CountryProfile{}, err
	}

	currencies, err := s.currencyService.ListCurrencies(ctx, CurrencyListInput{CountryAreaID: normalizedID})
	if err != nil {
		return models.CountryProfile{}, translateCountryProfileCompositionError("list country profile currencies", err)
	}
	if err := ctx.Err(); err != nil {
		return models.CountryProfile{}, err
	}

	timeZones, err := s.timeZoneService.ListTimeZones(ctx, TimeZoneListInput{CountryAreaID: normalizedID})
	if err != nil {
		return models.CountryProfile{}, translateCountryProfileCompositionError("list country profile time zones", err)
	}

	return buildCountryProfile(country, currencies, timeZones), nil
}

func normalizeCountryOrAreaID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || !geographyCountryOrAreaIDPattern.MatchString(value) {
		return "", ErrInvalidCountryOrAreaID
	}
	return value, nil
}

func buildCountryProfile(country models.CountryOrArea, currencies []models.Currency, timeZones []models.TimeZone) models.CountryProfile {
	profile := models.CountryProfile{
		ID:                     country.ID,
		Name:                   country.Name,
		Alpha2Code:             country.Alpha2Code,
		Alpha3Code:             country.Alpha3Code,
		NumericCode:            country.NumericCode,
		CallingCodes:           cloneStringSlice(country.CallingCodes),
		FlagEmoji:              country.FlagEmoji,
		FlagSVGURL:             country.FlagSVGURL,
		RegionCode:             country.RegionCode,
		RegionName:             country.RegionName,
		SubregionCode:          country.SubregionCode,
		SubregionName:          country.SubregionName,
		IntermediateRegionCode: country.IntermediateRegionCode,
		IntermediateRegionName: country.IntermediateRegionName,
		CurrencyIDs:            dedupeAndSortCurrencyIDs(currencies),
		TimeZoneIDs:            dedupeAndSortTimeZoneIDs(timeZones),
	}
	if profile.CallingCodes == nil {
		profile.CallingCodes = nil
	}
	return profile
}

func dedupeAndSortCurrencyIDs(currencies []models.Currency) []string {
	if len(currencies) == 0 {
		return make([]string, 0)
	}
	ids := make(map[string]struct{}, len(currencies))
	for _, currency := range currencies {
		ids[currency.ID] = struct{}{}
	}
	out := make([]string, 0, len(ids))
	for id := range ids {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func dedupeAndSortTimeZoneIDs(timeZones []models.TimeZone) []string {
	if len(timeZones) == 0 {
		return make([]string, 0)
	}
	ids := make(map[string]struct{}, len(timeZones))
	for _, timeZone := range timeZones {
		ids[timeZone.ID] = struct{}{}
	}
	out := make([]string, 0, len(ids))
	for id := range ids {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func cloneStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func translateCountryProfileLookupError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, ErrInvalidCountryOrAreaID):
		return ErrInvalidCountryOrAreaID
	case errors.Is(err, ErrCountryOrAreaNotFound):
		return ErrCountryOrAreaNotFound
	case errors.Is(err, interfaces.ErrCountryOrAreaNotFound):
		return ErrCountryOrAreaNotFound
	default:
		return fmt.Errorf("get country profile: repository unavailable")
	}
}

func translateCountryProfileCompositionError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("%s: repository unavailable", op)
	default:
		return fmt.Errorf("%s: repository unavailable", op)
	}
}
