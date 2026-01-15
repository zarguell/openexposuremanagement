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
// @Summary Query findings
// @Description Query findings with filters, aggregations, and sorting
// @Tags query
// @Accept json
// @Produce json
// @Param request body query.Query true "Query object"
// @Success 200 {object} map[string]interface{} "Query results"
// @Failure 400 {object} api.QueryError "Bad request"
// @Failure 401 {object} api.QueryError "Unauthorized"
// @Failure 422 {object} api.QueryError "Validation error"
// @Failure 500 {object} api.QueryError "Internal server error"
// @Security ApiKeyAuth
// @Router /query/findings [post]
func (h *QueryHandler) QueryFindings(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	h.executeQuery(w, r, requestID, "findings")
}

// QueryAssets handles POST /api/v1/query/assets
// @Summary Query assets
// @Description Query assets with filters, aggregations, and sorting
// @Tags query
// @Accept json
// @Produce json
// @Param request body query.Query true "Query object"
// @Success 200 {object} map[string]interface{} "Query results"
// @Failure 400 {object} api.QueryError "Bad request"
// @Failure 401 {object} api.QueryError "Unauthorized"
// @Failure 422 {object} api.QueryError "Validation error"
// @Failure 500 {object} api.QueryError "Internal server error"
// @Security ApiKeyAuth
// @Router /query/assets [post]
func (h *QueryHandler) QueryAssets(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	h.executeQuery(w, r, requestID, "assets")
}

// QuerySoftwareInventory handles POST /api/v1/query/software_inventory
// @Summary Query software inventory
// @Description Query software inventory with filters, aggregations, and sorting
// @Tags query
// @Accept json
// @Produce json
// @Param request body query.Query true "Query object"
// @Success 200 {object} map[string]interface{} "Query results"
// @Failure 400 {object} api.QueryError "Bad request"
// @Failure 401 {object} api.QueryError "Unauthorized"
// @Failure 422 {object} api.QueryError "Validation error"
// @Failure 500 {object} api.QueryError "Internal server error"
// @Security ApiKeyAuth
// @Router /query/software_inventory [post]
func (h *QueryHandler) QuerySoftwareInventory(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	h.executeQuery(w, r, requestID, "software_inventory")
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

// QueryUnified handles POST /api/v1/query/unified
// @Summary Execute unified query with dot-walking syntax
// @Description Execute cross-entity correlation queries using dot-walking syntax (e.g., software.vendor, findings.severity)
// @Tags query
// @Accept json
// @Produce json
// @Param request body string true "Unified query object (JSON with filters, aggregations, sort, limit, offset)"
// @Success 200 {object} map[string]interface{} "Query results"
// @Failure 400 {object} api.QueryError "Bad request"
// @Failure 401 {object} api.QueryError "Unauthorized"
// @Failure 422 {object} api.QueryError "Validation error"
// @Failure 500 {object} api.QueryError "Internal server error"
// @Security ApiKeyAuth
// @Router /query/unified [post]
func (h *QueryHandler) QueryUnified(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)

	// Get user context
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

	// Parse unified query (simplified format - no primary_entity or join needed)
	var q query.Query
	if err := json.Unmarshal(body, &q); err != nil {
		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "INVALID_JSON",
			Message: "Invalid JSON in request body",
			Details: map[string]interface{}{"error": err.Error()},
		}, requestID, http.StatusBadRequest)
		return
	}

	// For unified queries, we always query assets with dot-walking syntax
	// The translator will auto-detect joins from field prefixes (software.*, findings.*)
	primaryEntity := "assets"

	// Execute query
	tenantID := strconv.FormatInt(user.TenantID, 10)
	result, err := h.executor.Execute(r.Context(), tenantID, primaryEntity, &q)
	if err != nil {
		log.Error().Err(err).
			Str("request_id", requestID).
			Str("tenant_id", tenantID).
			Str("primary_entity", primaryEntity).
			Msg("unified query execution failed")

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
			Msg("failed to encode unified query response")
	}
}
