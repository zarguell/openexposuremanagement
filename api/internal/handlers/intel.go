package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/auth"
	"github.com/openexposuremanagement/oem/internal/intel"
	"github.com/rs/zerolog/log"
)

// RefreshIntel triggers a full threat intelligence sync
func RefreshIntel(db *sqlx.DB) http.HandlerFunc {
	return MethodsAllowed(http.MethodPost)(RequireRole("admin")(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get user context
		userCtx, err := auth.GetUserContext(r)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get user context")
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		log.Info().
			Str("user_id", userCtx.UserID).
			Int64("tenant_id", userCtx.TenantID).
			Msg("Admin triggered threat intel refresh")

		// Create syncer and run sync
		syncer := intel.NewSyncer(db)

		// Run sync in background (don't block the request)
		go func() {
			if _, err := syncer.SyncAll(ctx); err != nil {
				log.Error().Err(err).Msg("Background intel sync failed")
			}
		}()

		// Return immediately with 202 Accepted
		setJSONHeaders(w)
		respondJSON(w, http.StatusAccepted, map[string]interface{}{
			"status":  "accepted",
			"message": "Threat intelligence sync started",
		})
	}))
}

// GetIntelStatus returns the status of all threat intel syncs
func GetIntelStatus(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Only accept GET
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get sync status
		syncer := intel.NewSyncer(db)
		status, err := syncer.GetSyncStatus(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get intel sync status")
			respondWithError(w, http.StatusInternalServerError, "Failed to get sync status")
			return
		}

		respondWithIntelStatus(w, status)
	}
}

// respondWithIntelStatus sends the intel status response
func respondWithIntelStatus(w http.ResponseWriter, status *intel.SyncStatus) {
	setJSONHeaders(w)

	// Build response with per-source status
	response := map[string]interface{}{
		"sources": status.Sources,
	}

	respondJSON(w, http.StatusOK, response)
}
