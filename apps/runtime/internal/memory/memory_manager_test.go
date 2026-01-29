package memory

import (
	"context"
	"testing"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/store"
)

func TestGetMemoryUsage_Basic(t *testing.T) {
	ctx := context.Background()

	// Create in-memory SQLite database for testing
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	manager := NewManager(s, b)

	// Create a test session
	session, err := s.CreateSession("Test Session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add some test messages
	_, _ = s.AddMessage(session.ID, store.RoleUser, "This is a test message with some content")
	_, _ = s.AddMessage(session.ID, store.RoleAssistant, "This is another test message with more content")

	// Test memory usage calculation
	usage, err := manager.GetMemoryUsage(ctx, session.ID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if usage.MaxTokens != MaxContextTokens {
		t.Errorf("Expected MaxTokens %d, got: %d", MaxContextTokens, usage.MaxTokens)
	}

	if usage.UsedTokens <= 0 {
		t.Errorf("Expected positive used tokens, got: %d", usage.UsedTokens)
	}

	if usage.WarningLevel != "" {
		// Should not have warning for small amount of content
		t.Errorf("Expected no warning level for small content, got: %s", usage.WarningLevel)
	}
}

func TestGetMemoryUsage_HighUsage(t *testing.T) {
	ctx := context.Background()

	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	manager := NewManager(s, b)

	session, err := s.CreateSession("High Usage Session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add many messages to simulate high token usage
	for i := 0; i < 100; i++ {
		// Create a long message to increase token count
		longContent := ""
		for j := 0; j < 100; j++ {
			longContent += "This is a test message with lots of content to simulate token usage. "
		}
		_, _ = s.AddMessage(session.ID, store.RoleUser, longContent)
	}

	usage, err := manager.GetMemoryUsage(ctx, session.ID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// With many long messages, we should see high token usage
	if usage.UsedTokens <= 100 {
		t.Errorf("Expected high token usage, got: %d", usage.UsedTokens)
	}
}

func TestSummarizeSession_Basic(t *testing.T) {
	ctx := context.Background()

	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	manager := NewManager(s, b)

	session, err := s.CreateSession("Summarize Test Session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add some test messages
	for i := 0; i < 10; i++ {
		_, _ = s.AddMessage(session.ID, store.RoleUser, "Test message")
	}

	result, err := manager.SummarizeSession(ctx, session.ID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.CompressedCount == 0 {
		t.Error("Expected some compression to happen")
	}

	if result.SavedTokens == 0 {
		t.Error("Expected some tokens to be saved")
	}
}

func TestGetSessionMemory(t *testing.T) {
	ctx := context.Background()

	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	manager := NewManager(s, b)

	session, err := s.CreateSession("Memory Test Session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add test messages
	for i := 0; i < 5; i++ {
		_, _ = s.AddMessage(session.ID, store.RoleUser, "Test message content")
	}

	memory, err := manager.GetSessionMemory(ctx, session.ID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if memory.SessionID != session.ID {
		t.Errorf("Expected session ID %s, got: %s", session.ID, memory.SessionID)
	}

	if memory.Title != "Memory Test Session" {
		t.Errorf("Expected title 'Memory Test Session', got: %s", memory.Title)
	}

	if memory.MessagesCount != 5 {
		t.Errorf("Expected 5 messages, got: %d", memory.MessagesCount)
	}

	if memory.TotalTokens <= 0 {
		t.Errorf("Expected positive total tokens, got: %d", memory.TotalTokens)
	}
}

func TestAutoManageMemory(t *testing.T) {
	ctx := context.Background()

	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	manager := NewManager(s, b)

	session, err := s.CreateSession("Auto Manage Test")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add some messages
	for i := 0; i < 5; i++ {
		_, _ = s.AddMessage(session.ID, store.RoleUser, "Test message")
	}

	// This should not error even with low usage
	err = manager.AutoManageMemory(ctx, session.ID)
	if err != nil {
		t.Errorf("Expected no error from auto manage, got: %v", err)
	}
}

func TestCleanupOldSessions(t *testing.T) {
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	manager := NewManager(s, b)

	// Create some test sessions
	for i := 0; i < 3; i++ {
		_, _ = s.CreateSession("Old Session")
	}

	// Run cleanup
	archivedCount, err := manager.CleanupOldSessions(context.Background())
	if err != nil {
		t.Errorf("Expected no error from cleanup, got: %v", err)
	}

	if archivedCount < 0 {
		t.Errorf("Expected non-negative archived count, got: %d", archivedCount)
	}
}

func TestQueryRAG(t *testing.T) {
	ctx := context.Background()

	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	manager := NewManager(s, b)

	session, err := s.CreateSession("RAG Test Session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add some test messages
	_, _ = s.AddMessage(session.ID, store.RoleUser, "Hello world")
	_, _ = s.AddMessage(session.ID, store.RoleAssistant, "Hello! How can I help you?")

	result, err := manager.QueryRAG(ctx, session.ID, "hello")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result")
	}

	if result["session_id"] != session.ID {
		t.Errorf("Expected session ID %s in result, got: %v", session.ID, result["session_id"])
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"a", 1},
		{"hello world", 3},
		{"This is a test sentence with multiple words", 11}, // 43 chars / 4 = 10.75 -> 11
	}

	for _, test := range tests {
		result := estimateTokens(test.input)
		if result != test.expected {
			t.Errorf("estimateTokens(%q) = %d, expected %d", test.input, result, test.expected)
		}
	}
}

func TestConstants(t *testing.T) {
	// Test that constants are defined correctly
	if WarnThresholdPercent <= 0 || WarnThresholdPercent > 100 {
		t.Errorf("WarnThresholdPercent should be between 0 and 100, got: %d", WarnThresholdPercent)
	}

	if SummarizeThresholdPercent <= 0 || SummarizeThresholdPercent > 100 {
		t.Errorf("SummarizeThresholdPercent should be between 0 and 100, got: %d", SummarizeThresholdPercent)
	}

	if CompressionRatio <= 0 || CompressionRatio > 1 {
		t.Errorf("CompressionRatio should be between 0 and 1, got: %f", CompressionRatio)
	}

	if MaxContextTokens <= 0 {
		t.Errorf("MaxContextTokens should be positive, got: %d", MaxContextTokens)
	}

	if SessionArchiveDays <= 0 {
		t.Errorf("SessionArchiveDays should be positive, got: %d", SessionArchiveDays)
	}
}

func TestMemoryUsageStruct(t *testing.T) {
	usage := MemoryUsage{
		UsedTokens:   1000,
		MaxTokens:    128000,
		UsagePercent: 0.78,
		WarningLevel: "info",
	}

	if usage.UsedTokens != 1000 {
		t.Errorf("Expected UsedTokens 1000, got: %d", usage.UsedTokens)
	}

	if usage.MaxTokens != 128000 {
		t.Errorf("Expected MaxTokens 128000, got: %d", usage.MaxTokens)
	}

	if usage.UsagePercent != 0.78 {
		t.Errorf("Expected UsagePercent 0.78, got: %f", usage.UsagePercent)
	}

	if usage.WarningLevel != "info" {
		t.Errorf("Expected WarningLevel 'info', got: %s", usage.WarningLevel)
	}
}

func TestSessionMemoryStruct(t *testing.T) {
	now := time.Now()
	memory := SessionMemory{
		SessionID:        "test-id",
		ParentSessionID:  "parent-id",
		Title:            "Test Session",
		CreatedAt:        now,
		UpdatedAt:        now,
		MessagesCount:    10,
		TotalTokens:      500,
		CompressedTokens: 100,
		Archived:         false,
	}

	if memory.SessionID != "test-id" {
		t.Errorf("Expected SessionID 'test-id', got: %s", memory.SessionID)
	}

	if memory.MessagesCount != 10 {
		t.Errorf("Expected MessagesCount 10, got: %d", memory.MessagesCount)
	}

	if memory.Archived {
		t.Error("Expected Archived to be false")
	}
}

func TestCompressionResultStruct(t *testing.T) {
	result := CompressionResult{
		CompressedCount: 5,
		NewTotalTokens:  1000,
		SavedTokens:     200,
	}

	if result.CompressedCount != 5 {
		t.Errorf("Expected CompressedCount 5, got: %d", result.CompressedCount)
	}

	if result.SavedTokens != 200 {
		t.Errorf("Expected SavedTokens 200, got: %d", result.SavedTokens)
	}
}

func TestArchiveResultStruct(t *testing.T) {
	result := ArchiveResult{
		ArchivedCount:    3,
		ArchivedSessions: []string{"s1", "s2", "s3"},
	}

	if result.ArchivedCount != 3 {
		t.Errorf("Expected ArchivedCount 3, got: %d", result.ArchivedCount)
	}

	if len(result.ArchivedSessions) != 3 {
		t.Errorf("Expected 3 archived sessions, got: %d", len(result.ArchivedSessions))
	}
}
