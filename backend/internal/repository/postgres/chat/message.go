package chatpostgres

import (
	"context"
	"fmt"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
	sqlcChat "github.com/AvitoTeam49/ScamTrainer/backend/internal/repository/postgres/chat/sqlc"
)

type MessageRepository struct {
	queries sqlcChat.Querier
}

func NewMessageRepository(db DB) *MessageRepository {
	return &MessageRepository{queries: sqlcChat.New(db)}
}

func (r *MessageRepository) ListByChatID(ctx context.Context, chatID int64, cursor chatdomain.Cursor) ([]*chatdomain.Message, error) {
	if err := cursor.Validate(); err != nil {
		return nil, err
	}

	rows, err := r.queries.ListMessagesByChatID(ctx, sqlcChat.ListMessagesByChatIDParams{
		ChatID:  chatID,
		AfterID: cursor.AfterID,
		Lim:     int32(cursor.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list messages by chat id query: %w", err)
	}

	messages := make([]*chatdomain.Message, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, &chatdomain.Message{
			ID:         row.ID,
			ChatID:     row.ChatID,
			SenderType: chatdomain.SenderType(row.SenderType),
			Content:    row.Content,
			CreatedAt:  row.CreatedAt,
		})
	}

	return messages, nil
}

func (r *MessageRepository) Create(ctx context.Context, message *chatdomain.Message) error {
	row, err := r.queries.CreateMessage(ctx, sqlcChat.CreateMessageParams{
		ChatID:     message.ChatID,
		SenderType: string(message.SenderType),
		Content:    message.Content,
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return chatdomain.ErrChatNotFound
		}
		return fmt.Errorf("failed to create message query: %w", err)
	}

	message.ID = row.ID
	message.CreatedAt = row.CreatedAt

	return nil
}
