package channels

import (
	"context"
	"time"
)

type Status string

const (
	StatusDisconnected Status = "disconnected"
	StatusConnecting   Status = "connecting"
	StatusConnected    Status = "connected"
	StatusError        Status = "error"
)

type Message struct {
	ID        string            `json:"id"`
	Content   string            `json:"content"`
	ChannelID string            `json:"channel_id"`
	SenderID  string            `json:"sender_id"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type Channel interface {
	ID() string
	Type() string
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Send(ctx context.Context, msg Message) error
	Status() Status
}
