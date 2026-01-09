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
							Title:       "Apache HTTP Server vulnerabilities",
							Severity:    "High",
							CVEs:        []string{"CVE-2023-1234"},
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
							Title:       "Apache HTTP Server vulnerabilities",
							Severity:    "High",
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
							Title:       "Apache HTTP Server vulnerabilities",
							Severity:    "High",
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
							Title:       "Apache HTTP Server vulnerabilities",
							Severity:    "High",
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
