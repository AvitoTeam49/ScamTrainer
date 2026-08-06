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

func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/users", a.userServer.CreateUser)
	mux.HandleFunc("GET /v1/users/me", a.userServer.GetUser)
	mux.HandleFunc("GET /v1/users/me/progress", a.userServer.GetUserProgress)
	mux.HandleFunc("GET /v1/users/leaderboard", a.userServer.GetLeaderboard)
}
