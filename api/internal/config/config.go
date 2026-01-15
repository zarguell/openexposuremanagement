package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	DatabaseURL string
	Port        string
	Environment string
	LogLevel    int
	DemoMode    bool
	DemoAPIKey  string
	SwaggerEnabled bool

	// OIDC Configuration
	OIDCIssuer   string
	OIDCClientID string
}

// Load configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://oem:password@localhost:5432/oem?sslmode=disable"),
		Port:            getEnv("API_PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		LogLevel:        getLogLevel(getEnv("LOG_LEVEL", "info")),
		DemoMode:        getEnvBool("DEMO_MODE", false),
		DemoAPIKey:      getEnv("DEMO_API_KEY", ""),
		SwaggerEnabled:  getEnvBool("SWAGGER_ENABLED", true),
		OIDCIssuer:      getEnv("OIDC_ISSUER", ""),
		OIDCClientID:    getEnv("OIDC_CLIENT_ID", ""),
	}

	return cfg, nil
}

// Addr returns the address to listen on
func (c *Config) Addr() string {
	return ":" + c.Port
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		boolVal, err := strconv.ParseBool(value)
		if err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getLogLevel(level string) int {
	switch level {
	case "debug":
		return 0
	case "info":
		return 1
	case "warn":
		return 2
	case "error":
		return 3
	default:
		return 1
	}
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// Validate configuration
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if !c.DemoMode && c.OIDCIssuer == "" {
		return fmt.Errorf("OIDC_ISSUER is required when DEMO_MODE is false")
	}

	if !c.DemoMode && c.OIDCClientID == "" {
		return fmt.Errorf("OIDC_CLIENT_ID is required when DEMO_MODE is false")
	}

	return nil
}
