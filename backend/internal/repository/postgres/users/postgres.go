package userspostgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/users"
	userssqlc "github.com/AvitoTeam49/ScamTrainer/backend/internal/repository/postgres/users/sqlc"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	userssqlc.DBTX
}

type postgresRepository struct {
	queries userssqlc.Querier
	db      DB
}

func NewPostgresRepository(qdb DB) *postgresRepository {
	return &postgresRepository{
		queries: userssqlc.New(qdb),
		db:      qdb,
	}
}

func (r *postgresRepository) CreateUser(ctx context.Context, userID int64, username string) (*usersdomain.User, error) {
	u, err := r.queries.CreateUser(ctx, userssqlc.CreateUserParams{
		ID:       userID,
		Username: username,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, usersdomain.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user query: %w", err)
	}

	return &usersdomain.User{
		ID:        u.ID,
		Username:  u.Username,
		Score:     int(u.Score),
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}, nil
}

func (r *postgresRepository) GetUserByID(ctx context.Context, id int64) (*usersdomain.User, error) {
	u, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usersdomain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id query: %w", err)
	}

	return &usersdomain.User{
		ID:        u.ID,
		Username:  u.Username,
		Score:     int(u.Score),
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}, nil
}

func (r *postgresRepository) GetUserProgress(ctx context.Context, userID int64) (*usersdomain.UserProgress, error) {
	p, err := r.queries.GetUserProgress(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usersdomain.ErrProgressNotFound
		}
		return nil, fmt.Errorf("failed to get user progress query: %w", err)
	}

	return &usersdomain.UserProgress{
		UserID:             p.UserID,
		ScenariosCompleted: int(p.ScenariosCompleted),
		ScamsDetected:      int(p.ScamsDetected),
		FailedAttempts:     int(p.FailedAttempts),
	}, nil
}

func (r *postgresRepository) UpdateUserScore(ctx context.Context, userID int64, scoreDelta int) error {
	_, err := r.queries.UpdateUserScore(ctx, userssqlc.UpdateUserScoreParams{
		ScoreDelta: int32(scoreDelta),
		ID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usersdomain.ErrUserNotFound
		}
		return fmt.Errorf("failed to update user score query: %w", err)
	}

	return nil
}

func (r *postgresRepository) GetLeaderboard(ctx context.Context, limit, offset int) ([]usersdomain.LeaderboardEntry, int, error) {
	totalUsers, err := r.queries.CountUsers(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users query: %w", err)
	}

	rows, err := r.queries.GetLeaderboard(ctx, userssqlc.GetLeaderboardParams{
		Lim:  int32(limit),
		Offs: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get leaderboard query: %w", err)
	}

	entries := make([]usersdomain.LeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, usersdomain.LeaderboardEntry{
			UserID:   row.ID,
			Username: row.Username,
			Score:    int(row.Score),
		})
	}

	return entries, int(totalUsers), nil
}
