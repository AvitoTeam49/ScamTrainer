package usersrest

import (
	"fmt"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

const (
	defaultLeaderboardLimit = 10
	maxLeaderboardLimit     = 100
)

func (h *UserHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	limit := defaultLeaderboardLimit
	if limitStr := query.Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			h.logger.Warn("get leaderboard failed: invalid limit format", zap.String("limit", limitStr), zap.Error(err))
			h.respondError(w, http.StatusBadRequest, "limit must be a valid integer")
			return
		}
		if parsedLimit <= 0 {
			h.logger.Warn("get leaderboard failed: limit must be positive", zap.Int("limit", parsedLimit))
			h.respondError(w, http.StatusBadRequest, "limit must be greater than 0")
			return
		}
		if parsedLimit > maxLeaderboardLimit {
			h.logger.Warn("get leaderboard failed: limit is too large", zap.Int("limit", parsedLimit))
			h.respondError(w, http.StatusBadRequest, fmt.Sprintf("limit must not exceed %d", maxLeaderboardLimit))
			return
		}
		limit = parsedLimit
	}

	offset := 0
	if offsetStr := query.Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			h.logger.Warn("get leaderboard failed: invalid offset format", zap.String("offset", offsetStr), zap.Error(err))
			h.respondError(w, http.StatusBadRequest, "offset must be a valid integer")
			return
		}
		if parsedOffset < 0 {
			h.logger.Warn("get leaderboard failed: offset cannot be negative", zap.Int("offset", parsedOffset))
			h.respondError(w, http.StatusBadRequest, "offset must be non-negative")
			return
		}
		offset = parsedOffset
	}

	leaderboard, err := h.usersService.GetLeaderboard(ctx, limit, offset)
	if err != nil {
		h.logger.Error("failed to get leaderboard", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.respondJSON(w, http.StatusOK, leaderboard)
}
