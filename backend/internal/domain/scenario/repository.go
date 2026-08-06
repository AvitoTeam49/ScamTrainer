package scenariodomain

import (
	"context"
	"errors"
)

var ErrScenarioNotFound = errors.New("scenario not found")

type Repository interface {
	GetByID(ctx context.Context, id string) (*Scenario, error)
}
