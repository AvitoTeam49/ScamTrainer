package controller

import (
	"net/http"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/users/controller/users"
	"go.uber.org/zap"
)

type UserHandler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
	GetUserProgress(w http.ResponseWriter, r *http.Request)
	GetLeaderboard(w http.ResponseWriter, r *http.Request)
}

type API struct {
	userServer UserHandler
	logger     *zap.Logger
}

func New(userService users.UsersService, logger *zap.Logger) *API {
	return &API{
		userServer: users.NewUserHandler(userService, logger),
		logger:     logger,
	}
}

func Register(mux *http.ServeMux, h UserHandler) {
	mux.HandleFunc("POST /api/v1/users", h.CreateUser)
	mux.HandleFunc("GET /api/v1/users/me", h.GetUser)
	mux.HandleFunc("GET /api/v1/users/me/progress", h.GetUserProgress)
	mux.HandleFunc("GET /api/v1/users/leaderboard", h.GetLeaderboard)
}
