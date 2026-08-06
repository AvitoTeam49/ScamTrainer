package main

import (
	"context"
	"fmt"

	scenariodomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/scenario"
	chatusecase "github.com/AvitoTeam49/ScamTrainer/backend/internal/usecase/chat"
)

var _ chatusecase.ScenarioSource = (*graphScenarioSource)(nil)

// graphScenarioSource resolves the scenario id a chat stores to the scenario
// graph that drives the conversation and its scoring.
//
// The two domains disagree on how a scenario is identified: chat persists
// scenario_id as bigint, while the graphs are keyed by the string ids declared
// in their YAML files. The translation table comes from configuration and is
// resolved here, at the composition root, leaving both domains untouched.
type graphScenarioSource struct {
	scenarios scenariodomain.Repository
	ids       map[int64]string
}

func newGraphScenarioSource(
	scenarios scenariodomain.Repository,
	ids map[int64]string,
) *graphScenarioSource {
	return &graphScenarioSource{scenarios: scenarios, ids: ids}
}

func (s *graphScenarioSource) Scenario(
	ctx context.Context,
	scenarioID int64,
) (*scenariodomain.Scenario, error) {
	graphID, ok := s.ids[scenarioID]
	if !ok {
		return nil, fmt.Errorf("%w: scenario_id %d is not mapped", scenariodomain.ErrScenarioNotFound, scenarioID)
	}

	return s.scenarios.GetByID(ctx, graphID)
}

// verify fails fast when SCENARIOS_MAP points at a scenario that no YAML file
// declares, so a typo surfaces at startup instead of mid-conversation.
func (s *graphScenarioSource) verify(ctx context.Context) error {
	for scenarioID, graphID := range s.ids {
		if _, err := s.scenarios.GetByID(ctx, graphID); err != nil {
			return fmt.Errorf("scenario_id %d -> %q: %w", scenarioID, graphID, err)
		}
	}

	return nil
}
