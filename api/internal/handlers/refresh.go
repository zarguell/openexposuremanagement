package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/auth"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/rs/zerolog/log"
)

// RefreshDashboard handles POST /dashboard/refresh (admin only)
func RefreshDashboard(db *sqlx.DB) http.HandlerFunc {
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

		// Check for admin role
		if !hasRole(userCtx, auth.RoleAdmin) {
			respondWithError(w, http.StatusForbidden, "Admin role required")
			return
		}

		// Refresh materialized views
		dashRepo := repository.NewDashboardRepository(db)
		err = dashRepo.RefreshMaterializedViews(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to refresh materialized views")
			respondWithError(w, http.StatusInternalServerError, "Failed to refresh dashboard")
			return
		}

		setJSONHeaders(w)
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "refreshed",
			"message": "Dashboard materialized views refreshed successfully",
		})
	}
}

// hasRole checks if user has the specified role
func hasRole(userCtx *auth.UserContext, role string) bool {
	for _, r := range userCtx.Roles {
		if r == role {
			return true
		}
	}
	return false
}
