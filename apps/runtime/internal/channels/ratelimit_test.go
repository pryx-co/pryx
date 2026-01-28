package channels

import (
	"context"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	mock := &mockChannel{id: "rl-test", status: StatusConnected}
	// Limit: 2 messages per 100ms
	rlChannel := NewRateLimitedChannel(mock, 2, 100*time.Millisecond)

	ctx := context.Background()
	msg := Message{Content: "test"}

	// 1. First send - should pass
	if err := rlChannel.Send(ctx, msg); err != nil {
		t.Errorf("First send failed: %v", err)
	}

	// 2. Second send - should pass
	if err := rlChannel.Send(ctx, msg); err != nil {
		t.Errorf("Second send failed: %v", err)
	}

	// 3. Third send - should fail
	if err := rlChannel.Send(ctx, msg); err == nil {
		t.Error("Third send should have failed due to rate limit")
	}

	// 4. Wait for window to pass
	time.Sleep(150 * time.Millisecond)

	// 5. Send after window - should pass
	if err := rlChannel.Send(ctx, msg); err != nil {
		t.Errorf("Send after wait failed: %v", err)
	}
}
