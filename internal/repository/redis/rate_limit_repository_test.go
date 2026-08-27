package redis

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type evalFake struct {
	mu     sync.Mutex
	script string
	keys   []string
	args   []any
	result any
	err    error
	calls  int
}

func (f *evalFake) Eval(ctx context.Context, script string, keys []string, args ...any) *redis.Cmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.script = script
	f.keys = append([]string(nil), keys...)
	f.args = append([]any(nil), args...)
	cmd := redis.NewCmd(ctx, "eval")
	if f.err != nil {
		cmd.SetErr(f.err)
		return cmd
	}
	cmd.SetVal(f.result)
	return cmd
}

func TestNewRateLimitRepositoryRejectsNilClient(t *testing.T) {
	if _, err := NewRateLimitRepository(nil, "softdata"); err == nil {
		t.Fatal("NewRateLimitRepository() error = nil, want error")
	}
}

func TestAllowValidRequestBuildsSafeKeyAndAllowsFirstRequest(t *testing.T) {
	fake := &evalFake{result: []any{int64(1), int64(1), int64(60000), int64(1787756400), int64(0)}}
	repo, err := NewRateLimitRepository(fake, "softdata")
	if err != nil {
		t.Fatalf("NewRateLimitRepository() error = %v", err)
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

	if !result.Allowed || result.Limit != 300 || result.Remaining != 299 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.ResetAt.Location() != time.UTC {
		t.Fatalf("reset time must be UTC: %v", result.ResetAt.Location())
	}
	if fake.calls != 1 {
		t.Fatalf("unexpected eval call count: %d", fake.calls)
	}
	if len(fake.keys) != 1 {
		t.Fatalf("unexpected keys: %#v", fake.keys)
	}
	key := fake.keys[0]
	if strings.Contains(key, "api-key-123") || strings.Contains(key, "Bearer") || strings.Contains(key, "127.0.0.1") {
		t.Fatalf("unsafe value leaked into redis key: %q", key)
	}
	if !strings.HasPrefix(key, "softdata:ratelimit:v1:api_key:60000:") {
		t.Fatalf("unexpected key prefix: %q", key)
	}
}

func TestAllowAtLimitAndAboveLimit(t *testing.T) {
	cases := []struct {
		name          string
		result        []any
		wantAllowed   bool
		wantRemaining int64
	}{
		{
			name:          "at limit",
			result:        []any{int64(1), int64(300), int64(60000), int64(1787756400), int64(0)},
			wantAllowed:   true,
			wantRemaining: 0,
		},
		{
			name:          "above limit",
			result:        []any{int64(0), int64(301), int64(60000), int64(1787756400), int64(0)},
			wantAllowed:   false,
			wantRemaining: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := &evalFake{result: tc.result}
			repo, err := NewRateLimitRepository(fake, "softdata")
			if err != nil {
				t.Fatalf("NewRateLimitRepository() error = %v", err)
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
			if result.Allowed != tc.wantAllowed {
				t.Fatalf("unexpected allowed value: %#v", result)
			}
			if result.Remaining != tc.wantRemaining {
				t.Fatalf("unexpected remaining value: %#v", result)
			}
		})
	}
}

func TestAllowRejectsInvalidInput(t *testing.T) {
	repo, err := NewRateLimitRepository(&evalFake{}, "softdata")
	if err != nil {
		t.Fatalf("NewRateLimitRepository() error = %v", err)
	}

	cases := []interfaces.RateLimitRequest{
		{},
		{SubjectKind: interfaces.RateLimitSubjectAPIKey, Subject: "", Limit: 1, Window: time.Minute},
		{SubjectKind: interfaces.RateLimitSubjectAPIKey, Subject: "valid", Limit: 0, Window: time.Minute},
		{SubjectKind: interfaces.RateLimitSubjectAPIKey, Subject: "valid", Limit: 1, Window: 0},
		{SubjectKind: interfaces.RateLimitSubjectAPIKey, Subject: strings.Repeat("a", 257), Limit: 1, Window: time.Minute},
		{SubjectKind: "unsupported", Subject: "valid", Limit: 1, Window: time.Minute},
	}

	for _, req := range cases {
		if _, err := repo.Allow(context.Background(), req); !errors.Is(err, interfaces.ErrInvalidRateLimitInput) {
			t.Fatalf("Allow() error = %v, want ErrInvalidRateLimitInput", err)
		}
	}
}

func TestAllowPropagatesContextErrors(t *testing.T) {
	repo, err := NewRateLimitRepository(&evalFake{}, "softdata")
	if err != nil {
		t.Fatalf("NewRateLimitRepository() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repo.Allow(ctx, interfaces.RateLimitRequest{
		SubjectKind: interfaces.RateLimitSubjectDownload,
		Subject:     "download-1",
		Limit:       10,
		Window:      time.Minute,
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("Allow() error = %v, want context.Canceled", err)
	}
}

func TestAllowTranslatesRedisFailure(t *testing.T) {
	fake := &evalFake{err: errors.New("dial tcp 127.0.0.1:6379: connect: connection refused")}
	repo, err := NewRateLimitRepository(fake, "softdata")
	if err != nil {
		t.Fatalf("NewRateLimitRepository() error = %v", err)
	}

	_, err = repo.Allow(context.Background(), interfaces.RateLimitRequest{
		SubjectKind: interfaces.RateLimitSubjectAnonymous,
		Subject:     "anonymous-id",
		Limit:       60,
		Window:      time.Minute,
	})
	if !errors.Is(err, interfaces.ErrRateLimitUnavailable) {
		t.Fatalf("Allow() error = %v, want ErrRateLimitUnavailable", err)
	}
	if strings.Contains(err.Error(), "127.0.0.1") || strings.Contains(err.Error(), "connect") {
		t.Fatalf("unsafe redis details leaked into error: %v", err)
	}
}

func TestSubjectKindsDoNotCollide(t *testing.T) {
	anonymousKey, err := buildRateLimitKey("softdata", interfaces.RateLimitSubjectAnonymous, "subject-123", time.Minute)
	if err != nil {
		t.Fatalf("buildRateLimitKey() error = %v", err)
	}
	apiKey, err := buildRateLimitKey("softdata", interfaces.RateLimitSubjectAPIKey, "subject-123", time.Minute)
	if err != nil {
		t.Fatalf("buildRateLimitKey() error = %v", err)
	}
	downloadKey, err := buildRateLimitKey("softdata", interfaces.RateLimitSubjectDownload, "subject-123", time.Minute)
	if err != nil {
		t.Fatalf("buildRateLimitKey() error = %v", err)
	}

	if anonymousKey == apiKey || anonymousKey == downloadKey || apiKey == downloadKey {
		t.Fatal("subject kinds unexpectedly collided")
	}
}

func TestWindowDurationsDoNotCollide(t *testing.T) {
	first, err := buildRateLimitKey("softdata", interfaces.RateLimitSubjectAPIKey, "subject-123", time.Minute)
	if err != nil {
		t.Fatalf("buildRateLimitKey() error = %v", err)
	}
	second, err := buildRateLimitKey("softdata", interfaces.RateLimitSubjectAPIKey, "subject-123", 10*time.Minute)
	if err != nil {
		t.Fatalf("buildRateLimitKey() error = %v", err)
	}
	if first == second {
		t.Fatal("different windows unexpectedly collided")
	}
}

func TestAllowIsRaceSafeWithConcurrentCalls(t *testing.T) {
	fake := &evalFake{result: []any{int64(1), int64(1), int64(60000), int64(1787756400), int64(0)}}
	repo, err := NewRateLimitRepository(fake, "softdata")
	if err != nil {
		t.Fatalf("NewRateLimitRepository() error = %v", err)
	}

	const calls = 32
	var wg sync.WaitGroup
	for i := 0; i < calls; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = repo.Allow(context.Background(), interfaces.RateLimitRequest{
				SubjectKind: interfaces.RateLimitSubjectAPIKey,
				Subject:     "api-key-123",
				Limit:       300,
				Window:      time.Minute,
			})
		}()
	}
	wg.Wait()
}
