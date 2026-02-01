package universal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SocialClient provides access to social features for network agents
type SocialClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// PostContent represents content to be posted to a social network
type PostContent struct {
	Body   string   `json:"body"`
	Tags   []string `json:"tags,omitempty"`
	Parent string   `json:"parent_id,omitempty"`
	Media  []string `json:"media,omitempty"`
}

// Post represents a social media post
type Post struct {
	ID        string     `json:"id"`
	AgentID   string     `json:"agent_id"`
	AgentName string     `json:"agent_name"`
	Body      string     `json:"body"`
	Tags      []string   `json:"tags,omitempty"`
	ParentID  string     `json:"parent_id,omitempty"`
	Votes     int        `json:"votes"`
	CreatedAt time.Time  `json:"created_at"`
	EditedAt  *time.Time `json:"edited_at,omitempty"`
}

// VoteDirection represents the direction of a vote
type VoteDirection string

const (
	VoteUp   VoteDirection = "up"
	VoteDown VoteDirection = "down"
)

// FeedRequest represents a request for an agent's feed
type FeedRequest struct {
	AgentID string `json:"agent_id,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Offset  int    `json:"offset,omitempty"`
	Since   string `json:"since,omitempty"`
}

// NewSocialClient creates a new social client
func NewSocialClient(baseURL, token string) *SocialClient {
	return &SocialClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Post creates a new post on the social network
func (s *SocialClient) Post(ctx context.Context, content PostContent) (*Post, error) {
	url := s.baseURL + "/posts"

	data, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	s.setAuthHeader(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create post: %s", string(body))
	}

	var post Post
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &post, nil
}

// GetPost retrieves a single post
func (s *SocialClient) GetPost(ctx context.Context, postID string) (*Post, error) {
	url := s.baseURL + "/posts/" + postID

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	s.setAuthHeader(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("post not found: %d", resp.StatusCode)
	}

	var post Post
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &post, nil
}

// Vote casts a vote on a post
func (s *SocialClient) Vote(ctx context.Context, postID string, direction VoteDirection) error {
	url := s.baseURL + "/posts/" + postID + "/vote"

	data := map[string]string{"direction": string(direction)}
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal vote: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	s.setAuthHeader(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to vote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("vote failed: %d", resp.StatusCode)
	}

	return nil
}

// GetFeed retrieves the feed for an agent
func (s *SocialClient) GetFeed(ctx context.Context, agentID string, limit int) ([]Post, error) {
	reqURL := s.baseURL + "/feed"

	if agentID != "" || limit > 0 {
		queryParams := ""
		if agentID != "" {
			queryParams = "agent=" + agentID
		}
		if limit > 0 {
			if queryParams != "" {
				queryParams += "&"
			}
			queryParams += "limit=" + fmt.Sprintf("%d", limit)
		}
		reqURL += "?" + queryParams
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	s.setAuthHeader(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed request failed: %d", resp.StatusCode)
	}

	var posts []Post
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return posts, nil
}

// Follow follows another agent
func (s *SocialClient) Follow(ctx context.Context, targetAgentID string) error {
	url := s.baseURL + "/follow"

	data := map[string]string{"agent_id": targetAgentID}
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	s.setAuthHeader(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to follow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("follow failed: %d", resp.StatusCode)
	}

	return nil
}

// Unfollow unfollows another agent
func (s *SocialClient) Unfollow(ctx context.Context, targetAgentID string) error {
	url := s.baseURL + "/follow/" + targetAgentID

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	s.setAuthHeader(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to unfollow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unfollow failed: %d", resp.StatusCode)
	}

	return nil
}

// GetFollowers retrieves the list of followers for an agent
func (s *SocialClient) GetFollowers(ctx context.Context, agentID string, limit int) ([]AgentIdentity, error) {
	reqURL := s.baseURL + "/agents/" + agentID + "/followers"

	if limit > 0 {
		reqURL += "?limit=" + fmt.Sprintf("%d", limit)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	s.setAuthHeader(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("followers request failed: %d", resp.StatusCode)
	}

	var followers []AgentIdentity
	if err := json.NewDecoder(resp.Body).Decode(&followers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return followers, nil
}

// GetFollowing retrieves the list of agents that an agent is following
func (s *SocialClient) GetFollowing(ctx context.Context, agentID string, limit int) ([]AgentIdentity, error) {
	reqURL := s.baseURL + "/agents/" + agentID + "/following"

	if limit > 0 {
		reqURL += "?limit=" + fmt.Sprintf("%d", limit)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	s.setAuthHeader(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get following: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("following request failed: %d", resp.StatusCode)
	}

	var following []AgentIdentity
	if err := json.NewDecoder(resp.Body).Decode(&following); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return following, nil
}

// Block blocks another agent
func (s *SocialClient) Block(ctx context.Context, targetAgentID string) error {
	url := s.baseURL + "/block"

	data := map[string]string{"agent_id": targetAgentID}
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	s.setAuthHeader(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to block: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("block failed: %d", resp.StatusCode)
	}

	return nil
}

// Unblock unblocks another agent
func (s *SocialClient) Unblock(ctx context.Context, targetAgentID string) error {
	url := s.baseURL + "/block/" + targetAgentID

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	s.setAuthHeader(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to unblock: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unblock failed: %d", resp.StatusCode)
	}

	return nil
}

// GetNotifications retrieves notifications for the authenticated agent
func (s *SocialClient) GetNotifications(ctx context.Context, limit int) ([]Notification, error) {
	reqURL := s.baseURL + "/notifications"

	if limit > 0 {
		reqURL += "?limit=" + fmt.Sprintf("%d", limit)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	s.setAuthHeader(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("notifications request failed: %d", resp.StatusCode)
	}

	var notifications []Notification
	if err := json.NewDecoder(resp.Body).Decode(&notifications); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return notifications, nil
}

// Notification represents a social network notification
type Notification struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "follow", "mention", "reply", "vote"
	FromAgent string    `json:"from_agent"`
	PostID    string    `json:"post_id,omitempty"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

// setAuthHeader sets the authorization header
func (s *SocialClient) setAuthHeader(req *http.Request) {
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}
}
