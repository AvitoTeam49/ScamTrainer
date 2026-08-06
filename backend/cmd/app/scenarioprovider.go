package main

import (
	"context"

	chatdomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
	scenariodomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/scenario"
)

var _ chatdomain.ScenarioProvider = (*graphScenarioProvider)(nil)

// graphScenarioProvider adapts the scenario graph domain to the provider the
// chat use case expects, so that the scenario domain stays the single source of
// truth for scenario content.
//
// The two domains disagree on how a scenario is identified: chat persists
// scenario_id as bigint, while the graphs are keyed by the string ids declared
// in their YAML files. The translation table comes from configuration and is
// resolved here, at the composition root, leaving both domains untouched.
type graphScenarioProvider struct {
	scenarios scenariodomain.Repository
	ids       map[int64]string
}

func newGraphScenarioProvider(
	scenarios scenariodomain.Repository,
	ids map[int64]string,
) *graphScenarioProvider {
	return &graphScenarioProvider{scenarios: scenarios, ids: ids}
}

func (p *graphScenarioProvider) SystemPrompt(ctx context.Context, scenarioID int64) (string, error) {
	graphID, ok := p.ids[scenarioID]
	if !ok {
		return "", chatdomain.ErrScenarioNotFound
	}

	found, err := p.scenarios.GetByID(ctx, graphID)
	if err != nil {
		return "", chatdomain.ErrScenarioNotFound
	}

	return found.LLM.CharacterPrompt, nil
}
