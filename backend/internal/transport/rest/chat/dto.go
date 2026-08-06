package chatrest

import (
	"time"

	chatdomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
)

type errorResponse struct {
	Error string `json:"error"`
}

type startChatRequest struct {
	ScenarioID int64  `json:"scenario_id"`
	Title      string `json:"title"`
}

type sendMessageRequest struct {
	Content string `json:"content"`
}

type chatResponse struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	ScenarioID    int64      `json:"scenario_id"`
	Title         string     `json:"title"`
	Status        string     `json:"status"`
	Resume        string     `json:"resume"`
	Score         int64      `json:"score"`
	CurrentNodeID string     `json:"current_node_id"`
	CreatedAt     time.Time  `json:"created_at"`
	FinishedAt    *time.Time `json:"finished_at"`
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

type decisionResponse struct {
	ID           int64     `json:"id"`
	ChatID       int64     `json:"chat_id"`
	NodeID       string    `json:"node_id"`
	TransitionID string    `json:"transition_id"`
	TargetNodeID string    `json:"target_node_id"`
	ScoreDelta   int       `json:"score_delta"`
	Feedback     string    `json:"feedback"`
	CreatedAt    time.Time `json:"created_at"`
}

type decisionsResponse struct {
	Items       []decisionResponse `json:"items"`
	NextAfterID *int64             `json:"next_after_id"`
}

type eventResponse struct {
	Type     string            `json:"type"`
	ChatID   int64             `json:"chat_id"`
	Message  *messageResponse  `json:"message,omitempty"`
	Decision *decisionResponse `json:"decision,omitempty"`
	Chat     *chatResponse     `json:"chat,omitempty"`
	Reason   string            `json:"reason,omitempty"`
}

func chatFrom(chat *chatdomain.Chat) chatResponse {
	return chatResponse{
		ID:            chat.ID,
		UserID:        chat.UserID,
		ScenarioID:    chat.ScenarioID,
		Title:         chat.Title,
		Status:        string(chat.Status),
		Resume:        chat.Resume,
		Score:         chat.Score,
		CurrentNodeID: chat.CurrentNodeID,
		CreatedAt:     chat.CreatedAt,
		FinishedAt:    chat.FinishedAt,
	}
}

func messageFrom(message *chatdomain.Message) messageResponse {
	return messageResponse{
		ID:         message.ID,
		ChatID:     message.ChatID,
		SenderType: string(message.SenderType),
		Content:    message.Content,
		CreatedAt:  message.CreatedAt,
	}
}

func messagesFrom(messages []*chatdomain.Message, limit int) messagesResponse {
	items := make([]messageResponse, 0, len(messages))
	for _, message := range messages {
		items = append(items, messageFrom(message))
	}

	response := messagesResponse{Items: items}
	if len(items) == limit {
		response.NextAfterID = &items[len(items)-1].ID
	}

	return response
}

func decisionFrom(decision *chatdomain.Decision) decisionResponse {
	return decisionResponse{
		ID:           decision.ID,
		ChatID:       decision.ChatID,
		NodeID:       decision.NodeID,
		TransitionID: decision.TransitionID,
		TargetNodeID: decision.TargetNodeID,
		ScoreDelta:   decision.ScoreDelta,
		Feedback:     decision.Feedback,
		CreatedAt:    decision.CreatedAt,
	}
}

func decisionsFrom(decisions []*chatdomain.Decision, limit int) decisionsResponse {
	items := make([]decisionResponse, 0, len(decisions))
	for _, decision := range decisions {
		items = append(items, decisionFrom(decision))
	}

	response := decisionsResponse{Items: items}
	if len(items) == limit {
		response.NextAfterID = &items[len(items)-1].ID
	}

	return response
}

func eventFrom(event chatdomain.Event) eventResponse {
	response := eventResponse{
		Type:   string(event.Type),
		ChatID: event.ChatID,
		Reason: event.Reason,
	}

	if event.Message != nil {
		message := messageFrom(event.Message)
		response.Message = &message
	}

	if event.Decision != nil {
		decision := decisionFrom(event.Decision)
		response.Decision = &decision
	}

	if event.Chat != nil {
		chat := chatFrom(event.Chat)
		response.Chat = &chat
	}

	return response
}
