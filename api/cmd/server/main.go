package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/openexposuremanagement/oem/internal/config"
	"github.com/openexposuremanagement/oem/internal/database"
	"github.com/openexposuremanagement/oem/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// initializeDemoData creates default tenant and roles for demo mode
func initializeDemoData(db *sqlx.DB) error {
	// Check if tenant 1 exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM tenants WHERE id = 1)").Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check tenant existence: %w", err)
	}

	if !exists {
		// Create default tenant
		_, err = db.Exec("INSERT INTO tenants (id, name) VALUES (1, 'Demo Tenant') ON CONFLICT (id) DO NOTHING")
		if err != nil {
			return fmt.Errorf("failed to create demo tenant: %w", err)
		}

		// Create default roles
		_, err = db.Exec(`
			INSERT INTO roles (tenant_id, name) VALUES
			(1, 'admin'),
			(1, 'analyst'),
			(1, 'viewer')
			ON CONFLICT (tenant_id, name) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("failed to create demo roles: %w", err)
		}

		log.Info().Msg("Demo tenant and roles initialized")
	}

	return nil
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).Level(zerolog.Level(cfg.LogLevel)).With().Timestamp().Logger()

	log.Info().Msg("Open Exposure Management API - Starting...")
	log.Info().Str("version", "0.1.0").Msg("Version")
	log.Info().Str("environment", cfg.Environment).Msg("Environment")

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	log.Info().Msg("Database connection established")

	// Initialize demo mode if enabled
	if os.Getenv("DEMO_MODE") == "true" {
		log.Warn().Msg("🔓 DEMO MODE: Initializing demo tenant and roles")
		if err := initializeDemoData(db); err != nil {
			log.Error().Err(err).Msg("Failed to initialize demo data")
			os.Exit(1)
		}
	}

	// Create server
	srv := server.New(cfg, db)

	// Start server in goroutine
	go func() {
		log.Info().Str("port", cfg.Port).Msg("Starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown
	if err := srv.Shutdown(); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
