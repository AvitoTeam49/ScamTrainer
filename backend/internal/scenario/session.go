package scenario

import "time"

type SessionStatus string

const (
	SessionStatusInProgress SessionStatus = "in_progress"
	SessionStatusCompleted  SessionStatus = "completed"
)

type TrainingSession struct {
	ID         string
	UserID     string
	ScenarioID string

	CurrentNodeID string
	Status        SessionStatus
	Score         int

	StartedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}
