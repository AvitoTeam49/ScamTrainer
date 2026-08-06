package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultHTTPAddr            = ":8080"
	defaultHTTPPrefix          = "/api"
	defaultHTTPShutdownTimeout = 10 * time.Second
	defaultDeepSeekBaseURL     = "https://api.deepseek.com"
	defaultDeepSeekModel       = "deepseek-v4-flash"
	defaultDeepSeekTimeout     = 60 * time.Second
	defaultDeepSeekMaxRetries  = 2
)

var (
	ErrMissingEnv = errors.New("required environment variable is not set")
	ErrInvalidEnv = errors.New("invalid environment variable value")
)

type DeepSeekConfig struct {
	BaseURL    string
	APIKey     string
	Model      string
	Timeout    time.Duration
	MaxRetries int
}

type HTTPConfig struct {
	Addr            string
	Prefix          string
	ShutdownTimeout time.Duration
}

type PostgresConfig struct {
	DSN string
}

type Config struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
	DeepSeek DeepSeekConfig
}

func Load() (*Config, error) {
	httpConfig, err := loadHTTP()
	if err != nil {
		return nil, err
	}

	postgres, err := loadPostgres()
	if err != nil {
		return nil, err
	}

	deepSeek, err := loadDeepSeek()
	if err != nil {
		return nil, err
	}

	return &Config{
		HTTP:     httpConfig,
		Postgres: postgres,
		DeepSeek: deepSeek,
	}, nil
}

func loadHTTP() (HTTPConfig, error) {
	shutdownTimeout, err := durationEnv("HTTP_SHUTDOWN_TIMEOUT", defaultHTTPShutdownTimeout)
	if err != nil {
		return HTTPConfig{}, err
	}

	return HTTPConfig{
		Addr:            stringEnv("HTTP_ADDR", defaultHTTPAddr),
		Prefix:          stringEnv("HTTP_PREFIX", defaultHTTPPrefix),
		ShutdownTimeout: shutdownTimeout,
	}, nil
}

func loadPostgres() (PostgresConfig, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return PostgresConfig{}, fmt.Errorf("%w: DATABASE_URL", ErrMissingEnv)
	}

	return PostgresConfig{DSN: dsn}, nil
}

func loadDeepSeek() (DeepSeekConfig, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return DeepSeekConfig{}, fmt.Errorf("%w: DEEPSEEK_API_KEY", ErrMissingEnv)
	}

	timeout, err := durationEnv("DEEPSEEK_TIMEOUT", defaultDeepSeekTimeout)
	if err != nil {
		return DeepSeekConfig{}, err
	}

	maxRetries, err := intEnv("DEEPSEEK_MAX_RETRIES", defaultDeepSeekMaxRetries)
	if err != nil {
		return DeepSeekConfig{}, err
	}

	if maxRetries < 0 {
		return DeepSeekConfig{}, fmt.Errorf("%w: DEEPSEEK_MAX_RETRIES must not be negative", ErrInvalidEnv)
	}

	return DeepSeekConfig{
		BaseURL:    stringEnv("DEEPSEEK_BASE_URL", defaultDeepSeekBaseURL),
		APIKey:     apiKey,
		Model:      stringEnv("DEEPSEEK_MODEL", defaultDeepSeekModel),
		Timeout:    timeout,
		MaxRetries: maxRetries,
	}, nil
}

func stringEnv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}

func durationEnv(name string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%w: %s: %v", ErrInvalidEnv, name, err)
	}

	if value <= 0 {
		return 0, fmt.Errorf("%w: %s must be positive", ErrInvalidEnv, name)
	}

	return value, nil
}

func intEnv(name string, fallback int) (int, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%w: %s: %v", ErrInvalidEnv, name, err)
	}

	return value, nil
}
