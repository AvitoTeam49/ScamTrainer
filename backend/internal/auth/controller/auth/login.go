package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/auth/entity"
	"go.uber.org/zap"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

func (a *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		a.respondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	accessToken, refreshToken, err := a.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidCredentials):
			a.respondError(w, http.StatusUnauthorized, "invalid email or password")
			return
		default:
			a.logger.Error("failed to login user", zap.Error(err))
			a.respondError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	setAccessCookie(w, accessToken)
	a.respondJSON(w, http.StatusOK, LoginResponse{
		AccessToken: accessToken,
	})
}
