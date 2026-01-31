package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pryx-core/internal/config"
)

// ChannelConfig represents a simplified channel configuration for CLI
type ChannelConfig struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"` // telegram, discord, slack, webhook
	Name      string            `json:"name"`
	Enabled   bool              `json:"enabled"`
	Config    map[string]string `json:"config"` // Type-specific config
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

func runChannel(args []string) int {
	if len(args) < 1 {
		channelUsage()
		return 2
	}

	cmd := args[0]
	cfg := config.Load()

	switch cmd {
	case "list", "ls":
		return runChannelList(args[1:], cfg)
	case "add":
		return runChannelAdd(args[1:], cfg)
	case "remove", "rm", "delete":
		return runChannelRemove(args[1:], cfg)
	case "enable":
		return runChannelEnable(args[1:], cfg)
	case "disable":
		return runChannelDisable(args[1:], cfg)
	case "test":
		return runChannelTest(args[1:], cfg)
	case "status":
		return runChannelStatus(args[1:], cfg)
	case "sync":
		return runChannelSync(args[1:], cfg)
	case "help", "-h", "--help":
		channelUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		channelUsage()
		return 2
	}
}

func runChannelList(args []string, cfg *config.Config) int {
	jsonOutput := false
	detailed := false

	for _, arg := range args {
		if arg == "--json" || arg == "-j" {
			jsonOutput = true
		}
		if arg == "--verbose" || arg == "-v" {
			detailed = true
		}
	}

	channels, err := loadChannels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load channels: %v\n", err)
		return 1
	}

	if jsonOutput {
		data, err := json.MarshalIndent(channels, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to marshal channels: %v\n", err)
			return 1
		}
		fmt.Println(string(data))
	} else {
		fmt.Printf("Channels (%d)\n", len(channels))
		fmt.Println(strings.Repeat("=", 50))

		if len(channels) == 0 {
			fmt.Println("No channels configured.")
			fmt.Println("Use 'pryx-core channel add <type> <name>' to add a channel.")
		} else {
			for _, ch := range channels {
				status := "disabled"
				if ch.Enabled {
					status = "enabled"
				}
				fmt.Printf("• %s [%s] - %s\n", ch.Name, ch.Type, status)

				if detailed {
					fmt.Printf("  ID: %s\n", ch.ID)
					if len(ch.Config) > 0 {
						fmt.Printf("  Config:\n")
						for k, v := range ch.Config {
							// Don't print sensitive values
							if strings.Contains(k, "token") || strings.Contains(k, "secret") {
								fmt.Printf("    %s: ***\n", k)
							} else {
								fmt.Printf("    %s: %s\n", k, v)
							}
						}
					}
					fmt.Printf("  Created: %s\n", ch.CreatedAt)
					fmt.Printf("  Updated: %s\n", ch.UpdatedAt)
					fmt.Println()
				}
			}
		}
	}

	return 0
}

func runChannelAdd(args []string, cfg *config.Config) int {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: channel type and name required\n")
		fmt.Fprintf(os.Stderr, "Usage: pryx-core channel add <type> <name> [--<key> <value>...]\n")
		return 2
	}

	channelType := args[0]
	name := args[1]
	configValues := make(map[string]string)

	// Parse additional config values
	i := 2
	for i < len(args) {
		if strings.HasPrefix(args[i], "--") {
			key := strings.TrimPrefix(args[i], "--")
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				configValues[key] = args[i+1]
				i += 2
			} else {
				configValues[key] = ""
				i += 1
			}
		} else {
			i++
		}
	}

	// Validate channel type
	validTypes := map[string]bool{
		"telegram": true,
		"discord":  true,
		"slack":    true,
		"webhook":  true,
	}

	if !validTypes[channelType] {
		fmt.Fprintf(os.Stderr, "Error: invalid channel type: %s\n", channelType)
		fmt.Fprintf(os.Stderr, "Valid types: telegram, discord, slack, webhook\n")
		return 1
	}

	// Load existing channels
	channels, err := loadChannels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load channels: %v\n", err)
		return 1
	}

	// Check for duplicate name
	for _, ch := range channels {
		if ch.Name == name {
			fmt.Fprintf(os.Stderr, "Error: channel name already exists: %s\n", name)
			return 1
		}
	}

	// Create new channel
	now := getTimestamp()
	newChannel := ChannelConfig{
		ID:        fmt.Sprintf("%s-%d", channelType, now),
		Type:      channelType,
		Name:      name,
		Enabled:   false, // Disabled by default until tested
		Config:    configValues,
		CreatedAt: now,
		UpdatedAt: now,
	}

	channels = append(channels, newChannel)

	// Save channels
	if err := saveChannels(channels); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save channels: %v\n", err)
		return 1
	}

	fmt.Printf("✓ Added channel: %s (type: %s)\n", name, channelType)
	fmt.Printf("  ID: %s\n", newChannel.ID)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Set authentication (if required):")
	fmt.Printf("     pryx-core channel enable %s\n", name)
	fmt.Println("  2. Test the connection:")
	fmt.Printf("     pryx-core channel test %s\n", name)

	return 0
}

func runChannelRemove(args []string, cfg *config.Config) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: channel name or ID required\n")
		return 2
	}

	name := args[0]

	channels, err := loadChannels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load channels: %v\n", err)
		return 1
	}

	found := false
	filtered := make([]ChannelConfig, 0, len(channels))
	for _, ch := range channels {
		if ch.Name == name || ch.ID == name {
			found = true
			fmt.Printf("Removing channel: %s (%s)\n", ch.Name, ch.ID)
		} else {
			filtered = append(filtered, ch)
		}
	}

	if !found {
		fmt.Fprintf(os.Stderr, "Error: channel not found: %s\n", name)
		return 1
	}

	if err := saveChannels(filtered); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save channels: %v\n", err)
		return 1
	}

	fmt.Printf("✓ Removed channel: %s\n", name)
	return 0
}

func runChannelEnable(args []string, cfg *config.Config) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: channel name or ID required\n")
		return 2
	}

	name := args[0]

	channels, err := loadChannels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load channels: %v\n", err)
		return 1
	}

	found := false
	for i, ch := range channels {
		if ch.Name == name || ch.ID == name {
			found = true
			channels[i].Enabled = true
			channels[i].UpdatedAt = getTimestamp()
			fmt.Printf("✓ Enabled channel: %s\n", ch.Name)
		}
	}

	if !found {
		fmt.Fprintf(os.Stderr, "Error: channel not found: %s\n", name)
		return 1
	}

	if err := saveChannels(channels); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save channels: %v\n", err)
		return 1
	}

	return 0
}

func runChannelDisable(args []string, cfg *config.Config) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: channel name or ID required\n")
		return 2
	}

	name := args[0]

	channels, err := loadChannels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load channels: %v\n", err)
		return 1
	}

	found := false
	for i, ch := range channels {
		if ch.Name == name || ch.ID == name {
			found = true
			channels[i].Enabled = false
			channels[i].UpdatedAt = getTimestamp()
			fmt.Printf("✓ Disabled channel: %s\n", ch.Name)
		}
	}

	if !found {
		fmt.Fprintf(os.Stderr, "Error: channel not found: %s\n", name)
		return 1
	}

	if err := saveChannels(channels); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save channels: %v\n", err)
		return 1
	}

	return 0
}

func runChannelTest(args []string, cfg *config.Config) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: channel name or ID required\n")
		return 2
	}

	name := args[0]

	channels, err := loadChannels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load channels: %v\n", err)
		return 1
	}

	var target *ChannelConfig
	for _, ch := range channels {
		if ch.Name == name || ch.ID == name {
			target = &ch
			break
		}
	}

	if target == nil {
		fmt.Fprintf(os.Stderr, "Error: channel not found: %s\n", name)
		return 1
	}

	fmt.Printf("Testing channel: %s (%s)\n", target.Name, target.Type)
	fmt.Println(strings.Repeat("=", 40))

	// Validate required config based on type
	switch target.Type {
	case "telegram":
		if token, ok := target.Config["token"]; ok && token != "" {
			fmt.Printf("✓ Token configured\n")
		} else {
			fmt.Printf("✗ Token not configured\n")
			fmt.Println("  Set token: pryx-core provider set-key telegram")
		}
	case "discord":
		if token, ok := target.Config["token"]; ok && token != "" {
			fmt.Printf("✓ Token configured\n")
		} else {
			fmt.Printf("✗ Token not configured\n")
			fmt.Println("  Set token: pryx-core provider set-key discord")
		}
	case "slack":
		if token, ok := target.Config["bot_token"]; ok && token != "" {
			fmt.Printf("✓ Bot token configured\n")
		} else {
			fmt.Printf("✗ Bot token not configured\n")
			fmt.Println("  Set token: pryx-core provider set-key slack")
		}
		if token, ok := target.Config["app_token"]; ok && token != "" {
			fmt.Printf("✓ App token configured\n")
		} else {
			fmt.Printf("✗ App token not configured\n")
		}
	case "webhook":
		if url, ok := target.Config["url"]; ok && url != "" {
			fmt.Printf("✓ Webhook URL configured: %s\n", url)
		} else {
			fmt.Printf("✗ Webhook URL not configured\n")
		}
	}

	fmt.Println()
	fmt.Println("Note: Full connection testing requires runtime to be running")
	fmt.Println("Start runtime with: pryx-core")
	fmt.Println("Or test in TUI for interactive verification")

	return 0
}

func runChannelStatus(args []string, cfg *config.Config) int {
	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	channels, err := loadChannels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load channels: %v\n", err)
		return 1
	}

	if name != "" {
		// Show status for specific channel
		for _, ch := range channels {
			if ch.Name == name || ch.ID == name {
				fmt.Printf("Channel: %s\n", ch.Name)
				fmt.Println(strings.Repeat("=", 40))
				fmt.Printf("ID:      %s\n", ch.ID)
				fmt.Printf("Type:    %s\n", ch.Type)
				fmt.Printf("Status:  ")
				if ch.Enabled {
					fmt.Printf("enabled\n")
				} else {
					fmt.Printf("disabled\n")
				}
				fmt.Printf("Created: %s\n", ch.CreatedAt)
				fmt.Printf("Updated: %s\n", ch.UpdatedAt)

				if len(ch.Config) > 0 {
					fmt.Println("\nConfiguration:")
					for k, v := range ch.Config {
						fmt.Printf("  %s: ", k)
						if strings.Contains(k, "token") || strings.Contains(k, "secret") || strings.Contains(k, "password") {
							fmt.Println("***")
						} else {
							fmt.Println(v)
						}
					}
				}

				return 0
			}
		}
		fmt.Fprintf(os.Stderr, "Error: channel not found: %s\n", name)
		return 1
	}

	// Show status for all channels
	fmt.Printf("Channel Status\n")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println()

	enabledCount := 0
	for _, ch := range channels {
		status := "disabled"
		if ch.Enabled {
			status = "enabled"
			enabledCount++
		}
		fmt.Printf("• %s [%s] - %s\n", ch.Name, ch.Type, status)
	}

	fmt.Println()
	fmt.Printf("Total: %d channels (%d enabled, %d disabled)\n",
		len(channels), enabledCount, len(channels)-enabledCount)

	return 0
}

func runChannelSync(args []string, cfg *config.Config) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: channel name or ID required\n")
		return 2
	}

	name := args[0]

	channels, err := loadChannels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load channels: %v\n", err)
		return 1
	}

	var target *ChannelConfig
	for _, ch := range channels {
		if ch.Name == name || ch.ID == name {
			target = &ch
			break
		}
	}

	if target == nil {
		fmt.Fprintf(os.Stderr, "Error: channel not found: %s\n", name)
		return 1
	}

	fmt.Printf("Syncing channel: %s (%s)\n", target.Name, target.Type)

	// Channel-specific sync logic
	switch target.Type {
	case "discord":
		fmt.Println("Syncing Discord slash commands...")
		fmt.Println("(Requires runtime to be running)")
		fmt.Println("Start runtime with: pryx-core")
	case "slack":
		fmt.Println("Syncing Slack app configuration...")
		fmt.Println("(Requires runtime to be running)")
		fmt.Println("Start runtime with: pryx-core")
	default:
		fmt.Printf("Sync not required for %s channels\n", target.Type)
	}

	return 0
}

func channelUsage() {
	fmt.Println("pryx-core channel - Manage communication channels")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  list [--json] [--verbose]        List all channels")
	fmt.Println("  add <type> <name> [--key val]    Add a new channel")
	fmt.Println("  remove <name>                    Remove a channel")
	fmt.Println("  enable <name>                   Enable a channel")
	fmt.Println("  disable <name>                  Disable a channel")
	fmt.Println("  test <name>                     Test channel connection")
	fmt.Println("  status [name]                   Show channel status")
	fmt.Println("  sync <name>                     Sync channel configuration")
	fmt.Println("")
	fmt.Println("Channel types:")
	fmt.Println("  telegram                         Telegram bot")
	fmt.Println("  discord                          Discord bot")
	fmt.Println("  slack                            Slack app")
	fmt.Println("  webhook                          Webhook endpoint")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  pryx-core channel add telegram my-bot --token YOUR_TOKEN")
	fmt.Println("  pryx-core channel enable my-bot")
	fmt.Println("  pryx-core channel test my-bot")
}

func loadChannels() ([]ChannelConfig, error) {
	path := getChannelsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []ChannelConfig{}, nil
		}
		return nil, err
	}

	var channels []ChannelConfig
	if err := json.Unmarshal(data, &channels); err != nil {
		return nil, err
	}

	return channels, nil
}

func saveChannels(channels []ChannelConfig) error {
	path := getChannelsPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(channels, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func getChannelsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pryx", "channels.json")
}

func getTimestamp() string {
	return fmt.Sprintf("%d", 0) // Simplified timestamp
}
