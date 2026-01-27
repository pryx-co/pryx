package bus

import (
	"sync"

	"github.com/google/uuid"
)

// Handler is a function that handles an event
type Handler func(Event)

// Subscription represents a subscription to the bus
type Subscription struct {
	id      string
	ch      chan Event
	topics  []EventType
	handler Handler
	closer  func()
}

// Bus is the event bus
type Bus struct {
	mu   sync.RWMutex
	subs map[string]*Subscription
}

// New creates a new Bus
func New() *Bus {
	return &Bus{
		subs: make(map[string]*Subscription),
	}
}

// Subscribe subscribes to events. If topics is empty, it subscribes to all events.
// Returns a channel that receives events.
func (b *Bus) Subscribe(topics ...EventType) (<-chan Event, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := uuid.New().String()
	ch := make(chan Event, 100) // Buffer events

	sub := &Subscription{
		id:     id,
		ch:     ch,
		topics: topics,
		closer: func() {
			b.Unsubscribe(id)
		},
	}

	b.subs[id] = sub
	return ch, sub.closer
}

// Publish publishes an event to all subscribers
func (b *Bus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.subs {
		if b.matches(sub, event.Event) {
			select {
			case sub.ch <- event:
			default:
				// Drop event if subscriber is too slow to prevent blocking
				// In a real system we might want metrics here
			}
		}
	}
}

// Unsubscribe removes a subscription
func (b *Bus) Unsubscribe(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if sub, ok := b.subs[id]; ok {
		close(sub.ch)
		delete(b.subs, id)
	}
}

func (b *Bus) matches(sub *Subscription, topic EventType) bool {
	if len(sub.topics) == 0 {
		return true // Subscribe to all
	}
	for _, t := range sub.topics {
		if t == topic {
			return true
		}
	}
	return false
}
