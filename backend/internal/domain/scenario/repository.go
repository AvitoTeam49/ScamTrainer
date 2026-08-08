package scenariodomain

import (
	"context"
	"errors"
)

var ErrScenarioNotFound = errors.New("scenario not found")

type ScenarioRepository interface {
	GetById(ctx context.Context, id int) (*Scenario, error)
}

type ScenarioCatalog interface {
	List(
		ctx context.Context,
	) ([]ScenarioInfo, error)

	ListByDifficulty(
		ctx context.Context,
		difficulty Difficulty,
	) ([]ScenarioInfo, error)
}

type ScenarioInfo struct {
	ID         int
	Title      string
	Role       Role
	Difficulty Difficulty
}
