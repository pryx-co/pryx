package channels

import (
	"context"
	"fmt"
	"sync"
	"time"

	"pryx-core/internal/bus"
)

type ChannelManager struct {
	mu       sync.RWMutex
	channels map[string]Channel
	eventBus *bus.Bus

	ctx    context.Context
	cancel func()
}

func NewManager(eventBus *bus.Bus) *ChannelManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ChannelManager{
		channels: make(map[string]Channel),
		eventBus: eventBus,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (m *ChannelManager) Register(c Channel) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.channels[c.ID()]; exists {
		return fmt.Errorf("channel %s already registered", c.ID())
	}

	m.channels[c.ID()] = c

	// Start auto-reconnect loop
	go m.maintainConnection(c)

	return nil
}

func (m *ChannelManager) Get(id string) (Channel, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.channels[id]
	return c, ok
}

func (m *ChannelManager) List() []Channel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]Channel, 0, len(m.channels))
	for _, c := range m.channels {
		list = append(list, c)
	}
	return list
}

func (m *ChannelManager) Shutdown() {
	m.cancel()
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, c := range m.channels {
		c.Disconnect(context.Background())
	}
}

func (m *ChannelManager) maintainConnection(c Channel) {
	// Initial Connect
	if err := c.Connect(m.ctx); err != nil {
		m.publishError(c, err)
	} else {
		m.publishStatus(c, StatusConnected)
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			c.Disconnect(context.Background())
			return
		case <-ticker.C:
			currentStatus := c.Status()
			if currentStatus != StatusConnected && currentStatus != StatusConnecting {
				m.publishStatus(c, StatusConnecting)
				if err := c.Connect(m.ctx); err != nil {
					m.publishError(c, err)
					m.publishStatus(c, StatusError)
				} else {
					m.publishStatus(c, StatusConnected)
				}
			}
		}
	}
}

func (m *ChannelManager) publishError(c Channel, err error) {
	if m.eventBus != nil {
		m.eventBus.Publish(bus.NewEvent(bus.EventErrorOccurred, "", map[string]interface{}{
			"channel_id": c.ID(),
			"error":      err.Error(),
		}))
	}
}

func (m *ChannelManager) publishStatus(c Channel, status Status) {
	if m.eventBus != nil {
		m.eventBus.Publish(bus.NewEvent(bus.EventChannelStatus, "", map[string]interface{}{
			"channel_id": c.ID(),
			"status":     status,
		}))
	}
}
