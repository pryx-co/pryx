package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"pryx-core/internal/config"
	"pryx-core/internal/db"
	"pryx-core/internal/keychain"
	"pryx-core/internal/server"
)

var (
	Version   = "dev"
	BuildDate = "unknown"
)

func main() {
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

	// Start server in goroutine
	go func() {
		log.Printf("Listening on %s", cfg.ListenAddr)
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down...")
	// Graceful shutdown logic would go here
}
