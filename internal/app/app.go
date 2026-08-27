package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/AbdulQuayyum/softdata-api/internal/config"
)

// App owns the production HTTP server and the resources it depends on.
type App struct {
	cfg             *config.Config
	logger          *slog.Logger
	server          *http.Server
	runServer       func() error
	shutdownServer  func(context.Context) error
	closeRedis      func() error
	closePostgres   func()
	shutdownOnce    sync.Once
	shutdownErr     error
	shutdownStarted atomic.Bool
}

type appDependencies struct {
	server         *http.Server
	runServer      func() error
	shutdownServer func(context.Context) error
	closeRedis     func() error
	closePostgres  func()
}

// New builds the production app from configuration and a logger.
func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*App, error) {
	deps, err := buildDependencies(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}
	return newApp(cfg, logger, deps)
}

func newApp(cfg *config.Config, logger *slog.Logger, deps appDependencies) (*App, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if deps.runServer == nil {
		return nil, fmt.Errorf("server runner is required")
	}
	if deps.shutdownServer == nil {
		return nil, fmt.Errorf("server shutdown function is required")
	}

	return &App{
		cfg:            cfg,
		logger:         logger,
		server:         deps.server,
		runServer:      deps.runServer,
		shutdownServer: deps.shutdownServer,
		closeRedis:     deps.closeRedis,
		closePostgres:  deps.closePostgres,
	}, nil
}

// Run starts the configured HTTP server and waits for it to stop.
func (a *App) Run() error {
	if a == nil || a.runServer == nil {
		return nil
	}
	if a.logger != nil && a.server != nil {
		a.logger.Info("starting HTTP server", "addr", a.server.Addr)
	}

	err := a.runServer()
	if errors.Is(err, http.ErrServerClosed) && a.shutdownStarted.Load() {
		return nil
	}
	return err
}
