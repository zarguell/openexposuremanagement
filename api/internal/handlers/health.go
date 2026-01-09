package handlers

import (
	"encoding/json"
	"net/http"
)

// GetMe returns the current user information (placeholder)
func GetMe(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement actual user context from JWT
	response := map[string]interface{}{
		"user":    nil,
		"tenant":  nil,
		"roles":   []string{},
		"message": "User authentication not yet implemented",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
