package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test with default values
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.DatabaseURL == "" {
		t.Error("DATABASE_URL should not be empty")
	}

	if cfg.Port == "" {
		t.Error("Port should not be empty")
	}

	if cfg.Environment != "development" {
		t.Errorf("Expected Environment to be 'development', got '%s'", cfg.Environment)
	}
}

func TestLoadWithEnvOverrides(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("API_PORT", "9090")
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("API_PORT")
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.DatabaseURL != "postgres://test:test@localhost:5432/test" {
		t.Errorf("Expected DATABASE_URL to be 'postgres://test:test@localhost:5432/test', got '%s'", cfg.DatabaseURL)
	}

	if cfg.Port != "9090" {
		t.Errorf("Expected Port to be '9090', got '%s'", cfg.Port)
	}

	if cfg.Environment != "production" {
		t.Errorf("Expected Environment to be 'production', got '%s'", cfg.Environment)
	}

	if cfg.LogLevel != 0 {
		t.Errorf("Expected LogLevel to be 0 (debug), got %d", cfg.LogLevel)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				DatabaseURL:  "postgres://test:test@localhost:5432/test",
				DemoMode:     true,
				DemoAPIKey:   "test-key",
				OIDCIssuer:   "",
				OIDCClientID: "",
			},
			wantErr: false,
		},
		{
			name: "missing database URL",
			cfg: &Config{
				DatabaseURL:  "",
				DemoMode:     true,
				DemoAPIKey:   "test-key",
				OIDCIssuer:   "",
				OIDCClientID: "",
			},
			wantErr: true,
		},
		{
			name: "non-demo mode missing OIDC",
			cfg: &Config{
				DatabaseURL:  "postgres://test:test@localhost:5432/test",
				DemoMode:     false,
				DemoAPIKey:   "",
				OIDCIssuer:   "",
				OIDCClientID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
