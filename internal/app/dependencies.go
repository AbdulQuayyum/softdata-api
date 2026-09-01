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
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	redisclient "github.com/AbdulQuayyum/softdata-api/internal/redis"
	fileRepo "github.com/AbdulQuayyum/softdata-api/internal/repository/file"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	postgresrepo "github.com/AbdulQuayyum/softdata-api/internal/repository/postgres"
	redisrepo "github.com/AbdulQuayyum/softdata-api/internal/repository/redis"
	"github.com/AbdulQuayyum/softdata-api/internal/router"
	"github.com/AbdulQuayyum/softdata-api/internal/security"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
	redisv9 "github.com/redis/go-redis/v9"
)

const (
	defaultKeyPrefix                                       = "softdata"
	geographyStatesRelativePath                            = "geography/states.json"
	geographyGeopoliticalZonesRelativePath                 = "geography/geopolitical_zones.json"
	geographyLocalGovernmentUnitsRelativePath              = "geography/lgas.json"
	geographyCountriesAndAreasRelativePath                 = "geography/countries_and_areas.json"
	educationUniversitiesRelativePath                      = "education/universities.json"
	educationCollegesOfEducationRelativePath               = "education/colleges_of_education.json"
	financePaymentServiceProvidersRelativePath             = "finance/payment_service_providers.json"
	financeInternationalMoneyTransferOperatorsRelativePath = "finance/international_money_transfer_operators.json"
)

var approvedUniversityStateIDs = map[string]struct{}{
	"abia": {}, "adamawa": {}, "akwa-ibom": {}, "anambra": {}, "bauchi": {},
	"bayelsa": {}, "benue": {}, "borno": {}, "cross-river": {}, "delta": {},
	"ebonyi": {}, "edo": {}, "ekiti": {}, "enugu": {}, "fct": {}, "gombe": {},
	"imo": {}, "jigawa": {}, "kaduna": {}, "kano": {}, "katsina": {},
	"kebbi": {}, "kogi": {}, "kwara": {}, "lagos": {}, "nasarawa": {},
	"niger": {}, "ogun": {}, "ondo": {}, "osun": {}, "oyo": {}, "plateau": {},
	"rivers": {}, "sokoto": {}, "taraba": {}, "yobe": {}, "zamfara": {},
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
		runStartupCleanup(&err, cleanup)
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
	jsonRepository, err := fileRepo.NewJSONRepository(cfg.Datasets.Path, cfg.Datasets.JSONMaxBytes)
	if err != nil {
		return appDependencies{}, fmt.Errorf("initialize json repository: %w", err)
	}

	geographyHandler, err := buildGeographyHandlerFromJSONRepository(ctx, jsonRepository,
		func(repository interfaces.JSONFileRepository, statesPath, zonesPath, localGovernmentUnitsPath, countriesAndAreasPath string) (interfaces.GeographyRepository, error) {
			return fileRepo.NewGeographyRepository(repository, statesPath, zonesPath, localGovernmentUnitsPath, countriesAndAreasPath)
		},
		func(repository interfaces.GeographyRepository) (geographyService, error) {
			return services.NewGeographyService(repository)
		},
		func(service geographyService) (*handlers.GeographyHandler, error) {
			return handlers.NewGeographyHandler(service)
		},
	)
	if err != nil {
		return appDependencies{}, err
	}
	educationHandler, err := buildEducationHandlerFromJSONRepository(ctx, jsonRepository,
		func(repository interfaces.JSONFileRepository, universitiesPath, collegesOfEducationPath string) (interfaces.EducationRepository, error) {
			return fileRepo.NewEducationRepository(repository, universitiesPath, collegesOfEducationPath)
		},
		func(repository interfaces.EducationRepository) (educationService, error) {
			return services.NewEducationService(repository)
		},
		func(service educationService) (*handlers.EducationHandler, error) {
			return handlers.NewEducationHandler(service)
		},
	)
	if err != nil {
		return appDependencies{}, err
	}
	financeHandler, err := buildFinanceHandlerFromJSONRepository(ctx, jsonRepository,
		func(repository interfaces.JSONFileRepository, paymentServiceProvidersPath string) (interfaces.FinanceRepository, error) {
			return fileRepo.NewFinanceRepository(repository, paymentServiceProvidersPath, financeInternationalMoneyTransferOperatorsRelativePath)
		},
		func(repository interfaces.FinanceRepository) (financeService, error) {
			return services.NewFinanceService(repository)
		},
		func(service financeService) (*handlers.FinanceHandler, error) {
			return handlers.NewFinanceHandler(service)
		},
	)
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
		Geography: geographyHandler,
		Education: educationHandler,
		Finance:   financeHandler,
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

func runStartupCleanup(err *error, cleanup []func()) {
	if err == nil || *err == nil {
		return
	}
	for i := len(cleanup) - 1; i >= 0; i-- {
		if cleanup[i] != nil {
			cleanup[i]()
		}
	}
}

type geographyService interface {
	ListStates(context.Context) ([]models.State, error)
	GetState(context.Context, string) (models.State, error)
	ListGeopoliticalZones(context.Context) ([]models.GeopoliticalZone, error)
	GetGeopoliticalZone(context.Context, string) (models.GeopoliticalZone, error)
	ListLocalGovernmentUnits(context.Context) ([]models.LocalGovernmentUnit, error)
	ListLocalGovernmentUnitsByState(context.Context, string) ([]models.LocalGovernmentUnit, error)
	GetLocalGovernmentUnit(context.Context, string) (models.LocalGovernmentUnit, error)
	ListCountriesAndAreas(context.Context, services.CountryOrAreaListInput) ([]models.CountryOrArea, error)
	GetCountryOrArea(context.Context, string) (models.CountryOrArea, error)
}

type educationService interface {
	ListUniversities(context.Context, services.UniversityListInput) ([]models.University, error)
	GetUniversity(context.Context, string) (models.University, error)
	ListCollegesOfEducation(context.Context, services.CollegeOfEducationListInput) ([]models.CollegeOfEducation, error)
	GetCollegeOfEducation(context.Context, string) (models.CollegeOfEducation, error)
}

type financeService interface {
	ListPaymentServiceProviders(context.Context) ([]models.PaymentServiceProvider, error)
	ListPaymentServiceProvidersByType(context.Context, string) ([]models.PaymentServiceProvider, error)
	GetPaymentServiceProvider(context.Context, string) (models.PaymentServiceProvider, error)
	ListInternationalMoneyTransferOperators(context.Context) ([]models.InternationalMoneyTransferOperator, error)
	GetInternationalMoneyTransferOperator(context.Context, string) (models.InternationalMoneyTransferOperator, error)
	ListCurrencies(context.Context, services.CurrencyListInput) ([]models.Currency, error)
	GetCurrency(context.Context, string) (models.Currency, error)
}

func buildGeographyHandler(
	ctx context.Context,
	cfg *config.Config,
	newJSONRepository func(string, int64) (interfaces.JSONFileRepository, error),
	newGeographyRepository func(interfaces.JSONFileRepository, string, string, string, string) (interfaces.GeographyRepository, error),
	newGeographyService func(interfaces.GeographyRepository) (geographyService, error),
	newGeographyHandler func(geographyService) (*handlers.GeographyHandler, error),
) (*handlers.GeographyHandler, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if newJSONRepository == nil {
		return nil, fmt.Errorf("json repository factory is required")
	}
	if newGeographyRepository == nil {
		return nil, fmt.Errorf("geography repository factory is required")
	}
	if newGeographyService == nil {
		return nil, fmt.Errorf("geography service factory is required")
	}
	if newGeographyHandler == nil {
		return nil, fmt.Errorf("geography handler factory is required")
	}

	jsonRepository, err := newJSONRepository(cfg.Datasets.Path, cfg.Datasets.JSONMaxBytes)
	if err != nil {
		return nil, fmt.Errorf("initialize geography json repository: %w", err)
	}
	return buildGeographyHandlerFromJSONRepository(ctx, jsonRepository, newGeographyRepository, newGeographyService, newGeographyHandler)
}

func buildGeographyHandlerFromJSONRepository(
	ctx context.Context,
	jsonRepository interfaces.JSONFileRepository,
	newGeographyRepository func(interfaces.JSONFileRepository, string, string, string, string) (interfaces.GeographyRepository, error),
	newGeographyService func(interfaces.GeographyRepository) (geographyService, error),
	newGeographyHandler func(geographyService) (*handlers.GeographyHandler, error),
) (*handlers.GeographyHandler, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	if newGeographyRepository == nil {
		return nil, fmt.Errorf("geography repository factory is required")
	}
	if newGeographyService == nil {
		return nil, fmt.Errorf("geography service factory is required")
	}
	if newGeographyHandler == nil {
		return nil, fmt.Errorf("geography handler factory is required")
	}

	geographyRepository, err := newGeographyRepository(jsonRepository, geographyStatesRelativePath, geographyGeopoliticalZonesRelativePath, geographyLocalGovernmentUnitsRelativePath, geographyCountriesAndAreasRelativePath)
	if err != nil {
		return nil, fmt.Errorf("initialize geography repository: %w", err)
	}
	geographyService, err := newGeographyService(geographyRepository)
	if err != nil {
		return nil, fmt.Errorf("initialize geography service: %w", err)
	}
	if err := verifyGeographyDataset(ctx, geographyService); err != nil {
		return nil, err
	}
	geographyHandler, err := newGeographyHandler(geographyService)
	if err != nil {
		return nil, fmt.Errorf("initialize geography handler: %w", err)
	}
	return geographyHandler, nil
}

func buildEducationHandler(
	ctx context.Context,
	cfg *config.Config,
	newJSONRepository func(string, int64) (interfaces.JSONFileRepository, error),
	newEducationRepository func(interfaces.JSONFileRepository, string, string) (interfaces.EducationRepository, error),
	newEducationService func(interfaces.EducationRepository) (educationService, error),
	newEducationHandler func(educationService) (*handlers.EducationHandler, error),
) (*handlers.EducationHandler, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if newJSONRepository == nil {
		return nil, fmt.Errorf("json repository factory is required")
	}
	if newEducationRepository == nil {
		return nil, fmt.Errorf("education repository factory is required")
	}
	if newEducationService == nil {
		return nil, fmt.Errorf("education service factory is required")
	}
	if newEducationHandler == nil {
		return nil, fmt.Errorf("education handler factory is required")
	}

	jsonRepository, err := newJSONRepository(cfg.Datasets.Path, cfg.Datasets.JSONMaxBytes)
	if err != nil {
		return nil, fmt.Errorf("initialize education json repository: %w", err)
	}
	return buildEducationHandlerFromJSONRepository(ctx, jsonRepository, newEducationRepository, newEducationService, newEducationHandler)
}

func buildEducationHandlerFromJSONRepository(
	ctx context.Context,
	jsonRepository interfaces.JSONFileRepository,
	newEducationRepository func(interfaces.JSONFileRepository, string, string) (interfaces.EducationRepository, error),
	newEducationService func(interfaces.EducationRepository) (educationService, error),
	newEducationHandler func(educationService) (*handlers.EducationHandler, error),
) (*handlers.EducationHandler, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	if newEducationRepository == nil {
		return nil, fmt.Errorf("education repository factory is required")
	}
	if newEducationService == nil {
		return nil, fmt.Errorf("education service factory is required")
	}
	if newEducationHandler == nil {
		return nil, fmt.Errorf("education handler factory is required")
	}

	educationRepository, err := newEducationRepository(jsonRepository, educationUniversitiesRelativePath, educationCollegesOfEducationRelativePath)
	if err != nil {
		return nil, fmt.Errorf("initialize education repository: %w", err)
	}
	educationService, err := newEducationService(educationRepository)
	if err != nil {
		return nil, fmt.Errorf("initialize education service: %w", err)
	}
	if err := verifyEducationDataset(ctx, educationService); err != nil {
		return nil, err
	}
	if err := verifyCollegeOfEducationDataset(ctx, educationService); err != nil {
		return nil, err
	}
	educationHandler, err := newEducationHandler(educationService)
	if err != nil {
		return nil, fmt.Errorf("initialize education handler: %w", err)
	}
	return educationHandler, nil
}

func buildFinanceHandler(
	ctx context.Context,
	cfg *config.Config,
	newJSONRepository func(string, int64) (interfaces.JSONFileRepository, error),
	newFinanceRepository func(interfaces.JSONFileRepository, string) (interfaces.FinanceRepository, error),
	newFinanceService func(interfaces.FinanceRepository) (financeService, error),
	newFinanceHandler func(financeService) (*handlers.FinanceHandler, error),
) (*handlers.FinanceHandler, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if newJSONRepository == nil {
		return nil, fmt.Errorf("json repository factory is required")
	}
	if newFinanceRepository == nil {
		return nil, fmt.Errorf("finance repository factory is required")
	}
	if newFinanceService == nil {
		return nil, fmt.Errorf("finance service factory is required")
	}
	if newFinanceHandler == nil {
		return nil, fmt.Errorf("finance handler factory is required")
	}

	jsonRepository, err := newJSONRepository(cfg.Datasets.Path, cfg.Datasets.JSONMaxBytes)
	if err != nil {
		return nil, fmt.Errorf("initialize finance json repository: %w", err)
	}
	return buildFinanceHandlerFromJSONRepository(ctx, jsonRepository, newFinanceRepository, newFinanceService, newFinanceHandler)
}

func buildFinanceHandlerFromJSONRepository(
	ctx context.Context,
	jsonRepository interfaces.JSONFileRepository,
	newFinanceRepository func(interfaces.JSONFileRepository, string) (interfaces.FinanceRepository, error),
	newFinanceService func(interfaces.FinanceRepository) (financeService, error),
	newFinanceHandler func(financeService) (*handlers.FinanceHandler, error),
) (*handlers.FinanceHandler, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if jsonRepository == nil {
		return nil, fmt.Errorf("json repository is required")
	}
	if newFinanceRepository == nil {
		return nil, fmt.Errorf("finance repository factory is required")
	}
	if newFinanceService == nil {
		return nil, fmt.Errorf("finance service factory is required")
	}
	if newFinanceHandler == nil {
		return nil, fmt.Errorf("finance handler factory is required")
	}

	financeRepository, err := newFinanceRepository(jsonRepository, financePaymentServiceProvidersRelativePath)
	if err != nil {
		return nil, fmt.Errorf("initialize finance repository: %w", err)
	}
	financeService, err := newFinanceService(financeRepository)
	if err != nil {
		return nil, fmt.Errorf("initialize finance service: %w", err)
	}
	if err := verifyFinanceDataset(ctx, financeService); err != nil {
		return nil, err
	}
	if err := verifyCurrencyDataset(ctx, financeService); err != nil {
		return nil, err
	}
	financeHandler, err := newFinanceHandler(financeService)
	if err != nil {
		return nil, fmt.Errorf("initialize finance handler: %w", err)
	}
	return financeHandler, nil
}

func verifyGeographyDataset(ctx context.Context, service geographyService) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if service == nil {
		return fmt.Errorf("verify geography dataset: geography service is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	states, err := service.ListStates(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("verify geography dataset: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(states) != 37 {
		return fmt.Errorf("verify geography dataset: %w", interfaces.ErrInvalidDatasetFile)
	}

	stateCount := 0
	fctCount := 0
	for _, state := range states {
		switch state.AdministrativeType {
		case "federal_capital_territory":
			fctCount++
		default:
			stateCount++
		}
	}
	if stateCount != 36 || fctCount != 1 {
		return fmt.Errorf("verify geography dataset: %w", interfaces.ErrInvalidDatasetFile)
	}

	countries, err := service.ListCountriesAndAreas(ctx, services.CountryOrAreaListInput{})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("verify geography dataset: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateCountryOrAreasSnapshot(countries); err != nil {
		return fmt.Errorf("verify geography dataset: %w", interfaces.ErrInvalidDatasetFile)
	}
	return nil
}

func validateCountryOrAreasSnapshot(countries []models.CountryOrArea) error {
	if len(countries) != 248 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	seenIDs := make(map[string]struct{}, len(countries))
	seenAlpha2 := make(map[string]struct{}, len(countries))
	seenAlpha3 := make(map[string]struct{}, len(countries))
	seenNumeric := make(map[string]struct{}, len(countries))
	requiredNames := map[string]struct{}{
		"Holy See":           {},
		"State of Palestine": {},
		"Western Sahara":     {},
		"China, Hong Kong Special Administrative Region": {},
		"China, Macao Special Administrative Region":     {},
		"Antarctica": {},
	}
	foundRequiredNames := make(map[string]struct{}, len(requiredNames))
	foundNigeria := false
	foundAlgeria := false
	foundKosovo := false

	for _, country := range countries {
		if country.ID == "" || country.Name == "" || country.Alpha2Code == "" || country.Alpha3Code == "" || country.NumericCode == "" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !isLowerAlpha2Code(country.ID) || country.ID != strings.ToLower(country.Alpha2Code) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !isUpperAlpha2Code(country.Alpha2Code) || !isUpperAlpha3Code(country.Alpha3Code) || !isThreeDigitCode(country.NumericCode) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenIDs[country.ID]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenAlpha2[country.Alpha2Code]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenAlpha3[country.Alpha3Code]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenNumeric[country.NumericCode]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if (country.RegionCode == "") != (country.RegionName == "") {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if (country.SubregionCode == "") != (country.SubregionName == "") {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if (country.IntermediateRegionCode == "") != (country.IntermediateRegionName == "") {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if country.RegionCode != "" && !isThreeDigitCode(country.RegionCode) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if country.SubregionCode != "" && !isThreeDigitCode(country.SubregionCode) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if country.IntermediateRegionCode != "" && !isThreeDigitCode(country.IntermediateRegionCode) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}

		seenIDs[country.ID] = struct{}{}
		seenAlpha2[country.Alpha2Code] = struct{}{}
		seenAlpha3[country.Alpha3Code] = struct{}{}
		seenNumeric[country.NumericCode] = struct{}{}

		if country.ID == "ng" {
			if country.Name != "Nigeria" || country.Alpha2Code != "NG" || country.Alpha3Code != "NGA" || country.NumericCode != "566" {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			foundNigeria = true
		}
		if country.Name == "Algeria" {
			if country.NumericCode != "012" {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			foundAlgeria = true
		}
		if country.Name == "Kosovo" || country.ID == "xk" {
			foundKosovo = true
		}
		if _, ok := requiredNames[country.Name]; ok {
			foundRequiredNames[country.Name] = struct{}{}
		}
	}

	if !foundNigeria || !foundAlgeria || foundKosovo || len(foundRequiredNames) != len(requiredNames) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	return nil
}

func isLowerAlpha2Code(value string) bool {
	if len(value) != 2 {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] < 'a' || value[i] > 'z' {
			return false
		}
	}
	return true
}

func isUpperAlpha2Code(value string) bool {
	if len(value) != 2 {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] < 'A' || value[i] > 'Z' {
			return false
		}
	}
	return true
}

func isUpperAlpha3Code(value string) bool {
	if len(value) != 3 {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] < 'A' || value[i] > 'Z' {
			return false
		}
	}
	return true
}

func isThreeDigitCode(value string) bool {
	if len(value) != 3 {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}

func verifyEducationDataset(ctx context.Context, service educationService) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if service == nil {
		return fmt.Errorf("verify education dataset: education service is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	universities, err := service.ListUniversities(ctx, services.UniversityListInput{})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("verify education dataset: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(universities) != 328 {
		return fmt.Errorf("verify education dataset: %w", interfaces.ErrInvalidDatasetFile)
	}

	ownershipCounts := map[string]int{
		"federal": 0,
		"state":   0,
		"private": 0,
	}
	stateSeen := make(map[string]struct{}, 37)
	for _, university := range universities {
		if university.CountryCode != "NG" || university.StateID == "" {
			return fmt.Errorf("verify education dataset: %w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := ownershipCounts[university.OwnershipType]; !ok {
			return fmt.Errorf("verify education dataset: %w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := approvedUniversityStateIDs[university.StateID]; !ok {
			return fmt.Errorf("verify education dataset: %w", interfaces.ErrInvalidDatasetFile)
		}
		ownershipCounts[university.OwnershipType]++
		stateSeen[university.StateID] = struct{}{}
	}
	if ownershipCounts["federal"] != 77 || ownershipCounts["state"] != 69 || ownershipCounts["private"] != 182 {
		return fmt.Errorf("verify education dataset: %w", interfaces.ErrInvalidDatasetFile)
	}
	if len(stateSeen) != 37 {
		return fmt.Errorf("verify education dataset: %w", interfaces.ErrInvalidDatasetFile)
	}
	return nil
}

func verifyCollegeOfEducationDataset(ctx context.Context, service educationService) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if service == nil {
		return fmt.Errorf("verify education dataset: education service is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	colleges, err := service.ListCollegesOfEducation(ctx, services.CollegeOfEducationListInput{})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("verify education dataset: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(colleges) != 244 {
		return fmt.Errorf("verify education dataset: %w", interfaces.ErrInvalidDatasetFile)
	}

	ownershipCounts := map[string]int{
		"federal": 0,
		"state":   0,
		"private": 0,
	}
	stateSeen := make(map[string]struct{}, 37)
	for _, college := range colleges {
		if college.CountryCode != "NG" || college.ID == "" || college.Name == "" || college.StateID == "" {
			return fmt.Errorf("verify education dataset: %w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := ownershipCounts[college.OwnershipType]; !ok {
			return fmt.Errorf("verify education dataset: %w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := approvedUniversityStateIDs[college.StateID]; !ok {
			return fmt.Errorf("verify education dataset: %w", interfaces.ErrInvalidDatasetFile)
		}
		ownershipCounts[college.OwnershipType]++
		stateSeen[college.StateID] = struct{}{}
	}
	if ownershipCounts["federal"] != 28 || ownershipCounts["state"] != 48 || ownershipCounts["private"] != 168 {
		return fmt.Errorf("verify education dataset: %w", interfaces.ErrInvalidDatasetFile)
	}
	if len(stateSeen) != 37 {
		return fmt.Errorf("verify education dataset: %w", interfaces.ErrInvalidDatasetFile)
	}
	return nil
}

func verifyFinanceDataset(ctx context.Context, service financeService) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if service == nil {
		return fmt.Errorf("verify finance dataset: finance service is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	providers, err := service.ListPaymentServiceProviders(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("verify finance dataset: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(providers) != 255 {
		return fmt.Errorf("verify finance dataset: %w", interfaces.ErrInvalidDatasetFile)
	}

	expectedCounts := map[string]int{
		"mobile_money_operator":               17,
		"switching_and_processing_company":    19,
		"payment_solution_service_provider":   108,
		"payment_terminal_service_provider":   47,
		"super_agent":                         61,
		"payment_service_holding_company":     1,
		"payment_terminal_service_aggregator": 2,
	}
	counts := make(map[string]int, len(expectedCounts))
	for _, provider := range providers {
		if _, ok := expectedCounts[provider.InstitutionType]; !ok {
			return fmt.Errorf("verify finance dataset: %w", interfaces.ErrInvalidDatasetFile)
		}
		counts[provider.InstitutionType]++
	}
	for institutionType, want := range expectedCounts {
		if counts[institutionType] != want {
			return fmt.Errorf("verify finance dataset: %w", interfaces.ErrInvalidDatasetFile)
		}
	}

	operators, err := service.ListInternationalMoneyTransferOperators(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("verify finance dataset: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(operators) != 108 {
		return fmt.Errorf("verify finance dataset: %w", interfaces.ErrInvalidDatasetFile)
	}
	return nil
}

func verifyCurrencyDataset(ctx context.Context, service financeService) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if service == nil {
		return fmt.Errorf("verify currency dataset: finance service is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	currencies, err := service.ListCurrencies(ctx, services.CurrencyListInput{})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("verify currency dataset: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(currencies) != 155 {
		return fmt.Errorf("verify currency dataset: %w", interfaces.ErrInvalidDatasetFile)
	}

	const expectedRelationships = 252
	const expectedMappedCountryAreas = 245
	expectedSharedCurrencyCounts := map[string]int{
		"AUD": 8,
		"DKK": 3,
		"EUR": 36,
		"GBP": 4,
		"NZD": 5,
		"USD": 19,
		"XAF": 6,
		"XCD": 8,
		"XCG": 2,
		"XOF": 8,
		"XPF": 3,
	}
	zeroMapping := map[string]struct{}{
		"aq": {},
		"gs": {},
		"ps": {},
	}

	relationships := 0
	mappedCountryAreas := make(map[string]struct{}, expectedMappedCountryAreas)
	for _, currency := range currencies {
		if currency.ID == "" || currency.Name == "" || currency.AlphabeticCode == "" || currency.NumericCode == "" {
			return fmt.Errorf("verify currency dataset: %w", interfaces.ErrInvalidDatasetFile)
		}
		if len(currency.CountryAreaIDs) == 0 {
			if currency.AlphabeticCode != "TWD" {
				return fmt.Errorf("verify currency dataset: %w", interfaces.ErrInvalidDatasetFile)
			}
			continue
		}
		if expected, ok := expectedSharedCurrencyCounts[currency.AlphabeticCode]; ok && len(currency.CountryAreaIDs) != expected {
			return fmt.Errorf("verify currency dataset: %w", interfaces.ErrInvalidDatasetFile)
		}
		relationships += len(currency.CountryAreaIDs)
		for _, countryAreaID := range currency.CountryAreaIDs {
			mappedCountryAreas[countryAreaID] = struct{}{}
		}
	}

	if relationships != expectedRelationships {
		return fmt.Errorf("verify currency dataset: %w", interfaces.ErrInvalidDatasetFile)
	}
	if len(mappedCountryAreas) != expectedMappedCountryAreas {
		return fmt.Errorf("verify currency dataset: %w", interfaces.ErrInvalidDatasetFile)
	}
	for countryAreaID := range zeroMapping {
		if _, ok := mappedCountryAreas[countryAreaID]; ok {
			return fmt.Errorf("verify currency dataset: %w", interfaces.ErrInvalidDatasetFile)
		}
	}
	return nil
}

func buildRateLimitRepository(ctx context.Context, cfg *config.Config, logger *slog.Logger) (interfaces.RateLimitRepository, func() error, error) {
	return buildRateLimitRepositoryWith(
		ctx,
		cfg,
		logger,
		func(cfg config.RedisConfig) (redisBootstrapClient, error) {
			return redisclient.NewClient(cfg)
		},
		pingRedis,
		func(client redisBootstrapClient, prefix string) (interfaces.RateLimitRepository, error) {
			return redisrepo.NewRateLimitRepository(client, prefix)
		},
	)
}

type redisBootstrapClient interface {
	Ping(context.Context) *redisv9.StatusCmd
	Close() error
	Eval(context.Context, string, []string, ...any) *redisv9.Cmd
}

type redisClientFactory func(config.RedisConfig) (redisBootstrapClient, error)

type redisPingFunc func(context.Context, redisBootstrapClient) error

type redisRateLimitRepositoryFactory func(redisBootstrapClient, string) (interfaces.RateLimitRepository, error)

func buildRateLimitRepositoryWith(
	ctx context.Context,
	cfg *config.Config,
	logger *slog.Logger,
	newClient redisClientFactory,
	ping redisPingFunc,
	newRepo redisRateLimitRepositoryFactory,
) (interfaces.RateLimitRepository, func() error, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if cfg == nil {
		return nil, nil, fmt.Errorf("config is required")
	}
	if newClient == nil {
		return nil, nil, fmt.Errorf("redis client factory is required")
	}
	if ping == nil {
		return nil, nil, fmt.Errorf("redis ping function is required")
	}
	if newRepo == nil {
		return nil, nil, fmt.Errorf("rate limit repository factory is required")
	}

	prefix := strings.TrimSpace(cfg.Redis.KeyPrefix)
	if prefix == "" {
		prefix = defaultKeyPrefix
	}

	client, err := newClient(cfg.Redis)
	if err != nil {
		return nil, nil, fmt.Errorf("open redis client: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, redisStartupTimeout(cfg.Redis))
	defer cancel()

	if err := ping(pingCtx, client); err != nil {
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

	repo, err := newRepo(client, prefix)
	if err != nil {
		_ = client.Close()
		return nil, nil, fmt.Errorf("create rate limit repository: %w", err)
	}

	return repo, client.Close, nil
}

func pingRedis(ctx context.Context, client redisBootstrapClient) error {
	if client == nil {
		return fmt.Errorf("%w", redisclient.ErrInvalidRedisConfig)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := client.Ping(ctx).Err(); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("%w", redisclient.ErrRedisUnavailable)
	}
	return nil
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
