package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestID middleware adds a unique request ID to context and response header
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.New().String()
		w.Header().Set("X-Request-ID", reqID)

		ctx := context.WithValue(r.Context(), "request_id", reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
