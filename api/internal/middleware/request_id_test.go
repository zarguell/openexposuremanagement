package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/openexposuremanagement/oem/internal/middleware"
)

func TestRequestID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Context().Value(middleware.RequestIDKey)
		if reqID == nil {
			t.Fatal("request_id not in context")
		}

		// Verify it's a string
		reqIDStr, ok := reqID.(string)
		if !ok {
			t.Fatal("request_id is not a string")
		}

		// Verify it's valid UUID format
		if _, err := uuid.Parse(reqIDStr); err != nil {
			t.Fatalf("request_id is not valid UUID: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	h := middleware.RequestID(handler)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("got status %d, want 200", resp.StatusCode)
	}

	reqIDHdr := resp.Header.Get("X-Request-ID")
	if reqIDHdr == "" {
		t.Fatal("X-Request-ID header not set")
	}

	// Verify it's valid UUID format
	if _, err := uuid.Parse(reqIDHdr); err != nil {
		t.Fatalf("X-Request-ID header is not valid UUID: %v", err)
	}
}

func TestRequestID_ForwardsExistingClientRequestID(t *testing.T) {
	clientReqID := "client-request-id-123"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Context().Value(middleware.RequestIDKey)
		if reqID == nil {
			t.Fatal("request_id not in context")
		}

		reqIDStr, ok := reqID.(string)
		if !ok {
			t.Fatal("request_id is not a string")
		}

		if reqIDStr != clientReqID {
			t.Errorf("got request_id %q, want %q", reqIDStr, clientReqID)
		}

		w.WriteHeader(http.StatusOK)
	})

	h := middleware.RequestID(handler)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", clientReqID)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("got status %d, want 200", resp.StatusCode)
	}

	reqIDHdr := resp.Header.Get("X-Request-ID")
	if reqIDHdr != clientReqID {
		t.Errorf("got X-Request-ID %q, want %q", reqIDHdr, clientReqID)
	}
}
