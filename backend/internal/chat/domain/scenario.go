package domain

import "context"

type ScenarioProvider interface {
	SystemPrompt(ctx context.Context, scenarioID int64) (string, error)
}
