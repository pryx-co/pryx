package discord

import (
	"context"
	"fmt"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/channels"

	"github.com/bwmarrin/discordgo"
)

type DiscordChannel struct {
	id       string
	token    string
	session  *discordgo.Session
	eventBus *bus.Bus
	cancel   context.CancelFunc
	status   channels.Status
}

func NewDiscordChannel(id, token string, eventBus *bus.Bus) *DiscordChannel {
	return &DiscordChannel{
		id:       id,
		token:    token,
		eventBus: eventBus,
		status:   channels.StatusDisconnected,
	}
}

func (d *DiscordChannel) ID() string {
	return d.id
}

func (d *DiscordChannel) Type() string {
	return "discord"
}

func (d *DiscordChannel) Connect(ctx context.Context) error {
	session, err := discordgo.New("Bot " + d.token)
	if err != nil {
		d.status = channels.StatusError
		return fmt.Errorf("failed to create discord session: %w", err)
	}

	// Set up message handler
	session.AddHandler(d.handleMessage)

	// Open connection
	if err := session.Open(); err != nil {
		d.status = channels.StatusError
		return fmt.Errorf("failed to open discord connection: %w", err)
	}

	d.session = session
	d.status = channels.StatusConnected

	// Create context for management
	ctx, cancel := context.WithCancel(ctx)
	d.cancel = cancel

	// Subscribe to outbound messages
	if d.eventBus != nil {
		outbound, unsub := d.eventBus.Subscribe(bus.EventChannelOutboundMessage)
		go d.handleOutbound(ctx, outbound, unsub)
	}

	return nil
}

func (d *DiscordChannel) Disconnect(ctx context.Context) error {
	if d.cancel != nil {
		d.cancel()
		d.cancel = nil
	}

	if d.session != nil {
		if err := d.session.Close(); err != nil {
			return fmt.Errorf("failed to close discord session: %w", err)
		}
		d.session = nil
	}

	d.status = channels.StatusDisconnected
	return nil
}

func (d *DiscordChannel) Send(ctx context.Context, msg channels.Message) error {
	if d.session == nil {
		return fmt.Errorf("discord session not connected")
	}

	_, err := d.session.ChannelMessageSend(msg.ChannelID, msg.Content)
	return err
}

func (d *DiscordChannel) Status() channels.Status {
	return d.status
}

func (d *DiscordChannel) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Only handle direct messages or mentions
	isDM := m.GuildID == ""
	isMention := false
	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID {
			isMention = true
			break
		}
	}

	if !isDM && !isMention {
		return
	}

	// Create channel message
	msg := channels.Message{
		ID:        m.ID,
		Content:   m.Content,
		Source:    d.id,
		ChannelID: m.ChannelID,
		SenderID:  m.Author.ID,
		Metadata: map[string]string{
			"username":   m.Author.Username,
			"channel":    m.ChannelID,
			"is_dm":      fmt.Sprintf("%v", isDM),
			"is_mention": fmt.Sprintf("%v", isMention),
		},
		CreatedAt: time.Now(),
	}

	// Publish to event bus
	if d.eventBus != nil {
		d.eventBus.Publish(bus.Event{
			Type:    bus.EventChannelMessage,
			Payload: msg,
		})
	}
}

func (d *DiscordChannel) handleOutbound(ctx context.Context, outbound <-chan bus.Event, unsub func()) {
	defer unsub()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-outbound:
			if msg, ok := event.Payload.(channels.Message); ok {
				if msg.Source == d.id {
					// This is our message to send
					if err := d.Send(ctx, msg); err != nil {
						// Log error but don't crash
						fmt.Printf("Discord send error: %v\n", err)
					}
				}
			}
		}
	}
}
