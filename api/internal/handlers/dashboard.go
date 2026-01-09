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

		log.Info().Int64("tenant_id", userCtx.TenantID).Int("total_assets", data.Assets.TotalAssets).Int("open_findings", data.Findings.OpenCount).Msg("Dashboard data retrieved")

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
			log.Error().Err(err).Msg("Failed to refresh materialized views")
			respondWithError(w, http.StatusInternalServerError, "Failed to refresh materialized views")
			return
		}

		log.Info().Msg("Dashboard materialized views refreshed successfully")

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
