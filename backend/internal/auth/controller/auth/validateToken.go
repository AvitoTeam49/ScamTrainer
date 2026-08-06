package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/auth/entity"
	"go.uber.org/zap"
)

type ValidateTokenRequest struct {
	Token string `json:"token"`
}

type ValidateTokenResponse struct {
	IsValid bool   `json:"is_valid"`
	UserID  int64  `json:"user_id"`
	Role    string `json:"role"`
}

func (a *authHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var token string

	authHeader := r.Header.Get("Authorization")
	if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
		token = after
	}

	if token == "" && r.Body != nil {
		var req ValidateTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request payload", http.StatusBadRequest)
			return
		}
		token = req.Token
	}

	if token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	isValid, userID, role, err := a.authService.ValidateToken(r.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidToken):
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		default:
			a.logger.Error(
				"failed to validate token",
				zap.Error(err),
			)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(ValidateTokenResponse{
		IsValid: isValid,
		UserID:  userID,
		Role:    role,
	}); err != nil {
		a.logger.Error("failed to encode response json",
			zap.Error(err),
		)
	}
}
