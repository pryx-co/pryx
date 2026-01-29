package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pryx-core/internal/mcp"
)

func runMCP(args []string) int {
	if len(args) < 1 {
		mcpUsage()
		return 2
	}

	cmd := args[0]

	switch cmd {
	case "list", "ls":
		return runMCPList(args[1:])
	case "add":
		return runMCPAdd(args[1:])
	case "remove", "rm", "delete":
		return runMCPRemove(args[1:])
	case "test":
		return runMCPTest(args[1:])
	case "auth":
		return runMCPAuth(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		mcpUsage()
		return 2
	}
}

func runMCPList(args []string) int {
	jsonOutput := false

	for _, arg := range args {
		if arg == "--json" || arg == "-j" {
			jsonOutput = true
		}
	}

	cfg, path, err := mcp.LoadServersConfigFromFirstExisting(mcp.DefaultServersConfigPaths())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
		return 1
	}

	if path == "" {
		path = getDefaultMCPServerPath()
	}

	if jsonOutput {
		data, err := json.MarshalIndent(cfg.Servers, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to marshal servers: %v\n", err)
			return 1
		}
		fmt.Println(string(data))
	} else {
		fmt.Printf("MCP Servers (%d)\n", len(cfg.Servers))
		fmt.Println(strings.Repeat("=", 50))
		fmt.Printf("Config: %s\n\n", path)

		if len(cfg.Servers) == 0 {
			fmt.Println("No MCP servers configured.")
			fmt.Println("Use 'pryx-core mcp add <name> --url <url>' or")
			fmt.Println("       'pryx-core mcp add <name> --cmd <command>' to add a server.")
		} else {
			for name, server := range cfg.Servers {
				status := "configured"
				if server.Transport == "bundled" {
					status = "bundled"
				}
				fmt.Printf("• %s [%s]\n", name, status)
				if server.URL != "" {
					fmt.Printf("  URL: %s\n", server.URL)
				}
				if len(server.Command) > 0 {
					fmt.Printf("  Command: %s\n", strings.Join(server.Command, " "))
				}
				if server.Auth != nil {
					fmt.Printf("  Auth: %s\n", server.Auth.Type)
				}
				fmt.Println()
			}
		}
	}

	return 0
}

func runMCPAdd(args []string) int {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: server name and connection method required\n")
		mcpUsage()
		return 2
	}

	name := args[0]
	serverURL := ""
	var command []string
	var authType string
	var authTokenRef string

	// Parse flags
	i := 1
	for i < len(args) {
		arg := args[i]
		switch arg {
		case "--url", "-u":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --url requires a value\n")
				return 2
			}
			serverURL = args[i+1]
			i += 2
		case "--cmd", "-c":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --cmd requires a value\n")
				return 2
			}
			command = []string{args[i+1]}
			i += 2
		case "--auth":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --auth requires a value (bearer, basic, etc.)\n")
				return 2
			}
			authType = args[i+1]
			i += 2
		case "--token-ref":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --token-ref requires a value\n")
				return 2
			}
			authTokenRef = args[i+1]
			i += 2
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
			mcpUsage()
			return 2
		}
	}

	if serverURL == "" && len(command) == 0 {
		fmt.Fprintf(os.Stderr, "Error: either --url or --cmd required\n")
		return 2
	}

	// Load existing config
	cfg, _, err := mcp.LoadServersConfigFromFirstExisting(mcp.DefaultServersConfigPaths())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
		return 1
	}

	// Check if server already exists
	if _, exists := cfg.Servers[name]; exists {
		fmt.Printf("ℹ Server '%s' already exists, updating...\n", name)
	}

	// Create server config
	serverCfg := mcp.ServerConfig{
		Transport: "stdio",
		URL:       serverURL,
		Command:   command,
	}

	if authType != "" {
		serverCfg.Auth = &mcp.AuthConfig{
			Type:     authType,
			TokenRef: authTokenRef,
		}
	}

	cfg.Servers[name] = serverCfg

	// Save config
	path := getDefaultMCPServerPath()
	if err := saveMCPServerConfig(path, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save config: %v\n", err)
		return 1
	}

	fmt.Printf("✓ Added MCP server: %s\n", name)
	fmt.Printf("  Config saved to: %s\n", path)

	return 0
}

func runMCPRemove(args []string) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: server name required\n")
		return 2
	}

	name := args[0]

	cfg, _, err := mcp.LoadServersConfigFromFirstExisting(mcp.DefaultServersConfigPaths())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
		return 1
	}

	if _, exists := cfg.Servers[name]; !exists {
		fmt.Fprintf(os.Stderr, "Error: server '%s' not found\n", name)
		return 1
	}

	delete(cfg.Servers, name)

	path := getDefaultMCPServerPath()
	if err := saveMCPServerConfig(path, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save config: %v\n", err)
		return 1
	}

	fmt.Printf("✓ Removed MCP server: %s\n", name)

	return 0
}

func runMCPTest(args []string) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: server name required\n")
		return 2
	}

	name := args[0]

	cfg, _, err := mcp.LoadServersConfigFromFirstExisting(mcp.DefaultServersConfigPaths())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
		return 1
	}

	server, exists := cfg.Servers[name]
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: server '%s' not found\n", name)
		return 1
	}

	fmt.Printf("Testing MCP server: %s\n", name)
	fmt.Println(strings.Repeat("=", 40))

	// Test connection based on transport type
	if server.URL != "" {
		fmt.Printf("URL: %s\n", server.URL)
		fmt.Println("(HTTP transport testing not implemented in CLI)")
		fmt.Println("✓ Server configuration looks valid")
	} else if len(server.Command) > 0 {
		fmt.Printf("Command: %s\n", strings.Join(server.Command, " "))
		fmt.Println("(Stdio transport testing requires runtime context)")
		fmt.Println("✓ Server configuration looks valid")
	} else if server.Transport == "bundled" {
		fmt.Println("Bundled server - available when runtime starts")
		fmt.Println("✓ Server configuration looks valid")
	} else {
		fmt.Println("✗ No valid transport configured")
		return 1
	}

	return 0
}

func runMCPAuth(args []string) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: server name required\n")
		return 2
	}

	name := args[0]

	cfg, _, err := mcp.LoadServersConfigFromFirstExisting(mcp.DefaultServersConfigPaths())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
		return 1
	}

	server, exists := cfg.Servers[name]
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: server '%s' not found\n", name)
		return 1
	}

	fmt.Printf("MCP Server Authentication: %s\n", name)
	fmt.Println(strings.Repeat("=", 40))

	if server.Auth == nil {
		fmt.Println("No authentication configured.")
		fmt.Println("Use: pryx-core mcp auth <name> <type> [--token-ref <ref>]")
		fmt.Println("  Types: bearer, basic, api_key")
	} else {
		fmt.Printf("Type: %s\n", server.Auth.Type)
		if server.Auth.TokenRef != "" {
			fmt.Printf("Token Reference: %s\n", server.Auth.TokenRef)
			fmt.Println("(Token stored in OS keychain)")
		} else {
			fmt.Println("Token: [embedded in config - not recommended]")
		}
	}

	return 0
}

func mcpUsage() {
	fmt.Println("pryx-core mcp - Manage MCP servers")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  list                          List configured MCP servers")
	fmt.Println("  add <name> --url <url>        Add HTTP MCP server")
	fmt.Println("  add <name> --cmd <command>    Add stdio MCP server")
	fmt.Println("  remove <name>                 Remove an MCP server")
	fmt.Println("  test <name>                   Test MCP server connection")
	fmt.Println("  auth <name>                   Manage authentication")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --url, -u <url>               Server URL (for HTTP transport)")
	fmt.Println("  --cmd, -c <command>           Command (for stdio transport)")
	fmt.Println("  --auth <type>                 Authentication type (bearer, basic)")
	fmt.Println("  --token-ref <ref>             Token reference (keychain)")
	fmt.Println("  --json, -j                    Output in JSON format")
}

func getDefaultMCPServerPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pryx", "mcp", "servers.json")
}

func saveMCPServerConfig(path string, cfg *mcp.ServersConfig) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
