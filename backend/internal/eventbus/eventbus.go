package eventbus

import (
	"context"
	"sync"

	chatdomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
)

const defaultBuffer = 32

type Bus struct {
	mu          sync.RWMutex
	subscribers map[int64]map[uint64]chan chatdomain.Event
	nextID      uint64
	buffer      int
	closed      bool
}

func New(buffer int) *Bus {
	if buffer <= 0 {
		buffer = defaultBuffer
	}

	return &Bus{
		subscribers: make(map[int64]map[uint64]chan chatdomain.Event),
		buffer:      buffer,
	}
}

func (b *Bus) Subscribe(chatID int64) (<-chan chatdomain.Event, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	events := make(chan chatdomain.Event, b.buffer)

	if b.closed {
		close(events)
		return events, func() {}
	}

	b.nextID++
	id := b.nextID

	if _, ok := b.subscribers[chatID]; !ok {
		b.subscribers[chatID] = make(map[uint64]chan chatdomain.Event)
	}
	b.subscribers[chatID][id] = events

	return events, func() { b.unsubscribe(chatID, id) }
}

func (b *Bus) unsubscribe(chatID int64, id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	byChat, ok := b.subscribers[chatID]
	if !ok {
		return
	}

	events, ok := byChat[id]
	if !ok {
		return
	}

	delete(byChat, id)
	if len(byChat) == 0 {
		delete(b.subscribers, chatID)
	}

	close(events)
}

func (b *Bus) Publish(_ context.Context, event chatdomain.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, events := range b.subscribers[event.ChatID] {
		select {
		case events <- event:
		default:
		}
	}
}

func (b *Bus) Subscribers(chatID int64) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.subscribers[chatID])
}

func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}
	b.closed = true

	for chatID, byChat := range b.subscribers {
		for id, events := range byChat {
			close(events)
			delete(byChat, id)
		}
		delete(b.subscribers, chatID)
	}
}
