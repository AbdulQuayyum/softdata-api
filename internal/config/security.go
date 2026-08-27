package config

import (
	"fmt"
	"strings"
	"time"
)

type SecurityConfig struct {
	AuthTokenSecret   string
	JWTIssuer         string
	JWTAudience       string
	AnonymousIDSecret string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
}

func loadSecurityConfig(lookup LookupEnv) (SecurityConfig, error) {
	authTokenSecret, err := loadRequiredSecret(lookup, "AUTH_TOKEN_SECRET")
	if err != nil {
		return SecurityConfig{}, err
	}

	anonymousIDSecret, err := loadRequiredSecret(lookup, "ANONYMOUS_ID_SECRET")
	if err != nil {
		return SecurityConfig{}, err
	}

	accessTokenTTL, err := parsePositiveDuration("ACCESS_TOKEN_TTL", lookupString(lookup, "ACCESS_TOKEN_TTL"), 15*time.Minute)
	if err != nil {
		return SecurityConfig{}, err
	}

	refreshTokenTTL, err := parsePositiveDuration("REFRESH_TOKEN_TTL", lookupString(lookup, "REFRESH_TOKEN_TTL"), 720*time.Hour)
	if err != nil {
		return SecurityConfig{}, err
	}
	if refreshTokenTTL <= accessTokenTTL {
		return SecurityConfig{}, fmt.Errorf("REFRESH_TOKEN_TTL must be greater than ACCESS_TOKEN_TTL")
	}

	return SecurityConfig{
		AuthTokenSecret:   authTokenSecret,
		JWTIssuer:         lookupDefault(lookup, "JWT_ISSUER", "softdata-api"),
		JWTAudience:       lookupDefault(lookup, "JWT_AUDIENCE", "softdata-api"),
		AnonymousIDSecret: anonymousIDSecret,
		AccessTokenTTL:    accessTokenTTL,
		RefreshTokenTTL:   refreshTokenTTL,
	}, nil
}

func loadRequiredSecret(lookup LookupEnv, name string) (string, error) {
	secret := lookupString(lookup, name)
	if secret == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	if len(secret) < 32 {
		return "", fmt.Errorf("%s must be at least 32 bytes", name)
	}
	return secret, nil
}

func lookupDefault(lookup LookupEnv, name, defaultValue string) string {
	value := lookupString(lookup, name)
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}
