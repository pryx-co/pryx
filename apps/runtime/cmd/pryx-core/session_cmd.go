package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"pryx-core/internal/config"
	"pryx-core/internal/store"
)

func runSession(args []string) int {
	if len(args) < 1 {
		sessionUsage()
		return 2
	}

	cmd := args[0]
	cfg := config.Load()

	switch cmd {
	case "list", "ls":
		return runSessionList(args[1:], cfg)
	case "get", "show", "view":
		return runSessionGet(args[1:], cfg)
	case "delete", "remove", "rm":
		return runSessionDelete(args[1:], cfg)
	case "export":
		return runSessionExport(args[1:], cfg)
	case "fork":
		return runSessionFork(args[1:], cfg)
	case "help", "-h", "--help":
		sessionUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		sessionUsage()
		return 2
	}
}

func runSessionList(args []string, cfg *config.Config) int {
	jsonOutput := false
	limit := 100

	for _, arg := range args {
		if arg == "--json" || arg == "-j" {
			jsonOutput = true
		}
	}

	s, err := store.New(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize store: %v\n", err)
		return 1
	}
	defer s.Close()

	sessions, err := s.ListSessions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to list sessions: %v\n", err)
		return 1
	}

	if len(sessions) > limit {
		sessions = sessions[:limit]
	}

	if jsonOutput {
		data, err := json.MarshalIndent(sessions, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to marshal sessions: %v\n", err)
			return 1
		}
		fmt.Println(string(data))
	} else {
		fmt.Printf("Sessions (%d shown, max %d)\n", len(sessions), limit)
		fmt.Println(strings.Repeat("=", 50))

		if len(sessions) == 0 {
			fmt.Println("No sessions found.")
			fmt.Println("Start a conversation to create a session.")
		} else {
			for i, sess := range sessions {
				msgCount := 0
				if count, err := s.GetMessageCount(sess.ID); err == nil {
					msgCount = count
				}

				fmt.Printf("%d. %s\n", i+1, sess.Title)
				fmt.Printf("   ID:        %s\n", sess.ID)
				fmt.Printf("   Created:   %s\n", formatTime(sess.CreatedAt))
				fmt.Printf("   Updated:   %s\n", formatTime(sess.UpdatedAt))
				fmt.Printf("   Messages:  %d\n", msgCount)
				fmt.Println()
			}
		}
	}

	return 0
}

func runSessionGet(args []string, cfg *config.Config) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: session ID or name required\n")
		return 2
	}

	sessionID := args[0]
	jsonOutput := false
	detailed := false

	for _, arg := range args[1:] {
		if arg == "--json" || arg == "-j" {
			jsonOutput = true
		}
		if arg == "--verbose" || arg == "-v" {
			detailed = true
		}
	}

	s, err := store.New(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize store: %v\n", err)
		return 1
	}
	defer s.Close()

	sess, err := s.GetSession(sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: session not found: %v\n", err)
		return 1
	}

	if jsonOutput {
		data, err := json.MarshalIndent(sess, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to marshal session: %v\n", err)
			return 1
		}
		fmt.Println(string(data))
	} else {
		fmt.Printf("Session: %s\n", sess.Title)
		fmt.Println(strings.Repeat("=", 40))
		fmt.Printf("ID:        %s\n", sess.ID)
		fmt.Printf("Created:   %s\n", formatTime(sess.CreatedAt))
		fmt.Printf("Updated:   %s\n", formatTime(sess.UpdatedAt))

		if detailed {
			msgCount, _ := s.GetMessageCount(sess.ID)
			fmt.Printf("Messages:  %d\n", msgCount)
		}
	}

	return 0
}

func runSessionDelete(args []string, cfg *config.Config) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: session ID required\n")
		return 2
	}

	sessionID := args[0]
	force := false

	for _, arg := range args[1:] {
		if arg == "--force" || arg == "-f" {
			force = true
		}
	}

	if !force {
		fmt.Printf("Are you sure you want to delete session '%s'? This cannot be undone.\n", sessionID)
		fmt.Print("Type 'yes' to confirm: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "yes" {
			fmt.Println("Operation cancelled.")
			return 0
		}
	}

	s, err := store.New(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize store: %v\n", err)
		return 1
	}
	defer s.Close()

	_, err = s.GetSession(sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: session not found: %v\n", err)
		return 1
	}

	_, err = s.DB.Exec("DELETE FROM messages WHERE session_id = ?", sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to delete messages: %v\n", err)
		return 1
	}

	_, err = s.DB.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to delete session: %v\n", err)
		return 1
	}

	fmt.Printf("✓ Deleted session: %s\n", sessionID)
	return 0
}

func runSessionExport(args []string, cfg *config.Config) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: session ID required\n")
		return 2
	}

	sessionID := args[0]
	format := "json"
	outputFile := ""

	for i, arg := range args[1:] {
		if arg == "--format" && i+1 < len(args[1:]) {
			format = args[1:][i+1]
		}
		if arg == "--output" && i+1 < len(args[1:]) {
			outputFile = args[1:][i+1]
		}
	}

	s, err := store.New(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize store: %v\n", err)
		return 1
	}
	defer s.Close()

	sess, err := s.GetSession(sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: session not found: %v\n", err)
		return 1
	}

	// Get messages for session
	rows, err := s.DB.Query(`
		SELECT id, role, content, created_at 
		FROM messages 
		WHERE session_id = ? 
		ORDER BY created_at ASC
	`, sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get messages: %v\n", err)
		return 1
	}
	defer rows.Close()

	type Message struct {
		ID        string    `json:"id"`
		Role      string    `json:"role"`
		Content   string    `json:"content"`
		CreatedAt time.Time `json:"created_at"`
	}

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.Role, &msg.Content, &msg.CreatedAt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to scan message: %v\n", err)
			continue
		}
		messages = append(messages, msg)
	}

	type ExportSession struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	exportData := struct {
		Session  *ExportSession `json:"session"`
		Messages []Message      `json:"messages"`
	}{
		Session: &ExportSession{
			ID:        sess.ID,
			Title:     sess.Title,
			CreatedAt: sess.CreatedAt,
			UpdatedAt: sess.UpdatedAt,
		},
		Messages: messages,
	}

	if format == "json" {
		data, err := json.MarshalIndent(exportData, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to marshal export: %v\n", err)
			return 1
		}

		if outputFile != "" {
			if err := os.WriteFile(outputFile, data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to write file: %v\n", err)
				return 1
			}
			fmt.Printf("✓ Exported session to: %s\n", outputFile)
		} else {
			fmt.Println(string(data))
		}
	} else if format == "markdown" {
		output := fmt.Sprintf("# %s\n\n", sess.Title)
		output += fmt.Sprintf("Exported: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
		output += "---\n\n"
		for _, msg := range messages {
			role := strings.Title(msg.Role)
			output += fmt.Sprintf("## %s\n", role)
			output += fmt.Sprintf("%s\n\n", msg.Content)
		}

		if outputFile != "" {
			if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to write file: %v\n", err)
				return 1
			}
			fmt.Printf("✓ Exported session to: %s\n", outputFile)
		} else {
			fmt.Println(output)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: unsupported format: %s\n", format)
		fmt.Fprintf(os.Stderr, "Supported formats: json, markdown\n")
		return 1
	}

	return 0
}

func runSessionFork(args []string, cfg *config.Config) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: session ID required\n")
		return 2
	}

	sessionID := args[0]
	newTitle := ""

	for i, arg := range args[1:] {
		if arg == "--title" && i+1 < len(args[1:]) {
			newTitle = args[1:][i+1]
		}
	}

	s, err := store.New(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize store: %v\n", err)
		return 1
	}
	defer s.Close()

	originalSess, err := s.GetSession(sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: session not found: %v\n", err)
		return 1
	}

	if newTitle == "" {
		newTitle = fmt.Sprintf("Copy of %s", originalSess.Title)
	}

	newSess, err := s.CreateSession(newTitle)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create session: %v\n", err)
		return 1
	}

	// Copy messages to new session
	_, err = s.DB.Exec(`
		INSERT INTO messages (id, session_id, role, content, created_at, updated_at)
		SELECT 
			lower(hex(randomblob(16))),
			?,
			role, 
			content, 
			created_at, 
			updated_at
		FROM messages 
		WHERE session_id = ?
		ORDER BY created_at ASC
	`, newSess.ID, sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to copy messages: %v\n", err)
		return 1
	}

	fmt.Printf("✓ Forked session '%s' to '%s'\n", originalSess.Title, newTitle)
	fmt.Printf("  New ID: %s\n", newSess.ID)

	return 0
}

func sessionUsage() {
	fmt.Println("pryx-core session - Manage conversation sessions")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  list [--json]                   List all sessions")
	fmt.Println("  get <id> [--json] [--verbose]   Get session details")
	fmt.Println("  delete <id> [--force]            Delete a session")
	fmt.Println("  export <id> [--format]          Export session to file")
	fmt.Println("  fork <id> [--title]              Fork (copy) a session")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --json, -j                      Output in JSON format")
	fmt.Println("  --verbose, -v                    Show detailed information")
	fmt.Println("  --force, -f                     Skip confirmation for delete")
	fmt.Println("  --format <json|markdown>         Export format (default: json)")
	fmt.Println("  --output <file>                 Output file path")
	fmt.Println("  --title <name>                  New session title (for fork)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  pryx-core session list")
	fmt.Println("  pryx-core session get abc123")
	fmt.Println("  pryx-core session delete abc123 --force")
	fmt.Println("  pryx-core session export abc123 --format markdown --output chat.md")
	fmt.Println("  pryx-core session fork abc123 --title 'New Chat'")
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
