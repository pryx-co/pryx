package slack

import (
	"context"
	"fmt"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/channels"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type SlackChannel struct {
	id       string
	botToken string
	appToken string
	client   *slack.Client
	socket   *socketmode.Client
	eventBus *bus.Bus
	cancel   context.CancelFunc
	status   channels.Status
}

func NewSlackChannel(id, botToken, appToken string, eventBus *bus.Bus) *SlackChannel {
	return &SlackChannel{
		id:       id,
		botToken: botToken,
		appToken: appToken,
		eventBus: eventBus,
		status:   channels.StatusDisconnected,
	}
}

func (s *SlackChannel) ID() string {
	return s.id
}

func (s *SlackChannel) Type() string {
	return "slack"
}

func (s *SlackChannel) Connect(ctx context.Context) error {
	// Create Slack client
	client := slack.New(s.botToken, slack.OptionAppLevelToken(s.appToken))

	// Test auth
	_, err := client.AuthTest()
	if err != nil {
		s.status = channels.StatusError
		return fmt.Errorf("slack auth failed: %w", err)
	}

	s.client = client

	// Create socket mode client for real-time events
	s.socket = socketmode.New(
		client,
		socketmode.OptionDebug(false),
	)

	s.status = channels.StatusConnected

	// Create context for management
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	// Start socket mode in background
	go s.runSocketMode(ctx)

	// Subscribe to outbound messages
	if s.eventBus != nil {
		outbound, unsub := s.eventBus.Subscribe(bus.EventChannelOutboundMessage)
		go s.handleOutbound(ctx, outbound, unsub)
	}

	return nil
}

func (s *SlackChannel) Disconnect(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}

	s.status = channels.StatusDisconnected
	return nil
}

func (s *SlackChannel) Send(ctx context.Context, msg channels.Message) error {
	if s.client == nil {
		return fmt.Errorf("slack client not connected")
	}

	_, _, err := s.client.PostMessage(
		msg.ChannelID,
		slack.MsgOptionText(msg.Content, false),
	)
	return err
}

func (s *SlackChannel) Status() channels.Status {
	return s.status
}

func (s *SlackChannel) runSocketMode(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-s.socket.Events:
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				s.status = channels.StatusConnecting
			case socketmode.EventTypeConnectionError:
				s.status = channels.StatusError
			case socketmode.EventTypeConnected:
				s.status = channels.StatusConnected
			case socketmode.EventTypeEventsAPI:
				if enve, ok := evt.Data.(slackevents.EventsAPIEvent); ok {
					s.handleEventsAPI(enve)
				}
			}
		}
	}
}

func (s *SlackChannel) handleEventsAPI(event slackevents.EventsAPIEvent) {
	switch event.Type {
	case "message":
		if messageEvent, ok := event.Data.(*slackevents.MessageEvent); ok {
			// Ignore messages from the bot itself
			if messageEvent.BotID != "" {
				return
			}

			// Create channel message
			msg := channels.Message{
				ID:        messageEvent.TimeStamp,
				Content:   messageEvent.Text,
				Source:    s.id,
				ChannelID: messageEvent.Channel,
				SenderID:  messageEvent.User,
				Metadata: map[string]string{
					"channel": messageEvent.Channel,
				},
				CreatedAt: time.Now(),
			}

			// Publish to event bus
			if s.eventBus != nil {
				s.eventBus.Publish(bus.Event{
					Type:    bus.EventChannelMessage,
					Payload: msg,
				})
			}
		}

	case "app_mention":
		if mentionEvent, ok := event.Data.(*slackevents.AppMentionEvent); ok {
			msg := channels.Message{
				ID:        mentionEvent.TimeStamp,
				Content:   mentionEvent.Text,
				Source:    s.id,
				ChannelID: mentionEvent.Channel,
				SenderID:  mentionEvent.User,
				Metadata: map[string]string{
					"channel": mentionEvent.Channel,
					"mention": "true",
				},
				CreatedAt: time.Now(),
			}

			if s.eventBus != nil {
				s.eventBus.Publish(bus.Event{
					Type:    bus.EventChannelMessage,
					Payload: msg,
				})
			}
		}
	}
}

func (s *SlackChannel) handleOutbound(ctx context.Context, outbound <-chan bus.Event, unsub func()) {
	defer unsub()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-outbound:
			if msg, ok := event.Payload.(channels.Message); ok {
				if msg.Source == s.id {
					// This is our message to send
					if err := s.Send(ctx, msg); err != nil {
						// Log error but don't crash
						fmt.Printf("Slack send error: %v\n", err)
					}
				}
			}
		}
	}
}
