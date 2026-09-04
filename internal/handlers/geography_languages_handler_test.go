package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

func TestGeographyHandlerLanguagesContract(t *testing.T) {
	stub := &geographyHandlerStub{
		languageListFn: func(context.Context) ([]models.Language, error) {
			return nil, nil
		},
		languageGetFn: func(_ context.Context, id string) (models.Language, error) {
			if id != "en" {
				t.Fatalf("unexpected language id: %q", id)
			}
			return models.Language{ID: "en", Name: "English"}, nil
		},
		countryLanguageListFn: func(_ context.Context, input services.CountryLanguageListInput) ([]models.CountryLanguage, error) {
			if input != (services.CountryLanguageListInput{CountryAreaID: "ng", LanguageID: "yo", Status: "official"}) {
				t.Fatalf("unexpected relationship input: %#v", input)
			}
			return []models.CountryLanguage{{CountryAreaID: "ng", LanguageID: "yo", Status: "official"}}, nil
		},
	}
	h, err := NewGeographyHandler(stub)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/languages", nil)
	invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) { h.ListLanguages(w, r) }, rr)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"data":[]`) {
		t.Fatalf("unexpected empty language response: %d %s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/geography/languages/en", nil)
	req.SetPathValue("language_id", " en ")
	invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) { h.GetLanguage(w, r) }, rr)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected detail status: %d", rr.Code)
	}
	var detail map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &detail); err != nil || detail["data"].(map[string]any)["id"] != "en" {
		t.Fatalf("unexpected detail body: %s", rr.Body.String())
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/geography/country-languages?country_area_id=ng&language_id=yo&status=official", nil)
	invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) { h.ListCountryLanguages(w, r) }, rr)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"language_id":"yo"`) {
		t.Fatalf("unexpected relationship response: %d %s", rr.Code, rr.Body.String())
	}
}

func TestGeographyHandlerLanguageValidationAndErrors(t *testing.T) {
	stub := &geographyHandlerStub{
		languageGetFn: func(context.Context, string) (models.Language, error) {
			return models.Language{}, services.ErrLanguageNotFound
		},
		countryLanguageListFn: func(context.Context, services.CountryLanguageListInput) ([]models.CountryLanguage, error) {
			return nil, errors.New("filesystem path leaked")
		},
	}
	h, err := NewGeographyHandler(stub)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/languages/fat", nil)
	req.SetPathValue("language_id", "fat")
	h.GetLanguage(rr, req)
	if rr.Code != http.StatusUnprocessableEntity || stub.languageGetCalls != 0 {
		t.Fatalf("deprecated id was not rejected before service access: %d calls=%d", rr.Code, stub.languageGetCalls)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/geography/languages/zz", nil)
	req.SetPathValue("language_id", "zz")
	h.GetLanguage(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("unexpected not-found status: %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/geography/country-languages?status=bad", nil)
	h.ListCountryLanguages(rr, req)
	if rr.Code != http.StatusUnprocessableEntity || stub.countryLanguageListCalls != 0 {
		t.Fatalf("invalid status was not rejected before service access: %d calls=%d", rr.Code, stub.countryLanguageListCalls)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/geography/country-languages?country_area_id=ng", nil)
	h.ListCountryLanguages(rr, req)
	if rr.Code != http.StatusInternalServerError || strings.Contains(rr.Body.String(), "filesystem path leaked") {
		t.Fatalf("unexpected internal error response: %d %s", rr.Code, rr.Body.String())
	}
}

func TestGeographyHandlerLanguageMethodGuards(t *testing.T) {
	h, err := NewGeographyHandler(&geographyHandlerStub{})
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}
	checks := []func(http.ResponseWriter, *http.Request){h.ListLanguages, h.GetLanguage, h.ListCountryLanguages}
	for _, handler := range checks {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/geography/languages", nil)
		handler(rr, req)
		if rr.Code != http.StatusMethodNotAllowed || rr.Header().Get("Allow") != http.MethodGet {
			t.Fatalf("unexpected method guard response: %d allow=%q", rr.Code, rr.Header().Get("Allow"))
		}
	}
}
