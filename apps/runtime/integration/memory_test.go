package integration

import (
	"context"
	"testing"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/memory"
	"pryx-core/internal/store"
)

func TestMemoryManagement_Integration(t *testing.T) {
	// Create in-memory database
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	// Create event bus
	b := bus.New()

	// Create memory manager
	mgr := memory.NewManager(s, b)

	// Create a test session
	session, err := s.CreateSession("Integration Test Session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test 1: Get initial memory usage
	ctx := context.Background()
	usage, err := mgr.GetMemoryUsage(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetMemoryUsage failed: %v", err)
	}

	if usage.MaxTokens != memory.MaxContextTokens {
		t.Errorf("Expected MaxTokens %d, got: %d", memory.MaxContextTokens, usage.MaxTokens)
	}

	if usage.UsedTokens != 0 {
		t.Errorf("Expected 0 used tokens initially, got: %d", usage.UsedTokens)
	}

	// Test 2: Add messages and check memory usage
	_, err = s.AddMessage(session.ID, store.RoleUser, "Test message 1")
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	_, err = s.AddMessage(session.ID, store.RoleAssistant, "Test response 1")
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	usage, err = mgr.GetMemoryUsage(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetMemoryUsage after messages failed: %v", err)
	}

	if usage.UsedTokens <= 0 {
		t.Errorf("Expected positive used tokens after messages, got: %d", usage.UsedTokens)
	}

	// Test 3: Check warnings (should be below threshold)
	err = mgr.CheckAndWarn(ctx, session.ID)
	if err != nil {
		t.Errorf("CheckAndWarn failed: %v", err)
	}

	// Test 4: Get session memory details
	sessionMemory, err := mgr.GetSessionMemory(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetSessionMemory failed: %v", err)
	}

	if sessionMemory.SessionID != session.ID {
		t.Errorf("Expected session ID %s, got: %s", session.ID, sessionMemory.SessionID)
	}

	if sessionMemory.Title != "Integration Test Session" {
		t.Errorf("Expected title 'Integration Test Session', got: %s", sessionMemory.Title)
	}

	if sessionMemory.MessagesCount != 2 {
		t.Errorf("Expected 2 messages, got: %d", sessionMemory.MessagesCount)
	}

	// Test 5: Summarize session
	result, err := mgr.SummarizeSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("SummarizeSession failed: %v", err)
	}

	if result.CompressedCount == 0 {
		t.Error("Expected some compression to happen")
	}

	if result.SavedTokens <= 0 {
		t.Error("Expected some tokens to be saved")
	}

	// Test 6: Get all sessions memory
	allMemory, err := mgr.GetAllSessionsMemory(ctx)
	if err != nil {
		t.Fatalf("GetAllSessionsMemory failed: %v", err)
	}

	if len(allMemory) == 0 {
		t.Error("Expected at least one session in memory list")
	}

	// Test 7: RAG query
	ragResult, err := mgr.QueryRAG(ctx, session.ID, "test")
	if err != nil {
		t.Fatalf("QueryRAG failed: %v", err)
	}

	if ragResult["session_id"] != session.ID {
		t.Errorf("Expected session ID in RAG result, got: %v", ragResult["session_id"])
	}

	// Test 8: Auto manage memory
	err = mgr.AutoManageMemory(ctx, session.ID)
	if err != nil {
		t.Errorf("AutoManageMemory failed: %v", err)
	}

	// Test 9: Cleanup old sessions
	archivedCount, err := mgr.CleanupOldSessions(ctx)
	if err != nil {
		t.Errorf("CleanupOldSessions failed: %v", err)
	}

	// Should not error even if no sessions to archive
	_ = archivedCount

	// Test 10: Constants are valid
	if memory.WarnThresholdPercent <= 0 || memory.WarnThresholdPercent > 100 {
		t.Errorf("Invalid WarnThresholdPercent: %d", memory.WarnThresholdPercent)
	}

	if memory.SummarizeThresholdPercent <= 0 || memory.SummarizeThresholdPercent > 100 {
		t.Errorf("Invalid SummarizeThresholdPercent: %d", memory.SummarizeThresholdPercent)
	}

	if memory.CompressionRatio <= 0 || memory.CompressionRatio > 1 {
		t.Errorf("Invalid CompressionRatio: %f", memory.CompressionRatio)
	}

	if memory.MaxContextTokens <= 0 {
		t.Errorf("Invalid MaxContextTokens: %d", memory.MaxContextTokens)
	}

	if memory.SessionArchiveDays <= 0 {
		t.Errorf("Invalid SessionArchiveDays: %d", memory.SessionArchiveDays)
	}
}

func TestMemoryManagement_MultipleSessions(t *testing.T) {
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	mgr := memory.NewManager(s, b)
	ctx := context.Background()

	// Create multiple sessions
	sessions := []string{}
	for i := 0; i < 3; i++ {
		session, err := s.CreateSession("Test Session")
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
		sessions = append(sessions, session.ID)

		// Add messages
		for j := 0; j < 5; j++ {
			_, _ = s.AddMessage(session.ID, store.RoleUser, "Message content")
		}
	}

	// Get all sessions memory
	allMemory, err := mgr.GetAllSessionsMemory(ctx)
	if err != nil {
		t.Fatalf("GetAllSessionsMemory failed: %v", err)
	}

	if len(allMemory) != 3 {
		t.Errorf("Expected 3 sessions, got: %d", len(allMemory))
	}

	// Verify each session has memory data
	sessionMap := make(map[string]bool)
	for _, mem := range allMemory {
		sessionMap[mem.SessionID] = true
	}

	for _, sessionID := range sessions {
		if !sessionMap[sessionID] {
			t.Errorf("Session %s not found in memory list", sessionID)
		}
	}
}

func TestMemoryManagement_ChildSessions(t *testing.T) {
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	mgr := memory.NewManager(s, b)
	ctx := context.Background()

	// Create parent session
	parentSession, err := s.CreateSession("Parent Session")
	if err != nil {
		t.Fatalf("Failed to create parent session: %v", err)
	}

	// Create child session
	childSessionID, err := mgr.CreateChildSession(ctx, parentSession.ID, "Child Session")
	if err != nil {
		t.Fatalf("Failed to create child session: %v", err)
	}

	if childSessionID == "" {
		t.Error("Expected non-empty child session ID")
	}

	childMemory, err := mgr.GetSessionMemory(ctx, childSessionID)
	if err != nil {
		t.Fatalf("GetSessionMemory for child failed: %v", err)
	}

	if childMemory.SessionID != childSessionID {
		t.Errorf("Expected child session ID %s, got: %s", childSessionID, childMemory.SessionID)
	}
}

func TestMemoryManagement_Archive(t *testing.T) {
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	mgr := memory.NewManager(s, b)
	ctx := context.Background()

	// Create session
	session, err := s.CreateSession("Archive Test Session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Archive session
	result, err := mgr.ArchiveSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("ArchiveSession failed: %v", err)
	}

	if result.ArchivedCount != 1 {
		t.Errorf("Expected 1 archived session, got: %d", result.ArchivedCount)
	}

	if len(result.ArchivedSessions) != 1 {
		t.Errorf("Expected 1 archived session ID, got: %d", len(result.ArchivedSessions))
	}

	// Unarchive session
	err = mgr.UnarchiveSession(ctx, session.ID)
	if err != nil {
		t.Errorf("UnarchiveSession failed: %v", err)
	}
}

func TestMemoryManagement_StructValidations(t *testing.T) {
	// Test MemoryUsage struct
	usage := memory.MemoryUsage{
		UsedTokens:   1000,
		MaxTokens:    128000,
		UsagePercent: 0.78,
		WarningLevel: "info",
	}

	if usage.UsedTokens != 1000 {
		t.Errorf("Expected UsedTokens 1000, got: %d", usage.UsedTokens)
	}

	if usage.WarningLevel != "info" {
		t.Errorf("Expected WarningLevel 'info', got: %s", usage.WarningLevel)
	}

	// Test SessionMemory struct
	sessionMem := memory.SessionMemory{
		SessionID:        "test-id",
		Title:            "Test Session",
		MessagesCount:    10,
		TotalTokens:      500,
		CompressedTokens: 100,
		Archived:         false,
	}

	if sessionMem.Archived {
		t.Error("Expected Archived to be false")
	}

	if sessionMem.MessagesCount != 10 {
		t.Errorf("Expected MessagesCount 10, got: %d", sessionMem.MessagesCount)
	}

	// Test CompressionResult struct
	compResult := memory.CompressionResult{
		CompressedCount: 5,
		NewTotalTokens:  1000,
		SavedTokens:     200,
	}

	if compResult.CompressedCount != 5 {
		t.Errorf("Expected CompressedCount 5, got: %d", compResult.CompressedCount)
	}

	if compResult.SavedTokens != 200 {
		t.Errorf("Expected SavedTokens 200, got: %d", compResult.SavedTokens)
	}

	// Test ArchiveResult struct
	archResult := memory.ArchiveResult{
		ArchivedCount:    3,
		ArchivedSessions: []string{"s1", "s2", "s3"},
	}

	if archResult.ArchivedCount != 3 {
		t.Errorf("Expected ArchivedCount 3, got: %d", archResult.ArchivedCount)
	}

	if len(archResult.ArchivedSessions) != 3 {
		t.Errorf("Expected 3 archived sessions, got: %d", len(archResult.ArchivedSessions))
	}
}

func TestMemoryManagement_TokenEstimation(t *testing.T) {
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	mgr := memory.NewManager(s, b)
	ctx := context.Background()

	session, err := s.CreateSession("Token Estimation Test")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	usage1, err := mgr.GetMemoryUsage(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetMemoryUsage failed: %v", err)
	}

	_, err = s.AddMessage(session.ID, store.RoleUser, "Hi")
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	usage2, err := mgr.GetMemoryUsage(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetMemoryUsage after message failed: %v", err)
	}

	if usage2.UsedTokens <= usage1.UsedTokens {
		t.Errorf("Expected token count to increase after adding message, was %d, now %d",
			usage1.UsedTokens, usage2.UsedTokens)
	}

	_, err = s.AddMessage(session.ID, store.RoleUser, "This is a much longer test message that should result in more tokens being estimated")
	if err != nil {
		t.Fatalf("Failed to add long message: %v", err)
	}

	usage3, err := mgr.GetMemoryUsage(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetMemoryUsage after long message failed: %v", err)
	}

	if usage3.UsedTokens <= usage2.UsedTokens {
		t.Errorf("Expected token count to increase after adding long message, was %d, now %d",
			usage2.UsedTokens, usage3.UsedTokens)
	}

	if len("Hi") >= len("This is a much longer test message that should result in more tokens being estimated") {
		t.Error("Longer message should have more tokens")
	}
}

func TestMemoryManagement_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	mgr := memory.NewManager(s, b)
	ctx := context.Background()

	// Create session with many messages
	session, _ := s.CreateSession("Performance Test")
	for i := 0; i < 100; i++ {
		_, _ = s.AddMessage(session.ID, store.RoleUser, "Performance test message")
	}

	// Measure GetMemoryUsage performance
	start := time.Now()
	for i := 0; i < 100; i++ {
		_, _ = mgr.GetMemoryUsage(ctx, session.ID)
	}
	elapsed := time.Since(start)

	if elapsed > 5*time.Second {
		t.Errorf("GetMemoryUsage x100 took %v, expected < 5s", elapsed)
	}

	// Measure SummarizeSession performance
	start = time.Now()
	_, _ = mgr.SummarizeSession(ctx, session.ID)
	elapsed = time.Since(start)

	if elapsed > 2*time.Second {
		t.Errorf("SummarizeSession took %v, expected < 2s", elapsed)
	}
}
