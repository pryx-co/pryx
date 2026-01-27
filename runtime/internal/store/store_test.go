package store

import (
	"os"
	"testing"
)

func TestStore(t *testing.T) {
	// Use a temporary DB file
	dbPath := "test_pryx.db"
	defer os.Remove(dbPath)

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	// Test Session Creation
	sess, err := s.CreateSession("Test Session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	if sess.Title != "Test Session" {
		t.Errorf("Expected title 'Test Session', got '%s'", sess.Title)
	}

	// Test Get Session
	fetched, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	if fetched.ID != sess.ID {
		t.Errorf("Expected fetched ID to match created ID")
	}

	// Test Add Message
	msg, err := s.AddMessage(sess.ID, RoleUser, "Hello world")
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}
	if msg.Content != "Hello world" {
		t.Errorf("Expected content 'Hello world', got '%s'", msg.Content)
	}

	// Test Get Messages
	msgs, err := s.GetMessages(sess.ID)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}
	if len(msgs) != 1 {
		t.Errorf("Expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Content != "Hello world" {
		t.Errorf("Expected stored message content to match")
	}

	// Test List Sessions
	sessions, err := s.ListSessions()
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}
	if len(sessions) == 0 {
		t.Errorf("Expected at least one session")
	}
}
