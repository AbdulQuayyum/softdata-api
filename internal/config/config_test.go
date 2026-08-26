package config

import (
	"strings"
	"testing"
)

func TestLoadDefaultDevelopmentConfig(t *testing.T) {
	resetConfigEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Environment != AppEnvironmentDevelopment {
		t.Fatalf("unexpected environment: %s", cfg.App.Environment)
	}
	if cfg.Server.Port != 8080 {
		t.Fatalf("unexpected server port: %d", cfg.Server.Port)
	}
	if got := cfg.Server.ListenAddress(); got != "127.0.0.1:8080" {
		t.Fatalf("unexpected listen address: %s", got)
	}
	if cfg.Database.URL != nil {
		t.Fatalf("expected nil database url")
	}
	if cfg.Security.AuthTokenSecret != nil {
		t.Fatalf("expected nil auth secret")
	}
	if cfg.RateLimit.AnonymousRequestLimit != 60 || cfg.RateLimit.APIKeyRequestLimit != 300 || cfg.RateLimit.APIKeyMonthlyAllowance != 50000 {
		t.Fatalf("unexpected rate limits: %+v", cfg.RateLimit)
	}
	if cfg.RateLimit.Window.String() != "1m0s" {
		t.Fatalf("unexpected rate limit window: %s", cfg.RateLimit.Window)
	}
	if cfg.Datasets.Path != "datasets" {
		t.Fatalf("unexpected datasets path: %s", cfg.Datasets.Path)
	}
}

func TestLoadInvalidAppEnv(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("APP_ENV", "bogus")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "APP_ENV") {
		t.Fatalf("expected APP_ENV error, got %v", err)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("SERVER_PORT", "70000")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "SERVER_PORT") {
		t.Fatalf("expected SERVER_PORT error, got %v", err)
	}
}

func TestLoadInvalidDuration(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("SERVER_READ_TIMEOUT", "0s")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "SERVER_READ_TIMEOUT") {
		t.Fatalf("expected SERVER_READ_TIMEOUT error, got %v", err)
	}
}

func TestLoadInvalidRateLimit(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("ANONYMOUS_RATE_LIMIT", "0")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "ANONYMOUS_RATE_LIMIT") {
		t.Fatalf("expected ANONYMOUS_RATE_LIMIT error, got %v", err)
	}
}

func TestLoadInvalidDatabaseConnectionBounds(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("DATABASE_MIN_CONNECTIONS", "5")
	t.Setenv("DATABASE_MAX_CONNECTIONS", "4")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "DATABASE_MIN_CONNECTIONS") {
		t.Fatalf("expected connection bound error, got %v", err)
	}
}

func TestLoadMissingDatabaseURLInProduction(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("APP_ENV", string(AppEnvironmentProduction))
	t.Setenv("AUTH_TOKEN_SECRET", "dev-secret")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected DATABASE_URL error, got %v", err)
	}
}

func TestLoadMissingAuthSecretInProduction(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("APP_ENV", string(AppEnvironmentProduction))
	t.Setenv("DATABASE_URL", "postgres://example")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "AUTH_TOKEN_SECRET") {
		t.Fatalf("expected AUTH_TOKEN_SECRET error, got %v", err)
	}
}

func TestLoadProductionConfig(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("APP_ENV", string(AppEnvironmentProduction))
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("AUTH_TOKEN_SECRET", "dev-secret")
	t.Setenv("SERVER_HOST", "0.0.0.0")
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("DATASETS_PATH", "datasets/")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.App.Environment != AppEnvironmentProduction {
		t.Fatalf("unexpected environment: %s", cfg.App.Environment)
	}
	if cfg.Database.URL == nil || *cfg.Database.URL != "postgres://example" {
		t.Fatalf("unexpected database url: %+v", cfg.Database.URL)
	}
	if cfg.Security.AuthTokenSecret == nil || *cfg.Security.AuthTokenSecret != "dev-secret" {
		t.Fatalf("unexpected auth secret: %+v", cfg.Security.AuthTokenSecret)
	}
	if got := cfg.Server.ListenAddress(); got != "0.0.0.0:9090" {
		t.Fatalf("unexpected listen address: %s", got)
	}
	if cfg.Datasets.Path != "datasets" {
		t.Fatalf("unexpected datasets path: %s", cfg.Datasets.Path)
	}
}

func TestLoadDoesNotLeakSecretsInErrors(t *testing.T) {
	resetConfigEnv(t)
	t.Setenv("APP_ENV", string(AppEnvironmentProduction))
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("AUTH_TOKEN_SECRET", "supersecret")
	t.Setenv("ACCESS_TOKEN_TTL", "not-a-duration")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "supersecret") {
		t.Fatalf("error leaked secret value: %v", err)
	}
}

func resetConfigEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"APP_ENV",
		"SERVER_HOST",
		"SERVER_PORT",
		"SERVER_READ_HEADER_TIMEOUT",
		"SERVER_READ_TIMEOUT",
		"SERVER_WRITE_TIMEOUT",
		"SERVER_IDLE_TIMEOUT",
		"SERVER_SHUTDOWN_TIMEOUT",
		"DATABASE_URL",
		"DATABASE_MAX_CONNECTIONS",
		"DATABASE_MIN_CONNECTIONS",
		"DATABASE_MAX_CONNECTION_LIFETIME",
		"DATABASE_MAX_CONNECTION_IDLE_TIME",
		"DATABASE_HEALTH_CHECK_PERIOD",
		"AUTH_TOKEN_SECRET",
		"ACCESS_TOKEN_TTL",
		"REFRESH_TOKEN_TTL",
		"ANONYMOUS_RATE_LIMIT",
		"API_KEY_RATE_LIMIT",
		"API_KEY_MONTHLY_LIMIT",
		"RATE_LIMIT_WINDOW",
		"DATASETS_PATH",
	} {
		t.Setenv(key, "")
	}
}
