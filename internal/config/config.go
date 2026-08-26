package config

import (
	"fmt"
	"os"
	"strings"
)

type AppEnvironment string

const (
	AppEnvironmentDevelopment AppEnvironment = "development"
	AppEnvironmentTest        AppEnvironment = "test"
	AppEnvironmentStaging     AppEnvironment = "staging"
	AppEnvironmentProduction  AppEnvironment = "production"
)

type Config struct {
	App       AppConfig
	Server    ServerConfig
	Database  DatabaseConfig
	Security  SecurityConfig
	RateLimit RateLimitConfig
	Datasets  DatasetsConfig
}

type AppConfig struct {
	Environment AppEnvironment
}

func Load() (*Config, error) {
	appConfig, err := loadAppConfig()
	if err != nil {
		return nil, err
	}

	serverConfig, err := loadServerConfig()
	if err != nil {
		return nil, err
	}

	databaseConfig, err := loadDatabaseConfig(appConfig.Environment)
	if err != nil {
		return nil, err
	}

	securityConfig, err := loadSecurityConfig(appConfig.Environment)
	if err != nil {
		return nil, err
	}

	rateLimitConfig, err := loadRateLimitConfig()
	if err != nil {
		return nil, err
	}

	datasetsConfig, err := loadDatasetsConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		App:       appConfig,
		Server:    serverConfig,
		Database:  databaseConfig,
		Security:  securityConfig,
		RateLimit: rateLimitConfig,
		Datasets:  datasetsConfig,
	}, nil
}

func loadAppConfig() (AppConfig, error) {
	env := strings.ToLower(strings.TrimSpace(getEnv("APP_ENV")))
	if env == "" {
		env = string(AppEnvironmentDevelopment)
	}

	switch AppEnvironment(env) {
	case AppEnvironmentDevelopment, AppEnvironmentTest, AppEnvironmentStaging, AppEnvironmentProduction:
		return AppConfig{Environment: AppEnvironment(env)}, nil
	default:
		return AppConfig{}, fmt.Errorf("invalid APP_ENV value")
	}
}

func getEnv(name string) string {
	return os.Getenv(name)
}
