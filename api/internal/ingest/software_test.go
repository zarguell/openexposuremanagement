package ingest

import (
	"testing"
)

func TestInstalledSoftwareValidate(t *testing.T) {
	tests := []struct {
		name    string
		sw      InstalledSoftware
		wantErr bool
	}{
		{
			name: "valid software with vendor and product",
			sw: InstalledSoftware{
				Vendor:  "Adobe",
				Product: "Acrobat Reader",
				Version: "2023.001.20093",
			},
			wantErr: false,
		},
		{
			name: "valid software with CPE",
			sw: InstalledSoftware{
				CPE:     "cpe:2.3:a:adobe:acrobat_reader:2023.001.20093:*:*:*:*:*:*:*:*",
				Vendor:  "Adobe",
				Product: "Acrobat Reader",
			},
			wantErr: false,
		},
		{
			name: "invalid - missing vendor and product",
			sw: InstalledSoftware{
				Version: "2023.001.20093",
			},
			wantErr: true,
		},
		{
			name: "valid software with all fields",
			sw: InstalledSoftware{
				Vendor:     "Google",
				Product:    "Chrome",
				Version:    "120.0.6099.109",
				Edition:    "Stable",
				CPE:        "cpe:2.3:a:google:chrome:120.0.6099.109:*:*:*:*:*:*:*:*",
				InstallPath: "C:\\Program Files\\Google\\Chrome\\",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sw.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("InstalledSoftware.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVMAssetValidateWithSoftware(t *testing.T) {
	tests := []struct {
		name    string
		asset   VMAsset
		wantErr bool
	}{
		{
			name: "valid asset without software",
			asset: VMAsset{
				Hostname: "webserver01.example.com",
			},
			wantErr: false,
		},
		{
			name: "valid asset with empty software array",
			asset: VMAsset{
				Hostname:          "webserver01.example.com",
				InstalledSoftware: []InstalledSoftware{},
			},
			wantErr: false,
		},
		{
			name: "valid asset with valid software",
			asset: VMAsset{
				Hostname: "webserver01.example.com",
				InstalledSoftware: []InstalledSoftware{
					{
						Vendor:  "Adobe",
						Product: "Acrobat Reader",
						Version: "2023.001.20093",
					},
					{
						Vendor:  "Google",
						Product: "Chrome",
						Version: "120.0.6099.109",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid asset with invalid software",
			asset: VMAsset{
				Hostname: "webserver01.example.com",
				InstalledSoftware: []InstalledSoftware{
					{
						Version: "2023.001.20093", // Missing vendor and product
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAsset(&tt.asset)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAsset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcessInstalledSoftware(t *testing.T) {
	// This will be an integration test that requires a database
	// For now, we'll skip it and implement it later
	t.Skip("TODO: Implement integration test with test database")
}

func TestGetSoftwareCPE(t *testing.T) {
	tests := []struct {
		name     string
		sw       InstalledSoftware
		expected string
	}{
		{
			name: "use existing CPE",
			sw: InstalledSoftware{
				Vendor:  "Adobe",
				Product: "Acrobat Reader",
				Version: "2023.001.20093",
				CPE:     "cpe:2.3:a:adobe:acrobat_reader:2023.001.20093:*:*:*:*:*:*:*:*",
			},
			expected: "cpe:2.3:a:adobe:acrobat_reader:2023.001.20093:*:*:*:*:*:*:*:*",
		},
		{
			name: "generate CPE from components",
			sw: InstalledSoftware{
				Vendor:  "Google",
				Product: "Chrome",
				Version: "120.0.6099.109",
			},
			expected: "cpe:2.3:a:google:chrome:120.0.6099.109:*:*:*:*:*:*:*:*",
		},
		{
			name: "generate CPE with edition",
			sw: InstalledSoftware{
				Vendor:  "Microsoft",
				Product: "SQL Server",
				Version: "2019",
				Edition: "Enterprise",
			},
			expected: "cpe:2.3:a:microsoft:sql_server:2019:enterprise:*:*:*:*:*:*:*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.sw.GetCPE()
			if result != tt.expected {
				t.Errorf("InstalledSoftware.GetCPE() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSoftwareUpsertResult(t *testing.T) {
	// Test the result structure
	result := SoftwareUpsertResult{
		TotalSoftware:      5,
		SoftwareUpserted:   3,
		SoftwareCreated:    2,
		RelationsCreated:   4,
		RelationsUpdated:   1,
		RelationsDeleted:   2,
	}

	if result.TotalSoftware != 5 {
		t.Errorf("Expected TotalSoftware to be 5, got %d", result.TotalSoftware)
	}

	if result.SoftwareUpserted != 3 {
		t.Errorf("Expected SoftwareUpserted to be 3, got %d", result.SoftwareUpserted)
	}

	if result.RelationsDeleted != 2 {
		t.Errorf("Expected RelationsDeleted to be 2, got %d", result.RelationsDeleted)
	}
}

func TestGetFormattedTitle(t *testing.T) {
	tests := []struct {
		name     string
		sw       InstalledSoftware
		expected string
	}{
		{
			name: "complete software",
			sw: InstalledSoftware{
				Vendor:     "Adobe",
				Product:    "Acrobat Reader",
				Version:    "2023.001.20093",
				Edition:    "DC",
			},
			expected: "Adobe Acrobat Reader DC 2023.001.20093",
		},
		{
			name: "software without edition",
			sw: InstalledSoftware{
				Vendor:  "Google",
				Product: "Chrome",
				Version: "120.0.6099.109",
			},
			expected: "Google Chrome 120.0.6099.109",
		},
		{
			name: "minimal software",
			sw: InstalledSoftware{
				Vendor:  "Apache",
				Product: "HTTP Server",
			},
			expected: "Apache HTTP Server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.sw.GetFormattedTitle()
			if result != tt.expected {
				t.Errorf("InstalledSoftware.GetFormattedTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}
