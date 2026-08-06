package scenariodomain

import (
	"errors"
	"time"
)

var (
	ErrSessionCompleted       = errors.New("training session is completed")
	ErrScenarioMismatch       = errors.New("session belongs to another scenario")
	ErrCurrentNodeNotFound    = errors.New("current node not found")
	ErrTransitionNotAvailable = errors.New("transition is not available")
	ErrTargetNodeNotFound     = errors.New("target node not found")
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Start(s *Scenario, sessionID string, userID string) *TrainingSession {
	now := time.Now()
	return &TrainingSession{
		ID:            sessionID,
		UserID:        userID,
		ScenarioID:    s.ID,
		CurrentNodeID: s.StartNodeID,
		Status:        SessionStatusInProgress,
		Score:         0,
		StartedAt:     now,
		UpdatedAt:     now,
	}
}

func (e *Engine) ApplyChoice(s *Scenario, session *TrainingSession, transitionID string) (*Decision, error) {
	if session.Status == SessionStatusCompleted {
		return nil, ErrSessionCompleted
	}

	if session.ScenarioID != s.ID {
		return nil, ErrScenarioMismatch
	}

	curNode, ok := s.Nodes[session.CurrentNodeID]
	if !ok {
		return nil, ErrCurrentNodeNotFound
	}

	transition, ok := findTransition(curNode.Transitions, transitionID)
	if !ok {
		return nil, ErrTransitionNotAvailable
	}

	targetNode, ok := s.Nodes[transition.ToNodeID]
	if !ok {
		return nil, ErrTargetNodeNotFound
	}

	now := time.Now()

	prevNodeID := session.CurrentNodeID

	session.CurrentNodeID = targetNode.ID
	session.Score += transition.ScoreDelta
	session.UpdatedAt = now

	completed := targetNode.Type == NodeTypeEnding
	if completed {
		session.Status = SessionStatusCompleted
		session.CompletedAt = &now
	}

	return &Decision{
		SessionID:    session.ID,
		NodeID:       prevNodeID,
		TransitionID: transition.ID,
		TargetNodeID: targetNode.ID,
		ScoreDelta:   transition.ScoreDelta,
		Feedback:     transition.Feedback,
		CreatedAt:    now,
	}, nil
}

func findTransition(transitions []Transition, transitionID string) (*Transition, bool) {
	for i := range transitions {
		if transitions[i].ID == transitionID {
			return &transitions[i], true
		}
	}
	return nil, false
}
