package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type LookupEnv func(string) (string, bool)

type AppEnvironment string

const (
	AppEnvironmentDevelopment AppEnvironment = "development"
	AppEnvironmentTest        AppEnvironment = "test"
	AppEnvironmentStaging     AppEnvironment = "staging"
	AppEnvironmentProduction  AppEnvironment = "production"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	Security    SecurityConfig
	RateLimit   RateLimitConfig
	Datasets    DatasetConfig
}

// Load reads configuration from the process environment.
func Load() (*Config, error) {
	cfg, err := load(os.LookupEnv)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func load(lookup LookupEnv) (Config, error) {
	if lookup == nil {
		lookup = os.LookupEnv
	}

	env, err := loadEnvironment(lookup)
	if err != nil {
		return Config{}, err
	}

	server, err := loadServerConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	database, err := loadDatabaseConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	redisConfig, err := loadRedisConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	security, err := loadSecurityConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	rateLimit, err := loadRateLimitConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	datasets, err := loadDatasetsConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment: string(env),
		Server:      server,
		Database:    database,
		Redis:       redisConfig,
		Security:    security,
		RateLimit:   rateLimit,
		Datasets:    datasets,
	}, nil
}

func loadEnvironment(lookup LookupEnv) (AppEnvironment, error) {
	value := strings.ToLower(strings.TrimSpace(lookupString(lookup, "APP_ENV")))
	if value == "" {
		value = string(AppEnvironmentDevelopment)
	}

	switch AppEnvironment(value) {
	case AppEnvironmentDevelopment, AppEnvironmentTest, AppEnvironmentStaging, AppEnvironmentProduction:
		return AppEnvironment(value), nil
	default:
		return "", fmt.Errorf("invalid APP_ENV value")
	}
}

func lookupString(lookup LookupEnv, name string) string {
	if lookup == nil {
		return ""
	}
	value, ok := lookup(name)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func parsePositiveInt(name, raw string, defaultValue int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("invalid %s", name)
	}
	if value <= 0 {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return value, nil
}

func parseNonNegativeInt(name, raw string, defaultValue int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("invalid %s", name)
	}
	if value < 0 {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return value, nil
}

func parsePositiveInt64(name, raw string, defaultValue int64) (int64, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}

	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", name)
	}
	if value <= 0 {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return value, nil
}

func parseNonNegativeInt64(name, raw string, defaultValue int64) (int64, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}

	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", name)
	}
	if value < 0 {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return value, nil
}

func parsePositiveDuration(name, raw string, defaultValue time.Duration) (time.Duration, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}

	value, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("invalid %s", name)
	}
	if value <= 0 {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return value, nil
}
