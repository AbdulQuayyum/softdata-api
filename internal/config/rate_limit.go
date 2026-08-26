package config

import (
	"fmt"
	"strings"
	"time"
)

type RateLimitConfig struct {
	AnonymousRequestLimit  int
	APIKeyRequestLimit     int
	APIKeyMonthlyAllowance int
	Window                 time.Duration
}

func loadRateLimitConfig() (RateLimitConfig, error) {
	anonymousLimit, err := parsePositiveInt("ANONYMOUS_RATE_LIMIT", getEnv("ANONYMOUS_RATE_LIMIT"), 60)
	if err != nil {
		return RateLimitConfig{}, err
	}

	apiKeyLimit, err := parsePositiveInt("API_KEY_RATE_LIMIT", getEnv("API_KEY_RATE_LIMIT"), 300)
	if err != nil {
		return RateLimitConfig{}, err
	}

	apiKeyMonthlyAllowance, err := parsePositiveInt("API_KEY_MONTHLY_LIMIT", getEnv("API_KEY_MONTHLY_LIMIT"), 50000)
	if err != nil {
		return RateLimitConfig{}, err
	}

	window, err := parsePositiveDuration("RATE_LIMIT_WINDOW", getEnv("RATE_LIMIT_WINDOW"), time.Minute)
	if err != nil {
		return RateLimitConfig{}, err
	}

	return RateLimitConfig{
		AnonymousRequestLimit:  anonymousLimit,
		APIKeyRequestLimit:     apiKeyLimit,
		APIKeyMonthlyAllowance: apiKeyMonthlyAllowance,
		Window:                 window,
	}, nil
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
