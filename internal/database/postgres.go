package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	startupTimeout = 10 * time.Second
	pingTimeout    = 5 * time.Second
)

func NewPostgres(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if cfg.URL == nil || strings.TrimSpace(*cfg.URL) == "" {
		return nil, fmt.Errorf("database url is required")
	}

	poolConfig, err := pgxpool.ParseConfig(*cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: invalid DATABASE_URL")
	}

	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MinConnections)
	poolConfig.MaxConnLifetime = cfg.MaxConnectionLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnectionIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	startupCtx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(startupCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: initialization failed")
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, pingTimeout)
	defer pingCancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("create postgres pool: ping failed")
	}

	return pool, nil
}
