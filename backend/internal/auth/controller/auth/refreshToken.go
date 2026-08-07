package auth

import (
	"net/http"

	"go.uber.org/zap"
)

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

func (a *authHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		a.logger.Warn("refresh token cookie missing", zap.Error(err))
		a.respondError(w, http.StatusUnauthorized, "refresh token cookie missing")
		return
	}

	refreshToken := cookie.Value
	if refreshToken == "" {
		a.respondError(w, http.StatusUnauthorized, "refresh_token is empty")
		return
	}

	newAccess, newRefresh, err := a.authService.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		a.logger.Warn("failed to refresh token", zap.Error(err))
		a.respondError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefresh,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	a.respondJSON(w, http.StatusOK, RefreshResponse{
		AccessToken: newAccess,
	})
}
