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
