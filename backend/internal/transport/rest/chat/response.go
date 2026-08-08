package chatrest

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	chatdomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/middleware"
)

const (
	defaultLimit = 50
	maxLimit     = 100
)

var (
	errInvalidRequest = errors.New("invalid request")
	errUnauthorized   = errors.New("unauthorized")
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to write chat response", "error", err)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errUnauthorized):
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: err.Error()})
	case errors.Is(err, chatdomain.ErrChatNotFound), errors.Is(err, chatdomain.ErrScenarioNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, chatdomain.ErrChatAccessDenied):
		writeJSON(w, http.StatusForbidden, errorResponse{Error: err.Error()})
	case errors.Is(err, chatdomain.ErrChatFinished), errors.Is(err, chatdomain.ErrOwnerNotFound):
		writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})
	case errors.Is(err, errInvalidRequest),
		errors.Is(err, chatdomain.ErrInvalidCursor),
		errors.Is(err, chatdomain.ErrMessageEmpty):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		slog.ErrorContext(r.Context(), "chat request failed", "method", r.Method, "path", r.URL.Path, "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
	}
}

func decodeJSON(r *http.Request, target any) error {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return fmt.Errorf("%w: malformed json body", errInvalidRequest)
	}

	return nil
}

func pathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%w: %s must be a positive integer", errInvalidRequest, name)
	}

	return id, nil
}

func chatScope(r *http.Request) (chatID, userID int64, err error) {
	chatID, err = pathID(r, "chatID")
	if err != nil {
		return 0, 0, err
	}

	userID, err = requestUserID(r)
	if err != nil {
		return 0, 0, err
	}

	return chatID, userID, nil
}

func requestUserID(r *http.Request) (int64, error) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		return 0, errUnauthorized
	}

	return userID, nil
}

func cursorFrom(r *http.Request) (chatdomain.Cursor, error) {
	limit, err := intQuery(r, "limit", defaultLimit)
	if err != nil {
		return chatdomain.Cursor{}, err
	}

	if limit <= 0 || limit > maxLimit {
		return chatdomain.Cursor{}, fmt.Errorf("%w: limit must be between 1 and %d", errInvalidRequest, maxLimit)
	}

	afterID, err := intQuery(r, "after_id", 0)
	if err != nil {
		return chatdomain.Cursor{}, err
	}

	if afterID < 0 {
		return chatdomain.Cursor{}, fmt.Errorf("%w: after_id must not be negative", errInvalidRequest)
	}

	return chatdomain.Cursor{Limit: int(limit), AfterID: afterID}, nil
}

func intQuery(r *http.Request, name string, fallback int64) (int64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s must be an integer", errInvalidRequest, name)
	}

	return value, nil
}
