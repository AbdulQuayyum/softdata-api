package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/config"
	redisclient "github.com/AbdulQuayyum/softdata-api/internal/redis"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	redisrepo "github.com/AbdulQuayyum/softdata-api/internal/repository/redis"
	redisv9 "github.com/redis/go-redis/v9"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func validRedisBootstrapConfig(failOpen bool) *config.Config {
	return &config.Config{
		RateLimit: config.RateLimitConfig{
			AnonymousRequestLimit: 60,
			APIKeyRequestLimit:    300,
			DatasetDownloadLimit:  10,
			Window:                time.Minute,
			FailOpen:              failOpen,
		},
		Redis: config.RedisConfig{
			Address:      "127.0.0.1:63999",
			DB:           0,
			DialTimeout:  500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
			WriteTimeout: 500 * time.Millisecond,
			PoolSize:     1,
		},
	}
}

type fakeRedisBootstrapClient struct {
	mu       sync.Mutex
	pingErr  error
	closeErr error
	evalErr  error
	evalVal  any
	closed   bool
	pings    int
	evals    int
}

func (f *fakeRedisBootstrapClient) Ping(ctx context.Context) *redisv9.StatusCmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pings++
	cmd := redisv9.NewStatusCmd(ctx)
	if f.pingErr != nil {
		cmd.SetErr(f.pingErr)
	}
	return cmd
}

func (f *fakeRedisBootstrapClient) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return f.closeErr
}

func (f *fakeRedisBootstrapClient) Eval(ctx context.Context, script string, keys []string, args ...any) *redisv9.Cmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.evals++
	cmd := redisv9.NewCmd(ctx, "eval")
	if f.evalErr != nil {
		cmd.SetErr(f.evalErr)
		return cmd
	}
	cmd.SetVal(f.evalVal)
	return cmd
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

func TestBuildRateLimitRepositoryRejectsInvalidRedisConfigRegardlessOfFailOpen(t *testing.T) {
	t.Parallel()

	for _, failOpen := range []bool{false, true} {
		failOpen := failOpen
		t.Run(strconv.FormatBool(failOpen), func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				RateLimit: config.RateLimitConfig{FailOpen: failOpen},
				Redis:     config.RedisConfig{},
			}

			repo, closeFn, err := buildRateLimitRepository(context.Background(), cfg, testLogger())
			if err == nil {
				t.Fatal("buildRateLimitRepository() error = nil, want error")
			}
			if !errors.Is(err, redisclient.ErrInvalidRedisConfig) {
				t.Fatalf("buildRateLimitRepository() error = %v, want ErrInvalidRedisConfig", err)
			}
			if repo != nil {
				t.Fatalf("buildRateLimitRepository() repo = %#v, want nil", repo)
			}
			if closeFn != nil {
				t.Fatal("buildRateLimitRepository() closeFn = non-nil, want nil")
			}
		})
	}
}

func TestBuildRateLimitRepositoryUsesNormalRepositoryAfterSuccessfulPing(t *testing.T) {
	t.Parallel()

	client := &fakeRedisBootstrapClient{
		evalVal: []any{int64(1), int64(1), int64(60000), int64(1787756400), int64(0)},
	}
	cfg := validRedisBootstrapConfig(false)
	repo, closeFn, err := buildRateLimitRepositoryWith(
		context.Background(),
		cfg,
		testLogger(),
		func(config.RedisConfig) (redisBootstrapClient, error) {
			return client, nil
		},
		pingRedis,
		func(client redisBootstrapClient, prefix string) (interfaces.RateLimitRepository, error) {
			return redisrepo.NewRateLimitRepository(client, prefix)
		},
	)
	if err != nil {
		t.Fatalf("buildRateLimitRepository() error = %v", err)
	}
	if closeFn == nil {
		t.Fatal("buildRateLimitRepository() closeFn = nil, want redis close function")
	}
	if _, ok := repo.(*redisrepo.RateLimitRepository); !ok {
		t.Fatalf("buildRateLimitRepository() repo = %T, want *redisrepo.RateLimitRepository", repo)
	}

	result, err := repo.Allow(context.Background(), interfaces.RateLimitRequest{
		SubjectKind: interfaces.RateLimitSubjectAPIKey,
		Subject:     "api-key-123",
		Limit:       300,
		Window:      time.Minute,
	})
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if !result.Allowed || result.Remaining != 299 {
		t.Fatalf("unexpected rate limit result: %#v", result)
	}
	if client.evals == 0 {
		t.Fatal("expected repo.Allow() to use the real Redis repository")
	}

	if err := closeFn(); err != nil {
		t.Fatalf("closeFn() error = %v", err)
	}
}

func TestBuildRateLimitRepositoryKeepsRealRepositoryOnPingFailureWhenFailOpen(t *testing.T) {
	t.Parallel()

	client := &fakeRedisBootstrapClient{
		pingErr: errors.New("connection refused"),
		evalErr: errors.New("connection refused"),
	}
	cfg := validRedisBootstrapConfig(true)
	repo, closeFn, err := buildRateLimitRepositoryWith(
		context.Background(),
		cfg,
		testLogger(),
		func(config.RedisConfig) (redisBootstrapClient, error) {
			return client, nil
		},
		pingRedis,
		func(client redisBootstrapClient, prefix string) (interfaces.RateLimitRepository, error) {
			return redisrepo.NewRateLimitRepository(client, prefix)
		},
	)
	if err != nil {
		t.Fatalf("buildRateLimitRepository() error = %v", err)
	}
	if closeFn == nil {
		t.Fatal("buildRateLimitRepository() closeFn = nil, want redis close function")
	}
	if client.closed {
		t.Fatal("buildRateLimitRepository() closed the Redis client in fail-open mode")
	}

	if _, ok := repo.(*redisrepo.RateLimitRepository); !ok {
		t.Fatalf("buildRateLimitRepository() repo = %T, want *redisrepo.RateLimitRepository", repo)
	}

	if _, err := repo.Allow(context.Background(), interfaces.RateLimitRequest{
		SubjectKind: interfaces.RateLimitSubjectAPIKey,
		Subject:     "api-key-123",
		Limit:       1,
		Window:      time.Minute,
	}); !errors.Is(err, interfaces.ErrRateLimitUnavailable) {
		t.Fatalf("Allow() error = %v, want ErrRateLimitUnavailable", err)
	}

	if err := closeFn(); err != nil {
		t.Fatalf("closeFn() error = %v", err)
	}
	if !client.closed {
		t.Fatal("closeFn() did not close the retained Redis client")
	}
}

func TestBuildRateLimitRepositoryFailsClosedOnPingFailure(t *testing.T) {
	t.Parallel()

	client := &fakeRedisBootstrapClient{
		pingErr: errors.New("connection refused"),
	}
	cfg := validRedisBootstrapConfig(false)
	repo, closeFn, err := buildRateLimitRepositoryWith(
		context.Background(),
		cfg,
		testLogger(),
		func(config.RedisConfig) (redisBootstrapClient, error) {
			return client, nil
		},
		pingRedis,
		func(client redisBootstrapClient, prefix string) (interfaces.RateLimitRepository, error) {
			return redisrepo.NewRateLimitRepository(client, prefix)
		},
	)
	if err == nil {
		t.Fatal("buildRateLimitRepository() error = nil, want error")
	}
	if !errors.Is(err, redisclient.ErrRedisUnavailable) {
		t.Fatalf("buildRateLimitRepository() error = %v, want ErrRedisUnavailable", err)
	}
	if repo != nil {
		t.Fatalf("buildRateLimitRepository() repo = %#v, want nil", repo)
	}
	if closeFn != nil {
		t.Fatal("buildRateLimitRepository() closeFn = non-nil, want nil after startup failure")
	}
	if !client.closed {
		t.Fatal("buildRateLimitRepository() did not close the Redis client on fail-closed ping failure")
	}
}
