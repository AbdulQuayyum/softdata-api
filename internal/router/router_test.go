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

	return Handlers{
		Health:    handlers.NewHealthHandler(),
		Discovery: handlers.NewDiscoveryHandler(),
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
		{name: "health wrong method", method: http.MethodPost, target: "/health", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_405"},
		{name: "auth wrong method", method: http.MethodGet, target: "/v1/auth/login", allow: http.MethodPost, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_auth"},
		{name: "account wrong method", method: http.MethodPost, target: "/v1/account", allow: "GET, PATCH, DELETE", status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_account"},
		{name: "dataset wrong method", method: http.MethodPost, target: "/v1/datasets/ng-states", allow: http.MethodGet, status: http.StatusMethodNotAllowed, wantCode: "INVALID_REQUEST", wantRequest: "req_dataset"},
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
			if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
				t.Fatalf("unexpected content type: %q", got)
			}
			if got := rr.Header().Get("X-Request-ID"); got != tc.wantRequest {
				t.Fatalf("unexpected request id header: %q", got)
			}
			if tc.allow != "" {
				if got := rr.Header().Get("Allow"); got != tc.allow {
					t.Fatalf("unexpected allow header: %q", got)
				}
			}
			body := rr.Body.String()
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

func TestRouterRejectsHeadRequestsWithJson405(t *testing.T) {
	tests := []struct {
		name   string
		target string
		allow  string
	}{
		{name: "health", target: "/health", allow: http.MethodGet},
		{name: "account", target: "/v1/account", allow: "GET, PATCH, DELETE"},
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

func itoa(v int) string {
	return strconv.Itoa(v)
}
