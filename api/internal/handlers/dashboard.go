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

		// Fetch dashboard data
		dashRepo := repository.NewDashboardRepository(db)
		data, err := dashRepo.GetTenantData(ctx, userCtx.TenantID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get dashboard data")
			respondWithError(w, http.StatusInternalServerError, "Failed to get dashboard data")
			return
		}

		respondWithDashboard(w, data)
	}
}

// respondWithDashboard sends the dashboard data response
func respondWithDashboard(w http.ResponseWriter, data *repository.DashboardData) {
	setJSONHeaders(w)
	respondJSON(w, http.StatusOK, data)
}
