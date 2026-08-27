package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/config"
	redisv9 "github.com/redis/go-redis/v9"
)

var (
	// ErrInvalidRedisConfig reports a malformed or incomplete Redis config.
	ErrInvalidRedisConfig = errors.New("redis: invalid configuration")
	// ErrRedisUnavailable reports that Redis could not be reached during startup.
	ErrRedisUnavailable = errors.New("redis: unavailable")
)

type redisClient interface {
	Ping(context.Context) *redisv9.StatusCmd
	Close() error
}

func OpenRedis(ctx context.Context, cfg config.RedisConfig) (*redisv9.Client, error) {
	client, err := openRedis(ctx, cfg, func(options *redisv9.Options) redisClient {
		return redisv9.NewClient(options)
	})
	if err != nil {
		return nil, err
	}

	redisClient, ok := client.(*redisv9.Client)
	if !ok {
		return nil, fmt.Errorf("%w", ErrInvalidRedisConfig)
	}

	return redisClient, nil
}

func openRedis(ctx context.Context, cfg config.RedisConfig, newClient func(*redisv9.Options) redisClient) (redisClient, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	options, err := buildRedisOptions(cfg)
	if err != nil {
		return nil, err
	}

	client := newClient(options)
	pingCtx, cancel := context.WithTimeout(ctx, startupRedisTimeout(options))
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("%w", ErrRedisUnavailable)
	}

	return client, nil
}

func buildRedisOptions(cfg config.RedisConfig) (*redisv9.Options, error) {
	if cfg.DB < 0 || cfg.DialTimeout <= 0 || cfg.ReadTimeout <= 0 || cfg.WriteTimeout <= 0 || cfg.PoolSize <= 0 {
		return nil, fmt.Errorf("%w", ErrInvalidRedisConfig)
	}

	options, err := redisOptionsFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	options.DialTimeout = cfg.DialTimeout
	options.ReadTimeout = cfg.ReadTimeout
	options.WriteTimeout = cfg.WriteTimeout
	options.PoolSize = cfg.PoolSize

	return options, nil
}

func redisOptionsFromConfig(cfg config.RedisConfig) (*redisv9.Options, error) {
	if cfg.URL != nil && strings.TrimSpace(*cfg.URL) != "" {
		options, err := redisv9.ParseURL(strings.TrimSpace(*cfg.URL))
		if err != nil {
			return nil, fmt.Errorf("%w", ErrInvalidRedisConfig)
		}
		return options, nil
	}

	if strings.TrimSpace(cfg.Address) == "" {
		return nil, fmt.Errorf("%w", ErrInvalidRedisConfig)
	}

	return &redisv9.Options{
		Addr:         strings.TrimSpace(cfg.Address),
		Username:     derefString(cfg.Username),
		Password:     derefString(cfg.Password),
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
	}, nil
}

func startupRedisTimeout(options *redisv9.Options) time.Duration {
	timeout := options.DialTimeout
	if options.ReadTimeout > timeout {
		timeout = options.ReadTimeout
	}
	if options.WriteTimeout > timeout {
		timeout = options.WriteTimeout
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return timeout
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
