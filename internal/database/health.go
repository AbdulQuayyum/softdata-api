package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const readinessTimeout = 3 * time.Second

func Ready(ctx context.Context, pool *pgxpool.Pool) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if pool == nil {
		return fmt.Errorf("database is unavailable")
	}

	readyCtx, cancel := context.WithTimeout(ctx, readinessTimeout)
	defer cancel()

	if err := pool.Ping(readyCtx); err != nil {
		return fmt.Errorf("database is unavailable")
	}

	return nil
}
