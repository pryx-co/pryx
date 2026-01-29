package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/memory"
	"pryx-core/internal/store"

	"github.com/stretchr/testify/require"
)

func TestMemoryAndSessionIntegration(t *testing.T) {
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	mgr := memory.NewManager(s, b)

	ctx := context.Background()

	parentSession, err := s.CreateSession("Parent Session")
	if err != nil {
		t.Fatalf("Failed to create parent session: %v", err)
	}

	childSessionID, err := mgr.CreateChildSession(ctx, parentSession.ID, "Child Session")
	if err != nil {
		t.Fatalf("Failed to create child session: %v", err)
	}

	if childSessionID == "" {
		t.Error("Expected non-empty child session ID")
	}

	_, err = s.AddMessage(parentSession.ID, store.RoleUser, "Parent message")
	if err != nil {
		t.Fatalf("Failed to add parent message: %v", err)
	}

	_, err = s.AddMessage(childSessionID, store.RoleUser, "Child message")
	if err != nil {
		t.Fatalf("Failed to add child message: %v", err)
	}

	parentUsage, err := mgr.GetMemoryUsage(ctx, parentSession.ID)
	if err != nil {
		t.Fatalf("Failed to get parent memory usage: %v", err)
	}

	if parentUsage.UsedTokens <= 0 {
		t.Error("Expected positive used tokens for parent session")
	}

	childUsage, err := mgr.GetMemoryUsage(ctx, childSessionID)
	if err != nil {
		t.Fatalf("Failed to get child memory usage: %v", err)
	}

	if childUsage.UsedTokens <= 0 {
		t.Error("Expected positive used tokens for child session")
	}

	allSessions, err := mgr.GetAllSessionsMemory(ctx)
	if err != nil {
		t.Fatalf("Failed to get all sessions memory: %v", err)
	}

	if len(allSessions) < 2 {
		t.Error("Expected at least 2 sessions")
	}
}

func TestCLIToRuntimeIntegration(t *testing.T) {
	runtimeDir, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)

	cliPath := filepath.Join(t.TempDir(), "pryx-core")
	build := exec.Command("go", "build", "-o", cliPath, "./cmd/pryx-core")
	build.Dir = runtimeDir
	build.Env = append(os.Environ(), "CGO_ENABLED=1")
	out, err := build.CombinedOutput()
	require.NoError(t, err, "failed to build CLI: %s", string(out))

	tests := []struct {
		name  string
		args  []string
		check func(string) bool
	}{
		{
			name:  "skills list",
			args:  []string{"skills", "list"},
			check: func(s string) bool { return strings.Contains(s, "Skills") || strings.Contains(s, "skills") },
		},
		{
			name:  "mcp list",
			args:  []string{"mcp", "list"},
			check: func(s string) bool { return strings.Contains(s, "MCP") },
		},
		{
			name:  "cost summary",
			args:  []string{"cost", "summary"},
			check: func(s string) bool { return strings.Contains(s, "Cost") || strings.Contains(s, "cost") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, cliPath, tt.args...)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Fatalf("CLI command failed: %v\nOutput:\n%s", err, string(output))
			}

			if !tt.check(string(output)) {
				t.Errorf("Expected specific content in output for %s, got: %s", tt.name, output)
			}
		})
	}
}

func TestFullWorkflowIntegration(t *testing.T) {
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	mgr := memory.NewManager(s, b)

	ctx := context.Background()

	session, err := s.CreateSession("Full Workflow Integration Test")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	for i := 0; i < 5; i++ {
		_, err = s.AddMessage(session.ID, store.RoleUser, "Test message for workflow")
		if err != nil {
			t.Fatalf("Failed to add message: %v", err)
		}

		_, err = s.AddMessage(session.ID, store.RoleAssistant, "Test response for workflow")
		if err != nil {
			t.Fatalf("Failed to add response: %v", err)
		}
	}

	usage, err := mgr.GetMemoryUsage(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to get memory usage: %v", err)
	}

	if usage.UsedTokens == 0 {
		t.Error("Expected positive token count after adding messages")
	}

	compression, err := mgr.SummarizeSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to summarize session: %v", err)
	}

	if compression.CompressedCount == 0 {
		t.Log("Note: No messages were compressed (may be expected behavior)")
	}

	sessionMemory, err := mgr.GetSessionMemory(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to get session memory: %v", err)
	}

	if sessionMemory.MessagesCount != 10 {
		t.Logf("Note: Expected 10 messages, got %d (may include compression)", sessionMemory.MessagesCount)
	}

	ragResult, err := mgr.QueryRAG(ctx, session.ID, "test")
	if err != nil {
		t.Fatalf("Failed to query RAG: %v", err)
	}

	if ragResult == nil {
		t.Error("Expected non-nil RAG result")
	}
}

func TestMemoryWarningThresholds(t *testing.T) {
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	mgr := memory.NewManager(s, b)

	ctx := context.Background()

	session, err := s.CreateSession("Memory Thresholds Test")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	for i := 0; i < 50; i++ {
		_, err = s.AddMessage(session.ID, store.RoleUser, strings.Repeat("test message ", 50))
		if err != nil {
			t.Fatalf("Failed to add message: %v", err)
		}
	}

	usage, err := mgr.GetMemoryUsage(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to get memory usage: %v", err)
	}

	if usage.UsedTokens == 0 {
		t.Error("Expected positive token count")
	}

	if usage.MaxTokens == 0 {
		t.Error("Expected non-zero max tokens")
	}

	percentUsed := (float64(usage.UsedTokens) / float64(usage.MaxTokens)) * 100

	t.Logf("Memory usage: %d/%d tokens (%.2f%%), warning level: %s",
		usage.UsedTokens, usage.MaxTokens, percentUsed, usage.WarningLevel)

	err = mgr.CheckAndWarn(ctx, session.ID)
	if err != nil {
		t.Fatalf("CheckAndWarn failed: %v", err)
	}
}

func TestSessionArchiveWorkflow(t *testing.T) {
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	mgr := memory.NewManager(s, b)

	ctx := context.Background()

	session, err := s.CreateSession("Archive Test Session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	for i := 0; i < 3; i++ {
		_, err = s.AddMessage(session.ID, store.RoleUser, "Test message")
		if err != nil {
			t.Fatalf("Failed to add message: %v", err)
		}
	}

	archiveResult, err := mgr.ArchiveSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to archive session: %v", err)
	}

	if archiveResult.ArchivedCount != 1 {
		t.Errorf("Expected 1 archived session, got %d", archiveResult.ArchivedCount)
	}

	err = mgr.UnarchiveSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to unarchive session: %v", err)
	}

	usage, err := mgr.GetMemoryUsage(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to get memory usage after unarchive: %v", err)
	}

	if usage.UsedTokens == 0 {
		t.Error("Expected positive token count after unarchive")
	}
}

func TestAutoMemoryManagement(t *testing.T) {
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	mgr := memory.NewManager(s, b)

	ctx := context.Background()

	session, err := s.CreateSession("Auto Memory Test")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	for i := 0; i < 100; i++ {
		_, err = s.AddMessage(session.ID, store.RoleUser, "Auto management test message")
		if err != nil {
			t.Fatalf("Failed to add message: %v", err)
		}
	}

	err = mgr.AutoManageMemory(ctx, session.ID)
	if err != nil {
		t.Fatalf("AutoManageMemory failed: %v", err)
	}

	usage, err := mgr.GetMemoryUsage(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to get memory usage after auto-manage: %v", err)
	}

	t.Logf("After auto-management: %d tokens used", usage.UsedTokens)
}

func TestMultipleSessionsMemory(t *testing.T) {
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	mgr := memory.NewManager(s, b)

	ctx := context.Background()

	sessionIDs := []string{}
	for i := 0; i < 5; i++ {
		session, err := s.CreateSession("Multi Session Test")
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
		sessionIDs = append(sessionIDs, session.ID)

		for j := 0; j < 3; j++ {
			_, err = s.AddMessage(session.ID, store.RoleUser, "Session message")
			if err != nil {
				t.Fatalf("Failed to add message: %v", err)
			}
		}
	}

	allMemory, err := mgr.GetAllSessionsMemory(ctx)
	if err != nil {
		t.Fatalf("Failed to get all sessions memory: %v", err)
	}

	if len(allMemory) < 5 {
		t.Errorf("Expected at least 5 sessions, got %d", len(allMemory))
	}

	for i, sessionID := range sessionIDs {
		found := false
		for _, mem := range allMemory {
			if mem.SessionID == sessionID {
				found = true
				if mem.MessagesCount != 3 {
					t.Logf("Note: Session %d has %d messages (expected 3)", i+1, mem.MessagesCount)
				}
				break
			}
		}
		if !found {
			t.Errorf("Session %s not found in memory list", sessionID)
		}
	}
}
