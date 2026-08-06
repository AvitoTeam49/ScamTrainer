package auth

import (
	"context"

	"go.uber.org/zap"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (userID int64, message string, err error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	ValidateToken(ctx context.Context, token string) (isValid bool, userID int64, role string, err error)
}

type authHandler struct {
	authService AuthService
	logger      *zap.Logger
}

func NewAuthHandler(
	authService AuthService,
	logger *zap.Logger,
) *authHandler {
	return &authHandler{
		authService: authService,
		logger:      logger,
	}
}
