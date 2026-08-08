package usersrest

import (
	"errors"
	"net/http"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/users"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/middleware"
	"go.uber.org/zap"
)

func (h *UserHandler) GetUserProgress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		h.logger.Warn("create user failed: unauthorized or missing user_id in context")
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	progress, err := h.usersService.GetUserProgress(ctx, userID)
	if err != nil {
		if errors.Is(err, usersdomain.ErrProgressNotFound) || errors.Is(err, usersdomain.ErrUserNotFound) {
			h.respondError(w, http.StatusNotFound, "user progress not found")
			return
		}
		h.logger.Error("failed to get user progress", zap.Int64("user_id", userID), zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.respondJSON(w, http.StatusOK, progress)
}
