package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/users/entity"
	sqlcUsers "github.com/AvitoTeam49/ScamTrainer/backend/internal/users/repository/postgres/sqlc"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	sqlcUsers.DBTX
}

type postgresRepository struct {
	queries sqlcUsers.Querier
	db      DB
}

func NewPostgresRepository(qdb DB) *postgresRepository {
	return &postgresRepository{
		queries: sqlcUsers.New(qdb),
		db:      qdb,
	}
}

func (r *postgresRepository) CreateUser(ctx context.Context, username string) (*entity.User, error) {
	u, err := r.queries.CreateUser(ctx, username)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, entity.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user query: %w", err)
	}

	return &entity.User{
		ID:        u.ID,
		Username:  u.Username,
		Score:     int(u.Score),
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}, nil
}

func (r *postgresRepository) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	u, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id query: %w", err)
	}

	return &entity.User{
		ID:        u.ID,
		Username:  u.Username,
		Score:     int(u.Score),
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}, nil
}

func (r *postgresRepository) GetUserProgress(ctx context.Context, userID int64) (*entity.UserProgress, error) {
	p, err := r.queries.GetUserProgress(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrProgressNotFound
		}
		return nil, fmt.Errorf("failed to get user progress query: %w", err)
	}

	return &entity.UserProgress{
		UserID:             p.UserID,
		ScenariosCompleted: int(p.ScenariosCompleted),
		ScamsDetected:      int(p.ScamsDetected),
		FailedAttempts:     int(p.FailedAttempts),
	}, nil
}

func (r *postgresRepository) UpdateUserScore(ctx context.Context, userID int64, scoreDelta int) error {
	_, err := r.queries.UpdateUserScore(ctx, sqlcUsers.UpdateUserScoreParams{
		ScoreDelta: int32(scoreDelta),
		ID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ErrUserNotFound
		}
		return fmt.Errorf("failed to update user score query: %w", err)
	}

	return nil
}

func (r *postgresRepository) GetLeaderboard(ctx context.Context, limit, offset int) ([]entity.LeaderboardEntry, int, error) {
	totalUsers, err := r.queries.CountUsers(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users query: %w", err)
	}

	rows, err := r.queries.GetLeaderboard(ctx, sqlcUsers.GetLeaderboardParams{
		Lim:  int32(limit),
		Offs: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get leaderboard query: %w", err)
	}

	entries := make([]entity.LeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, entity.LeaderboardEntry{
			UserID:   row.ID,
			Username: row.Username,
			Score:    int(row.Score),
		})
	}

	return entries, int(totalUsers), nil
}
