package webhook

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/channels"
)

func TestWebhookChannel_Send_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "application/json") {
			t.Errorf("expected JSON content type, got %s", ct)
		}

		sig := r.Header.Get("X-Webhook-Signature")
		if sig != "" {
			t.Logf("received signature: %s", sig)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	eventBus := bus.New()
	config := WebhookConfig{
		ID:        "test-webhook",
		Secret:    "test-secret",
		TargetURL: server.URL,
		Retries:   3,
	}

	w := NewWebhookChannel(config, eventBus)
	// Mock connected status for sending (since manual connect isn't called)
	// Send checks if TargetURL is set, doesn't strictly check w.status currently,
	// but strictly speaking we should probably be connected.
	// The implementation of Send uses w.config directly.

	msg := channels.Message{
		ID:        "msg-1",
		Content:   "test content",
		ChannelID: "test-channel",
		CreatedAt: time.Now(),
		Metadata: map[string]string{
			"Content-Type": "application/json",
		},
	}

	if err := w.Send(context.Background(), msg); err != nil {
		t.Fatalf("failed to send message: %v", err)
	}
}

func TestWebhookChannel_Send_FormData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "application/x-www-form-urlencoded") {
			t.Errorf("expected form-data content type, got %s", ct)
		}

		if err := r.ParseForm(); err != nil {
			t.Errorf("failed to parse form: %v", err)
			return
		}

		if r.FormValue("content") != "test content" {
			t.Errorf("unexpected content: %s", r.FormValue("content"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	eventBus := bus.New()
	config := WebhookConfig{
		ID:        "test-webhook",
		Secret:    "test-secret",
		TargetURL: server.URL,
		Retries:   3,
	}

	w := NewWebhookChannel(config, eventBus)

	msg := channels.Message{
		ID:        "msg-1",
		Content:   "test content",
		ChannelID: "test-channel",
		CreatedAt: time.Now(),
		Metadata: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
	}

	ctx := context.Background()
	if err := w.Send(ctx, msg); err != nil {
		t.Fatalf("failed to send message: %v", err)
	}
}

func TestWebhookChannel_RetryLogic(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		if attempts <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	eventBus := bus.New()
	config := WebhookConfig{
		ID:        "test-webhook",
		Secret:    "test-secret",
		TargetURL: server.URL,
		Retries:   3,
	}

	w := NewWebhookChannel(config, eventBus)

	msg := channels.Message{
		ID:        "msg-1",
		Content:   "test content",
		ChannelID: "test",
		CreatedAt: time.Now(),
		Metadata: map[string]string{
			"Content-Type": "application/json",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := w.Send(ctx, msg); err != nil {
		t.Logf("send failed after retries: %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 retry attempts, got %d", attempts)
	}
}
