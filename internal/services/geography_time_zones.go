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

var geographyTimeZoneIDPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._+-]*(?:/[A-Za-z0-9._+-]+)+$`)
var geographyTimeZoneCountryAreaIDPattern = regexp.MustCompile(`^[a-z]{2}$`)

// TimeZoneListInput carries an optional country/area filter.
type TimeZoneListInput struct {
	CountryAreaID string
}

func (s *GeographyService) ListTimeZones(ctx context.Context, input TimeZoneListInput) ([]models.TimeZone, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	countryAreaID := strings.TrimSpace(input.CountryAreaID)
	if countryAreaID != "" && !geographyTimeZoneCountryAreaIDPattern.MatchString(countryAreaID) {
		return nil, ErrInvalidTimeZoneCountryAreaID
	}

	timeZones, err := s.repository.ListTimeZones(ctx, interfaces.TimeZoneFilter{CountryAreaID: countryAreaID})
	if err != nil {
		return nil, translateGeographyTimeZoneServiceError("list time zones", err)
	}
	return cloneTimeZoneList(timeZones), nil
}

func (s *GeographyService) GetTimeZone(ctx context.Context, timeZoneID string) (models.TimeZone, error) {
	if err := ctx.Err(); err != nil {
		return models.TimeZone{}, err
	}

	timeZoneID = strings.TrimSpace(timeZoneID)
	if timeZoneID == "" || !geographyTimeZoneIDPattern.MatchString(timeZoneID) {
		return models.TimeZone{}, ErrInvalidTimeZoneID
	}

	timeZone, err := s.repository.GetTimeZone(ctx, timeZoneID)
	if err != nil {
		return models.TimeZone{}, translateGeographyTimeZoneLookupError(err)
	}
	return timeZone, nil
}

func cloneTimeZoneList(timeZones []models.TimeZone) []models.TimeZone {
	if len(timeZones) == 0 {
		return make([]models.TimeZone, 0)
	}
	cloned := make([]models.TimeZone, len(timeZones))
	for i, timeZone := range timeZones {
		cloned[i] = cloneTimeZone(timeZone)
	}
	return cloned
}

func cloneTimeZone(timeZone models.TimeZone) models.TimeZone {
	timeZone.CountryAreaIDs = cloneTimeZoneCountryAreaIDs(timeZone.CountryAreaIDs)
	return timeZone
}

func cloneTimeZoneCountryAreaIDs(ids []string) []string {
	if len(ids) == 0 {
		return make([]string, 0)
	}
	cloned := make([]string, len(ids))
	copy(cloned, ids)
	return cloned
}

func translateGeographyTimeZoneLookupError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, interfaces.ErrTimeZoneNotFound):
		return ErrTimeZoneNotFound
	case errors.Is(err, interfaces.ErrInvalidDatasetFile), errors.Is(err, interfaces.ErrDatasetFileNotFound), errors.Is(err, interfaces.ErrDatasetFileUnavailable):
		return fmt.Errorf("get time zone: repository unavailable")
	default:
		return fmt.Errorf("get time zone: repository unavailable")
	}
}

func translateGeographyTimeZoneServiceError(op string, err error) error {
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
