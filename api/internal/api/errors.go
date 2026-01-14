package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// QueryError represents a structured error response
type QueryError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	RequestID string                 `json:"request_id"`
	Timestamp string                 `json:"timestamp"`
}

// WriteErrorResponse writes a structured error response
func WriteErrorResponse(w http.ResponseWriter, err *QueryError, requestID string) {
	err.RequestID = requestID
	err.Timestamp = time.Now().Format(time.RFC3339)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err,
	})
}
