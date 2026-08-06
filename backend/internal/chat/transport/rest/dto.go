package rest

import (
	"time"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/domain"
)

type errorResponse struct {
	Error string `json:"error"`
}

type chatResponse struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	ScenarioID int64      `json:"scenario_id"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
	Resume     string     `json:"resume"`
	Score      int64      `json:"score"`
	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at"`
}

type messageResponse struct {
	ID         int64     `json:"id"`
	ChatID     int64     `json:"chat_id"`
	SenderType string    `json:"sender_type"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type messagesResponse struct {
	Items       []messageResponse `json:"items"`
	NextAfterID *int64            `json:"next_after_id"`
}

func chatFrom(chat *domain.Chat) chatResponse {
	return chatResponse{
		ID:         chat.ID,
		UserID:     chat.UserID,
		ScenarioID: chat.ScenarioID,
		Title:      chat.Title,
		Status:     string(chat.Status),
		Resume:     chat.Resume,
		Score:      chat.Score,
		CreatedAt:  chat.CreatedAt,
		FinishedAt: chat.FinishedAt,
	}
}

func messagesFrom(messages []*domain.Message, limit int) messagesResponse {
	items := make([]messageResponse, 0, len(messages))
	for _, message := range messages {
		items = append(items, messageResponse{
			ID:         message.ID,
			ChatID:     message.ChatID,
			SenderType: string(message.SenderType),
			Content:    message.Content,
			CreatedAt:  message.CreatedAt,
		})
	}

	response := messagesResponse{Items: items}
	if len(items) == limit {
		response.NextAfterID = &items[len(items)-1].ID
	}

	return response
}
