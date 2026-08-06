package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/agent/deepseek"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/domain"
	chatpg "github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/repository/postgres"
	chatscenario "github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/scenario"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/service"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/transport/rest"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/config"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/scenario"
	userscontroller "github.com/AvitoTeam49/ScamTrainer/backend/internal/users/controller"
	userspg "github.com/AvitoTeam49/ScamTrainer/backend/internal/users/repository/postgres"
	usersusecase "github.com/AvitoTeam49/ScamTrainer/backend/internal/users/usecase"
	db "github.com/AvitoTeam49/ScamTrainer/backend/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger, err := newLogger()
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	pool, err := pgxpool.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		return fmt.Errorf("failed to create postgres pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to reach postgres: %w", err)
	}

	if err := db.Up(ctx, pool); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	scenarios, err := newScenarioProvider(cfg.Scenarios, logger)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	chatService := service.New(
		chatpg.NewChatRepository(pool),
		chatpg.NewMessageRepository(pool),
		chatpg.NewIncidentRepository(pool),
		scenarios,
		deepseek.New(deepseek.Config{
			BaseURL:    cfg.DeepSeek.BaseURL,
			APIKey:     cfg.DeepSeek.APIKey,
			Model:      cfg.DeepSeek.Model,
			Timeout:    cfg.DeepSeek.Timeout,
			MaxRetries: cfg.DeepSeek.MaxRetries,
		}),
		service.Options{},
	)
	rest.NewHandler(chatService).Register(mux)

	usersService := usersusecase.NewUserService(userspg.NewPostgresRepository(pool), logger)
	userscontroller.New(usersService, logger).Register(mux)

	server := &http.Server{
		Addr:    cfg.HTTP.Addr,
		Handler: http.StripPrefix(cfg.HTTP.Prefix, mux),
	}

	return serve(ctx, server, cfg.HTTP.ShutdownTimeout)
}

// newLogger builds the application logger. The users domain is written against
// zap while the chat transport logs through log/slog, so slog is pointed at the
// same JSON stream to keep a single log format on stderr.
func newLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stderr"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	return logger, nil
}

// newScenarioProvider prefers the scenario graphs from internal/scenario, which
// is the source of truth for scenario content. When SCENARIOS_DIR is not set the
// chat domain keeps using its built-in static prompt.
func newScenarioProvider(
	cfg config.ScenariosConfig,
	logger *zap.Logger,
) (domain.ScenarioProvider, error) {
	if !cfg.Enabled() {
		logger.Info("scenario graphs disabled, falling back to the static prompt")

		return chatscenario.NewStaticProvider(""), nil
	}

	// Graphs are parsed and validated here so that a broken YAML file stops the
	// process before it starts accepting traffic.
	graphs, err := scenario.NewYAMLRepository(cfg.Dir)
	if err != nil {
		return nil, fmt.Errorf("failed to load scenarios: %w", err)
	}

	logger.Info("scenario graphs loaded",
		zap.String("dir", cfg.Dir),
		zap.Int("mapped_scenarios", len(cfg.Map)),
	)

	return newGraphScenarioProvider(graphs, cfg.Map), nil
}

func serve(ctx context.Context, server *http.Server, shutdownTimeout time.Duration) error {
	failed := make(chan error, 1)

	go func() {
		slog.Info("http server started", "addr", server.Addr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			failed <- err
		}
	}()

	select {
	case err := <-failed:
		return fmt.Errorf("http server failed: %w", err)
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shut down http server: %w", err)
	}

	return nil
}
