package memory

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	schema := `
CREATE TABLE IF NOT EXISTS memory_entries (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    date TEXT,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    access_count INTEGER DEFAULT 0,
    last_accessed DATETIME
);

CREATE TABLE IF NOT EXISTS memory_sources (
    id TEXT PRIMARY KEY,
    entry_id TEXT,
    source_type TEXT,
    source_path TEXT,
    FOREIGN KEY (entry_id) REFERENCES memory_entries(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS memory_vectors (
    entry_id TEXT PRIMARY KEY,
    embedding BLOB,
    FOREIGN KEY (entry_id) REFERENCES memory_entries(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_memory_type ON memory_entries(type);
CREATE INDEX IF NOT EXISTS idx_memory_date ON memory_entries(date);
CREATE INDEX IF NOT EXISTS idx_memory_created_at ON memory_entries(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_memory_sources_entry_id ON memory_sources(entry_id);
`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func TestRAGManager_WriteDaily(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mgr := NewRAGManager(db, true)

	sources := []MemorySource{
		{SourceType: "conversation", SourcePath: "session_123"},
	}

	entryID, err := mgr.WriteDaily("Today we implemented the RAG memory system", sources)
	if err != nil {
		t.Fatalf("WriteDaily failed: %v", err)
	}

	if entryID == "" {
		t.Error("WriteDaily returned empty entryID")
	}

	entry, err := mgr.Get(entryID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if entry.Type != MemoryTypeDaily {
		t.Errorf("Expected type 'daily', got '%s'", entry.Type)
	}

	if entry.Content != "Today we implemented the RAG memory system" {
		t.Errorf("Content mismatch: %s", entry.Content)
	}
}

func TestRAGManager_WriteLongterm(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mgr := NewRAGManager(db, true)

	sources := []MemorySource{
		{SourceType: "file", SourcePath: "/docs/architecture.md"},
	}

	entryID, err := mgr.WriteLongterm("Key architecture decision: Use SQLite for local-first design", sources)
	if err != nil {
		t.Fatalf("WriteLongterm failed: %v", err)
	}

	entry, err := mgr.Get(entryID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if entry.Type != MemoryTypeLongterm {
		t.Errorf("Expected type 'longterm', got '%s'", entry.Type)
	}
}

func TestRAGManager_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mgr := NewRAGManager(db, true)

	mgr.WriteDaily("Daily log entry 1", nil)
	mgr.WriteDaily("Daily log entry 2", nil)
	mgr.WriteLongterm("Longterm memory", nil)

	entries, err := mgr.List(SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	dailyEntries, err := mgr.List(SearchOptions{Type: MemoryTypeDaily, Limit: 10})
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}

	if len(dailyEntries) != 2 {
		t.Errorf("Expected 2 daily entries, got %d", len(dailyEntries))
	}
}

func TestRAGManager_Search(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mgr := NewRAGManager(db, true)

	mgr.WriteLongterm("Authentication implementation using OAuth2", nil)
	mgr.WriteLongterm("Database schema design with SQLite", nil)
	mgr.WriteDaily("Working on authentication module", nil)

	results, err := mgr.Search(context.Background(), "authentication", SearchOptions{Limit: 10, IncludeFTS: true})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected search results, got none")
	}

	found := false
	for _, r := range results {
		if r.Entry.Type == MemoryTypeLongterm {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find longterm memory in results")
	}
}

func TestRAGManager_Stats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mgr := NewRAGManager(db, true)

	mgr.WriteDaily("Daily 1", nil)
	mgr.WriteDaily("Daily 2", nil)
	mgr.WriteLongterm("Longterm 1", nil)

	stats, err := mgr.Stats()
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}

	if stats.TotalEntries != 3 {
		t.Errorf("Expected 3 total entries, got %d", stats.TotalEntries)
	}

	if stats.DailyEntries != 2 {
		t.Errorf("Expected 2 daily entries, got %d", stats.DailyEntries)
	}

	if stats.LongtermEntries != 1 {
		t.Errorf("Expected 1 longterm entry, got %d", stats.LongtermEntries)
	}
}

func TestRAGManager_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mgr := NewRAGManager(db, true)

	entryID, _ := mgr.WriteDaily("To be deleted", nil)

	err := mgr.Delete(entryID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = mgr.Get(entryID)
	if err == nil {
		t.Error("Expected error when getting deleted entry, got nil")
	}
}

func TestRAGManager_Disabled(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mgr := NewRAGManager(db, false)

	_, err := mgr.WriteDaily("Test", nil)
	if err == nil {
		t.Error("Expected error when writing to disabled memory, got nil")
	}

	_, err = mgr.List(SearchOptions{})
	if err == nil {
		t.Error("Expected error when listing disabled memory, got nil")
	}
}

func TestAutoFlush_ShouldFlush(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	af := NewAutoFlush(db)

	if !af.ShouldFlush(90000, 100000, 80000) {
		t.Error("Expected ShouldFlush to return true when tokens exceed threshold")
	}

	if af.ShouldFlush(50000, 100000, 80000) {
		t.Error("Expected ShouldFlush to return false when tokens below threshold")
	}
}

func TestAutoFlush_FlushSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	af := NewAutoFlush(db)

	sources := []MemorySource{
		{SourceType: "conversation", SourcePath: "session_456"},
	}

	entryID, err := af.FlushSession("session_456", "Important session summary", sources)
	if err != nil {
		t.Fatalf("FlushSession failed: %v", err)
	}

	if entryID == "" {
		t.Error("FlushSession returned empty entryID")
	}
}
