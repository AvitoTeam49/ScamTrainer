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
			tokenStr, found := bearerToken(r)
			if !found {
				tokenStr, found = cookieToken(r)
			}

			if !found || tokenStr == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"missing or invalid authorization header"}`))
				return
			}

			isValid, userID, err := authService.ValidateToken(r.Context(), tokenStr)
			if err != nil || !isValid {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"invalid or expired token"}`))
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

func bearerToken(r *http.Request) (string, bool) {
	return strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
}

// EventSource не умеет отправлять заголовок Authorization, поэтому SSE авторизуется кукой.
func cookieToken(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(auth.AccessTokenCookie)
	if err != nil {
		return "", false
	}

	return cookie.Value, cookie.Value != ""
}
