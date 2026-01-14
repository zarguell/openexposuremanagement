package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/openexposuremanagement/oem/internal/api"
	"github.com/openexposuremanagement/oem/internal/auth"
	"github.com/openexposuremanagement/oem/internal/services/query"
	"github.com/rs/zerolog/log"
)

// QueryExecutor interface for testing and dependency injection
type QueryExecutor interface {
	Execute(ctx context.Context, tenantID string, entityType string, q *query.Query) (*query.QueryResult, error)
}

// QueryHandler handles query-related endpoints
type QueryHandler struct {
	executor QueryExecutor
}

// NewQueryHandler creates a new QueryHandler
func NewQueryHandler(executor QueryExecutor) *QueryHandler {
	return &QueryHandler{executor: executor}
}

// ServeHTTP dispatches requests to the appropriate handler
func (h *QueryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = "unknown"
	}

	path := r.URL.Path

	// Route to appropriate handler
	switch {
	case path == "/api/v1/queries/execute" && r.Method == http.MethodPost:
		h.handleExecute(w, r, requestID)
	case path == "/api/v1/queries/saved":
		switch r.Method {
		case http.MethodGet:
			h.handleListSaved(w, r, requestID)
		case http.MethodPost:
			h.handleCreateSaved(w, r, requestID)
		default:
			h.writeMethodNotAllowed(w, requestID)
		}
	case strings.HasPrefix(path, "/api/v1/queries/saved/"):
		// Extract ID from path
		idStr := strings.TrimPrefix(path, "/api/v1/queries/saved/")
		if idStr == "" {
			api.WriteErrorResponse(w, &api.QueryError{
				Code:    "INVALID_PATH",
				Message: "Invalid saved query path",
			}, requestID, http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.handleGetSaved(w, r, requestID, idStr)
		case http.MethodDelete:
			h.handleDeleteSaved(w, r, requestID, idStr)
		default:
			h.writeMethodNotAllowed(w, requestID)
		}
	default:
		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "NOT_FOUND",
			Message: "Query endpoint not found",
		}, requestID, http.StatusNotFound)
	}
}

// handleExecute executes a query against the specified entity type
func (h *QueryHandler) handleExecute(w http.ResponseWriter, r *http.Request, requestID string) {
	// Get user context from auth middleware
	userCtx := r.Context().Value(auth.UserContextKey)
	if userCtx == nil {
		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "UNAUTHORIZED",
			Message: "User context not found",
		}, requestID, http.StatusUnauthorized)
		return
	}

	user, ok := userCtx.(*auth.UserContext)
	if !ok {
		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "INVALID_CONTEXT",
			Message: "Invalid user context",
		}, requestID, http.StatusInternalServerError)
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "INVALID_REQUEST",
			Message: "Failed to read request body",
			Details: map[string]interface{}{"error": err.Error()},
		}, requestID, http.StatusBadRequest)
		return
	}

	// Parse query request
	var queryReq struct {
		EntityType string        `json:"entity_type"`
		Query      *query.Query  `json:"query"`
	}

	if err := json.Unmarshal(body, &queryReq); err != nil {
		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "INVALID_JSON",
			Message: "Invalid JSON in request body",
			Details: map[string]interface{}{"error": err.Error()},
		}, requestID, http.StatusBadRequest)
		return
	}

	// Validate entity_type
	if queryReq.EntityType == "" {
		queryReq.EntityType = "findings" // default
	}

	if queryReq.EntityType != "assets" && queryReq.EntityType != "findings" {
		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "INVALID_ENTITY_TYPE",
			Message: "entity_type must be 'assets' or 'findings'",
		}, requestID, http.StatusBadRequest)
		return
	}

	// Validate query is present
	if queryReq.Query == nil {
		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "MISSING_QUERY",
			Message: "query field is required",
		}, requestID, http.StatusBadRequest)
		return
	}

	// Convert tenantID to string for executor
	tenantID := strconv.FormatInt(user.TenantID, 10)

	// Execute query
	result, err := h.executor.Execute(r.Context(), tenantID, queryReq.EntityType, queryReq.Query)
	if err != nil {
		log.Error().Err(err).
			Str("request_id", requestID).
			Str("tenant_id", tenantID).
			Str("entity_type", queryReq.EntityType).
			Msg("query execution failed")

		// Check if it's a validation error
		errMsg := err.Error()
		if strings.HasPrefix(errMsg, "validation error:") {
			api.WriteErrorResponse(w, &api.QueryError{
				Code:    "VALIDATION_ERROR",
				Message: "Query validation failed",
				Details: map[string]interface{}{"error": errMsg},
			}, requestID, http.StatusUnprocessableEntity)
			return
		}

		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "QUERY_FAILED",
			Message: "Query execution failed",
			Details: map[string]interface{}{"error": err.Error()},
		}, requestID, http.StatusInternalServerError)
		return
	}

	// Return results
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": result.Data,
		"meta": result.Meta,
	}); err != nil {
		log.Error().Err(err).
			Str("request_id", requestID).
			Msg("failed to encode query response")
	}
}

// handleListSaved lists saved queries (stub for next task)
func (h *QueryHandler) handleListSaved(w http.ResponseWriter, r *http.Request, requestID string) {
	api.WriteErrorResponse(w, &api.QueryError{
		Code:    "NOT_IMPLEMENTED",
		Message: "Saved query listing not yet implemented",
	}, requestID, http.StatusNotImplemented)
}

// handleCreateSaved creates a new saved query (stub for next task)
func (h *QueryHandler) handleCreateSaved(w http.ResponseWriter, r *http.Request, requestID string) {
	api.WriteErrorResponse(w, &api.QueryError{
		Code:    "NOT_IMPLEMENTED",
		Message: "Saved query creation not yet implemented",
	}, requestID, http.StatusNotImplemented)
}

// handleGetSaved retrieves a saved query by ID (stub for next task)
func (h *QueryHandler) handleGetSaved(w http.ResponseWriter, r *http.Request, requestID string, id string) {
	api.WriteErrorResponse(w, &api.QueryError{
		Code:    "NOT_IMPLEMENTED",
		Message: "Saved query retrieval not yet implemented",
	}, requestID, http.StatusNotImplemented)
}

// handleDeleteSaved deletes a saved query by ID (stub for next task)
func (h *QueryHandler) handleDeleteSaved(w http.ResponseWriter, r *http.Request, requestID string, id string) {
	api.WriteErrorResponse(w, &api.QueryError{
		Code:    "NOT_IMPLEMENTED",
		Message: "Saved query deletion not yet implemented",
	}, requestID, http.StatusNotImplemented)
}

// writeMethodNotAllowed writes a 405 Method Not Allowed error
func (h *QueryHandler) writeMethodNotAllowed(w http.ResponseWriter, requestID string) {
	api.WriteErrorResponse(w, &api.QueryError{
		Code:    "METHOD_NOT_ALLOWED",
		Message: "Method not allowed",
	}, requestID, http.StatusMethodNotAllowed)
}
