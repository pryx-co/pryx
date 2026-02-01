package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"nhooyr.io/websocket"
	"pryx-core/internal/bus"
	"pryx-core/internal/validation"
)

const (
	WebSocketBufferSize       = 256
	defaultMaxMessageSize     = 10 * 1024 * 1024 // 10MB
	defaultMaxConnections     = 1000
	defaultRateLimitPerMinute = 60
)

// wsConnectionPool tracks active WebSocket connections
var (
	activeConnections   = make(map[string]bool)
	connectionPoolMutex sync.RWMutex
	rateLimiters        = make(map[string]*rate.Limiter)
	rateLimitMutex      sync.RWMutex
)

// getRateLimiter returns or creates a rate limiter for an IP
func getRateLimiter(ip string, rps rate.Limit) *rate.Limiter {
	rateLimitMutex.Lock()
	defer rateLimitMutex.Unlock()

	if limiter, exists := rateLimiters[ip]; exists {
		return limiter
	}

	// Create new limiter with burst of 5
	limiter := rate.NewLimiter(rps, 5)
	rateLimiters[ip] = limiter
	return limiter
}

// cleanupOldRateLimiters removes stale rate limiters periodically
func cleanupOldRateLimiters() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rateLimitMutex.Lock()
		// Clear old entries - this is a simple approach
		// In production, you'd track last access time
		if len(rateLimiters) > 10000 {
			rateLimiters = make(map[string]*rate.Limiter)
		}
		rateLimitMutex.Unlock()
	}
}

func init() {
	go cleanupOldRateLimiters()
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	cfg := s.cfg

	// Apply rate limiting
	ip := getClientIP(r)
	rateLimitPerMinute := cfg.WebSocketRateLimitPerMinute
	if rateLimitPerMinute <= 0 {
		rateLimitPerMinute = defaultRateLimitPerMinute
	}

	rps := rate.Limit(float64(rateLimitPerMinute) / 60.0)
	limiter := getRateLimiter(ip, rps)

	if !limiter.Allow() {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	// Check max connections limit
	maxConns := cfg.MaxWebSocketConnections
	if maxConns <= 0 {
		maxConns = defaultMaxConnections
	}

	connectionPoolMutex.Lock()
	if len(activeConnections) >= maxConns {
		connectionPoolMutex.Unlock()
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	connID := generateConnectionID(ip)
	activeConnections[connID] = true
	connectionPoolMutex.Unlock()

	defer func() {
		connectionPoolMutex.Lock()
		delete(activeConnections, connID)
		connectionPoolMutex.Unlock()
	}()

	// Setup WebSocket accept options with origin validation
	allowedOrigins := cfg.WebSocketAllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = cfg.AllowedOrigins
	}

	// Accept WebSocket with origin validation
	acceptOpts := &websocket.AcceptOptions{
		InsecureSkipVerify: false,
	}

	// Only set OriginPatterns if we have specific origins configured
	// Otherwise, use default behavior which validates against the Host header
	if len(allowedOrigins) > 0 {
		acceptOpts.OriginPatterns = allowedOrigins
	}

	c, err := websocket.Accept(w, r, acceptOpts)
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal error")

	// Set read limit for message size
	maxMessageSize := cfg.MaxWebSocketMessageSize
	if maxMessageSize <= 0 {
		maxMessageSize = defaultMaxMessageSize
	}
	c.SetReadLimit(maxMessageSize)

	query := r.URL.Query()
	surface := strings.TrimSpace(query.Get("surface"))
	sessionFilter := strings.TrimSpace(query.Get("session_id"))
	eventFilters := query["event"]

	validator := validation.NewValidator()
	if err := validator.ValidateSessionID(sessionFilter); err != nil {
		return
	}

	var topics []bus.EventType
	for _, ev := range eventFilters {
		ev = strings.TrimSpace(ev)
		if ev == "" {
			continue
		}
		topics = append(topics, bus.EventType(ev))
	}

	var events <-chan bus.Event
	var cancel func()
	if len(topics) == 0 {
		events, cancel = s.bus.Subscribe()
	} else {
		events, cancel = s.bus.Subscribe(topics...)
	}
	defer cancel()

	ctx := r.Context()

	s.bus.Publish(bus.NewEvent(bus.EventTraceEvent, sessionFilter, map[string]interface{}{
		"kind":        "ws.connected",
		"remote_addr": r.RemoteAddr,
		"surface":     surface,
	}))

	// Use buffered channel for event distribution
	eventCh := make(chan bus.Event, WebSocketBufferSize)

	// Event pump goroutine with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.bus.Publish(bus.NewEvent(bus.EventErrorOccurred, sessionFilter, map[string]interface{}{
					"kind":  "ws.event_pump.panic",
					"error": r,
				}))
			}
			close(eventCh)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-events:
				if !ok {
					return
				}
				if sessionFilter != "" && evt.SessionID != sessionFilter {
					continue
				}
				select {
				case eventCh <- evt:
				default:
					// Channel full, drop event
				}
			}
		}
	}()

	// Writer goroutine with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.bus.Publish(bus.NewEvent(bus.EventErrorOccurred, sessionFilter, map[string]interface{}{
					"kind":  "ws.writer.panic",
					"error": r,
				}))
			}
		}()

		for evt := range eventCh {
			bytes, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			if err := c.Write(ctx, websocket.MessageText, bytes); err != nil {
				return
			}
		}
	}()

	// Main read loop with panic recovery
	defer func() {
		if r := recover(); r != nil {
			s.bus.Publish(bus.NewEvent(bus.EventErrorOccurred, sessionFilter, map[string]interface{}{
				"kind":  "ws.reader.panic",
				"error": r,
			}))
		}
	}()

	for {
		msgType, data, err := c.Read(ctx)
		if err != nil {
			break
		}
		if msgType != websocket.MessageText {
			continue
		}

		// Check message size
		if int64(len(data)) > maxMessageSize {
			s.bus.Publish(bus.NewEvent(bus.EventErrorOccurred, sessionFilter, map[string]interface{}{
				"kind":     "ws.message_too_large",
				"size":     len(data),
				"max_size": maxMessageSize,
			}))
			continue
		}

		in := struct {
			Event      string                 `json:"event"`
			Type       string                 `json:"type"`
			SessionID  string                 `json:"session_id"`
			Payload    map[string]interface{} `json:"payload"`
			ApprovalID string                 `json:"approval_id"`
			Approved   bool                   `json:"approved"`
		}{}
		if err := json.Unmarshal(data, &in); err != nil {
			continue
		}

		eventType := in.Event
		if eventType == "" {
			eventType = in.Type
		}

		switch eventType {
		case "approval.resolve":
			approvalID := strings.TrimSpace(in.ApprovalID)
			if err := validator.ValidateID("approval_id", approvalID); err == nil {
				_ = s.mcp.ResolveApproval(approvalID, in.Approved)
			}
		case "chat.send":
			if in.Payload != nil && in.Payload["content"] != nil {
				if content, ok := in.Payload["content"].(string); ok {
					if err := validator.ValidateChatContent(content); err == nil {
						s.bus.Publish(bus.NewEvent(bus.EventChatRequest, sessionFilter, in.Payload))
					}
				}
			}
		}
	}

	s.bus.Publish(bus.NewEvent(bus.EventTraceEvent, sessionFilter, map[string]interface{}{
		"kind":        "ws.disconnected",
		"remote_addr": r.RemoteAddr,
		"surface":     surface,
	}))

	c.Close(websocket.StatusNormalClosure, "")
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-Ip header
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// generateConnectionID creates a unique connection ID
func generateConnectionID(ip string) string {
	return ip + "-" + time.Now().Format("20060102150405") + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
}
