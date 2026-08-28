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
	listFn     func(context.Context) ([]models.State, error)
	getFn      func(context.Context, string) (models.State, error)
	zoneListFn func(context.Context) ([]models.GeopoliticalZone, error)
	zoneGetFn  func(context.Context, string) (models.GeopoliticalZone, error)

	listCalls     int
	getCalls      int
	zoneListCalls int
	zoneGetCalls  int
	lastID        string
	lastZoneID    string
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
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/states/missing", nil)
		req.SetPathValue("state_id", "missing")
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
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/states/missing", nil)
		req.SetPathValue("state_id", "missing")
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

func invokeWithRequestID(t *testing.T, req *http.Request, fn func(http.ResponseWriter, *http.Request), rr *httptest.ResponseRecorder) {
	t.Helper()

	req.Header.Set("X-Request-ID", "req_test")
	middlewares.RequestID(http.HandlerFunc(fn)).ServeHTTP(rr, req)
}

func fmtWrappedErr(err error) error {
	return fmt.Errorf("wrapped: %w", err)
}
