package users

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/users/entity"
	"go.uber.org/zap"
)

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		userIDStr = r.URL.Query().Get("user_id")
	}

	if userIDStr == "" {
		h.logger.Warn("get user failed: missing user_id")
		h.respondError(w, http.StatusBadRequest, "user_id header or query param is required")
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		h.logger.Warn("get user failed: invalid user_id format", zap.String("raw_id", userIDStr))
		h.respondError(w, http.StatusBadRequest, "user_id must be a valid positive integer")
		return
	}

	user, err := h.usersService.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			h.respondError(w, http.StatusNotFound, "user not found")
			return
		}
		h.logger.Error("failed to get user", zap.Int64("id", userID), zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}
