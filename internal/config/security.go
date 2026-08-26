package config

import (
	"fmt"
	"strings"
	"time"
)

const (
	developmentAuthTokenSecretFallback   = "dev-only-insecure-auth-token-secret"
	developmentAnonymousIDSecretFallback = "dev-only-insecure-anonymous-id-secret"
)

type SecurityConfig struct {
	AuthTokenSecret   *string
	AnonymousIDSecret *string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
}

func loadSecurityConfig(environment AppEnvironment) (SecurityConfig, error) {
	authTokenSecret, err := loadSecretForEnvironment(
		"AUTH_TOKEN_SECRET",
		environment,
		developmentAuthTokenSecretFallback,
		true,
	)
	if err != nil {
		return SecurityConfig{}, err
	}

	anonymousIDSecret, err := loadSecretForEnvironment(
		"ANONYMOUS_ID_SECRET",
		environment,
		developmentAnonymousIDSecretFallback,
		true,
	)
	if err != nil {
		return SecurityConfig{}, err
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
		AuthTokenSecret:   authTokenSecret,
		AnonymousIDSecret: anonymousIDSecret,
		AccessTokenTTL:    accessTokenTTL,
		RefreshTokenTTL:   refreshTokenTTL,
	}, nil
}

func loadSecretForEnvironment(name string, environment AppEnvironment, fallback string, requireMinimumLength bool) (*string, error) {
	secret := strings.TrimSpace(getEnv(name))

	switch environment {
	case AppEnvironmentDevelopment, AppEnvironmentTest:
		if secret == "" {
			secret = fallback
		}
	case AppEnvironmentStaging, AppEnvironmentProduction:
		if secret == "" {
			return nil, fmt.Errorf("%s is required in %s", name, environment)
		}
		if requireMinimumLength && len(secret) < 32 {
			return nil, fmt.Errorf("%s must be at least 32 bytes in %s", name, environment)
		}
	default:
		return nil, fmt.Errorf("invalid APP_ENV value")
	}

	return &secret, nil
}
