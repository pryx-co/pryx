package bus

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBus(t *testing.T) {
	b := New()
	assert.NotNil(t, b)
	assert.NotNil(t, b.subs)
	assert.Empty(t, b.subs)
}

func TestBus_Subscribe_Unsubscribe(t *testing.T) {
	b := New()

	ch, cancel := b.Subscribe()
	require.NotNil(t, ch)
	require.NotNil(t, cancel)

	// Should have one subscriber
	b.mu.RLock()
	assert.Len(t, b.subs, 1)
	b.mu.RUnlock()

	// Unsubscribe
	cancel()

	// Should have no subscribers
	b.mu.RLock()
	assert.Len(t, b.subs, 0)
	b.mu.RUnlock()
}

func TestBus_Subscribe_WithTopics(t *testing.T) {
	b := New()

	ch, cancel := b.Subscribe(EventTraceEvent, EventErrorOccurred)
	defer cancel()

	require.NotNil(t, ch)

	// Verify subscription has topics
	b.mu.RLock()
	var sub *Subscription
	for _, s := range b.subs {
		sub = s
		break
	}
	b.mu.RUnlock()

	require.NotNil(t, sub)
	assert.Equal(t, []EventType{EventTraceEvent, EventErrorOccurred}, sub.topics)
}

func TestBus_Publish_SubscribesToAll(t *testing.T) {
	b := New()

	ch, cancel := b.Subscribe() // Subscribe to all
	defer cancel()

	// Publish different event types
	event1 := NewEvent(EventTraceEvent, "session1", map[string]interface{}{"key": "value1"})
	event2 := NewEvent(EventErrorOccurred, "session2", map[string]interface{}{"key": "value2"})

	b.Publish(event1)
	b.Publish(event2)

	// Both should be received
	received1 := <-ch
	received2 := <-ch

	assert.Equal(t, EventTraceEvent, received1.Event)
	assert.Equal(t, EventErrorOccurred, received2.Event)
}

func TestBus_Publish_TopicFiltering(t *testing.T) {
	b := New()

	// Subscribe only to Trace events
	ch, cancel := b.Subscribe(EventTraceEvent)
	defer cancel()

	// Publish Trace event - should be received
	traceEvent := NewEvent(EventTraceEvent, "session1", map[string]interface{}{"type": "trace"})
	b.Publish(traceEvent)

	select {
	case received := <-ch:
		assert.Equal(t, EventTraceEvent, received.Event)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected to receive trace event")
	}

	// Publish Error event - should NOT be received
	errorEvent := NewEvent(EventErrorOccurred, "session2", map[string]interface{}{"type": "error"})
	b.Publish(errorEvent)

	select {
	case <-ch:
		t.Fatal("Should not receive error event")
	case <-time.After(100 * time.Millisecond):
		// Expected - no event received
	}
}

func TestBus_Publish_MultipleSubscribers(t *testing.T) {
	b := New()

	ch1, cancel1 := b.Subscribe()
	defer cancel1()

	ch2, cancel2 := b.Subscribe()
	defer cancel2()

	event := NewEvent(EventTraceEvent, "session1", map[string]interface{}{"key": "value"})
	b.Publish(event)

	// Both subscribers should receive
	<-ch1
	<-ch2
	// If we get here without blocking, both received the event
}

func TestBus_Publish_VersionIncrement(t *testing.T) {
	b := New()

	ch, cancel := b.Subscribe()
	defer cancel()

	// Publish multiple events
	for i := 0; i < 5; i++ {
		event := NewEvent(EventTraceEvent, "session", map[string]interface{}{"index": i})
		b.Publish(event)
	}

	// Verify versions are monotonically increasing
	var lastVersion int
	for i := 0; i < 5; i++ {
		received := <-ch
		if i > 0 {
			assert.Greater(t, received.Version, lastVersion, "Version should increment")
		}
		lastVersion = received.Version
	}
}

func TestBus_Publish_BufferFull(t *testing.T) {
	b := New()

	// Create subscriber that doesn't read
	ch, cancel := b.Subscribe()
	defer cancel()

	// Fill buffer (100 events) + some extra
	for i := 0; i < 150; i++ {
		event := NewEvent(EventTraceEvent, "session", map[string]interface{}{"index": i})
		// Should not block even if subscriber is slow
		done := make(chan struct{})
		go func() {
			b.Publish(event)
			close(done)
		}()

		select {
		case <-done:
			// Good - publish didn't block
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Publish blocked on full buffer")
		}
	}

	// Drain the channel to clean up
	go func() {
		for range ch {
		}
	}()
}

func TestBus_ConcurrentSubscribePublish(t *testing.T) {
	b := New()
	var wg sync.WaitGroup

	// Concurrent subscriptions
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch, cancel := b.Subscribe()
			time.Sleep(10 * time.Millisecond)
			cancel()
			_ = ch // Prevent unused variable warning
		}()
	}

	// Concurrent publishes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			event := NewEvent(EventTraceEvent, "session", map[string]interface{}{"index": index})
			b.Publish(event)
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Concurrent operations timed out")
	}
}

func TestBus_Unsubscribe_Idempotent(t *testing.T) {
	b := New()

	ch, cancel := b.Subscribe()
	cancel()

	// Unsubscribe again should not panic
	cancel()

	// Channel should be closed
	_, ok := <-ch
	assert.False(t, ok, "Channel should be closed")
}

func TestBus_Matches(t *testing.T) {
	b := New()

	tests := []struct {
		name     string
		topics   []EventType
		event    EventType
		expected bool
	}{
		{"empty topics matches all", []EventType{}, EventTraceEvent, true},
		{"single topic match", []EventType{EventTraceEvent}, EventTraceEvent, true},
		{"single topic no match", []EventType{EventTraceEvent}, EventErrorOccurred, false},
		{"multiple topics match", []EventType{EventTraceEvent, EventErrorOccurred}, EventErrorOccurred, true},
		{"multiple topics no match", []EventType{EventTraceEvent, EventChatRequest}, EventErrorOccurred, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscription{topics: tt.topics}
			result := b.matches(sub, tt.event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewEvent(t *testing.T) {
	sessionID := "test-session"
	payload := map[string]interface{}{
		"key": "value",
		"num": 42,
	}

	event := NewEvent(EventTraceEvent, sessionID, payload)

	assert.Equal(t, EventTraceEvent, event.Event)
	assert.Equal(t, sessionID, event.SessionID)
	assert.Equal(t, payload, event.Payload)
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, 0, event.Version) // Version set on publish
}

func BenchmarkBus_Publish(b *testing.B) {
	bus := New()
	ch, cancel := bus.Subscribe()
	defer cancel()

	// Drain channel
	go func() {
		for range ch {
		}
	}()

	event := NewEvent(EventTraceEvent, "session", map[string]interface{}{"key": "value"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(event)
	}
}

func BenchmarkBus_Subscribe(b *testing.B) {
	bus := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, cancel := bus.Subscribe()
		cancel()
	}
}
