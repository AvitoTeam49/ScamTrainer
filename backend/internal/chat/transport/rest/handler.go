package rest

import (
	"context"
	"net/http"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/chat/domain"
)

type ChatService interface {
	GetChat(ctx context.Context, chatID int64) (*domain.Chat, error)
	ListMessages(ctx context.Context, chatID int64, cursor domain.Cursor) ([]*domain.Message, error)
}

type Handler struct {
	chats ChatService
}

func NewHandler(chats ChatService) *Handler {
	return &Handler{chats: chats}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /chats/{chatID}", h.getChat)
	mux.HandleFunc("GET /chats/{chatID}/messages", h.listMessages)
}

func (h *Handler) getChat(w http.ResponseWriter, r *http.Request) {
	chatID, err := pathID(r, "chatID")
	if err != nil {
		writeError(w, r, err)
		return
	}

	chat, err := h.chats.GetChat(r.Context(), chatID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, chatFrom(chat))
}

func (h *Handler) listMessages(w http.ResponseWriter, r *http.Request) {
	chatID, err := pathID(r, "chatID")
	if err != nil {
		writeError(w, r, err)
		return
	}

	cursor, err := cursorFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	messages, err := h.chats.ListMessages(r.Context(), chatID, cursor)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, messagesFrom(messages, cursor.Limit))
}
