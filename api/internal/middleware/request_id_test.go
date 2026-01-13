package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openexposuremanagement/oem/internal/middleware"
)

func TestRequestID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Context().Value("request_id")
		if reqID == nil {
			t.Fatal("request_id not in context")
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
		t.Error("X-Request-ID header not set")
	}
}
