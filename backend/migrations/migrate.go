package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

const migrationsDir = "migrations"

func Up(ctx context.Context, pool *pgxpool.Pool) error {
	return withGoose(ctx, pool, goose.UpContext)
}

func Down(ctx context.Context, pool *pgxpool.Pool) error {
	return withGoose(ctx, pool, goose.DownContext)
}

func withGoose(
	ctx context.Context,
	pool *pgxpool.Pool,
	action func(ctx context.Context, db *sql.DB, dir string, opts ...goose.OptionsFunc) error,
) error {
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer func() {
		_ = sqlDB.Close()
	}()

	if err := action(ctx, sqlDB, migrationsDir); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
