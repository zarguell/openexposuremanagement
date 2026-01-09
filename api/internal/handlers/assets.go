package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/auth"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/rs/zerolog/log"
)

const defaultLimit = 50
const maxLimit = 500

// ListAssets handles GET /assets
func ListAssets(db *sqlx.DB) http.HandlerFunc {
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
		query := r.URL.Query().Get("query")
		limit, err := parseLimit(r.URL.Query().Get("limit"))
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid limit parameter")
			return
		}
		offset, err := parseOffset(r.URL.Query().Get("offset"))
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid offset parameter")
			return
		}

		// Query assets
		repo := repository.NewAssetRepository(db)
		params := repository.AssetListParams{
			TenantID: userCtx.TenantID,
			Query:    query,
			Limit:    limit,
			Offset:   offset,
		}

		result, err := repo.List(ctx, params)
		if err != nil {
			log.Error().Err(err).Msg("Failed to list assets")
			respondWithError(w, http.StatusInternalServerError, "Failed to list assets")
			return
		}

		respondWithAssetList(w, result)
	}
}

// GetAssetByID handles GET /assets/{id}
func GetAssetByID(db *sqlx.DB) http.HandlerFunc {
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

		// Parse asset ID from URL
		assetID, err := parseAssetID(r.URL.Path)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid asset ID")
			return
		}

		// Get asset with details
		repo := repository.NewAssetRepository(db)
		assetDetail, err := repo.GetWithDetails(ctx, userCtx.TenantID, assetID)
		if err != nil {
			log.Error().Err(err).Int64("asset_id", assetID).Msg("Failed to get asset")
			respondWithError(w, http.StatusInternalServerError, "Failed to get asset")
			return
		}

		if assetDetail == nil {
			respondWithError(w, http.StatusNotFound, "Asset not found")
			return
		}

		respondWithAssetDetail(w, assetDetail)
	}
}

// parseLimit parses and validates the limit parameter
func parseLimit(limitStr string) (int, error) {
	if limitStr == "" {
		return defaultLimit, nil
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return 0, fmt.Errorf("invalid limit: %w", err)
	}

	if limit < 0 {
		return 0, fmt.Errorf("limit cannot be negative")
	}

	if limit > maxLimit {
		return maxLimit, nil
	}

	return limit, nil
}

// parseOffset parses and validates the offset parameter
func parseOffset(offsetStr string) (int, error) {
	if offsetStr == "" {
		return 0, nil
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return 0, fmt.Errorf("invalid offset: %w", err)
	}

	if offset < 0 {
		return 0, fmt.Errorf("offset cannot be negative")
	}

	return offset, nil
}

// parseAssetID extracts the asset ID from the URL path
func parseAssetID(path string) (int64, error) {
	// Expected path format: /api/v1/assets/{id}
	// Need to extract the ID from the path
	// For now, we'll use a simple approach
	var id int64
	_, err := fmt.Sscanf(path, "/api/v1/assets/%d", &id)
	if err != nil {
		return 0, fmt.Errorf("invalid asset ID format: %w", err)
	}
	return id, nil
}

// respondWithAssetList sends a paginated asset list response
func respondWithAssetList(w http.ResponseWriter, result *repository.AssetListResult) {
	setJSONHeaders(w)
	response := map[string]interface{}{
		"assets": result.Assets,
		"pagination": map[string]interface{}{
			"total":  result.Total,
			"limit":  result.Limit,
			"offset": result.Offset,
		},
	}
	respondJSON(w, http.StatusOK, response)
}

// respondWithAssetDetail sends a single asset detail response
func respondWithAssetDetail(w http.ResponseWriter, asset *repository.AssetDetail) {
	setJSONHeaders(w)
	respondJSON(w, http.StatusOK, asset)
}
