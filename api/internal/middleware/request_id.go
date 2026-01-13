package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// RequestIDKey is the key used to store request ID in the request context
const RequestIDKey contextKey = "request_id"

// RequestID middleware adds a unique request ID to context and response header.
// If a request ID is already present in the incoming request headers, it will be
// preserved and forwarded. Otherwise, a new UUID will be generated.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Forward existing request ID if present, otherwise generate new one
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", reqID)

		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
