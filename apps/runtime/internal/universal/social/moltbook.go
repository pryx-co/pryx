package social

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// MoltbookAdapter implements SocialAdapter for the Moltbook social network
type MoltbookAdapter struct {
	baseURL    string
	token      string
	httpClient *http.Client
	manifest   *NetworkManifest
	mu         sync.RWMutex

	// Cached data
	cachedFeed []FeedItem
	cacheTime  time.Time
	cacheTTL   time.Duration
}

// NewMoltbookAdapter creates a new Moltbook adapter
func NewMoltbookAdapter(baseURL string) *MoltbookAdapter {
	return &MoltbookAdapter{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cacheTTL: 5 * time.Minute,
		manifest: &NetworkManifest{
			Name:        "moltbook",
			Version:     "1.0.0",
			DisplayName: "Moltbook",
			Description: "Social network for AI agents",
			BaseURL:     baseURL,
			APIVersion:  "v1",
			SocialFeatures: SocialCapabilities{
				SupportsPost:   true,
				SupportsVote:   true,
				SupportsFollow: true,
				SupportsFeed:   true,
				SupportsReply:  true,
				SupportsRepost: true,
				Endpoints: map[string]string{
					"post":          "/api/v1/posts",
					"vote":          "/api/v1/posts/{post_id}/vote",
					"follow":        "/api/v1/agents/{agent_id}/follow",
					"unfollow":      "/api/v1/agents/{agent_id}/unfollow",
					"feed":          "/api/v1/feed",
					"notifications": "/api/v1/notifications",
					"profile":       "/api/v1/agents/{agent_id}",
					"timeline":      "/api/v1/timeline",
				},
			},
			Auth: AuthConfig{
				Type:     "bearer",
				Required: true,
			},
		},
	}
}

// Name returns the adapter name
func (m *MoltbookAdapter) Name() string {
	return "moltbook"
}

// Version returns the adapter version
func (m *MoltbookAdapter) Version() string {
	return "1.0.0"
}

// DisplayName returns the human-readable name
func (m *MoltbookAdapter) DisplayName() string {
	return "Moltbook"
}

// Capabilities returns the social capabilities
func (m *MoltbookAdapter) Capabilities() SocialCapabilities {
	return m.manifest.SocialFeatures
}

// GetManifest returns the network manifest
func (m *MoltbookAdapter) GetManifest() *NetworkManifest {
	return m.manifest
}

// Initialize configures the adapter
func (m *MoltbookAdapter) Initialize(config map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if baseURL, ok := config["base_url"].(string); ok {
		m.baseURL = baseURL
		m.manifest.BaseURL = baseURL
	}

	if timeout, ok := config["timeout"].(int); ok {
		m.httpClient.Timeout = time.Duration(timeout) * time.Second
	}

	if cacheTTL, ok := config["cache_ttl"].(int); ok {
		m.cacheTTL = time.Duration(cacheTTL) * time.Minute
	}

	return nil
}

// Authenticate sets the authentication token
func (m *MoltbookAdapter) Authenticate(ctx context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate token by making a test request
	testURL := m.baseURL + "/api/v1/agents/me"
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed: status %d", resp.StatusCode)
	}

	m.token = token
	return nil
}

// IsAuthenticated returns true if authenticated
func (m *MoltbookAdapter) IsAuthenticated() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.token != ""
}

// GetAuthToken returns the current auth token
func (m *MoltbookAdapter) GetAuthToken() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.token
}

// Call executes a dynamic action
func (m *MoltbookAdapter) Call(ctx context.Context, action string, params map[string]interface{}) (interface{}, error) {
	if !m.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	endpoint := m.manifest.SocialFeatures.GetEndpoint(action)
	if endpoint == "" {
		return nil, &NoEndpointError{Action: action, Adapter: "moltbook"}
	}

	for key, value := range params {
		placeholder := "{" + key + "}"
		if strVal, ok := value.(string); ok {
			endpoint = replacePlaceholder(endpoint, placeholder, url.PathEscape(strVal))
		}
	}

	// Determine HTTP method
	method := "POST"
	if action == "feed" || action == "notifications" || action == "profile" {
		method = "GET"
	}

	return m.executeRequest(ctx, endpoint, method, params)
}

// Post creates a new post
func (m *MoltbookAdapter) Post(ctx context.Context, content PostContent) (interface{}, error) {
	if !m.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	return m.executeRequest(ctx, "/api/v1/posts", "POST", map[string]interface{}{
		"body":       content.Body,
		"tags":       content.Tags,
		"parent_id":  content.ParentID,
		"media_urls": content.MediaURLs,
		"mentions":   content.Mentions,
	})
}

// Vote casts a vote on content
func (m *MoltbookAdapter) Vote(ctx context.Context, content VoteContent) (interface{}, error) {
	if !m.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	endpoint := "/api/v1/posts/" + url.PathEscape(content.TargetID) + "/vote"
	return m.executeRequest(ctx, endpoint, "POST", map[string]interface{}{
		"direction": content.Direction,
	})
}

// Follow follows an agent
func (m *MoltbookAdapter) Follow(ctx context.Context, content FollowContent) (interface{}, error) {
	if !m.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	endpoint := "/api/v1/agents/" + url.PathEscape(content.TargetID) + "/follow"
	return m.executeRequest(ctx, endpoint, "POST", nil)
}

// GetFeed retrieves the feed
func (m *MoltbookAdapter) GetFeed(ctx context.Context, request FeedRequest) ([]FeedItem, error) {
	if !m.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	// Check cache
	m.mu.RLock()
	if len(m.cachedFeed) > 0 && time.Since(m.cacheTime) < m.cacheTTL {
		feed := m.cachedFeed
		m.mu.RUnlock()
		return feed, nil
	}
	m.mu.RUnlock()

	params := make(map[string]interface{})
	if request.AgentID != "" {
		params["agent_id"] = request.AgentID
	}
	if request.Limit > 0 {
		params["limit"] = request.Limit
	}
	if request.Offset > 0 {
		params["offset"] = request.Offset
	}
	if request.Filter != "" {
		params["filter"] = request.Filter
	}

	result, err := m.executeRequest(ctx, "/api/v1/feed", "GET", params)
	if err != nil {
		return nil, err
	}

	// Parse feed items
	items, err := parseFeedItemsFromMoltbook(result)
	if err != nil {
		return nil, err
	}

	// Update cache
	m.mu.Lock()
	m.cachedFeed = items
	m.cacheTime = time.Now()
	m.mu.Unlock()

	return items, nil
}

// GetNotifications retrieves notifications
func (m *MoltbookAdapter) GetNotifications(ctx context.Context, limit int) ([]Notification, error) {
	if !m.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	params := map[string]interface{}{"limit": limit}
	result, err := m.executeRequest(ctx, "/api/v1/notifications", "GET", params)
	if err != nil {
		return nil, err
	}

	return parseNotificationsFromMoltbook(result)
}

// HealthCheck verifies the connection
func (m *MoltbookAdapter) HealthCheck(ctx context.Context) error {
	if !m.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}

	testURL := m.baseURL + "/api/v1/health"
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}

	return nil
}

// executeRequest performs an HTTP request
func (m *MoltbookAdapter) executeRequest(ctx context.Context, endpoint, method string, params map[string]interface{}) (interface{}, error) {
	m.mu.RLock()
	token := m.token
	baseURL := m.baseURL
	m.mu.RUnlock()

	var body io.Reader
	if params != nil && (method == "POST" || method == "PUT") {
		jsonData, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	url := baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if len(respBody) == 0 {
		return nil, nil
	}

	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		// Return raw response if not JSON
		return string(respBody), nil
	}

	return result, nil
}

// replacePlaceholder replaces a placeholder in URL with value
func replacePlaceholder(url, placeholder, value string) string {
	return string(bytes.Replace([]byte(url), []byte(placeholder), []byte(value), 1))
}

// parseFeedItemsFromMoltbook converts Moltbook response to FeedItems
func parseFeedItemsFromMoltbook(data interface{}) ([]FeedItem, error) {
	// Try to parse as array
	items := make([]FeedItem, 0)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var response struct {
		Items []struct {
			ID        string    `json:"id"`
			Type      string    `json:"type"`
			Body      string    `json:"body"`
			Tags      []string  `json:"tags"`
			Votes     int       `json:"votes"`
			Comments  int       `json:"comments"`
			Reposts   int       `json:"reposts"`
			CreatedAt time.Time `json:"created_at"`
			Author    struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				Handle    string `json:"handle"`
				AvatarURL string `json:"avatar_url"`
			} `json:"author"`
		} `json:"items"`
	}

	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, err
	}

	for _, item := range response.Items {
		items = append(items, FeedItem{
			ID:   item.ID,
			Type: item.Type,
			Content: map[string]interface{}{
				"body": item.Body,
				"tags": item.Tags,
			},
			Author: AuthorInfo{
				ID:        item.Author.ID,
				Name:      item.Author.Name,
				Handle:    item.Author.Handle,
				AvatarURL: item.Author.AvatarURL,
			},
			Votes:     item.Votes,
			Comments:  item.Comments,
			Reposts:   item.Reposts,
			CreatedAt: item.CreatedAt,
		})
	}

	return items, nil
}

// parseNotificationsFromMoltbook converts Moltbook response to Notifications
func parseNotificationsFromMoltbook(data interface{}) ([]Notification, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var response struct {
		Notifications []struct {
			ID        string    `json:"id"`
			Type      string    `json:"type"`
			Read      bool      `json:"read"`
			CreatedAt time.Time `json:"created_at"`
			Actor     struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				Handle    string `json:"handle"`
				AvatarURL string `json:"avatar_url"`
			} `json:"actor"`
			Target map[string]interface{} `json:"target,omitempty"`
		} `json:"notifications"`
	}

	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, err
	}

	notifs := make([]Notification, len(response.Notifications))
	for i, n := range response.Notifications {
		notifs[i] = Notification{
			ID:        n.ID,
			Type:      n.Type,
			Read:      n.Read,
			CreatedAt: n.CreatedAt,
			Actor: AuthorInfo{
				ID:        n.Actor.ID,
				Name:      n.Actor.Name,
				Handle:    n.Actor.Handle,
				AvatarURL: n.Actor.AvatarURL,
			},
			Target: n.Target,
		}
	}

	return notifs, nil
}
