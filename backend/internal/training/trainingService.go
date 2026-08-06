package training

import (
	"context"
	"fmt"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/scenario"
)

type TrainingService struct {
	scenarios scenario.ScenarioRepository
	sessions  SessionRepository
	engine    *scenario.Engine
	ids       IDGenerator
}

func NewService(
	scenarios scenario.ScenarioRepository,
	sessions SessionRepository,
	engine *scenario.Engine,
	ids IDGenerator,
) *TrainingService {
	return &TrainingService{
		scenarios: scenarios,
		sessions:  sessions,
		engine:    engine,
		ids:       ids,
	}
}

type StartResult struct {
	Session *scenario.TrainingSession
	Node    *scenario.Node
}

func (s *TrainingService) Start(
	ctx context.Context,
	userID string,
	scenarioID string,
) (*StartResult, error) {
	trainingScenario, err := s.scenarios.GetById(
		ctx,
		scenarioID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get scenario %q: %w",
			scenarioID,
			err,
		)
	}

	startNode, exists := trainingScenario.Nodes[trainingScenario.StartNodeID]
	if !exists {
		return nil, fmt.Errorf(
			"start node %q not found in scenario %q",
			trainingScenario.StartNodeID,
			trainingScenario.ID,
		)
	}

	session := s.engine.Start(
		trainingScenario,
		s.ids.NewID(),
		userID,
	)

	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf(
			"create training session: %w",
			err,
		)
	}

	return &StartResult{
		Session: session,
		Node:    startNode,
	}, nil
}
