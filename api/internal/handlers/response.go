package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/openexposuremanagement/oem/internal/ingest"
	"github.com/rs/zerolog/log"
)

const (
	defaultLimit = 50
	maxLimit     = 500
)

// setJSONHeaders sets common JSON response headers
func setJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, status int, message string) {
	setJSONHeaders(w)
	response := map[string]interface{}{
		"error": message,
	}
	respondJSON(w, status, response)
}

// respondWithSuccess sends a success response with optional data
func respondWithSuccess(w http.ResponseWriter, data interface{}) {
	setJSONHeaders(w)
	response := map[string]interface{}{
		"status": "success",
		"data":   data,
	}
	respondJSON(w, http.StatusOK, response)
}

// respondWithValidationError sends a validation error response
func respondWithValidationError(w http.ResponseWriter, err error) {
	setJSONHeaders(w)

	// Check if it's a ValidationError
	var validationErr ingest.ValidationError
	if errors.As(err, &validationErr) {
		response := map[string]interface{}{
			"error": "Validation failed",
			"details": map[string]interface{}{
				"field":   validationErr.Field,
				"message": validationErr.Message,
				"index":   validationErr.Index,
			},
		}
		respondJSON(w, http.StatusBadRequest, response)
		return
	}

	// Generic error
	response := map[string]interface{}{
		"error":   "Validation failed",
		"message": err.Error(),
	}
	respondJSON(w, http.StatusBadRequest, response)
}

// PaginationResponse wraps a paginated response with metadata
type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// respondWithPagination sends a paginated response
func respondWithPagination(w http.ResponseWriter, data interface{}, total, limit, offset int) {
	setJSONHeaders(w)
	response := PaginationResponse{
		Data: data,
		Pagination: Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}
	respondJSON(w, http.StatusOK, response)
}

// parseLimit parses and validates the limit parameter
func parseLimit(limitStr string) (int, error) {
	if limitStr == "" {
		return defaultLimit, nil
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return 0, fmt.Errorf("invalid limit: %w", err)
	}

	if limit < 0 {
		return 0, fmt.Errorf("limit cannot be negative")
	}

	if limit > maxLimit {
		return maxLimit, nil
	}

	return limit, nil
}

// parseOffset parses and validates the offset parameter
func parseOffset(offsetStr string) (int, error) {
	if offsetStr == "" {
		return 0, nil
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return 0, fmt.Errorf("invalid offset: %w", err)
	}

	if offset < 0 {
		return 0, fmt.Errorf("offset cannot be negative")
	}

	return offset, nil
}
