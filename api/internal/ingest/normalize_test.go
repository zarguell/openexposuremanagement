package ingest

import (
	"testing"
)

func TestNormalizeHostname(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple hostname",
			input:    "WebServer01",
			expected: "webserver01",
		},
		{
			name:     "hostname with trailing dot",
			input:    "webserver01.example.com.",
			expected: "webserver01.example.com",
		},
		{
			name:     "hostname with spaces",
			input:    "  webserver01.example.com  ",
			expected: "webserver01.example.com",
		},
		{
			name:     "fully qualified hostname",
			input:    "WebServer01.Example.COM",
			expected: "webserver01.example.com",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeHostname(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeHostname() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeShortname(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple hostname",
			input:    "webserver01",
			expected: "webserver01",
		},
		{
			name:     "fully qualified hostname",
			input:    "webserver01.example.com",
			expected: "webserver01",
		},
		{
			name:     "hostname with trailing dot",
			input:    "WebServer01.Example.COM.",
			expected: "webserver01",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeShortname(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeShortname() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidHostname(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid hostname",
			input:    "webserver01.example.com",
			expected: true,
		},
		{
			name:     "valid hostname with numbers",
			input:    "web-server-01.example.com",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "hostname with space",
			input:    "web server.example.com",
			expected: false,
		},
		{
			name:     "hostname starting with dash",
			input:    "-webserver.example.com",
			expected: false,
		},
		{
			name:     "hostname with underscore",
			input:    "web_server.example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidHostname(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidHostname() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidIPv4(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid IPv4",
			input:    "192.168.1.1",
			expected: true,
		},
		{
			name:     "valid IPv4 with spaces",
			input:    " 192.168.1.1 ",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid - not enough octets",
			input:    "192.168.1",
			expected: false,
		},
		{
			name:     "invalid - too many octets",
			input:    "192.168.1.1.1",
			expected: false,
		},
		{
			name:     "invalid - leading zeros",
			input:    "192.168.01.1",
			expected: false,
		},
		{
			name:     "invalid - out of range",
			input:    "192.168.1.256",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidIPv4(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidIPv4() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeIP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "IPv4 address",
			input:    "192.168.1.1",
			expected: "192.168.1.1",
		},
		{
			name:     "IPv6 address",
			input:    "2001:DB8::1",
			expected: "2001:db8::1",
		},
		{
			name:     "IP with spaces",
			input:    "  192.168.1.1  ",
			expected: "192.168.1.1",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeIP(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeIP() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeExternalID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "external ID",
			input:    "ABC123",
			expected: "abc123",
		},
		{
			name:     "external ID with spaces",
			input:    "  ABC123  ",
			expected: "abc123",
		},
		{
			name:     "mixed case",
			input:    "AbC-123_def",
			expected: "abc-123_def",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeExternalID(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeExternalID() = %v, want %v", result, tt.expected)
			}
		})
	}
}
