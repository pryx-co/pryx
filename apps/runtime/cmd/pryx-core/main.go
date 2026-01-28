package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/config"
	"pryx-core/internal/db"
	"pryx-core/internal/doctor"
	"pryx-core/internal/keychain"
	"pryx-core/internal/mcp"
	"pryx-core/internal/server"
)

var (
	Version   = "dev"
	BuildDate = "unknown"
)

func main() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "mcp":
			os.Exit(runMCPServer(os.Args[2:]))
		case "doctor":
			os.Exit(runDoctor())
		case "help", "-h", "--help":
			usage()
			return
		}
	}

	log.Printf("Starting pryx-core version %s (built %s)", Version, BuildDate)

	cfg := config.Load()

	// Initialize database
	database, err := db.Init(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize keychain integration
	kc := keychain.New("pryx")

	// Initialize server
	srv := server.New(cfg, database, kc)
	srv.Bus().Publish(bus.NewEvent(bus.EventTraceEvent, "", map[string]interface{}{
		"kind":      "runtime.started",
		"version":   Version,
		"buildDate": BuildDate,
	}))

	l, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}
	actualAddr := l.Addr().String()
	host, portStr, err := net.SplitHostPort(actualAddr)
	if err != nil || strings.TrimSpace(portStr) == "" {
		fmt.Printf("PRYX_CORE_LISTEN_ADDR=%s\n", actualAddr)
	} else {
		port, _ := strconv.Atoi(portStr)
		_ = host
		fmt.Printf("PRYX_CORE_LISTEN_ADDR=http://127.0.0.1:%d\n", port)
	}

	go func() {
		log.Printf("Listening on %s", actualAddr)
		if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
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
	log.Println("  pryx-core doctor")
	log.Println("  pryx-core mcp <filesystem|shell|browser|clipboard>")
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
		return 2
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
