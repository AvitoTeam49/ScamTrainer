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

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/agent/deepseek"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/config"
	chatpostgres "github.com/AvitoTeam49/ScamTrainer/backend/internal/repository/postgres/chat"
	userspostgres "github.com/AvitoTeam49/ScamTrainer/backend/internal/repository/postgres/users"
	scenarioyaml "github.com/AvitoTeam49/ScamTrainer/backend/internal/repository/yaml/scenario"
	chatrest "github.com/AvitoTeam49/ScamTrainer/backend/internal/transport/rest/chat"
	usersrest "github.com/AvitoTeam49/ScamTrainer/backend/internal/transport/rest/users"
	chatusecase "github.com/AvitoTeam49/ScamTrainer/backend/internal/usecase/chat"
	usersusecase "github.com/AvitoTeam49/ScamTrainer/backend/internal/usecase/users"
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

	scenarios, err := newScenarioSource(ctx, cfg.Scenarios, logger)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	chatService := chatusecase.New(
		chatpostgres.NewChatRepository(pool),
		chatpostgres.NewMessageRepository(pool),
		chatpostgres.NewDecisionRepository(pool),
		scenarios,
		deepseek.New(deepseek.Config{
			BaseURL:    cfg.DeepSeek.BaseURL,
			APIKey:     cfg.DeepSeek.APIKey,
			Model:      cfg.DeepSeek.Model,
			Timeout:    cfg.DeepSeek.Timeout,
			MaxRetries: cfg.DeepSeek.MaxRetries,
		}),
		chatusecase.Options{},
	)
	chatrest.NewHandler(chatService).Register(mux)

	usersService := usersusecase.NewUserService(userspostgres.NewPostgresRepository(pool), logger)
	usersrest.New(usersService, logger).Register(mux)

	server := &http.Server{
		Addr:    cfg.HTTP.Addr,
		Handler: http.StripPrefix(cfg.HTTP.Prefix, mux),
	}

	return serve(ctx, server, cfg.HTTP.ShutdownTimeout)
}

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

// newScenarioSource loads the scenario graphs. They are parsed and validated
// here so that a broken YAML file stops the process before it starts accepting
// traffic, and they are mandatory: the graphs drive both the conversation and
// its scoring.
func newScenarioSource(
	ctx context.Context,
	cfg config.ScenariosConfig,
	logger *zap.Logger,
) (*graphScenarioSource, error) {
	graphs, err := scenarioyaml.NewYAMLRepository(cfg.Dir)
	if err != nil {
		return nil, fmt.Errorf("failed to load scenarios: %w", err)
	}

	source := newGraphScenarioSource(graphs, cfg.Map)
	if err := source.verify(ctx); err != nil {
		return nil, fmt.Errorf("failed to verify scenario mapping: %w", err)
	}

	logger.Info("scenario graphs loaded",
		zap.String("dir", cfg.Dir),
		zap.Int("mapped_scenarios", len(cfg.Map)),
	)

	return source, nil
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
