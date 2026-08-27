package config

type UsageConfig struct {
	APIKeyMonthlyAllowance int64
}

func loadUsageConfig(lookup LookupEnv) (UsageConfig, error) {
	allowance, err := parsePositiveInt64("API_KEY_MONTHLY_LIMIT", lookupString(lookup, "API_KEY_MONTHLY_LIMIT"), 50000)
	if err != nil {
		return UsageConfig{}, err
	}

	return UsageConfig{
		APIKeyMonthlyAllowance: allowance,
	}, nil
}
