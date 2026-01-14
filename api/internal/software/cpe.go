package software

import (
	"fmt"
	"regexp"
	"strings"
)

// CPEComponents represents the components of a CPE 2.3 string
type CPEComponents struct {
	Vendor   string
	Product  string
	Version  string
	Update   string
	Edition  string
	Language string
	TargetSW string
	TargetHW string
}

// Common vendor name patterns for normalization
var vendorSuffixes = []string{
	"incorporated", "inc", "llc", "llc.", "ltd", "ltd.", "corp", "corp.",
	"corporation", "systems", "technologies", "software", "foundation",
	"solutions", "group", "labs", "technologies",
}

// vendorCache is a simple cache for normalized vendor names
var vendorCache = map[string]string{
	"adobe systems incorporated":  "adobe",
	"adobe systems":               "adobe",
	"google inc":                  "google",
	"google inc.":                 "google",
	"microsoft corporation":       "microsoft",
	"microsoft":                   "microsoft",
	"oracle corporation":          "oracle",
	"apache software foundation":  "apache",
	"apache":                      "apache",
	"mozilla foundation":          "mozilla",
	"mozilla":                     "mozilla",
	"red hat":                     "redhat",
	"canonical":                   "canonical",
	"docker inc":                  "docker",
	"vmware":                      "vmware",
	"ibm":                         "ibm",
	"cisco systems":               "cisco",
	"intel corporation":           "intel",
	"amazon web services":         "amazon",
	"amazon":                      "amazon",
}

// NormalizeVendor normalizes a vendor name to a simple form
func NormalizeVendor(vendor string) string {
	if vendor == "" {
		return ""
	}

	// Trim and lowercase
	vendor = strings.TrimSpace(vendor)
	vendor = strings.ToLower(vendor)

	// Check cache first
	if cached, ok := vendorCache[vendor]; ok {
		return cached
	}

	// Remove common suffixes
	for _, suffix := range vendorSuffixes {
		if strings.HasSuffix(vendor, suffix) {
			vendor = strings.TrimSuffix(vendor, suffix)
			vendor = strings.TrimSpace(vendor)
			break
		}
	}

	// Remove trailing dots and commas
	vendor = strings.TrimSuffix(vendor, ".")
	vendor = strings.TrimSuffix(vendor, ",")

	return vendor
}

// NormalizeProduct normalizes a product name for CPE
func NormalizeProduct(product string) string {
	if product == "" {
		return ""
	}

	// Trim and lowercase
	product = strings.TrimSpace(product)
	product = strings.ToLower(product)

	// Remove version numbers from product name (common pattern: "Windows Server 2019" -> "windows_server")
	// Look for 4-digit years or version patterns at the end
	versionReg := regexp.MustCompile(`[\s_]+(20\d{2}|20\d{2}|r2|sp\d|\.?\d+(\.\d+){1,3})$`)
	product = versionReg.ReplaceAllString(product, "")

	// Replace spaces with underscores
	product = strings.ReplaceAll(product, " ", "_")

	// Replace special characters with underscore or escaped versions
	reg := regexp.MustCompile(`[^a-z0-9_\.+:-]`)
	product = reg.ReplaceAllString(product, "_")

	// Remove trailing underscore
	product = strings.TrimSuffix(product, "_")

	return product
}

// NormalizeVersion normalizes a version string
func NormalizeVersion(version string) string {
	if version == "" {
		return ""
	}

	// Trim whitespace only - keep version as-is
	version = strings.TrimSpace(version)

	return version
}

// NormalizeEdition normalizes an edition string
func NormalizeEdition(edition string) string {
	if edition == "" {
		return ""
	}

	// Trim, lowercase, and remove spaces
	edition = strings.TrimSpace(edition)
	edition = strings.ToLower(edition)
	edition = strings.ReplaceAll(edition, " ", "_")

	// Remove "edition" suffix completely
	edition = strings.ReplaceAll(edition, "_edition", "")
	edition = strings.ReplaceAll(edition, "edition", "")

	// Remove "ed" suffix
	edition = strings.ReplaceAll(edition, "_ed", "")
	edition = strings.ReplaceAll(edition, "ed", "")

	// Trim underscores again
	edition = strings.Trim(edition, "_")

	return edition
}

// BuildCPE constructs a CPE 2.3 string from components
// Format: cpe:2.3:part:vendor:product:version:update:edition:lang:sw_edition:target_sw:target_hw:other
func BuildCPE(components CPEComponents) string {
	part := "a" // Application (default for software)

	vendor := components.Vendor
	if vendor == "" {
		vendor = "*"
	}

	product := components.Product
	if product == "" {
		product = "*"
	}

	version := components.Version
	if version == "" {
		version = "*"
	}

	// In CPE, edition comes before update (sw_edition field is different)
	update := components.Edition
	if update == "" {
		update = "*"
	}

	edition := "*" // This field is not used in our simplified model

	language := components.Language
	if language == "" {
		language = "*"
	}

	swEdition := "*"
	targetSW := components.TargetSW
	if targetSW == "" {
		targetSW = "*"
	}

	targetHW := components.TargetHW
	if targetHW == "" {
		targetHW = "*"
	}

	other := "*"
	// Some CPE formats include an additional field
	other2 := "*"

	// CPE 2.3 format has 12 variable fields after "cpe:2.3:"
	// cpe:2.3:part:vendor:product:version:update:edition:lang:sw_edition:target_sw:target_hw:other:other2
	cpe := fmt.Sprintf("cpe:2.3:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s",
		part, vendor, product, version, update, edition, language, swEdition, targetSW, targetHW, other, other2)

	return cpe
}

// NormalizeToCPE normalizes software information to a CPE 2.3 string
// If an existing CPE is provided and is valid, it will be used
// Otherwise, a CPE is built from the vendor, product, version, and edition
func NormalizeToCPE(vendor, product, version, edition, existingCPE string) string {
	// If scanner provided a valid CPE, use it
	if existingCPE != "" && IsValidCPE(existingCPE) {
		return existingCPE
	}

	// Otherwise, build CPE from components
	normalizedVendor := NormalizeVendor(vendor)
	normalizedProduct := NormalizeProduct(product)
	normalizedVersion := NormalizeVersion(version)
	normalizedEdition := NormalizeEdition(edition)

	components := CPEComponents{
		Vendor:  normalizedVendor,
		Product: normalizedProduct,
		Version: normalizedVersion,
		Edition: normalizedEdition,
	}

	return BuildCPE(components)
}

// FormatTitle creates a human-readable title for software
func FormatTitle(vendor, product, version, edition string) string {
	if vendor == "" && product == "" {
		return ""
	}

	parts := []string{}

	if vendor != "" {
		parts = append(parts, strings.TrimSpace(vendor))
	}

	if product != "" {
		parts = append(parts, strings.TrimSpace(product))
	}

	if edition != "" {
		parts = append(parts, strings.TrimSpace(edition))
	}

	if version != "" {
		parts = append(parts, strings.TrimSpace(version))
	}

	return strings.Join(parts, " ")
}

// IsValidCPE performs basic validation of a CPE 2.3 string
func IsValidCPE(cpe string) bool {
	if cpe == "" {
		return false
	}

	// CPE 2.3 must start with "cpe:2.3:"
	if !strings.HasPrefix(cpe, "cpe:2.3:") {
		return false
	}

	// Must have at least 11 parts (part:vendor:product:version:update:edition:lang:sw_edition:target_sw:target_hw:other)
	parts := strings.Split(strings.TrimPrefix(cpe, "cpe:2.3:"), ":")
	if len(parts) < 11 {
		return false
	}

	// Part should be 'a' (application), 'o' (operating system), or 'h' (hardware)
	part := parts[0]
	if part != "a" && part != "o" && part != "h" {
		return false
	}

	return true
}
