package chatdomain

import (
	"context"
	"time"
)

type Decision struct {
	ID           int64
	ChatID       int64
	NodeID       string
	TransitionID string
	TargetNodeID string
	ScoreDelta   int
	Feedback     string
	CreatedAt    time.Time
}

type DecisionRepository interface {
	ListByChatID(ctx context.Context, chatID int64, cursor Cursor) ([]*Decision, error)
	Create(ctx context.Context, decision *Decision, score int64) error
}
