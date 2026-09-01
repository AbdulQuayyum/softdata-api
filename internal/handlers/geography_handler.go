package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
	"github.com/AbdulQuayyum/softdata-api/internal/validators"
)

type geographyService interface {
	ListStates(context.Context) ([]models.State, error)
	GetState(context.Context, string) (models.State, error)
	ListGeopoliticalZones(context.Context) ([]models.GeopoliticalZone, error)
	GetGeopoliticalZone(context.Context, string) (models.GeopoliticalZone, error)
	ListLocalGovernmentUnits(context.Context) ([]models.LocalGovernmentUnit, error)
	ListLocalGovernmentUnitsByState(context.Context, string) ([]models.LocalGovernmentUnit, error)
	GetLocalGovernmentUnit(context.Context, string) (models.LocalGovernmentUnit, error)
	ListTimeZones(context.Context, services.TimeZoneListInput) ([]models.TimeZone, error)
	GetTimeZone(context.Context, string) (models.TimeZone, error)
	ListCountriesAndAreas(context.Context, services.CountryOrAreaListInput) ([]models.CountryOrArea, error)
	GetCountryOrArea(context.Context, string) (models.CountryOrArea, error)
}

type GeographyHandler struct {
	service geographyService
}

func NewGeographyHandler(service geographyService) (*GeographyHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("geography service is required")
	}
	return &GeographyHandler{service: service}, nil
}

func (h *GeographyHandler) ListStates(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	states, err := h.service.ListStates(r.Context())
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, states)
}

func (h *GeographyHandler) GetState(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	stateID, err := validators.ValidateStateID(r.PathValue("state_id"))
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	state, err := h.service.GetState(r.Context(), stateID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, state)
}

func (h *GeographyHandler) ListGeopoliticalZones(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	zones, err := h.service.ListGeopoliticalZones(r.Context())
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, zones)
}

func (h *GeographyHandler) GetGeopoliticalZone(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	zoneID, err := validators.ValidateGeopoliticalZoneID(r.PathValue("zone_id"))
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	zone, err := h.service.GetGeopoliticalZone(r.Context(), zoneID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, zone)
}

func (h *GeographyHandler) ListLocalGovernmentUnits(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	query, err := validators.ValidateLocalGovernmentUnitListQuery(r.URL.Query())
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	var units []models.LocalGovernmentUnit
	if query.StateID == nil {
		units, err = h.service.ListLocalGovernmentUnits(r.Context())
	} else {
		units, err = h.service.ListLocalGovernmentUnitsByState(r.Context(), *query.StateID)
	}
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, units)
}

func (h *GeographyHandler) GetLocalGovernmentUnit(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	unitID, err := validators.ValidateLocalGovernmentUnitID(r.PathValue("lga_id"))
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	unit, err := h.service.GetLocalGovernmentUnit(r.Context(), unitID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, unit)
}

// ListTimeZones handles GET /v1/geography/time-zones.
func (h *GeographyHandler) ListTimeZones(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	query, err := validators.ValidateTimeZoneListQuery(r.URL.Query())
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	timeZones, err := h.service.ListTimeZones(r.Context(), query)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, timeZones)
}

// GetTimeZone handles GET /v1/geography/time-zones/{time_zone_id}.
func (h *GeographyHandler) GetTimeZone(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	timeZoneID, err := validators.ValidateTimeZoneID(r.PathValue("time_zone_id"))
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	timeZone, err := h.service.GetTimeZone(r.Context(), timeZoneID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, timeZone)
}

func (h *GeographyHandler) ListCountriesAndAreas(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	query, err := validators.ValidateCountryOrAreaListQuery(r.URL.Query())
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	countries, err := h.service.ListCountriesAndAreas(r.Context(), query)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.List(w, http.StatusOK, countries)
}

func (h *GeographyHandler) GetCountryOrArea(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	requestID := requestIDFromContext(r.Context())
	countryID, err := validators.ValidateCountryOrAreaID(r.PathValue("country_id"))
	if err != nil {
		if validationErr, ok := validationErrorsFrom(err); ok {
			_ = response.Validation(w, requestID, validationErrorsToResponse(validationErr))
			return
		}
		_ = response.Error(w, err, requestID)
		return
	}

	country, err := h.service.GetCountryOrArea(r.Context(), countryID)
	if err != nil {
		_ = response.Error(w, err, requestID)
		return
	}

	_ = response.Success(w, http.StatusOK, country)
}
