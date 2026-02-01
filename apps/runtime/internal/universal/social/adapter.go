package social

import (
	"context"
)

// SocialAdapter defines the interface for social network implementations
// Each social network (Moltbook, Discord, etc.) implements this interface
type SocialAdapter interface {
	// Identity
	Name() string        // Unique name: "moltbook", "discord", etc.
	Version() string     // Adapter version
	DisplayName() string // Human readable: "Moltbook", "Discord"

	// Capabilities - declare what this adapter supports
	Capabilities() SocialCapabilities

	// Manifest - get the network manifest
	GetManifest() *NetworkManifest

	// Initialize with configuration
	Initialize(config map[string]interface{}) error

	// Authentication
	Authenticate(ctx context.Context, token string) error
	IsAuthenticated() bool
	GetAuthToken() string

	// Core social operations - return dynamic results based on capabilities
	Call(ctx context.Context, action string, params map[string]interface{}) (interface{}, error)

	// Named operations for type safety
	Post(ctx context.Context, content PostContent) (interface{}, error)
	Vote(ctx context.Context, content VoteContent) (interface{}, error)
	Follow(ctx context.Context, content FollowContent) (interface{}, error)
	GetFeed(ctx context.Context, request FeedRequest) ([]FeedItem, error)
	GetNotifications(ctx context.Context, limit int) ([]Notification, error)

	// Health check
	HealthCheck(ctx context.Context) error
}

// AdapterResult represents the result of a social operation
type AdapterResult struct {
	Success   bool                   `json:"success"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	ErrorCode string                 `json:"error_code,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// CallOptions provides additional options for adapter calls
type CallOptions struct {
	Timeout  int               // Timeout in seconds
	Headers  map[string]string // Additional headers
	Retries  int               // Number of retries
	Endpoint string            // Override endpoint
	Method   string            // HTTP method override
}

// ExecuteCall performs the actual HTTP call - to be implemented by adapters
type ExecuteCallFunc func(ctx context.Context, endpoint, method string, params map[string]interface{}, opts CallOptions) (interface{}, error)

// Call executes a social action on the adapter
func Call(ctx context.Context, adapter SocialAdapter, action string, params map[string]interface{}) (interface{}, error) {
	// Check if action is supported
	caps := adapter.Capabilities()
	if !caps.HasCapability(action) {
		return nil, &UnsupportedActionError{
			Action:       action,
			Adapter:      adapter.Name(),
			Capabilities: caps,
		}
	}

	// Get endpoint for action
	endpoint := caps.GetEndpoint(action)
	if endpoint == "" {
		return nil, &NoEndpointError{Action: action, Adapter: adapter.Name()}
	}

	return adapter.Call(ctx, action, params)
}

// CreatePost creates a new post
func CreatePost(ctx context.Context, adapter SocialAdapter, content PostContent) (interface{}, error) {
	if !adapter.Capabilities().SupportsPost {
		return nil, &UnsupportedActionError{Action: "post", Adapter: adapter.Name()}
	}

	return adapter.Post(ctx, content)
}

// Vote casts a vote on content
func Vote(ctx context.Context, adapter SocialAdapter, content VoteContent) (interface{}, error) {
	if !adapter.Capabilities().SupportsVote {
		return nil, &UnsupportedActionError{Action: "vote", Adapter: adapter.Name()}
	}

	return adapter.Vote(ctx, content)
}

// Follow follows an agent or user
func Follow(ctx context.Context, adapter SocialAdapter, content FollowContent) (interface{}, error) {
	if !adapter.Capabilities().SupportsFollow {
		return nil, &UnsupportedActionError{Action: "follow", Adapter: adapter.Name()}
	}

	return adapter.Follow(ctx, content)
}

// GetFeed retrieves a feed
func GetFeed(ctx context.Context, adapter SocialAdapter, request FeedRequest) ([]FeedItem, error) {
	if !adapter.Capabilities().SupportsFeed {
		return nil, &UnsupportedActionError{Action: "feed", Adapter: adapter.Name()}
	}

	return adapter.GetFeed(ctx, request)
}

// GetNotifications retrieves notifications
func GetNotifications(ctx context.Context, adapter SocialAdapter, limit int) ([]Notification, error) {
	result, err := adapter.Call(ctx, "notifications", map[string]interface{}{
		"limit": limit,
	})
	if err != nil {
		return nil, err
	}

	// Convert result to notifications
	notifs, ok := result.([]Notification)
	if !ok {
		// Try to parse from map
		return parseNotifications(result)
	}
	return notifs, nil
}

// parseNotifications converts dynamic result to Notification slice
func parseNotifications(result interface{}) ([]Notification, error) {
	// Default implementation - adapters should override
	return nil, &NotImplementedError{Method: "parseNotifications"}
}

// Error types for social adapter operations
type UnsupportedActionError struct {
	Action       string
	Adapter      string
	Capabilities SocialCapabilities
}

func (e *UnsupportedActionError) Error() string {
	return "action '" + e.Action + "' not supported by adapter '" + e.Adapter + "'"
}

type NoEndpointError struct {
	Action  string
	Adapter string
}

func (e *NoEndpointError) Error() string {
	return "no endpoint configured for action '" + e.Action + "' in adapter '" + e.Adapter + "'"
}

type NotImplementedError struct {
	Method string
}

func (e *NotImplementedError) Error() string {
	return "method '" + e.Method + "' not implemented"
}

// AdapterOption functional option for adapter configuration
type AdapterOption func(*AdapterConfig)

type AdapterConfig struct {
	Timeout   int
	Retries   int
	CacheSize int
	Headers   map[string]string
}

// WithTimeout sets the request timeout
func WithTimeout(seconds int) AdapterOption {
	return func(c *AdapterConfig) {
		c.Timeout = seconds
	}
}

// WithRetries sets the number of retries
func WithRetries(retries int) AdapterOption {
	return func(c *AdapterConfig) {
		c.Retries = retries
	}
}

// WithHeaders sets additional headers
func WithHeaders(headers map[string]string) AdapterOption {
	return func(c *AdapterConfig) {
		c.Headers = headers
	}
}
