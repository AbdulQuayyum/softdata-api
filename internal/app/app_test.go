package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/config"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestShutdownClosesResourcesInOrder(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	calls := make([]string, 0, 3)
	app := &App{
		shutdownServer: func(context.Context) error {
			mu.Lock()
			calls = append(calls, "server")
			mu.Unlock()
			return nil
		},
		closeRedis: func() error {
			mu.Lock()
			calls = append(calls, "redis")
			mu.Unlock()
			return nil
		},
		closePostgres: func() {
			mu.Lock()
			calls = append(calls, "postgres")
			mu.Unlock()
		},
	}

	if err := app.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := app.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() second call error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	want := []string{"server", "redis", "postgres"}
	if strings.Join(calls, ",") != strings.Join(want, ",") {
		t.Fatalf("unexpected shutdown order: got %v want %v", calls, want)
	}
}

func TestRunTreatsServerClosedAsNilAfterShutdown(t *testing.T) {
	t.Parallel()

	app := &App{
		runServer: func() error {
			return http.ErrServerClosed
		},
	}
	app.shutdownStarted.Store(true)

	if err := app.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestUnavailableRateLimitRepositoryPreservesContextCancellation(t *testing.T) {
	t.Parallel()

	repo := unavailableRateLimitRepository{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := repo.Allow(ctx, interfaces.RateLimitRequest{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("Allow() error = %v, want context.Canceled", err)
	}

	_, err := repo.Allow(context.Background(), interfaces.RateLimitRequest{})
	if !errors.Is(err, interfaces.ErrRateLimitUnavailable) {
		t.Fatalf("Allow() error = %v, want ErrRateLimitUnavailable", err)
	}
}

func TestBuildRateLimitRepositoryFallsBackWhenFailOpenAndConfigIsMissing(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{FailOpen: true},
		Redis:     config.RedisConfig{},
	}

	repo, closeFn, err := buildRateLimitRepository(context.Background(), cfg, testLogger())
	if err != nil {
		t.Fatalf("buildRateLimitRepository() error = %v", err)
	}
	if closeFn != nil {
		t.Fatal("buildRateLimitRepository() closeFn = non-nil, want nil")
	}
	if repo == nil {
		t.Fatal("buildRateLimitRepository() repo = nil, want fallback repository")
	}

	if _, err := repo.Allow(context.Background(), interfaces.RateLimitRequest{}); !errors.Is(err, interfaces.ErrRateLimitUnavailable) {
		t.Fatalf("Allow() error = %v, want ErrRateLimitUnavailable", err)
	}
}

func TestBuildRateLimitRepositoryRejectsMissingRedisConfigWhenFailOpenIsDisabled(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{FailOpen: false},
		Redis:     config.RedisConfig{},
	}

	_, _, err := buildRateLimitRepository(context.Background(), cfg, testLogger())
	if err == nil {
		t.Fatal("buildRateLimitRepository() error = nil, want error")
	}
}
