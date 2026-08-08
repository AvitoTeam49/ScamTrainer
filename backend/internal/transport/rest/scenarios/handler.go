package scenariosrest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	scenariodomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/scenario"
)

var errInvalidRequest = errors.New("invalid request")

type Handler struct {
	catalog scenariodomain.ScenarioCatalog
}

func NewHandler(catalog scenariodomain.ScenarioCatalog) *Handler {
	return &Handler{catalog: catalog}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/scenarios", h.listScenarios)
}

func (h *Handler) listScenarios(w http.ResponseWriter, r *http.Request) {
	scenarios, err := h.list(r.Context(), r.URL.Query().Get("difficulty"))
	if err != nil {
		writeError(w, r, err)
		return
	}

	items := make([]scenarioResponse, 0, len(scenarios))
	for _, scenario := range scenarios {
		items = append(items, scenarioFrom(scenario))
	}

	writeJSON(w, http.StatusOK, scenariosResponse{Items: items})
}

func (h *Handler) list(ctx context.Context, rawDifficulty string) ([]scenariodomain.ScenarioInfo, error) {
	if rawDifficulty == "" {
		return h.catalog.List(ctx)
	}

	difficulty, err := parseDifficulty(rawDifficulty)
	if err != nil {
		return nil, err
	}

	return h.catalog.ListByDifficulty(ctx, difficulty)
}

func parseDifficulty(raw string) (scenariodomain.Difficulty, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%w: difficulty must be an integer", errInvalidRequest)
	}

	difficulty := scenariodomain.Difficulty(value)
	switch difficulty {
	case scenariodomain.DifficultyEasy, scenariodomain.DifficultyMedium, scenariodomain.DifficultyHard:
		return difficulty, nil
	default:
		return 0, fmt.Errorf("%w: unknown difficulty %d", errInvalidRequest, value)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to write scenarios response", "error", err)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, errInvalidRequest) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	slog.ErrorContext(r.Context(), "scenarios request failed", "path", r.URL.Path, "error", err)
	writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
}
