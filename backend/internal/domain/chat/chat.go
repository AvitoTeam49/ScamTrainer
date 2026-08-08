package chatdomain

import (
	"context"
	"time"
)

type ChatStatus string

const (
	ChatStatusActive    ChatStatus = "active"
	ChatStatusFinished  ChatStatus = "finished"
	ChatStatusAbandoned ChatStatus = "abandoned"
)

func (s ChatStatus) Valid() bool {
	switch s {
	case ChatStatusActive, ChatStatusFinished, ChatStatusAbandoned:
		return true
	default:
		return false
	}
}

type Chat struct {
	ID            int64
	UserID        int64
	ScenarioID    int64
	SessionID     string
	Title         string
	Status        ChatStatus
	Resume        string
	Score         int64
	CurrentNodeID string
	CreatedAt     time.Time
	FinishedAt    *time.Time
}

func NewChat(userID, scenarioID int64, sessionID, title, startNodeID string) *Chat {
	return &Chat{
		UserID:        userID,
		ScenarioID:    scenarioID,
		SessionID:     sessionID,
		Title:         title,
		Status:        ChatStatusActive,
		Score:         0,
		CurrentNodeID: startNodeID,
		CreatedAt:     time.Now(),
	}
}

func (c *Chat) IsActive() bool {
	return c.Status == ChatStatusActive
}

func (c *Chat) ApplyDecision(scoreDelta int, targetNodeID string) error {
	if !c.IsActive() {
		return ErrChatFinished
	}

	c.Score += int64(scoreDelta)
	c.CurrentNodeID = targetNodeID

	return nil
}

func (c *Chat) Finish(resume string) error {
	if !c.IsActive() {
		return ErrChatFinished
	}

	now := time.Now()
	c.Status = ChatStatusFinished
	c.Resume = resume
	c.FinishedAt = &now

	return nil
}

func (c *Chat) Abandon() error {
	if !c.IsActive() {
		return ErrChatFinished
	}

	now := time.Now()
	c.Status = ChatStatusAbandoned
	c.FinishedAt = &now

	return nil
}

type ChatRepository interface {
	GetByID(ctx context.Context, id int64) (*Chat, error)
	ListByUserID(ctx context.Context, userID int64, cursor Cursor) ([]*Chat, error)
	Create(ctx context.Context, chat *Chat) error
	// Close сохраняет терминальный статус чата (finished или abandoned).
	// Возвращает false, если чат уже был закрыт — например, параллельным ходом агента.
	Close(ctx context.Context, chat *Chat) (bool, error)
	Delete(ctx context.Context, id int64) error
}
