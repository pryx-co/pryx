//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

// TestSessionManagement tests session lifecycle operations
func TestSessionManagement(t *testing.T) {
	bin := buildPryxCore(t)
	home := t.TempDir()

	// Start pryx-core
	ctx, cancel := startPryxCore(t, bin, home)
	defer cancel()
	waitForServer(t, 5*time.Second)

	// Test 1: Create a new session
	t.Run("create_session", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "E2E Test Session",
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post("http://localhost:3000/api/v1/sessions", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected 201, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["id"] == nil {
			t.Fatal("Expected session ID in response")
		}

		t.Logf("✓ Session created: %s", result["id"])
	})

	// Test 2: List sessions
	t.Run("list_sessions", func(t *testing.T) {
		resp, err := http.Get("http://localhost:3000/api/v1/sessions")
		if err != nil {
			t.Fatalf("Failed to list sessions: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		sessions, ok := result["sessions"].([]interface{})
		if !ok {
			t.Fatal("Expected sessions array")
		}

		t.Logf("✓ Found %d sessions", len(sessions))
	})

	// Test 3: Fork session
	t.Run("fork_session", func(t *testing.T) {
		// First create a session
		payload := map[string]interface{}{"name": "Parent Session"}
		body, _ := json.Marshal(payload)
		resp, err := http.Post("http://localhost:3000/api/v1/sessions", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to create parent session: %v", err)
		}

		var parent map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&parent)
		resp.Body.Close()

		parentID := parent["id"].(string)

		// Fork it
		forkPayload := map[string]interface{}{
			"source_id": parentID,
		}
		body, _ = json.Marshal(forkPayload)
		resp, err = http.Post("http://localhost:3000/api/v1/sessions/fork", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to fork session: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected 201, got %d", resp.StatusCode)
		}

		var forked map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&forked); err != nil {
			t.Fatalf("Failed to decode fork response: %v", err)
		}

		if forked["id"] == nil {
			t.Fatal("Expected forked session ID")
		}

		t.Logf("✓ Session forked: %s -> %s", parentID, forked["id"])
	})

	// Test 4: Delete session
	t.Run("delete_session", func(t *testing.T) {
		// Create a session
		payload := map[string]interface{}{"name": "Session to Delete"}
		body, _ := json.Marshal(payload)
		resp, err := http.Post("http://localhost:3000/api/v1/sessions", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		var session map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&session)
		resp.Body.Close()

		sessionID := session["id"].(string)

		// Delete it
		req, _ := http.NewRequest("DELETE", "http://localhost:3000/api/v1/sessions/"+sessionID, nil)
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to delete session: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 204 or 200, got %d", resp.StatusCode)
		}

		t.Logf("✓ Session deleted: %s", sessionID)
	})
}
