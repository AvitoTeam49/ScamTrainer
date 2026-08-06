package main

import (
	"context"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/domain"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/scenario"
)

var _ domain.ScenarioProvider = (*graphScenarioProvider)(nil)

type graphScenarioProvider struct {
	scenarios scenario.Repository
	ids       map[int64]string
}

func newGraphScenarioProvider(
	scenarios scenario.Repository,
	ids map[int64]string,
) *graphScenarioProvider {
	return &graphScenarioProvider{scenarios: scenarios, ids: ids}
}

func (p *graphScenarioProvider) SystemPrompt(ctx context.Context, scenarioID int64) (string, error) {
	graphID, ok := p.ids[scenarioID]
	if !ok {
		return "", domain.ErrScenarioNotFound
	}

	found, err := p.scenarios.GetByID(ctx, graphID)
	if err != nil {
		return "", domain.ErrScenarioNotFound
	}

	return found.LLM.CharacterPrompt, nil
}
