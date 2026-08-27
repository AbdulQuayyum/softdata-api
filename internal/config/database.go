package config

import (
	"fmt"
	"time"
)

type DatabaseConfig struct {
	URL                   *string
	MaxConnections        int
	MinConnections        int
	MaxConnectionLifetime time.Duration
	MaxConnectionIdleTime time.Duration
	HealthCheckPeriod     time.Duration
	ConnectTimeout        time.Duration
}

func loadDatabaseConfig(lookup LookupEnv) (DatabaseConfig, error) {
	url := lookupString(lookup, "DATABASE_URL")
	if url == "" {
		return DatabaseConfig{}, fmt.Errorf("DATABASE_URL is required")
	}

	databaseURL := &url

	maxConnections, err := parsePositiveInt("DATABASE_MAX_CONNECTIONS", lookupString(lookup, "DATABASE_MAX_CONNECTIONS"), 10)
	if err != nil {
		return DatabaseConfig{}, err
	}

	minConnections, err := parsePositiveInt("DATABASE_MIN_CONNECTIONS", lookupString(lookup, "DATABASE_MIN_CONNECTIONS"), 2)
	if err != nil {
		return DatabaseConfig{}, err
	}

	if minConnections > maxConnections {
		return DatabaseConfig{}, fmt.Errorf("DATABASE_MIN_CONNECTIONS must not exceed DATABASE_MAX_CONNECTIONS")
	}

	maxConnectionLifetime, err := parsePositiveDuration("DATABASE_MAX_CONNECTION_LIFETIME", lookupString(lookup, "DATABASE_MAX_CONNECTION_LIFETIME"), 30*time.Minute)
	if err != nil {
		return DatabaseConfig{}, err
	}

	maxConnectionIdleTime, err := parsePositiveDuration("DATABASE_MAX_CONNECTION_IDLE_TIME", lookupString(lookup, "DATABASE_MAX_CONNECTION_IDLE_TIME"), 5*time.Minute)
	if err != nil {
		return DatabaseConfig{}, err
	}

	healthCheckPeriod, err := parsePositiveDuration("DATABASE_HEALTH_CHECK_PERIOD", lookupString(lookup, "DATABASE_HEALTH_CHECK_PERIOD"), time.Minute)
	if err != nil {
		return DatabaseConfig{}, err
	}

	connectTimeout, err := parsePositiveDuration("DATABASE_CONNECT_TIMEOUT", lookupString(lookup, "DATABASE_CONNECT_TIMEOUT"), 10*time.Second)
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
		ConnectTimeout:        connectTimeout,
	}, nil
}
