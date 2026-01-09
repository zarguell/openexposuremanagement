package auth

import (
	"testing"
)

func TestJWTValidatorDemoMode(t *testing.T) {
	// Demo mode (no issuer URL)
	validator := NewJWTValidator("", "")

	token := "dummy-token"
	claims, err := validator.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed in demo mode: %v", err)
	}

	if claims.Subject != "demo-user" {
		t.Errorf("Expected Subject 'demo-user', got '%s'", claims.Subject)
	}

	if claims.Email != "demo@example.com" {
		t.Errorf("Expected Email 'demo@example.com', got '%s'", claims.Email)
	}
}

func TestJWTValidatorInvalidToken(t *testing.T) {
	validator := NewJWTValidator("https://example.com", "client-id")

	// Invalid token format
	_, err := validator.ValidateToken("not-a-jwt")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}
