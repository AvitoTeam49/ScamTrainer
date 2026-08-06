package chatrest

import (
	"context"
	"net/http"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
)

type ChatService interface {
	GetChat(ctx context.Context, chatID int64) (*chatdomain.Chat, error)
	ListMessages(ctx context.Context, chatID int64, cursor chatdomain.Cursor) ([]*chatdomain.Message, error)
}

type Handler struct {
	chats ChatService
}

func NewHandler(chats ChatService) *Handler {
	return &Handler{chats: chats}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/chats/{chatID}", h.getChat)
	mux.HandleFunc("GET /v1/chats/{chatID}/messages", h.listMessages)
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
