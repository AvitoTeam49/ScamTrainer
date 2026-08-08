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
	authrest "github.com/AvitoTeam49/ScamTrainer/backend/internal/auth/controller"
	authpostgres "github.com/AvitoTeam49/ScamTrainer/backend/internal/auth/repository"
	authusecase "github.com/AvitoTeam49/ScamTrainer/backend/internal/auth/usecase/auth"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/config"
	scenariodomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/scenario"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/eventbus"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/middleware"
	chatpostgres "github.com/AvitoTeam49/ScamTrainer/backend/internal/repository/postgres/chat"
	userspostgres "github.com/AvitoTeam49/ScamTrainer/backend/internal/repository/postgres/users"
	scenarioyaml "github.com/AvitoTeam49/ScamTrainer/backend/internal/repository/yaml/scenario"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/training"
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
	protected := http.NewServeMux()

	authService := authusecase.NewAuthService(
		authpostgres.NewPostgresRepository(pool),
		logger,
		cfg.Auth.JWTSecret,
	)
	authrest.New(authService, logger).Register(mux)

	mux.Handle("/v1/", middleware.Auth(authService)(protected))

	bus := eventbus.New(0)

	usersService := usersusecase.NewUserService(userspostgres.NewPostgresRepository(pool), logger)

	sessions := training.NewInMemorySessionRepository()
	trainingService := training.NewService(
		scenarios,
		sessions,
		scenariodomain.NewEngine(),
		training.UUIDGenerator{},
	)

	chatService := chatusecase.New(
		chatpostgres.NewChatRepository(pool),
		chatpostgres.NewMessageRepository(pool),
		chatpostgres.NewDecisionRepository(pool),
		scenarios,
		sessions,
		trainingService,
		bus,
		usersService,
		deepseek.New(deepseek.Config{
			BaseURL:    cfg.DeepSeek.BaseURL,
			APIKey:     cfg.DeepSeek.APIKey,
			Model:      cfg.DeepSeek.Model,
			Timeout:    cfg.DeepSeek.Timeout,
			MaxRetries: cfg.DeepSeek.MaxRetries,
		}),
		chatusecase.Options{},
	)
	chatrest.NewHandler(chatService, bus).Register(protected)

	usersrest.New(usersService, logger).Register(protected)

	server := &http.Server{
		Addr:    cfg.HTTP.Addr,
		Handler: http.StripPrefix(cfg.HTTP.Prefix, mux),
	}

	return serve(ctx, server, cfg.HTTP.ShutdownTimeout, bus.Close)
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

func newScenarioSource(
	ctx context.Context,
	cfg config.ScenariosConfig,
	logger *zap.Logger,
) (*scenarioyaml.YAMLRepository, error) {
	graphs, err := scenarioyaml.NewYAMLRepository(cfg.Dir)
	if err != nil {
		return nil, fmt.Errorf("failed to load scenarios: %w", err)
	}

	loaded, err := graphs.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list scenarios: %w", err)
	}

	logger.Info("scenario graphs loaded",
		zap.String("dir", cfg.Dir),
		zap.Int("scenarios", len(loaded)),
	)

	return graphs, nil
}

func serve(
	ctx context.Context,
	server *http.Server,
	shutdownTimeout time.Duration,
	beforeShutdown func(),
) error {
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

	beforeShutdown()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shut down http server: %w", err)
	}

	return nil
}
