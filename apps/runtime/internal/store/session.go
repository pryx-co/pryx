package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Store) CreateSession(title string) (*Session, error) {
	id := uuid.New().String()
	now := time.Now().UTC()

	sess := &Session{
		ID:        id,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `INSERT INTO sessions (id, title, created_at, updated_at) VALUES (?, ?, ?, ?)`
	_, err := s.DB.Exec(query, sess.ID, sess.Title, sess.CreatedAt, sess.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func (s *Store) GetSession(id string) (*Session, error) {
	sess := &Session{}
	query := `SELECT id, title, created_at, updated_at FROM sessions WHERE id = ?`
	err := s.DB.QueryRow(query, id).Scan(&sess.ID, &sess.Title, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *Store) ListSessions() ([]*Session, error) {
	query := `SELECT id, title, created_at, updated_at FROM sessions ORDER BY updated_at DESC LIMIT 100` // Cap for now
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		sess := &Session{}
		if err := rows.Scan(&sess.ID, &sess.Title, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

func (s *Store) EnsureSession(id string, title string) (*Session, error) {
	if id == "" {
		return nil, sql.ErrNoRows
	}
	if title == "" {
		title = "Session"
	}

	now := time.Now().UTC()
	_, err := s.DB.Exec(
		`INSERT OR IGNORE INTO sessions (id, title, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		id,
		title,
		now,
		now,
	)
	if err != nil {
		return nil, err
	}
	_, _ = s.DB.Exec(`UPDATE sessions SET updated_at = ? WHERE id = ?`, now, id)

	return s.GetSession(id)
}

func (s *Store) DeleteSession(id string) error {
	if id == "" {
		return sql.ErrNoRows
	}
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM messages WHERE session_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM sessions WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}
