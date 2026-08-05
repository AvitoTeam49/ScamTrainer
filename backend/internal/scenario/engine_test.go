package scenario

import (
	"errors"
	"testing"
)

func TestEngine_Start(t *testing.T) {
	s := &Scenario{
		ID:          "seller-scam",
		StartNodeID: "start",
		Nodes: map[string]*Node{
			"start": {
				ID:   "start",
				Type: NodeTypeDecision,
			},
		},
	}

	engine := NewEngine()

	session := engine.Start(
		s,
		"session-1",
		"user-1",
	)

	if session.ID != "session-1" {
		t.Errorf("ID = %q, want %q", session.ID, "session-1")
	}

	if session.UserID != "user-1" {
		t.Errorf("UserID = %q, want %q", session.UserID, "user-1")
	}

	if session.ScenarioID != s.ID {
		t.Errorf(
			"ScenarioID = %q, want %q",
			session.ScenarioID,
			s.ID,
		)
	}

	if session.CurrentNodeID != s.StartNodeID {
		t.Errorf(
			"CurrentNodeID = %q, want %q",
			session.CurrentNodeID,
			s.StartNodeID,
		)
	}

	if session.Status != SessionStatusInProgress {
		t.Errorf(
			"Status = %q, want %q",
			session.Status,
			SessionStatusInProgress,
		)
	}

	if session.Score != 0 {
		t.Errorf("Score = %d, want 0", session.Score)
	}

	if session.StartedAt.IsZero() {
		t.Error("StartedAt is zero")
	}

	if !session.StartedAt.Equal(session.UpdatedAt) {
		t.Errorf(
			"StartedAt = %v, UpdatedAt = %v, want equal",
			session.StartedAt,
			session.UpdatedAt,
		)
	}

	if session.CompletedAt != nil {
		t.Errorf(
			"CompletedAt = %v, want nil",
			session.CompletedAt,
		)
	}
}

func TestEngine_ApplyChoice_MovesToNextNode(t *testing.T) {
	s := &Scenario{
		ID:          "seller-scam",
		StartNodeID: "start",
		Nodes: map[string]*Node{
			"start": {
				ID:   "start",
				Type: NodeTypeDecision,
				Transitions: []Transition{
					{
						ID:         "stay_safe",
						ToNodeID:   "pressure",
						ScoreDelta: 10,
						Feedback:   "Безопасное решение",
					},
				},
			},
			"pressure": {
				ID:   "pressure",
				Type: NodeTypeDecision,
			},
		},
	}

	engine := NewEngine()
	session := engine.Start(s, "session-1", "user-1")

	previousUpdatedAt := session.UpdatedAt

	decision, err := engine.ApplyChoice(
		s,
		session,
		"stay_safe",
	)
	if err != nil {
		t.Fatalf("ApplyChoice() error = %v", err)
	}

	// Проверяем новое состояние сессии.

	if session.CurrentNodeID != "pressure" {
		t.Errorf(
			"CurrentNodeID = %q, want %q",
			session.CurrentNodeID,
			"pressure",
		)
	}

	if session.Score != 10 {
		t.Errorf("Score = %d, want 10", session.Score)
	}

	if session.Status != SessionStatusInProgress {
		t.Errorf(
			"Status = %q, want %q",
			session.Status,
			SessionStatusInProgress,
		)
	}

	if session.CompletedAt != nil {
		t.Errorf(
			"CompletedAt = %v, want nil",
			session.CompletedAt,
		)
	}

	if session.UpdatedAt.Before(previousUpdatedAt) {
		t.Errorf(
			"UpdatedAt = %v, must not be before %v",
			session.UpdatedAt,
			previousUpdatedAt,
		)
	}

	// Проверяем описание произошедшего шага.

	if decision.SessionID != session.ID {
		t.Errorf(
			"Decision.SessionID = %q, want %q",
			decision.SessionID,
			session.ID,
		)
	}

	if decision.NodeID != "start" {
		t.Errorf(
			"Decision.NodeID = %q, want %q",
			decision.NodeID,
			"start",
		)
	}

	if decision.TransitionID != "stay_safe" {
		t.Errorf(
			"Decision.TransitionID = %q, want %q",
			decision.TransitionID,
			"stay_safe",
		)
	}

	if decision.TargetNodeID != "pressure" {
		t.Errorf(
			"Decision.TargetNodeID = %q, want %q",
			decision.TargetNodeID,
			"pressure",
		)
	}

	if decision.ScoreDelta != 10 {
		t.Errorf(
			"Decision.ScoreDelta = %d, want 10",
			decision.ScoreDelta,
		)
	}

	if decision.Feedback != "Безопасное решение" {
		t.Errorf(
			"Decision.Feedback = %q, want %q",
			decision.Feedback,
			"Безопасное решение",
		)
	}

	if !decision.CreatedAt.Equal(session.UpdatedAt) {
		t.Errorf(
			"Decision.CreatedAt = %v, session.UpdatedAt = %v, want equal",
			decision.CreatedAt,
			session.UpdatedAt,
		)
	}
}

func TestEngine_ApplyChoice_CompletesSession(t *testing.T) {
	s := &Scenario{
		ID:          "seller-scam",
		StartNodeID: "start",
		Nodes: map[string]*Node{
			"start": {
				ID:   "start",
				Type: NodeTypeDecision,
				Transitions: []Transition{
					{
						ID:         "finish_safe",
						ToNodeID:   "safe_ending",
						ScoreDelta: 20,
						Feedback:   "Сценарий завершён безопасно",
					},
				},
			},
			"safe_ending": {
				ID:   "safe_ending",
				Type: NodeTypeEnding,
			},
		},
	}

	engine := NewEngine()
	session := engine.Start(s, "session-1", "user-1")

	decision, err := engine.ApplyChoice(
		s,
		session,
		"finish_safe",
	)
	if err != nil {
		t.Fatalf("ApplyChoice() error = %v", err)
	}

	if session.CurrentNodeID != "safe_ending" {
		t.Errorf(
			"CurrentNodeID = %q, want %q",
			session.CurrentNodeID,
			"safe_ending",
		)
	}

	if session.Status != SessionStatusCompleted {
		t.Errorf(
			"Status = %q, want %q",
			session.Status,
			SessionStatusCompleted,
		)
	}

	if session.CompletedAt == nil {
		t.Fatal("CompletedAt = nil, want completion time")
	}

	if !session.CompletedAt.Equal(session.UpdatedAt) {
		t.Errorf(
			"CompletedAt = %v, UpdatedAt = %v, want equal",
			session.CompletedAt,
			session.UpdatedAt,
		)
	}

	if decision.TargetNodeID != "safe_ending" {
		t.Errorf(
			"Decision.TargetNodeID = %q, want %q",
			decision.TargetNodeID,
			"safe_ending",
		)
	}
}

func TestEngine_ApplyChoice_RejectsUnknownTransition(t *testing.T) {
	s := &Scenario{
		ID:          "seller-scam",
		StartNodeID: "start",
		Nodes: map[string]*Node{
			"start": {
				ID:   "start",
				Type: NodeTypeDecision,
			},
		},
	}

	engine := NewEngine()
	session := engine.Start(s, "session-1", "user-1")

	previousNodeID := session.CurrentNodeID
	previousScore := session.Score
	previousUpdatedAt := session.UpdatedAt

	decision, err := engine.ApplyChoice(
		s,
		session,
		"unknown-transition",
	)

	if !errors.Is(err, ErrTransitionNotAvailable) {
		t.Fatalf(
			"error = %v, want %v",
			err,
			ErrTransitionNotAvailable,
		)
	}

	if decision != nil {
		t.Errorf("decision = %#v, want nil", decision)
	}

	if session.CurrentNodeID != previousNodeID {
		t.Errorf(
			"CurrentNodeID changed from %q to %q",
			previousNodeID,
			session.CurrentNodeID,
		)
	}

	if session.Score != previousScore {
		t.Errorf(
			"Score changed from %d to %d",
			previousScore,
			session.Score,
		)
	}

	if !session.UpdatedAt.Equal(previousUpdatedAt) {
		t.Errorf(
			"UpdatedAt changed from %v to %v",
			previousUpdatedAt,
			session.UpdatedAt,
		)
	}
}
