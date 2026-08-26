package config

import (
	"fmt"
	"net"
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
	ShutdownTimeout   time.Duration
}

func (s ServerConfig) ListenAddress() string {
	host := strings.TrimSpace(s.Host)
	return net.JoinHostPort(host, strconv.Itoa(s.Port))
}

func loadServerConfig() (ServerConfig, error) {
	host := strings.TrimSpace(getEnv("SERVER_HOST"))
	if host == "" {
		host = "127.0.0.1"
	}

	port, err := parsePort(getEnv("SERVER_PORT"))
	if err != nil {
		return ServerConfig{}, err
	}

	readHeaderTimeout, err := parsePositiveDuration("SERVER_READ_HEADER_TIMEOUT", getEnv("SERVER_READ_HEADER_TIMEOUT"), 5*time.Second)
	if err != nil {
		return ServerConfig{}, err
	}

	readTimeout, err := parsePositiveDuration("SERVER_READ_TIMEOUT", getEnv("SERVER_READ_TIMEOUT"), 10*time.Second)
	if err != nil {
		return ServerConfig{}, err
	}

	writeTimeout, err := parsePositiveDuration("SERVER_WRITE_TIMEOUT", getEnv("SERVER_WRITE_TIMEOUT"), 10*time.Second)
	if err != nil {
		return ServerConfig{}, err
	}

	idleTimeout, err := parsePositiveDuration("SERVER_IDLE_TIMEOUT", getEnv("SERVER_IDLE_TIMEOUT"), 60*time.Second)
	if err != nil {
		return ServerConfig{}, err
	}

	shutdownTimeout, err := parsePositiveDuration("SERVER_SHUTDOWN_TIMEOUT", getEnv("SERVER_SHUTDOWN_TIMEOUT"), 10*time.Second)
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
		ShutdownTimeout:   shutdownTimeout,
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
