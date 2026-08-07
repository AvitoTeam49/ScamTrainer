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

func newApplyChoiceTestScenario() *scenario.Scenario {
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
					Text:   "Перейдите по ссылке",
				},
				Transitions: []scenario.Transition{
					{
						ID:          "stay_safe",
						Description: "Отказаться от перехода",
						ToNodeID:    "pressure",
						ScoreDelta:  10,
						Feedback:    "Безопасное решение",
					},
					{
						ID:          "finish_safe",
						Description: "Завершить разговор",
						ToNodeID:    "safe_ending",
						ScoreDelta:  20,
						Feedback:    "Сценарий завершён безопасно",
					},
				},
			},
			"pressure": {
				ID:   "pressure",
				Type: scenario.NodeTypeDecision,
				Message: scenario.Message{
					Author: "scammer",
					Text:   "Почему вы мне не доверяете?",
				},
			},
			"safe_ending": {
				ID:      "safe_ending",
				Type:    scenario.NodeTypeEnding,
				Title:   "Безопасное завершение",
				Outcome: "Пользователь не перешёл по опасной ссылке",
			},
		},
	}
}

func TestTrainingService_ApplyChoice(t *testing.T) {
	ctx := context.Background()

	trainingScenario := newApplyChoiceTestScenario()
	sessionRepository := NewInMemorySessionRepository()
	engine := scenario.NewEngine()

	session := engine.Start(
		trainingScenario,
		"session-1",
		"user-1",
	)

	if err := sessionRepository.Create(
		ctx,
		session,
	); err != nil {
		t.Fatalf("prepare session: %v", err)
	}

	service := NewService(
		&scenarioRepositoryStub{
			scenario: trainingScenario,
		},
		sessionRepository,
		engine,
		fixedIDGenerator{id: "unused"},
	)

	result, err := service.ApplyChoice(
		ctx,
		"session-1",
		"stay_safe",
	)
	if err != nil {
		t.Fatalf("ApplyChoice() error = %v", err)
	}

	if result == nil {
		t.Fatal("ApplyChoice() result = nil")
	}

	if result.Session.CurrentNodeID != "pressure" {
		t.Errorf(
			"CurrentNodeID = %q, want %q",
			result.Session.CurrentNodeID,
			"pressure",
		)
	}

	if result.Session.Score != 10 {
		t.Errorf(
			"Score = %d, want 10",
			result.Session.Score,
		)
	}

	if result.Node == nil {
		t.Fatal("Node = nil")
	}

	if result.Node.ID != "pressure" {
		t.Errorf(
			"Node.ID = %q, want %q",
			result.Node.ID,
			"pressure",
		)
	}

	if result.Decision == nil {
		t.Fatal("Decision = nil")
	}

	if result.Decision.TransitionID != "stay_safe" {
		t.Errorf(
			"Decision.TransitionID = %q, want %q",
			result.Decision.TransitionID,
			"stay_safe",
		)
	}

	if result.Decision.ScoreDelta != 10 {
		t.Errorf(
			"Decision.ScoreDelta = %d, want 10",
			result.Decision.ScoreDelta,
		)
	}

	if result.Decision.Feedback != "Безопасное решение" {
		t.Errorf(
			"Decision.Feedback = %q, want %q",
			result.Decision.Feedback,
			"Безопасное решение",
		)
	}

	stored, err := sessionRepository.GetById(
		ctx,
		"session-1",
	)
	if err != nil {
		t.Fatalf("GetById() error = %v", err)
	}

	if stored.CurrentNodeID != "pressure" {
		t.Errorf(
			"stored CurrentNodeID = %q, want %q",
			stored.CurrentNodeID,
			"pressure",
		)
	}

	if stored.Score != 10 {
		t.Errorf(
			"stored Score = %d, want 10",
			stored.Score,
		)
	}
}

func TestTrainingService_ApplyChoiceReturnsTransitionError(
	t *testing.T,
) {
	ctx := context.Background()

	trainingScenario := newApplyChoiceTestScenario()
	sessionRepository := NewInMemorySessionRepository()
	engine := scenario.NewEngine()

	session := engine.Start(
		trainingScenario,
		"session-1",
		"user-1",
	)

	if err := sessionRepository.Create(
		ctx,
		session,
	); err != nil {
		t.Fatalf("prepare session: %v", err)
	}

	service := NewService(
		&scenarioRepositoryStub{
			scenario: trainingScenario,
		},
		sessionRepository,
		engine,
		fixedIDGenerator{id: "unused"},
	)

	result, err := service.ApplyChoice(
		ctx,
		"session-1",
		"unknown-transition",
	)

	if result != nil {
		t.Fatalf(
			"ApplyChoice() result = %#v, want nil",
			result,
		)
	}

	if !errors.Is(
		err,
		scenario.ErrTransitionNotAvailable,
	) {
		t.Fatalf(
			"ApplyChoice() error = %v, want %v",
			err,
			scenario.ErrTransitionNotAvailable,
		)
	}

	stored, err := sessionRepository.GetById(
		ctx,
		"session-1",
	)
	if err != nil {
		t.Fatalf("GetById() error = %v", err)
	}

	if stored.CurrentNodeID != "start" {
		t.Errorf(
			"stored CurrentNodeID = %q, want %q",
			stored.CurrentNodeID,
			"start",
		)
	}

	if stored.Score != 0 {
		t.Errorf(
			"stored Score = %d, want 0",
			stored.Score,
		)
	}
}

func TestTrainingService_ApplyChoiceReturnsSessionNotFound(
	t *testing.T,
) {
	service := NewService(
		&scenarioRepositoryStub{
			scenario: newApplyChoiceTestScenario(),
		},
		NewInMemorySessionRepository(),
		scenario.NewEngine(),
		fixedIDGenerator{id: "unused"},
	)

	result, err := service.ApplyChoice(
		context.Background(),
		"missing-session",
		"stay_safe",
	)

	if result != nil {
		t.Fatalf(
			"ApplyChoice() result = %#v, want nil",
			result,
		)
	}

	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf(
			"ApplyChoice() error = %v, want %v",
			err,
			ErrSessionNotFound,
		)
	}
}
