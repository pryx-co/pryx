package social

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestMoltbookAdapter tests the Moltbook adapter
func TestMoltbookAdapter(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/agents/me":
			// Auth validation endpoint
			if r.Method == "GET" && r.Header.Get("Authorization") != "" {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":   "test-agent",
					"name": "Test Agent",
				})
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		case "/api/v1/posts":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":         "new-post-1",
					"agent_id":   "test-agent",
					"body":       "Test post content",
					"votes":      0,
					"created_at": time.Now().Format(time.RFC3339),
				})
			}
		case "/api/v1/feed":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"id":    "feed-post-1",
						"type":  "post",
						"body":  "Feed post",
						"votes": 5,
						"author": map[string]interface{}{
							"id":   "author-1",
							"name": "Test Author",
						},
						"created_at": time.Now().Format(time.RFC3339),
					},
				},
			})
		case "/api/v1/health":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create adapter
	adapter := NewMoltbookAdapter(server.URL)
	ctx := context.Background()

	// Test capabilities
	caps := adapter.Capabilities()
	if !caps.SupportsPost {
		t.Error("expected post capability to be supported")
	}
	if !caps.SupportsFeed {
		t.Error("expected feed capability to be supported")
	}

	// Test authenticate
	err := adapter.Authenticate(ctx, "test-token")
	if err != nil {
		t.Fatalf("authenticate failed: %v", err)
	}
	if !adapter.IsAuthenticated() {
		t.Error("expected adapter to be authenticated")
	}

	// Test post
	post, err := adapter.Post(ctx, PostContent{
		Body: "Test post content",
		Tags: []string{"test"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	if post == nil {
		t.Fatal("expected post result")
	}

	// Test get feed
	feed, err := adapter.GetFeed(ctx, FeedRequest{Limit: 10})
	if err != nil {
		t.Fatalf("get feed failed: %v", err)
	}
	if len(feed) == 0 {
		t.Error("expected feed items")
	}

	// Test health check
	err = adapter.HealthCheck(ctx)
	if err != nil {
		t.Errorf("health check failed: %v", err)
	}
}

// TestHub tests the social hub
func TestHub(t *testing.T) {
	hub := NewHub()

	// Test empty hub
	if hub.AdapterCount() != 0 {
		t.Error("expected empty hub")
	}
	if len(hub.ListAdapters()) != 0 {
		t.Error("expected empty adapter list")
	}

	// Create and register test adapter
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/posts":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":    "post-1",
					"body":  "Test",
					"votes": 0,
				})
			}
		case "/api/v1/feed":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"items": []map[string]interface{}{},
			})
		case "/api/v1/health":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	adapter := NewMoltbookAdapter(server.URL)
	err := hub.RegisterAdapter(adapter)
	if err != nil {
		t.Fatalf("register adapter failed: %v", err)
	}

	// Test adapter registered
	if hub.AdapterCount() != 1 {
		t.Errorf("expected 1 adapter, got %d", hub.AdapterCount())
	}
	if !hub.HasAdapter("moltbook") {
		t.Error("expected moltbook adapter to be registered")
	}

	// Test capabilities
	caps, err := hub.Capabilities("moltbook")
	if err != nil {
		t.Fatalf("get capabilities failed: %v", err)
	}
	if !caps.SupportsPost {
		t.Error("expected post capability")
	}

	// Test supports check
	if !hub.Supports("moltbook", "post") {
		t.Error("expected moltbook to support post")
	}
	if hub.Supports("moltbook", "nonexistent") {
		t.Error("did not expect nonexistent action to be supported")
	}

	// Test get adapter
	retrievedAdapter, ok := hub.GetAdapter("moltbook")
	if !ok {
		t.Error("expected to get moltbook adapter")
	}
	if retrievedAdapter.Name() != "moltbook" {
		t.Errorf("expected name 'moltbook', got '%s'", retrievedAdapter.Name())
	}

	// Test unregister
	hub.UnregisterAdapter("moltbook")
	if hub.AdapterCount() != 0 {
		t.Error("expected empty hub after unregister")
	}
}

// TestRegistry tests the adapter registry
func TestRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test empty registry
	if registry.Count() != 0 {
		t.Error("expected empty registry")
	}

	// Register Moltbook adapter directly to the test registry
	err := registry.Register("moltbook", func() SocialAdapter {
		return NewMoltbookAdapter("https://moltbook.com")
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Test registered
	if !registry.HasAdapter("moltbook") {
		t.Error("expected moltbook to be registered")
	}

	// Test create
	adapter, err := registry.Create("moltbook")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if adapter.Name() != "moltbook" {
		t.Errorf("expected name 'moltbook', got '%s'", adapter.Name())
	}

	// Test capabilities
	caps, err := registry.Capabilities("moltbook")
	if err != nil {
		t.Fatalf("get capabilities failed: %v", err)
	}
	if !caps.HasCapability("post") {
		t.Error("expected post capability")
	}

	// Test unregister
	registry.Unregister("moltbook")
	if registry.HasAdapter("moltbook") {
		t.Error("expected moltbook to be unregistered")
	}
}

// TestManifest tests manifest parsing
func TestManifest(t *testing.T) {
	// Test capability checking
	caps := SocialCapabilities{
		SupportsPost:   true,
		SupportsVote:   true,
		SupportsFollow: false,
		Endpoints: map[string]string{
			"post": "/api/v1/posts",
			"vote": "/api/v1/vote",
		},
	}

	if !caps.HasCapability("post") {
		t.Error("expected post capability")
	}
	if !caps.HasCapability("vote") {
		t.Error("expected vote capability")
	}
	if caps.HasCapability("follow") {
		t.Error("did not expect follow capability")
	}

	// Test endpoint lookup
	if caps.GetEndpoint("post") != "/api/v1/posts" {
		t.Errorf("expected '/api/v1/posts', got '%s'", caps.GetEndpoint("post"))
	}
	if caps.GetEndpoint("vote") != "/api/v1/vote" {
		t.Errorf("expected '/api/v1/vote', got '%s'", caps.GetEndpoint("vote"))
	}
	if caps.GetEndpoint("nonexistent") != "" {
		t.Error("expected empty string for nonexistent endpoint")
	}
}

// TestPostContent tests post content
func TestPostContent(t *testing.T) {
	content := PostContent{
		Body:      "Hello world",
		Tags:      []string{"test", "pryx"},
		ParentID:  "parent-123",
		MediaURLs: []string{"https://example.com/image.jpg"},
		Mentions:  []string{"@user1", "@user2"},
	}

	if content.Body != "Hello world" {
		t.Errorf("expected 'Hello world', got '%s'", content.Body)
	}
	if len(content.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(content.Tags))
	}
}

// TestFeedRequest tests feed request
func TestFeedRequest(t *testing.T) {
	limit := 20
	filter := "following"

	request := FeedRequest{
		AgentID: "agent-123",
		Limit:   limit,
		Filter:  filter,
	}

	if request.Limit != limit {
		t.Errorf("expected limit %d, got %d", limit, request.Limit)
	}
	if request.Filter != filter {
		t.Errorf("expected filter '%s', got '%s'", filter, request.Filter)
	}
}
