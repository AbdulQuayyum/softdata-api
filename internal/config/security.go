package config

import (
	"fmt"
	"strings"
	"time"
)

type SecurityConfig struct {
	AuthTokenSecret *string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func loadSecurityConfig(environment AppEnvironment) (SecurityConfig, error) {
	secret := strings.TrimSpace(getEnv("AUTH_TOKEN_SECRET"))
	switch environment {
	case AppEnvironmentStaging, AppEnvironmentProduction:
		if secret == "" {
			return SecurityConfig{}, fmt.Errorf("AUTH_TOKEN_SECRET is required in %s", environment)
		}
	case AppEnvironmentDevelopment, AppEnvironmentTest:
		// Optional in local and test environments.
	default:
		return SecurityConfig{}, fmt.Errorf("invalid APP_ENV value")
	}

	var authTokenSecret *string
	if secret != "" {
		authTokenSecret = &secret
	}

	accessTokenTTL, err := parsePositiveDuration("ACCESS_TOKEN_TTL", getEnv("ACCESS_TOKEN_TTL"), 15*time.Minute)
	if err != nil {
		return SecurityConfig{}, err
	}

	refreshTokenTTL, err := parsePositiveDuration("REFRESH_TOKEN_TTL", getEnv("REFRESH_TOKEN_TTL"), 720*time.Hour)
	if err != nil {
		return SecurityConfig{}, err
	}

	return SecurityConfig{
		AuthTokenSecret: authTokenSecret,
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
	}, nil
}
