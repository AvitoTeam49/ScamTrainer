package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/auth/controller/auth"
)

type contextKey string

const UserIDKey contextKey = "userID"

func Auth(authService auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			tokenStr, found := strings.CutPrefix(authHeader, "Bearer ")
			if !found || tokenStr == "" {
				http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			isValid, userID, _, err := authService.ValidateToken(r.Context(), tokenStr)
			if err != nil || !isValid {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}
