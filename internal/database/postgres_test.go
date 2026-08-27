package database

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type fakePostgresPool struct {
	mu      sync.Mutex
	pingErr error
	closed  bool
	pings   int
}

func (f *fakePostgresPool) Ping(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pings++
	return f.pingErr
}

func (f *fakePostgresPool) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
}

func TestBuildPostgresPoolConfigMapsAllFields(t *testing.T) {
	url := "postgres://user:pass@127.0.0.1:5432/softdata?sslmode=disable"
	cfg := validPostgresConfig()
	cfg.URL = &url

	poolConfig, err := buildPostgresPoolConfig(cfg)
	if err != nil {
		t.Fatalf("buildPostgresPoolConfig() error = %v", err)
	}

	if poolConfig.MaxConns != int32(cfg.MaxConnections) || poolConfig.MinConns != int32(cfg.MinConnections) {
		t.Fatalf("unexpected connection bounds: %#v", poolConfig)
	}
	if poolConfig.MaxConnLifetime != cfg.MaxConnectionLifetime || poolConfig.MaxConnIdleTime != cfg.MaxConnectionIdleTime || poolConfig.HealthCheckPeriod != cfg.HealthCheckPeriod {
		t.Fatalf("unexpected duration mapping: %#v", poolConfig)
	}
	if poolConfig.ConnConfig.ConnectTimeout != cfg.ConnectTimeout {
		t.Fatalf("unexpected connect timeout: %#v", poolConfig.ConnConfig)
	}
	if *cfg.URL != url {
		t.Fatalf("configuration URL was mutated: got %q want %q", *cfg.URL, url)
	}
}

func TestBuildPostgresPoolConfigRejectsEmptyURL(t *testing.T) {
	cfg := validPostgresConfig()
	cfg.URL = nil

	_, err := buildPostgresPoolConfig(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidDatabaseConfig) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildPostgresPoolConfigRejectsInvalidBounds(t *testing.T) {
	cfg := validPostgresConfig()
	cfg.MinConnections = cfg.MaxConnections + 1

	_, err := buildPostgresPoolConfig(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidDatabaseConfig) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildPostgresPoolConfigRejectsInvalidDuration(t *testing.T) {
	cfg := validPostgresConfig()
	cfg.ConnectTimeout = 0

	_, err := buildPostgresPoolConfig(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidDatabaseConfig) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenPostgresClosesPoolOnPingFailure(t *testing.T) {
	client := &fakePostgresPool{pingErr: errors.New("connection refused")}
	opened := false

	result, err := openPostgres(context.Background(), validPostgresConfig(), func(context.Context, *pgxpool.Config) (postgresPool, error) {
		opened = true
		return client, nil
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if result != nil {
		t.Fatalf("expected nil pool, got %#v", result)
	}
	if !opened || client.pings != 1 || !client.closed {
		t.Fatalf("unexpected open/close state: opened=%v pings=%d closed=%v", opened, client.pings, client.closed)
	}
	if !errors.Is(err, ErrDatabaseUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenPostgresPreservesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opened := false
	_, err := openPostgres(ctx, validPostgresConfig(), func(context.Context, *pgxpool.Config) (postgresPool, error) {
		opened = true
		return &fakePostgresPool{}, nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error: %v", err)
	}
	if opened {
		t.Fatal("opener should not be called for canceled context")
	}
}

func TestOpenPostgresPreservesContextDeadline(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	opened := false
	_, err := openPostgres(ctx, validPostgresConfig(), func(context.Context, *pgxpool.Config) (postgresPool, error) {
		opened = true
		return &fakePostgresPool{}, nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected error: %v", err)
	}
	if opened {
		t.Fatal("opener should not be called for expired context")
	}
}

func TestOpenPostgresConcurrentUsageHasNoSharedState(t *testing.T) {
	const workers = 8
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			_, err := buildPostgresPoolConfig(validPostgresConfig())
			if err != nil {
				t.Errorf("buildPostgresPoolConfig() error = %v", err)
			}
		}()
	}

	wg.Wait()
}

func TestNewPostgresRejectsMalformedURLWithoutLeakingCredentials(t *testing.T) {
	cfg := validPostgresConfig()
	badURL := "postgres://user:secret@%zz"
	cfg.URL = &badURL

	_, err := NewPostgres(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidDatabaseConfig) {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("error leaked credentials: %v", err)
	}
}

func validPostgresConfig() config.DatabaseConfig {
	url := "postgres://user:pass@127.0.0.1:5432/softdata?sslmode=disable"
	return config.DatabaseConfig{
		URL:                   &url,
		MaxConnections:        4,
		MinConnections:        1,
		MaxConnectionLifetime: 30 * time.Minute,
		MaxConnectionIdleTime: 5 * time.Minute,
		HealthCheckPeriod:     time.Minute,
		ConnectTimeout:        2 * time.Second,
	}
}
