package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/handlers"
	"github.com/AbdulQuayyum/softdata-api/internal/middlewares"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type routerRecorder struct {
	mu    sync.Mutex
	calls []string
}

func (r *routerRecorder) add(value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, value)
}

func (r *routerRecorder) snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.calls))
	copy(out, r.calls)
	return out
}

func recordingMiddleware(rec *routerRecorder, name string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rec != nil {
				rec.add(name)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func recordingUsageFactory(rec *routerRecorder) UsageMiddlewareFactory {
	return func(endpoint, datasetGroup string) (MiddlewareFunc, error) {
		return recordingMiddleware(rec, "usage:"+endpoint+"|"+datasetGroup), nil
	}
}

func testMiddleware(rec *routerRecorder) Middleware {
	return Middleware{
		RequestID: func(next http.Handler) http.Handler {
			return middlewares.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if rec != nil {
					rec.add("request_id")
				}
				next.ServeHTTP(w, r)
			}))
		},
		Recovery:        recordingMiddleware(rec, "recovery"),
		Logger:          recordingMiddleware(rec, "logger"),
		SecurityHeaders: recordingMiddleware(rec, "security_headers"),
		CORS:            recordingMiddleware(rec, "cors"),
		BodyLimit:       recordingMiddleware(rec, "body_limit"),
		Timeout:         recordingMiddleware(rec, "timeout"),
		Authentication: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rec.add("authentication")
				ctx := middlewares.WithAccountIdentity(r.Context(), middlewares.AccountIdentity{AccountID: "acc_123"})
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		},
		OptionalAPIKey: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rec.add("optional_api_key")
				ctx := middlewares.WithAPIKeyIdentity(r.Context(), services.APIKeyIdentity{APIKeyID: "key_123", AccountID: "acc_123"})
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		},
		StandardLimit: recordingMiddleware(rec, "rate_limit"),
		UsageTracking: recordingUsageFactory(rec),
	}
}

func testHandlers(t *testing.T, rec *routerRecorder) Handlers {
	t.Helper()

	auth := &routerAuthStub{rec: rec}
	account := &routerAccountStub{rec: rec}
	apiKey := &routerAPIKeyStub{rec: rec}
	usage := &routerUsageStub{rec: rec}
	dataset := &routerDatasetStub{rec: rec}
	geography := &routerGeographyStub{rec: rec}
	education := &routerEducationStub{rec: rec}
	finance := &routerFinanceStub{rec: rec}

	authHandler, err := handlers.NewAuthHandler(auth, auth)
	if err != nil {
		t.Fatalf("NewAuthHandler() error = %v", err)
	}
	accountHandler, err := handlers.NewAccountHandler(account)
	if err != nil {
		t.Fatalf("NewAccountHandler() error = %v", err)
	}
	apiKeyHandler, err := handlers.NewAPIKeyHandler(apiKey)
	if err != nil {
		t.Fatalf("NewAPIKeyHandler() error = %v", err)
	}
	usageHandler, err := handlers.NewUsageHandler(usage)
	if err != nil {
		t.Fatalf("NewUsageHandler() error = %v", err)
	}
	datasetHandler, err := handlers.NewDatasetHandler(dataset)
	if err != nil {
		t.Fatalf("NewDatasetHandler() error = %v", err)
	}
	geographyHandler, err := handlers.NewGeographyHandler(geography, geography)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}
	educationHandler, err := handlers.NewEducationHandler(education)
	if err != nil {
		t.Fatalf("NewEducationHandler() error = %v", err)
	}
	financeHandler, err := handlers.NewFinanceHandler(finance)
	if err != nil {
		t.Fatalf("NewFinanceHandler() error = %v", err)
	}

	return Handlers{
		Health:    handlers.NewHealthHandler(),
		Discovery: handlers.NewDiscoveryHandler(),
		Geography: geographyHandler,
		Education: educationHandler,
		Finance:   financeHandler,
		Auth:      authHandler,
		Account:   accountHandler,
		APIKey:    apiKeyHandler,
		Usage:     usageHandler,
		Dataset:   datasetHandler,
	}
}

func testHandlersWithGeography(t *testing.T, rec *routerRecorder, geography *routerGeographyStub) Handlers {
	t.Helper()

	auth := &routerAuthStub{rec: rec}
	account := &routerAccountStub{rec: rec}
	apiKey := &routerAPIKeyStub{rec: rec}
	usage := &routerUsageStub{rec: rec}
	dataset := &routerDatasetStub{rec: rec}
	education := &routerEducationStub{rec: rec}
	finance := &routerFinanceStub{rec: rec}

	authHandler, err := handlers.NewAuthHandler(auth, auth)
	if err != nil {
		t.Fatalf("NewAuthHandler() error = %v", err)
	}
	accountHandler, err := handlers.NewAccountHandler(account)
	if err != nil {
		t.Fatalf("NewAccountHandler() error = %v", err)
	}
	apiKeyHandler, err := handlers.NewAPIKeyHandler(apiKey)
	if err != nil {
		t.Fatalf("NewAPIKeyHandler() error = %v", err)
	}
	usageHandler, err := handlers.NewUsageHandler(usage)
	if err != nil {
		t.Fatalf("NewUsageHandler() error = %v", err)
	}
	datasetHandler, err := handlers.NewDatasetHandler(dataset)
	if err != nil {
		t.Fatalf("NewDatasetHandler() error = %v", err)
	}
	geographyHandler, err := handlers.NewGeographyHandler(geography, geography)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}
	educationHandler, err := handlers.NewEducationHandler(education)
	if err != nil {
		t.Fatalf("NewEducationHandler() error = %v", err)
	}
	financeHandler, err := handlers.NewFinanceHandler(finance)
	if err != nil {
		t.Fatalf("NewFinanceHandler() error = %v", err)
	}

	return Handlers{
		Health:    handlers.NewHealthHandler(),
		Discovery: handlers.NewDiscoveryHandler(),
		Geography: geographyHandler,
		Education: educationHandler,
		Finance:   financeHandler,
		Auth:      authHandler,
		Account:   accountHandler,
		APIKey:    apiKeyHandler,
		Usage:     usageHandler,
		Dataset:   datasetHandler,
	}
}

func newTestRouter(t *testing.T, rec *routerRecorder) http.Handler {
	t.Helper()

	router, err := New(testHandlers(t, rec), testMiddleware(rec))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return router
}

func newPreflightRouter(t *testing.T, rec *routerRecorder) http.Handler {
	t.Helper()

	cors, err := middlewares.NewCORS(middlewares.CORSOptions{
		AllowedOrigins: []string{"https://example.com"},
	})
	if err != nil {
		t.Fatalf("NewCORS() error = %v", err)
	}

	router, err := New(testHandlers(t, rec), Middleware{
		RequestID:       middlewares.RequestID,
		Recovery:        func(next http.Handler) http.Handler { return next },
		Logger:          func(next http.Handler) http.Handler { return next },
		SecurityHeaders: func(next http.Handler) http.Handler { return next },
		CORS:            cors,
		BodyLimit:       func(next http.Handler) http.Handler { return next },
		Timeout:         func(next http.Handler) http.Handler { return next },
		Authentication:  func(next http.Handler) http.Handler { return next },
		OptionalAPIKey:  func(next http.Handler) http.Handler { return next },
		StandardLimit:   func(next http.Handler) http.Handler { return next },
		UsageTracking: func(endpoint, datasetGroup string) (MiddlewareFunc, error) {
			return func(next http.Handler) http.Handler { return next }, nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return router
}

func TestNewRejectsMissingDependencies(t *testing.T) {
	if _, err := New(Handlers{}, Middleware{}); err == nil {
		t.Fatal("New() error = nil, want error")
	}
}

func TestNewRejectsMissingGeographyHandler(t *testing.T) {
	rec := &routerRecorder{}
	handlers := testHandlers(t, rec)
	handlers.Geography = nil

	if _, err := New(handlers, testMiddleware(rec)); err == nil || !strings.Contains(err.Error(), "geography handler") {
		t.Fatalf("New() error = %v, want geography handler error", err)
	}
}

func TestNewRejectsMissingEducationHandler(t *testing.T) {
	rec := &routerRecorder{}
	handlers := testHandlers(t, rec)
	handlers.Education = nil

	if _, err := New(handlers, testMiddleware(rec)); err == nil || !strings.Contains(err.Error(), "education handler") {
		t.Fatalf("New() error = %v, want education handler error", err)
	}
}

func TestRouterAppliesGlobalAndDatasetRouteMiddlewareOrder(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	values := url.Values{}
	values.Set("search", " states ")
	values.Set("page", "2")
	values.Set("limit", "5")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/datasets?"+values.Encode(), nil)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}

	got := rec.snapshot()
	want := []string{
		"request_id",
		"recovery",
		"logger",
		"security_headers",
		"cors",
		"body_limit",
		"timeout",
		"optional_api_key",
		"rate_limit",
		"usage:/v1/datasets|",
		"dataset.list:states:2:5",
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("unexpected middleware order:\n got: %v\nwant: %v", got, want)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	data := body["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("unexpected dataset list payload: %#v", data)
	}
}

func TestRouterRegistersGeographyRoutes(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	t.Run("list", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/states", nil)
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		got := strings.Join(rec.snapshot(), ",")
		for _, want := range []string{
			"optional_api_key",
			"rate_limit",
			"usage:/v1/geography/states|geography",
			"geography.list",
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("expected %q in middleware sequence: %v", want, rec.snapshot())
			}
		}
	})

	t.Run("detail", func(t *testing.T) {
		rec = &routerRecorder{}
		router := newTestRouter(t, rec)

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/states/abia", nil)
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		got := strings.Join(rec.snapshot(), ",")
		for _, want := range []string{
			"optional_api_key",
			"rate_limit",
			"usage:/v1/geography/states/{state_id}|geography",
			"geography.get:abia",
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("expected %q in middleware sequence: %v", want, rec.snapshot())
			}
		}
	})

	t.Run("fct", func(t *testing.T) {
		rec := &routerRecorder{}
		router := newTestRouter(t, rec)

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/states/fct", nil)
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		got := strings.Join(rec.snapshot(), ",")
		if !strings.Contains(got, "geography.get:fct") {
			t.Fatalf("expected FCT lookup in middleware sequence: %v", rec.snapshot())
		}
	})

	t.Run("countries", func(t *testing.T) {
		t.Run("list", func(t *testing.T) {
			rec := &routerRecorder{}
			router := newTestRouter(t, rec)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries?region_code=002&subregion_code=015", nil)
			router.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			got := strings.Join(rec.snapshot(), ",")
			for _, want := range []string{
				"optional_api_key",
				"rate_limit",
				"usage:/v1/geography/countries|geography",
				"geography.country.list",
			} {
				if !strings.Contains(got, want) {
					t.Fatalf("expected %q in middleware sequence: %v", want, rec.snapshot())
				}
			}
		})

		t.Run("detail", func(t *testing.T) {
			rec := &routerRecorder{}
			router := newTestRouter(t, rec)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries/ng", nil)
			router.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			got := strings.Join(rec.snapshot(), ",")
			for _, want := range []string{
				"optional_api_key",
				"rate_limit",
				"usage:/v1/geography/countries/{country_id}|geography",
				"geography.country.get:ng",
			} {
				if !strings.Contains(got, want) {
					t.Fatalf("expected %q in middleware sequence: %v", want, rec.snapshot())
				}
			}
		})
	})

	t.Run("country flags", func(t *testing.T) {
		rec := &routerRecorder{}
		router := newTestRouter(t, rec)

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/assets/flags/ng.svg", nil)
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		got := strings.Join(rec.snapshot(), ",")
		for _, want := range []string{
			"optional_api_key",
			"rate_limit",
			"usage:/v1/assets/flags/{country_id}.svg|geography",
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("expected %q in middleware sequence: %v", want, rec.snapshot())
			}
		}
		if got := rr.Header().Get("Content-Type"); got != "image/svg+xml" {
			t.Fatalf("unexpected content type: %q", got)
		}
	})

	t.Run("geopolitical zones", func(t *testing.T) {
		t.Run("list", func(t *testing.T) {
			rec := &routerRecorder{}
			router := newTestRouter(t, rec)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones", nil)
			router.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			got := strings.Join(rec.snapshot(), ",")
			for _, want := range []string{
				"optional_api_key",
				"rate_limit",
				"usage:/v1/geography/geopolitical-zones|geography",
				"geography.zone.list",
			} {
				if !strings.Contains(got, want) {
					t.Fatalf("expected %q in middleware sequence: %v", want, rec.snapshot())
				}
			}
		})

		t.Run("detail", func(t *testing.T) {
			cases := []string{
				"north-central",
				"north-east",
				"north-west",
				"south-east",
				"south-south",
				"south-west",
			}
			for _, zoneID := range cases {
				t.Run(zoneID, func(t *testing.T) {
					rec := &routerRecorder{}
					router := newTestRouter(t, rec)

					rr := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones/"+zoneID, nil)
					router.ServeHTTP(rr, req)
					if rr.Code != http.StatusOK {
						t.Fatalf("unexpected status: %d", rr.Code)
					}
					got := strings.Join(rec.snapshot(), ",")
					for _, want := range []string{
						"optional_api_key",
						"rate_limit",
						"usage:/v1/geography/geopolitical-zones/{zone_id}|geography",
						"geography.zone.get:" + zoneID,
					} {
						if !strings.Contains(got, want) {
							t.Fatalf("expected %q in middleware sequence: %v", want, rec.snapshot())
						}
					}
				})
			}
		})

		t.Run("list not captured by detail", func(t *testing.T) {
			rec := &routerRecorder{}
			router := newTestRouter(t, rec)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones", nil)
			router.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			got := strings.Join(rec.snapshot(), ",")
			if strings.Contains(got, "geography.zone.get:") {
				t.Fatalf("list route must not hit detail handler: %v", rec.snapshot())
			}
		})
	})

	t.Run("allow and path value", func(t *testing.T) {
		rec := &routerRecorder{}
		router := newTestRouter(t, rec)

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodHead, "/v1/geography/states/abia", nil)
		req.Header.Set("X-Request-ID", "req_geo_head")
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if got := rr.Header().Get("Allow"); got != http.MethodGet {
			t.Fatalf("unexpected allow header: %q", got)
		}
	})
}

func TestRouterRegistersTimeZoneRoutes(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		rec := &routerRecorder{}
		geography := &routerGeographyStub{rec: rec}
		router, err := New(testHandlersWithGeography(t, rec, geography), testMiddleware(rec))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/time-zones?country_area_id=ng", nil)
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if geography.timeZoneListCalls != 1 {
			t.Fatalf("unexpected time zone list calls: %d", geography.timeZoneListCalls)
		}
		if geography.lastTimeZoneInput.CountryAreaID != "ng" {
			t.Fatalf("unexpected time zone filter: %#v", geography.lastTimeZoneInput)
		}
		got := strings.Join(rec.snapshot(), ",")
		for _, want := range []string{
			"optional_api_key",
			"rate_limit",
			"usage:/v1/geography/time-zones|geography",
			"geography.time-zone.list",
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("expected %q in middleware sequence: %v", want, rec.snapshot())
			}
		}
	})

	t.Run("detail", func(t *testing.T) {
		rec := &routerRecorder{}
		geography := &routerGeographyStub{rec: rec}
		router, err := New(testHandlersWithGeography(t, rec, geography), testMiddleware(rec))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/time-zones/America/Argentina/Buenos_Aires", nil)
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if geography.timeZoneGetCalls != 1 {
			t.Fatalf("unexpected time zone get calls: %d", geography.timeZoneGetCalls)
		}
		if geography.lastTimeZoneID != "America/Argentina/Buenos_Aires" {
			t.Fatalf("unexpected time zone id: %q", geography.lastTimeZoneID)
		}
		got := strings.Join(rec.snapshot(), ",")
		for _, want := range []string{
			"optional_api_key",
			"rate_limit",
			"usage:/v1/geography/time-zones/{time_zone_id...}|geography",
			"geography.time-zone.get:America/Argentina/Buenos_Aires",
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("expected %q in middleware sequence: %v", want, rec.snapshot())
			}
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		rec := &routerRecorder{}
		geography := &routerGeographyStub{rec: rec}
		router, err := New(testHandlersWithGeography(t, rec, geography), testMiddleware(rec))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/geography/time-zones/Africa/Lagos", nil)
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if got := rr.Header().Get("Allow"); got != http.MethodGet {
			t.Fatalf("unexpected allow header: %q", got)
		}
	})
}

func TestRouterRejectsUnknownRoutesAndUnsupportedMethods(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		target      string
		allow       string
		status      int
		wantCode    string
		wantRequest string
	}{
		{name: "unknown route", method: http.MethodGet, target: "/v1/unknown?token=abc", status: http.StatusNotFound, wantCode: "RESOURCE_NOT_FOUND", wantRequest: "req_404"},
		{name: "unknown nested route", method: http.MethodGet, target: "/v1/datasets/ng-states/unknown", status: http.StatusNotFound, wantCode: "RESOURCE_NOT_FOUND", wantRequest: "req_nested"},
		{name: "unknown geography nested route", method: http.MethodGet, target: "/v1/geography/states/abia/unknown", status: http.StatusNotFound, wantCode: "RESOURCE_NOT_FOUND", wantRequest: "req_geo_nested"},
		{name: "health wrong method", method: http.MethodPost, target: "/health", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_405"},
		{name: "auth wrong method", method: http.MethodGet, target: "/v1/auth/login", allow: http.MethodPost, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_auth"},
		{name: "account wrong method", method: http.MethodPost, target: "/v1/account", allow: "GET, PATCH, DELETE", status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_account"},
		{name: "dataset wrong method", method: http.MethodPost, target: "/v1/datasets/ng-states", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_dataset"},
		{name: "geography wrong method", method: http.MethodPost, target: "/v1/geography/states", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_geo"},
		{name: "geography detail wrong method", method: http.MethodDelete, target: "/v1/geography/states/abia", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_geo_detail"},
		{name: "geopolitical zones wrong method", method: http.MethodPost, target: "/v1/geography/geopolitical-zones", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_geo_zones"},
		{name: "geopolitical zone detail wrong method", method: http.MethodDelete, target: "/v1/geography/geopolitical-zones/north-central", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_geo_zone_detail"},
		{name: "geopolitical zone nested route", method: http.MethodGet, target: "/v1/geography/geopolitical-zones/north-central/extra", status: http.StatusNotFound, wantCode: "RESOURCE_NOT_FOUND", wantRequest: "req_geo_zone_nested"},
		{name: "lga wrong method", method: http.MethodPost, target: "/v1/geography/lgas", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_lga"},
		{name: "lga detail wrong method", method: http.MethodDelete, target: "/v1/geography/lgas/lagos-ikeja", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_lga_detail"},
		{name: "lga nested route", method: http.MethodGet, target: "/v1/geography/lgas/lagos-ikeja/extra", status: http.StatusNotFound, wantCode: "RESOURCE_NOT_FOUND", wantRequest: "req_lga_nested"},
		{name: "finance wrong method", method: http.MethodPost, target: "/v1/finance/payment-service-providers", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_finance"},
		{name: "finance detail wrong method", method: http.MethodDelete, target: "/v1/finance/payment-service-providers/super-agent-fairmoney", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_finance_detail"},
		{name: "finance nested route", method: http.MethodGet, target: "/v1/finance/payment-service-providers/super-agent-fairmoney/extra", status: http.StatusNotFound, wantCode: "RESOURCE_NOT_FOUND", wantRequest: "req_finance_nested"},
		{name: "imto wrong method", method: http.MethodPost, target: "/v1/finance/international-money-transfer-operators", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_imto"},
		{name: "imto detail wrong method", method: http.MethodDelete, target: "/v1/finance/international-money-transfer-operators/olive-monies-express-limited", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_imto_detail"},
		{name: "imto nested route", method: http.MethodGet, target: "/v1/finance/international-money-transfer-operators/olive-monies-express-limited/extra", status: http.StatusNotFound, wantCode: "RESOURCE_NOT_FOUND", wantRequest: "req_imto_nested"},
		{name: "currency wrong method", method: http.MethodPost, target: "/v1/finance/currencies", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_currency"},
		{name: "currency detail wrong method", method: http.MethodDelete, target: "/v1/finance/currencies/ngn", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_currency_detail"},
		{name: "currency nested route", method: http.MethodGet, target: "/v1/finance/currencies/ngn/extra", status: http.StatusNotFound, wantCode: "RESOURCE_NOT_FOUND", wantRequest: "req_currency_nested"},
		{name: "ready not registered", method: http.MethodGet, target: "/ready", status: http.StatusNotFound, wantCode: "RESOURCE_NOT_FOUND", wantRequest: "req_ready"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := &routerRecorder{}
			router := newTestRouter(t, rec)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.target, nil)
			req.Header.Set("X-Request-ID", tc.wantRequest)
			router.ServeHTTP(rr, req)

			if rr.Code != tc.status {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			if tc.target != "/v1/geography/geopolitical-zones/" {
				if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
					t.Fatalf("unexpected content type: %q", got)
				}
				if got := rr.Header().Get("X-Request-ID"); got != tc.wantRequest {
					t.Fatalf("unexpected request id header: %q", got)
				}
			}
			if tc.allow != "" {
				if got := rr.Header().Get("Allow"); got != tc.allow {
					t.Fatalf("unexpected allow header: %q", got)
				}
			}
			body := rr.Body.String()
			if tc.target == "/v1/geography/geopolitical-zones/" {
				if !strings.Contains(body, "404 page not found") {
					t.Fatalf("unexpected trailing-slash behavior: %s", body)
				}
				return
			}
			if strings.Contains(body, "404 page not found") || strings.Contains(body, "Method Not Allowed") {
				t.Fatalf("plain-text router error leaked: %s", body)
			}
			if !strings.Contains(body, `"success":false`) {
				t.Fatalf("expected error envelope: %s", body)
			}
			if !strings.Contains(body, `"code":"`+tc.wantCode+`"`) {
				t.Fatalf("unexpected error code: %s", body)
			}
			if !strings.Contains(body, `"request_id":"`+tc.wantRequest+`"`) {
				t.Fatalf("unexpected request id in body: %s", body)
			}
			if strings.Contains(body, "token=abc") {
				t.Fatalf("raw query leaked into error body: %s", body)
			}
		})
	}
}

func TestRouterPreservesTrailingSlashPolicyForGeopoliticalZones(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones/", nil)
	req.Header.Set("X-Request-ID", "req_geo_zones_slash")
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("unexpected content type: %q", got)
	}
	if !strings.Contains(rr.Body.String(), "404 page not found") {
		t.Fatalf("unexpected trailing-slash response: %s", rr.Body.String())
	}
}

func TestRouterPreservesTrailingSlashPolicyForLGAs(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas/", nil)
	req.Header.Set("X-Request-ID", "req_lga_slash")
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("unexpected content type: %q", got)
	}
	if !strings.Contains(rr.Body.String(), "404 page not found") {
		t.Fatalf("unexpected trailing-slash response: %s", rr.Body.String())
	}
}

func TestRouterRejectsHeadRequestsWithJson405(t *testing.T) {
	tests := []struct {
		name   string
		target string
		allow  string
	}{
		{name: "health", target: "/health", allow: http.MethodGet},
		{name: "account", target: "/v1/account", allow: "GET, PATCH, DELETE"},
		{name: "geography", target: "/v1/geography/states", allow: http.MethodGet},
		{name: "geopolitical zones", target: "/v1/geography/geopolitical-zones", allow: http.MethodGet},
		{name: "lgas", target: "/v1/geography/lgas", allow: http.MethodGet},
		{name: "lga detail", target: "/v1/geography/lgas/lagos-ikeja", allow: http.MethodGet},
		{name: "finance", target: "/v1/finance/payment-service-providers", allow: http.MethodGet},
		{name: "finance detail", target: "/v1/finance/payment-service-providers/super-agent-fairmoney", allow: http.MethodGet},
		{name: "imto", target: "/v1/finance/international-money-transfer-operators", allow: http.MethodGet},
		{name: "imto detail", target: "/v1/finance/international-money-transfer-operators/olive-monies-express-limited", allow: http.MethodGet},
		{name: "currency", target: "/v1/finance/currencies", allow: http.MethodGet},
		{name: "currency detail", target: "/v1/finance/currencies/ngn", allow: http.MethodGet},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := &routerRecorder{}
			router := newTestRouter(t, rec)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodHead, tc.target, nil)
			req.Header.Set("X-Request-ID", "req_head")
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusMethodNotAllowed {
				t.Fatalf("unexpected status: %d", rr.Code)
			}
			if got := rr.Header().Get("Allow"); got != tc.allow {
				t.Fatalf("unexpected allow header: %q", got)
			}
			body := rr.Body.String()
			if strings.Contains(body, "404 page not found") || strings.Contains(body, "Method Not Allowed") {
				t.Fatalf("plain-text head rejection leaked: %s", body)
			}
			if !strings.Contains(body, `"request_id":"req_head"`) {
				t.Fatalf("expected request id in body: %s", body)
			}
		})
	}
}

func TestRouterHandlesOptionsPreflightBeforeMethodRejection(t *testing.T) {
	rec := &routerRecorder{}
	router := newPreflightRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/v1/datasets", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("unexpected allow origin: %q", got)
	}
	got := strings.Join(rec.snapshot(), ",")
	if strings.Contains(got, "body_limit") || strings.Contains(got, "timeout") || strings.Contains(got, "dataset.list") {
		t.Fatalf("preflight should have short-circuited before downstream handlers: %v", rec.snapshot())
	}
}

func TestRouterExcludesRateLimitedDatasetRequestsFromUsageTracking(t *testing.T) {
	rec := &routerRecorder{}
	router, err := New(testHandlers(t, rec), Middleware{
		RequestID:       recordingMiddleware(rec, "request_id"),
		Recovery:        recordingMiddleware(rec, "recovery"),
		Logger:          recordingMiddleware(rec, "logger"),
		SecurityHeaders: recordingMiddleware(rec, "security_headers"),
		CORS:            recordingMiddleware(rec, "cors"),
		BodyLimit:       recordingMiddleware(rec, "body_limit"),
		Timeout:         recordingMiddleware(rec, "timeout"),
		Authentication:  testMiddleware(rec).Authentication,
		OptionalAPIKey:  testMiddleware(rec).OptionalAPIKey,
		StandardLimit: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rec.add("rate_limit")
				w.WriteHeader(http.StatusTooManyRequests)
			})
		},
		UsageTracking: recordingUsageFactory(rec),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/datasets", nil)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	got := strings.Join(rec.snapshot(), ",")
	if strings.Contains(got, "usage:/v1/datasets|") {
		t.Fatalf("usage tracking should not record rejected dataset request: %v", rec.snapshot())
	}
	if strings.Contains(got, "dataset.list:") {
		t.Fatalf("dataset handler should not run on rate-limited request: %v", rec.snapshot())
	}
}

func TestRouterUsesGeographyMiddlewarePolicy(t *testing.T) {
	t.Parallel()

	harness := newGeographyPolicyRouter(t)

	t.Run("anonymous", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/states", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("User-Agent", "TestAgent/1.0")
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.rateLimit.calls == 0 {
			t.Fatal("expected rate limit repository to be used")
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAnonymous {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.rateLimit.request.Limit != 60 {
			t.Fatalf("unexpected anonymous limit: %d", harness.rateLimit.request.Limit)
		}
		if harness.usage.calls == 0 || harness.usage.input.Route != "/v1/geography/states" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "geography" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.geography.listCalls != 1 {
			t.Fatalf("unexpected geography list calls: %d", harness.geography.listCalls)
		}
	})

	t.Run("api key", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/states/abia", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("X-API-Key", "sd_live_example")
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.auth.calls == 0 {
			t.Fatal("expected API key authenticator to be used")
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAPIKey {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.rateLimit.request.Limit != 300 {
			t.Fatalf("unexpected api-key limit: %d", harness.rateLimit.request.Limit)
		}
		if harness.rateLimit.request.Subject != "key_123" {
			t.Fatalf("unexpected subject: %q", harness.rateLimit.request.Subject)
		}
		if harness.usage.input.Route != "/v1/geography/states/{state_id}" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "geography" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.geography.lastStateID != "abia" {
			t.Fatalf("unexpected state id seen by handler: %q", harness.geography.lastStateID)
		}
		if !harness.geography.lastHadAPIKey || harness.geography.lastAPIKeyIdentity.APIKeyID != "key_123" || harness.geography.lastAPIKeyIdentity.AccountID != "acc_123" {
			t.Fatalf("unexpected api key identity seen by handler: %#v", harness.geography.lastAPIKeyIdentity)
		}
	})

	t.Run("zone anonymous", func(t *testing.T) {
		harness = newGeographyPolicyRouter(t)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("User-Agent", "TestAgent/1.0")
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAnonymous {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.rateLimit.request.Limit != 60 {
			t.Fatalf("unexpected anonymous limit: %d", harness.rateLimit.request.Limit)
		}
		if harness.usage.input.Route != "/v1/geography/geopolitical-zones" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "geography" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.geography.zoneListCalls != 1 {
			t.Fatalf("unexpected zone list calls: %d", harness.geography.zoneListCalls)
		}
	})

	t.Run("zone api key", func(t *testing.T) {
		harness = newGeographyPolicyRouter(t)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones/north-central", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("X-API-Key", "sd_live_example")
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAPIKey {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.rateLimit.request.Limit != 300 {
			t.Fatalf("unexpected api-key limit: %d", harness.rateLimit.request.Limit)
		}
		if harness.rateLimit.request.Subject != "key_123" {
			t.Fatalf("unexpected subject: %q", harness.rateLimit.request.Subject)
		}
		if harness.usage.input.Route != "/v1/geography/geopolitical-zones/{zone_id}" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "geography" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.geography.zoneGetCalls != 1 {
			t.Fatalf("unexpected zone get calls: %d", harness.geography.zoneGetCalls)
		}
		if harness.geography.lastZoneID != "north-central" {
			t.Fatalf("unexpected zone id seen by handler: %q", harness.geography.lastZoneID)
		}
		if !harness.geography.lastHadAPIKey || harness.geography.lastAPIKeyIdentity.APIKeyID != "key_123" || harness.geography.lastAPIKeyIdentity.AccountID != "acc_123" {
			t.Fatalf("unexpected api key identity seen by handler: %#v", harness.geography.lastAPIKeyIdentity)
		}
	})

	t.Run("lga anonymous", func(t *testing.T) {
		harness = newGeographyPolicyRouter(t)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("User-Agent", "TestAgent/1.0")
		req.URL.RawQuery = "state_id=fct"
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAnonymous {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.rateLimit.request.Limit != 60 {
			t.Fatalf("unexpected anonymous limit: %d", harness.rateLimit.request.Limit)
		}
		if harness.usage.input.Route != "/v1/geography/lgas" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "geography" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.geography.lgaListByCalls != 1 {
			t.Fatalf("unexpected lga list-by calls: %d", harness.geography.lgaListByCalls)
		}
		if harness.geography.lastLGAStateID != "fct" {
			t.Fatalf("unexpected state id seen by handler: %q", harness.geography.lastLGAStateID)
		}
	})

	t.Run("lga api key", func(t *testing.T) {
		harness = newGeographyPolicyRouter(t)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas/lagos-ikeja", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("X-API-Key", "sd_live_example")
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAPIKey {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.rateLimit.request.Limit != 300 {
			t.Fatalf("unexpected api-key limit: %d", harness.rateLimit.request.Limit)
		}
		if harness.rateLimit.request.Subject != "key_123" {
			t.Fatalf("unexpected subject: %q", harness.rateLimit.request.Subject)
		}
		if harness.usage.input.Route != "/v1/geography/lgas/{lga_id}" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "geography" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.geography.lgaGetCalls != 1 {
			t.Fatalf("unexpected lga get calls: %d", harness.geography.lgaGetCalls)
		}
		if harness.geography.lastLGAID != "lagos-ikeja" {
			t.Fatalf("unexpected lga id seen by handler: %q", harness.geography.lastLGAID)
		}
		if !harness.geography.lastHadAPIKey || harness.geography.lastAPIKeyIdentity.APIKeyID != "key_123" || harness.geography.lastAPIKeyIdentity.AccountID != "acc_123" {
			t.Fatalf("unexpected api key identity seen by handler: %#v", harness.geography.lastAPIKeyIdentity)
		}
	})

	t.Run("lga empty state query reaches handler", func(t *testing.T) {
		harness = newGeographyPolicyRouter(t)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("User-Agent", "TestAgent/1.0")
		req.URL.RawQuery = "state_id="
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.geography.lgaListCalls != 0 || harness.geography.lgaListByCalls != 0 {
			t.Fatalf("handler validation should fail before service calls: %#v", harness.geography)
		}
		if harness.usage.calls != 1 {
			t.Fatalf("usage should record the rejected handler response once: %d", harness.usage.calls)
		}
		if harness.usage.input.Route != "/v1/geography/lgas" {
			t.Fatalf("unexpected usage route: %q", harness.usage.input.Route)
		}
	})

	t.Run("rate limited", func(t *testing.T) {
		harness := newGeographyPolicyRouter(t)
		harness.rateLimit.result = interfaces.RateLimitResult{
			Allowed:   false,
			Limit:     60,
			Remaining: 0,
			ResetAt:   time.Now().UTC().Add(time.Minute),
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/states", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusTooManyRequests {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.geography.listCalls != 0 {
			t.Fatalf("handler should not run on rate-limited request: %d", harness.geography.listCalls)
		}
		if harness.usage.calls != 0 {
			t.Fatalf("usage should not run on rate-limited request: %d", harness.usage.calls)
		}
	})

	t.Run("zone rate limited", func(t *testing.T) {
		harness := newGeographyPolicyRouter(t)
		harness.rateLimit.result = interfaces.RateLimitResult{
			Allowed:   false,
			Limit:     60,
			Remaining: 0,
			ResetAt:   time.Now().UTC().Add(time.Minute),
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones/north-central", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusTooManyRequests {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.geography.zoneGetCalls != 0 {
			t.Fatalf("handler should not run on rate-limited request: %d", harness.geography.zoneGetCalls)
		}
		if harness.usage.calls != 0 {
			t.Fatalf("usage should not run on rate-limited request: %d", harness.usage.calls)
		}
	})

	t.Run("lga rate limited", func(t *testing.T) {
		harness := newGeographyPolicyRouter(t)
		harness.rateLimit.result = interfaces.RateLimitResult{
			Allowed:   false,
			Limit:     60,
			Remaining: 0,
			ResetAt:   time.Now().UTC().Add(time.Minute),
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas/lagos-ikeja", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusTooManyRequests {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.geography.lgaGetCalls != 0 {
			t.Fatalf("handler should not run on rate-limited request: %d", harness.geography.lgaGetCalls)
		}
		if harness.usage.calls != 0 {
			t.Fatalf("usage should not run on rate-limited request: %d", harness.usage.calls)
		}
	})

	t.Run("invalid key", func(t *testing.T) {
		harness.auth.err = services.ErrAPIKeyNotFound
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/geography/states", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("X-API-Key", "bad-key")
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.geography.listCalls != 0 {
			t.Fatalf("handler should not run on invalid api key, calls: %d", harness.geography.listCalls)
		}
	})

	if harness.rateLimit.request.SubjectKind == interfaces.RateLimitSubjectDownload {
		t.Fatal("geography routes must not use download rate limit policy")
	}
}

func TestRouterUsesFinanceMiddlewarePolicy(t *testing.T) {
	t.Parallel()

	harness := newFinancePolicyRouter(t)

	t.Run("anonymous", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("User-Agent", "TestAgent/1.0")
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAnonymous {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.rateLimit.request.Limit != 60 {
			t.Fatalf("unexpected anonymous limit: %d", harness.rateLimit.request.Limit)
		}
		if harness.usage.input.Route != "/v1/finance/payment-service-providers" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "finance" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.finance.listCalls != 1 {
			t.Fatalf("unexpected finance list calls: %d", harness.finance.listCalls)
		}
	})

	t.Run("api key", func(t *testing.T) {
		harness = newFinancePolicyRouter(t)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("X-API-Key", "sd_live_example")
		req.URL.RawQuery = "institution_type=super_agent"
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAPIKey {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.rateLimit.request.Limit != 300 {
			t.Fatalf("unexpected api-key limit: %d", harness.rateLimit.request.Limit)
		}
		if harness.rateLimit.request.Subject != "key_123" {
			t.Fatalf("unexpected subject: %q", harness.rateLimit.request.Subject)
		}
		if harness.usage.input.Route != "/v1/finance/payment-service-providers" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "finance" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.finance.lastInstitutionType != "super_agent" {
			t.Fatalf("unexpected institution type seen by handler: %q", harness.finance.lastInstitutionType)
		}
	})

	t.Run("detail", func(t *testing.T) {
		harness = newFinancePolicyRouter(t)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers/super-agent-fairmoney", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("X-API-Key", "sd_live_example")
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAPIKey {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.rateLimit.request.Limit != 300 {
			t.Fatalf("unexpected api-key limit: %d", harness.rateLimit.request.Limit)
		}
		if harness.rateLimit.request.Subject != "key_123" {
			t.Fatalf("unexpected subject: %q", harness.rateLimit.request.Subject)
		}
		if harness.usage.input.Route != "/v1/finance/payment-service-providers/{provider_id}" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "finance" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.finance.lastProviderID != "super-agent-fairmoney" {
			t.Fatalf("unexpected provider id seen by handler: %q", harness.finance.lastProviderID)
		}
		if !harness.finance.lastHadAPIKey || harness.finance.lastAPIKeyIdentity.APIKeyID != "key_123" || harness.finance.lastAPIKeyIdentity.AccountID != "acc_123" {
			t.Fatalf("unexpected api key identity seen by handler: %#v", harness.finance.lastAPIKeyIdentity)
		}
	})

	t.Run("imto list", func(t *testing.T) {
		harness = newFinancePolicyRouter(t)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/international-money-transfer-operators", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAnonymous {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.usage.input.Route != "/v1/finance/international-money-transfer-operators" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "finance" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.finance.listIMTOCalls != 1 {
			t.Fatalf("unexpected imto list calls: %d", harness.finance.listIMTOCalls)
		}
	})

	t.Run("imto detail", func(t *testing.T) {
		harness = newFinancePolicyRouter(t)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/international-money-transfer-operators/olive-monies-express-limited", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("X-API-Key", "sd_live_example")
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAPIKey {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.rateLimit.request.Subject != "key_123" {
			t.Fatalf("unexpected subject: %q", harness.rateLimit.request.Subject)
		}
		if harness.usage.input.Route != "/v1/finance/international-money-transfer-operators/{operator_id}" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "finance" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.finance.lastOperatorID != "olive-monies-express-limited" {
			t.Fatalf("unexpected operator id seen by handler: %q", harness.finance.lastOperatorID)
		}
		if !harness.finance.lastHadAPIKey || harness.finance.lastAPIKeyIdentity.APIKeyID != "key_123" || harness.finance.lastAPIKeyIdentity.AccountID != "acc_123" {
			t.Fatalf("unexpected api key identity seen by handler: %#v", harness.finance.lastAPIKeyIdentity)
		}
	})

	t.Run("currency list", func(t *testing.T) {
		harness = newFinancePolicyRouter(t)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/currencies", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("X-API-Key", "sd_live_example")
		req.URL.RawQuery = "country_area_id=ng"
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAPIKey {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.rateLimit.request.Subject != "key_123" {
			t.Fatalf("unexpected subject: %q", harness.rateLimit.request.Subject)
		}
		if harness.usage.input.Route != "/v1/finance/currencies" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "finance" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.finance.lastCurrencyFilter != "ng" {
			t.Fatalf("unexpected currency filter seen by handler: %q", harness.finance.lastCurrencyFilter)
		}
	})

	t.Run("currency detail", func(t *testing.T) {
		harness = newFinancePolicyRouter(t)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/currencies/ngn", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("X-API-Key", "sd_live_example")
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAPIKey {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.usage.input.Route != "/v1/finance/currencies/{currency_id}" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "finance" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.finance.lastCurrencyID != "ngn" {
			t.Fatalf("unexpected currency id seen by handler: %q", harness.finance.lastCurrencyID)
		}
	})

	t.Run("invalid query reaches handler", func(t *testing.T) {
		harness = newFinancePolicyRouter(t)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.URL.RawQuery = "institution_type="
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.finance.listCalls != 0 || harness.finance.listByTypeCalls != 0 {
			t.Fatalf("service should not be called for invalid query: %#v", harness.finance)
		}
		if harness.usage.calls != 1 {
			t.Fatalf("usage should record the rejected handler response once: %d", harness.usage.calls)
		}
		if harness.usage.input.Route != "/v1/finance/payment-service-providers" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "finance" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
	})

	t.Run("rate limited", func(t *testing.T) {
		harness = newFinancePolicyRouter(t)
		harness.rateLimit.result = interfaces.RateLimitResult{
			Allowed:   false,
			Limit:     60,
			Remaining: 0,
			ResetAt:   time.Now().UTC().Add(time.Minute),
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusTooManyRequests {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.finance.listCalls != 0 {
			t.Fatalf("handler should not run on rate-limited request: %d", harness.finance.listCalls)
		}
		if harness.usage.calls != 0 {
			t.Fatalf("usage should not run on rate-limited request: %d", harness.usage.calls)
		}
	})

	t.Run("currency invalid query reaches handler", func(t *testing.T) {
		harness = newFinancePolicyRouter(t)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/currencies", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.URL.RawQuery = "country_area_id=zz"
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.finance.listCurrencyCalls != 0 {
			t.Fatalf("service should not be called for invalid currency query: %#v", harness.finance)
		}
		if harness.usage.calls != 1 {
			t.Fatalf("usage should record the rejected handler response once: %d", harness.usage.calls)
		}
		if harness.usage.input.Route != "/v1/finance/currencies" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "finance" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
	})

	t.Run("currency rate limited", func(t *testing.T) {
		harness = newFinancePolicyRouter(t)
		harness.rateLimit.result = interfaces.RateLimitResult{
			Allowed:   false,
			Limit:     60,
			Remaining: 0,
			ResetAt:   time.Now().UTC().Add(time.Minute),
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/currencies", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusTooManyRequests {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.finance.listCurrencyCalls != 0 {
			t.Fatalf("handler should not run on rate-limited request: %d", harness.finance.listCurrencyCalls)
		}
		if harness.usage.calls != 0 {
			t.Fatalf("usage should not run on rate-limited request: %d", harness.usage.calls)
		}
	})

	t.Run("imto rate limited", func(t *testing.T) {
		harness = newFinancePolicyRouter(t)
		harness.rateLimit.result = interfaces.RateLimitResult{
			Allowed:   false,
			Limit:     60,
			Remaining: 0,
			ResetAt:   time.Now().UTC().Add(time.Minute),
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/international-money-transfer-operators", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusTooManyRequests {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.finance.listIMTOCalls != 0 {
			t.Fatalf("handler should not run on rate-limited request: %d", harness.finance.listIMTOCalls)
		}
		if harness.usage.calls != 0 {
			t.Fatalf("usage should not run on rate-limited request: %d", harness.usage.calls)
		}
	})

	if harness.rateLimit.request.SubjectKind == interfaces.RateLimitSubjectDownload {
		t.Fatal("finance routes must not use download rate limit policy")
	}
}

func TestRouterUsesEducationMiddlewarePolicy(t *testing.T) {
	t.Parallel()

	harness := newEducationPolicyRouter(t)

	t.Run("anonymous list", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/colleges-of-education", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("User-Agent", "TestAgent/1.0")
		req.URL.RawQuery = "state_id=lagos&ownership_type=private"
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAnonymous {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.rateLimit.request.Limit != 60 {
			t.Fatalf("unexpected anonymous limit: %d", harness.rateLimit.request.Limit)
		}
		if harness.usage.input.Route != "/v1/education/colleges-of-education" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "education" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.education.collegeListCalls != 1 {
			t.Fatalf("unexpected college list calls: %d", harness.education.collegeListCalls)
		}
		if harness.education.lastCollegeInput.StateID != "lagos" || harness.education.lastCollegeInput.OwnershipType != "private" {
			t.Fatalf("unexpected college query: %#v", harness.education.lastCollegeInput)
		}
	})

	t.Run("api key detail", func(t *testing.T) {
		harness = newEducationPolicyRouter(t)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/colleges-of-education/federal-college-of-education-zaria", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("X-API-Key", "sd_live_example")
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.rateLimit.request.SubjectKind != interfaces.RateLimitSubjectAPIKey {
			t.Fatalf("unexpected subject kind: %#v", harness.rateLimit.request.SubjectKind)
		}
		if harness.rateLimit.request.Limit != 300 {
			t.Fatalf("unexpected api-key limit: %d", harness.rateLimit.request.Limit)
		}
		if harness.rateLimit.request.Subject != "key_123" {
			t.Fatalf("unexpected subject: %q", harness.rateLimit.request.Subject)
		}
		if harness.usage.input.Route != "/v1/education/colleges-of-education/{college_id}" || harness.usage.input.DatasetGroup == nil || *harness.usage.input.DatasetGroup != "education" {
			t.Fatalf("unexpected usage record: %#v", harness.usage.input)
		}
		if harness.education.lastCollegeID != "federal-college-of-education-zaria" {
			t.Fatalf("unexpected college id seen by handler: %q", harness.education.lastCollegeID)
		}
		if !harness.education.lastHadAPIKey || harness.education.lastAPIKeyIdentity.APIKeyID != "key_123" || harness.education.lastAPIKeyIdentity.AccountID != "acc_123" {
			t.Fatalf("unexpected api key identity seen by handler: %#v", harness.education.lastAPIKeyIdentity)
		}
	})

	t.Run("rate limited", func(t *testing.T) {
		harness = newEducationPolicyRouter(t)
		harness.rateLimit.result = interfaces.RateLimitResult{
			Allowed:   false,
			Limit:     60,
			Remaining: 0,
			ResetAt:   time.Now().UTC().Add(time.Minute),
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/education/colleges-of-education", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		harness.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusTooManyRequests {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if harness.education.collegeListCalls != 0 {
			t.Fatalf("handler should not run on rate-limited request: %d", harness.education.collegeListCalls)
		}
		if harness.usage.calls != 0 {
			t.Fatalf("usage should not run on rate-limited request: %d", harness.usage.calls)
		}
	})

	if harness.rateLimit.request.SubjectKind == interfaces.RateLimitSubjectDownload {
		t.Fatal("education routes must not use download rate limit policy")
	}
}

type geographyPolicyHarness struct {
	router    http.Handler
	auth      *routerAPIKeyAuthenticatorStub
	rateLimit *routerRateLimitRepoStub
	usage     *routerUsageRecorderStub
	geography *routerGeographyStub
	anonymous *routerAnonymousIdentifierStub
}

type financePolicyHarness struct {
	router    http.Handler
	auth      *routerAPIKeyAuthenticatorStub
	rateLimit *routerRateLimitRepoStub
	usage     *routerUsageRecorderStub
	finance   *routerFinanceStub
	anonymous *routerAnonymousIdentifierStub
}

type educationPolicyHarness struct {
	router    http.Handler
	auth      *routerAPIKeyAuthenticatorStub
	rateLimit *routerRateLimitRepoStub
	usage     *routerUsageRecorderStub
	education *routerEducationStub
	anonymous *routerAnonymousIdentifierStub
}

func newGeographyPolicyRouter(t *testing.T) geographyPolicyHarness {
	t.Helper()

	cors, err := middlewares.NewCORS(middlewares.CORSOptions{
		AllowedOrigins: []string{"https://example.com"},
	})
	if err != nil {
		t.Fatalf("NewCORS() error = %v", err)
	}

	auth := &routerAPIKeyAuthenticatorStub{identity: services.APIKeyIdentity{APIKeyID: "key_123", AccountID: "acc_123"}}
	anonymous := &routerAnonymousIdentifierStub{value: "anon-opaque"}
	rateLimit := &routerRateLimitRepoStub{}
	usage := &routerUsageRecorderStub{}
	geography := &routerGeographyStub{}
	geographyHandler, err := handlers.NewGeographyHandler(geography, geography)
	if err != nil {
		t.Fatalf("NewGeographyHandler() error = %v", err)
	}
	baseHandlers := testHandlers(t, &routerRecorder{})
	rateLimitMiddleware, err := middlewares.RateLimit(rateLimit, anonymous, middlewares.RateLimitPolicy{
		AnonymousLimit: 60,
		APIKeyLimit:    300,
		DownloadLimit:  10,
		Window:         time.Minute,
	})
	if err != nil {
		t.Fatalf("RateLimit() error = %v", err)
	}

	routerHandler, err := New(Handlers{
		Health:    handlers.NewHealthHandler(),
		Discovery: handlers.NewDiscoveryHandler(),
		Geography: geographyHandler,
		Education: baseHandlers.Education,
		Finance:   baseHandlers.Finance,
		Auth:      baseHandlers.Auth,
		Account:   baseHandlers.Account,
		APIKey:    baseHandlers.APIKey,
		Usage:     baseHandlers.Usage,
		Dataset:   baseHandlers.Dataset,
	}, Middleware{
		RequestID:       middlewares.RequestID,
		Recovery:        func(next http.Handler) http.Handler { return next },
		Logger:          func(next http.Handler) http.Handler { return next },
		SecurityHeaders: func(next http.Handler) http.Handler { return next },
		CORS:            cors,
		BodyLimit:       func(next http.Handler) http.Handler { return next },
		Timeout:         func(next http.Handler) http.Handler { return next },
		Authentication:  func(next http.Handler) http.Handler { return next },
		OptionalAPIKey:  middlewares.OptionalAPIKey(auth),
		StandardLimit:   rateLimitMiddleware,
		UsageTracking: func(endpoint, datasetGroup string) (MiddlewareFunc, error) {
			switch endpoint {
			case "/v1/geography/states":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			case "/v1/geography/states/{state_id}":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			case "/v1/geography/geopolitical-zones":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			case "/v1/geography/geopolitical-zones/{zone_id}":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			case "/v1/geography/lgas":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			case "/v1/geography/lgas/{lga_id}":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			default:
				return func(next http.Handler) http.Handler { return next }, nil
			}
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return geographyPolicyHarness{
		router:    routerHandler,
		auth:      auth,
		rateLimit: rateLimit,
		usage:     usage,
		geography: geography,
		anonymous: anonymous,
	}
}

func newFinancePolicyRouter(t *testing.T) financePolicyHarness {
	t.Helper()

	cors, err := middlewares.NewCORS(middlewares.CORSOptions{
		AllowedOrigins: []string{"https://example.com"},
	})
	if err != nil {
		t.Fatalf("NewCORS() error = %v", err)
	}

	auth := &routerAPIKeyAuthenticatorStub{identity: services.APIKeyIdentity{APIKeyID: "key_123", AccountID: "acc_123"}}
	anonymous := &routerAnonymousIdentifierStub{value: "anon-opaque"}
	rateLimit := &routerRateLimitRepoStub{}
	usage := &routerUsageRecorderStub{}
	finance := &routerFinanceStub{}
	financeHandler, err := handlers.NewFinanceHandler(finance)
	if err != nil {
		t.Fatalf("NewFinanceHandler() error = %v", err)
	}
	baseHandlers := testHandlers(t, &routerRecorder{})
	rateLimitMiddleware, err := middlewares.RateLimit(rateLimit, anonymous, middlewares.RateLimitPolicy{
		AnonymousLimit: 60,
		APIKeyLimit:    300,
		DownloadLimit:  10,
		Window:         time.Minute,
	})
	if err != nil {
		t.Fatalf("RateLimit() error = %v", err)
	}

	routerHandler, err := New(Handlers{
		Health:    handlers.NewHealthHandler(),
		Discovery: handlers.NewDiscoveryHandler(),
		Geography: baseHandlers.Geography,
		Education: baseHandlers.Education,
		Finance:   financeHandler,
		Auth:      baseHandlers.Auth,
		Account:   baseHandlers.Account,
		APIKey:    baseHandlers.APIKey,
		Usage:     baseHandlers.Usage,
		Dataset:   baseHandlers.Dataset,
	}, Middleware{
		RequestID:       middlewares.RequestID,
		Recovery:        func(next http.Handler) http.Handler { return next },
		Logger:          func(next http.Handler) http.Handler { return next },
		SecurityHeaders: func(next http.Handler) http.Handler { return next },
		CORS:            cors,
		BodyLimit:       func(next http.Handler) http.Handler { return next },
		Timeout:         func(next http.Handler) http.Handler { return next },
		Authentication:  func(next http.Handler) http.Handler { return next },
		OptionalAPIKey:  middlewares.OptionalAPIKey(auth),
		StandardLimit:   rateLimitMiddleware,
		UsageTracking: func(endpoint, datasetGroup string) (MiddlewareFunc, error) {
			switch endpoint {
			case "/v1/finance/payment-service-providers":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			case "/v1/finance/payment-service-providers/{provider_id}":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			case "/v1/finance/international-money-transfer-operators":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			case "/v1/finance/international-money-transfer-operators/{operator_id}":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			case "/v1/finance/currencies":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			case "/v1/finance/currencies/{currency_id}":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			default:
				return func(next http.Handler) http.Handler { return next }, nil
			}
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return financePolicyHarness{
		router:    routerHandler,
		auth:      auth,
		rateLimit: rateLimit,
		usage:     usage,
		finance:   finance,
		anonymous: anonymous,
	}
}

func newEducationPolicyRouter(t *testing.T) educationPolicyHarness {
	t.Helper()

	cors, err := middlewares.NewCORS(middlewares.CORSOptions{
		AllowedOrigins: []string{"https://example.com"},
	})
	if err != nil {
		t.Fatalf("NewCORS() error = %v", err)
	}

	auth := &routerAPIKeyAuthenticatorStub{identity: services.APIKeyIdentity{APIKeyID: "key_123", AccountID: "acc_123"}}
	anonymous := &routerAnonymousIdentifierStub{value: "anon-opaque"}
	rateLimit := &routerRateLimitRepoStub{}
	usage := &routerUsageRecorderStub{}
	education := &routerEducationStub{}
	educationHandler, err := handlers.NewEducationHandler(education)
	if err != nil {
		t.Fatalf("NewEducationHandler() error = %v", err)
	}
	baseHandlers := testHandlers(t, &routerRecorder{})
	rateLimitMiddleware, err := middlewares.RateLimit(rateLimit, anonymous, middlewares.RateLimitPolicy{
		AnonymousLimit: 60,
		APIKeyLimit:    300,
		DownloadLimit:  10,
		Window:         time.Minute,
	})
	if err != nil {
		t.Fatalf("RateLimit() error = %v", err)
	}

	routerHandler, err := New(Handlers{
		Health:    handlers.NewHealthHandler(),
		Discovery: handlers.NewDiscoveryHandler(),
		Geography: baseHandlers.Geography,
		Education: educationHandler,
		Finance:   baseHandlers.Finance,
		Auth:      baseHandlers.Auth,
		Account:   baseHandlers.Account,
		APIKey:    baseHandlers.APIKey,
		Usage:     baseHandlers.Usage,
		Dataset:   baseHandlers.Dataset,
	}, Middleware{
		RequestID:       middlewares.RequestID,
		Recovery:        func(next http.Handler) http.Handler { return next },
		Logger:          func(next http.Handler) http.Handler { return next },
		SecurityHeaders: func(next http.Handler) http.Handler { return next },
		CORS:            cors,
		BodyLimit:       func(next http.Handler) http.Handler { return next },
		Timeout:         func(next http.Handler) http.Handler { return next },
		Authentication:  func(next http.Handler) http.Handler { return next },
		OptionalAPIKey:  middlewares.OptionalAPIKey(auth),
		StandardLimit:   rateLimitMiddleware,
		UsageTracking: func(endpoint, datasetGroup string) (MiddlewareFunc, error) {
			switch endpoint {
			case "/v1/education/universities":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			case "/v1/education/universities/{university_id}":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			case "/v1/education/colleges-of-education":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			case "/v1/education/colleges-of-education/{college_id}":
				return middlewares.UsageTracking(usage, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
					Timeout:             time.Second,
					AnonymousIdentifier: anonymous,
				})
			default:
				return func(next http.Handler) http.Handler { return next }, nil
			}
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return educationPolicyHarness{
		router:    routerHandler,
		auth:      auth,
		rateLimit: rateLimit,
		usage:     usage,
		education: education,
		anonymous: anonymous,
	}
}

type routerAuthStub struct {
	rec *routerRecorder
}

func (s *routerAuthStub) Register(ctx context.Context, input models.AccountCreateInput) (models.AccountResponse, error) {
	s.rec.add("auth.register")
	return models.AccountResponse{ID: "acc_123", Username: input.Username, Status: models.AccountStatusActive, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}, nil
}

func (s *routerAuthStub) Login(ctx context.Context, username, password string) (models.LoginResult, error) {
	s.rec.add("auth.login")
	return models.LoginResult{
		Account: models.AccountResponse{ID: "acc_123", Username: username, Status: models.AccountStatusActive, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		Tokens:  models.TokenPair{AccessToken: "access", RefreshToken: "refresh", TokenType: "Bearer", ExpiresIn: 900},
	}, nil
}

func (s *routerAuthStub) Refresh(ctx context.Context, refreshToken string) (models.TokenPair, error) {
	s.rec.add("auth.refresh")
	return models.TokenPair{AccessToken: "access", RefreshToken: "refresh", TokenType: "Bearer", ExpiresIn: 900}, nil
}

func (s *routerAuthStub) Logout(ctx context.Context, refreshToken string) error {
	s.rec.add("auth.logout")
	return nil
}

type routerAccountStub struct {
	rec *routerRecorder
}

func (s *routerAccountStub) GetCurrent(ctx context.Context, accountID string) (models.AccountResponse, error) {
	s.rec.add("account.get")
	now := time.Now().UTC()
	return models.AccountResponse{ID: accountID, Username: "alice", Status: models.AccountStatusActive, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *routerAccountStub) UpdateCurrent(ctx context.Context, accountID string, input models.AccountUpdateInput) (models.AccountResponse, error) {
	s.rec.add("account.patch")
	now := time.Now().UTC()
	return models.AccountResponse{ID: accountID, Username: "alice", Status: models.AccountStatusActive, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *routerAccountStub) DeactivateCurrent(ctx context.Context, accountID string) error {
	s.rec.add("account.delete")
	return nil
}

type routerAPIKeyStub struct {
	rec *routerRecorder
}

func (s *routerAPIKeyStub) CreateKey(ctx context.Context, accountID string, input models.APIKeyCreateInput) (models.APIKeyCreatedResponse, error) {
	s.rec.add("apikey.create")
	return models.APIKeyCreatedResponse{Key: "sd_live_abc123", APIKey: models.APIKeyMetadata{ID: "key_123", Name: input.Name, Status: models.APIKeyStatusActive, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}}, nil
}

func (s *routerAPIKeyStub) ListKeys(ctx context.Context, accountID string) ([]models.APIKeyMetadata, error) {
	s.rec.add("apikey.list")
	return []models.APIKeyMetadata{}, nil
}

func (s *routerAPIKeyStub) RevokeKey(ctx context.Context, accountID, keyID string) error {
	s.rec.add("apikey.delete:" + keyID)
	return nil
}

func (s *routerAPIKeyStub) RotateKey(ctx context.Context, accountID, keyID string) (models.APIKeyCreatedResponse, error) {
	s.rec.add("apikey.rotate:" + keyID)
	return models.APIKeyCreatedResponse{Key: "sd_live_new", APIKey: models.APIKeyMetadata{ID: "key_456", Name: "Portfolio", Status: models.APIKeyStatusActive, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}}, nil
}

type routerUsageStub struct {
	rec *routerRecorder
}

func (s *routerUsageStub) GetUsageSummary(ctx context.Context, accountID string, apiKeyID *string, start, end time.Time) (models.UsageSummaryReportResponse, error) {
	s.rec.add("usage.summary")
	return models.UsageSummaryReportResponse{
		TotalRequests:      1,
		SuccessfulRequests: 1,
		ErrorCount:         0,
		CurrentAllowance:   50000,
		RemainingAllowance: 49999,
		PeriodStart:        start,
		PeriodEnd:          end,
	}, nil
}

func (s *routerUsageStub) GetUsageHistory(ctx context.Context, accountID string, start, end time.Time) ([]models.UsageDailyResponse, error) {
	s.rec.add("usage.history")
	return []models.UsageDailyResponse{}, nil
}

func (s *routerUsageStub) GetAPIKeyUsageHistory(ctx context.Context, accountID, apiKeyID string, start, end time.Time) ([]models.UsageDailyResponse, error) {
	s.rec.add("usage.history.api_key")
	return []models.UsageDailyResponse{}, nil
}

func (s *routerUsageStub) ListEndpointUsage(ctx context.Context, accountID string, start, end time.Time) ([]models.EndpointUsageResponse, error) {
	s.rec.add("usage.endpoints")
	return []models.EndpointUsageResponse{}, nil
}

func (s *routerUsageStub) ListAPIKeyEndpointUsage(ctx context.Context, accountID, apiKeyID string, start, end time.Time) ([]models.EndpointUsageResponse, error) {
	s.rec.add("usage.endpoints.api_key")
	return []models.EndpointUsageResponse{}, nil
}

func (s *routerUsageStub) GetDatasetGroupUsage(ctx context.Context, accountID string, start, end time.Time) ([]models.DatasetGroupUsageResponse, error) {
	s.rec.add("usage.groups")
	return []models.DatasetGroupUsageResponse{}, nil
}

func (s *routerUsageStub) GetAPIKeyDatasetGroupUsage(ctx context.Context, accountID, apiKeyID string, start, end time.Time) ([]models.DatasetGroupUsageResponse, error) {
	s.rec.add("usage.groups.api_key")
	return []models.DatasetGroupUsageResponse{}, nil
}

type routerDatasetStub struct {
	rec *routerRecorder
}

func (s *routerDatasetStub) ListDatasets(ctx context.Context, search string, page, limit int) (services.DatasetListResult, error) {
	s.rec.add("dataset.list:" + search + ":" + itoa(page) + ":" + itoa(limit))
	now := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
	return services.DatasetListResult{
		Datasets: []models.DatasetResponse{{
			ID:            "ng-states",
			Name:          "Nigerian States",
			Group:         "geography",
			Version:       "1.0.0",
			Status:        models.DatasetStatusActive,
			PrimaryFormat: "json",
			IsPublic:      true,
			CreatedAt:     now,
			UpdatedAt:     now,
		}},
		Total:      1,
		Page:       page,
		Limit:      limit,
		TotalPages: 1,
	}, nil
}

func (s *routerDatasetStub) GetDataset(ctx context.Context, datasetKey string) (models.DatasetResponse, error) {
	s.rec.add("dataset.get:" + datasetKey)
	now := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
	return models.DatasetResponse{ID: datasetKey, Name: "Nigerian States", Group: "geography", Version: "1.0.0", Status: models.DatasetStatusActive, PrimaryFormat: "json", IsPublic: true, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *routerDatasetStub) ListDatasetSources(ctx context.Context, datasetKey string) ([]models.DatasetSourceResponse, error) {
	s.rec.add("dataset.sources:" + datasetKey)
	now := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
	return []models.DatasetSourceResponse{{ID: "source-example", Name: "Official source", IsOfficial: true, CreatedAt: now, UpdatedAt: now}}, nil
}

func (s *routerDatasetStub) ListDatasetVersions(ctx context.Context, datasetKey string) ([]models.DatasetVersionResponse, error) {
	s.rec.add("dataset.versions:" + datasetKey)
	now := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
	return []models.DatasetVersionResponse{{Version: "1.0.0", Format: "json", Status: models.DatasetVersionStatusPublished, CreatedAt: now, UpdatedAt: now}}, nil
}

type routerGeographyStub struct {
	rec                      *routerRecorder
	mu                       sync.Mutex
	listCalls                int
	getCalls                 int
	zoneListCalls            int
	zoneGetCalls             int
	lgaListCalls             int
	lgaListByCalls           int
	lgaGetCalls              int
	timeZoneListCalls        int
	timeZoneGetCalls         int
	lastTimeZoneInput        services.TimeZoneListInput
	countryListCalls         int
	countryGetCalls          int
	languageListCalls        int
	languageGetCalls         int
	countryLanguageListCalls int
	profileCalls             int
	lastStateID              string
	lastZoneID               string
	lastLGAID                string
	lastTimeZoneID           string
	lastLGAStateID           string
	lastCountryID            string
	lastLanguageID           string
	lastCountryLanguageInput services.CountryLanguageListInput
	lastProfileID            string
	lastHadAPIKey            bool
	lastAPIKeyIdentity       services.APIKeyIdentity
	profileFn                func(context.Context, string) (models.CountryProfile, error)
}

type routerFinanceStub struct {
	rec                 *routerRecorder
	mu                  sync.Mutex
	listCalls           int
	listByTypeCalls     int
	getCalls            int
	listIMTOCalls       int
	getIMTOCalls        int
	listCurrencyCalls   int
	getCurrencyCalls    int
	lastInstitutionType string
	lastProviderID      string
	lastOperatorID      string
	lastCurrencyFilter  string
	lastCurrencyID      string
	lastHadAPIKey       bool
	lastAPIKeyIdentity  services.APIKeyIdentity
}

func (s *routerFinanceStub) ListPaymentServiceProviders(ctx context.Context) ([]models.PaymentServiceProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listCalls++
	if s.rec != nil {
		s.rec.add("finance.list")
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return []models.PaymentServiceProvider{{ID: "mobile-money-operator-abeg-technologies-limited", Name: "Abeg Technologies Limited", InstitutionType: "mobile_money_operator", CountryCode: "NG"}}, nil
}

func (s *routerFinanceStub) ListPaymentServiceProvidersByType(ctx context.Context, institutionType string) ([]models.PaymentServiceProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listByTypeCalls++
	s.lastInstitutionType = institutionType
	if s.rec != nil {
		s.rec.add("finance.list-by:" + institutionType)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return []models.PaymentServiceProvider{{ID: institutionType + "-example", Name: "Example", InstitutionType: institutionType, CountryCode: "NG"}}, nil
}

func (s *routerFinanceStub) GetPaymentServiceProvider(ctx context.Context, providerID string) (models.PaymentServiceProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getCalls++
	s.lastProviderID = providerID
	if s.rec != nil {
		s.rec.add("finance.get:" + providerID)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return models.PaymentServiceProvider{ID: providerID, Name: "Example", InstitutionType: "super_agent", CountryCode: "NG"}, nil
}

func (s *routerFinanceStub) ListInternationalMoneyTransferOperators(ctx context.Context) ([]models.InternationalMoneyTransferOperator, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listIMTOCalls++
	if s.rec != nil {
		s.rec.add("finance.imto.list")
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return []models.InternationalMoneyTransferOperator{{ID: "olive-monies-express-limited", Name: "OLIVE MONIES EXPRESS LIMITED"}}, nil
}

func (s *routerFinanceStub) GetInternationalMoneyTransferOperator(ctx context.Context, operatorID string) (models.InternationalMoneyTransferOperator, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getIMTOCalls++
	s.lastOperatorID = operatorID
	if s.rec != nil {
		s.rec.add("finance.imto.get:" + operatorID)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return models.InternationalMoneyTransferOperator{ID: operatorID, Name: "OLIVE MONIES EXPRESS LIMITED"}, nil
}

func (s *routerFinanceStub) ListCurrencies(ctx context.Context, input services.CurrencyListInput) ([]models.Currency, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listCurrencyCalls++
	s.lastCurrencyFilter = input.CountryAreaID
	if s.rec != nil {
		s.rec.add("finance.currency.list:" + input.CountryAreaID)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return []models.Currency{{ID: "ngn", Name: "Naira", AlphabeticCode: "NGN", NumericCode: "566", MinorUnit: 2, CountryAreaIDs: []string{"ng"}}}, nil
}

func (s *routerFinanceStub) GetCurrency(ctx context.Context, currencyID string) (models.Currency, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getCurrencyCalls++
	s.lastCurrencyID = currencyID
	if s.rec != nil {
		s.rec.add("finance.currency.get:" + currencyID)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return models.Currency{ID: currencyID, Name: "Naira", AlphabeticCode: "NGN", NumericCode: "566", MinorUnit: 2, CountryAreaIDs: []string{"ng"}}, nil
}

func (s *routerGeographyStub) ListStates(ctx context.Context) ([]models.State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listCalls++
	if s.rec != nil {
		s.rec.add("geography.list")
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return []models.State{{ID: "abia", Name: "Abia"}}, nil
}

func (s *routerGeographyStub) GetState(ctx context.Context, stateID string) (models.State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getCalls++
	s.lastStateID = stateID
	if s.rec != nil {
		s.rec.add("geography.get:" + stateID)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return models.State{ID: stateID, Name: strings.Title(stateID)}, nil
}

func (s *routerGeographyStub) ListGeopoliticalZones(ctx context.Context) ([]models.GeopoliticalZone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.zoneListCalls++
	if s.rec != nil {
		s.rec.add("geography.zone.list")
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return []models.GeopoliticalZone{{ID: "north-central", Name: "North Central"}}, nil
}

func (s *routerGeographyStub) GetGeopoliticalZone(ctx context.Context, zoneID string) (models.GeopoliticalZone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.zoneGetCalls++
	s.lastZoneID = zoneID
	if s.rec != nil {
		s.rec.add("geography.zone.get:" + zoneID)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return models.GeopoliticalZone{ID: zoneID, Name: strings.Title(strings.ReplaceAll(zoneID, "-", " "))}, nil
}

func (s *routerGeographyStub) ListLocalGovernmentUnits(ctx context.Context) ([]models.LocalGovernmentUnit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lgaListCalls++
	if s.rec != nil {
		s.rec.add("geography.lga.list")
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return []models.LocalGovernmentUnit{{ID: "lagos-ikeja", Name: "Ikeja", StateID: "lagos", CountryCode: "NG", AdministrativeType: "local_government_area"}}, nil
}

func (s *routerGeographyStub) ListLocalGovernmentUnitsByState(ctx context.Context, stateID string) ([]models.LocalGovernmentUnit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lgaListByCalls++
	s.lastLGAStateID = stateID
	if s.rec != nil {
		s.rec.add("geography.lga.list-by:" + stateID)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return []models.LocalGovernmentUnit{{ID: stateID + "-example", Name: "Example", StateID: stateID, CountryCode: "NG", AdministrativeType: "local_government_area"}}, nil
}

func (s *routerGeographyStub) GetLocalGovernmentUnit(ctx context.Context, unitID string) (models.LocalGovernmentUnit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lgaGetCalls++
	s.lastLGAID = unitID
	if s.rec != nil {
		s.rec.add("geography.lga.get:" + unitID)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return models.LocalGovernmentUnit{ID: unitID, Name: "Example", StateID: "lagos", CountryCode: "NG", AdministrativeType: "local_government_area"}, nil
}

func (s *routerGeographyStub) ListLanguages(ctx context.Context) ([]models.Language, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.languageListCalls++
	if s.rec != nil {
		s.rec.add("geography.language.list")
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return []models.Language{{ID: "en", Name: "English"}}, nil
}

func (s *routerGeographyStub) GetLanguage(ctx context.Context, languageID string) (models.Language, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.languageGetCalls++
	s.lastLanguageID = languageID
	if s.rec != nil {
		s.rec.add("geography.language.get:" + languageID)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return models.Language{ID: languageID, Name: "English"}, nil
}

func (s *routerGeographyStub) ListCountryLanguages(ctx context.Context, input services.CountryLanguageListInput) ([]models.CountryLanguage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.countryLanguageListCalls++
	s.lastCountryLanguageInput = input
	if s.rec != nil {
		s.rec.add("geography.country-language.list")
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return []models.CountryLanguage{{CountryAreaID: "ng", LanguageID: "yo", Status: "official"}}, nil
}

func (s *routerGeographyStub) ListTimeZones(ctx context.Context, input services.TimeZoneListInput) ([]models.TimeZone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timeZoneListCalls++
	s.lastTimeZoneInput = input
	if s.rec != nil {
		s.rec.add("geography.time-zone.list")
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return []models.TimeZone{{ID: "Africa/Lagos", CountryAreaIDs: []string{"ng"}}}, nil
}

func (s *routerGeographyStub) GetTimeZone(ctx context.Context, timeZoneID string) (models.TimeZone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timeZoneGetCalls++
	s.lastTimeZoneID = timeZoneID
	if s.rec != nil {
		s.rec.add("geography.time-zone.get:" + timeZoneID)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return models.TimeZone{ID: timeZoneID, CountryAreaIDs: []string{"ng"}}, nil
}

func (s *routerGeographyStub) ListCountriesAndAreas(ctx context.Context, input services.CountryOrAreaListInput) ([]models.CountryOrArea, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.countryListCalls++
	if s.rec != nil {
		s.rec.add("geography.country.list")
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return []models.CountryOrArea{{ID: "ng", Name: "Nigeria", Alpha2Code: "NG", Alpha3Code: "NGA", NumericCode: "566"}}, nil
}

func (s *routerGeographyStub) GetCountryOrArea(ctx context.Context, countryID string) (models.CountryOrArea, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.countryGetCalls++
	s.lastCountryID = countryID
	if s.rec != nil {
		s.rec.add("geography.country.get:" + countryID)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return models.CountryOrArea{ID: countryID, Name: "Nigeria", Alpha2Code: "NG", Alpha3Code: "NGA", NumericCode: "566"}, nil
}

func (s *routerGeographyStub) GetCountryProfile(ctx context.Context, countryID string) (models.CountryProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profileCalls++
	s.lastProfileID = countryID
	if s.rec != nil {
		s.rec.add("geography.country.profile:" + countryID)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	if s.profileFn != nil {
		return s.profileFn(ctx, countryID)
	}
	return models.CountryProfile{ID: countryID, Name: "Nigeria", Alpha2Code: "NG", Alpha3Code: "NGA", NumericCode: "566", CurrencyIDs: []string{}, TimeZoneIDs: []string{}}, nil
}

type routerEducationStub struct {
	rec                *routerRecorder
	mu                 sync.Mutex
	listCalls          int
	getCalls           int
	collegeListCalls   int
	collegeGetCalls    int
	lastInput          services.UniversityListInput
	lastUniversityID   string
	lastCollegeInput   services.CollegeOfEducationListInput
	lastCollegeID      string
	lastHadAPIKey      bool
	lastAPIKeyIdentity services.APIKeyIdentity
}

func (s *routerEducationStub) ListUniversities(ctx context.Context, input services.UniversityListInput) ([]models.University, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listCalls++
	s.lastInput = input
	if s.rec != nil {
		s.rec.add("education.list")
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return []models.University{{ID: "ahmadu-bello-university-zaria", Name: "Ahmadu Bello University, Zaria", OwnershipType: "federal", StateID: "kaduna", CountryCode: "NG"}}, nil
}

func (s *routerEducationStub) GetUniversity(ctx context.Context, universityID string) (models.University, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getCalls++
	s.lastUniversityID = universityID
	if s.rec != nil {
		s.rec.add("education.get:" + universityID)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return models.University{ID: universityID, Name: "Example University", OwnershipType: "state", StateID: "taraba", CountryCode: "NG"}, nil
}

func (s *routerEducationStub) ListCollegesOfEducation(ctx context.Context, input services.CollegeOfEducationListInput) ([]models.CollegeOfEducation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collegeListCalls++
	s.lastCollegeInput = input
	if s.rec != nil {
		s.rec.add("education.college.list")
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return []models.CollegeOfEducation{{ID: "federal-college-of-education-zaria", Name: "Federal College of Education, Zaria", OwnershipType: "federal", StateID: "kaduna", CountryCode: "NG"}}, nil
}

func (s *routerEducationStub) GetCollegeOfEducation(ctx context.Context, collegeID string) (models.CollegeOfEducation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collegeGetCalls++
	s.lastCollegeID = collegeID
	if s.rec != nil {
		s.rec.add("education.college.get:" + collegeID)
	}
	if identity, ok := middlewares.APIKeyIdentityFromContext(ctx); ok {
		s.lastHadAPIKey = true
		s.lastAPIKeyIdentity = identity
	}
	return models.CollegeOfEducation{ID: collegeID, Name: "Example College of Education", OwnershipType: "state", StateID: "taraba", CountryCode: "NG"}, nil
}

type routerAPIKeyAuthenticatorStub struct {
	mu       sync.Mutex
	identity services.APIKeyIdentity
	err      error
	calls    int
	last     string
}

func (s *routerAPIKeyAuthenticatorStub) Authenticate(ctx context.Context, plaintext string) (services.APIKeyIdentity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.last = plaintext
	if s.err != nil {
		return services.APIKeyIdentity{}, s.err
	}
	return s.identity, nil
}

type routerAnonymousIdentifierStub struct {
	mu    sync.Mutex
	value string
	err   error
	calls int
}

func (s *routerAnonymousIdentifierStub) Identify(r *http.Request) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	if s.err != nil {
		return "", s.err
	}
	return s.value, nil
}

type routerRateLimitRepoStub struct {
	mu      sync.Mutex
	request interfaces.RateLimitRequest
	result  interfaces.RateLimitResult
	calls   int
}

func (s *routerRateLimitRepoStub) Allow(ctx context.Context, request interfaces.RateLimitRequest) (interfaces.RateLimitResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.request = request
	if s.result.Limit != 0 {
		return s.result, nil
	}
	return interfaces.RateLimitResult{
		Allowed:   true,
		Limit:     request.Limit,
		Remaining: request.Limit - 1,
		ResetAt:   time.Now().UTC().Add(request.Window),
	}, nil
}

type routerUsageRecorderStub struct {
	mu    sync.Mutex
	input services.RequestRecordInput
	calls int
}

func (s *routerUsageRecorderStub) RecordRequest(ctx context.Context, input services.RequestRecordInput) (models.APIRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.input = input
	return models.APIRequest{ID: int64(s.calls)}, nil
}

func itoa(v int) string {
	return strconv.Itoa(v)
}
