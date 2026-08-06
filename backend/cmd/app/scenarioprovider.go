package main

import (
	"context"
	"fmt"

	scenariodomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/scenario"
	chatusecase "github.com/AvitoTeam49/ScamTrainer/backend/internal/usecase/chat"
)

var _ chatusecase.ScenarioSource = (*graphScenarioSource)(nil)

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

func (s *graphScenarioSource) verify(ctx context.Context) error {
	for scenarioID, graphID := range s.ids {
		if _, err := s.scenarios.GetByID(ctx, graphID); err != nil {
			return fmt.Errorf("scenario_id %d -> %q: %w", scenarioID, graphID, err)
		}
	}

	return nil
}
