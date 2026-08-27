package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/app"
	"github.com/AbdulQuayyum/softdata-api/internal/config"
)

const startupTimeout = 30 * time.Second

func main() {
	logger := newLogger()
	if err := run(logger); err != nil {
		logger.Error("softdata api exited", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	if logger == nil {
		logger = newLogger()
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	startupCtx, cancel := context.WithTimeout(rootCtx, startupTimeout)
	defer cancel()

	application, err := app.New(startupCtx, cfg, logger)
	if err != nil {
		return err
	}

	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- application.Run()
	}()

	select {
	case err := <-runErrCh:
		if err == nil {
			return nil
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()
		_ = application.Shutdown(shutdownCtx)
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-rootCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()
		if err := application.Shutdown(shutdownCtx); err != nil {
			return err
		}
		if err := <-runErrCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
