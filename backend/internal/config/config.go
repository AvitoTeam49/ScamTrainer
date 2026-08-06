package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
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

// ScenariosConfig wires the YAML scenario graphs into the chat domain. The
// graphs drive both the conversation and its scoring, so they are required.
//
// Chat stores scenario_id as bigint, while the scenario graphs are keyed by
// string ids taken from their YAML files. The mapping is declared explicitly
// instead of being derived from the directory listing: chats.scenario_id is
// already persisted, so an id that shifts when a YAML file is added would
// silently repoint existing chats at a different scenario.
type ScenariosConfig struct {
	Dir string
	Map map[int64]string
}

func (c ScenariosConfig) Enabled() bool {
	return c.Dir != ""
}

type Config struct {
	HTTP      HTTPConfig
	Postgres  PostgresConfig
	DeepSeek  DeepSeekConfig
	Scenarios ScenariosConfig
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

	scenarios, err := loadScenarios()
	if err != nil {
		return nil, err
	}

	return &Config{
		HTTP:      httpConfig,
		Postgres:  postgres,
		DeepSeek:  deepSeek,
		Scenarios: scenarios,
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

func loadScenarios() (ScenariosConfig, error) {
	dir := os.Getenv("SCENARIOS_DIR")
	if dir == "" {
		return ScenariosConfig{}, fmt.Errorf("%w: SCENARIOS_DIR", ErrMissingEnv)
	}

	mapping, err := parseScenarioMap(os.Getenv("SCENARIOS_MAP"))
	if err != nil {
		return ScenariosConfig{}, err
	}

	return ScenariosConfig{Dir: dir, Map: mapping}, nil
}

// parseScenarioMap reads pairs like "1:seller_fake_delivery,2:buyer_prepay".
func parseScenarioMap(raw string) (map[int64]string, error) {
	mapping := make(map[int64]string)

	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		id, scenarioID, found := strings.Cut(pair, ":")
		if !found {
			return nil, fmt.Errorf(
				"%w: SCENARIOS_MAP entry %q must look like <id>:<scenario_id>", ErrInvalidEnv, pair)
		}

		parsed, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: SCENARIOS_MAP entry %q: %v", ErrInvalidEnv, pair, err)
		}

		if parsed <= 0 {
			return nil, fmt.Errorf(
				"%w: SCENARIOS_MAP entry %q must use a positive id", ErrInvalidEnv, pair)
		}

		scenarioID = strings.TrimSpace(scenarioID)
		if scenarioID == "" {
			return nil, fmt.Errorf(
				"%w: SCENARIOS_MAP entry %q has an empty scenario id", ErrInvalidEnv, pair)
		}

		if _, exists := mapping[parsed]; exists {
			return nil, fmt.Errorf("%w: SCENARIOS_MAP contains duplicate id %d", ErrInvalidEnv, parsed)
		}

		mapping[parsed] = scenarioID
	}

	return mapping, nil
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
