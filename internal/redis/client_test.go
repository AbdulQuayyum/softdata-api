package redis

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/config"
	redisv9 "github.com/redis/go-redis/v9"
)

type fakeRedisClient struct {
	mu       sync.Mutex
	pingErr  error
	closeErr error
	closed   bool
	pings    int
}

func (f *fakeRedisClient) Ping(ctx context.Context) *redisv9.StatusCmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pings++
	cmd := redisv9.NewStatusCmd(ctx)
	if f.pingErr != nil {
		cmd.SetErr(f.pingErr)
	}
	return cmd
}

func (f *fakeRedisClient) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return f.closeErr
}

func TestBuildRedisOptionsUsesAddressConfiguration(t *testing.T) {
	url := "redis://ignored:ignored@localhost:6379/1"
	username := "user"
	password := "pass"
	options, err := buildRedisOptions(config.RedisConfig{
		Address:      "127.0.0.1:6380",
		Username:     &username,
		Password:     &password,
		DB:           4,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 4 * time.Second,
		PoolSize:     12,
		URL:          &url,
	})
	if err != nil {
		t.Fatalf("buildRedisOptions() error = %v", err)
	}

	if options.Addr != "localhost:6379" {
		t.Fatalf("unexpected addr: %s", options.Addr)
	}
	if options.Username != "ignored" || options.Password != "ignored" || options.DB != 1 {
		t.Fatalf("URL precedence not preserved: %#v", options)
	}
	if options.DialTimeout != 2*time.Second || options.ReadTimeout != 3*time.Second || options.WriteTimeout != 4*time.Second || options.PoolSize != 12 {
		t.Fatalf("unexpected timeout/pool settings: %#v", options)
	}
}

func TestBuildRedisOptionsUsesStandaloneAddressWhenNoURL(t *testing.T) {
	username := "user"
	password := "pass"
	options, err := buildRedisOptions(config.RedisConfig{
		Address:      "127.0.0.1:6380",
		Username:     &username,
		Password:     &password,
		DB:           4,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 4 * time.Second,
		PoolSize:     12,
	})
	if err != nil {
		t.Fatalf("buildRedisOptions() error = %v", err)
	}

	if options.Addr != "127.0.0.1:6380" || options.Username != "user" || options.Password != "pass" || options.DB != 4 {
		t.Fatalf("unexpected standalone config mapping: %#v", options)
	}
}

func TestNewClientConstructsWithoutConnectivityCheck(t *testing.T) {
	t.Parallel()

	client, err := NewClient(validRedisConfig())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewClient() client = nil, want redis client")
	}

	if err := client.Close(); err != nil {
		t.Fatalf("client.Close() error = %v", err)
	}
}

func TestBuildRedisOptionsRejectsInvalidConfig(t *testing.T) {
	_, err := buildRedisOptions(config.RedisConfig{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRedisConfig) {
		t.Fatalf("unexpected error: %v", err)
	}

	url := "redis://user:secret@%zz"
	_, err = buildRedisOptions(config.RedisConfig{
		URL:          &url,
		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		PoolSize:     1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRedisConfig) {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("error leaked credentials: %v", err)
	}
}

func TestOpenRedisClosesClientOnPingFailure(t *testing.T) {
	client := &fakeRedisClient{pingErr: errors.New("connection refused")}
	opened := false

	result, err := openRedis(context.Background(), validRedisConfig(), func(*redisv9.Options) redisClient {
		opened = true
		return client
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if result != nil {
		t.Fatalf("expected nil client, got %#v", result)
	}
	if !opened || client.pings != 1 || !client.closed {
		t.Fatalf("unexpected open/close state: opened=%v pings=%d closed=%v", opened, client.pings, client.closed)
	}
	if !errors.Is(err, ErrRedisUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenRedisPreservesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opened := false
	_, err := openRedis(ctx, validRedisConfig(), func(*redisv9.Options) redisClient {
		opened = true
		return &fakeRedisClient{}
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error: %v", err)
	}
	if opened {
		t.Fatal("opener should not be called for canceled context")
	}
}

func TestOpenRedisPreservesContextDeadline(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	opened := false
	_, err := openRedis(ctx, validRedisConfig(), func(*redisv9.Options) redisClient {
		opened = true
		return &fakeRedisClient{}
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected error: %v", err)
	}
	if opened {
		t.Fatal("opener should not be called for expired context")
	}
}

func TestBuildRedisOptionsRejectsNegativeDB(t *testing.T) {
	cfg := validRedisConfig()
	cfg.DB = -1

	_, err := buildRedisOptions(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRedisConfig) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenRedisConcurrentUsageHasNoSharedState(t *testing.T) {
	const workers = 8
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			_, err := buildRedisOptions(validRedisConfig())
			if err != nil {
				t.Errorf("buildRedisOptions() error = %v", err)
			}
		}()
	}
	wg.Wait()
}

func validRedisConfig() config.RedisConfig {
	return config.RedisConfig{
		Address:      "127.0.0.1:6379",
		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		PoolSize:     10,
	}
}
