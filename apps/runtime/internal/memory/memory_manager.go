package memory

import (
	"context"
	"fmt"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/store"
)

type Manager struct {
	store *store.Store
	bus   *bus.Bus
}

func NewManager(store *store.Store, bus *bus.Bus) *Manager {
	return &Manager{
		store: store,
		bus:   bus,
	}
}

func (m *Manager) GetMemoryUsage(ctx context.Context, sessionID string) (MemoryUsage, error) {
	messages, err := m.store.GetMessages(sessionID)
	if err != nil {
		return MemoryUsage{}, err
	}

	// Estimate tokens from message content
	totalTokens := 0
	for _, msg := range messages {
		totalTokens += estimateTokens(msg.Content)
	}

	// Calculate token usage
	usagePercent := 0.0
	if MaxContextTokens > 0 {
		usagePercent = float64(totalTokens) / float64(MaxContextTokens) * 100.0
	}

	// Determine warning level
	warningLevel := ""
	if usagePercent >= 100 {
		warningLevel = "critical"
	} else if usagePercent >= float64(SummarizeThresholdPercent) {
		warningLevel = "warn"
	} else if usagePercent >= float64(WarnThresholdPercent) {
		warningLevel = "info"
	}

	return MemoryUsage{
		UsedTokens:   totalTokens,
		MaxTokens:    MaxContextTokens,
		UsagePercent: usagePercent,
		WarningLevel: warningLevel,
	}, nil
}

func (m *Manager) CheckAndWarn(ctx context.Context, sessionID string) error {
	usage, err := m.GetMemoryUsage(ctx, sessionID)
	if err != nil {
		return err
	}

	// Publish warning if needed
	if usage.UsagePercent >= float64(WarnThresholdPercent) {
		if m.bus != nil {
			event := bus.NewEvent("memory.warning", sessionID, map[string]interface{}{
				"usage_percent": usage.UsagePercent,
				"used_tokens":   usage.UsedTokens,
				"max_tokens":    usage.MaxTokens,
				"warning_level": usage.WarningLevel,
			})
			m.bus.Publish(event)
		}
	}

	if usage.UsagePercent >= float64(SummarizeThresholdPercent) {
		if m.bus != nil {
			event := bus.NewEvent("memory.summarize_request", sessionID, map[string]interface{}{
				"oldest_messages":   usage.UsedTokens / 2,
				"compression_ratio": CompressionRatio,
			})
			m.bus.Publish(event)
		}
	}

	return nil
}

func (m *Manager) SummarizeSession(ctx context.Context, sessionID string) (*CompressionResult, error) {
	messages, err := m.store.GetMessages(sessionID)
	if err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return &CompressionResult{
			CompressedCount: 0,
			NewTotalTokens:  0,
			SavedTokens:     0,
		}, nil
	}

	// Compress oldest messages (20% of messages)
	compressCount := int(float64(len(messages)) * CompressionRatio)
	if compressCount == 0 {
		compressCount = 1
	}

	// Create summary of compressed messages
	totalTokens := 0
	for i, msg := range messages {
		if i < compressCount {
			totalTokens += estimateTokens(msg.Content)
		}
	}

	summary := fmt.Sprintf("Compressed %d messages (%d tokens)", compressCount, totalTokens)

	// Note: In a real implementation, you would archive the oldest messages
	// and replace them with a summary message. For now, we just return the result.

	if m.bus != nil {
		event := bus.NewEvent("memory.summarized", sessionID, map[string]interface{}{
			"compressed_count": compressCount,
			"saved_tokens":     totalTokens,
			"summary":          summary,
		})
		m.bus.Publish(event)
	}

	return &CompressionResult{
		CompressedCount: compressCount,
		NewTotalTokens:  len(messages) - compressCount,
		SavedTokens:     totalTokens,
	}, nil
}

func (m *Manager) ArchiveSession(ctx context.Context, sessionID string) (*ArchiveResult, error) {
	_, err := m.store.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	archivedSessions := []string{sessionID}

	// Note: In a real implementation, you would mark the session as archived
	// and move it to an archive table or mark it in the database

	if m.bus != nil {
		event := bus.NewEvent("session.archived", sessionID, map[string]interface{}{
			"archived_count":    1,
			"archived_sessions": archivedSessions,
		})
		m.bus.Publish(event)
	}

	return &ArchiveResult{
		ArchivedCount:    1,
		ArchivedSessions: archivedSessions,
	}, nil
}

func (m *Manager) UnarchiveSession(ctx context.Context, sessionID string) error {
	// Note: In a real implementation, you would unarchive the session
	return nil
}

func (m *Manager) CreateChildSession(ctx context.Context, parentSessionID, title string) (string, error) {
	session, err := m.store.CreateSession(title)
	if err != nil {
		return "", err
	}

	if m.bus != nil {
		event := bus.NewEvent("session.created", session.ID, map[string]interface{}{
			"parent_session_id": parentSessionID,
			"title":             title,
		})
		m.bus.Publish(event)
	}

	return session.ID, nil
}

func (m *Manager) GetSessionMemory(ctx context.Context, sessionID string) (SessionMemory, error) {
	session, err := m.store.GetSession(sessionID)
	if err != nil {
		return SessionMemory{}, err
	}

	messages, err := m.store.GetMessages(sessionID)
	if err != nil {
		return SessionMemory{SessionID: sessionID}, nil
	}

	totalTokens := 0
	for _, msg := range messages {
		totalTokens += estimateTokens(msg.Content)
	}

	return SessionMemory{
		SessionID:        session.ID,
		Title:            session.Title,
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
		MessagesCount:    len(messages),
		TotalTokens:      totalTokens,
		CompressedTokens: 0,
		Archived:         false,
	}, nil
}

func (m *Manager) GetAllSessionsMemory(ctx context.Context) ([]SessionMemory, error) {
	sessions, err := m.store.ListSessions()
	if err != nil {
		return nil, err
	}

	var sessionMemories []SessionMemory
	for _, session := range sessions {
		messages, err := m.store.GetMessages(session.ID)
		if err != nil {
			continue
		}

		totalTokens := 0
		for _, msg := range messages {
			totalTokens += estimateTokens(msg.Content)
		}

		sessionMemories = append(sessionMemories, SessionMemory{
			SessionID:        session.ID,
			Title:            session.Title,
			CreatedAt:        session.CreatedAt,
			UpdatedAt:        session.UpdatedAt,
			MessagesCount:    len(messages),
			TotalTokens:      totalTokens,
			CompressedTokens: 0,
			Archived:         false,
		})
	}

	return sessionMemories, nil
}

func (m *Manager) AutoManageMemory(ctx context.Context, sessionID string) error {
	usage, err := m.GetMemoryUsage(ctx, sessionID)
	if err != nil {
		return err
	}

	// Auto-summarize at threshold
	if usage.UsagePercent >= float64(SummarizeThresholdPercent) {
		_, err := m.SummarizeSession(ctx, sessionID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) CleanupOldSessions(ctx context.Context) (int, error) {
	sessions, err := m.store.ListSessions()
	if err != nil {
		return 0, err
	}

	archiveThreshold := time.Now().AddDate(0, 0, -SessionArchiveDays)
	archivedCount := 0

	for _, session := range sessions {
		// Archive sessions older than threshold
		if session.UpdatedAt.Before(archiveThreshold) {
			_, err := m.ArchiveSession(ctx, session.ID)
			if err == nil {
				archivedCount++
			}
		}
	}

	// Publish cleanup event
	if m.bus != nil {
		event := bus.NewEvent("sessions.cleaned", "", map[string]interface{}{
			"archived_count": archivedCount,
		})
		m.bus.Publish(event)
	}

	return archivedCount, nil
}

func (m *Manager) QueryRAG(ctx context.Context, sessionID, query string) (map[string]interface{}, error) {
	// Placeholder for RAG integration
	// TODO: Implement actual RAG backend integration

	messages, err := m.store.GetMessages(sessionID)
	if err != nil {
		return nil, err
	}

	// Simple keyword-based search for now
	var relevantMessages []string
	for _, msg := range messages {
		if len(msg.Content) > 0 {
			relevantMessages = append(relevantMessages, msg.Content)
		}
	}

	response := map[string]interface{}{
		"query":      query,
		"session_id": sessionID,
		"results": []map[string]interface{}{
			{
				"source":  "context_memory",
				"content": fmt.Sprintf("Found %d relevant messages for query", len(relevantMessages)),
			},
		},
		"message_count": len(relevantMessages),
	}

	return response, nil
}

func estimateTokens(text string) int {
	// Simple estimation: ~4 chars per token
	if len(text) == 0 {
		return 0
	}
	return (len(text) + 3) / 4
}
