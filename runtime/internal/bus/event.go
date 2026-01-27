package bus

import "time"

// EventType represents the type of event
type EventType string

const (
	EventSessionMessage   EventType = "session.message"
	EventSessionTyping    EventType = "session.typing"
	EventToolRequest      EventType = "tool.request"
	EventToolExecuting    EventType = "tool.executing"
	EventToolComplete     EventType = "tool.complete"
	EventApprovalNeeded   EventType = "approval.needed"
	EventApprovalResolved EventType = "approval.resolved"
	EventTraceEvent       EventType = "trace.event"
	EventErrorOccurred    EventType = "error.occurred"
)

// Event represents a single event in the system
type Event struct {
	Type      EventType   `json:"type"`
	Event     EventType   `json:"event"` // Redundant but consistent with PRD example
	SessionID string      `json:"session_id,omitempty"`
	Surface   string      `json:"surface,omitempty"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
	Version   int         `json:"version"`
}

// NewEvent creates a new event with the current timestamp
func NewEvent(eventType EventType, sessionID string, payload interface{}) Event {
	return Event{
		Type:      "event",   // envelope type
		Event:     eventType, // specific event type
		SessionID: sessionID,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
		Version:   1,
	}
}
