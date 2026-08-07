package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/auth/entity"
	"go.uber.org/zap"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

func (a *authHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, message, err := a.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrUserAlreadyExists):
			a.respondError(w, http.StatusConflict, "user already exists")
			return
		case errors.Is(err, entity.ErrInvalidPassword):
			a.respondError(w, http.StatusBadRequest, "password is too short or invalid")
			return
		case errors.Is(err, entity.ErrInvalidEmail):
			a.respondError(w, http.StatusBadRequest, "invalid email format")
			return
		default:
			a.logger.Error(
				"failed to register user",
				zap.Error(err),
			)
			a.respondError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	a.respondJSON(w, http.StatusCreated, RegisterResponse{
		UserID:  userID,
		Message: message,
	})
}
