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
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/repository/postgres"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/scenario"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/service"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/transport/rest"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
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

	chatService := service.New(
		postgres.NewChatRepository(pool),
		postgres.NewMessageRepository(pool),
		postgres.NewIncidentRepository(pool),
		scenario.NewStaticProvider(""),
		deepseek.New(deepseek.Config{
			BaseURL:    cfg.DeepSeek.BaseURL,
			APIKey:     cfg.DeepSeek.APIKey,
			Model:      cfg.DeepSeek.Model,
			Timeout:    cfg.DeepSeek.Timeout,
			MaxRetries: cfg.DeepSeek.MaxRetries,
		}),
		service.Options{},
	)

	mux := http.NewServeMux()
	rest.NewHandler(chatService).Register(mux)

	server := &http.Server{
		Addr:    cfg.HTTP.Addr,
		Handler: http.StripPrefix(cfg.HTTP.Prefix, mux),
	}

	return serve(ctx, server, cfg.HTTP.ShutdownTimeout)
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
