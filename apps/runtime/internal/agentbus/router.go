package agentbus

import (
	"context"
	"sync"

	"pryx-core/internal/bus"
)

// MessageRouter routes messages between agents
type MessageRouter struct {
	mu          sync.RWMutex
	bus         *bus.Bus
	logger      *StructuredLogger
	routes      map[string]*Route
	subscribers map[string][]chan *UniversalMessage
	broadcast   chan *UniversalMessage
	running     bool
	stopCh      chan struct{}
}

// Route represents a message route
type Route struct {
	FromAgent string
	ToAgent   string
	Pattern   string
	Priority  int
	Handler   func(*UniversalMessage) error
}

// NewMessageRouter creates a new message router
func NewMessageRouter(b *bus.Bus) *MessageRouter {
	return &MessageRouter{
		bus:         b,
		logger:      NewStructuredLogger("router", "info"),
		routes:      make(map[string]*Route),
		subscribers: make(map[string][]chan *UniversalMessage),
		broadcast:   make(chan *UniversalMessage, 1000),
		stopCh:      make(chan struct{}),
	}
}

// Start initializes the message router
func (mr *MessageRouter) Start(ctx context.Context) error {
	mr.mu.Lock()
	if mr.running {
		mr.mu.Unlock()
		return nil
	}
	mr.running = true
	mr.mu.Unlock()

	mr.logger.Info("message router started", nil)
	mr.bus.Publish(bus.NewEvent("agentbus.router.started", "", nil))

	return nil
}

// Stop gracefully shuts down the message router
func (mr *MessageRouter) Stop(ctx context.Context) error {
	mr.mu.Lock()
	if !mr.running {
		mr.mu.Unlock()
		return nil
	}
	mr.running = false
	mr.mu.Unlock()

	close(mr.stopCh)

	mr.logger.Info("message router stopped", nil)
	mr.bus.Publish(bus.NewEvent("agentbus.router.stopped", "", nil))

	return nil
}

// Route routes a message to its destination
func (mr *MessageRouter) Route(ctx context.Context, msg *UniversalMessage) (bool, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	// Direct routing by agent ID
	routeKey := mr.getRouteKey(msg.From.ID, msg.To.ID)
	if route, exists := mr.routes[routeKey]; exists {
		if route.Handler != nil {
			if err := route.Handler(msg); err != nil {
				return false, err
			}
		}
		return true, nil
	}

	// Check for wildcard routes
	wildcardKey := mr.getRouteKey(msg.From.ID, "*")
	if route, exists := mr.routes[wildcardKey]; exists {
		if route.Handler != nil {
			if err := route.Handler(msg); err != nil {
				return false, err
			}
		}
		return true, nil
	}

	// Broadcast routing
	mr.broadcast <- msg

	// Notify subscribers
	mr.notifySubscribers(msg)

	return false, nil // No direct route found
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
func (mr *MessageRouter) AddRoute(fromAgent, toAgent, pattern string, priority int, handler func(*UniversalMessage) error) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	routeKey := mr.getRouteKey(fromAgent, toAgent)
	mr.routes[routeKey] = &Route{
		FromAgent: fromAgent,
		ToAgent:   toAgent,
		Pattern:   pattern,
		Priority:  priority,
		Handler:   handler,
	}

	mr.logger.Debug("route added", map[string]interface{}{
		"from":    fromAgent,
		"to":      toAgent,
		"pattern": pattern,
	})
}

// RemoveRoute removes a route
func (mr *MessageRouter) RemoveRoute(fromAgent, toAgent string) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	routeKey := mr.getRouteKey(fromAgent, toAgent)
	delete(mr.routes, routeKey)
}

// GetRoutes returns all routes
func (mr *MessageRouter) GetRoutes() []*Route {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	routes := make([]*Route, 0, len(mr.routes))
	for _, route := range mr.routes {
		routes = append(routes, route)
	}

	return routes
}

// Broadcast sends a message to all subscribers
func (mr *MessageRouter) Broadcast(msg *UniversalMessage) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	mr.broadcast <- msg
	mr.notifySubscribers(msg)
}

// notifySubscribers notifies all subscribers matching the message
func (mr *MessageRouter) notifySubscribers(msg *UniversalMessage) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	for pattern, subscribers := range mr.subscribers {
		if mr.matchesPattern(msg, pattern) {
			for _, ch := range subscribers {
				select {
				case ch <- msg:
				default:
					// Subscriber buffer full, skip
				}
			}
		}
	}
}

// matchesPattern checks if a message matches a subscription pattern
func (mr *MessageRouter) matchesPattern(msg *UniversalMessage, pattern string) bool {
	// Simple pattern matching - can be extended for more complex patterns
	if pattern == "*" {
		return true
	}
	if pattern == msg.To.ID {
		return true
	}
	if pattern == msg.From.ID {
		return true
	}
	if pattern == msg.Action {
		return true
	}
	return false
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

// SubscriberCount returns the number of subscribers
func (mr *MessageRouter) SubscriberCount() int {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return len(mr.subscribers)
}
