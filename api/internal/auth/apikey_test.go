package auth

import (
	"testing"

	"github.com/openexposuremanagement/oem/internal/repository"
)

func TestHashAPIKey(t *testing.T) {
	key1 := "test-api-key"
	key2 := "test-api-key"
	key3 := "different-key"

	hash1 := hashAPIKey(key1)
	hash2 := hashAPIKey(key2)
	hash3 := hashAPIKey(key3)

	// Same keys should produce same hash
	if hash1 != hash2 {
		t.Error("Same API keys should produce same hash")
	}

	// Different keys should produce different hashes
	if hash1 == hash3 {
		t.Error("Different API keys should produce different hashes")
	}
}

func TestEnforceSourceBinding(t *testing.T) {
	tests := []struct {
		name          string
		boundSource   *string
		payloadSource string
		wantErr       bool
	}{
		{
			name:          "no binding - should allow any source",
			boundSource:   nil,
			payloadSource: "tenable",
			wantErr:       false,
		},
		{
			name:          "matching source - should allow",
			boundSource:   stringPtr("tenable"),
			payloadSource: "tenable",
			wantErr:       false,
		},
		{
			name:          "mismatched source - should reject",
			boundSource:   stringPtr("tenable"),
			payloadSource: "qualys",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := &repository.APIKey{
				BoundSource: tt.boundSource,
			}

			err := EnforceSourceBinding(key, tt.payloadSource)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnforceSourceBinding() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
