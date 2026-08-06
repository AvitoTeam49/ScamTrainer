package usersrest

import (
	"net/http"

	"go.uber.org/zap"
)

type API struct {
	userServer *UserHandler
	logger     *zap.Logger
}

func New(userService UsersService, logger *zap.Logger) *API {
	return &API{
		userServer: NewUserHandler(userService, logger),
		logger:     logger,
	}
}

func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/users", a.userServer.CreateUser)
	mux.HandleFunc("GET /v1/users/me", a.userServer.GetUser)
	mux.HandleFunc("GET /v1/users/me/progress", a.userServer.GetUserProgress)
	mux.HandleFunc("GET /v1/users/leaderboard", a.userServer.GetLeaderboard)
}
