package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/domain"
)

const (
	defaultLimit = 50
	maxLimit     = 100
)

var errInvalidRequest = errors.New("invalid request")

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to write chat response", "error", err)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrChatNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrChatAccessDenied):
		writeJSON(w, http.StatusForbidden, errorResponse{Error: err.Error()})
	case errors.Is(err, errInvalidRequest), errors.Is(err, domain.ErrInvalidCursor):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		slog.ErrorContext(r.Context(), "chat request failed", "method", r.Method, "path", r.URL.Path, "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
	}
}

func pathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%w: %s must be a positive integer", errInvalidRequest, name)
	}

	return id, nil
}

func cursorFrom(r *http.Request) (domain.Cursor, error) {
	limit, err := intQuery(r, "limit", defaultLimit)
	if err != nil {
		return domain.Cursor{}, err
	}

	if limit <= 0 || limit > maxLimit {
		return domain.Cursor{}, fmt.Errorf("%w: limit must be between 1 and %d", errInvalidRequest, maxLimit)
	}

	afterID, err := intQuery(r, "after_id", 0)
	if err != nil {
		return domain.Cursor{}, err
	}

	if afterID < 0 {
		return domain.Cursor{}, fmt.Errorf("%w: after_id must not be negative", errInvalidRequest)
	}

	return domain.Cursor{Limit: int(limit), AfterID: afterID}, nil
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
