package store

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

// Default message limits
const (
	// DefaultMaxMessagesPerSession is the default limit for messages per session.
	DefaultMaxMessagesPerSession = 1000
	// MaxAllowedMessages is the absolute hard limit for messages per session.
	MaxAllowedMessages = 10000
)

// Store provides database access with message limits
type Store struct {
	DB          *sql.DB
	maxMessages int
}

func NewFromDB(db *sql.DB) *Store {
	s := &Store{DB: db}
	s.maxMessages = s.loadMaxMessages()
	return s
}

// New creates a new Store with the given database path
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	// For in-memory databases, use single connection to ensure all operations use the same database
	// For file-based databases, use connection pooling for better performance
	if dbPath == ":memory:" {
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
	} else {
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(25)
		db.SetConnMaxLifetime(5 * 60 * 1000 * 1000000) // 5 minutes
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s := &Store{DB: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Load max messages from environment or use default
	s.maxMessages = s.loadMaxMessages()

	return s, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.DB.Close()
}

// SetMaxMessages sets the maximum messages per session (0 = unlimited)
func (s *Store) SetMaxMessages(max int) {
	if max < 0 {
		max = 0
	}
	if max > MaxAllowedMessages {
		max = MaxAllowedMessages
	}
	s.maxMessages = max
}

// GetMaxMessages returns the current maximum messages setting
func (s *Store) GetMaxMessages() int {
	return s.maxMessages
}

// loadMaxMessages loads max messages from environment or returns default
func (s *Store) loadMaxMessages() int {
	if v := os.Getenv("PRYX_MAX_MESSAGES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			if n > MaxAllowedMessages {
				return MaxAllowedMessages
			}
			return n
		}
	}
	return DefaultMaxMessagesPerSession
}

// CleanupOldMessages removes excess messages for a session
func (s *Store) CleanupOldMessages(sessionID string) error {
	if s.maxMessages <= 0 {
		return nil
	}

	// Count messages
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM messages WHERE session_id = ?", sessionID).Scan(&count)
	if err != nil {
		return err
	}

	if count <= s.maxMessages {
		return nil
	}

	deleteQuery := `
		DELETE FROM messages 
		WHERE session_id = ? 
		AND id IN (
			SELECT id FROM messages 
			WHERE session_id = ? 
			ORDER BY created_at ASC 
			LIMIT ?
		)`

	toDelete := count - s.maxMessages
	_, err = s.DB.Exec(deleteQuery, sessionID, sessionID, toDelete)
	return err
}

// GetMessageCount returns the number of messages in a session
func (s *Store) GetMessageCount(sessionID string) (int, error) {
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM messages WHERE session_id = ?", sessionID).Scan(&count)
	return count, err
}

func (s *Store) migrate() error {
	_, err := s.DB.Exec(schema)
	if err != nil {
		return err
	}

	// Create indexes for better query performance
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON sessions(updated_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at DESC)`,
	}

	for _, idx := range indexes {
		if _, err := s.DB.Exec(idx); err != nil {
			continue
		}
	}

	return nil
}
