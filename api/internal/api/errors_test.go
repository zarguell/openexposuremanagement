package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openexposuremanagement/oem/internal/api"
)

func TestErrorResponse(t *testing.T) {
	t.Run("validation error response", func(t *testing.T) {
		err := &api.QueryError{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid query parameter",
			Details: map[string]interface{}{
				"field": "severity",
				"issue": "must be one of: critical, high, medium, low",
			},
		}

		w := httptest.NewRecorder()
		api.WriteErrorResponse(w, err, "abc-123", http.StatusBadRequest)

		if w.Code != 400 {
			t.Errorf("got status %d, want 400", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		errResp, ok := resp["error"].(map[string]interface{})
		if !ok {
			t.Fatal("response missing error object")
		}

		if errResp["code"] != "VALIDATION_ERROR" {
			t.Errorf("got code %v, want VALIDATION_ERROR", errResp["code"])
		}

		if errResp["request_id"] != "abc-123" {
			t.Errorf("got request_id %v, want abc-123", errResp["request_id"])
		}

		// Verify timestamp is present
		if errResp["timestamp"] == nil {
			t.Error("timestamp field is missing")
		}

		// Verify message
		if errResp["message"] != "Invalid query parameter" {
			t.Errorf("got message %v, want 'Invalid query parameter'", errResp["message"])
		}

		// Verify details are present
		if errResp["details"] == nil {
			t.Error("details field is missing")
		}
	})

	t.Run("error without details", func(t *testing.T) {
		err := &api.QueryError{
			Code:    "NOT_FOUND",
			Message: "Asset not found",
		}

		w := httptest.NewRecorder()
		api.WriteErrorResponse(w, err, "xyz-789", http.StatusNotFound)

		if w.Code != 404 {
			t.Errorf("got status %d, want 400", w.Code)
		}

		// Verify content type
		ct := w.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("got content-type %s, want application/json", ct)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		errResp, ok := resp["error"].(map[string]interface{})
		if !ok {
			t.Fatal("response missing error object")
		}

		if errResp["code"] != "NOT_FOUND" {
			t.Errorf("got code %v, want NOT_FOUND", errResp["code"])
		}

		// Details should be omitted when empty
		if _, exists := errResp["details"]; exists {
			t.Error("details field should be omitted when empty")
		}
	})

	t.Run("response is valid JSON", func(t *testing.T) {
		err := &api.QueryError{
			Code:    "INTERNAL_ERROR",
			Message: "Database connection failed",
			Details: map[string]interface{}{
				"retry_after": 5,
			},
		}

		w := httptest.NewRecorder()
		api.WriteErrorResponse(w, err, "req-456", http.StatusInternalServerError)

		// Verify body is valid JSON
		body := w.Body.String()
		if !strings.HasPrefix(body, "{") {
			t.Errorf("response body should be JSON object, got: %s", body)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		if _, ok := resp["error"]; !ok {
			t.Error("response missing 'error' key")
		}
	})

	t.Run("configurable status codes", func(t *testing.T) {
		tests := []struct {
			name       string
			statusCode int
			errCode    string
		}{
			{"bad request", http.StatusBadRequest, "BAD_REQUEST"},
			{"not found", http.StatusNotFound, "NOT_FOUND"},
			{"internal server error", http.StatusInternalServerError, "INTERNAL_ERROR"},
			{"unauthorized", http.StatusUnauthorized, "UNAUTHORIZED"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := &api.QueryError{
					Code:    tt.errCode,
					Message: "Test error message",
				}

				w := httptest.NewRecorder()
				api.WriteErrorResponse(w, err, "test-req-id", tt.statusCode)

				if w.Code != tt.statusCode {
					t.Errorf("got status %d, want %d", w.Code, tt.statusCode)
				}
			})
		}
	})
}
