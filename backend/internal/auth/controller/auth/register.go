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
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID, message, err := a.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrUserAlreadyExists):
			http.Error(w, "user already exists", http.StatusConflict)
			return
		case errors.Is(err, entity.ErrInvalidPassword):
			http.Error(w, "password is too short or invalid", http.StatusBadRequest)
			return
		case errors.Is(err, entity.ErrInvalidEmail):
			http.Error(w, "invalid email format", http.StatusBadRequest)
			return
		default:
			a.logger.Error(
				"failed to register user",
				zap.Error(err),
			)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(RegisterResponse{
		UserID:  userID,
		Message: message,
	}); err != nil {
		a.logger.Error("failed to encode response json",
			zap.Error(err),
		)
	}
}
