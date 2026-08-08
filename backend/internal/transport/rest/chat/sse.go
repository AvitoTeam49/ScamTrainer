package chatrest

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const keepAliveInterval = 20 * time.Second

func (h *Handler) streamEvents(w http.ResponseWriter, r *http.Request) {
	chatID, userID, err := chatScope(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	if _, err := h.chats.GetChat(r.Context(), chatID, userID); err != nil {
		writeError(w, r, err)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "streaming unsupported"})
		return
	}

	events, unsubscribe := h.events.Subscribe(chatID)
	defer unsubscribe()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	keepAlive := time.NewTicker(keepAliveInterval)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-keepAlive.C:
			if _, err := fmt.Fprint(w, ": keep-alive\n\n"); err != nil {
				return
			}
			flusher.Flush()

		case event, open := <-events:
			if !open {
				return
			}

			payload, err := json.Marshal(eventFrom(event))
			if err != nil {
				slog.ErrorContext(r.Context(), "failed to encode chat event",
					"chat_id", chatID, "type", event.Type, "error", err)
				continue
			}

			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, payload); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
