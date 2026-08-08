package chatrest

import (
	"context"
	"net/http"
	"time"

	chatdomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
)

const agentTurnTimeout = 5 * time.Minute

type ChatService interface {
	StartChat(ctx context.Context, userID, scenarioID int64, title string) (*chatdomain.Chat, error)
	SendMessage(ctx context.Context, chatID, userID int64, content string) (*chatdomain.Message, error)
	RunAgentTurn(ctx context.Context, chatID int64) error
	GetChat(ctx context.Context, chatID, userID int64) (*chatdomain.Chat, error)
	ListMessages(ctx context.Context, chatID, userID int64, cursor chatdomain.Cursor) ([]*chatdomain.Message, error)
	ListDecisions(ctx context.Context, chatID, userID int64, cursor chatdomain.Cursor) ([]*chatdomain.Decision, error)
}

type EventSubscriber interface {
	Subscribe(chatID int64) (<-chan chatdomain.Event, func())
}

type Handler struct {
	chats  ChatService
	events EventSubscriber
}

func NewHandler(chats ChatService, events EventSubscriber) *Handler {
	return &Handler{chats: chats, events: events}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/chats", h.startChat)
	mux.HandleFunc("GET /v1/chats/{chatID}", h.getChat)
	mux.HandleFunc("POST /v1/chats/{chatID}/messages", h.sendMessage)
	mux.HandleFunc("GET /v1/chats/{chatID}/messages", h.listMessages)
	mux.HandleFunc("GET /v1/chats/{chatID}/decisions", h.listDecisions)
	mux.HandleFunc("GET /v1/chats/{chatID}/events", h.streamEvents)
}

func (h *Handler) startChat(w http.ResponseWriter, r *http.Request) {
	userID, err := requestUserID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	var request startChatRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, err)
		return
	}

	if request.ScenarioID <= 0 {
		writeError(w, r, errInvalidRequest)
		return
	}

	chat, err := h.chats.StartChat(r.Context(), userID, request.ScenarioID, request.Title)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, chatFrom(chat))
}

func (h *Handler) getChat(w http.ResponseWriter, r *http.Request) {
	chatID, userID, err := chatScope(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	chat, err := h.chats.GetChat(r.Context(), chatID, userID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, chatFrom(chat))
}

func (h *Handler) sendMessage(w http.ResponseWriter, r *http.Request) {
	chatID, userID, err := chatScope(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	var request sendMessageRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, err)
		return
	}

	message, err := h.chats.SendMessage(r.Context(), chatID, userID, request.Content)
	if err != nil {
		writeError(w, r, err)
		return
	}

	h.startAgentTurn(r, chatID)

	writeJSON(w, http.StatusAccepted, messageFrom(message))
}

func (h *Handler) startAgentTurn(r *http.Request, chatID int64) {
	turnCtx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), agentTurnTimeout)

	go func() {
		defer cancel()

		_ = h.chats.RunAgentTurn(turnCtx, chatID)
	}()
}

func (h *Handler) listMessages(w http.ResponseWriter, r *http.Request) {
	chatID, userID, err := chatScope(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	cursor, err := cursorFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	messages, err := h.chats.ListMessages(r.Context(), chatID, userID, cursor)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, messagesFrom(messages, cursor.Limit))
}

func (h *Handler) listDecisions(w http.ResponseWriter, r *http.Request) {
	chatID, userID, err := chatScope(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	cursor, err := cursorFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	decisions, err := h.chats.ListDecisions(r.Context(), chatID, userID, cursor)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, decisionsFrom(decisions, cursor.Limit))
}
