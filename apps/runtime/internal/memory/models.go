package memory

import "time"

// MemoryType represents the type of memory entry
type MemoryType string

const (
	MemoryTypeDaily    MemoryType = "daily"
	MemoryTypeLongterm MemoryType = "longterm"
	MemoryTypeSession  MemoryType = "session"
)

// MemoryEntry represents a single memory entry in the RAG system
type MemoryEntry struct {
	ID           string         `json:"id"`
	Type         MemoryType     `json:"type"`
	Date         string         `json:"date,omitempty"` // YYYY-MM-DD for daily logs
	Content      string         `json:"content"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	AccessCount  int            `json:"access_count"`
	LastAccessed *time.Time     `json:"last_accessed,omitempty"`
	Sources      []MemorySource `json:"sources,omitempty"`
}

// MemorySource represents the source of a memory entry
type MemorySource struct {
	ID         string `json:"id"`
	EntryID    string `json:"entry_id"`
	SourceType string `json:"source_type"` // 'file', 'tool', 'conversation'
	SourcePath string `json:"source_path"`
}

// SearchResult represents a search result with relevance scoring
type SearchResult struct {
	Entry       MemoryEntry `json:"entry"`
	FTSScore    float64     `json:"fts_score"`
	VectorScore float64     `json:"vector_score,omitempty"`
	HybridScore float64     `json:"hybrid_score"`
}

// SearchOptions provides options for memory search
type SearchOptions struct {
	Type          MemoryType // Filter by memory type
	Date          string     // Filter by date (YYYY-MM-DD)
	Limit         int        // Maximum results
	IncludeFTS    bool       // Include full-text search
	IncludeVector bool       // Include vector search (placeholder for future)
}

// FlushOptions provides options for auto-flush behavior
type FlushOptions struct {
	SessionID        string
	TokenCount       int
	MaxTokens        int
	ThresholdPercent float64
}

// WriteRequest represents a request to write to memory
type WriteRequest struct {
	Type    MemoryType
	Content string
	Date    string // For daily logs
	Sources []MemorySource
}

// MemoryStats provides statistics about the memory system
type MemoryStats struct {
	TotalEntries    int64      `json:"total_entries"`
	DailyEntries    int64      `json:"daily_entries"`
	LongtermEntries int64      `json:"longterm_entries"`
	SessionEntries  int64      `json:"session_entries"`
	LastFlushAt     *time.Time `json:"last_flush_at,omitempty"`
}
