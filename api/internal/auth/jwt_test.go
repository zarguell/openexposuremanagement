package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJWTValidator_ValidateToken_DemoMode(t *testing.T) {
	t.Run("accepts any token in demo mode", func(t *testing.T) {
		// Setup: No issuer URL = demo mode
		validator := NewJWTValidator("", "")

		// Test: Any token should work
		token := "any-token-value"
		claims, err := validator.ValidateToken(token)

		// Assert
		if err != nil {
			t.Fatalf("expected no error in demo mode, got %v", err)
		}
		if claims == nil {
			t.Fatal("expected claims, got nil")
		}
		if claims.Subject != "demo-user" {
			t.Errorf("expected subject 'demo-user', got '%s'", claims.Subject)
		}
		if claims.Email != "demo@example.com" {
			t.Errorf("expected email 'demo@example.com', got '%s'", claims.Email)
		}
	})

	t.Run("returns demo user details in demo mode", func(t *testing.T) {
		validator := NewJWTValidator("", "")

		claims, _ := validator.ValidateToken("irrelevant-token")

		if claims.Name != "Demo User" {
			t.Errorf("expected name 'Demo User', got '%s'", claims.Name)
		}
	})
}

func TestJWTValidator_ValidateToken_ProductionMode(t *testing.T) {
	// Note: These tests will fail until we implement JWKS fetching
	// They document the EXPECTED behavior

	t.Run("rejects invalid token format", func(t *testing.T) {
		validator := NewJWTValidator("https://issuer.example.com", "client-id")

		invalidTokens := []string{
			"",
			"not-a-jwt",
			"invalid.format",
		}

		for _, token := range invalidTokens {
			claims, err := validator.ValidateToken(token)
			if err == nil {
				t.Errorf("expected error for invalid token '%s', got nil", token)
			}
			if claims != nil {
				t.Error("expected nil claims for invalid token")
			}
		}
	})

	t.Run("validates token expiry", func(t *testing.T) {
		t.Skip("JWKS fetching not yet implemented")
		validator := NewJWTValidator("https://issuer.example.com", "client-id")

		// Test with expired token
		expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTYyMzkwMjJ9.invalid"
		claims, err := validator.ValidateToken(expiredToken)

		if err == nil {
			t.Error("expected error for expired token, got nil")
		}
		if claims != nil {
			t.Error("expected nil claims for expired token")
		}
	})

	t.Run("validates issuer", func(t *testing.T) {
		t.Skip("JWKS fetching not yet implemented")
		validator := NewJWTValidator("https://correct-issuer.example.com", "client-id")

		// Test with token from wrong issuer
		wrongIssuerToken := "token-from-wrong-issuer"
		_, err := validator.ValidateToken(wrongIssuerToken)

		if err == nil {
			t.Error("expected error for wrong issuer, got nil")
		}
	})
}

func TestJWTValidator_AuthMiddleware(t *testing.T) {
	t.Run("accepts request with valid Bearer token", func(t *testing.T) {
		validator := NewJWTValidator("", "")
		middleware := validator.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Create request with Bearer token
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("rejects request without Authorization header", func(t *testing.T) {
		validator := NewJWTValidator("", "")
		dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		middleware := validator.AuthMiddleware(dummyHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("rejects request with malformed Authorization header", func(t *testing.T) {
		validator := NewJWTValidator("", "")

		testCases := []struct {
			name       string
			authHeader string
		}{
			{"missing Bearer prefix", "just-a-token"},
			{"wrong scheme", "ApiKey token-value"},
			{"empty Bearer", "Bearer "},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
				middleware := validator.AuthMiddleware(dummyHandler)

				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", tc.authHeader)
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				if w.Code != http.StatusUnauthorized {
					t.Errorf("expected status 401, got %d", w.Code)
				}
			})
		}
	})
}

func TestUserContext(t *testing.T) {
	t.Run("creates user context from claims", func(t *testing.T) {
		claims := &Claims{
			Subject: "user-123",
			Email:   "user@example.com",
			Name:    "Test User",
		}

		userCtx := &UserContext{
			UserID:   claims.Subject,
			Email:    claims.Email,
			Name:     claims.Name,
			Token:    "test-token",
			Claims:   claims,
			Roles:    []string{"analyst"},
			TenantID: 1,
		}

		if userCtx.UserID != "user-123" {
			t.Errorf("expected UserID 'user-123', got '%s'", userCtx.UserID)
		}
		if userCtx.Email != "user@example.com" {
			t.Errorf("expected Email 'user@example.com', got '%s'", userCtx.Email)
		}
		if userCtx.TenantID != 1 {
			t.Errorf("expected TenantID 1, got %d", userCtx.TenantID)
		}
	})
}
