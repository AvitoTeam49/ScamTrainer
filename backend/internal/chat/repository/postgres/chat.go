package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/domain"
	sqlcChat "github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/repository/postgres/sqlc"
	"github.com/jackc/pgx/v5"
)

type ChatRepository struct {
	queries sqlcChat.Querier
}

func NewChatRepository(db DB) *ChatRepository {
	return &ChatRepository{queries: sqlcChat.New(db)}
}

func (r *ChatRepository) GetByID(ctx context.Context, id int64) (*domain.Chat, error) {
	row, err := r.queries.GetChatByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrChatNotFound
		}
		return nil, fmt.Errorf("failed to get chat by id query: %w", err)
	}

	return chatFromRow(row), nil
}

func (r *ChatRepository) ListByUserID(ctx context.Context, userID int64, cursor domain.Cursor) ([]*domain.Chat, error) {
	if err := cursor.Validate(); err != nil {
		return nil, err
	}

	rows, err := r.queries.ListChatsByUserID(ctx, sqlcChat.ListChatsByUserIDParams{
		UserID:  userID,
		AfterID: cursor.AfterID,
		Lim:     int32(cursor.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list chats by user id query: %w", err)
	}

	chats := make([]*domain.Chat, 0, len(rows))
	for _, row := range rows {
		chats = append(chats, chatFromRow(row))
	}

	return chats, nil
}

func (r *ChatRepository) Create(ctx context.Context, chat *domain.Chat) error {
	id, err := r.queries.CreateChat(ctx, sqlcChat.CreateChatParams{
		UserID:     chat.UserID,
		ScenarioID: chat.ScenarioID,
		Title:      chat.Title,
		Status:     string(chat.Status),
		Resume:     chat.Resume,
		Score:      chat.Score,
		CreatedAt:  chat.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to create chat query: %w", err)
	}

	chat.ID = id

	return nil
}

func (r *ChatRepository) Update(ctx context.Context, chat *domain.Chat) error {
	affected, err := r.queries.UpdateChat(ctx, sqlcChat.UpdateChatParams{
		ID:         chat.ID,
		Title:      chat.Title,
		Status:     string(chat.Status),
		Resume:     chat.Resume,
		Score:      chat.Score,
		FinishedAt: chat.FinishedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to update chat query: %w", err)
	}

	if affected == 0 {
		return domain.ErrChatNotFound
	}

	return nil
}

func (r *ChatRepository) Delete(ctx context.Context, id int64) error {
	affected, err := r.queries.DeleteChat(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete chat query: %w", err)
	}

	if affected == 0 {
		return domain.ErrChatNotFound
	}

	return nil
}

func chatFromRow(row sqlcChat.Chat) *domain.Chat {
	return &domain.Chat{
		ID:         row.ID,
		UserID:     row.UserID,
		ScenarioID: row.ScenarioID,
		Title:      row.Title,
		Status:     domain.ChatStatus(row.Status),
		Resume:     row.Resume,
		Score:      row.Score,
		CreatedAt:  row.CreatedAt,
		FinishedAt: row.FinishedAt,
	}
}
