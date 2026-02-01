package social

import (
	"encoding/json"
	"time"
)

// SocialCapabilities declares what social features a network supports
type SocialCapabilities struct {
	// Feature support flags
	SupportsPost    bool `json:"supports_post"`
	SupportsVote    bool `json:"supports_vote"`
	SupportsFollow  bool `json:"supports_follow"`
	SupportsFeed    bool `json:"supports_feed"`
	SupportsReply   bool `json:"supports_reply"`
	SupportsRepost  bool `json:"supports_repost"`
	SupportsMessage bool `json:"supports_message"`

	// Dynamic endpoint discovery
	Endpoints map[string]string `json:"endpoints"`

	// Rate limits
	RateLimit struct {
		PostsPerMinute   int `json:"posts_per_minute"`
		VotesPerMinute   int `json:"votes_per_minute"`
		FollowsPerMinute int `json:"follows_per_minute"`
	} `json:"rate_limit"`
}

// NetworkManifest represents the manifest for a social network
type NetworkManifest struct {
	Name           string             `json:"name"`
	Version        string             `json:"version"`
	DisplayName    string             `json:"display_name"`
	Description    string             `json:"description"`
	SocialFeatures SocialCapabilities `json:"social_features"`

	// Authentication configuration
	Auth AuthConfig `json:"auth"`

	// API versioning
	APIVersion string `json:"api_version"`

	// Base URL for the network
	BaseURL string `json:"base_url"`

	// Feature flags
	Features map[string]bool `json:"features"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthConfig defines authentication requirements
type AuthConfig struct {
	Type     string   `json:"type"` // "bearer", "api_key", "oauth2"
	Required bool     `json:"required"`
	Scopes   []string `json:"scopes,omitempty"`
	Endpoint string   `json:"endpoint,omitempty"`
}

// PostContent represents content to be posted to a social network
type PostContent struct {
	Body      string   `json:"body"`
	Tags      []string `json:"tags,omitempty"`
	ParentID  string   `json:"parent_id,omitempty"`
	MediaURLs []string `json:"media_urls,omitempty"`
	Mentions  []string `json:"mentions,omitempty"`
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

// VoteContent represents a vote action
type VoteContent struct {
	TargetID   string `json:"target_id"`
	TargetType string `json:"target_type"` // "post", "comment"
	Direction  string `json:"direction"`   // "up", "down"
}

// FollowContent represents a follow action
type FollowContent struct {
	TargetID   string `json:"target_id"`
	TargetType string `json:"target_type"` // "agent", "user"
}

// FeedRequest represents a request for an agent's feed
type FeedRequest struct {
	AgentID   string     `json:"agent_id,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
	SinceTime *time.Time `json:"since_time,omitempty"`
	Filter    string     `json:"filter,omitempty"` // "all", "following", "trending"
}

// FeedItem represents a single item in a feed
type FeedItem struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // "post", "repost", "reply"
	Content   map[string]interface{} `json:"content"`
	Author    AuthorInfo             `json:"author"`
	Votes     int                    `json:"votes"`
	Comments  int                    `json:"comments"`
	Reposts   int                    `json:"reposts"`
	CreatedAt time.Time              `json:"created_at"`
}

// AuthorInfo represents information about a post author
type AuthorInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Handle    string `json:"handle,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// Notification represents a social notification
type Notification struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // "vote", "follow", "reply", "mention"
	Actor     AuthorInfo             `json:"actor"`
	Target    map[string]interface{} `json:"target,omitempty"`
	Read      bool                   `json:"read"`
	CreatedAt time.Time              `json:"created_at"`
}

// ParseManifest parses a JSON manifest into NetworkManifest
func ParseManifest(data []byte) (*NetworkManifest, error) {
	var manifest NetworkManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

// ToJSON converts manifest to JSON bytes
func (m *NetworkManifest) ToJSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

// HasCapability checks if a specific capability is supported
func (c SocialCapabilities) HasCapability(name string) bool {
	switch name {
	case "post":
		return c.SupportsPost
	case "vote":
		return c.SupportsVote
	case "follow":
		return c.SupportsFollow
	case "feed":
		return c.SupportsFeed
	case "reply":
		return c.SupportsReply
	case "repost":
		return c.SupportsRepost
	case "message":
		return c.SupportsMessage
	default:
		return false
	}
}

// GetEndpoint returns the endpoint URL for a given action
func (c SocialCapabilities) GetEndpoint(action string) string {
	if c.Endpoints == nil {
		return ""
	}
	return c.Endpoints[action]
}

// Endpoint represents a discovered API endpoint
type Endpoint struct {
	Path        string                 `json:"path"`
	Method      string                 `json:"method"`
	Description string                 `json:"description"`
	Parameters  []Parameter            `json:"parameters"`
	Response    map[string]interface{} `json:"response"`
}

// Parameter represents an API parameter
type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}
