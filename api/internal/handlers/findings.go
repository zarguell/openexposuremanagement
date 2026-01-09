package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/auth"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/rs/zerolog/log"
)

// ListFindings handles GET /findings
func ListFindings(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Only accept GET
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract user context
		userCtx, err := auth.GetUserContext(r)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get user context")
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Parse query parameters
		query := r.URL.Query()
		source := query.Get("source")
		severity := query.Get("severity")
		effectiveStatus := query.Get("effective_status")
		cve := query.Get("cve")
		assetName := query.Get("asset")
		includeIntel := query.Get("include_intel") != "" // Default to true for MVP

		limit, err := parseLimit(query.Get("limit"))
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid limit parameter")
			return
		}

		offset, err := parseOffset(query.Get("offset"))
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid offset parameter")
			return
		}

		// Validate effective_status if provided
		if effectiveStatus != "" {
			validStatuses := map[string]bool{
				"open":  true,
				"fixed": true,
			}
			if !validStatuses[effectiveStatus] {
				respondWithError(w, http.StatusBadRequest, "Invalid effective_status. Must be 'open' or 'fixed'")
				return
			}
		}

		// Query enriched findings
		repo := repository.NewFindingInstanceRepository(db)
		params := repository.FindingListParams{
			TenantID:        userCtx.TenantID,
			Source:          source,
			Severity:        severity,
			EffectiveStatus: effectiveStatus,
			CVE:             cve,
			AssetName:       assetName,
			Limit:           limit,
			Offset:          offset,
		}

		result, err := repo.ListEnriched(ctx, params, includeIntel)
		if err != nil {
			log.Error().Err(err).Msg("Failed to list findings")
			respondWithError(w, http.StatusInternalServerError, "Failed to list findings")
			return
		}

		respondWithEnrichedFindingsList(w, result)
	}
}

// respondWithEnrichedFindingsList sends a paginated enriched findings list response
func respondWithEnrichedFindingsList(w http.ResponseWriter, result *repository.EnrichedFindingListResult) {
	setJSONHeaders(w)
	response := map[string]interface{}{
		"findings": result.Findings,
		"pagination": Pagination{
			Total:  result.Total,
			Limit:  result.Limit,
			Offset: result.Offset,
		},
	}
	respondJSON(w, http.StatusOK, response)
}
