package bus

import (
	"testing"
)

func TestBus_AssignsMonotonicVersions(t *testing.T) {
	b := New()

	ch, cancel := b.Subscribe()
	defer cancel()

	b.Publish(NewEvent(EventTraceEvent, "s1", map[string]interface{}{"n": 1}))
	b.Publish(NewEvent(EventTraceEvent, "s1", map[string]interface{}{"n": 2}))

	first := <-ch
	second := <-ch

	if first.Version <= 0 {
		t.Fatalf("expected first event version > 0, got %d", first.Version)
	}
	if second.Version != first.Version+1 {
		t.Fatalf("expected second version %d, got %d", first.Version+1, second.Version)
	}
}
