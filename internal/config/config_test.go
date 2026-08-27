package config

import (
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
	if cfg.Datasets.Path != "datasets" {
		t.Fatalf("unexpected datasets path: %s", cfg.Datasets.Path)
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

func lookupFromMap(values map[string]string) LookupEnv {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}
