package ingest

import (
	"strconv"
	"strings"
	"unicode"
)

// NormalizeHostname normalizes a hostname according to the rules:
// - Convert to lowercase
// - Trim whitespace
// - Remove trailing dot
func NormalizeHostname(hostname string) string {
	if hostname == "" {
		return ""
	}

	// Trim whitespace
	hostname = strings.TrimSpace(hostname)

	// Convert to lowercase
	hostname = strings.ToLower(hostname)

	// Remove trailing dot
	hostname = strings.TrimSuffix(hostname, ".")

	return hostname
}

// NormalizeShortname extracts and normalizes the shortname from a hostname:
// - Takes substring before first dot
// - Applies hostname normalization rules
func NormalizeShortname(hostname string) string {
	if hostname == "" {
		return ""
	}

	// Normalize hostname first
	hostname = NormalizeHostname(hostname)

	// Extract shortname (substring before first dot)
	parts := strings.SplitN(hostname, ".", 2)
	if len(parts) > 0 {
		return parts[0]
	}

	return hostname
}

// NormalizeIP normalizes an IP address string
func NormalizeIP(ip string) string {
	if ip == "" {
		return ""
	}

	// Trim whitespace
	ip = strings.TrimSpace(ip)

	// Convert to lowercase (for IPv6)
	ip = strings.ToLower(ip)

	return ip
}

// NormalizeExternalID normalizes an external ID
func NormalizeExternalID(id string) string {
	if id == "" {
		return ""
	}

	// Trim whitespace
	id = strings.TrimSpace(id)

	// Convert to lowercase
	id = strings.ToLower(id)

	return id
}

// IsValidHostname performs basic validation of a hostname
func IsValidHostname(hostname string) bool {
	if hostname == "" {
		return false
	}

	hostname = NormalizeHostname(hostname)

	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	// Check each label
	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}

		// Check that label starts and ends with alphanumeric
		if !unicode.IsLetter(rune(label[0])) && !unicode.IsDigit(rune(label[0])) {
			return false
		}

		if !unicode.IsLetter(rune(label[len(label)-1])) && !unicode.IsDigit(rune(label[len(label)-1])) {
			return false
		}

		// Check each character in label
		for _, r := range label {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' {
				return false
			}
		}
	}

	return true
}

// IsValidIPv4 performs basic validation of an IPv4 address
func IsValidIPv4(ip string) bool {
	ip = NormalizeIP(ip)
	if ip == "" {
		return false
	}

	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		// Check each character is a digit
		for _, r := range part {
			if !unicode.IsDigit(r) {
				return false
			}
		}

		// No leading zeros (unless it's just "0")
		if len(part) > 1 && part[0] == '0' {
			return false
		}

		// Check that the string isn't empty and isn't too long
		if len(part) == 0 || len(part) > 3 {
			return false
		}

		// Parse and check range (0-255)
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 || num > 255 {
			return false
		}
	}

	return true
}
