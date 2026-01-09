package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/auth"
	"github.com/openexposuremanagement/oem/internal/ingest"
	"github.com/rs/zerolog/log"
)

// IngestVMFindings handles POST /ingest/vm/findings
func IngestVMFindings(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Only accept POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Decode payload
		var payload ingest.VMFindingsPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			log.Error().Err(err).Msg("Failed to decode request body")
			respondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		// Validate payload
		err = payload.Validate()
		if err != nil {
			log.Error().Err(err).Msg("Payload validation failed")
			respondWithValidationError(w, err)
			return
		}

		// Extract tenant context from request
		userCtx, err := auth.GetUserContext(r)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get user context")
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// TODO: Validate API key has ingest:vm scope
		// TODO: Enforce source binding if API key has bound_source

		// Parse scanned_at if provided
		scannedAt := time.Now()
		if !payload.ScannedAt.IsZero() {
			scannedAt = payload.ScannedAt
		}

		// Process findings
		summary, err := processVMFindings(ctx, db, userCtx.TenantID, payload.Source, payload.Findings, scannedAt)
		if err != nil {
			log.Error().Err(err).Msg("Failed to process findings")
			respondWithError(w, http.StatusInternalServerError, "Failed to process findings")
			return
		}

		// Respond with success
		respondWithSuccess(w, summary)
	}
}

// processVMFindings processes a batch of VM findings
func processVMFindings(ctx context.Context, db *sqlx.DB, tenantID int64, source string, findings []ingest.VMFinding, scannedAt time.Time) (*IngestSummary, error) {
	summary := &IngestSummary{
		TotalFindings: len(findings),
	}

	// Process each finding
	for i, finding := range findings {
		log.Debug().
			Int("index", i).
			Str("source", source).
			Str("hostname", finding.Asset.Hostname).
			Msg("Processing finding")

		// 1. Upsert asset
		assetResult, err := ingest.UpsertAsset(ctx, db, tenantID, source, &finding.Asset, scannedAt)
		if err != nil {
			return nil, wrapFindingError(i, "asset", err)
		}

		if assetResult.NewAsset {
			summary.AssetsCreated++
		} else {
			summary.AssetsUpdated++
		}

		// 2. Upsert definition and CVE aliases
		err = ingest.UpsertDefinitionWithAliases(ctx, db, source, &finding)
		if err != nil {
			return nil, wrapFindingError(i, "definition", err)
		}

		// Generate definition UID
		definitionUID := ingest.GenerateDefinitionUID(source, finding.Finding.DefinitionID)
		summary.DefinitionsProcessed++

		// 3. Upsert finding instance
		err = ingest.UpsertFindingInstance(ctx, db, tenantID, assetResult.Asset.ID, definitionUID, &finding)
		if err != nil {
			return nil, wrapFindingError(i, "finding instance", err)
		}

		summary.FindingsUpserted++

		log.Info().
			Int64("asset_id", assetResult.Asset.ID).
			Str("definition_uid", definitionUID).
			Str("match_reason", string(assetResult.Reason)).
			Int("index", i).
			Msg("Successfully processed finding")
	}

	return summary, nil
}

// wrapFindingError creates a consistent error message for finding processing failures
func wrapFindingError(index int, operation string, err error) error {
	log.Error().
		Err(err).
		Int("index", index).
		Str("operation", operation).
		Msgf("Failed to %s", operation)
	return fmt.Errorf("failed to %s for finding %d: %w", operation, index, err)
}

// IngestSummary represents the summary of ingestion results
type IngestSummary struct {
	TotalFindings        int `json:"total_findings"`
	AssetsCreated        int `json:"assets_created"`
	AssetsUpdated        int `json:"assets_updated"`
	DefinitionsProcessed int `json:"definitions_processed"`
	FindingsUpserted     int `json:"findings_upserted"`
}

// setJSONHeaders sets common JSON response headers
func setJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

// respondJSON sends a JSON response with the given status code
func respondJSON(w http.ResponseWriter, status int, response interface{}) error {
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(response)
}

// respondWithSuccess sends a successful response
func respondWithSuccess(w http.ResponseWriter, summary *IngestSummary) {
	setJSONHeaders(w)
	response := map[string]interface{}{
		"status":  "success",
		"message": "Findings ingested successfully",
		"summary": summary,
	}
	respondJSON(w, http.StatusOK, response)
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, status int, message string) {
	setJSONHeaders(w)
	response := map[string]interface{}{
		"error": message,
	}
	respondJSON(w, status, response)
}

// respondWithValidationError sends a validation error response
func respondWithValidationError(w http.ResponseWriter, err error) {
	setJSONHeaders(w)

	// Check if it's a ValidationError
	var validationErr ingest.ValidationError
	if errors.As(err, &validationErr) {
		response := map[string]interface{}{
			"error": "Validation failed",
			"details": map[string]interface{}{
				"field":   validationErr.Field,
				"message": validationErr.Message,
				"index":   validationErr.Index,
			},
		}
		respondJSON(w, http.StatusBadRequest, response)
		return
	}

	// Generic error
	response := map[string]interface{}{
		"error":   "Validation failed",
		"message": err.Error(),
	}
	respondJSON(w, http.StatusBadRequest, response)
}
