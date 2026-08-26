package database

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/config"
)

func TestNewPostgresMissingURL(t *testing.T) {
	cfg := validTestDatabaseConfig()
	cfg.URL = nil

	_, err := NewPostgres(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "database url is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewPostgresMalformedURL(t *testing.T) {
	cfg := validTestDatabaseConfig()
	badURL := "postgres://user:secret@%zz"
	cfg.URL = &badURL

	_, err := NewPostgres(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("error leaked credentials: %v", err)
	}
}

func TestReadyNilPool(t *testing.T) {
	err := Ready(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "database is unavailable") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func validTestDatabaseConfig() config.DatabaseConfig {
	url := "postgres://user:pass@127.0.0.1:5432/softdata?sslmode=disable"
	return config.DatabaseConfig{
		URL:                   &url,
		MaxConnections:        4,
		MinConnections:        1,
		MaxConnectionLifetime: 30 * time.Minute,
		MaxConnectionIdleTime: 5 * time.Minute,
		HealthCheckPeriod:     time.Minute,
	}
}
