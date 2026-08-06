package training

import (
	"context"
	"errors"
	"testing"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/scenario"
)

type fixedIDGenerator struct {
	id string
}

func (g fixedIDGenerator) NewID() string {
	return g.id
}

type scenarioRepositoryStub struct {
	scenario    *scenario.Scenario
	err         error
	requestedID string
}

func (r *scenarioRepositoryStub) GetById(
	_ context.Context,
	id string,
) (*scenario.Scenario, error) {
	r.requestedID = id

	if r.err != nil {
		return nil, r.err
	}

	return r.scenario, nil
}

func TestTrainingService_Start(t *testing.T) {
	ctx := context.Background()

	trainingScenario := newServiceTestScenario()

	scenarioRepository := &scenarioRepositoryStub{
		scenario: trainingScenario,
	}

	sessionRepository := NewInMemorySessionRepository()

	service := NewService(
		scenarioRepository,
		sessionRepository,
		scenario.NewEngine(),
		fixedIDGenerator{id: "session-1"},
	)

	result, err := service.Start(
		ctx,
		"user-1",
		"scenario-1",
	)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if result == nil {
		t.Fatal("Start() result = nil")
	}

	if scenarioRepository.requestedID != "scenario-1" {
		t.Errorf(
			"requested scenario ID = %q, want %q",
			scenarioRepository.requestedID,
			"scenario-1",
		)
	}

	session := result.Session

	if session.ID != "session-1" {
		t.Errorf(
			"session ID = %q, want %q",
			session.ID,
			"session-1",
		)
	}

	if session.UserID != "user-1" {
		t.Errorf(
			"UserID = %q, want %q",
			session.UserID,
			"user-1",
		)
	}

	if session.ScenarioID != "scenario-1" {
		t.Errorf(
			"ScenarioID = %q, want %q",
			session.ScenarioID,
			"scenario-1",
		)
	}

	if session.CurrentNodeID != "start" {
		t.Errorf(
			"CurrentNodeID = %q, want %q",
			session.CurrentNodeID,
			"start",
		)
	}

	if session.Status != scenario.SessionStatusInProgress {
		t.Errorf(
			"Status = %q, want %q",
			session.Status,
			scenario.SessionStatusInProgress,
		)
	}

	if result.Node == nil {
		t.Fatal("result Node = nil")
	}

	if result.Node.ID != "start" {
		t.Errorf(
			"Node ID = %q, want %q",
			result.Node.ID,
			"start",
		)
	}

	stored, err := sessionRepository.GetById(
		ctx,
		"session-1",
	)
	if err != nil {
		t.Fatalf("GetById() error = %v", err)
	}

	requireSessionsEqual(t, stored, session)
}

func TestTrainingService_StartReturnsScenarioError(
	t *testing.T,
) {
	service := NewService(
		&scenarioRepositoryStub{
			err: scenario.ErrScenarioNotFound,
		},
		NewInMemorySessionRepository(),
		scenario.NewEngine(),
		fixedIDGenerator{id: "session-1"},
	)

	result, err := service.Start(
		context.Background(),
		"user-1",
		"missing-scenario",
	)

	if result != nil {
		t.Fatalf("Start() result = %#v, want nil", result)
	}

	if !errors.Is(err, scenario.ErrScenarioNotFound) {
		t.Fatalf(
			"Start() error = %v, want %v",
			err,
			scenario.ErrScenarioNotFound,
		)
	}
}

func TestTrainingService_StartReturnsDuplicateSessionError(
	t *testing.T,
) {
	ctx := context.Background()

	trainingScenario := newServiceTestScenario()
	sessionRepository := NewInMemorySessionRepository()
	engine := scenario.NewEngine()

	existing := engine.Start(
		trainingScenario,
		"session-1",
		"existing-user",
	)

	if err := sessionRepository.Create(
		ctx,
		existing,
	); err != nil {
		t.Fatalf("prepare existing session: %v", err)
	}

	service := NewService(
		&scenarioRepositoryStub{
			scenario: trainingScenario,
		},
		sessionRepository,
		engine,
		fixedIDGenerator{id: "session-1"},
	)

	result, err := service.Start(
		ctx,
		"user-1",
		trainingScenario.ID,
	)

	if result != nil {
		t.Fatalf("Start() result = %#v, want nil", result)
	}

	if !errors.Is(err, ErrSessionAlreadyExists) {
		t.Fatalf(
			"Start() error = %v, want %v",
			err,
			ErrSessionAlreadyExists,
		)
	}
}

func newServiceTestScenario() *scenario.Scenario {
	return &scenario.Scenario{
		ID:          "scenario-1",
		Title:       "Test scenario",
		Role:        scenario.RoleSeller,
		StartNodeID: "start",
		Nodes: map[string]*scenario.Node{
			"start": {
				ID:   "start",
				Type: scenario.NodeTypeDecision,
				Message: scenario.Message{
					Author: "scammer",
					Text:   "Test message",
				},
			},
		},
	}
}
