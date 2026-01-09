package ingest

import (
	"errors"
	"time"
)

// VMFindingsPayload represents the ingestion payload for VM findings
type VMFindingsPayload struct {
	Source    string      `json:"source" validate:"required"`
	Findings  []VMFinding `json:"findings" validate:"required,min=1"`
	ScannedAt time.Time   `json:"scanned_at" validate:"required"`
	Scanner   string      `json:"scanner"`
}

// VMFinding represents a single vulnerability finding
type VMFinding struct {
	// Asset identification
	Asset VMAsset `json:"asset" validate:"required"`
	// Finding details
	Finding VMFindingDetails `json:"finding" validate:"required"`
	// Status and timestamps
	Status     string    `json:"status" validate:"required,oneof=open fixed fixed_by_verification"`
	FirstFound time.Time `json:"first_found"`
	LastFound  time.Time `json:"last_found" validate:"required"`
	// Additional evidence
	Evidence map[string]interface{} `json:"evidence,omitempty"`
}

// VMAsset represents asset identifiers in a finding
type VMAsset struct {
	// External IDs (e.g., AWS instance ID, Azure resource ID)
	ExternalIDs map[string]string `json:"external_ids,omitempty"`
	// Hostname (primary locator)
	Hostname string `json:"hostname,omitempty"`
	// Shortname (optional, for DHCP-heavy environments)
	Shortname string `json:"shortname,omitempty"`
	// IP addresses (optional, conditional matching)
	IPAddresses []string `json:"ip_addresses,omitempty"`
	// Whether this asset has a static IP (for conditional IP matching)
	StaticIP bool `json:"static_ip,omitempty"`
}

// VMFindingDetails represents finding definition details
type VMFindingDetails struct {
	// Scanner-specific definition ID
	DefinitionID string `json:"definition_id" validate:"required"`
	// Title/name of the vulnerability
	Title string `json:"title" validate:"required"`
	// Severity level
	Severity string `json:"severity" validate:"required,oneof=Critical High Medium Low Info"`
	// CVE IDs associated with this finding
	CVEs []string `json:"cves,omitempty"`
	// References (URLs, advisories, etc.)
	References []string `json:"references,omitempty"`
	// Description (optional)
	Description string `json:"description,omitempty"`
	// Solution text (optional)
	Solution string `json:"solution,omitempty"`
}

// Valid statuses for scanner findings
var validStatuses = map[string]bool{
	"open":                  true,
	"fixed":                 true,
	"fixed_by_verification": true,
}

// Validate performs basic validation on the payload
func (p *VMFindingsPayload) Validate() error {
	// Check required fields
	if p.Source == "" {
		return ValidationError{Field: "source", Message: "source is required"}
	}

	if len(p.Findings) == 0 {
		return ValidationError{Field: "findings", Message: "at least one finding is required"}
	}

	// Validate each finding
	for i, finding := range p.Findings {
		if err := finding.Validate(); err != nil {
			return ValidationError{
				Field:   "findings",
				Message: err.Error(),
				Index:   &i,
			}
		}
	}

	return nil
}

// Validate performs validation on a single finding
func (f *VMFinding) Validate() error {
	// Validate asset
	if err := validateAsset(&f.Asset); err != nil {
		return err
	}

	// Validate finding details
	if err := validateFindingDetails(&f.Finding); err != nil {
		return err
	}

	// Validate status
	if !validStatuses[f.Status] {
		return ValidationError{
			Field:   "status",
			Message: "status must be one of: open, fixed, fixed_by_verification",
		}
	}

	return nil
}

// validateAsset validates the asset portion of a finding
func validateAsset(asset *VMAsset) error {
	if asset.Hostname == "" && len(asset.ExternalIDs) == 0 {
		return ValidationError{
			Field:   "asset",
			Message: "asset must have at least hostname or external_id",
		}
	}
	return nil
}

// validateFindingDetails validates the finding details
func validateFindingDetails(details *VMFindingDetails) error {
	if details.DefinitionID == "" {
		return ValidationError{
			Field:   "finding.definition_id",
			Message: "definition_id is required",
		}
	}

	if details.Title == "" {
		return ValidationError{
			Field:   "finding.title",
			Message: "title is required",
		}
	}

	return nil
}

// ErrValidation is the base error for validation failures
var ErrValidation = errors.New("validation error")

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Index   *int
}

func (e ValidationError) Error() string {
	if e.Index != nil {
		return e.Message
	}
	return e.Field + ": " + e.Message
}

// Unwrap returns the underlying error for error chaining
func (e ValidationError) Unwrap() error {
	return ErrValidation
}
