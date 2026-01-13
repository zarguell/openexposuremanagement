package ingest

import (
	"testing"
	"time"
)

func TestVMFindingsPayloadValidate(t *testing.T) {
	tests := []struct {
		name    string
		payload *VMFindingsPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			payload: &VMFindingsPayload{
				Source:    "tenable",
				ScannedAt: time.Now(),
				Findings: []VMFinding{
					{
						Asset: VMAsset{
							Hostname: "web-server-01.example.com",
						},
						Finding: VMFindingDetails{
							DefinitionID: "12345",
							Title:        "Apache HTTP Server vulnerabilities",
							Severity:     "High",
							CVEs:         []string{"CVE-2023-1234"},
						},
						Status:    "open",
						LastFound: time.Now(),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing source",
			payload: &VMFindingsPayload{
				ScannedAt: time.Now(),
				Findings: []VMFinding{
					{
						Asset: VMAsset{
							Hostname: "web-server-01.example.com",
						},
						Finding: VMFindingDetails{
							DefinitionID: "12345",
							Title:        "Apache HTTP Server vulnerabilities",
							Severity:     "High",
						},
						Status:    "open",
						LastFound: time.Now(),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty findings array",
			payload: &VMFindingsPayload{
				Source:    "tenable",
				ScannedAt: time.Now(),
				Findings:  []VMFinding{},
			},
			wantErr: true,
		},
		{
			name: "asset without identifiers",
			payload: &VMFindingsPayload{
				Source:    "tenable",
				ScannedAt: time.Now(),
				Findings: []VMFinding{
					{
						Asset: VMAsset{},
						Finding: VMFindingDetails{
							DefinitionID: "12345",
							Title:        "Apache HTTP Server vulnerabilities",
							Severity:     "High",
						},
						Status:    "open",
						LastFound: time.Now(),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			payload: &VMFindingsPayload{
				Source:    "tenable",
				ScannedAt: time.Now(),
				Findings: []VMFinding{
					{
						Asset: VMAsset{
							Hostname: "web-server-01.example.com",
						},
						Finding: VMFindingDetails{
							DefinitionID: "12345",
							Title:        "Apache HTTP Server vulnerabilities",
							Severity:     "High",
						},
						Status:    "invalid",
						LastFound: time.Now(),
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVMFindingsPayloadValidate_MissingScannedAt(t *testing.T) {
	payload := &VMFindingsPayload{
		Source: "tenable",
		Findings: []VMFinding{
			{
				Asset: VMAsset{
					Hostname: "web-server-01.example.com",
				},
				Finding: VMFindingDetails{
					DefinitionID: "12345",
					Title:        "Apache HTTP Server vulnerabilities",
					Severity:     "High",
				},
				Status:    "open",
				LastFound: time.Now(),
			},
		},
	}

	err := payload.Validate()
	if err == nil {
		t.Fatal("Expected error for missing scanned_at, got nil")
	}

	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "scanned_at" {
		t.Errorf("Expected field 'scanned_at', got '%s'", validationErr.Field)
	}
}

func TestVMFindingValidate_InvalidSeverity(t *testing.T) {
	finding := VMFinding{
		Asset: VMAsset{
			Hostname: "web-server-01.example.com",
		},
		Finding: VMFindingDetails{
			DefinitionID: "12345",
			Title:        "Apache HTTP Server vulnerabilities",
			Severity:     "InvalidSeverity", // Invalid severity
		},
		Status:    "open",
		LastFound: time.Now(),
	}

	err := finding.Validate()
	if err == nil {
		t.Fatal("Expected error for invalid severity, got nil")
	}

	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "finding.severity" {
		t.Errorf("Expected field 'finding.severity', got '%s'", validationErr.Field)
	}

	expectedMsg := "severity must be one of: Critical, High, Medium, Low, Info"
	if validationErr.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, validationErr.Message)
	}
}

func TestVMFindingValidate_MissingSeverity(t *testing.T) {
	finding := VMFinding{
		Asset: VMAsset{
			Hostname: "web-server-01.example.com",
		},
		Finding: VMFindingDetails{
			DefinitionID: "12345",
			Title:        "Apache HTTP Server vulnerabilities",
			Severity:     "", // Missing severity
		},
		Status:    "open",
		LastFound: time.Now(),
	}

	err := finding.Validate()
	if err == nil {
		t.Fatal("Expected error for missing severity, got nil")
	}

	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "finding.severity" {
		t.Errorf("Expected field 'finding.severity', got '%s'", validationErr.Field)
	}
}

func TestVMFindingValidate_FirstFoundAfterLastFound(t *testing.T) {
	now := time.Now()
	finding := VMFinding{
		Asset: VMAsset{
			Hostname: "web-server-01.example.com",
		},
		Finding: VMFindingDetails{
			DefinitionID: "12345",
			Title:        "Apache HTTP Server vulnerabilities",
			Severity:     "High",
		},
		Status:     "open",
		FirstFound: now.Add(24 * time.Hour), // First found AFTER last found
		LastFound:  now,
	}

	err := finding.Validate()
	if err == nil {
		t.Fatal("Expected error for first_found after last_found, got nil")
	}

	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "first_found" {
		t.Errorf("Expected field 'first_found', got '%s'", validationErr.Field)
	}
}

func TestVMAssetValidate_InvalidIPv4(t *testing.T) {
	finding := VMFinding{
		Asset: VMAsset{
			Hostname:    "web-server-01.example.com",
			IPAddresses: []string{"999.999.999.999"}, // Invalid IP
		},
		Finding: VMFindingDetails{
			DefinitionID: "12345",
			Title:        "Apache HTTP Server vulnerabilities",
			Severity:     "High",
		},
		Status:    "open",
		LastFound: time.Now(),
	}

	err := finding.Validate()
	if err == nil {
		t.Fatal("Expected error for invalid IP address, got nil")
	}

	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "asset.ip_addresses" {
		t.Errorf("Expected field 'asset.ip_addresses', got '%s'", validationErr.Field)
	}
}

func TestVMFindingValidate_ValidSeverities(t *testing.T) {
	validSeverities := []string{"Critical", "High", "Medium", "Low", "Info"}

	for _, severity := range validSeverities {
		t.Run(severity, func(t *testing.T) {
			finding := VMFinding{
				Asset: VMAsset{
					Hostname: "web-server-01.example.com",
				},
				Finding: VMFindingDetails{
					DefinitionID: "12345",
					Title:        "Apache HTTP Server vulnerabilities",
					Severity:     severity,
				},
				Status:    "open",
				LastFound: time.Now(),
			}

			err := finding.Validate()
			if err != nil {
				t.Errorf("Severity '%s' should be valid, got error: %v", severity, err)
			}
		})
	}
}

func TestVMAssetValidate_ValidIPAddresses(t *testing.T) {
	validIPs := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"8.8.8.8",
		"127.0.0.1",
	}

	finding := VMFinding{
		Asset: VMAsset{
			Hostname:    "web-server-01.example.com",
			IPAddresses: validIPs,
		},
		Finding: VMFindingDetails{
			DefinitionID: "12345",
			Title:        "Apache HTTP Server vulnerabilities",
			Severity:     "High",
		},
		Status:    "open",
		LastFound: time.Now(),
	}

	err := finding.Validate()
	if err != nil {
		t.Errorf("Valid IP addresses should pass validation, got error: %v", err)
	}
}

func TestVMFindingValidate_EmptyCVEs(t *testing.T) {
	// Empty CVEs slice should be valid (CVEs are optional)
	finding := VMFinding{
		Asset: VMAsset{
			Hostname: "web-server-01.example.com",
		},
		Finding: VMFindingDetails{
			DefinitionID: "12345",
			Title:        "Apache HTTP Server vulnerabilities",
			Severity:     "High",
			CVEs:         []string{}, // Empty CVEs
		},
		Status:    "open",
		LastFound: time.Now(),
	}

	err := finding.Validate()
	if err != nil {
		t.Errorf("Empty CVEs slice should be valid, got error: %v", err)
	}
}

func TestVMFindingValidate_ValidCVEs(t *testing.T) {
	validCVEs := []string{
		"CVE-2023-1234",
		"CVE-2023-56789",
		"CVE-2024-1",
	}

	finding := VMFinding{
		Asset: VMAsset{
			Hostname: "web-server-01.example.com",
		},
		Finding: VMFindingDetails{
			DefinitionID: "12345",
			Title:        "Apache HTTP Server vulnerabilities",
			Severity:     "High",
			CVEs:         validCVEs,
		},
		Status:    "open",
		LastFound: time.Now(),
	}

	err := finding.Validate()
	if err != nil {
		t.Errorf("Valid CVEs should pass validation, got error: %v", err)
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	err := ValidationError{
		Field:   "test_field",
		Message: "test message",
	}

	unwrapped := err.Unwrap()
	if unwrapped != ErrValidation {
		t.Errorf("Expected unwrapped error to be ErrValidation, got %v", unwrapped)
	}
}

func TestValidationError_Error_WithIndex(t *testing.T) {
	index := 5
	err := ValidationError{
		Field:   "test_field",
		Message: "test message",
		Index:   &index,
	}

	expected := "test message"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestValidationError_Error_WithoutIndex(t *testing.T) {
	err := ValidationError{
		Field:   "test_field",
		Message: "test message",
	}

	expected := "test_field: test message"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestVMFindingsPayloadValidate_WithMultipleFindings(t *testing.T) {
	payload := &VMFindingsPayload{
		Source:    "tenable",
		ScannedAt: time.Now(),
		Findings: []VMFinding{
			{
				Asset: VMAsset{
					Hostname: "web-server-01.example.com",
				},
				Finding: VMFindingDetails{
					DefinitionID: "12345",
					Title:        "Apache HTTP Server vulnerabilities",
					Severity:     "High",
				},
				Status:    "open",
				LastFound: time.Now(),
			},
			{
				Asset: VMAsset{
					Hostname: "db-server-01.example.com",
				},
				Finding: VMFindingDetails{
					DefinitionID: "67890",
					Title:        "PostgreSQL vulnerabilities",
					Severity:     "Critical",
					CVEs:         []string{"CVE-2023-1234", "CVE-2023-5678"},
				},
				Status:    "open",
				LastFound: time.Now(),
			},
		},
	}

	err := payload.Validate()
	if err != nil {
		t.Errorf("Valid payload with multiple findings should pass validation, got error: %v", err)
	}
}
