package telegram

import (
	"context"
	"testing"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/channels"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestTelegramChannel_Lifecycle(t *testing.T) {
	b := bus.New()
	tc := NewTelegramChannel("telegram-1", "invalid-token", b)

	if tc.ID() != "telegram-1" {
		t.Errorf("Expected ID telegram-1, got %s", tc.ID())
	}
	if tc.Type() != "telegram" {
		t.Errorf("Expected Type telegram, got %s", tc.Type())
	}
	if tc.Status() != channels.StatusDisconnected {
		t.Errorf("Expected Initial Status Disconnected, got %s", tc.Status())
	}

	// Connect should fail with invalid token (and no network to telegram)
	// This verifies Connect attempts to use the token
	err := tc.Connect(context.Background())
	if err == nil {
		t.Error("Expected error connecting with invalid token")
	}
	if tc.Status() != channels.StatusError {
		t.Errorf("Expected Status Error after failed connect, got %s", tc.Status())
	}
}

func TestTelegramChannel_HandleMessage(t *testing.T) {
	b := bus.New()
	tc := NewTelegramChannel("telegram-1", "token", b)

	// Subscribe to bus
	msgCh, _ := b.Subscribe(bus.EventChannelMessage)

	// Manually invoke handleMessage (internal method, accessible in same package test)
	now := int(time.Now().Unix())
	tgMsg := &tgbotapi.Message{
		MessageID: 123,
		Text:      "Hello World",
		Chat:      &tgbotapi.Chat{ID: 456},
		From:      &tgbotapi.User{ID: 789, UserName: "testuser", FirstName: "Test"},
		Date:      now,
	}

	tc.handleMessage(tgMsg)

	select {
	case event := <-msgCh:
		msg, ok := event.Payload.(channels.Message)
		if !ok {
			t.Fatalf("Expected channels.Message payload")
		}
		if msg.Content != "Hello World" {
			t.Errorf("Expected content 'Hello World', got '%s'", msg.Content)
		}
		if msg.ChannelID != "456" {
			t.Errorf("Expected ChannelID 456, got %s", msg.ChannelID)
		}
		if msg.SenderID != "789" {
			t.Errorf("Expected SenderID 789, got %s", msg.SenderID)
		}
		if msg.Metadata["username"] != "testuser" {
			t.Errorf("Expected metadata username testuser, got %s", msg.Metadata["username"])
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for event")
	}
}

func TestTelegramChannel_HandleMedia(t *testing.T) {
	b := bus.New()
	tc := NewTelegramChannel("telegram-1", "token", b)

	msgCh, _ := b.Subscribe(bus.EventChannelMessage)
	now := int(time.Now().Unix())

	tests := []struct {
		name     string
		msg      *tgbotapi.Message
		expected string
	}{
		{
			name: "Photo",
			msg: &tgbotapi.Message{
				MessageID: 1,
				Chat:      &tgbotapi.Chat{ID: 1},
				From:      &tgbotapi.User{ID: 1},
				Date:      now,
				Photo:     []tgbotapi.PhotoSize{{FileID: "123"}},
				Caption:   "Look at this",
			},
			expected: "[Media: Photo] Look at this",
		},
		{
			name: "Voice",
			msg: &tgbotapi.Message{
				MessageID: 2,
				Chat:      &tgbotapi.Chat{ID: 1},
				From:      &tgbotapi.User{ID: 1},
				Date:      now,
				Voice:     &tgbotapi.Voice{FileID: "123"},
			},
			expected: "[Media: Voice]",
		},
		{
			name: "Unknown",
			msg: &tgbotapi.Message{
				MessageID: 3,
				Chat:      &tgbotapi.Chat{ID: 1},
				From:      &tgbotapi.User{ID: 1},
				Date:      now,
			},
			expected: "[Media: Unknown]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc.handleMessage(tt.msg)
			select {
			case event := <-msgCh:
				msg, ok := event.Payload.(channels.Message)
				if !ok {
					t.Fatalf("Expected channels.Message payload")
				}
				if msg.Content != tt.expected {
					t.Errorf("Expected '%s', got '%s'", tt.expected, msg.Content)
				}
			case <-time.After(100 * time.Millisecond):
				t.Error("Timeout waiting for event")
			}
		})
	}
}
