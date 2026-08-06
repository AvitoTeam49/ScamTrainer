package eventbus

import (
	"context"
	"testing"
	"time"

	chatdomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
)

func TestBus_FansOutToEverySubscriberOfTheChat(t *testing.T) {
	bus := New(4)

	first, closeFirst := bus.Subscribe(1)
	defer closeFirst()
	second, closeSecond := bus.Subscribe(1)
	defer closeSecond()

	bus.Publish(context.Background(), chatdomain.ErrorEvent(1, "boom"))

	for name, events := range map[string]<-chan chatdomain.Event{"first": first, "second": second} {
		event := receive(t, events)
		if event.Reason != "boom" {
			t.Fatalf("%s subscriber got %q", name, event.Reason)
		}
	}
}

func TestBus_DoesNotDeliverEventsOfOtherChats(t *testing.T) {
	bus := New(4)

	events, unsubscribe := bus.Subscribe(1)
	defer unsubscribe()

	bus.Publish(context.Background(), chatdomain.ErrorEvent(2, "other chat"))

	select {
	case event := <-events:
		t.Fatalf("unexpected event: %+v", event)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestBus_UnsubscribeStopsDeliveryAndClosesChannel(t *testing.T) {
	bus := New(4)

	events, unsubscribe := bus.Subscribe(1)
	unsubscribe()

	if _, open := <-events; open {
		t.Fatal("channel must be closed after unsubscribe")
	}

	if got := bus.Subscribers(1); got != 0 {
		t.Fatalf("subscribers: got %d, want 0", got)
	}

	bus.Publish(context.Background(), chatdomain.ErrorEvent(1, "after unsubscribe"))
}

func TestBus_UnsubscribeIsIdempotent(t *testing.T) {
	bus := New(4)

	_, unsubscribe := bus.Subscribe(1)
	unsubscribe()
	unsubscribe()
}

func TestBus_SlowSubscriberDoesNotBlockPublisher(t *testing.T) {
	bus := New(1)

	slow, unsubscribe := bus.Subscribe(1)
	defer unsubscribe()

	done := make(chan struct{})
	go func() {
		defer close(done)

		for i := 0; i < 100; i++ {
			bus.Publish(context.Background(), chatdomain.ErrorEvent(1, "flood"))
		}
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("publisher blocked on a slow subscriber")
	}

	if len(slow) != 1 {
		t.Fatalf("buffered events: got %d, want 1", len(slow))
	}
}

func TestBus_CloseClosesEverySubscriber(t *testing.T) {
	bus := New(4)

	first, _ := bus.Subscribe(1)
	second, _ := bus.Subscribe(2)

	bus.Close()

	if _, open := <-first; open {
		t.Fatal("first subscriber must be closed")
	}

	if _, open := <-second; open {
		t.Fatal("second subscriber must be closed")
	}

	after, _ := bus.Subscribe(3)
	if _, open := <-after; open {
		t.Fatal("subscribing to a closed bus must yield a closed channel")
	}

	bus.Close()
}

func receive(t *testing.T, events <-chan chatdomain.Event) chatdomain.Event {
	t.Helper()

	select {
	case event := <-events:
		return event
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for an event")
		return chatdomain.Event{}
	}
}
