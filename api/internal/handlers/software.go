package handlers

import (
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/auth"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/rs/zerolog/log"
)

// GetSoftwareCatalog handles GET /api/v1/software
// Returns a paginated list of software with optional filters
// @Summary List software catalog
// @Description Get a paginated list of software with optional filters for vendor, product, version, and CPE
// @Tags software
// @Accept json
// @Produce json
// @Param vendor query string false "Filter by vendor"
// @Param product query string false "Filter by product name"
// @Param version query string false "Filter by version"
// @Param cpe query string false "Filter by CPE string"
// @Param limit query int false "Maximum number of results (default: 100)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Success 200 {object} map[string]interface{} "Software list with pagination"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router /software [get]
func GetSoftwareCatalog(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Only accept GET
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get user context
		userCtx, err := auth.GetUserContext(r)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Parse query parameters
		params := repository.SoftwareListParams{
			TenantID: userCtx.TenantID,
		}

		// Optional filters
		if vendor := r.URL.Query().Get("vendor"); vendor != "" {
			params.Vendor = vendor
		}

		if product := r.URL.Query().Get("product"); product != "" {
			params.Product = product
		}

		if version := r.URL.Query().Get("version"); version != "" {
			params.Version = version
		}

		if cpe := r.URL.Query().Get("cpe"); cpe != "" {
			params.CPE = cpe
		}

		// Pagination
		limitStr := r.URL.Query().Get("limit")
		params.Limit, err = parseLimit(limitStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		offsetStr := r.URL.Query().Get("offset")
		params.Offset, err = parseOffset(offsetStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Query software
		swRepo := repository.NewSoftwareRepository(db)
		software, total, err := swRepo.List(ctx, params)
		if err != nil {
			log.Error().Err(err).Msg("Failed to query software catalog")
			respondWithError(w, http.StatusInternalServerError, "Failed to query software catalog")
			return
		}

		// Respond with pagination
		respondWithSoftwareList(w, software, total, params.Limit, params.Offset)
	}
}

// respondWithSoftwareList sends a paginated software list response
func respondWithSoftwareList(w http.ResponseWriter, software []repository.SoftwareSummary, total, limit, offset int) {
	setJSONHeaders(w)
	response := map[string]interface{}{
		"software": software,
		"pagination": Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}
	respondJSON(w, http.StatusOK, response)
}

// GetSoftwareByID handles GET /api/v1/software/{id}
// Returns detailed information about a specific software including affected assets and related findings
// @Summary Get software by ID
// @Description Get detailed information about a specific software including affected assets and related findings
// @Tags software
// @Accept json
// @Produce json
// @Param id path int true "Software ID"
// @Success 200 {object} map[string]interface{} "Software details"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router /software/{id} [get]
func GetSoftwareByID(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Only accept GET
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get user context
		userCtx, err := auth.GetUserContext(r)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Parse software ID from URL
		idStr := r.PathValue("id")
		softwareID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid software ID")
			return
		}

		// Query software details
		swRepo := repository.NewSoftwareRepository(db)
		details, err := swRepo.GetSoftwareDetails(ctx, userCtx.TenantID, softwareID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to query software details")
			respondWithError(w, http.StatusInternalServerError, "Failed to query software details")
			return
		}

		respondWithSuccess(w, details)
	}
}

// GetSoftwareForAsset handles GET /api/v1/assets/{id}/software
// Returns all software installed on a specific asset
// @Summary Get software for asset
// @Description Get all software installed on a specific asset
// @Tags assets,software
// @Accept json
// @Produce json
// @Param id path int true "Asset ID"
// @Success 200 {object} map[string]interface{} "Software list for asset"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router /assets/{id}/software [get]
func GetSoftwareForAsset(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Only accept GET
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get user context
		userCtx, err := auth.GetUserContext(r)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Parse asset ID from URL
		idStr := r.PathValue("id")
		assetID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid asset ID")
			return
		}

		// Verify asset belongs to tenant
		assetRepo := repository.NewAssetRepository(db)
		asset, err := assetRepo.GetByID(ctx, assetID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Asset not found")
			return
		}

		if asset.TenantID != userCtx.TenantID {
			respondWithError(w, http.StatusForbidden, "Access denied")
			return
		}

		// Query software for asset
		swRepo := repository.NewSoftwareRepository(db)
		software, err := swRepo.GetSoftwareForAsset(ctx, userCtx.TenantID, assetID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to query software for asset")
			respondWithError(w, http.StatusInternalServerError, "Failed to query software for asset")
			return
		}

		respondWithSuccess(w, software)
	}
}
