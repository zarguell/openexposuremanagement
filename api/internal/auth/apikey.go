package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/openexposuremanagement/oem/internal/repository"
)

// APIKeyAuthMiddleware creates API key authentication middleware
func APIKeyAuthMiddleware(repo *repository.APIKeyRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := extractAPIKey(r)
			if apiKey == "" {
				http.Error(w, "Missing API key", http.StatusUnauthorized)
				return
			}

			// Hash the API key
			_ = hashAPIKey(apiKey)

			// Look up API key in database
			// TODO: Fetch from repository using keyHash
			// For MVP demo mode, accept demo key
			if apiKey == "demo-key-for-development-only" {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, "Invalid API key", http.StatusUnauthorized)
		})
	}
}

// extractAPIKey extracts API key from request
func extractAPIKey(r *http.Request) string {
	// Try Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return authHeader[7:]
		}
		if strings.HasPrefix(authHeader, "ApiKey ") {
			return authHeader[7:]
		}
	}

	// Try query parameter
	return r.URL.Query().Get("api_key")
}

// hashAPIKey creates a SHA256 hash of an API key
func hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// RequireScope creates middleware that checks for required scope
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Extract API key from context and check scopes
			// For MVP demo mode, allow all
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateSourceBinding validates that the source matches the API key's bound source
func ValidateSourceBinding(repo *repository.APIKeyRepository, source string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Extract API key from context and validate bound source
			// For MVP, skip validation
			next.ServeHTTP(w, r)
		})
	}
}

// EnforceSourceBinding checks if the source in the payload matches the API key's bound source
func EnforceSourceBinding(key *repository.APIKey, payloadSource string) error {
	if key.BoundSource == nil {
		return nil // No binding, any source is allowed
	}

	if payloadSource != *key.BoundSource {
		return fmt.Errorf("source mismatch: API key bound to '%s', but payload has '%s'",
			*key.BoundSource, payloadSource)
	}

	return nil
}
