package agentbus

import (
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	// CircuitBreakerClosed allows requests to pass through
	CircuitBreakerClosed CircuitBreakerState = "closed"
	// CircuitBreakerOpen blocks all requests
	CircuitBreakerOpen CircuitBreakerState = "open"
	// CircuitBreakerHalfOpen allows limited requests to test recovery
	CircuitBreakerHalfOpen CircuitBreakerState = "half_open"
)

// CircuitBreaker prevents cascading failures by stopping requests to failing services
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	lastSuccessTime time.Time
	config          CircuitBreakerConfig
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 5
	}
	if config.RecoveryTimeout == 0 {
		config.RecoveryTimeout = 30 * time.Second
	}
	if config.HalfOpenRequests == 0 {
		config.HalfOpenRequests = 3
	}

	return &CircuitBreaker{
		state:  CircuitBreakerClosed,
		config: config,
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// AllowRequest checks if a request should be allowed
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		// Check if recovery timeout has elapsed
		if time.Since(cb.lastFailureTime) > cb.config.RecoveryTimeout {
			cb.state = CircuitBreakerHalfOpen
			cb.successCount = 0
			cb.failureCount = 0
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		return cb.successCount < cb.config.HalfOpenRequests
	}
	return false
}

// RecordSuccess marks a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastSuccessTime = time.Now()
	cb.successCount++

	switch cb.state {
	case CircuitBreakerHalfOpen:
		// If enough successes, close the circuit
		if cb.successCount >= cb.config.HalfOpenRequests {
			cb.state = CircuitBreakerClosed
			cb.failureCount = 0
		}
	case CircuitBreakerClosed:
		// Reset failure count on success (optional strategy)
		// cb.failureCount = 0
	}
}

// RecordFailure marks a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureTime = time.Now()
	cb.failureCount++

	switch cb.state {
	case CircuitBreakerClosed:
		// Open the circuit if threshold reached
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.state = CircuitBreakerOpen
		}
	case CircuitBreakerHalfOpen:
		// Any failure in half-open state opens the circuit
		cb.state = CircuitBreakerOpen
	}
}

// FailureCount returns the current failure count
func (cb *CircuitBreaker) FailureCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failureCount
}

// SuccessCount returns the current success count
func (cb *CircuitBreaker) SuccessCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.successCount
}

// Reset clears the circuit breaker state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitBreakerClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastFailureTime = time.Time{}
	cb.lastSuccessTime = time.Time{}
}

// ForceOpen forces the circuit breaker to open
func (cb *CircuitBreaker) ForceOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitBreakerOpen
	cb.lastFailureTime = time.Now()
}

// ForceClose forces the circuit breaker to close
func (cb *CircuitBreaker) ForceClose() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitBreakerClosed
	cb.failureCount = 0
	cb.successCount = 0
}

// String returns a string representation of the circuit breaker
func (cb *CircuitBreaker) String() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return string(cb.state)
}
