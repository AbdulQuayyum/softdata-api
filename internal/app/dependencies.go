package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/config"
	"github.com/AbdulQuayyum/softdata-api/internal/database"
	"github.com/AbdulQuayyum/softdata-api/internal/handlers"
	"github.com/AbdulQuayyum/softdata-api/internal/middlewares"
	redisclient "github.com/AbdulQuayyum/softdata-api/internal/redis"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	postgresrepo "github.com/AbdulQuayyum/softdata-api/internal/repository/postgres"
	redisrepo "github.com/AbdulQuayyum/softdata-api/internal/repository/redis"
	"github.com/AbdulQuayyum/softdata-api/internal/router"
	"github.com/AbdulQuayyum/softdata-api/internal/security"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

const (
	defaultKeyPrefix = "softdata"
)

type unavailableRateLimitRepository struct{}

func (unavailableRateLimitRepository) Allow(ctx context.Context, request interfaces.RateLimitRequest) (interfaces.RateLimitResult, error) {
	if err := ctx.Err(); err != nil {
		return interfaces.RateLimitResult{}, err
	}
	return interfaces.RateLimitResult{}, interfaces.ErrRateLimitUnavailable
}

func buildDependencies(ctx context.Context, cfg *config.Config, logger *slog.Logger) (deps appDependencies, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if cfg == nil {
		return appDependencies{}, fmt.Errorf("config is required")
	}
	if logger == nil {
		return appDependencies{}, fmt.Errorf("logger is required")
	}

	cleanup := make([]func(), 0, 2)
	defer func() {
		if err == nil {
			return
		}
		for i := len(cleanup) - 1; i >= 0; i-- {
			if cleanup[i] != nil {
				cleanup[i]()
			}
		}
	}()

	pool, err := database.NewPostgres(ctx, cfg.Database)
	if err != nil {
		return appDependencies{}, fmt.Errorf("initialize postgres: %w", err)
	}
	cleanup = append(cleanup, pool.Close)

	redisRepo, redisClose, err := buildRateLimitRepository(ctx, cfg, logger)
	if err != nil {
		return appDependencies{}, err
	}
	if redisClose != nil {
		cleanup = append(cleanup, func() {
			_ = redisClose()
		})
	}

	accountRepo := postgresrepo.NewAccountRepository(pool)
	sessionRepo := postgresrepo.NewSessionRepository(pool)
	apiKeyRepo := postgresrepo.NewAPIKeyRepository(pool)
	usageRepo := postgresrepo.NewUsageRepository(pool)
	datasetRepo := postgresrepo.NewDatasetRepository(pool)

	passwordHasher := services.NewSecurityPasswordHasher()
	refreshTokens := services.NewSecurityRefreshTokenGenerator()
	accessTokens, err := services.NewSecurityAccessTokenIssuer(cfg.Security.AuthTokenSecret)
	if err != nil {
		return appDependencies{}, err
	}
	clock := services.NewSystemClock()

	accountService, err := services.NewAccountService(accountRepo, passwordHasher)
	if err != nil {
		return appDependencies{}, err
	}
	authService, err := services.NewAuthService(
		accountRepo,
		sessionRepo,
		passwordHasher,
		accessTokens,
		refreshTokens,
		clock,
		cfg.Security.AccessTokenTTL,
		cfg.Security.RefreshTokenTTL,
	)
	if err != nil {
		return appDependencies{}, err
	}
	apiKeyService, err := services.NewAPIKeyService(apiKeyRepo, services.NewSecurityAPIKeyGenerator())
	if err != nil {
		return appDependencies{}, err
	}
	usageService, err := services.NewUsageService(usageRepo, apiKeyRepo, clock, cfg.Usage.APIKeyMonthlyAllowance)
	if err != nil {
		return appDependencies{}, err
	}
	datasetService, err := services.NewDatasetService(datasetRepo)
	if err != nil {
		return appDependencies{}, err
	}

	healthHandler := handlers.NewHealthHandler()
	discoveryHandler := handlers.NewDiscoveryHandler()
	authHandler, err := handlers.NewAuthHandler(accountService, authService)
	if err != nil {
		return appDependencies{}, err
	}
	accountHandler, err := handlers.NewAccountHandler(accountService)
	if err != nil {
		return appDependencies{}, err
	}
	apiKeyHandler, err := handlers.NewAPIKeyHandler(apiKeyService)
	if err != nil {
		return appDependencies{}, err
	}
	usageHandler, err := handlers.NewUsageHandler(usageService)
	if err != nil {
		return appDependencies{}, err
	}
	datasetHandler, err := handlers.NewDatasetHandler(datasetService)
	if err != nil {
		return appDependencies{}, err
	}

	requestIDMiddleware := middlewares.RequestID
	recoveryMiddleware := middlewares.Recovery()
	loggerMiddleware, err := middlewares.NewLogger(logger)
	if err != nil {
		return appDependencies{}, err
	}
	securityHeadersMiddleware, err := middlewares.NewSecurityHeaders(middlewares.SecurityHeadersOptions{
		EnableHSTS:            strings.EqualFold(cfg.Environment, string(config.AppEnvironmentProduction)),
		HSTSMaxAge:            365 * 24 * time.Hour,
		HSTSIncludeSubdomains: true,
	})
	if err != nil {
		return appDependencies{}, err
	}
	corsMiddleware, err := middlewares.NewCORS(middlewares.CORSOptions{
		AllowedOrigins: cfg.Server.AllowedOrigins,
	})
	if err != nil {
		return appDependencies{}, err
	}
	bodyLimitMiddleware, err := middlewares.NewBodyLimit(cfg.Server.MaxBodyBytes)
	if err != nil {
		return appDependencies{}, err
	}
	timeoutMiddleware, err := middlewares.NewTimeout(cfg.Server.RequestTimeout)
	if err != nil {
		return appDependencies{}, err
	}
	anonymousIdentifier, err := middlewares.NewSecurityAnonymousIdentifier(cfg.Security.AnonymousIDSecret)
	if err != nil {
		return appDependencies{}, err
	}
	optionalAPIKeyMiddleware := middlewares.OptionalAPIKey(apiKeyService)
	authenticationMiddleware := middlewares.Authentication(accessTokenVerifier{secret: cfg.Security.AuthTokenSecret})
	rateLimitMiddleware, err := middlewares.RateLimit(redisRepo, anonymousIdentifier, middlewares.RateLimitPolicy{
		AnonymousLimit: int64(cfg.RateLimit.AnonymousRequestLimit),
		APIKeyLimit:    int64(cfg.RateLimit.APIKeyRequestLimit),
		DownloadLimit:  int64(cfg.RateLimit.DatasetDownloadLimit),
		Window:         cfg.RateLimit.Window,
		FailOpen:       cfg.RateLimit.FailOpen,
	})
	if err != nil {
		return appDependencies{}, err
	}
	usageFactory := func(endpoint, datasetGroup string) (router.MiddlewareFunc, error) {
		return middlewares.UsageTracking(usageService, endpoint, datasetGroup, middlewares.UsageTrackingOptions{
			Timeout:             cfg.Server.RequestTimeout,
			Now:                 clock.Now,
			AnonymousIdentifier: anonymousIdentifier,
		})
	}

	routerHandler, err := router.New(router.Handlers{
		Health:    healthHandler,
		Discovery: discoveryHandler,
		Auth:      authHandler,
		Account:   accountHandler,
		APIKey:    apiKeyHandler,
		Usage:     usageHandler,
		Dataset:   datasetHandler,
	}, router.Middleware{
		RequestID:       requestIDMiddleware,
		Recovery:        recoveryMiddleware,
		Logger:          loggerMiddleware,
		SecurityHeaders: securityHeadersMiddleware,
		CORS:            corsMiddleware,
		BodyLimit:       bodyLimitMiddleware,
		Timeout:         timeoutMiddleware,
		Authentication:  authenticationMiddleware,
		OptionalAPIKey:  optionalAPIKeyMiddleware,
		StandardLimit:   rateLimitMiddleware,
		UsageTracking:   usageFactory,
	})
	if err != nil {
		return appDependencies{}, err
	}

	server := &http.Server{
		Addr:              cfg.Server.ListenAddress(),
		Handler:           routerHandler,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
	}

	deps = appDependencies{
		server: server,
		runServer: func() error {
			return server.ListenAndServe()
		},
		shutdownServer: func(ctx context.Context) error {
			if err := server.Shutdown(ctx); err != nil {
				_ = server.Close()
				return err
			}
			return nil
		},
		closeRedis:    redisClose,
		closePostgres: pool.Close,
	}
	return deps, nil
}

func buildRateLimitRepository(ctx context.Context, cfg *config.Config, logger *slog.Logger) (interfaces.RateLimitRepository, func() error, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if cfg == nil {
		return nil, nil, fmt.Errorf("config is required")
	}

	prefix := strings.TrimSpace(cfg.Redis.KeyPrefix)
	if prefix == "" {
		prefix = defaultKeyPrefix
	}

	client, err := redisclient.NewClient(cfg.Redis)
	if err != nil {
		if cfg.RateLimit.FailOpen {
			if logger != nil {
				logger.Warn("redis unavailable at startup; continuing with fail-open rate limiting")
			}
			return unavailableRateLimitRepository{}, nil, nil
		}
		return nil, nil, fmt.Errorf("open redis client: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, redisStartupTimeout(cfg.Redis))
	defer cancel()

	if err := redisclient.Ping(pingCtx, client); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			_ = client.Close()
			return nil, nil, err
		}
		if cfg.RateLimit.FailOpen {
			if logger != nil {
				logger.Warn("redis unavailable at startup; continuing with fail-open rate limiting")
			}
		} else {
			_ = client.Close()
			return nil, nil, fmt.Errorf("ping redis: %w", err)
		}
	}

	repo, err := redisrepo.NewRateLimitRepository(client, prefix)
	if err != nil {
		_ = client.Close()
		return nil, nil, fmt.Errorf("create rate limit repository: %w", err)
	}

	return repo, client.Close, nil
}

func redisStartupTimeout(cfg config.RedisConfig) time.Duration {
	timeout := cfg.DialTimeout
	if cfg.ReadTimeout > timeout {
		timeout = cfg.ReadTimeout
	}
	if cfg.WriteTimeout > timeout {
		timeout = cfg.WriteTimeout
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return timeout
}

type accessTokenVerifier struct {
	secret string
}

func (v accessTokenVerifier) ValidateAccessToken(token string) (*security.AccessTokenClaims, error) {
	return security.ValidateAccessToken(token, v.secret)
}
