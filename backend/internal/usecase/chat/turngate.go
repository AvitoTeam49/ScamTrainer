package chatusecase

import (
	"context"
	"sync"
)

// turnGate пропускает не больше одного хода агента на чат.
//
// Ход запускается горутиной на каждое входящее сообщение, поэтому два сообщения подряд
// давали два параллельных хода: оба успевали увидеть чат активным, и второй дописывал
// реплику уже после завершения сценария. Под воротами проверка chat.IsActive() снова
// становится осмысленной — второй ход читает чат уже завершённым и выходит.
//
// Ворота живут в памяти процесса, как и шина событий с сессиями тренировки:
// для нескольких реплик приложения понадобится внешняя блокировка.
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

// acquire ждёт своей очереди по чату и возвращает функцию освобождения.
// Ожидание прерывается вместе с контекстом, чтобы зависший ход не копил горутины.
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
