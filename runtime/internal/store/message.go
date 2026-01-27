package store

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Store) AddMessage(sessionID string, role Role, content string) (*Message, error) {
	id := uuid.New().String()
	now := time.Now().UTC()

	msg := &Message{
		ID:        id,
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: now,
	}

	query := `INSERT INTO messages (id, session_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, msg.ID, msg.SessionID, msg.Role, msg.Content, msg.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Update session timestamp
	_, _ = s.db.Exec(`UPDATE sessions SET updated_at = ? WHERE id = ?`, now, sessionID)

	return msg, nil
}

func (s *Store) GetMessages(sessionID string) ([]*Message, error) {
	query := `SELECT id, session_id, role, content, created_at FROM messages WHERE session_id = ? ORDER BY created_at ASC`
	rows, err := s.db.Query(query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
