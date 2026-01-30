//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

// TestAgentSpawning tests the agent spawning API
func TestAgentSpawning(t *testing.T) {
	bin := buildPryxCore(t)
	home := t.TempDir()

	// Start pryx-core in background
	ctx, cancel := startPryxCore(t, bin, home)
	defer cancel()

	// Wait for server to be ready
	waitForServer(t, 5*time.Second)

	// Test 1: Spawn a new agent
	t.Run("spawn_agent", func(t *testing.T) {
		payload := map[string]interface{}{
			"task":    "Test task for E2E",
			"model":   "gpt-4",
			"context": map[string]string{"test": "data"},
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post("http://localhost:3000/api/v1/agents/spawn", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to spawn agent: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["id"] == nil {
			t.Fatal("Expected agent ID in response")
		}

		t.Logf("✓ Agent spawned with ID: %s", result["id"])
	})

	// Test 2: List agents
	t.Run("list_agents", func(t *testing.T) {
		resp, err := http.Get("http://localhost:3000/api/v1/agents")
		if err != nil {
			t.Fatalf("Failed to list agents: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		agents, ok := result["agents"].([]interface{})
		if !ok {
			t.Fatal("Expected agents array in response")
		}

		if len(agents) == 0 {
			t.Fatal("Expected at least one agent")
		}

		t.Logf("✓ Found %d agents", len(agents))
	})

	// Test 3: Get agent status
	t.Run("get_agent_status", func(t *testing.T) {
		// First spawn an agent
		resp, err := http.Get("http://localhost:3000/api/v1/agents")
		if err != nil {
			t.Fatalf("Failed to list agents: %v", err)
		}

		var listResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&listResult)
		resp.Body.Close()

		agents := listResult["agents"].([]interface{})
		if len(agents) == 0 {
			t.Skip("No agents to check status")
		}

		firstAgent := agents[0].(map[string]interface{})
		agentID := firstAgent["id"].(string)

		resp, err = http.Get("http://localhost:3000/api/v1/agents/" + agentID)
		if err != nil {
			t.Fatalf("Failed to get agent: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}

		var agent map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
			t.Fatalf("Failed to decode agent: %v", err)
		}

		if agent["id"] != agentID {
			t.Fatal("Agent ID mismatch")
		}

		t.Logf("✓ Agent status retrieved: %s", agent["status"])
	})

	// Test 4: Cancel agent
	t.Run("cancel_agent", func(t *testing.T) {
		// Spawn a long-running agent
		payload := map[string]interface{}{
			"task": "Long running task for cancellation test",
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post("http://localhost:3000/api/v1/agents/spawn", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to spawn agent: %v", err)
		}

		var spawnResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&spawnResult)
		resp.Body.Close()

		agentID := spawnResult["id"].(string)

		// Cancel it
		req, _ := http.NewRequest("POST", "http://localhost:3000/api/v1/agents/"+agentID+"/cancel", nil)
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to cancel agent: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}

		t.Logf("✓ Agent cancelled successfully")
	})
}
