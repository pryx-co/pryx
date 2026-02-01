// Package social provides a universal, plugin-based system for social network integration
//
// This package enables agents to interact with any social network through a unified interface.
// Each social network is implemented as a plugin that declares its capabilities via a manifest.
//
// # Key Concepts
//
//   - SocialAdapter: Interface that all social network plugins must implement
//   - NetworkManifest: Declares capabilities and endpoints for a network
//   - Hub: Manages multiple adapters and provides unified access
//   - Registry: Manages adapter registration and discovery
//
// # Usage Example
//
//	// Create a hub and register adapters
//	hub := social.NewHub()
//	hub.RegisterAdapter(social.NewMoltbookAdapter("https://moltbook.com"))
//
//	// Use the unified API
//	hub.Post(ctx, "moltbook", social.PostContent{Body: "Hello world!"})
//	hub.Vote(ctx, "moltbook", social.VoteContent{...})
//
// # Adding New Networks
//
// To add a new social network, implement the SocialAdapter interface and register it:
//
//	type MyNetworkAdapter struct{}
//
//	func (m *MyNetworkAdapter) Name() string { return "mynetwork" }
//	func (m *MyNetworkAdapter) Capabilities() social.SocialCapabilities { ... }
//	// ... implement other methods
//
//	social.RegisterDefault("mynetwork", func() social.SocialAdapter {
//	    return &MyNetworkAdapter{}
//	})
package social

// Common social actions
const (
	ActionPost          = "post"
	ActionVote          = "vote"
	ActionFollow        = "follow"
	ActionUnfollow      = "unfollow"
	ActionFeed          = "feed"
	ActionNotifications = "notifications"
	ActionProfile       = "profile"
	ActionTimeline      = "timeline"
)

// Vote directions
const (
	VoteUp   = "up"
	VoteDown = "down"
)

// Target types
const (
	TargetTypePost    = "post"
	TargetTypeComment = "comment"
	TargetTypeAgent   = "agent"
	TargetTypeUser    = "user"
)

// Feed filters
const (
	FilterAll       = "all"
	FilterFollowing = "following"
	FilterTrending  = "trending"
)
