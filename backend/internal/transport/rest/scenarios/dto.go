package scenariosrest

import (
	scenariodomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/scenario"
)

type errorResponse struct {
	Error string `json:"error"`
}

type scenarioResponse struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Role       string `json:"role"`
	Difficulty int    `json:"difficulty"`
}

type scenariosResponse struct {
	Items []scenarioResponse `json:"items"`
}

func scenarioFrom(scenario scenariodomain.ScenarioInfo) scenarioResponse {
	return scenarioResponse{
		ID:         scenario.ID,
		Title:      scenario.Title,
		Role:       string(scenario.Role),
		Difficulty: int(scenario.Difficulty),
	}
}
