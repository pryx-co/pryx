package memory

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RAGManager manages the RAG memory system
type RAGManager struct {
	db      *sql.DB
	enabled bool
	fts     *FTSSearch
	flush   *AutoFlush
}

// NewRAGManager creates a new RAG memory manager
func NewRAGManager(db *sql.DB, enabled bool) *RAGManager {
	m := &RAGManager{
		db:      db,
		enabled: enabled,
	}
	if enabled {
		m.fts = NewFTSSearch(db)
		m.flush = NewAutoFlush(db)
	}
	return m
}

// WriteDaily writes to the daily log (append-only)
func (m *RAGManager) WriteDaily(content string, sources []MemorySource) (string, error) {
	if !m.enabled {
		return "", fmt.Errorf("memory system is disabled")
	}

	date := time.Now().Format("2006-01-02")
	return m.writeEntry(MemoryTypeDaily, content, date, sources)
}

// WriteLongterm writes to long-term memory (curated knowledge)
func (m *RAGManager) WriteLongterm(content string, sources []MemorySource) (string, error) {
	if !m.enabled {
		return "", fmt.Errorf("memory system is disabled")
	}

	return m.writeEntry(MemoryTypeLongterm, content, "", sources)
}

// writeEntry creates a new memory entry
func (m *RAGManager) writeEntry(entryType MemoryType, content string, date string, sources []MemorySource) (string, error) {
	entryID := uuid.New().String()

	tx, err := m.db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"INSERT INTO memory_entries (id, type, date, content) VALUES (?, ?, ?, ?)",
		entryID, entryType, date, content,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert memory entry: %w", err)
	}

	for _, source := range sources {
		sourceID := uuid.New().String()
		_, _ = tx.Exec(
			"INSERT INTO memory_sources (id, entry_id, source_type, source_path) VALUES (?, ?, ?, ?)",
			sourceID, entryID, source.SourceType, source.SourcePath,
		)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return entryID, nil
}

// List returns memory entries with optional filtering
func (m *RAGManager) List(opts SearchOptions) ([]MemoryEntry, error) {
	if !m.enabled {
		return nil, fmt.Errorf("memory system is disabled")
	}

	query := "SELECT id, type, date, content, created_at, updated_at, access_count, last_accessed FROM memory_entries WHERE 1=1"
	var args []interface{}

	if opts.Type != "" {
		query += " AND type = ?"
		args = append(args, opts.Type)
	}

	if opts.Date != "" {
		query += " AND date = ?"
		args = append(args, opts.Date)
	}

	query += " ORDER BY created_at DESC"

	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}
	defer rows.Close()

	var entries []MemoryEntry
	for rows.Next() {
		var entry MemoryEntry
		var lastAccessed sql.NullTime
		err := rows.Scan(
			&entry.ID,
			&entry.Type,
			&entry.Date,
			&entry.Content,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&entry.AccessCount,
			&lastAccessed,
		)
		if err != nil {
			continue
		}
		if lastAccessed.Valid {
			entry.LastAccessed = &lastAccessed.Time
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// Get retrieves a single memory entry by ID
func (m *RAGManager) Get(entryID string) (*MemoryEntry, error) {
	if !m.enabled {
		return nil, fmt.Errorf("memory system is disabled")
	}

	var entry MemoryEntry
	var lastAccessed sql.NullTime

	err := m.db.QueryRow(
		"SELECT id, type, date, content, created_at, updated_at, access_count, last_accessed FROM memory_entries WHERE id = ?",
		entryID,
	).Scan(
		&entry.ID,
		&entry.Type,
		&entry.Date,
		&entry.Content,
		&entry.CreatedAt,
		&entry.UpdatedAt,
		&entry.AccessCount,
		&lastAccessed,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("memory entry not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get memory entry: %w", err)
	}

	if lastAccessed.Valid {
		entry.LastAccessed = &lastAccessed.Time
	}

	entry.Sources, _ = m.getSources(entryID)

	_, _ = m.db.Exec(
		"UPDATE memory_entries SET access_count = access_count + 1, last_accessed = CURRENT_TIMESTAMP WHERE id = ?",
		entryID,
	)

	return &entry, nil
}

// Delete removes a memory entry
func (m *RAGManager) Delete(entryID string) error {
	if !m.enabled {
		return fmt.Errorf("memory system is disabled")
	}

	_, err := m.db.Exec("DELETE FROM memory_entries WHERE id = ?", entryID)
	return err
}

// Stats returns statistics about the memory system
func (m *RAGManager) Stats() (*MemoryStats, error) {
	if !m.enabled {
		return nil, fmt.Errorf("memory system is disabled")
	}

	stats := &MemoryStats{}

	_ = m.db.QueryRow("SELECT COUNT(*) FROM memory_entries").Scan(&stats.TotalEntries)
	_ = m.db.QueryRow("SELECT COUNT(*) FROM memory_entries WHERE type = 'daily'").Scan(&stats.DailyEntries)
	_ = m.db.QueryRow("SELECT COUNT(*) FROM memory_entries WHERE type = 'longterm'").Scan(&stats.LongtermEntries)
	_ = m.db.QueryRow("SELECT COUNT(*) FROM memory_entries WHERE type = 'session'").Scan(&stats.SessionEntries)

	return stats, nil
}

// Enabled returns whether the memory system is enabled
func (m *RAGManager) Enabled() bool {
	return m.enabled
}

// AutoFlush returns the auto-flush manager
func (m *RAGManager) AutoFlush() *AutoFlush {
	return m.flush
}
