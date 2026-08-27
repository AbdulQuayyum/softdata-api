package config

import "time"

type RedisConfig struct {
	URL          *string
	Address      string
	Username     *string
	Password     *string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	KeyPrefix    string
}

func loadRedisConfig(lookup LookupEnv) (RedisConfig, error) {
	url := lookupString(lookup, "REDIS_URL")
	address := lookupString(lookup, "REDIS_ADDRESS")

	username := lookupString(lookup, "REDIS_USERNAME")
	password := lookupString(lookup, "REDIS_PASSWORD")

	db, err := parseNonNegativeInt("REDIS_DB", lookupString(lookup, "REDIS_DB"), 0)
	if err != nil {
		return RedisConfig{}, err
	}

	dialTimeout, err := parsePositiveDuration("REDIS_DIAL_TIMEOUT", lookupString(lookup, "REDIS_DIAL_TIMEOUT"), 5*time.Second)
	if err != nil {
		return RedisConfig{}, err
	}

	readTimeout, err := parsePositiveDuration("REDIS_READ_TIMEOUT", lookupString(lookup, "REDIS_READ_TIMEOUT"), 3*time.Second)
	if err != nil {
		return RedisConfig{}, err
	}

	writeTimeout, err := parsePositiveDuration("REDIS_WRITE_TIMEOUT", lookupString(lookup, "REDIS_WRITE_TIMEOUT"), 3*time.Second)
	if err != nil {
		return RedisConfig{}, err
	}

	poolSize, err := parsePositiveInt("REDIS_POOL_SIZE", lookupString(lookup, "REDIS_POOL_SIZE"), 10)
	if err != nil {
		return RedisConfig{}, err
	}

	keyPrefix := lookupString(lookup, "REDIS_KEY_PREFIX")

	cfg := RedisConfig{
		Address:      address,
		DB:           db,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		PoolSize:     poolSize,
		KeyPrefix:    keyPrefix,
	}
	if url != "" {
		cfg.URL = &url
	}
	if username != "" {
		cfg.Username = &username
	}
	if password != "" {
		cfg.Password = &password
	}

	return cfg, nil
}
