package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type DatabaseConfig struct {
	URL                   *string
	MaxConnections        int
	MinConnections        int
	MaxConnectionLifetime time.Duration
	MaxConnectionIdleTime time.Duration
	HealthCheckPeriod     time.Duration
}

func loadDatabaseConfig(environment AppEnvironment) (DatabaseConfig, error) {
	url := strings.TrimSpace(getEnv("DATABASE_URL"))
	switch environment {
	case AppEnvironmentStaging, AppEnvironmentProduction:
		if url == "" {
			return DatabaseConfig{}, fmt.Errorf("DATABASE_URL is required in %s", environment)
		}
	case AppEnvironmentDevelopment, AppEnvironmentTest:
		// Optional in local and test environments.
	default:
		return DatabaseConfig{}, fmt.Errorf("invalid APP_ENV value")
	}

	var databaseURL *string
	if url != "" {
		databaseURL = &url
	}

	maxConnections, err := parsePositiveInt("DATABASE_MAX_CONNECTIONS", getEnv("DATABASE_MAX_CONNECTIONS"), 10)
	if err != nil {
		return DatabaseConfig{}, err
	}

	minConnections, err := parsePositiveInt("DATABASE_MIN_CONNECTIONS", getEnv("DATABASE_MIN_CONNECTIONS"), 2)
	if err != nil {
		return DatabaseConfig{}, err
	}

	if minConnections > maxConnections {
		return DatabaseConfig{}, fmt.Errorf("DATABASE_MIN_CONNECTIONS must not exceed DATABASE_MAX_CONNECTIONS")
	}

	maxConnectionLifetime, err := parsePositiveDuration("DATABASE_MAX_CONNECTION_LIFETIME", getEnv("DATABASE_MAX_CONNECTION_LIFETIME"), 30*time.Minute)
	if err != nil {
		return DatabaseConfig{}, err
	}

	maxConnectionIdleTime, err := parsePositiveDuration("DATABASE_MAX_CONNECTION_IDLE_TIME", getEnv("DATABASE_MAX_CONNECTION_IDLE_TIME"), 5*time.Minute)
	if err != nil {
		return DatabaseConfig{}, err
	}

	healthCheckPeriod, err := parsePositiveDuration("DATABASE_HEALTH_CHECK_PERIOD", getEnv("DATABASE_HEALTH_CHECK_PERIOD"), time.Minute)
	if err != nil {
		return DatabaseConfig{}, err
	}

	return DatabaseConfig{
		URL:                   databaseURL,
		MaxConnections:        maxConnections,
		MinConnections:        minConnections,
		MaxConnectionLifetime: maxConnectionLifetime,
		MaxConnectionIdleTime: maxConnectionIdleTime,
		HealthCheckPeriod:     healthCheckPeriod,
	}, nil
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
