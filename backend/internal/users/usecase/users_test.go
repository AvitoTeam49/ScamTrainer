package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/users/entity"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/users/usecase/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestUserService_CreateUser_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	repo := mocks.NewMockRepository(ctrl)
	logger := zap.NewNop()

	srv := NewUserService(repo, logger)

	userID := int64(10)
	username := "testuser"
	expectedUser := &entity.User{
		ID:       userID,
		Username: username,
	}

	repo.EXPECT().
		CreateUser(gomock.Any(), userID, username).
		Return(expectedUser, nil)

	// Act
	user, err := srv.CreateUser(context.Background(), username, userID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, expectedUser, user)
}

func TestUserService_CreateUser_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mocks.NewMockRepository(ctrl)
	logger := zap.NewNop()

	srv := NewUserService(repo, logger)

	userID := int64(10)
	username := "testuser"
	repoErr := errors.New("db insert error")

	repo.EXPECT().
		CreateUser(gomock.Any(), userID, username).
		Return(nil, repoErr)

	user, err := srv.CreateUser(context.Background(), username, userID)

	require.Error(t, err)
	require.Nil(t, user)
	require.ErrorIs(t, err, repoErr)
}

func TestUserService_GetUserByID_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mocks.NewMockRepository(ctrl)
	logger := zap.NewNop()

	srv := NewUserService(repo, logger)

	userID := int64(42)
	expectedUser := &entity.User{
		ID:       userID,
		Username: "john_doe",
	}

	repo.EXPECT().
		GetUserByID(gomock.Any(), userID).
		Return(expectedUser, nil)

	user, err := srv.GetUserByID(context.Background(), userID)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, expectedUser, user)
}

func TestUserService_GetUserByID_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mocks.NewMockRepository(ctrl)
	logger := zap.NewNop()

	srv := NewUserService(repo, logger)

	userID := int64(42)
	repoErr := errors.New("user not found")

	repo.EXPECT().
		GetUserByID(gomock.Any(), userID).
		Return(nil, repoErr)

	user, err := srv.GetUserByID(context.Background(), userID)

	require.Error(t, err)
	require.Nil(t, user)
	require.ErrorIs(t, err, repoErr)
}

func TestUserService_GetUserProgress_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mocks.NewMockRepository(ctrl)
	logger := zap.NewNop()

	srv := NewUserService(repo, logger)

	userID := int64(1)
	expectedProgress := &entity.UserProgress{
		UserID:             userID,
		ScenariosCompleted: 10,
		ScamsDetected:      8,
		FailedAttempts:     2,
	}

	repo.EXPECT().
		GetUserProgress(gomock.Any(), userID).
		Return(expectedProgress, nil)

	progress, err := srv.GetUserProgress(context.Background(), userID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, progress)
	require.Equal(t, expectedProgress, progress)
}
func TestUserService_GetUserProgress_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mocks.NewMockRepository(ctrl)
	logger := zap.NewNop()

	srv := NewUserService(repo, logger)

	userID := int64(1)
	repoErr := errors.New("progress not found")

	repo.EXPECT().
		GetUserProgress(gomock.Any(), userID).
		Return(nil, repoErr)

	progress, err := srv.GetUserProgress(context.Background(), userID)

	require.Error(t, err)
	require.Nil(t, progress)
	require.ErrorIs(t, err, repoErr)
}

func TestUserService_UpdateUserScore_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mocks.NewMockRepository(ctrl)
	logger := zap.NewNop()

	srv := NewUserService(repo, logger)

	userID := int64(99)
	scoreDelta := 15

	repo.EXPECT().
		UpdateUserScore(gomock.Any(), userID, scoreDelta).
		Return(nil)

	err := srv.UpdateUserScore(context.Background(), userID, scoreDelta)

	require.NoError(t, err)
}

func TestUserService_UpdateUserScore_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mocks.NewMockRepository(ctrl)
	logger := zap.NewNop()

	srv := NewUserService(repo, logger)

	userID := int64(99)
	scoreDelta := 15
	repoErr := errors.New("failed to update score")

	repo.EXPECT().
		UpdateUserScore(gomock.Any(), userID, scoreDelta).
		Return(repoErr)

	err := srv.UpdateUserScore(context.Background(), userID, scoreDelta)

	require.Error(t, err)
	require.ErrorIs(t, err, repoErr)
}

func TestUserService_GetLeaderboard_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mocks.NewMockRepository(ctrl)
	logger := zap.NewNop()

	srv := NewUserService(repo, logger)

	limit := 10
	offset := 5
	totalUsers := 100

	mockEntries := []entity.LeaderboardEntry{
		{UserID: 1, Score: 500},
		{UserID: 2, Score: 450},
	}

	repo.EXPECT().
		GetLeaderboard(gomock.Any(), limit, offset).
		Return(mockEntries, totalUsers, nil)

	// Act
	leaderboard, err := srv.GetLeaderboard(context.Background(), limit, offset)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, leaderboard)
	require.Equal(t, totalUsers, leaderboard.TotalUsers)
	require.Len(t, leaderboard.TopUsers, 2)

	require.Equal(t, 6, leaderboard.TopUsers[0].Rank)
	require.Equal(t, int64(1), leaderboard.TopUsers[0].UserID)

	require.Equal(t, 7, leaderboard.TopUsers[1].Rank)
	require.Equal(t, int64(2), leaderboard.TopUsers[1].UserID)
}

func TestUserService_GetLeaderboard_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mocks.NewMockRepository(ctrl)
	logger := zap.NewNop()

	srv := NewUserService(repo, logger)

	limit := 10
	offset := 0
	repoErr := errors.New("db error on leaderboard fetch")

	repo.EXPECT().
		GetLeaderboard(gomock.Any(), limit, offset).
		Return(nil, 0, repoErr)

	leaderboard, err := srv.GetLeaderboard(context.Background(), limit, offset)

	require.Error(t, err)
	require.Nil(t, leaderboard)
	require.ErrorIs(t, err, repoErr)
}
