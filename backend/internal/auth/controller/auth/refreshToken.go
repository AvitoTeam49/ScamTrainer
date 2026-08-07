package auth

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (a *authHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.logger.Warn("invalid refresh request body", zap.Error(err))
		a.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		a.respondError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	newAccess, newRefresh, err := a.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		a.logger.Warn("failed to refresh token", zap.Error(err))
		a.respondError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	a.respondJSON(w, http.StatusOK, RefreshResponse{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
	})
}
