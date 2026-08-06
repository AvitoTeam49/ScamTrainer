package scenariodomain

import "time"

type Decision struct {
	SessionID string

	NodeID       string
	TransitionID string
	TargetNodeID string

	ScoreDelta int
	Feedback   string

	CreatedAt time.Time
}
