package chatusecase

import (
	"context"
	"sync"
)

type turnGate struct {
	mu    sync.Mutex
	slots map[int64]*turnSlot
}

type turnSlot struct {
	busy chan struct{}
	refs int
}

func newTurnGate() *turnGate {
	return &turnGate{slots: make(map[int64]*turnSlot)}
}

func (g *turnGate) acquire(ctx context.Context, chatID int64) (func(), error) {
	slot := g.reserve(chatID)

	select {
	case slot.busy <- struct{}{}:
		return func() {
			<-slot.busy
			g.release(chatID)
		}, nil

	case <-ctx.Done():
		g.release(chatID)

		return nil, ctx.Err()
	}
}

func (g *turnGate) reserve(chatID int64) *turnSlot {
	g.mu.Lock()
	defer g.mu.Unlock()

	slot, ok := g.slots[chatID]
	if !ok {
		slot = &turnSlot{busy: make(chan struct{}, 1)}
		g.slots[chatID] = slot
	}

	slot.refs++

	return slot
}

func (g *turnGate) release(chatID int64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	slot, ok := g.slots[chatID]
	if !ok {
		return
	}

	slot.refs--
	if slot.refs == 0 {
		delete(g.slots, chatID)
	}
}
