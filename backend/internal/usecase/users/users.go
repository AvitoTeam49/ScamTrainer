package usersusecase

import (
	"context"
	"fmt"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/users"
	"go.uber.org/zap"
)

//go:generate mockgen -source=users.go -destination=mocks/users_mocks.go -package=mocks
type (
	Repository interface {
		CreateUser(ctx context.Context, userID int64, username string) (*usersdomain.User, error)
		GetUserByID(ctx context.Context, id int64) (*usersdomain.User, error)
		GetUserProgress(ctx context.Context, userID int64) (*usersdomain.UserProgress, error)
		UpdateUserScore(ctx context.Context, userID int64, scoreDelta int) error
		GetLeaderboard(ctx context.Context, limit, offset int) ([]usersdomain.LeaderboardEntry, int, error)
	}
)

type userService struct {
	repo   Repository
	logger *zap.Logger
}

func NewUserService(repo Repository, logger *zap.Logger) *userService {
	return &userService{
		repo:   repo,
		logger: logger,
	}
}

func (s *userService) CreateUser(ctx context.Context, userID int64, username string) (*usersdomain.User, error) {
	user, err := s.repo.CreateUser(ctx, userID, username)
	if err != nil {
		return nil, fmt.Errorf("create user repo err: %w", err)
	}

	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, userID int64) (*usersdomain.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id repo err: %w", err)
	}
	return user, nil
}

func (s *userService) GetUserProgress(ctx context.Context, userID int64) (*usersdomain.UserProgress, error) {
	progress, err := s.repo.GetUserProgress(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user progress repo err: %w", err)
	}

	return progress, nil
}

func (s *userService) UpdateUserScore(ctx context.Context, userID int64, scoreDelta int) error {
	if err := s.repo.UpdateUserScore(ctx, userID, scoreDelta); err != nil {
		return fmt.Errorf("update user score repo err: %w", err)
	}

	return nil
}

func (s *userService) GetLeaderboard(ctx context.Context, limit, offset int) (*usersdomain.Leaderboard, error) {
	entries, totalUsers, err := s.repo.GetLeaderboard(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get leaderboard repo err: %w", err)
	}

	for i := range entries {
		entries[i].Rank = offset + i + 1
	}

	return &usersdomain.Leaderboard{
		TopUsers:   entries,
		TotalUsers: totalUsers,
	}, nil
}
