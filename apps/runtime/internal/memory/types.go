package memory

import "time"

// Memory usage types
type MemoryUsage struct {
	UsedTokens   int     `json:"used_tokens"`
	MaxTokens    int     `json:"max_tokens"`
	UsagePercent float64 `json:"usage_percent"`
	WarningLevel string  `json:"warning_level"`
}

// Session memory details
type SessionMemory struct {
	SessionID        string    `json:"session_id"`
	ParentSessionID  string    `json:"parent_session_id,omitempty"`
	Title            string    `json:"title"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	MessagesCount    int       `json:"messages_count"`
	TotalTokens      int       `json:"total_tokens"`
	CompressedTokens int       `json:"compressed_tokens"`
	Archived         bool      `json:"archived"`
}

// Summary request type
type SummaryRequest struct {
	OldestMessages   int     `json:"oldest_messages"`
	CompressionRatio float64 `json:"compression_ratio"`
}

// Compression result type
type CompressionResult struct {
	CompressedCount int `json:"compressed_count"`
	NewTotalTokens  int `json:"new_total_tokens"`
	SavedTokens     int `json:"saved_tokens"`
}

// Archive result type
type ArchiveResult struct {
	ArchivedCount    int      `json:"archived_count"`
	ArchivedSessions []string `json:"archived_sessions"`
}

// Memory constants
const (
	WarnThresholdPercent      = 80
	SummarizeThresholdPercent = 90
	CompressionRatio          = 0.2
	MaxContextTokens          = 128000
	SessionArchiveDays        = 7
)
