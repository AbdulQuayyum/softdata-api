package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type RateLimitConfig struct {
	AnonymousRequestLimit int
	APIKeyRequestLimit    int
	DatasetDownloadLimit  int
	Window                time.Duration
	FailOpen              bool
}

func loadRateLimitConfig(lookup LookupEnv) (RateLimitConfig, error) {
	anonymousLimit, err := parsePositiveInt("ANONYMOUS_RATE_LIMIT", lookupString(lookup, "ANONYMOUS_RATE_LIMIT"), 60)
	if err != nil {
		return RateLimitConfig{}, err
	}

	apiKeyLimit, err := parsePositiveInt("API_KEY_RATE_LIMIT", lookupString(lookup, "API_KEY_RATE_LIMIT"), 300)
	if err != nil {
		return RateLimitConfig{}, err
	}

	datasetDownloadLimit, err := parsePositiveInt("DATASET_DOWNLOAD_RATE_LIMIT", lookupString(lookup, "DATASET_DOWNLOAD_RATE_LIMIT"), 10)
	if err != nil {
		return RateLimitConfig{}, err
	}

	window, err := parsePositiveDuration("RATE_LIMIT_WINDOW", lookupString(lookup, "RATE_LIMIT_WINDOW"), time.Minute)
	if err != nil {
		return RateLimitConfig{}, err
	}

	failOpen := true
	if raw := lookupString(lookup, "RATE_LIMIT_FAIL_OPEN"); raw != "" {
		parsed, err := parseBool("RATE_LIMIT_FAIL_OPEN", raw)
		if err != nil {
			return RateLimitConfig{}, err
		}
		failOpen = parsed
	}

	return RateLimitConfig{
		AnonymousRequestLimit: anonymousLimit,
		APIKeyRequestLimit:    apiKeyLimit,
		DatasetDownloadLimit:  datasetDownloadLimit,
		Window:                window,
		FailOpen:              failOpen,
	}, nil
}

func parseBool(name, raw string) (bool, error) {
	parsed, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return false, fmt.Errorf("invalid %s", name)
	}
	return parsed, nil
}
