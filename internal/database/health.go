package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const readinessTimeout = 3 * time.Second

func Ready(ctx context.Context, pool *pgxpool.Pool) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	if pool == nil {
		return fmt.Errorf("%w", ErrDatabaseUnavailable)
	}

	readyCtx, cancel := context.WithTimeout(ctx, readinessTimeout)
	defer cancel()

	if err := pool.Ping(readyCtx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("%w", ErrDatabaseUnavailable)
	}

	return nil
}
