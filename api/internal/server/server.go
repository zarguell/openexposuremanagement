package server

import (
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/config"
	"github.com/openexposuremanagement/oem/internal/handlers"
	"github.com/rs/zerolog/log"
)

// Server represents the HTTP server
type Server struct {
	cfg        *config.Config
	db         *sqlx.DB
	router     *http.ServeMux
	httpServer *http.Server
}

// New creates a new server instance
func New(cfg *config.Config, db *sqlx.DB) *Server {
	s := &Server{
		cfg:    cfg,
		db:     db,
		router: http.NewServeMux(),
	}

	s.registerRoutes()
	return s
}

// registerRoutes sets up all HTTP routes
func (s *Server) registerRoutes() {
	// Health check endpoint (no auth required)
	s.router.HandleFunc("/healthz", s.handleHealthz)
	s.router.HandleFunc("/readyz", s.handleReadyz)

	// API v1 routes
	apiV1 := http.NewServeMux()
	apiV1.HandleFunc("/me", handlers.GetMe)

	// Mount API v1
	s.router.Handle("/api/v1/", http.StripPrefix("/api/v1", apiV1))
}

// handleHealthz handles health check requests
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

// handleReadyz handles readiness check requests
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := s.db.Ping(); err != nil {
		log.Error().Err(err).Msg("Database health check failed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not ready","error":"database unavailable"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

// ListenAndServe starts the HTTP server
func (s *Server) ListenAndServe() error {
	s.httpServer = &http.Server{
		Addr:         s.cfg.Addr(),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info().Str("addr", s.httpServer.Addr).Msg("Server listening")
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(nil)
	}
	return nil
}
