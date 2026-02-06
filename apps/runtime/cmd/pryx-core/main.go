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

	"pryx-core/internal/agent"
	"pryx-core/internal/agent/spawn"
	"pryx-core/internal/auth"
	"pryx-core/internal/bus"
	"pryx-core/internal/channels"
	channelsSlack "pryx-core/internal/channels/slack"
	"pryx-core/internal/channels/telegram"
	"pryx-core/internal/config"
	"pryx-core/internal/constraints"
	"pryx-core/internal/doctor"
	"pryx-core/internal/keychain"
	"pryx-core/internal/mesh"
	"pryx-core/internal/models"
	"pryx-core/internal/performance"
	"pryx-core/internal/server"
	"pryx-core/internal/store"
	"pryx-core/internal/telemetry"
)

// Global variables set during build time.
var (
	// Version is the current version of the application.
	Version = "1.0.0"
	// BuildDate is the date when the application was built.
	BuildDate = "unknown"
)

// main is the entry point of the pryx-core application.
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
		case "config":
			os.Exit(runConfig(os.Args[2:]))
		case "provider":
			os.Exit(runProvider(os.Args[2:]))
		case "channel":
			os.Exit(runChannel(os.Args[2:]))
		case "session":
			os.Exit(runSession(os.Args[2:]))
		case "login":
			os.Exit(runLogin())
		case "install-service":
			os.Exit(runInstallService())
		case "uninstall-service":
			os.Exit(runUninstallService())
		case "help", "-h", "--help":
			usage()
			return
		}
	}

	log.Printf("Starting pryx-core version %s (built %s)", Version, BuildDate)

	// Initialize startup profiler
	profiler := performance.NewStartupProfiler()
	defer profiler.MarkComplete()
	defer profiler.PrintReport()

	// Load configuration
	var cfg *config.Config
	if err := profiler.TimeFunc("config.load", func() error {
		cfg = config.Load()
		return nil
	}); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize store (database)
	var s *store.Store
	if err := profiler.TimeFunc("store.init", func() error {
		var err error
		s, err = store.New(cfg.DatabasePath)
		if err != nil {
			return err
		}
		if cfg.MaxMessagesPerSession > 0 {
			s.SetMaxMessages(cfg.MaxMessagesPerSession)
		}
		return nil
	}); err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer s.Close()

	var memProfiler *performance.MemoryProfiler
	if cfg.EnableMemoryProfiling {
		memProfiler = performance.NewMemoryProfiler()
		memProfiler.SetLimits(performance.DefaultMemoryLimits)
		memProfiler.SetCallbacks(
			func(usage performance.MemorySnapshot, limit performance.MemoryLimit) {
				log.Printf("⚡ Memory warning: %d bytes allocated (%.1f%%)",
					usage.AllocBytes,
					float64(usage.AllocBytes)/float64(limit.MaxAllocBytes)*100)
			},
			func(usage performance.MemorySnapshot, limit performance.MemoryLimit) {
				log.Printf("⚠ Memory critical: %d bytes allocated (%.1f%%) - running GC",
					usage.AllocBytes,
					float64(usage.AllocBytes)/float64(limit.MaxAllocBytes)*100)
				memProfiler.ForceGC()
			},
		)
		memProfiler.StartMonitoring()
		defer memProfiler.PrintReport()
		log.Println("Memory profiling enabled")
	}

	// Initialize keychain
	var kc *keychain.Keychain
	profiler.TimeFunc("keychain.init", func() error {
		kc = keychain.New("pryx")
		return nil
	})

	// Initialize telemetry (async - non-blocking)
	profiler.StartPhase("telemetry.init")
	go func() {
		telProvider, err := telemetry.NewProvider(cfg, kc)
		if err != nil {
			log.Printf("Warning: Failed to initialize telemetry: %v", err)
		} else if telProvider.Enabled() {
			log.Printf("Telemetry enabled (device: %s)", telProvider.DeviceID())
			defer telProvider.Shutdown(context.Background())
		}
		profiler.EndPhase("telemetry.init", err)
	}()

	// Load models catalog (may be slow - load async)
	profiler.StartPhase("models.load")
	var catalog *models.Catalog
	catalogLoaded := make(chan *models.Catalog, 1)
	go func() {
		modelsService := models.NewService()
		cat, err := modelsService.Load()
		if err != nil {
			log.Printf("Warning: Failed to load models catalog: %v", err)
			profiler.EndPhase("models.load", err)
			catalogLoaded <- nil
			return
		}
		catalog = cat
		catalogLoaded <- catalog
		log.Printf("Loaded %d providers and %d models from catalog", len(catalog.Providers), len(catalog.Models))
		profiler.EndPhase("models.load", nil)
	}()

	// Initialize constraints catalog with models.dev data when catalog is loaded
	_ = constraints.FromModelsDevCatalog

	// Initialize server
	var srv *server.Server
	if err := profiler.TimeFunc("server.init", func() error {
		srv = server.New(cfg, s.DB, kc)
		// Set catalog if already loaded (rare race condition)
		if catalog != nil {
			srv.SetCatalog(catalog)
		}
		return nil
	}); err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}
	b := srv.Bus()

	schedulerCtx, schedulerCancel := context.WithCancel(context.Background())
	if err := srv.Scheduler().Start(schedulerCtx); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	defer func() {
		schedulerCancel()
		srv.Scheduler().Stop()
	}()

	// Wait for catalog to load after server starts and update it
	go func() {
		select {
		case cat := <-catalogLoaded:
			if cat != nil {
				srv.SetCatalog(cat)
				log.Printf("Catalog updated on server after async load")
			}
		case <-time.After(5 * time.Second):
			log.Printf("Catalog load timed out, continuing without it")
		}
	}()

	// Initialize mesh manager
	profiler.StartPhase("mesh.init")
	meshMgr := mesh.NewManager(cfg, b, s, kc)
	go func() {
		meshMgr.Start(context.Background())
		profiler.EndPhase("mesh.init", nil)
	}()
	defer meshMgr.Stop()

	// Initialize channels
	var chanMgr *channels.ChannelManager
	profiler.TimeFunc("channels.init", func() error {
		chanMgr = channels.NewManager(b)
		if cfg.TelegramEnabled && cfg.TelegramToken != "" {
			log.Println("Starting Telegram Bot...")
			tg := telegram.NewTelegramChannel("telegram-main", cfg.TelegramToken, b)
			if err := chanMgr.Register(tg); err != nil {
				log.Printf("Failed to register Telegram: %v", err)
			}
		}
		if cfg.SlackEnabled && cfg.SlackAppToken != "" && cfg.SlackBotToken != "" {
			log.Println("Starting Slack App...")
			slackCh := channelsSlack.NewSlackChannel("slack-main", cfg.SlackBotToken, cfg.SlackAppToken, b)
			if err := chanMgr.Register(slackCh); err != nil {
				log.Printf("Failed to register Slack: %v", err)
			}
		}
		return nil
	})
	defer chanMgr.Shutdown()

	// Initialize agent (AI Orchestrator) - heavy operation, defer if possible
	profiler.StartPhase("agent.init")
	var agt *agent.Agent
	go func() {
		var err error
		agt, err = agent.New(cfg, b, kc, catalog, srv.Skills(), srv.MCP(), srv.Agents(), srv.Memory())
		if err != nil {
			log.Printf("Warning: Failed to initialize Agent: %v", err)
			profiler.EndPhase("agent.init", err)
			return
		}
		log.Println("Starting AI Agent...")
		go agt.Run(context.Background())
		profiler.EndPhase("agent.init", nil)
	}()

	// Initialize spawner
	profiler.TimeFunc("spawner.init", func() error {
		spawner := spawn.NewSpawner(cfg, b, kc, s)
		spawnTool := spawn.NewSpawnTool(spawner, b)
		srv.SetSpawnTool(spawnTool)
		log.Println("Sub-agent spawner initialized (max agents: 10)")

		// Start cleanup goroutine
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

		// Store cleanupCancel for shutdown
		_ = cleanupCancel
		return nil
	})

	srv.Bus().Publish(bus.NewEvent(bus.EventTraceEvent, "", map[string]interface{}{
		"kind":      "runtime.started",
		"version":   Version,
		"buildDate": BuildDate,
	}))

	// Start server in background (with dynamic port allocation)
	profiler.StartPhase("server.start")
	serverErrCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			serverErrCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Start JSON-RPC bridge for host communication
	startRPCServer(context.Background(), srv)
	// Give server a moment to start, then mark phase complete
	go func() {
		time.Sleep(100 * time.Millisecond)
		profiler.EndPhase("server.start", nil)
		profiler.MarkComplete()
		profiler.PrintReport()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal or server error
	select {
	case <-stop:
		log.Println("Shutting down (received signal)...")
	case err := <-serverErrCh:
		log.Printf("Server error encountered: %v", err)
		log.Println("Shutting down (server error)...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	schedulerCancel()
	srv.Scheduler().Stop()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}

func usage() {
	log.Println("pryx-core")
	log.Println("")
	log.Println("Usage:")
	log.Println("  pryx-core")
	log.Println("  pryx-core skills <command>")
	log.Println("  pryx-core mcp <filesystem|shell|browser|clipboard>")
	log.Println("  pryx-core channel <command>")
	log.Println("  pryx-core session <command>")
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
	log.Println("    list                                List MCP servers")
	log.Println("    add <name>                          Add MCP server")
	log.Println("    remove <name>                       Remove MCP server")
	log.Println("    test <name>                         Test MCP server")
	log.Println("    auth <name>                         Manage authentication")
	log.Println("")
	log.Println("  channel")
	log.Println("    list [--json]                        List all channels")
	log.Println("    add <type> <name>                  Add a new channel")
	log.Println("    remove <name>                        Remove a channel")
	log.Println("    enable <name>                       Enable a channel")
	log.Println("    disable <name>                      Disable a channel")
	log.Println("    test <name>                         Test channel connection")
	log.Println("    status [name]                       Show channel status")
	log.Println("    sync <name>                         Sync channel configuration")
	log.Println("")
	log.Println("  session")
	log.Println("    list [--json]                       List all sessions")
	log.Println("    get <id> [--verbose]               Get session details")
	log.Println("    delete <id> [--force]               Delete a session")
	log.Println("    export <id> [--format]             Export session to file")
	log.Println("    fork <id> [--title]                Fork (copy) a session")
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
	log.Println("    oauth <provider>                     Authenticate via OAuth (Google)")
	log.Println("")
	log.Println("  doctor                               Run diagnostics")
	log.Println("  login                                Log in to Pryx Cloud")
	log.Println("  install-service                      Install as system service")
	log.Println("  uninstall-service                    Remove system service")
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

func runLogin() int {
	cfg := config.Load()
	kc := keychain.New("pryx")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Println("Attempting to log in to Pryx Cloud...")

	// Use PKCE-enabled device flow for enhanced security (RFC 7636)
	res, pkce, err := auth.StartDeviceFlowWithPKCE(cfg.CloudAPIUrl)
	if err != nil {
		log.Printf("\nLogin failed: %v", err)
		return 1
	}

	fmt.Printf("\nVerification URL: %s\n", res.VerificationURI)
	fmt.Printf("User Code: %s\n", res.UserCode)
	fmt.Println("Please open the URL above and enter the code to authorize this device.")
	fmt.Println("Waiting for authorization...")

	tokenCh := make(chan *auth.TokenResponse, 1)
	errCh := make(chan error, 1)
	go func() {
		// Use PKCE verifier when polling for token
		token, err := auth.PollForTokenWithPKCE(ctx, cfg.CloudAPIUrl, res.DeviceCode, res.Interval, pkce.CodeVerifier)
		if err != nil {
			errCh <- err
			return
		}
		tokenCh <- token
	}()

	select {
	case <-ctx.Done():
		log.Printf("\nLogin cancelled")
		return 1
	case err := <-errCh:
		log.Printf("\nLogin failed: %v", err)
		return 1
	case token := <-tokenCh:
		if err := kc.Set("cloud_access_token", token.AccessToken); err != nil {
			log.Printf("\nFailed to store token: %v", err)
			return 1
		}
	}

	fmt.Println("\nSuccessfully logged in!")
	return 0
}
