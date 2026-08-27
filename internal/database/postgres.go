package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrInvalidDatabaseConfig reports a malformed or incomplete PostgreSQL config.
	ErrInvalidDatabaseConfig = errors.New("database: invalid configuration")
	// ErrDatabaseUnavailable reports that PostgreSQL could not be reached during startup.
	ErrDatabaseUnavailable = errors.New("database: unavailable")
)

type postgresPool interface {
	Ping(context.Context) error
	Close()
}

func NewPostgres(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	pool, err := openPostgres(ctx, cfg, func(ctx context.Context, poolConfig *pgxpool.Config) (postgresPool, error) {
		return pgxpool.NewWithConfig(ctx, poolConfig)
	})
	if err != nil {
		return nil, err
	}

	pgPool, ok := pool.(*pgxpool.Pool)
	if !ok {
		return nil, fmt.Errorf("%w", ErrInvalidDatabaseConfig)
	}

	return pgPool, nil
}

func openPostgres(ctx context.Context, cfg config.DatabaseConfig, opener func(context.Context, *pgxpool.Config) (postgresPool, error)) (postgresPool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	poolConfig, err := buildPostgresPoolConfig(cfg)
	if err != nil {
		return nil, err
	}

	connectCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	pool, err := opener(connectCtx, poolConfig)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("%w", ErrDatabaseUnavailable)
	}

	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("%w", ErrDatabaseUnavailable)
	}

	return pool, nil
}

func buildPostgresPoolConfig(cfg config.DatabaseConfig) (*pgxpool.Config, error) {
	if cfg.URL == nil || strings.TrimSpace(*cfg.URL) == "" {
		return nil, fmt.Errorf("%w", ErrInvalidDatabaseConfig)
	}
	if cfg.MaxConnections <= 0 || cfg.MinConnections <= 0 || cfg.MinConnections > cfg.MaxConnections {
		return nil, fmt.Errorf("%w", ErrInvalidDatabaseConfig)
	}
	if cfg.MaxConnectionLifetime <= 0 || cfg.MaxConnectionIdleTime <= 0 || cfg.HealthCheckPeriod <= 0 || cfg.ConnectTimeout <= 0 {
		return nil, fmt.Errorf("%w", ErrInvalidDatabaseConfig)
	}

	poolConfig, err := pgxpool.ParseConfig(strings.TrimSpace(*cfg.URL))
	if err != nil {
		return nil, fmt.Errorf("%w", ErrInvalidDatabaseConfig)
	}

	poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MinConnections)
	poolConfig.MaxConnLifetime = cfg.MaxConnectionLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnectionIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	return poolConfig, nil
}
