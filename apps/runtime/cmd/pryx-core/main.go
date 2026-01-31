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
	"pryx-core/internal/agent/spawn"
	"pryx-core/internal/bus"
	"pryx-core/internal/channels"
	"pryx-core/internal/channels/telegram"
	"pryx-core/internal/config"
	"pryx-core/internal/doctor"
	"pryx-core/internal/keychain"
	"pryx-core/internal/mesh"
	"pryx-core/internal/models"
	"pryx-core/internal/server"
	"pryx-core/internal/store"
	"pryx-core/internal/telemetry"
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
		case "cost":
			os.Exit(runCost(os.Args[2:]))
			// 		case "login":
			// 			os.Exit(runLogin())
		case "config":
			os.Exit(runConfig(os.Args[2:]))
		case "provider":
			os.Exit(runProvider(os.Args[2:]))
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

	telProvider, err := telemetry.NewProvider(cfg, kc)
	if err != nil {
		log.Printf("Warning: Failed to initialize telemetry: %v", err)
	} else if telProvider.Enabled() {
		log.Printf("Telemetry enabled (device: %s)", telProvider.DeviceID())
		defer telProvider.Shutdown(context.Background())
	}

	modelsService := models.NewService()
	catalog, err := modelsService.Load()
	if err != nil {
		log.Printf("Warning: Failed to load models catalog: %v", err)
	} else {
		log.Printf("Loaded %d providers and %d models from catalog", len(catalog.Providers), len(catalog.Models))
	}

	srv := server.New(cfg, s.DB, kc)
	if catalog != nil {
		srv.SetCatalog(catalog)
	}
	b := srv.Bus()

	meshMgr := mesh.NewManager(cfg, b, s, kc)
	meshMgr.Start(context.Background())
	defer meshMgr.Stop()

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
	agt, err := agent.New(cfg, b, kc)
	if err != nil {
		log.Printf("Warning: Failed to initialize Agent: %v", err)
	} else {
		log.Println("Starting AI Agent...")
		go agt.Run(context.Background())
	}

	spawner := spawn.NewSpawner(cfg, b, kc)
	spawnTool := spawn.NewSpawnTool(spawner, b)
	srv.SetSpawnTool(spawnTool)
	log.Println("Sub-agent spawner initialized (max agents: 10)")

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				spawner.Cleanup(1 * time.Hour)
			case <-cleanupCtx.Done():
				return
			}
		}
	}()

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
	cleanupCancel()
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
	log.Println("  pryx-core cost <command>")
	log.Println("  pryx-core login")
	log.Println("  pryx-core config <set|get|list>")
	log.Println("  pryx-core provider <list|add|remove|use|test>")
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
	log.Println("  cost")
	log.Println("    summary                              Show total cost summary")
	log.Println("    daily [days]                         Show daily cost breakdown")
	log.Println("    monthly [months]                      Show monthly cost breakdown")
	log.Println("    budget                               Manage cost budget")
	log.Println("    pricing                              Show model pricing")
	log.Println("    optimize                             Show optimization suggestions")
	log.Println("")
	log.Println("  config")
	log.Println("    list                                 Show all configuration values")
	log.Println("    get <key>                            Get a configuration value")
	log.Println("    set <key> <value>                    Set a configuration value")
	log.Println("")
	log.Println("  provider")
	log.Println("    list                                 List all configured providers")
	log.Println("    add <name>                           Add new provider interactively")
	log.Println("    set-key <name>                       Set API key for provider")
	log.Println("    remove <name>                        Remove provider config")
	log.Println("    use <name>                           Set as active/default provider")
	log.Println("    test <name>                          Test connection to provider")
	log.Println("")
	log.Println("  doctor                               Run diagnostics")
	log.Println("  login                                Log in to Pryx Cloud")
	log.Println("  config                               Manage configuration")
	log.Println("  provider                             Manage LLM providers")
	log.Println("  help, -h, --help                    Show this help message")
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
