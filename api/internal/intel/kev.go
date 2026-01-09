package intel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// CISA KEV catalog URL
	kevDataURL = "https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json"
	kevTimeout = 2 * time.Minute
)

// KEVClient fetches CISA Known Exploited Vulnerabilities catalog
type KEVClient struct {
	DataURL string
	Timeout time.Duration
}

// NewKEVClient creates a new KEV client
func NewKEVClient() *KEVClient {
	return &KEVClient{
		DataURL: kevDataURL,
		Timeout: kevTimeout,
	}
}

// KEVCatalog represents the CISA KEV catalog structure
type KEVCatalog struct {
	Title           string     `json:"title"`
	CatalogVersion  string     `json:"catalogVersion"`
	DateReleased    string     `json:"dateReleased"`
	Count           int        `json:"count"`
	Vulnerabilities []KEVVuln  `json:"vulnerabilities"`
}

// KEVVuln represents a single vulnerability in the KEV catalog
type KEVVuln struct {
	CVEID             string `json:"cveID"`
	VendorProject     string `json:"vendorProject"`
	Product           string `json:"product"`
	VulnerabilityName string `json:"vulnerabilityName"`
	ShortDescription  string `json:"shortDescription"`
	RequiredAction    string `json:"requiredAction"`
	DueDate           string `json:"dueDate"`
	KnownRansomwareCampaignUse bool `json:"knownRansomwareCampaignUse"`
	Notes             string `json:"notes"`
}

// Fetch fetches the KEV catalog from CISA
func (c *KEVClient) Fetch(ctx context.Context) (*KEVCatalog, error) {
	log.Info().Str("url", c.DataURL).Msg("Fetching CISA KEV catalog")

	// Create HTTP client with timeout
	client := &http.Client{Timeout: c.Timeout}

	req, err := http.NewRequestWithContext(ctx, "GET", c.DataURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch KEV catalog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("KEV catalog fetch failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var catalog KEVCatalog
	if err := json.Unmarshal(body, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse KEV catalog JSON: %w", err)
	}

	log.Info().
		Int("vulnerabilities", len(catalog.Vulnerabilities)).
		Str("catalog_version", catalog.CatalogVersion).
		Msg("Successfully fetched CISA KEV catalog")

	return &catalog, nil
}

// ToEPSSRecords converts KEV vulnerabilities to EPSSRecord format for compatibility
// This allows us to use similar processing logic
func (c *KEVClient) ToEPSSRecords(catalog *KEVCatalog) []KEVRecord {
	records := make([]KEVRecord, 0, len(catalog.Vulnerabilities))

	for _, vuln := range catalog.Vulnerabilities {
		records = append(records, KEVRecord{
			CVE:        vuln.CVEID,
			DateAdded:  extractDateAdded(vuln),
			DueDate:    parseDueDate(vuln.DueDate),
			IsKEV:      true,
		})
	}

	return records
}

// KEVRecord represents a simplified KEV record
type KEVRecord struct {
	CVE       string
	DateAdded string
	DueDate   *string
	IsKEV     bool
}

// extractDateAdded attempts to extract the date added from the vulnerability data
// CISA doesn't provide an explicit date_added field, so we use the catalog release date
func extractDateAdded(vuln KEVVuln) string {
	// For now, we'll return empty - the sync layer will handle setting this
	// based on when we first saw the CVE in our database
	return ""
}

// parseDueDate parses the due date string
func parseDueDate(dueDate string) *string {
	if dueDate == "" {
		return nil
	}
	return &dueDate
}
