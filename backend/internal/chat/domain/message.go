package domain

import (
	"context"
	"time"
)

type SenderType string

const (
	SenderTypeAgent SenderType = "agent"
	SenderTypeUser  SenderType = "user"
)

func (s SenderType) Valid() bool {
	switch s {
	case SenderTypeAgent, SenderTypeUser:
		return true
	default:
		return false
	}
}

type Message struct {
	ID         int64
	ChatID     int64
	SenderType SenderType
	Content    string
	CreatedAt  time.Time
}

type MessageRepository interface {
	ListByChatID(ctx context.Context, chatID int64, cursor Cursor) ([]*Message, error)
	Create(ctx context.Context, message *Message) error
}
