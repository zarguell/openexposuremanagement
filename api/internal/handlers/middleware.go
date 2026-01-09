package handlers

import (
	"context"
	"net/http"
	"os"

	"github.com/openexposuremanagement/oem/internal/auth"
	"github.com/rs/zerolog/log"
)

// RequireAuth wraps a handler requiring authenticated user context
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if demo mode is enabled (for MVP demonstration)
		if os.Getenv("DEMO_MODE") == "true" {
			log.Warn().Msg("🔓 DEMO MODE: Authentication bypassed for demonstration purposes")
			// Create a mock user context for demo mode
			mockCtx := &auth.UserContext{
				UserID:   "demo-user",
				TenantID: 1, // Default tenant
				Email:    "demo@example.com",
				Roles:    []string{"admin", "analyst", "viewer"},
			}
			ctx := context.WithValue(r.Context(), auth.UserContextKey, mockCtx)
			r = r.WithContext(ctx)
			next(w, r)
			return
		}

		_, err := auth.GetUserContext(r)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get user context")
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next(w, r)
	}
}

// RequireRole wraps a handler requiring specific role(s)
func RequireRole(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check if demo mode is enabled
			if os.Getenv("DEMO_MODE") == "true" {
				log.Warn().Msg("🔓 DEMO MODE: Role check bypassed for demonstration purposes")
				next(w, r)
				return
			}

			userCtx, err := auth.GetUserContext(r)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get user context")
				respondWithError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			if !hasAnyRole(userCtx, roles...) {
				respondWithError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next(w, r)
		}
	}
}

// hasAnyRole checks if user has any of the specified roles
func hasAnyRole(userCtx *auth.UserContext, roles ...string) bool {
	for _, role := range roles {
		for _, userRole := range userCtx.Roles {
			if userRole == role {
				return true
			}
		}
	}
	return false
}

// MethodsAllowed checks if the request method is in the allowed list
func MethodsAllowed(methods ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			allowed := false
			for _, method := range methods {
				if r.Method == method {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			next(w, r)
		}
	}
}
