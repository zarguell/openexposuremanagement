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

// ExecutorInterface defines the query executor interface (matches spec)
// Note: Using query.QueryExecutor interface for consistency with query service
type ExecutorInterface interface {
	Execute(ctx context.Context, tenantID string, entityType string, q *query.Query) (*query.QueryResult, error)
}

// QueryHandler handles query-related endpoints
type QueryHandler struct {
	executor ExecutorInterface
}

// NewQueryHandler creates a new QueryHandler
func NewQueryHandler(executor ExecutorInterface) *QueryHandler {
	return &QueryHandler{executor: executor}
}

// getRequestID extracts request ID from context (set by middleware)
func getRequestID(r *http.Request) string {
	reqID := r.Context().Value("request_id")
	if reqID == nil {
		return "unknown"
	}
	return reqID.(string)
}

// QueryFindings handles POST /api/v1/query/findings
func (h *QueryHandler) QueryFindings(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	h.executeQuery(w, r, requestID, "findings")
}

// QueryAssets handles POST /api/v1/query/assets
func (h *QueryHandler) QueryAssets(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	h.executeQuery(w, r, requestID, "assets")
}

// executeQuery executes a query against the specified entity type
func (h *QueryHandler) executeQuery(w http.ResponseWriter, r *http.Request, requestID string, entityType string) {
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

	// Parse query (no entity_type field in spec - it's determined by endpoint)
	var q query.Query
	if err := json.Unmarshal(body, &q); err != nil {
		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "INVALID_JSON",
			Message: "Invalid JSON in request body",
			Details: map[string]interface{}{"error": err.Error()},
		}, requestID, http.StatusBadRequest)
		return
	}

	// Note: We allow empty filters to return all data (useful for query pages)
	// The frontend can add filters as needed to narrow down results

	// Convert tenantID to string for executor
	tenantID := strconv.FormatInt(user.TenantID, 10)

	// Execute query
	result, err := h.executor.Execute(r.Context(), tenantID, entityType, &q)
	if err != nil {
		log.Error().Err(err).
			Str("request_id", requestID).
			Str("tenant_id", tenantID).
			Str("entity_type", entityType).
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

// ListSavedQueries handles GET /api/v1/query/saved (stub for next task)
func (h *QueryHandler) ListSavedQueries(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	api.WriteErrorResponse(w, &api.QueryError{
		Code:    "NOT_IMPLEMENTED",
		Message: "Saved query listing not yet implemented",
	}, requestID, http.StatusNotImplemented)
}

// CreateSavedQuery handles POST /api/v1/query/saved (stub for next task)
func (h *QueryHandler) CreateSavedQuery(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	api.WriteErrorResponse(w, &api.QueryError{
		Code:    "NOT_IMPLEMENTED",
		Message: "Saved query creation not yet implemented",
	}, requestID, http.StatusNotImplemented)
}

// GetSavedQuery handles GET /api/v1/query/saved/{name} (stub for next task)
func (h *QueryHandler) GetSavedQuery(w http.ResponseWriter, r *http.Request, name string) {
	requestID := getRequestID(r)
	api.WriteErrorResponse(w, &api.QueryError{
		Code:    "NOT_IMPLEMENTED",
		Message: "Saved query retrieval not yet implemented",
	}, requestID, http.StatusNotImplemented)
}

// DeleteSavedQuery handles DELETE /api/v1/query/saved/{name} (stub for next task)
func (h *QueryHandler) DeleteSavedQuery(w http.ResponseWriter, r *http.Request, name string) {
	requestID := getRequestID(r)
	api.WriteErrorResponse(w, &api.QueryError{
		Code:    "NOT_IMPLEMENTED",
		Message: "Saved query deletion not yet implemented",
	}, requestID, http.StatusNotImplemented)
}
