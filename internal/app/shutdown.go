package app

import "context"

// Shutdown stops the HTTP server first, then closes Redis and PostgreSQL resources.
func (a *App) Shutdown(ctx context.Context) error {
	if a == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var shutdownErr error
	a.shutdownOnce.Do(func() {
		a.shutdownStarted.Store(true)

		if a.shutdownServer != nil {
			shutdownErr = a.shutdownServer(ctx)
		}
		if a.closeRedis != nil {
			if err := a.closeRedis(); err != nil && shutdownErr == nil {
				shutdownErr = err
			}
		}
		if a.closePostgres != nil {
			a.closePostgres()
		}
		a.shutdownErr = shutdownErr
	})

	return a.shutdownErr
}
