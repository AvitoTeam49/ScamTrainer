package controller

import (
	"net/http"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/auth/controller/auth"
	"go.uber.org/zap"
)

type AuthHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	ValidateToken(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
}

type API struct {
	authServer AuthHandler
	logger     *zap.Logger
}

func New(authService auth.AuthService, logger *zap.Logger) *API {
	return &API{
		authServer: auth.NewAuthHandler(authService, logger),
		logger:     logger,
	}
}

func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/auth/register", a.authServer.Register)
	mux.HandleFunc("POST /v1/auth/login", a.authServer.Login)
	mux.HandleFunc("POST /v1/auth/validate", a.authServer.ValidateToken)
	mux.HandleFunc("POST /v1/auth/refresh", a.authServer.RefreshToken)
}
