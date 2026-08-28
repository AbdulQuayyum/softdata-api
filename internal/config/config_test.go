package config

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestLoadDevelopmentConfig(t *testing.T) {
	cfg, err := load(lookupFromMap(map[string]string{
		"APP_ENV":             string(AppEnvironmentDevelopment),
		"DATABASE_URL":        "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":   strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET": strings.Repeat("b", 32),
		"DATASETS_PATH":       "datasets/",
	}))
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	if cfg.Environment != string(AppEnvironmentDevelopment) {
		t.Fatalf("unexpected environment: %s", cfg.Environment)
	}
	if got := cfg.Server.ListenAddress(); got != "127.0.0.1:8080" {
		t.Fatalf("unexpected listen address: %s", got)
	}
	if cfg.Server.MaxBodyBytes != 1<<20 {
		t.Fatalf("unexpected max body bytes: %d", cfg.Server.MaxBodyBytes)
	}
	if cfg.Database.URL == nil || *cfg.Database.URL != "postgres://localhost/softdata" {
		t.Fatalf("unexpected database url: %#v", cfg.Database.URL)
	}
	if cfg.Database.ConnectTimeout.String() != "10s" {
		t.Fatalf("unexpected database connect timeout: %s", cfg.Database.ConnectTimeout)
	}
	if cfg.Redis.URL != nil || cfg.Redis.Address != "" || cfg.Redis.PoolSize != 10 {
		t.Fatalf("unexpected redis config: %#v", cfg.Redis)
	}
	if cfg.Security.AuthTokenSecret != strings.Repeat("a", 32) {
		t.Fatalf("unexpected auth token secret: %q", cfg.Security.AuthTokenSecret)
	}
	if cfg.Security.JWTIssuer != "softdata-api" || cfg.Security.JWTAudience != "softdata-api" {
		t.Fatalf("unexpected jwt defaults: %#v", cfg.Security)
	}
	if cfg.RateLimit.AnonymousRequestLimit != 60 || cfg.RateLimit.APIKeyRequestLimit != 300 || cfg.RateLimit.DatasetDownloadLimit != 10 {
		t.Fatalf("unexpected rate limits: %#v", cfg.RateLimit)
	}
	if !cfg.RateLimit.FailOpen {
		t.Fatalf("expected fail-open rate limiting by default")
	}
	if cfg.Usage.APIKeyMonthlyAllowance != 50000 {
		t.Fatalf("unexpected usage allowance: %#v", cfg.Usage)
	}
	if cfg.Datasets.Path != "datasets" {
		t.Fatalf("unexpected datasets path: %s", cfg.Datasets.Path)
	}
	if cfg.Datasets.JSONMaxBytes != 16777216 {
		t.Fatalf("unexpected datasets json max bytes: %d", cfg.Datasets.JSONMaxBytes)
	}
}

func TestLoadRejectsInvalidEnvironment(t *testing.T) {
	_, err := load(lookupFromMap(map[string]string{
		"APP_ENV":             "bogus",
		"DATABASE_URL":        "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":   strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET": strings.Repeat("b", 32),
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "APP_ENV") {
		t.Fatalf("expected APP_ENV error, got %v", err)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	_, err := load(lookupFromMap(map[string]string{
		"SERVER_PORT":         "70000",
		"DATABASE_URL":        "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":   strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET": strings.Repeat("b", 32),
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "SERVER_PORT") {
		t.Fatalf("expected SERVER_PORT error, got %v", err)
	}
}

func TestLoadRejectsInvalidDuration(t *testing.T) {
	_, err := load(lookupFromMap(map[string]string{
		"SERVER_READ_TIMEOUT": "0s",
		"DATABASE_URL":        "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":   strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET": strings.Repeat("b", 32),
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "SERVER_READ_TIMEOUT") {
		t.Fatalf("expected SERVER_READ_TIMEOUT error, got %v", err)
	}
}

func TestLoadRejectsInvalidDatabaseBounds(t *testing.T) {
	_, err := load(lookupFromMap(map[string]string{
		"DATABASE_URL":             "postgres://localhost/softdata",
		"DATABASE_MIN_CONNECTIONS": "5",
		"DATABASE_MAX_CONNECTIONS": "4",
		"AUTH_TOKEN_SECRET":        strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET":      strings.Repeat("b", 32),
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "DATABASE_MIN_CONNECTIONS") {
		t.Fatalf("expected connection bound error, got %v", err)
	}
}

func TestLoadRejectsMissingRequiredDatabaseURL(t *testing.T) {
	_, err := load(lookupFromMap(map[string]string{
		"AUTH_TOKEN_SECRET":   strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET": strings.Repeat("b", 32),
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected DATABASE_URL error, got %v", err)
	}
}

func TestLoadRejectsMissingRequiredSecrets(t *testing.T) {
	_, err := load(lookupFromMap(map[string]string{
		"DATABASE_URL":        "postgres://localhost/softdata",
		"ANONYMOUS_ID_SECRET": strings.Repeat("b", 32),
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "AUTH_TOKEN_SECRET") {
		t.Fatalf("expected AUTH_TOKEN_SECRET error, got %v", err)
	}

	_, err = load(lookupFromMap(map[string]string{
		"DATABASE_URL":      "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET": strings.Repeat("a", 32),
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "ANONYMOUS_ID_SECRET") {
		t.Fatalf("expected ANONYMOUS_ID_SECRET error, got %v", err)
	}
}

func TestLoadRejectsShortSecrets(t *testing.T) {
	_, err := load(lookupFromMap(map[string]string{
		"DATABASE_URL":        "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":   "short",
		"ANONYMOUS_ID_SECRET": strings.Repeat("b", 32),
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "AUTH_TOKEN_SECRET") {
		t.Fatalf("expected AUTH_TOKEN_SECRET length error, got %v", err)
	}

	_, err = load(lookupFromMap(map[string]string{
		"DATABASE_URL":        "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":   strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET": "short",
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "ANONYMOUS_ID_SECRET") {
		t.Fatalf("expected ANONYMOUS_ID_SECRET length error, got %v", err)
	}
}

func TestLoadRejectsInvalidRefreshTTL(t *testing.T) {
	_, err := load(lookupFromMap(map[string]string{
		"DATABASE_URL":        "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":   strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET": strings.Repeat("b", 32),
		"ACCESS_TOKEN_TTL":    "15m",
		"REFRESH_TOKEN_TTL":   "10m",
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "REFRESH_TOKEN_TTL") {
		t.Fatalf("expected REFRESH_TOKEN_TTL error, got %v", err)
	}
}

func TestLoadRejectsInvalidOrigins(t *testing.T) {
	_, err := load(lookupFromMap(map[string]string{
		"DATABASE_URL":           "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":      strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET":    strings.Repeat("b", 32),
		"SERVER_ALLOWED_ORIGINS": "https://example.com/path",
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "SERVER_ALLOWED_ORIGINS") {
		t.Fatalf("expected origin error, got %v", err)
	}
}

func TestLoadRejectsInvalidRedisValues(t *testing.T) {
	_, err := load(lookupFromMap(map[string]string{
		"DATABASE_URL":        "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":   strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET": strings.Repeat("b", 32),
		"REDIS_DB":            "-1",
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "REDIS_DB") {
		t.Fatalf("expected REDIS_DB error, got %v", err)
	}

	_, err = load(lookupFromMap(map[string]string{
		"DATABASE_URL":        "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":   strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET": strings.Repeat("b", 32),
		"REDIS_POOL_SIZE":     "0",
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "REDIS_POOL_SIZE") {
		t.Fatalf("expected REDIS_POOL_SIZE error, got %v", err)
	}
}

func TestLoadUsesUsageConfigForMonthlyAllowance(t *testing.T) {
	cfg, err := load(lookupFromMap(map[string]string{
		"DATABASE_URL":                "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":           strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET":         strings.Repeat("b", 32),
		"API_KEY_MONTHLY_LIMIT":       "123456",
		"ANONYMOUS_RATE_LIMIT":        "61",
		"API_KEY_RATE_LIMIT":          "301",
		"DATASET_DOWNLOAD_RATE_LIMIT": "11",
	}))
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	if cfg.Usage.APIKeyMonthlyAllowance != 123456 {
		t.Fatalf("unexpected monthly allowance: %#v", cfg.Usage)
	}
	if cfg.RateLimit.AnonymousRequestLimit != 61 || cfg.RateLimit.APIKeyRequestLimit != 301 || cfg.RateLimit.DatasetDownloadLimit != 11 {
		t.Fatalf("unexpected rate limits: %#v", cfg.RateLimit)
	}
}

func TestLoadRejectsInvalidMonthlyAllowance(t *testing.T) {
	_, err := load(lookupFromMap(map[string]string{
		"DATABASE_URL":          "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":     strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET":   strings.Repeat("b", 32),
		"API_KEY_MONTHLY_LIMIT": "0",
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "API_KEY_MONTHLY_LIMIT") {
		t.Fatalf("expected API_KEY_MONTHLY_LIMIT error, got %v", err)
	}
}

func TestLoadDoesNotDeriveMonthlyAllowanceFromRateLimit(t *testing.T) {
	cfg, err := load(lookupFromMap(map[string]string{
		"DATABASE_URL":                "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":           strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET":         strings.Repeat("b", 32),
		"ANONYMOUS_RATE_LIMIT":        "60",
		"API_KEY_RATE_LIMIT":          "300",
		"DATASET_DOWNLOAD_RATE_LIMIT": "10",
	}))
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	if cfg.Usage.APIKeyMonthlyAllowance != 50000 {
		t.Fatalf("expected monthly allowance default to remain independent, got %#v", cfg.Usage)
	}
	if cfg.RateLimit.AnonymousRequestLimit != 60 || cfg.RateLimit.APIKeyRequestLimit != 300 || cfg.RateLimit.DatasetDownloadLimit != 10 {
		t.Fatalf("unexpected rate limits: %#v", cfg.RateLimit)
	}
}

func TestLoadDoesNotLeakSecretsInErrors(t *testing.T) {
	_, err := load(lookupFromMap(map[string]string{
		"DATABASE_URL":        "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":   "supersecret-supersecret-supersecret",
		"ANONYMOUS_ID_SECRET": strings.Repeat("b", 32),
		"ACCESS_TOKEN_TTL":    "not-a-duration",
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "supersecret") {
		t.Fatalf("error leaked secret value: %v", err)
	}
}

func TestLoadDatasetsConfigUsesDefaultJsonMaxBytes(t *testing.T) {
	cfg, err := loadDatasetsConfig(lookupFromMap(map[string]string{}))
	if err != nil {
		t.Fatalf("loadDatasetsConfig() error = %v", err)
	}
	if cfg.JSONMaxBytes != 16777216 {
		t.Fatalf("unexpected JSONMaxBytes: %d", cfg.JSONMaxBytes)
	}
}

func TestLoadDatasetsConfigParsesExplicitJsonMaxBytes(t *testing.T) {
	cfg, err := loadDatasetsConfig(lookupFromMap(map[string]string{
		"DATASETS_JSON_MAX_BYTES": "33554432",
	}))
	if err != nil {
		t.Fatalf("loadDatasetsConfig() error = %v", err)
	}
	if cfg.JSONMaxBytes != 33554432 {
		t.Fatalf("unexpected JSONMaxBytes: %d", cfg.JSONMaxBytes)
	}
}

func TestLoadDatasetsConfigAcceptsOneByteLimit(t *testing.T) {
	cfg, err := loadDatasetsConfig(lookupFromMap(map[string]string{
		"DATASETS_JSON_MAX_BYTES": "1",
	}))
	if err != nil {
		t.Fatalf("loadDatasetsConfig() error = %v", err)
	}
	if cfg.JSONMaxBytes != 1 {
		t.Fatalf("unexpected JSONMaxBytes: %d", cfg.JSONMaxBytes)
	}
}

func TestLoadDatasetsConfigRejectsInvalidJsonMaxBytes(t *testing.T) {
	tests := []struct {
		name string
		val  string
	}{
		{name: "zero", val: "0"},
		{name: "negative", val: "-1"},
		{name: "malformed", val: "16MB"},
		{name: "overflow", val: "9223372036854775808"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := loadDatasetsConfig(lookupFromMap(map[string]string{
				"DATASETS_JSON_MAX_BYTES": tc.val,
			}))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "DATASETS_JSON_MAX_BYTES") {
				t.Fatalf("expected DATASETS_JSON_MAX_BYTES error, got %v", err)
			}
		})
	}
}

func TestLoadDatasetsConfigDoesNotReadServerBodyLimit(t *testing.T) {
	lookup := func(name string) (string, bool) {
		switch name {
		case "DATASETS_JSON_MAX_BYTES":
			return "1234", true
		case "SERVER_BODY_LIMIT":
			t.Fatal("loadDatasetsConfig() should not read SERVER_BODY_LIMIT")
		}
		return "", false
	}

	cfg, err := loadDatasetsConfig(lookup)
	if err != nil {
		t.Fatalf("loadDatasetsConfig() error = %v", err)
	}
	if cfg.JSONMaxBytes != 1234 {
		t.Fatalf("unexpected JSONMaxBytes: %d", cfg.JSONMaxBytes)
	}
}

func TestLoadDoesNotMutateProcessEnvironment(t *testing.T) {
	before := append([]string(nil), os.Environ()...)
	_, err := load(lookupFromMap(map[string]string{
		"DATABASE_URL":                "postgres://localhost/softdata",
		"AUTH_TOKEN_SECRET":           strings.Repeat("a", 32),
		"ANONYMOUS_ID_SECRET":         strings.Repeat("b", 32),
		"DATASETS_JSON_MAX_BYTES":     "16777217",
		"DATASETS_PATH":               "datasets/",
		"ANONYMOUS_RATE_LIMIT":        "60",
		"API_KEY_RATE_LIMIT":          "300",
		"DATASET_DOWNLOAD_RATE_LIMIT": "10",
	}))
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}
	after := os.Environ()
	if !reflect.DeepEqual(before, after) {
		t.Fatal("process environment changed during config loading")
	}
}

func lookupFromMap(values map[string]string) LookupEnv {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}
