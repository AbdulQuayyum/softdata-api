package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ServerConfig struct {
	Host              string
	Port              int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	RequestTimeout    time.Duration
	ShutdownTimeout   time.Duration
	MaxBodyBytes      int64
	AllowedOrigins    []string
}

func (s ServerConfig) ListenAddress() string {
	host := strings.TrimSpace(s.Host)
	return net.JoinHostPort(host, strconv.Itoa(s.Port))
}

func loadServerConfig(lookup LookupEnv) (ServerConfig, error) {
	host := lookupString(lookup, "SERVER_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	port, err := parsePort(lookupString(lookup, "SERVER_PORT"))
	if err != nil {
		return ServerConfig{}, err
	}

	readHeaderTimeout, err := parsePositiveDuration("SERVER_READ_HEADER_TIMEOUT", lookupString(lookup, "SERVER_READ_HEADER_TIMEOUT"), 5*time.Second)
	if err != nil {
		return ServerConfig{}, err
	}

	readTimeout, err := parsePositiveDuration("SERVER_READ_TIMEOUT", lookupString(lookup, "SERVER_READ_TIMEOUT"), 10*time.Second)
	if err != nil {
		return ServerConfig{}, err
	}

	writeTimeout, err := parsePositiveDuration("SERVER_WRITE_TIMEOUT", lookupString(lookup, "SERVER_WRITE_TIMEOUT"), 10*time.Second)
	if err != nil {
		return ServerConfig{}, err
	}

	idleTimeout, err := parsePositiveDuration("SERVER_IDLE_TIMEOUT", lookupString(lookup, "SERVER_IDLE_TIMEOUT"), 60*time.Second)
	if err != nil {
		return ServerConfig{}, err
	}

	requestTimeout, err := parsePositiveDuration("SERVER_REQUEST_TIMEOUT", lookupString(lookup, "SERVER_REQUEST_TIMEOUT"), 30*time.Second)
	if err != nil {
		return ServerConfig{}, err
	}

	shutdownTimeout, err := parsePositiveDuration("SERVER_SHUTDOWN_TIMEOUT", lookupString(lookup, "SERVER_SHUTDOWN_TIMEOUT"), 10*time.Second)
	if err != nil {
		return ServerConfig{}, err
	}

	maxBodyBytes, err := parsePositiveInt64("SERVER_BODY_LIMIT", lookupString(lookup, "SERVER_BODY_LIMIT"), 1<<20)
	if err != nil {
		return ServerConfig{}, err
	}

	allowedOrigins, err := parseExactOrigins(lookupString(lookup, "SERVER_ALLOWED_ORIGINS"))
	if err != nil {
		return ServerConfig{}, err
	}

	return ServerConfig{
		Host:              host,
		Port:              port,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		RequestTimeout:    requestTimeout,
		ShutdownTimeout:   shutdownTimeout,
		MaxBodyBytes:      maxBodyBytes,
		AllowedOrigins:    allowedOrigins,
	}, nil
}

func parsePort(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return 8080, nil
	}

	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("invalid SERVER_PORT")
	}
	if value < 1 || value > 65535 {
		return 0, fmt.Errorf("invalid SERVER_PORT")
	}
	return value, nil
}

func parseExactOrigins(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		normalized, err := normalizeExactOrigin(origin)
		if err != nil {
			return nil, err
		}
		origins = append(origins, normalized)
	}
	return origins, nil
}

func normalizeExactOrigin(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("invalid SERVER_ALLOWED_ORIGINS value")
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid SERVER_ALLOWED_ORIGINS value")
	}
	if parsed.User != nil || parsed.Path != "" && parsed.Path != "/" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("invalid SERVER_ALLOWED_ORIGINS value")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("invalid SERVER_ALLOWED_ORIGINS value")
	}

	return strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host), nil
}
