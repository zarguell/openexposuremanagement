package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/config"
	_ "github.com/openexposuremanagement/oem/docs" // Import docs for swagger
	"github.com/openexposuremanagement/oem/internal/handlers"
	"github.com/openexposuremanagement/oem/internal/services/query"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger/v2"
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
	// Health check endpoints (no auth required)
	s.router.HandleFunc("/healthz", s.handleHealthz)
	s.router.HandleFunc("/healthz/live", s.handleLive)
	s.router.HandleFunc("/healthz/ready", s.handleReady)
	s.router.HandleFunc("/readyz", s.handleReadyz) // Deprecated: use /healthz/ready

	// API v1 routes
	apiV1 := http.NewServeMux()

	// User endpoints
	apiV1.HandleFunc("/me", handlers.RequireAuth(handlers.GetMe))

	// Ingestion endpoints (may not need auth in demo mode)
	apiV1.HandleFunc("/ingest/vm/findings", handlers.RequireAuth(handlers.IngestVMFindings(s.db)))

	// Asset endpoints
	apiV1.HandleFunc("/assets", handlers.RequireAuth(handlers.ListAssets(s.db)))
	// Software for asset endpoint (must be before /assets/ to avoid pattern conflicts)
	apiV1.HandleFunc("/assets/{id}/software", handlers.RequireAuth(handlers.GetSoftwareForAsset(s.db)))
	apiV1.HandleFunc("/assets/", handlers.RequireAuth(handlers.GetAssetByID(s.db)))

	// Findings endpoints
	apiV1.HandleFunc("/findings", handlers.RequireAuth(handlers.ListFindings(s.db)))

	// Software endpoints
	apiV1.HandleFunc("/software", handlers.RequireAuth(handlers.GetSoftwareCatalog(s.db)))
	apiV1.HandleFunc("/software/{id}", handlers.RequireAuth(handlers.GetSoftwareByID(s.db)))

	// Query endpoints (require auth)
	var queryExecutor query.QueryExecutor = query.NewExecutor(s.db)
	// Wrap executor with logging in development mode or when log level is debug
	if s.cfg.LogLevel == 0 || s.cfg.IsDevelopment() {
		queryExecutor = query.WithLogging(queryExecutor, log.Logger)
	}
	// Get translator for explain endpoint
	queryTranslator := query.NewTranslator()
	queryHandler := handlers.NewQueryHandler(queryExecutor, queryTranslator)
	apiV1.HandleFunc("/query/findings", handlers.RequireAuth(queryHandler.QueryFindings))
	apiV1.HandleFunc("/query/assets", handlers.RequireAuth(queryHandler.QueryAssets))
	apiV1.HandleFunc("/query/software_inventory", handlers.RequireAuth(queryHandler.QuerySoftwareInventory))
	apiV1.HandleFunc("/query/unified", handlers.RequireAuth(queryHandler.QueryUnified))
	apiV1.HandleFunc("/query/oql", handlers.RequireAuth(queryHandler.QueryOQL))
	apiV1.HandleFunc("/query/oql/validate", handlers.RequireAuth(queryHandler.ValidateOQL))
	apiV1.HandleFunc("/query/oql/explain", handlers.RequireAuth(queryHandler.ExplainOQL))

	// Saved query detail endpoints with name parameter (stubs for next task)
	// NOTE: Must register /query/saved/ before /query/saved to avoid pattern conflicts
	apiV1.HandleFunc("/query/saved/", handlers.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		// Extract name from path
		// After StripPrefix("/api/v1"), path becomes /query/saved/{name}
		prefix := "/query/saved/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			http.NotFound(w, r)
			return
		}
		name := strings.TrimPrefix(r.URL.Path, prefix)
		if name == "" {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case "GET":
			queryHandler.GetSavedQuery(w, r, name)
		case "DELETE":
			queryHandler.DeleteSavedQuery(w, r, name)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Saved query endpoints (stubs for next task)
	apiV1.HandleFunc("/query/saved", handlers.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			queryHandler.ListSavedQueries(w, r)
		} else if r.Method == http.MethodPost {
			queryHandler.CreateSavedQuery(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Dashboard endpoints
	apiV1.HandleFunc("/dashboard", handlers.RequireAuth(handlers.GetDashboard(s.db)))
	apiV1.HandleFunc("/dashboard/refresh", handlers.RequireAuth(handlers.RefreshDashboardViews(s.db)))

	// Intel endpoints
	apiV1.HandleFunc("/intel/status", handlers.RequireAuth(handlers.GetIntelStatus(s.db)))
	apiV1.HandleFunc("/intel/refresh", handlers.RequireRole("admin")(handlers.RefreshIntel(s.db)))

	// Mount API v1
	s.router.Handle("/api/v1/", http.StripPrefix("/api/v1", apiV1))

	// Swagger documentation (optional, enabled via config)
	if s.cfg.SwaggerEnabled {
		s.router.Handle("/swagger/", httpSwagger.WrapHandler)
		s.router.Handle("/swagger/doc.json", httpSwagger.WrapHandler)
		log.Info().Msg("Swagger documentation enabled at /swagger/")
	}
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

// handleLive handles liveness probe requests
func (s *Server) handleLive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Liveness probe - service is running
	// Always return 200 if the process is up
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"alive"}`))
}

// handleReady handles readiness probe requests
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Readiness probe - check dependencies
	if s.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not ready","error":"database not connected"}`))
		return
	}

	// Check database connection
	if err := s.db.Ping(); err != nil {
		log.Error().Err(err).Msg("Database readiness check failed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not ready","error":"database unavailable"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.statusCode).
			Dur("duration", duration).
			Msg("HTTP request")
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ListenAndServe starts the HTTP server
func (s *Server) ListenAndServe() error {
	s.httpServer = &http.Server{
		Addr:         s.cfg.Addr(),
		Handler:      loggingMiddleware(s.router),
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
