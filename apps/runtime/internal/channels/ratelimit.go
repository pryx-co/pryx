package channels

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type RateLimiter struct {
	limit    int
	interval time.Duration

	mu          sync.Mutex
	count       int
	windowStart time.Time
}

func NewRateLimiter(limit int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:       limit,
		interval:    interval,
		count:       0,
		windowStart: time.Now(),
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if now.Sub(rl.windowStart) > rl.interval {
		rl.windowStart = now
		rl.count = 0
	}

	if rl.count < rl.limit {
		rl.count++
		return true
	}

	return false
}

// RateLimitedChannel wraps a Channel with rate limiting on Send
type RateLimitedChannel struct {
	Channel
	limiter *RateLimiter
}

func NewRateLimitedChannel(c Channel, limit int, interval time.Duration) *RateLimitedChannel {
	return &RateLimitedChannel{
		Channel: c,
		limiter: NewRateLimiter(limit, interval),
	}
}

func (rc *RateLimitedChannel) Send(ctx context.Context, msg Message) error {
	if !rc.limiter.Allow() {
		return fmt.Errorf("rate limit exceeded")
	}
	return rc.Channel.Send(ctx, msg)
}
