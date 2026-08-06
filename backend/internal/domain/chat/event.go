package chatdomain

type EventType string

const (
	EventTypeMessage  EventType = "message"
	EventTypeDecision EventType = "decision"
	EventTypeChat     EventType = "chat"
	EventTypeError    EventType = "error"
)

type Event struct {
	Type     EventType
	ChatID   int64
	Message  *Message
	Decision *Decision
	Chat     *Chat
	Reason   string
}

func MessageEvent(message *Message) Event {
	return Event{Type: EventTypeMessage, ChatID: message.ChatID, Message: message}
}

func DecisionEvent(decision *Decision) Event {
	return Event{Type: EventTypeDecision, ChatID: decision.ChatID, Decision: decision}
}

func ChatEvent(chat *Chat) Event {
	return Event{Type: EventTypeChat, ChatID: chat.ID, Chat: chat}
}

func ErrorEvent(chatID int64, reason string) Event {
	return Event{Type: EventTypeError, ChatID: chatID, Reason: reason}
}
