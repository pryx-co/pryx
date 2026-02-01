package store

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *Store) CopySession(sourceSessionID string, newTitle string) (*Session, error) {
	_, err := s.GetSession(sourceSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source session: %w", err)
	}

	newSession, err := s.CreateSession(newTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to create new session: %w", err)
	}

	messages, err := s.GetMessages(sourceSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source messages: %w", err)
	}

	for _, msg := range messages {
		newMsg := &Message{
			ID:        uuid.New().String(),
			SessionID: newSession.ID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: time.Now().UTC(),
		}

		query := `INSERT INTO messages (id, session_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`
		_, err := s.DB.Exec(query, newMsg.ID, newMsg.SessionID, newMsg.Role, newMsg.Content, newMsg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to copy message: %w", err)
		}
	}

	now := time.Now().UTC()
	_, _ = s.DB.Exec(`UPDATE sessions SET updated_at = ? WHERE id = ?`, now, newSession.ID)

	return newSession, nil
}

func (s *Store) GetSessionMessages(sessionID string) ([]*Message, error) {
	query := `SELECT id, session_id, role, content, created_at FROM messages 
		WHERE session_id = ? ORDER BY created_at ASC`

	rows, err := s.DB.Query(query, sessionID)
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
