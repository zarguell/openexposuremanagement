package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/auth"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/rs/zerolog/log"
)

// GetDashboard handles GET /dashboard
func GetDashboard(db *sqlx.DB) http.HandlerFunc {
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

		log.Info().Int64("tenant_id", userCtx.TenantID).Msg("Fetching dashboard data")

		// Fetch dashboard data
		dashRepo := repository.NewDashboardRepository(db)
		data, err := dashRepo.GetTenantData(ctx, userCtx.TenantID)
		if err != nil {
			log.Error().Err(err).Int64("tenant_id", userCtx.TenantID).Msg("Failed to get dashboard data")
			respondWithError(w, http.StatusInternalServerError, "Failed to get dashboard data")
			return
		}

		log.Info().
			Int64("tenant_id", userCtx.TenantID).
			Int("total_assets", data.Assets.TotalAssets).
			Int("active_assets", data.Assets.ActiveAssets).
			Int("open_findings", data.Findings.OpenCount).
			Int("suppressed_findings", data.Findings.SuppressedCount).
			Str("intel_sync_status", data.IntelSync.Status).
			Msg("Dashboard data retrieved successfully")

		// Log warnings for potential issues
		if data.Assets.TotalAssets == 0 {
			log.Warn().Int64("tenant_id", userCtx.TenantID).Msg("Dashboard shows 0 total assets - materialized views may not be populated")
		}
		if data.Findings.OpenCount == 0 && data.Assets.TotalAssets > 0 {
			log.Info().Int64("tenant_id", userCtx.TenantID).Msg("Dashboard shows 0 open findings - this may be expected if no vulnerabilities found")
		}

		respondWithDashboard(w, data)
	}
}

// RefreshDashboardViews handles POST /dashboard/refresh
func RefreshDashboardViews(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Only accept POST
		if r.Method != http.MethodPost {
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

		log.Info().Int64("tenant_id", userCtx.TenantID).Msg("Refreshing dashboard materialized views")

		// Refresh materialized views
		dashRepo := repository.NewDashboardRepository(db)
		err = dashRepo.RefreshMaterializedViews(ctx)
		if err != nil {
			log.Error().Err(err).Int64("tenant_id", userCtx.TenantID).Msg("Failed to refresh materialized views")
			respondWithError(w, http.StatusInternalServerError, "Failed to refresh materialized views")
			return
		}

		log.Info().Int64("tenant_id", userCtx.TenantID).Msg("Dashboard materialized views refreshed successfully")

		// Verify the refresh worked by checking the data
		data, err := dashRepo.GetTenantData(ctx, userCtx.TenantID)
		if err != nil {
			log.Warn().Err(err).Int64("tenant_id", userCtx.TenantID).Msg("Could not verify refreshed data")
		} else {
			log.Info().Int64("tenant_id", userCtx.TenantID).Int("assets", data.Assets.TotalAssets).Int("findings", data.Findings.OpenCount).Msg("Refreshed data verification")
		}

		setJSONHeaders(w)
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": "Dashboard materialized views refreshed successfully",
		})
	}
}

// respondWithDashboard sends the dashboard data response
func respondWithDashboard(w http.ResponseWriter, data *repository.DashboardData) {
	setJSONHeaders(w)
	respondJSON(w, http.StatusOK, data)
}
