package server

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting for HTTP requests
// Uses token bucket algorithm per client IP
// Configuration:
//   - Rate: requests per second (default: 10)
//   - Burst: maximum burst size (default: 20)
//   - TTL: how long to keep inactive limiters (default: 5 minutes)
type RateLimiter struct {
	limiters map[string]*clientLimiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	ttl      time.Duration
}

// clientLimiter tracks rate limiter and last seen time for a client
type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter with specified parameters
func NewRateLimiter(r rate.Limit, burst int, ttl time.Duration) *RateLimiter {
	if r <= 0 {
		r = 10 // Default: 10 requests per second
	}
	if burst <= 0 {
		burst = 20 // Default: burst of 20
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute // Default: 5 minute TTL
	}

	rl := &RateLimiter{
		limiters: make(map[string]*clientLimiter),
		rate:     r,
		burst:    burst,
		ttl:      ttl,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Middleware returns an HTTP middleware that applies rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := rl.getClientID(r)

		limiter := rl.getLimiter(clientID)
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientID extracts a client identifier from the request
// Uses X-Forwarded-For header if present, otherwise RemoteAddr
func (rl *RateLimiter) getClientID(r *http.Request) string {
	// Check for forwarded IP (behind proxy/load balancer)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the chain
		if idx := len(forwarded); idx > 0 {
			for i := 0; i < len(forwarded); i++ {
				if forwarded[i] == ',' {
					forwarded = forwarded[:i]
					break
				}
			}
			return forwarded
		}
	}

	// Fall back to remote address
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// getLimiter returns (creating if necessary) a rate limiter for the client
func (rl *RateLimiter) getLimiter(clientID string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cl, exists := rl.limiters[clientID]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[clientID] = &clientLimiter{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	cl.lastSeen = time.Now()
	return cl.limiter
}

// cleanupLoop periodically removes inactive limiters
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes limiters that haven't been used recently
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for clientID, cl := range rl.limiters {
		if now.Sub(cl.lastSeen) > rl.ttl {
			delete(rl.limiters, clientID)
		}
	}
}

// DefaultRateLimiter creates a rate limiter with default settings
// 10 requests per second, burst of 20, 5 minute TTL
func DefaultRateLimiter() *RateLimiter {
	return NewRateLimiter(10, 20, 5*time.Minute)
}

// StrictRateLimiter creates a strict rate limiter for sensitive endpoints
// 1 request per second, burst of 3, 10 minute TTL
func StrictRateLimiter() *RateLimiter {
	return NewRateLimiter(1, 3, 10*time.Minute)
}
