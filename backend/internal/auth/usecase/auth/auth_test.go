package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/auth/entity"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/auth/usecase/auth/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	testJWTSecret = []byte("super_secret_key")
)

func TestAuthService_Register_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	authRepo := mocks.NewMockauthRepository(ctrl)
	logger := zap.NewNop()

	srv := NewAuthService(authRepo, logger, testJWTSecret)

	email := "test@example.com"
	password := "password123"
	expectedUserID := int64(42)

	authRepo.EXPECT().
		GetUserByEmail(gomock.Any(), email).
		Return(entity.User{}, entity.ErrUserNotFound)

	authRepo.EXPECT().
		CreateUser(gomock.Any(), email, gomock.Any()).
		Return(entity.User{ID: expectedUserID, Email: email}, nil)

	// Act
	userID, msg, err := srv.Register(context.Background(), email, password)

	// Assert
	require.NoError(t, err)
	require.Equal(t, expectedUserID, userID)
	require.Equal(t, "user successfully registered", msg)
}

func TestAuthService_Register_Err_Gomock(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("db connection error")

	tests := []struct {
		name        string
		email       string
		password    string
		mockSetup   func(authRepo *mocks.MockauthRepository)
		expectedErr error
	}{
		{
			name:        "invalid email",
			email:       "invalid-email",
			password:    "password123",
			mockSetup:   func(authRepo *mocks.MockauthRepository) {},
			expectedErr: entity.ErrInvalidEmail,
		},
		{
			name:        "short password",
			email:       "test@example.com",
			password:    "short",
			mockSetup:   func(authRepo *mocks.MockauthRepository) {},
			expectedErr: entity.ErrInvalidPassword,
		},
		{
			name:     "user already exists",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func(authRepo *mocks.MockauthRepository) {
				authRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(entity.User{ID: 1}, nil)
			},
			expectedErr: entity.ErrUserAlreadyExists,
		},
		{
			name:     "get user repo error",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func(authRepo *mocks.MockauthRepository) {
				authRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(entity.User{}, repoErr)
			},
			expectedErr: repoErr,
		},
		{
			name:     "create user repo error",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func(authRepo *mocks.MockauthRepository) {
				authRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(entity.User{}, entity.ErrUserNotFound)

				authRepo.EXPECT().
					CreateUser(gomock.Any(), "test@example.com", gomock.Any()).
					Return(entity.User{}, repoErr)
			},
			expectedErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			authRepo := mocks.NewMockauthRepository(ctrl)
			logger := zap.NewNop()

			srv := NewAuthService(authRepo, logger, testJWTSecret)

			tt.mockSetup(authRepo)

			userID, msg, err := srv.Register(context.Background(), tt.email, tt.password)

			require.Error(t, err)
			require.EqualValues(t, 0, userID)
			require.Empty(t, msg)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestAuthService_Login_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	authRepo := mocks.NewMockauthRepository(ctrl)
	logger := zap.NewNop()

	srv := NewAuthService(authRepo, logger, testJWTSecret)

	email := "test@example.com"
	password := "password123"
	expectedUserID := int64(777)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)

	authRepo.EXPECT().
		GetUserByEmail(gomock.Any(), email).
		Return(entity.User{ID: expectedUserID, Email: email, PasswordHash: string(hashedPassword)}, nil)

	// Act
	accessToken, refreshToken, err := srv.Login(context.Background(), email, password)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)
	require.NotEmpty(t, refreshToken)
}

func TestAuthService_Login_Err_Gomock(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("db error")

	validHash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)

	tests := []struct {
		name        string
		email       string
		password    string
		mockSetup   func(authRepo *mocks.MockauthRepository)
		expectedErr error
	}{
		{
			name:        "invalid input format",
			email:       "invalid",
			password:    "pass1234",
			mockSetup:   func(authRepo *mocks.MockauthRepository) {},
			expectedErr: entity.ErrInvalidCredentials,
		},
		{
			name:     "user not found",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func(authRepo *mocks.MockauthRepository) {
				authRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(entity.User{}, entity.ErrUserNotFound)
			},
			expectedErr: entity.ErrInvalidCredentials,
		},
		{
			name:     "repo error on get user",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func(authRepo *mocks.MockauthRepository) {
				authRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(entity.User{}, repoErr)
			},
			expectedErr: repoErr,
		},
		{
			name:     "wrong password",
			email:    "test@example.com",
			password: "wrong-password",
			mockSetup: func(authRepo *mocks.MockauthRepository) {
				authRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(entity.User{ID: 1, PasswordHash: string(validHash)}, nil)
			},
			expectedErr: entity.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			authRepo := mocks.NewMockauthRepository(ctrl)
			logger := zap.NewNop()

			srv := NewAuthService(authRepo, logger, testJWTSecret)

			tt.mockSetup(authRepo)

			accessToken, refreshToken, err := srv.Login(context.Background(), tt.email, tt.password)

			require.Error(t, err)
			require.Empty(t, accessToken)
			require.Empty(t, refreshToken)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestAuthService_ValidateToken_Success(t *testing.T) {
	t.Parallel()

	srv := NewAuthService(nil, zap.NewNop(), testJWTSecret)
	expectedUserID := int64(10)

	tokenString, err := srv.generateToken(expectedUserID, 15*time.Minute)
	require.NoError(t, err)

	// Act
	isValid, userID, err := srv.ValidateToken(context.Background(), tokenString)

	// Assert
	require.NoError(t, err)
	require.True(t, isValid)
	require.Equal(t, expectedUserID, userID)
}

func TestAuthService_ValidateToken_Err(t *testing.T) {
	t.Parallel()

	srv := NewAuthService(nil, zap.NewNop(), testJWTSecret)

	expiredToken, err := srv.generateToken(1, -1*time.Minute)
	require.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		expectedErr error
	}{
		{
			name:        "empty token",
			token:       "",
			expectedErr: entity.ErrInvalidToken,
		},
		{
			name:        "invalid token format",
			token:       "invalid.token.string",
			expectedErr: entity.ErrInvalidToken,
		},
		{
			name:        "expired token",
			token:       expiredToken,
			expectedErr: entity.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, userID, err := srv.ValidateToken(context.Background(), tt.token)

			require.Error(t, err)
			require.False(t, isValid)
			require.EqualValues(t, 0, userID)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	t.Parallel()

	srv := NewAuthService(nil, zap.NewNop(), testJWTSecret)
	expectedUserID := int64(99)

	oldRefreshToken, err := srv.generateToken(expectedUserID, 7*24*time.Hour)
	require.NoError(t, err)

	// Act
	newAccessToken, newRefreshToken, err := srv.RefreshToken(context.Background(), oldRefreshToken)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, newAccessToken)
	require.NotEmpty(t, newRefreshToken)
}
