package universal

import (
	"encoding/json"
	"time"
)

// Translator converts between different message formats
type Translator struct{}

// NewTranslator creates a new translator
func NewTranslator() *Translator {
	return &Translator{}
}

// OpenClawMessage represents an OpenClaw gateway message
type OpenClawMessage struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id,omitempty"`
	From      string                 `json:"from,omitempty"`
	To        string                 `json:"to,omitempty"`
	Action    string                 `json:"action"`
	Payload   map[string]interface{} `json:"payload"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
}

// ToUniversal converts an OpenClaw message to universal format
func (t *Translator) ToUniversal(openclaw *OpenClawMessage) *UniversalMessage {
	return &UniversalMessage{
		ID:          openclaw.ID,
		TraceID:     CorrelationID(),
		SpanID:      CorrelationID(),
		Protocol:    "openclaw",
		MessageType: t.mapMessageType(openclaw.Type),
		Action:      openclaw.Action,
		Payload:     openclaw.Payload,
		Metadata:    map[string]interface{}{"openclaw_type": openclaw.Type},
		Timestamp:   time.Now().UTC(),
		Context:     openclaw.Metadata,
		From: AgentIdentity{
			ID:   openclaw.From,
			Name: openclaw.From,
		},
		To: AgentIdentity{
			ID:   openclaw.To,
			Name: openclaw.To,
		},
	}
}

// ToOpenClaw converts a universal message to OpenClaw format
func (t *Translator) ToOpenClaw(universal *UniversalMessage) *OpenClawMessage {
	openclawType := "message"
	if universal.MessageType == MessageTypeRequest {
		openclawType = "request"
	} else if universal.MessageType == MessageTypeResponse {
		openclawType = "response"
	} else if universal.MessageType == MessageTypeEvent {
		openclawType = "event"
	}

	return &OpenClawMessage{
		Type:      openclawType,
		ID:        universal.ID,
		SessionID: "",
		From:      universal.From.ID,
		To:        universal.To.ID,
		Action:    universal.Action,
		Payload:   universal.Payload,
		Metadata:  universal.Context,
	}
}

// mapMessageType maps OpenClaw message types to universal types
func (t *Translator) mapMessageType(openclawType string) string {
	switch openclawType {
	case "request":
		return MessageTypeRequest
	case "response":
		return MessageTypeResponse
	case "event":
		return MessageTypeEvent
	case "stream":
		return MessageTypeStream
	default:
		return MessageTypeRequest
	}
}

// DetectVersion attempts to detect the protocol version from raw data
func (t *Translator) DetectVersion(data []byte) string {
	var msg map[string]interface{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return "unknown"
	}

	// Check for version field
	if version, ok := msg["version"].(string); ok {
		return version
	}

	// Check for common version indicators
	if _, ok := msg["handshake"]; ok {
		return "1.0"
	}

	return "1.0"
}

// NegotiateCapabilities compares capabilities and returns intersection
func (t *Translator) NegotiateCapabilities(local, remote []string) []string {
	capabilitySet := make(map[string]bool)
	for _, cap := range local {
		capabilitySet[cap] = true
	}

	var intersection []string
	for _, cap := range remote {
		if capabilitySet[cap] {
			intersection = append(intersection, cap)
		}
	}

	return intersection
}
