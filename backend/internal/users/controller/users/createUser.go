package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/middleware"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/users/entity"
	"go.uber.org/zap"
)

type createUserRequest struct {
	Username string `json:"username"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		h.logger.Warn("create user failed: unauthorized or missing user_id in context")
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode create user request", zap.Error(err))
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	username := strings.TrimSpace(req.Username)
	if username == "" {
		h.logger.Warn("create user validation failed: empty username")
		h.respondError(w, http.StatusBadRequest, "username is required")
		return
	}

	user, err := h.usersService.CreateUser(ctx, userID, username)
	if err != nil {
		if errors.Is(err, entity.ErrUserAlreadyExists) {
			h.logger.Warn("user already exists", zap.String("username", username))
			h.respondError(w, http.StatusConflict, "user with this username already exists")
			return
		}
		h.logger.Error("failed to create user", zap.String("username", username), zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.respondJSON(w, http.StatusCreated, user)
}
