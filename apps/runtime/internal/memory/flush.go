package memory

import (
	"database/sql"
	"fmt"
	"time"
)

// AutoFlush manages automatic memory flushing before context compaction
type AutoFlush struct {
	db *sql.DB
}

// NewAutoFlush creates a new auto-flush manager
func NewAutoFlush(db *sql.DB) *AutoFlush {
	return &AutoFlush{db: db}
}

// ShouldFlush checks if memory should be flushed based on token count
func (af *AutoFlush) ShouldFlush(tokenCount, maxTokens, thresholdTokens int) bool {
	if thresholdTokens <= 0 {
		thresholdTokens = int(float64(maxTokens) * 0.8)
	}
	return tokenCount >= thresholdTokens
}

// FlushSession flushes session memory to long-term storage
func (af *AutoFlush) FlushSession(sessionID string, content string, sources []MemorySource) (string, error) {
	entryID := generateID()

	tx, err := af.db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"INSERT INTO memory_entries (id, type, content) VALUES (?, ?, ?)",
		entryID, MemoryTypeSession, content,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert session memory: %w", err)
	}

	for _, source := range sources {
		sourceID := generateID()
		_, _ = tx.Exec(
			"INSERT INTO memory_sources (id, entry_id, source_type, source_path) VALUES (?, ?, ?, ?)",
			sourceID, entryID, source.SourceType, source.SourcePath,
		)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}

	return entryID, nil
}

// GetFlushReminder returns a reminder message for the agent to store durable memories
func (af *AutoFlush) GetFlushReminder() string {
	return `Memory Auto-Flush Triggered: Context window approaching limit.
Consider what should be preserved for long-term memory:
- Key learnings or insights from this session
- Important decisions or conclusions
- Context that may be needed in future sessions

Use the memory system to store durable knowledge before context compaction occurs.`
}

// GetMemoryContextForAgent retrieves relevant memory for agent context
func (af *AutoFlush) GetMemoryContextForAgent(query string, limit int) (string, error) {
	fts := NewFTSSearch(af.db)

	opts := SearchOptions{
		Limit:      limit,
		IncludeFTS: true,
	}

	results, err := fts.SearchWithFilter(query, opts)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", nil
	}

	context := "=== RELEVANT MEMORY ===\n"
	for i, r := range results {
		context += fmt.Sprintf("\n[%d] %s (%s):\n%s\n", i+1, r.Entry.Type, r.Entry.CreatedAt.Format("2006-01-02"), r.Entry.Content)
	}

	return context, nil
}

// generateID creates a unique ID
func generateID() string {
	return fmt.Sprintf("mem_%d", time.Now().UnixNano())
}
