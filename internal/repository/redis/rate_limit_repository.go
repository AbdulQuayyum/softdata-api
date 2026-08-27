package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	redis "github.com/redis/go-redis/v9"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const rateLimitKeyPrefix = "softdata:ratelimit:v1"

var _ interfaces.RateLimitRepository = (*RateLimitRepository)(nil)

type redisEvaluator interface {
	Eval(ctx context.Context, script string, keys []string, args ...any) *redis.Cmd
}

type RateLimitRepository struct {
	client redisEvaluator
	prefix string
}

func NewRateLimitRepository(client redisEvaluator, prefix string) (*RateLimitRepository, error) {
	if client == nil {
		return nil, fmt.Errorf("rate limit client is required")
	}
	prefix = strings.TrimSpace(prefix)
	if !isSafeKeySegment(prefix) {
		return nil, fmt.Errorf("rate limit prefix is required")
	}
	return &RateLimitRepository{client: client, prefix: prefix}, nil
}

func (r *RateLimitRepository) Allow(ctx context.Context, request interfaces.RateLimitRequest) (interfaces.RateLimitResult, error) {
	if err := ctx.Err(); err != nil {
		return interfaces.RateLimitResult{}, err
	}
	if err := validateRequest(request); err != nil {
		return interfaces.RateLimitResult{}, err
	}

	key, err := buildRateLimitKey(r.prefix, request.SubjectKind, request.Subject, request.Window)
	if err != nil {
		return interfaces.RateLimitResult{}, err
	}

	windowMs, err := durationToMilliseconds(request.Window)
	if err != nil {
		return interfaces.RateLimitResult{}, err
	}

	cmd := r.client.Eval(ctx, rateLimitScript, []string{key}, request.Limit, windowMs)
	value, err := cmd.Result()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return interfaces.RateLimitResult{}, err
		}
		return interfaces.RateLimitResult{}, fmt.Errorf("%w", interfaces.ErrRateLimitUnavailable)
	}

	allowed, count, ttlMs, nowUnix, nowMicro, err := parseRateLimitEvalResult(value)
	if err != nil {
		return interfaces.RateLimitResult{}, fmt.Errorf("%w", interfaces.ErrRateLimitUnavailable)
	}
	resetAt, err := resetTimeFromRedisTime(nowUnix, nowMicro, ttlMs)
	if err != nil {
		return interfaces.RateLimitResult{}, fmt.Errorf("%w", interfaces.ErrRateLimitUnavailable)
	}

	remaining := request.Limit - count
	if remaining < 0 {
		remaining = 0
	}

	return interfaces.RateLimitResult{
		Allowed:   allowed,
		Limit:     request.Limit,
		Remaining: remaining,
		ResetAt:   resetAt,
	}, nil
}

const rateLimitScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])

local current = redis.call("INCR", key)
if current == 1 then
  redis.call("PEXPIRE", key, window_ms)
end

local ttl = redis.call("PTTL", key)
if ttl < 0 then
  return redis.error_reply("rate limit counter has invalid ttl")
end

local now = redis.call("TIME")
local allowed = 0
if current <= limit then
  allowed = 1
end

return {allowed, current, ttl, now[1], now[2]}
`

func validateRequest(request interfaces.RateLimitRequest) error {
	if request.SubjectKind != interfaces.RateLimitSubjectAnonymous && request.SubjectKind != interfaces.RateLimitSubjectAPIKey && request.SubjectKind != interfaces.RateLimitSubjectDownload {
		return fmt.Errorf("%w", interfaces.ErrInvalidRateLimitInput)
	}
	if !isSafeSubject(request.Subject) {
		return fmt.Errorf("%w", interfaces.ErrInvalidRateLimitInput)
	}
	if request.Limit <= 0 {
		return fmt.Errorf("%w", interfaces.ErrInvalidRateLimitInput)
	}
	if request.Window <= 0 {
		return fmt.Errorf("%w", interfaces.ErrInvalidRateLimitInput)
	}
	if _, err := durationToMilliseconds(request.Window); err != nil {
		return err
	}
	return nil
}

func buildRateLimitKey(prefix string, subjectKind interfaces.RateLimitSubjectKind, subject string, window time.Duration) (string, error) {
	if !isSafeKeySegment(prefix) {
		return "", fmt.Errorf("%w", interfaces.ErrInvalidRateLimitInput)
	}
	if !isSafeSubject(subject) {
		return "", fmt.Errorf("%w", interfaces.ErrInvalidRateLimitInput)
	}
	if subjectKind != interfaces.RateLimitSubjectAnonymous && subjectKind != interfaces.RateLimitSubjectAPIKey && subjectKind != interfaces.RateLimitSubjectDownload {
		return "", fmt.Errorf("%w", interfaces.ErrInvalidRateLimitInput)
	}
	windowMs, err := durationToMilliseconds(window)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(subject))
	return strings.Join([]string{
		prefix,
		"ratelimit",
		"v1",
		string(subjectKind),
		strconv.FormatInt(windowMs, 10),
		hex.EncodeToString(hash[:]),
	}, ":"), nil
}

func parseRateLimitEvalResult(value any) (allowed bool, count int64, ttlMs int64, nowUnix int64, nowMicro int64, err error) {
	items, ok := value.([]any)
	if !ok {
		return false, 0, 0, 0, 0, fmt.Errorf("unexpected rate limit script result")
	}
	if len(items) != 5 {
		return false, 0, 0, 0, 0, fmt.Errorf("unexpected rate limit script result")
	}

	allowedInt, err := toInt64(items[0])
	if err != nil {
		return false, 0, 0, 0, 0, err
	}
	count, err = toInt64(items[1])
	if err != nil {
		return false, 0, 0, 0, 0, err
	}
	ttlMs, err = toInt64(items[2])
	if err != nil {
		return false, 0, 0, 0, 0, err
	}
	nowUnix, err = toInt64(items[3])
	if err != nil {
		return false, 0, 0, 0, 0, err
	}
	nowMicro, err = toInt64(items[4])
	if err != nil {
		return false, 0, 0, 0, 0, err
	}

	return allowedInt != 0, count, ttlMs, nowUnix, nowMicro, nil
}

func resetTimeFromRedisTime(unixSeconds, microseconds, ttlMs int64) (time.Time, error) {
	if ttlMs < 0 {
		return time.Time{}, fmt.Errorf("%w", interfaces.ErrRateLimitUnavailable)
	}
	if microseconds < 0 || microseconds >= int64(time.Second/time.Microsecond) {
		return time.Time{}, fmt.Errorf("%w", interfaces.ErrRateLimitUnavailable)
	}
	now := time.Unix(unixSeconds, microseconds*int64(time.Microsecond)).UTC()
	return now.Add(time.Duration(ttlMs) * time.Millisecond), nil
}

func durationToMilliseconds(window time.Duration) (int64, error) {
	if window <= 0 {
		return 0, fmt.Errorf("%w", interfaces.ErrInvalidRateLimitInput)
	}
	if window > maxRedisDuration() {
		return 0, fmt.Errorf("%w", interfaces.ErrInvalidRateLimitInput)
	}
	return int64(window / time.Millisecond), nil
}

func maxRedisDuration() time.Duration {
	return time.Duration(math.MaxInt64/int64(time.Millisecond)) * time.Millisecond
}

func isSafeKeySegment(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return false
		}
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

func isSafeSubject(value string) bool {
	if value == "" || len(value) > 256 {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func toInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("unexpected rate limit script result type %T", value)
	}
}
