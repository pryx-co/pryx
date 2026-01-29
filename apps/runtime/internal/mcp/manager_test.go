package mcp

import (
	"context"
	"testing"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/keychain"
	"pryx-core/internal/policy"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	b := bus.New()
	p := policy.NewEngine(nil)
	kc := keychain.New("test")

	mgr := NewManager(b, p, kc)

	assert.NotNil(t, mgr)
	assert.Equal(t, b, mgr.bus)
	assert.Equal(t, p, mgr.policy)
	assert.Equal(t, kc, mgr.keychain)
	assert.NotNil(t, mgr.clients)
	assert.NotNil(t, mgr.cache)
	assert.NotNil(t, mgr.pendingApprovals)
}

func TestNewManager_NilPolicy(t *testing.T) {
	b := bus.New()
	kc := keychain.New("test")

	mgr := NewManager(b, nil, kc)

	assert.NotNil(t, mgr)
	assert.NotNil(t, mgr.policy)
}

func TestManager_ResolveApproval(t *testing.T) {
	b := bus.New()
	mgr := NewManager(b, nil, nil)

	// Setup pending approval
	approvalID := "test-approval-123"
	ch := make(chan bool, 1)
	mgr.pendingApprovals[approvalID] = pendingApproval{
		ch:        ch,
		sessionID: "session-1",
		tool:      "test-tool",
	}

	// Resolve as approved
	result := mgr.ResolveApproval(approvalID, true)
	assert.True(t, result)

	// Verify channel received value
	select {
	case approved := <-ch:
		assert.True(t, approved)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected approval value in channel")
	}

	// Verify removed from pending
	_, exists := mgr.pendingApprovals[approvalID]
	assert.False(t, exists)
}

func TestManager_ResolveApproval_NotFound(t *testing.T) {
	b := bus.New()
	mgr := NewManager(b, nil, nil)

	result := mgr.ResolveApproval("nonexistent", true)
	assert.False(t, result)
}

func TestManager_ResolveApproval_BusEvent(t *testing.T) {
	b := bus.New()
	events, cancel := b.Subscribe()
	defer cancel()

	mgr := NewManager(b, nil, nil)

	approvalID := "approval-456"
	mgr.pendingApprovals[approvalID] = pendingApproval{
		ch:        make(chan bool, 1),
		sessionID: "session-2",
		tool:      "test-tool",
	}

	mgr.ResolveApproval(approvalID, false)

	// Verify event was published
	select {
	case evt := <-events:
		assert.Equal(t, bus.EventApprovalResolved, evt.Event)
		assert.Equal(t, "session-2", evt.SessionID)
		payload := evt.Payload.(map[string]interface{})
		assert.Equal(t, approvalID, payload["approval_id"])
		assert.False(t, payload["approved"].(bool))
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected event to be published")
	}
}

func TestSplitToolName(t *testing.T) {
	tests := []struct {
		name         string
		full         string
		expectedSrv  string
		expectedName string
	}{
		{"colon format", "filesystem:readFile", "filesystem", "readFile"},
		{"slash format", "filesystem/readFile", "filesystem", "readFile"},
		{"empty string", "", "", ""},
		{"whitespace", "  filesystem:readFile  ", "filesystem", "readFile"},
		{"only server", "filesystem:", "filesystem", ""},
		{"no separator", "toolname", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, name := splitToolName(tt.full)
			assert.Equal(t, tt.expectedSrv, srv)
			assert.Equal(t, tt.expectedName, name)
		})
	}
}

func TestManager_ListTools_Empty(t *testing.T) {
	b := bus.New()
	mgr := NewManager(b, nil, nil)

	ctx := context.Background()
	tools, err := mgr.ListTools(ctx, false)

	assert.NoError(t, err)
	assert.Empty(t, tools)
}

func TestManager_ListToolsFlat_Empty(t *testing.T) {
	b := bus.New()
	mgr := NewManager(b, nil, nil)

	ctx := context.Background()
	tools, err := mgr.ListToolsFlat(ctx, false)

	assert.NoError(t, err)
	assert.Empty(t, tools)
}

func TestManager_CacheHit(t *testing.T) {
	b := bus.New()
	mgr := NewManager(b, nil, nil)

	// Pre-populate cache
	mgr.cache["test-server"] = cachedTools{
		fetchedAt: time.Now(),
		tools: []Tool{
			{Name: "tool1"},
			{Name: "tool2"},
		},
	}

	ctx := context.Background()
	tools, err := mgr.listToolsCached(ctx, "test-server", nil, false)

	assert.NoError(t, err)
	assert.Len(t, tools, 2)
	assert.Equal(t, "tool1", tools[0].Name)
}

func TestManager_ApplyAuth_NonOAuth(t *testing.T) {
	b := bus.New()
	mgr := NewManager(b, nil, nil)

	headers := map[string]string{}
	auth := AuthConfig{Type: "apikey"}

	err := mgr.applyAuth(headers, auth)
	assert.NoError(t, err)
	assert.Empty(t, headers["Authorization"])
}

func TestManager_ApplyAuth_OAuth_MissingTokenRef(t *testing.T) {
	b := bus.New()
	mgr := NewManager(b, nil, nil)

	headers := map[string]string{}
	auth := AuthConfig{Type: "oauth"}

	err := mgr.applyAuth(headers, auth)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token_ref")
}

func TestManager_ApplyAuth_OAuth_InvalidTokenRef(t *testing.T) {
	b := bus.New()
	mgr := NewManager(b, nil, nil)

	headers := map[string]string{}
	auth := AuthConfig{
		Type:     "oauth",
		TokenRef: "invalid:ref",
	}

	err := mgr.applyAuth(headers, auth)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported token_ref")
}

func TestManager_ApplyAuth_OAuth_NoKeychain(t *testing.T) {
	b := bus.New()
	mgr := NewManager(b, nil, nil) // No keychain

	headers := map[string]string{}
	auth := AuthConfig{
		Type:     "oauth",
		TokenRef: "keychain:token",
	}

	err := mgr.applyAuth(headers, auth)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "keychain not available")
}

func BenchmarkSplitToolName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = splitToolName("filesystem:readFile")
	}
}

func BenchmarkManager_ResolveApproval(b *testing.B) {
	mgr := NewManager(nil, nil, nil)

	// Pre-populate
	for i := 0; i < 100; i++ {
		mgr.pendingApprovals[string(rune(i))] = pendingApproval{
			ch: make(chan bool, 1),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := string(rune(i % 100))
		mgr.ResolveApproval(id, true)
		// Re-add for next iteration
		mgr.pendingApprovals[id] = pendingApproval{ch: make(chan bool, 1)}
	}
}
