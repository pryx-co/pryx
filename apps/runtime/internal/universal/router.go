package universal

import (
	"context"
	"sync"
)

// MessageRouter routes messages between agents
type MessageRouter struct {
	mu          sync.RWMutex
	routes      map[string]*Route
	subscribers map[string][]chan *UniversalMessage
	running     bool
	stopCh      chan struct{}
}

// Route represents a message route
type Route struct {
	FromAgent string
	ToAgent   string
	Pattern   string
	Handler   func(*UniversalMessage) error
}

// NewMessageRouter creates a new message router
func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		routes:      make(map[string]*Route),
		subscribers: make(map[string][]chan *UniversalMessage),
		stopCh:      make(chan struct{}),
	}
}

// Start initializes the message router
func (mr *MessageRouter) Start(ctx context.Context) {
	mr.running = true
}

// Stop gracefully shuts down the message router
func (mr *MessageRouter) Stop(ctx context.Context) {
	mr.running = false
	close(mr.stopCh)
}

// Route routes a message to its destination
func (mr *MessageRouter) Route(ctx context.Context, msg *UniversalMessage) (bool, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	routeKey := mr.getRouteKey(msg.From.ID, msg.To.ID)
	if route, exists := mr.routes[routeKey]; exists {
		if route.Handler != nil {
			if err := route.Handler(msg); err != nil {
				return false, err
			}
		}
		return true, nil
	}

	return false, nil
}

// Subscribe subscribes to messages matching a pattern
func (mr *MessageRouter) Subscribe(pattern string) chan *UniversalMessage {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	ch := make(chan *UniversalMessage, 100)
	mr.subscribers[pattern] = append(mr.subscribers[pattern], ch)
	return ch
}

// Unsubscribe removes a subscription
func (mr *MessageRouter) Unsubscribe(pattern string, ch chan *UniversalMessage) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	subscribers := mr.subscribers[pattern]
	for i, sub := range subscribers {
		if sub == ch {
			mr.subscribers[pattern] = append(subscribers[:i], subscribers[i+1:]...)
			close(ch)
			break
		}
	}
}

// AddRoute adds a static route
func (mr *MessageRouter) AddRoute(fromAgent, toAgent, pattern string, handler func(*UniversalMessage) error) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	routeKey := mr.getRouteKey(fromAgent, toAgent)
	mr.routes[routeKey] = &Route{
		FromAgent: fromAgent,
		ToAgent:   toAgent,
		Pattern:   pattern,
		Handler:   handler,
	}
}

// RemoveRoute removes a route
func (mr *MessageRouter) RemoveRoute(fromAgent, toAgent string) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	routeKey := mr.getRouteKey(fromAgent, toAgent)
	delete(mr.routes, routeKey)
}

// getRouteKey generates a route key
func (mr *MessageRouter) getRouteKey(fromAgent, toAgent string) string {
	return fromAgent + "->" + toAgent
}

// RouteCount returns the number of routes
func (mr *MessageRouter) RouteCount() int {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return len(mr.routes)
}
