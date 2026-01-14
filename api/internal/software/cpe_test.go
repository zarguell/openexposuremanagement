package software

import (
	"testing"
)

func TestNormalizeVendor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple vendor",
			input:    "Adobe",
			expected: "adobe",
		},
		{
			name:     "vendor with legal suffix",
			input:    "Adobe Systems Incorporated",
			expected: "adobe",
		},
		{
			name:     "vendor with Inc",
			input:    "Google Inc.",
			expected: "google",
		},
		{
			name:     "vendor with LLC",
			input:    "Microsoft LLC",
			expected: "microsoft",
		},
		{
			name:     "vendor with Corporation",
			input:    "Oracle Corporation",
			expected: "oracle",
		},
		{
			name:     "vendor with spaces and mixed case",
			input:    "  Apache Software Foundation  ",
			expected: "apache",
		},
		{
			name:     "vendor already normalized",
			input:    "mozilla",
			expected: "mozilla",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeVendor(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeVendor() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeProduct(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple product",
			input:    "Acrobat Reader",
			expected: "acrobat_reader",
		},
		{
			name:     "product with spaces",
			input:    "SQL Server",
			expected: "sql_server",
		},
		{
			name:     "product with special chars",
			input:    "Gimp",
			expected: "gimp",
		},
		{
			name:     "product with version",
			input:    "Windows Server 2019",
			expected: "windows_server",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeProduct(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeProduct() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple version",
			input:    "2023.001.20093",
			expected: "2023.001.20093",
		},
		{
			name:     "version with spaces",
			input:    "  120.0.6099.109  ",
			expected: "120.0.6099.109",
		},
		{
			name:     "version with build info",
			input:    "16.0.1.234",
			expected: "16.0.1.234",
		},
		{
			name:     "empty version",
			input:    "",
			expected: "",
		},
		{
			name:     "version wildcard",
			input:    "*",
			expected: "*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeVersion(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeVersion() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeEdition(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple edition",
			input:    "DC",
			expected: "dc",
		},
		{
			name:     "edition with spaces",
			input:    "Enterprise Edition",
			expected: "enterprise",
		},
		{
			name:     "empty edition",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeEdition(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeEdition() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildCPE(t *testing.T) {
	tests := []struct {
		name     string
		input    CPEComponents
		expected string
	}{
		{
			name: "complete software",
			input: CPEComponents{
				Vendor:     "adobe",
				Product:    "acrobat_reader",
				Version:    "2023.001.20093",
				Edition:    "dc",
				TargetSW:   "windows",
				TargetHW:   "x64",
				Language:   "en-us",
			},
			expected: "cpe:2.3:a:adobe:acrobat_reader:2023.001.20093:dc:*:en-us:*:windows:x64:*:*",
		},
		{
			name: "minimal software",
			input: CPEComponents{
				Vendor:  "apache",
				Product: "http_server",
			},
			expected: "cpe:2.3:a:apache:http_server:*:*:*:*:*:*:*:*:*",
		},
		{
			name: "software with version only",
			input: CPEComponents{
				Vendor:  "google",
				Product: "chrome",
				Version: "120.0.6099.109",
			},
			expected: "cpe:2.3:a:google:chrome:120.0.6099.109:*:*:*:*:*:*:*:*",
		},
		{
			name: "software with edition",
			input: CPEComponents{
				Vendor:  "microsoft",
				Product: "sql_server",
				Version: "2019",
				Edition: "enterprise",
			},
			expected: "cpe:2.3:a:microsoft:sql_server:2019:enterprise:*:*:*:*:*:*:*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildCPE(tt.input)
			if result != tt.expected {
				t.Errorf("BuildCPE() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeToCPE(t *testing.T) {
	tests := []struct {
		name           string
		vendor         string
		product        string
		version        string
		edition        string
		existingCPE    string
		expectedResult string
	}{
		{
			name:           "use existing valid CPE",
			vendor:         "Adobe",
			product:        "Acrobat Reader",
			version:        "2023.001.20093",
			edition:        "DC",
			existingCPE:    "cpe:2.3:a:adobe:acrobat_reader:2023.001.20093:dc:*:*:*:*:*:*:*:*:*",
			expectedResult: "cpe:2.3:a:adobe:acrobat_reader:2023.001.20093:dc:*:*:*:*:*:*:*:*:*",
		},
		{
			name:           "build CPE from components",
			vendor:         "Google",
			product:        "Chrome",
			version:        "120.0.6099.109",
			edition:        "",
			existingCPE:    "",
			expectedResult: "cpe:2.3:a:google:chrome:120.0.6099.109:*:*:*:*:*:*:*:*",
		},
		{
			name:           "build CPE with vendor normalization",
			vendor:         "Adobe Systems Incorporated",
			product:        "Acrobat Reader DC",
			version:        "2023.001.20093",
			edition:        "",
			existingCPE:    "",
			expectedResult: "cpe:2.3:a:adobe:acrobat_reader_dc:2023.001.20093:*:*:*:*:*:*:*:*",
		},
		{
			name:           "minimal CPE when no version",
			vendor:         "Apache",
			product:        "HTTP Server",
			version:        "",
			edition:        "",
			existingCPE:    "",
			expectedResult: "cpe:2.3:a:apache:http_server:*:*:*:*:*:*:*:*:*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeToCPE(tt.vendor, tt.product, tt.version, tt.edition, tt.existingCPE)
			if result != tt.expectedResult {
				t.Errorf("NormalizeToCPE() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestFormatTitle(t *testing.T) {
	tests := []struct {
		name     string
		vendor   string
		product  string
		version  string
		edition  string
		expected string
	}{
		{
			name:     "complete software",
			vendor:   "Adobe",
			product:  "Acrobat Reader",
			version:  "2023.001.20093",
			edition:  "DC",
			expected: "Adobe Acrobat Reader DC 2023.001.20093",
		},
		{
			name:     "software without edition",
			vendor:   "Google",
			product:  "Chrome",
			version:  "120.0.6099.109",
			edition:  "",
			expected: "Google Chrome 120.0.6099.109",
		},
		{
			name:     "software without version",
			vendor:   "Apache",
			product:  "HTTP Server",
			version:  "",
			edition:  "",
			expected: "Apache HTTP Server",
		},
		{
			name:     "empty input",
			vendor:   "",
			product:  "",
			version:  "",
			edition:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTitle(tt.vendor, tt.product, tt.version, tt.edition)
			if result != tt.expected {
				t.Errorf("FormatTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidCPE(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid CPE 2.3",
			input:    "cpe:2.3:a:adobe:acrobat_reader:2023.001.20093:*:*:*:*:*:*:*:*:*",
			expected: true,
		},
		{
			name:     "valid CPE 2.3 with all fields",
			input:    "cpe:2.3:a:adobe:acrobat_reader:2023.001.20093:dc:*:en-us:windows:x64:*:*:*:*",
			expected: true,
		},
		{
			name:     "invalid CPE - missing prefix",
			input:    "adobe:acrobat_reader:2023.001.20093",
			expected: false,
		},
		{
			name:     "invalid CPE - wrong version",
			input:    "cpe:2.2:a:adobe:acrobat_reader:2023.001.20093:*:*:*:*:*:*:*:*:*",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid CPE - too few parts",
			input:    "cpe:2.3:a:adobe",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidCPE(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidCPE() = %v, want %v", result, tt.expected)
			}
		})
	}
}
