package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/middlewares"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type geographyHandlerStub struct {
	listFn                func(context.Context) ([]models.State, error)
	getFn                 func(context.Context, string) (models.State, error)
	zoneListFn            func(context.Context) ([]models.GeopoliticalZone, error)
	zoneGetFn             func(context.Context, string) (models.GeopoliticalZone, error)
	lgaListFn             func(context.Context) ([]models.LocalGovernmentUnit, error)
	lgaListByFn           func(context.Context, string) ([]models.LocalGovernmentUnit, error)
	lgaGetFn              func(context.Context, string) (models.LocalGovernmentUnit, error)
	languageListFn        func(context.Context) ([]models.Language, error)
	languageGetFn         func(context.Context, string) (models.Language, error)
	countryLanguageListFn func(context.Context, services.CountryLanguageListInput) ([]models.CountryLanguage, error)
	timeZoneListFn        func(context.Context, services.TimeZoneListInput) ([]models.TimeZone, error)
	timeZoneGetFn         func(context.Context, string) (models.TimeZone, error)
	countryListFn         func(context.Context, services.CountryOrAreaListInput) ([]models.CountryOrArea, error)
	countryGetFn          func(context.Context, string) (models.CountryOrArea, error)
	profileFn             func(context.Context, string) (models.CountryProfile, error)

	listCalls                int
	getCalls                 int
	zoneListCalls            int
	zoneGetCalls             int
	lgaListCalls             int
	lgaListByCalls           int
	lgaGetCalls              int
	languageListCalls        int
	languageGetCalls         int
	countryLanguageListCalls int
	timeZoneListCalls        int
	timeZoneGetCalls         int
	countryListCalls         int
	countryGetCalls          int
	profileCalls             int
	lastID                   string
	lastZoneID               string
	lastLGAID                string
	lastLanguageID           string
	lastCountryLanguageQuery services.CountryLanguageListInput
	lastTimeZoneID           string
	lastTimeZoneQuery        services.TimeZoneListInput
	lastStateID              string
	lastCountryID            string
	lastCountryQuery         services.CountryOrAreaListInput
	lastProfileID            string
}

func (s *geographyHandlerStub) ListStates(ctx context.Context) ([]models.State, error) {
	s.listCalls++
	if s.listFn != nil {
		return s.listFn(ctx)
	}
	return nil, nil
}

func (s *geographyHandlerStub) GetState(ctx context.Context, stateID string) (models.State, error) {
	s.getCalls++
	s.lastID = stateID
	if s.getFn != nil {
		return s.getFn(ctx, stateID)
	}
	return models.State{}, nil
}

func (s *geographyHandlerStub) ListGeopoliticalZones(ctx context.Context) ([]models.GeopoliticalZone, error) {
	s.zoneListCalls++
	if s.zoneListFn != nil {
		return s.zoneListFn(ctx)
	}
	return nil, nil
}

func (s *geographyHandlerStub) GetGeopoliticalZone(ctx context.Context, zoneID string) (models.GeopoliticalZone, error) {
	s.zoneGetCalls++
	s.lastZoneID = zoneID
	if s.zoneGetFn != nil {
		return s.zoneGetFn(ctx, zoneID)
	}
	return models.GeopoliticalZone{}, nil
}

func (s *geographyHandlerStub) ListLocalGovernmentUnits(ctx context.Context) ([]models.LocalGovernmentUnit, error) {
	s.lgaListCalls++
	if s.lgaListFn != nil {
		return s.lgaListFn(ctx)
	}
	return nil, nil
}

func (s *geographyHandlerStub) ListLocalGovernmentUnitsByState(ctx context.Context, stateID string) ([]models.LocalGovernmentUnit, error) {
	s.lgaListByCalls++
	s.lastStateID = stateID
	if s.lgaListByFn != nil {
		return s.lgaListByFn(ctx, stateID)
	}
	return nil, nil
}

func (s *geographyHandlerStub) GetLocalGovernmentUnit(ctx context.Context, unitID string) (models.LocalGovernmentUnit, error) {
	s.lgaGetCalls++
	s.lastLGAID = unitID
	if s.lgaGetFn != nil {
		return s.lgaGetFn(ctx, unitID)
	}
	return models.LocalGovernmentUnit{}, nil
}

func (s *geographyHandlerStub) ListLanguages(ctx context.Context) ([]models.Language, error) {
	s.languageListCalls++
	if s.languageListFn != nil {
		return s.languageListFn(ctx)
	}
	return nil, nil
}

func (s *geographyHandlerStub) GetLanguage(ctx context.Context, languageID string) (models.Language, error) {
	s.languageGetCalls++
	s.lastLanguageID = languageID
	if s.languageGetFn != nil {
		return s.languageGetFn(ctx, languageID)
	}
	return models.Language{}, nil
}

func (s *geographyHandlerStub) ListCountryLanguages(ctx context.Context, input services.CountryLanguageListInput) ([]models.CountryLanguage, error) {
	s.countryLanguageListCalls++
	s.lastCountryLanguageQuery = input
	if s.countryLanguageListFn != nil {
		return s.countryLanguageListFn(ctx, input)
	}
	return nil, nil
}

func (s *geographyHandlerStub) ListTimeZones(ctx context.Context, input services.TimeZoneListInput) ([]models.TimeZone, error) {
	s.timeZoneListCalls++
	s.lastTimeZoneQuery = input
	if s.timeZoneListFn != nil {
		return s.timeZoneListFn(ctx, input)
	}
	return nil, nil
}

func (s *geographyHandlerStub) GetTimeZone(ctx context.Context, timeZoneID string) (models.TimeZone, error) {
	s.timeZoneGetCalls++
	s.lastTimeZoneID = timeZoneID
	if s.timeZoneGetFn != nil {
		return s.timeZoneGetFn(ctx, timeZoneID)
	}
	return models.TimeZone{}, nil
}

func (s *geographyHandlerStub) ListCountriesAndAreas(ctx context.Context, input services.CountryOrAreaListInput) ([]models.CountryOrArea, error) {
	s.countryListCalls++
	s.lastCountryQuery = input
	if s.countryListFn != nil {
		return s.countryListFn(ctx, input)
	}
	return nil, nil
}

func (s *geographyHandlerStub) GetCountryOrArea(ctx context.Context, countryID string) (models.CountryOrArea, error) {
	s.countryGetCalls++
	s.lastCountryID = countryID
	if s.countryGetFn != nil {
		return s.countryGetFn(ctx, countryID)
	}
	return models.CountryOrArea{}, nil
}

func (s *geographyHandlerStub) GetCountryProfile(ctx context.Context, countryID string) (models.CountryProfile, error) {
	s.profileCalls++
	s.lastProfileID = countryID
	if s.profileFn != nil {
		return s.profileFn(ctx, countryID)
	}
	return models.CountryProfile{}, nil
}

func TestNewGeographyHandlerRejectsNilService(t *testing.T) {
	if _, err := NewGeographyHandler(nil); err == nil {
		t.Fatal("expected nil service to be rejected")
	}
}

func TestGeographyHandlerListStates(t *testing.T) {
	stub := &geographyHandlerStub{
		listFn: func(ctx context.Context) ([]models.State, error) {
			if _, ok := middlewares.RequestIDFromContext(ctx); !ok {
				t.Fatal("request context was not preserved")
			}
			return []models.State{{ID: "abia", Name: "Abia", OfficialName: "Abia State", AdministrativeType: "state", Capital: "Umuahia", GeopoliticalZoneID: "south-east", CountryCode: "NG"}}, nil
		},
	}
	h, err := NewGeographyHandler(stub)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/states", nil)
	invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
		h.ListStates(w, r)
	}, rr)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.listCalls != 1 {
		t.Fatalf("unexpected list call count: %d", stub.listCalls)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body["success"] != true {
		t.Fatalf("unexpected success flag: %#v", body["success"])
	}
	data := body["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("unexpected list payload: %#v", data)
	}
	item := data[0].(map[string]any)
	if len(item) != 7 {
		t.Fatalf("unexpected state field count: %#v", item)
	}
	if item["id"] != "abia" || item["name"] != "Abia" {
		t.Fatalf("unexpected state payload: %#v", item)
	}
}

func TestGeographyHandlerListStatesNormalizesEmptySlices(t *testing.T) {
	stub := &geographyHandlerStub{listFn: func(context.Context) ([]models.State, error) { return nil, nil }}
	h, err := NewGeographyHandler(stub)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/states", nil)
	invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
		h.ListStates(w, r)
	}, rr)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if _, ok := body["data"].([]any); !ok {
		t.Fatalf("expected array data, got %#v", body["data"])
	}
}

func TestGeographyHandlerGetState(t *testing.T) {
	stub := &geographyHandlerStub{
		getFn: func(ctx context.Context, stateID string) (models.State, error) {
			if _, ok := middlewares.RequestIDFromContext(ctx); !ok {
				t.Fatal("request context was not preserved")
			}
			if stateID != "akwa-ibom" {
				t.Fatalf("unexpected state id: %q", stateID)
			}
			return models.State{ID: "akwa-ibom", Name: "Akwa Ibom", OfficialName: "Akwa Ibom State", AdministrativeType: "state", Capital: "Uyo", GeopoliticalZoneID: "south-south", CountryCode: "NG"}, nil
		},
	}
	h, err := NewGeographyHandler(stub)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/states/akwa-ibom", nil)
	req.SetPathValue("state_id", " akwa-ibom ")
	invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
		h.GetState(w, r)
	}, rr)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.getCalls != 1 {
		t.Fatalf("unexpected get call count: %d", stub.getCalls)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	data := body["data"].(map[string]any)
	if data["id"] != "akwa-ibom" || data["country_code"] != "NG" {
		t.Fatalf("unexpected state response: %#v", data)
	}
}

func TestGeographyHandlerRejectsInvalidStateIDs(t *testing.T) {
	stub := &geographyHandlerStub{}
	h, err := NewGeographyHandler(stub)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}

	for _, value := range []string{"Abia", "akwa_ibom", "../abia", ""} {
		t.Run(value, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/geography/states/invalid", nil)
			req.SetPathValue("state_id", value)
			invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
				h.GetState(w, r)
			}, rr)

			if rr.Code != http.StatusUnprocessableEntity {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			if stub.getCalls != 0 {
				t.Fatalf("service should not be called for invalid ids")
			}
		})
	}
}

func TestGeographyHandlerErrorsAndMethodGuard(t *testing.T) {
	t.Run("missing state", func(t *testing.T) {
		stub := &geographyHandlerStub{
			getFn: func(context.Context, string) (models.State, error) {
				return models.State{}, services.ErrStateNotFound
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/states/abia", nil)
		req.SetPathValue("state_id", "abia")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetState(w, r)
		}, rr)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "req_test") {
			t.Fatalf("request id missing from error response: %s", rr.Body.String())
		}
	})

	t.Run("wrapped missing state", func(t *testing.T) {
		stub := &geographyHandlerStub{
			getFn: func(context.Context, string) (models.State, error) {
				return models.State{}, fmtWrappedErr(services.ErrStateNotFound)
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/states/abia", nil)
		req.SetPathValue("state_id", "abia")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetState(w, r)
		}, rr)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
	})

	t.Run("unexpected service error", func(t *testing.T) {
		stub := &geographyHandlerStub{
			getFn: func(context.Context, string) (models.State, error) {
				return models.State{}, errors.New("database down")
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/states/abia", nil)
		req.SetPathValue("state_id", "abia")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetState(w, r)
		}, rr)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if strings.Contains(rr.Body.String(), "database down") {
			t.Fatalf("internal error details leaked: %s", rr.Body.String())
		}
	})

	t.Run("method guard", func(t *testing.T) {
		h, err := NewGeographyHandler(&geographyHandlerStub{})
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/geography/states", nil)
		h.ListStates(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if got := rr.Header().Get("Allow"); got != http.MethodGet {
			t.Fatalf("unexpected allow header: %q", got)
		}
	})
}

func TestGeographyHandlerListGeopoliticalZones(t *testing.T) {
	stub := &geographyHandlerStub{
		zoneListFn: func(ctx context.Context) ([]models.GeopoliticalZone, error) {
			if _, ok := middlewares.RequestIDFromContext(ctx); !ok {
				t.Fatal("request context was not preserved")
			}
			return []models.GeopoliticalZone{{ID: "north-central", Name: "North Central", CountryCode: "NG"}}, nil
		},
	}
	h, err := NewGeographyHandler(stub)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones", nil)
	invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
		h.ListGeopoliticalZones(w, r)
	}, rr)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.zoneListCalls != 1 {
		t.Fatalf("unexpected zone list call count: %d", stub.zoneListCalls)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body["success"] != true {
		t.Fatalf("unexpected success flag: %#v", body["success"])
	}
	data := body["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("unexpected zone payload: %#v", data)
	}
	item := data[0].(map[string]any)
	if len(item) != 3 {
		t.Fatalf("unexpected zone field count: %#v", item)
	}
	if item["id"] != "north-central" || item["name"] != "North Central" {
		t.Fatalf("unexpected zone payload: %#v", item)
	}
}

func TestGeographyHandlerGetGeopoliticalZone(t *testing.T) {
	stub := &geographyHandlerStub{
		zoneGetFn: func(ctx context.Context, zoneID string) (models.GeopoliticalZone, error) {
			if _, ok := middlewares.RequestIDFromContext(ctx); !ok {
				t.Fatal("request context was not preserved")
			}
			if zoneID != "north-central" {
				t.Fatalf("unexpected zone id: %q", zoneID)
			}
			return models.GeopoliticalZone{ID: "north-central", Name: "North Central", CountryCode: "NG"}, nil
		},
	}
	h, err := NewGeographyHandler(stub)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones/north-central", nil)
	req.SetPathValue("zone_id", " north-central ")
	invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
		h.GetGeopoliticalZone(w, r)
	}, rr)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.zoneGetCalls != 1 {
		t.Fatalf("unexpected zone get call count: %d", stub.zoneGetCalls)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	data := body["data"].(map[string]any)
	if data["id"] != "north-central" || data["country_code"] != "NG" {
		t.Fatalf("unexpected zone response: %#v", data)
	}
}

func TestGeographyHandlerRejectsInvalidGeopoliticalZoneIDs(t *testing.T) {
	stub := &geographyHandlerStub{}
	h, err := NewGeographyHandler(stub)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}

	for _, value := range []string{"North Central", "north_central", "../north-central", "", "central"} {
		t.Run(value, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones/invalid", nil)
			req.SetPathValue("zone_id", value)
			invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
				h.GetGeopoliticalZone(w, r)
			}, rr)

			if rr.Code != http.StatusUnprocessableEntity {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			if stub.zoneGetCalls != 0 {
				t.Fatalf("service should not be called for invalid ids")
			}
		})
	}
}

func TestGeographyHandlerGeopoliticalZoneErrors(t *testing.T) {
	t.Run("missing zone", func(t *testing.T) {
		stub := &geographyHandlerStub{
			zoneGetFn: func(context.Context, string) (models.GeopoliticalZone, error) {
				return models.GeopoliticalZone{}, services.ErrGeopoliticalZoneNotFound
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones/north-central", nil)
		req.SetPathValue("zone_id", "north-central")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetGeopoliticalZone(w, r)
		}, rr)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
	})

	t.Run("wrapped missing zone", func(t *testing.T) {
		stub := &geographyHandlerStub{
			zoneGetFn: func(context.Context, string) (models.GeopoliticalZone, error) {
				return models.GeopoliticalZone{}, fmtWrappedErr(services.ErrGeopoliticalZoneNotFound)
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones/north-central", nil)
		req.SetPathValue("zone_id", "north-central")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetGeopoliticalZone(w, r)
		}, rr)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
	})

	t.Run("invalid service id", func(t *testing.T) {
		stub := &geographyHandlerStub{
			zoneGetFn: func(context.Context, string) (models.GeopoliticalZone, error) {
				return models.GeopoliticalZone{}, services.ErrInvalidGeopoliticalZoneID
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones/north-central", nil)
		req.SetPathValue("zone_id", "north-central")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetGeopoliticalZone(w, r)
		}, rr)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
	})

	t.Run("unexpected service error", func(t *testing.T) {
		stub := &geographyHandlerStub{
			zoneGetFn: func(context.Context, string) (models.GeopoliticalZone, error) {
				return models.GeopoliticalZone{}, errors.New("database down")
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones/north-central", nil)
		req.SetPathValue("zone_id", "north-central")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetGeopoliticalZone(w, r)
		}, rr)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if strings.Contains(rr.Body.String(), "database down") {
			t.Fatalf("internal error details leaked: %s", rr.Body.String())
		}
	})
}

func TestGeographyHandlerListLocalGovernmentUnits(t *testing.T) {
	t.Run("list all", func(t *testing.T) {
		stub := &geographyHandlerStub{
			lgaListFn: func(ctx context.Context) ([]models.LocalGovernmentUnit, error) {
				if _, ok := middlewares.RequestIDFromContext(ctx); !ok {
					t.Fatal("request context was not preserved")
				}
				return []models.LocalGovernmentUnit{
					{ID: "lagos-ikeja", Name: "Ikeja", StateID: "lagos", CountryCode: "NG", AdministrativeType: "local_government_area"},
				}, nil
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListLocalGovernmentUnits(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.lgaListCalls != 1 || stub.lgaListByCalls != 0 {
			t.Fatalf("unexpected call counts: list=%d by=%d", stub.lgaListCalls, stub.lgaListByCalls)
		}
		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		data := body["data"].([]any)
		if len(data) != 1 {
			t.Fatalf("unexpected lga payload: %#v", data)
		}
		item := data[0].(map[string]any)
		if len(item) != 5 {
			t.Fatalf("unexpected lga field count: %#v", item)
		}
		if item["id"] != "lagos-ikeja" || item["administrative_type"] != "local_government_area" {
			t.Fatalf("unexpected lga payload: %#v", item)
		}
	})

	t.Run("normalize nil slice", func(t *testing.T) {
		stub := &geographyHandlerStub{
			lgaListFn: func(context.Context) ([]models.LocalGovernmentUnit, error) {
				return nil, nil
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListLocalGovernmentUnits(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		data := body["data"].([]any)
		if len(data) != 0 {
			t.Fatalf("unexpected nil normalization payload: %#v", data)
		}
	})

	t.Run("list by state", func(t *testing.T) {
		stub := &geographyHandlerStub{
			lgaListByFn: func(ctx context.Context, stateID string) ([]models.LocalGovernmentUnit, error) {
				if _, ok := middlewares.RequestIDFromContext(ctx); !ok {
					t.Fatal("request context was not preserved")
				}
				if stateID != "fct" {
					t.Fatalf("unexpected state id: %q", stateID)
				}
				return []models.LocalGovernmentUnit{
					{ID: "fct-abaji", Name: "Abaji", StateID: "fct", CountryCode: "NG", AdministrativeType: "area_council"},
				}, nil
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas?state_id=fct", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListLocalGovernmentUnits(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.lgaListCalls != 0 || stub.lgaListByCalls != 1 {
			t.Fatalf("unexpected call counts: list=%d by=%d", stub.lgaListCalls, stub.lgaListByCalls)
		}
		if stub.lastStateID != "fct" {
			t.Fatalf("unexpected state id: %q", stub.lastStateID)
		}
		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		item := body["data"].([]any)[0].(map[string]any)
		if item["administrative_type"] != "area_council" {
			t.Fatalf("unexpected lga payload: %#v", item)
		}
	})

	t.Run("reject invalid query", func(t *testing.T) {
		stub := &geographyHandlerStub{}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		for _, query := range []string{"state_id=", "state_id=%20%20%20", "state_id=Lagos", "state_id=lagos&state_id=fct", "state_id=north-central"} {
			t.Run(query, func(t *testing.T) {
				rr := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas?"+query, nil)
				invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
					h.ListLocalGovernmentUnits(w, r)
				}, rr)
				if rr.Code != http.StatusUnprocessableEntity {
					t.Fatalf("unexpected status: %d", rr.Code)
				}
				if !strings.Contains(rr.Body.String(), "req_test") {
					t.Fatalf("request id missing from validation response: %s", rr.Body.String())
				}
				if stub.lgaListCalls != 0 || stub.lgaListByCalls != 0 {
					t.Fatalf("service should not be called for invalid query")
				}
			})
		}
	})

	t.Run("method guard", func(t *testing.T) {
		h, err := NewGeographyHandler(&geographyHandlerStub{})
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/geography/lgas", nil)
		h.ListLocalGovernmentUnits(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if got := rr.Header().Get("Allow"); got != http.MethodGet {
			t.Fatalf("unexpected allow header: %q", got)
		}
	})
}

func TestGeographyHandlerGetLocalGovernmentUnit(t *testing.T) {
	stub := &geographyHandlerStub{
		lgaGetFn: func(ctx context.Context, unitID string) (models.LocalGovernmentUnit, error) {
			if _, ok := middlewares.RequestIDFromContext(ctx); !ok {
				t.Fatal("request context was not preserved")
			}
			if unitID != "akwa-ibom-urue-offong-oruko" {
				t.Fatalf("unexpected lga id: %q", unitID)
			}
			return models.LocalGovernmentUnit{
				ID:                 unitID,
				Name:               "Urue-Offong/Oruko",
				StateID:            "akwa-ibom",
				CountryCode:        "NG",
				AdministrativeType: "local_government_area",
			}, nil
		},
	}
	h, err := NewGeographyHandler(stub)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas/akwa-ibom-urue-offong-oruko", nil)
	req.SetPathValue("lga_id", " akwa-ibom-urue-offong-oruko ")
	invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
		h.GetLocalGovernmentUnit(w, r)
	}, rr)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.lgaGetCalls != 1 {
		t.Fatalf("unexpected get call count: %d", stub.lgaGetCalls)
	}
	if stub.lastLGAID != "akwa-ibom-urue-offong-oruko" {
		t.Fatalf("unexpected lga id: %q", stub.lastLGAID)
	}
	if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("unexpected content type: %q", got)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	data := body["data"].(map[string]any)
	if len(data) != 5 {
		t.Fatalf("unexpected lga response: %#v", data)
	}
	if data["name"] != "Urue-Offong/Oruko" || data["administrative_type"] != "local_government_area" {
		t.Fatalf("unexpected lga response: %#v", data)
	}
}

func TestGeographyHandlerGetLocalGovernmentUnitMethodGuard(t *testing.T) {
	h, err := NewGeographyHandler(&geographyHandlerStub{})
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/geography/lgas/lagos-ikeja", nil)
	req.SetPathValue("lga_id", "lagos-ikeja")
	h.GetLocalGovernmentUnit(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("unexpected allow header: %q", got)
	}
}

func TestGeographyHandlerListCountriesAndAreas(t *testing.T) {
	t.Run("list all", func(t *testing.T) {
		stub := &geographyHandlerStub{
			countryListFn: func(ctx context.Context, input services.CountryOrAreaListInput) ([]models.CountryOrArea, error) {
				if _, ok := middlewares.RequestIDFromContext(ctx); !ok {
					t.Fatal("request context was not preserved")
				}
				if input.RegionCode != "" || input.SubregionCode != "" {
					t.Fatalf("unexpected query input: %#v", input)
				}
				return []models.CountryOrArea{{ID: "ng", Name: "Nigeria", Alpha2Code: "NG", Alpha3Code: "NGA", NumericCode: "566", FlagEmoji: "🇳🇬", FlagSVGURL: "/v1/assets/flags/ng.svg", RegionCode: "002", RegionName: "Africa", SubregionCode: "015", SubregionName: "Northern Africa", IntermediateRegionCode: "014", IntermediateRegionName: "Eastern Africa"}}, nil
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListCountriesAndAreas(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.countryListCalls != 1 || stub.countryGetCalls != 0 {
			t.Fatalf("unexpected call counts: list=%d get=%d", stub.countryListCalls, stub.countryGetCalls)
		}
		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		data := body["data"].([]any)
		if len(data) != 1 {
			t.Fatalf("unexpected country payload: %#v", data)
		}
		item := data[0].(map[string]any)
		if len(item) != 13 {
			t.Fatalf("unexpected country field count: %#v", item)
		}
		if item["id"] != "ng" || item["alpha_2_code"] != "NG" || item["numeric_code"] != "566" || item["flag_emoji"] != "🇳🇬" || item["flag_svg_url"] != "/v1/assets/flags/ng.svg" {
			t.Fatalf("unexpected country payload: %#v", item)
		}
	})

	t.Run("filter by region", func(t *testing.T) {
		stub := &geographyHandlerStub{
			countryListFn: func(ctx context.Context, input services.CountryOrAreaListInput) ([]models.CountryOrArea, error) {
				if input.RegionCode != "002" || input.SubregionCode != "" {
					t.Fatalf("unexpected query input: %#v", input)
				}
				return []models.CountryOrArea{{ID: "ng", Name: "Nigeria", Alpha2Code: "NG", Alpha3Code: "NGA", NumericCode: "566", FlagEmoji: "🇳🇬", FlagSVGURL: "/v1/assets/flags/ng.svg"}}, nil
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries?region_code=002", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListCountriesAndAreas(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.countryListCalls != 1 || stub.countryGetCalls != 0 {
			t.Fatalf("unexpected call counts: list=%d get=%d", stub.countryListCalls, stub.countryGetCalls)
		}
	})

	t.Run("filter by subregion", func(t *testing.T) {
		stub := &geographyHandlerStub{
			countryListFn: func(ctx context.Context, input services.CountryOrAreaListInput) ([]models.CountryOrArea, error) {
				if input.RegionCode != "" || input.SubregionCode != "015" {
					t.Fatalf("unexpected query input: %#v", input)
				}
				return []models.CountryOrArea{{ID: "ng", Name: "Nigeria", Alpha2Code: "NG", Alpha3Code: "NGA", NumericCode: "566", FlagEmoji: "🇳🇬", FlagSVGURL: "/v1/assets/flags/ng.svg"}}, nil
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries?subregion_code=015", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListCountriesAndAreas(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
	})

	t.Run("filter by both", func(t *testing.T) {
		stub := &geographyHandlerStub{
			countryListFn: func(ctx context.Context, input services.CountryOrAreaListInput) ([]models.CountryOrArea, error) {
				if input.RegionCode != "002" || input.SubregionCode != "015" {
					t.Fatalf("unexpected query input: %#v", input)
				}
				return []models.CountryOrArea{{ID: "ng", Name: "Nigeria", Alpha2Code: "NG", Alpha3Code: "NGA", NumericCode: "566", FlagEmoji: "🇳🇬", FlagSVGURL: "/v1/assets/flags/ng.svg"}}, nil
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries?region_code=002&subregion_code=015", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListCountriesAndAreas(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
	})

	t.Run("reject invalid query", func(t *testing.T) {
		stub := &geographyHandlerStub{}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		for _, query := range []string{"region_code=", "region_code=02", "region_code=002&region_code=019", "subregion_code=", "subregion_code=2"} {
			t.Run(query, func(t *testing.T) {
				rr := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries?"+query, nil)
				invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
					h.ListCountriesAndAreas(w, r)
				}, rr)
				if rr.Code != http.StatusUnprocessableEntity {
					t.Fatalf("unexpected status: %d", rr.Code)
				}
				if stub.countryListCalls != 0 {
					t.Fatalf("service should not be called for invalid query")
				}
			})
		}
	})

	t.Run("method guard", func(t *testing.T) {
		h, err := NewGeographyHandler(&geographyHandlerStub{})
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/geography/countries", nil)
		h.ListCountriesAndAreas(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if got := rr.Header().Get("Allow"); got != http.MethodGet {
			t.Fatalf("unexpected allow header: %q", got)
		}
	})
}

func TestGeographyHandlerGetCountryOrArea(t *testing.T) {
	stub := &geographyHandlerStub{
		countryGetFn: func(ctx context.Context, countryID string) (models.CountryOrArea, error) {
			if _, ok := middlewares.RequestIDFromContext(ctx); !ok {
				t.Fatal("request context was not preserved")
			}
			if countryID != "ng" {
				t.Fatalf("unexpected country id: %q", countryID)
			}
			return models.CountryOrArea{ID: "ng", Name: "Nigeria", Alpha2Code: "NG", Alpha3Code: "NGA", NumericCode: "566", FlagEmoji: "🇳🇬", FlagSVGURL: "/v1/assets/flags/ng.svg"}, nil
		},
	}
	h, err := NewGeographyHandler(stub)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries/ng", nil)
	req.SetPathValue("country_id", " ng ")
	invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
		h.GetCountryOrArea(w, r)
	}, rr)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.countryGetCalls != 1 {
		t.Fatalf("unexpected get call count: %d", stub.countryGetCalls)
	}
	if stub.lastCountryID != "ng" {
		t.Fatalf("unexpected country id: %q", stub.lastCountryID)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	data := body["data"].(map[string]any)
	if len(data) != 7 {
		t.Fatalf("unexpected country response: %#v", data)
	}
	if _, ok := data["region_code"]; ok {
		t.Fatalf("unexpected optional field: %#v", data)
	}
	if data["alpha_2_code"] != "NG" || data["alpha_3_code"] != "NGA" || data["numeric_code"] != "566" || data["flag_emoji"] != "🇳🇬" || data["flag_svg_url"] != "/v1/assets/flags/ng.svg" {
		t.Fatalf("unexpected country response: %#v", data)
	}
}

func TestGeographyHandlerGetCountryProfile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		stub := &geographyHandlerStub{
			profileFn: func(ctx context.Context, countryID string) (models.CountryProfile, error) {
				if _, ok := middlewares.RequestIDFromContext(ctx); !ok {
					t.Fatal("request context was not preserved")
				}
				if countryID != "ng" {
					t.Fatalf("unexpected country id: %q", countryID)
				}
				return models.CountryProfile{
					ID:          "ng",
					Name:        "Nigeria",
					Alpha2Code:  "NG",
					Alpha3Code:  "NGA",
					NumericCode: "566",
					FlagEmoji:   "🇳🇬",
					FlagSVGURL:  "/v1/assets/flags/ng.svg",
					CurrencyIDs: []string{},
					TimeZoneIDs: []string{},
					LanguageIDs: []string{"ann", "en", "ha", "yo"},
				}, nil
			},
		}
		h, err := NewGeographyHandler(stub, stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries/ng/profile", nil)
		req.SetPathValue("country_id", " ng ")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetCountryProfile(w, r)
		}, rr)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.profileCalls != 1 {
			t.Fatalf("unexpected profile call count: %d", stub.profileCalls)
		}
		if stub.lastProfileID != "ng" {
			t.Fatalf("unexpected profile id: %q", stub.lastProfileID)
		}

		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		data := body["data"].(map[string]any)
		if len(data) != 10 {
			t.Fatalf("unexpected profile field count: %#v", data)
		}
		if data["id"] != "ng" || data["name"] != "Nigeria" {
			t.Fatalf("unexpected profile response: %#v", data)
		}
		if _, ok := data["currency_ids"].([]any); !ok {
			t.Fatalf("expected currency_ids array, got %#v", data["currency_ids"])
		}
		if _, ok := data["time_zone_ids"].([]any); !ok {
			t.Fatalf("expected time_zone_ids array, got %#v", data["time_zone_ids"])
		}
		if got, ok := data["language_ids"].([]any); !ok || len(got) != 4 || got[0] != "ann" || got[3] != "yo" {
			t.Fatalf("expected language_ids array, got %#v", data["language_ids"])
		}
	})

	t.Run("reject invalid id", func(t *testing.T) {
		stub := &geographyHandlerStub{}
		h, err := NewGeographyHandler(stub, stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries/NG/profile", nil)
		req.SetPathValue("country_id", "NG")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetCountryProfile(w, r)
		}, rr)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.profileCalls != 0 {
			t.Fatalf("service should not be called for invalid ids")
		}
	})

	t.Run("errors", func(t *testing.T) {
		stub := &geographyHandlerStub{
			profileFn: func(context.Context, string) (models.CountryProfile, error) {
				return models.CountryProfile{}, services.ErrCountryOrAreaNotFound
			},
		}
		h, err := NewGeographyHandler(stub, stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries/ng/profile", nil)
		req.SetPathValue("country_id", "ng")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetCountryProfile(w, r)
		}, rr)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}

		stub.profileFn = func(context.Context, string) (models.CountryProfile, error) {
			return models.CountryProfile{}, errors.New("database down")
		}
		rr = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/v1/geography/countries/ng/profile", nil)
		req.SetPathValue("country_id", "ng")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetCountryProfile(w, r)
		}, rr)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if strings.Contains(rr.Body.String(), "database down") {
			t.Fatalf("internal error details leaked: %s", rr.Body.String())
		}
	})

	t.Run("method guard", func(t *testing.T) {
		h, err := NewGeographyHandler(&geographyHandlerStub{}, &geographyHandlerStub{})
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/geography/countries/ng/profile", nil)
		req.SetPathValue("country_id", "ng")
		h.GetCountryProfile(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if got := rr.Header().Get("Allow"); got != http.MethodGet {
			t.Fatalf("unexpected allow header: %q", got)
		}
	})
}

func TestGeographyHandlerGetCountryOrAreaErrors(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		stub := &geographyHandlerStub{
			countryGetFn: func(context.Context, string) (models.CountryOrArea, error) {
				return models.CountryOrArea{}, services.ErrCountryOrAreaNotFound
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries/ng", nil)
		req.SetPathValue("country_id", "ng")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetCountryOrArea(w, r)
		}, rr)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
	})

	t.Run("wrapped not found", func(t *testing.T) {
		stub := &geographyHandlerStub{
			countryGetFn: func(context.Context, string) (models.CountryOrArea, error) {
				return models.CountryOrArea{}, fmtWrappedErr(services.ErrCountryOrAreaNotFound)
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries/ng", nil)
		req.SetPathValue("country_id", "ng")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetCountryOrArea(w, r)
		}, rr)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
	})

	t.Run("invalid service id", func(t *testing.T) {
		stub := &geographyHandlerStub{
			countryGetFn: func(context.Context, string) (models.CountryOrArea, error) {
				return models.CountryOrArea{}, services.ErrInvalidCountryOrAreaID
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries/ng", nil)
		req.SetPathValue("country_id", "ng")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetCountryOrArea(w, r)
		}, rr)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
	})

	t.Run("unexpected service error", func(t *testing.T) {
		stub := &geographyHandlerStub{
			countryGetFn: func(context.Context, string) (models.CountryOrArea, error) {
				return models.CountryOrArea{}, errors.New("database down")
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries/ng", nil)
		req.SetPathValue("country_id", "ng")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetCountryOrArea(w, r)
		}, rr)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if strings.Contains(rr.Body.String(), "database down") {
			t.Fatalf("internal error details leaked: %s", rr.Body.String())
		}
	})

	t.Run("invalid id rejected before service", func(t *testing.T) {
		stub := &geographyHandlerStub{}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries/invalid", nil)
		req.SetPathValue("country_id", "NG")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetCountryOrArea(w, r)
		}, rr)
		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.countryGetCalls != 0 {
			t.Fatalf("service should not be called for invalid ids")
		}
	})

	t.Run("method guard", func(t *testing.T) {
		h, err := NewGeographyHandler(&geographyHandlerStub{})
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/geography/countries/ng", nil)
		req.SetPathValue("country_id", "ng")
		h.GetCountryOrArea(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if got := rr.Header().Get("Allow"); got != http.MethodGet {
			t.Fatalf("unexpected allow header: %q", got)
		}
	})
}

func TestGeographyHandlerRejectsInvalidLocalGovernmentUnitIDs(t *testing.T) {
	stub := &geographyHandlerStub{}
	h, err := NewGeographyHandler(stub)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}

	for _, value := range []string{"Aba North", "LAGOS-IKEJA", "lagos_ikeja", "lagos/ikeja", "fct", "north-central", "550e8400-e29b-41d4-a716-446655440000", ""} {
		t.Run(value, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas/invalid", nil)
			req.SetPathValue("lga_id", value)
			invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
				h.GetLocalGovernmentUnit(w, r)
			}, rr)

			if rr.Code != http.StatusUnprocessableEntity {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			if stub.lgaGetCalls != 0 {
				t.Fatalf("service should not be called for invalid ids")
			}
		})
	}
}

func TestGeographyHandlerLocalGovernmentUnitErrors(t *testing.T) {
	t.Run("missing unit", func(t *testing.T) {
		stub := &geographyHandlerStub{
			lgaGetFn: func(context.Context, string) (models.LocalGovernmentUnit, error) {
				return models.LocalGovernmentUnit{}, services.ErrLocalGovernmentUnitNotFound
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas/lagos-ikeja", nil)
		req.SetPathValue("lga_id", "lagos-ikeja")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetLocalGovernmentUnit(w, r)
		}, rr)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "req_test") {
			t.Fatalf("request id missing from error response: %s", rr.Body.String())
		}
	})

	t.Run("wrapped missing unit", func(t *testing.T) {
		stub := &geographyHandlerStub{
			lgaGetFn: func(context.Context, string) (models.LocalGovernmentUnit, error) {
				return models.LocalGovernmentUnit{}, fmtWrappedErr(services.ErrLocalGovernmentUnitNotFound)
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas/lagos-ikeja", nil)
		req.SetPathValue("lga_id", "lagos-ikeja")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetLocalGovernmentUnit(w, r)
		}, rr)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
	})

	t.Run("invalid service id", func(t *testing.T) {
		stub := &geographyHandlerStub{
			lgaGetFn: func(context.Context, string) (models.LocalGovernmentUnit, error) {
				return models.LocalGovernmentUnit{}, services.ErrInvalidLocalGovernmentUnitID
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas/lagos-ikeja", nil)
		req.SetPathValue("lga_id", "lagos-ikeja")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetLocalGovernmentUnit(w, r)
		}, rr)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
	})

	t.Run("unexpected service error", func(t *testing.T) {
		stub := &geographyHandlerStub{
			lgaGetFn: func(context.Context, string) (models.LocalGovernmentUnit, error) {
				return models.LocalGovernmentUnit{}, errors.New("database down")
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas/lagos-ikeja", nil)
		req.SetPathValue("lga_id", "lagos-ikeja")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetLocalGovernmentUnit(w, r)
		}, rr)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if strings.Contains(rr.Body.String(), "database down") {
			t.Fatalf("internal error details leaked: %s", rr.Body.String())
		}
	})
}

func TestGeographyHandlerListTimeZones(t *testing.T) {
	t.Run("list all", func(t *testing.T) {
		stub := &geographyHandlerStub{
			timeZoneListFn: func(ctx context.Context, input services.TimeZoneListInput) ([]models.TimeZone, error) {
				if input != (services.TimeZoneListInput{}) {
					t.Fatalf("unexpected input: %#v", input)
				}
				return []models.TimeZone{{ID: "Africa/Lagos", CountryAreaIDs: []string{"ng"}}}, nil
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/time-zones", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListTimeZones(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.timeZoneListCalls != 1 || stub.timeZoneGetCalls != 0 {
			t.Fatalf("unexpected call counts: list=%d get=%d", stub.timeZoneListCalls, stub.timeZoneGetCalls)
		}
		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		data := body["data"].([]any)
		if len(data) != 1 {
			t.Fatalf("unexpected payload: %#v", data)
		}
		item := data[0].(map[string]any)
		if item["id"] != "Africa/Lagos" {
			t.Fatalf("unexpected item: %#v", item)
		}
	})

	t.Run("filter and zero results", func(t *testing.T) {
		stub := &geographyHandlerStub{
			timeZoneListFn: func(ctx context.Context, input services.TimeZoneListInput) ([]models.TimeZone, error) {
				return []models.TimeZone{{ID: "Africa/Lagos", CountryAreaIDs: []string{"ng"}}}, nil
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/time-zones?country_area_id=ng", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListTimeZones(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.timeZoneListCalls != 1 || stub.lastTimeZoneQuery.CountryAreaID != "ng" {
			t.Fatalf("unexpected list call state: calls=%d input=%#v", stub.timeZoneListCalls, stub.lastTimeZoneQuery)
		}

		stub.timeZoneListCalls = 0
		rr = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/v1/geography/time-zones?country_area_id=bv", nil)
		stub.timeZoneListFn = func(ctx context.Context, input services.TimeZoneListInput) ([]models.TimeZone, error) {
			if input.CountryAreaID != "bv" {
				t.Fatalf("unexpected input: %#v", input)
			}
			return nil, nil
		}
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListTimeZones(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		data := body["data"].([]any)
		if len(data) != 0 {
			t.Fatalf("expected empty array, got %#v", data)
		}
	})

	t.Run("reject invalid query", func(t *testing.T) {
		stub := &geographyHandlerStub{}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		for _, rawQuery := range []string{"country_area_id=", "country_area_id=ZZ", "country_area_id=ng&country_area_id=bv", "country_area_id=zz"} {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/geography/time-zones?"+rawQuery, nil)
			invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
				h.ListTimeZones(w, r)
			}, rr)
			if rr.Code != http.StatusUnprocessableEntity {
				t.Fatalf("unexpected status for %q: %d", rawQuery, rr.Code)
			}
			if stub.timeZoneListCalls != 0 {
				t.Fatalf("service should not be called for invalid query")
			}
		}
	})

	t.Run("method guard", func(t *testing.T) {
		h, err := NewGeographyHandler(&geographyHandlerStub{})
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/geography/time-zones", nil)
		h.ListTimeZones(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if got := rr.Header().Get("Allow"); got != http.MethodGet {
			t.Fatalf("unexpected allow header: %q", got)
		}
	})
}

func TestGeographyHandlerGetTimeZone(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		stub := &geographyHandlerStub{
			timeZoneGetFn: func(ctx context.Context, timeZoneID string) (models.TimeZone, error) {
				if _, ok := middlewares.RequestIDFromContext(ctx); !ok {
					t.Fatal("request context was not preserved")
				}
				if timeZoneID != "America/Argentina/Buenos_Aires" {
					t.Fatalf("unexpected time zone id: %q", timeZoneID)
				}
				return models.TimeZone{ID: "America/Argentina/Buenos_Aires", CountryAreaIDs: []string{"ar"}}, nil
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/time-zones/America/Argentina/Buenos_Aires", nil)
		req.SetPathValue("time_zone_id", "America/Argentina/Buenos_Aires")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetTimeZone(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.timeZoneGetCalls != 1 {
			t.Fatalf("unexpected get call count: %d", stub.timeZoneGetCalls)
		}
		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		data := body["data"].(map[string]any)
		if data["id"] != "America/Argentina/Buenos_Aires" {
			t.Fatalf("unexpected payload: %#v", data)
		}
	})

	t.Run("reject invalid id", func(t *testing.T) {
		stub := &geographyHandlerStub{}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		for _, value := range []string{"", " ", "Africa", "/Africa/Lagos", "Africa/Lagos/", "Africa//Lagos", "Africa/./Lagos", "Africa/../Lagos", "Africa\\Lagos", "Africa%2FLagos", "Africa/Lagos?x=1", "Factory", "Etc/UTC"} {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/geography/time-zones/invalid", nil)
			req.SetPathValue("time_zone_id", value)
			invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
				h.GetTimeZone(w, r)
			}, rr)
			if rr.Code != http.StatusUnprocessableEntity {
				t.Fatalf("unexpected status for %q: %d", value, rr.Code)
			}
			if stub.timeZoneGetCalls != 0 {
				t.Fatalf("service should not be called for invalid id")
			}
		}
	})

	t.Run("not found and unexpected errors", func(t *testing.T) {
		stub := &geographyHandlerStub{
			timeZoneGetFn: func(context.Context, string) (models.TimeZone, error) {
				return models.TimeZone{}, services.ErrTimeZoneNotFound
			},
		}
		h, err := NewGeographyHandler(stub)
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/time-zones/Africa/Lagos", nil)
		req.SetPathValue("time_zone_id", "Africa/Lagos")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetTimeZone(w, r)
		}, rr)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}

		stub.timeZoneGetFn = func(context.Context, string) (models.TimeZone, error) {
			return models.TimeZone{}, errors.New("explode")
		}
		rr = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/v1/geography/time-zones/Africa/Lagos", nil)
		req.SetPathValue("time_zone_id", "Africa/Lagos")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetTimeZone(w, r)
		}, rr)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if strings.Contains(rr.Body.String(), "explode") {
			t.Fatalf("internal error details leaked: %s", rr.Body.String())
		}
	})

	t.Run("method guard", func(t *testing.T) {
		h, err := NewGeographyHandler(&geographyHandlerStub{})
		if err != nil {
			t.Fatalf("NewGeographyHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/geography/time-zones/Africa/Lagos", nil)
		req.SetPathValue("time_zone_id", "Africa/Lagos")
		h.GetTimeZone(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if got := rr.Header().Get("Allow"); got != http.MethodGet {
			t.Fatalf("unexpected allow header: %q", got)
		}
	})
}

func invokeWithRequestID(t *testing.T, req *http.Request, fn func(http.ResponseWriter, *http.Request), rr *httptest.ResponseRecorder) {
	t.Helper()

	req.Header.Set("X-Request-ID", "req_test")
	middlewares.RequestID(http.HandlerFunc(fn)).ServeHTTP(rr, req)
}

func fmtWrappedErr(err error) error {
	return fmt.Errorf("wrapped: %w", err)
}
