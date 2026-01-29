package telegram

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/channels"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramChannel struct {
	id       string
	token    string
	bot      *tgbotapi.BotAPI
	eventBus *bus.Bus
	cancel   context.CancelFunc
	status   channels.Status
}

func NewTelegramChannel(id, token string, eventBus *bus.Bus) *TelegramChannel {
	return &TelegramChannel{
		id:       id,
		token:    token,
		eventBus: eventBus,
		status:   channels.StatusDisconnected,
	}
}

func (t *TelegramChannel) ID() string {
	return t.id
}

func (t *TelegramChannel) Type() string {
	return "telegram"
}

func (t *TelegramChannel) Connect(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI(t.token)
	if err != nil {
		t.status = channels.StatusError
		return fmt.Errorf("failed to create bot: %w", err)
	}

	t.bot = bot
	t.status = channels.StatusConnected

	// Create context for polling loop
	ctx, cancel := context.WithCancel(ctx)
	t.cancel = cancel

	// Subscribe to outbound messages
	if t.eventBus != nil {
		outbound, unsub := t.eventBus.Subscribe(bus.EventChannelOutboundMessage)
		go t.handleOutbound(ctx, outbound, unsub)
	}

	// Start polling in background
	go t.poll(ctx)

	return nil
}

func (t *TelegramChannel) Disconnect(ctx context.Context) error {
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}
	t.status = channels.StatusDisconnected
	return nil
}

func (t *TelegramChannel) Send(ctx context.Context, msg channels.Message) error {
	if t.bot == nil {
		return fmt.Errorf("bot not connected")
	}

	chatID, err := strconv.ParseInt(msg.ChannelID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid channel ID (chat ID) %s: %w", msg.ChannelID, err)
	}

	tgMsg := tgbotapi.NewMessage(chatID, msg.Content)
	_, err = t.bot.Send(tgMsg)
	return err
}

func (t *TelegramChannel) Status() channels.Status {
	return t.status
}

func (t *TelegramChannel) poll(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Manual polling loop for better control over errors and backoff
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			updates, err := t.bot.GetUpdates(u)
			if err != nil {
				// Log error via bus or status
				t.eventBus.Publish(bus.NewEvent(bus.EventErrorOccurred, "", map[string]interface{}{
					"channel_id": t.id,
					"error":      fmt.Sprintf("polling error: %v", err),
				}))

				// If it's a critical error (e.g. 401 Unauthorized), we should disconnect
				// For network errors, we simple retry (ticks continue)
				continue
			}

			for _, update := range updates {
				if update.UpdateID >= u.Offset {
					u.Offset = update.UpdateID + 1
				}

				if update.Message == nil {
					continue
				}

				t.handleMessage(update.Message)
			}
		}
	}
}

func (t *TelegramChannel) handleMessage(msg *tgbotapi.Message) {
	if t.eventBus == nil {
		return
	}

	content := msg.Text

	// Handle media types if Text is empty
	if content == "" {
		if msg.Photo != nil {
			content = fmt.Sprintf("[Media: Photo] %s", msg.Caption)
		} else if msg.Document != nil {
			content = fmt.Sprintf("[Media: Document] %s", msg.Caption)
		} else if msg.Voice != nil {
			content = "[Media: Voice]"
		} else if msg.Audio != nil {
			content = fmt.Sprintf("[Media: Audio] %s", msg.Caption)
		} else if msg.Sticker != nil {
			content = fmt.Sprintf("[Media: Sticker] %s", msg.Sticker.Emoji)
		} else if msg.Video != nil {
			content = fmt.Sprintf("[Media: Video] %s", msg.Caption)
		} else {
			content = "[Media: Unknown]"
		}
	}

	channelMsg := channels.Message{
		ID:        strconv.Itoa(msg.MessageID),
		Content:   content,
		Source:    t.id,
		ChannelID: strconv.FormatInt(msg.Chat.ID, 10),
		SenderID:  strconv.FormatInt(msg.From.ID, 10),
		CreatedAt: time.Unix(int64(msg.Date), 0),
		Metadata: map[string]string{
			"username":   msg.From.UserName,
			"first_name": msg.From.FirstName,
			"last_name":  msg.From.LastName,
		},
	}

	t.eventBus.Publish(bus.NewEvent(bus.EventChannelMessage, "", channelMsg))
}

func (t *TelegramChannel) handleOutbound(ctx context.Context, events <-chan bus.Event, unsub func()) {
	defer unsub()
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-events:
			if !ok {
				return
			}
			// Parse payload
			if payload, ok := evt.Payload.(map[string]interface{}); ok {
				source, _ := payload["source"].(string)
				if source != t.id {
					continue // Not for this channel instance
				}
				chatID, _ := payload["channel_id"].(string)
				content, _ := payload["content"].(string)
				if chatID != "" && content != "" {
					// Send message
					msg := channels.Message{
						ChannelID: chatID,
						Content:   content,
					}
					_ = t.Send(context.Background(), msg)
				}
			}
		}
	}
}
