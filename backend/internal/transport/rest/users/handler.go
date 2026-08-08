package usersrest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/users"
	"go.uber.org/zap"
)

type UsersService interface {
	CreateUser(ctx context.Context, userID int64, username string) (*usersdomain.User, error)

	GetUserByID(ctx context.Context, userID int64) (*usersdomain.User, error)

	GetUserProgress(ctx context.Context, userID int64) (*usersdomain.UserProgress, error)

	UpdateUserScore(ctx context.Context, userID int64, scoreDelta int) error

	GetLeaderboard(ctx context.Context, limit, offset int) (*usersdomain.Leaderboard, error)
}

type UserHandler struct {
	usersService UsersService
	logger       *zap.Logger
}

func NewUserHandler(
	usersService UsersService,
	logger *zap.Logger,
) *UserHandler {
	return &UserHandler{
		usersService: usersService,
		logger:       logger,
	}
}

func (h *UserHandler) respondJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			h.logger.Error("failed to encode response json",
				zap.Int("status_code", code),
				zap.Error(err),
			)
		}
	}
}

func (h *UserHandler) respondError(w http.ResponseWriter, code int, message string) {
	h.respondJSON(w, code, errorResponse{Error: message})
}
