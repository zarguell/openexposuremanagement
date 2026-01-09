package auth

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTValidator handles JWT token validation
type JWTValidator struct {
	issuerURL    string
	clientID     string
	publicKey    *rsa.PublicKey
	keyCache     map[string]*rsa.PublicKey
	cacheExpiry  time.Time
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(issuerURL, clientID string) *JWTValidator {
	return &JWTValidator{
		issuerURL:   issuerURL,
		clientID:    clientID,
		keyCache:    make(map[string]*rsa.PublicKey),
		cacheExpiry: time.Now(),
	}
}

// Claims represents JWT claims
type Claims struct {
	Issuer   string                 `json:"iss"`
	Subject  string                 `json:"sub"`
	Audience []string               `json:"aud"`
	Expiry   int64                  `json:"exp"`
	IssuedAt int64                  `json:"iat"`
	Email    string                 `json:"email"`
	Name     string                 `json:"name"`
	Extra    map[string]interface{} `json:"-"`
}

// ValidateToken validates a JWT token and returns the claims
func (v *JWTValidator) ValidateToken(tokenString string) (*Claims, error) {
	// For demo mode, skip validation
	if v.issuerURL == "" {
		return &Claims{
			Subject: "demo-user",
			Email:   "demo@example.com",
			Name:    "Demo User",
		}, nil
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// TODO: Fetch JWKS from issuer URL
		// For MVP, we'll implement a simple JWKS fetcher
		return nil, fmt.Errorf("JWKS fetching not yet implemented")
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}

	claims := &Claims{
		Extra: make(map[string]interface{}),
	}

	// Extract standard claims
	if iss, ok := mapClaims["iss"].(string); ok {
		claims.Issuer = iss
	}
	if sub, ok := mapClaims["sub"].(string); ok {
		claims.Subject = sub
	}
	if exp, ok := mapClaims["exp"].(float64); ok {
		claims.Expiry = int64(exp)
	}
	if iat, ok := mapClaims["iat"].(float64); ok {
		claims.IssuedAt = int64(iat)
	}
	if email, ok := mapClaims["email"].(string); ok {
		claims.Email = email
	}
	if name, ok := mapClaims["name"].(string); ok {
		claims.Name = name
	}

	// Validate issuer
	if claims.Issuer != v.issuerURL {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", v.issuerURL, claims.Issuer)
	}

	// Validate expiry
	if claims.Expiry < time.Now().Unix() {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

// UserContext represents authenticated user context
type UserContext struct {
	UserID    string
	Email     string
	Name      string
	TenantID  string
	Roles     []string
	Token     string
	Claims    *Claims
}

// AuthMiddleware creates authentication middleware
func (v *JWTValidator) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Extract Bearer token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[7:]

		// Check for empty token
		if tokenString == "" {
			http.Error(w, "Empty bearer token", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := v.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		// Create user context
		userCtx := &UserContext{
			UserID:   claims.Subject,
			Email:    claims.Email,
			Name:     claims.Name,
			Token:    tokenString,
			Claims:   claims,
			Roles:    []string{"analyst"}, // TODO: Fetch from database
			TenantID: "1",                 // TODO: Fetch from database
		}

		// Add to request context
		// TODO: Use proper context package
		_ = userCtx

		next.ServeHTTP(w, r)
	})
}

// JWKSResponse represents JWKS response from OIDC provider
type JWKSResponse struct {
	Keys []JWKSKey `json:"keys"`
}

// JWKSKey represents a single key in JWKS
type JWKSKey struct {
	Kty string `json:"kty"` // Key type
	Kid string `json:"kid"` // Key ID
	Use string `json:"use"` // Public key use parameter
	N   string `json:"n"`   // Modulus
	E   string `json:"e"`   // Exponent
}

// fetchJWKS fetches JWKS from the issuer
func (v *JWTValidator) fetchJWKS() (*JWKSResponse, error) {
	// TODO: Implement JWKS fetching
	// This should:
	// 1. Fetch from {issuerURL}/.well-known/jwks.json
	// 2. Cache the response
	// 3. Parse RSA keys
	// 4. Handle key rotation

	return nil, fmt.Errorf("JWKS fetching not yet implemented")
}
