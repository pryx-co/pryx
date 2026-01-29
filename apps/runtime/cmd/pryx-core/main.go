package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// 	"pryx-core/internal/auth"

	"pryx-core/internal/agent"
	"pryx-core/internal/bus"
	"pryx-core/internal/channels"
	"pryx-core/internal/channels/telegram"
	"pryx-core/internal/config"
	"pryx-core/internal/doctor"
	"pryx-core/internal/keychain"
	"pryx-core/internal/mcp"
	"pryx-core/internal/mesh"
	"pryx-core/internal/server"
	"pryx-core/internal/store"
)

var (
	Version   = "dev"
	BuildDate = "unknown"
)

func main() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "skills":
			os.Exit(runSkills(os.Args[2:]))
		case "mcp":
			os.Exit(runMCP(os.Args[2:]))
		case "doctor":
			os.Exit(runDoctor())
			// 		case "login":
			// 			os.Exit(runLogin())
		case "config":
			os.Exit(runConfig(os.Args[2:]))
		case "help", "-h", "--help":
			usage()
			return
		}
	}

	log.Printf("Starting pryx-core version %s (built %s)", Version, BuildDate)

	cfg := config.Load()

	s, err := store.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer s.Close()

	kc := keychain.New("pryx")

	// server.New creates the Bus internally
	srv := server.New(cfg, s.DB, kc)
	b := srv.Bus()

	meshMgr := mesh.NewManager(cfg, b, s, kc)
	meshMgr.Start(context.Background())

	// Channels
	chanMgr := channels.NewManager(b)
	if cfg.TelegramEnabled && cfg.TelegramToken != "" {
		log.Println("Starting Telegram Bot...")
		tg := telegram.NewTelegramChannel("telegram-main", cfg.TelegramToken, b)
		if err := chanMgr.Register(tg); err != nil {
			log.Printf("Failed to register Telegram: %v", err)
		}
	}
	defer chanMgr.Shutdown()

	// Agent (AI Orchestrator)
	agt, err := agent.New(cfg, b)
	if err != nil {
		log.Printf("Warning: Failed to initialize Agent: %v", err)
	} else {
		log.Println("Starting AI Agent...")
		go agt.Run(context.Background())
	}

	srv.Bus().Publish(bus.NewEvent(bus.EventTraceEvent, "", map[string]interface{}{
		"kind":      "runtime.started",
		"version":   Version,
		"buildDate": BuildDate,
	}))

	// Start server in background (with dynamic port allocation)
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func usage() {
	log.Println("pryx-core")
	log.Println("")
	log.Println("Usage:")
	log.Println("  pryx-core")
	log.Println("  pryx-core skills <command>")
	log.Println("  pryx-core mcp <filesystem|shell|browser|clipboard>")
	log.Println("  pryx-core doctor")
	log.Println("  pryx-core doctor")
	log.Println("  pryx-core login")
	log.Println("  pryx-core config <set|get|list>")
	log.Println("")
	log.Println("Commands:")
	log.Println("  skills")
	log.Println("    list [--eligible] [--json]          List available skills")
	log.Println("    info <name>                         Show skill details")
	log.Println("    check                                Check all skills for issues")
	log.Println("    enable <name>                       Enable a skill")
	log.Println("    disable <name>                      Disable a skill")
	log.Println("    install <name>                       Install a skill")
	log.Println("")
	log.Println("  mcp")
	log.Println("    <name> <subcommand>                 Run MCP server")
	log.Println("")
	log.Println("  doctor                               Run diagnostics")
	log.Println("  login                                Log in to Pryx Cloud")
	log.Println("  help, -h, --help                    Show this help message")
}

func runMCPServer(args []string) int {
	if len(args) < 1 || strings.TrimSpace(args[0]) == "" {
		usage()
		return 2
	}
	name := strings.TrimSpace(args[0])

	provider, err := mcp.BundledProvider(name)
	if err != nil {
		log.Printf("unknown mcp server: %s", name)
		return 1
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := mcp.ServeStdio(ctx, provider); err != nil {
		if err == context.Canceled {
			return 0
		}
		log.Printf("mcp server error: %v", err)
		return 1
	}
	return 0
}

func runDoctor() int {
	cfg := config.Load()
	kc := keychain.New("pryx")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	rep, exitCode := doctor.Run(ctx, cfg, kc)
	for _, c := range rep.Checks {
		status := strings.ToUpper(string(c.Status))
		if c.Detail != "" {
			fmt.Printf("%-16s %s - %s\n", c.Name, status, c.Detail)
		} else {
			fmt.Printf("%-16s %s\n", c.Name, status)
		}
		if c.Suggestion != "" && (c.Status == doctor.StatusWarn || c.Status == doctor.StatusFail) {
			fmt.Printf("%-16s %s\n", "", c.Suggestion)
		}
	}
	return exitCode
}

// func runLogin() int {
// 	cfg := config.Load()
// 	kc := keychain.New("pryx")
// 	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
// 	defer cancel()

// 	fmt.Println("Attempting to log in to Pryx Cloud...")
// 	res, err := auth.StartDeviceFlow(cfg.CloudAPIUrl)
// 	if err != nil {
// 		log.Printf("\nLogin failed: %v", err)
// 		return 1
// 	}

// 	fmt.Printf("\nVerification URL: %s\n", res.VerificationURI)
// 	fmt.Printf("User Code: %s\n", res.UserCode)
// 	fmt.Println("Please open the URL above and enter the code to authorize this device.")
// 	fmt.Println("Waiting for authorization...")

// 	token, err := auth.PollForToken(cfg.CloudAPIUrl, res.DeviceCode)
// 	if err != nil {
// 		log.Printf("\nLogin failed: %v", err)
// 		return 1
// 	}

// 	if err := kc.Set("cloud_access_token", token.AccessToken) {
// 		log.Printf("\nFailed to store token: %v", err)
// 		return 1
// 	}

// 	fmt.Println("\nSuccessfully logged in!")
// 	return 0
// }
