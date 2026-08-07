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

type TurnResult struct {
	Session  *scenario.TrainingSession
	Node     *scenario.Node
	Decision *scenario.Decision
}

func (s *TrainingService) Start(
	ctx context.Context,
	userID int64,
	scenarioID int,
) (*StartResult, error) {
	trainingScenario, err := s.scenarios.GetById(
		ctx,
		scenarioID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get scenario %d: %w",
			scenarioID,
			err,
		)
	}

	startNode, exists := trainingScenario.Nodes[trainingScenario.StartNodeID]
	if !exists {
		return nil, fmt.Errorf(
			"start node %q not found in scenario %d",
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

func (s *TrainingService) ApplyChoice(
	ctx context.Context,
	sessionID string,
	transitionID string,
) (*TurnResult, error) {
	session, err := s.sessions.GetById(
		ctx,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get training session %q: %w",
			sessionID,
			err,
		)
	}

	trainingScenario, err := s.scenarios.GetById(
		ctx,
		session.ScenarioID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get scenario %d: %w",
			session.ScenarioID,
			err,
		)
	}

	decision, err := s.engine.ApplyChoice(
		trainingScenario,
		session,
		transitionID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"apply transition %q to session %q: %w",
			transitionID,
			sessionID,
			err,
		)
	}

	node, exists := trainingScenario.Nodes[session.CurrentNodeID]
	if !exists {
		return nil, fmt.Errorf(
			"current node %q not found after transition %q",
			session.CurrentNodeID,
			transitionID,
		)
	}

	if err := s.sessions.Update(ctx, session); err != nil {
		return nil, fmt.Errorf(
			"update training session %q: %w",
			sessionID,
			err,
		)
	}

	return &TurnResult{
		Session:  session,
		Node:     node,
		Decision: decision,
	}, nil
}
