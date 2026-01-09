package intel

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	nvdAPIBaseURL = "https://services.nvd.nist.gov/rest/json/cves/2.0"
	defaultRateLimit = 5 * time.Second
)

// NVDClient fetches vulnerability data from NVD API v2.0
type NVDClient struct {
	BaseURL    string
	HTTPClient *http.Client
	RateLimit  time.Duration
	lastFetch  time.Time
	mu         sync.Mutex
}

// NewNVDClient creates a new NVD client with default settings
func NewNVDClient() *NVDClient {
	return &NVDClient{
		BaseURL:    nvdAPIBaseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		RateLimit:  defaultRateLimit,
	}
}

// CVEDetail represents detailed information about a single CVE
type CVEDetail struct {
	ID           string
	CVSSScore    *float64
	CVSSVector   string
	Descriptions []string
	References   []string
}

// FetchParams represents parameters for fetching CVEs
type FetchParams struct {
	LastModStartDate time.Time
	LastModEndDate   time.Time
	ResultsPerPage   int
	StartIndex       int
}

// NVDResponse represents the NVD API v2.0 response structure
type NVDResponse struct {
	TotalResults int            `json:"totalResults"`
	Vulnerabilities []NVDEntry `json:"vulnerabilities"`
}

// NVDEntry wraps a CVE in the NVD API response
type NVDEntry struct {
	CVE NVDCVE `json:"cve"`
}

// NVDCVE represents a vulnerability from NVD
type NVDCVE struct {
	ID           string        `json:"id"`
	Metrics      *NVDMetrics   `json:"metrics"`
	Descriptions []Description `json:"descriptions"`
	References   []Reference   `json:"references"`
}

// NVDMetrics wraps CVSS metrics in NVD format
type NVDMetrics struct {
	CVSSMetricV31 []CVSSMetricV31 `json:"cvssMetricV31"`
}

// CVSSMetricV31 represents CVSS v3.1 metrics
type CVSSMetricV31 struct {
	CVSSData CVSSData `json:"cvssData"`
}

// CVSSData contains the actual CVSS score and vector
type CVSSData struct {
	Score       float64 `json:"baseScore"`
	VectorString string `json:"vectorString"`
}

// CVSMetrics represents CVSS score data (simplified for internal use)
type CVSMetrics struct {
	Score  float64
	Vector string
}

// Description represents a CVE description
type Description struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

// Reference represents a CVE reference
type Reference struct {
	Source string `json:"source"`
	URL    string `json:"url"`
	Tags   []string `json:"tags"`
}

// FetchCVE fetches details for a specific CVE by ID
func (c *NVDClient) FetchCVE(cveID string) (*CVEDetail, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Respect rate limiting
	if !c.lastFetch.IsZero() {
		elapsed := time.Since(c.lastFetch)
		if elapsed < c.RateLimit {
			waitTime := c.RateLimit - elapsed
			time.Sleep(waitTime)
		}
	}

	url := fmt.Sprintf("%s?cveId=%s", c.BaseURL, cveID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.lastFetch = time.Now()
		return nil, fmt.Errorf("failed to fetch CVE: %w", err)
	}
	defer resp.Body.Close()

	c.lastFetch = time.Now()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, errors.New("rate limited by NVD API")
		}
		return nil, fmt.Errorf("NVD API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var nvdResp NVDResponse
	if err := json.Unmarshal(body, &nvdResp); err != nil {
		return nil, fmt.Errorf("failed to parse NVD response: %w", err)
	}

	if len(nvdResp.Vulnerabilities) == 0 {
		return nil, fmt.Errorf("CVE %s not found", cveID)
	}

	entry := nvdResp.Vulnerabilities[0]
	cve := entry.CVE
	return &CVEDetail{
		ID:           cve.ID,
		CVSSScore:    extractCVSSScore(cve.Metrics),
		CVSSVector:   extractCVSSVector(cve.Metrics),
		Descriptions: extractDescriptions(cve.Descriptions),
		References:   extractReferences(cve.References),
	}, nil
}

// Fetch fetches CVEs matching the given parameters
func (c *NVDClient) Fetch(params *FetchParams) (*NVDResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Respect rate limiting
	if !c.lastFetch.IsZero() {
		elapsed := time.Since(c.lastFetch)
		if elapsed < c.RateLimit {
			waitTime := c.RateLimit - elapsed
			time.Sleep(waitTime)
		}
	}

	url := c.buildURL(params)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.lastFetch = time.Now()
		return nil, fmt.Errorf("failed to fetch CVEs: %w", err)
	}
	defer resp.Body.Close()

	c.lastFetch = time.Now()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, errors.New("rate limited by NVD API")
		}
		return nil, fmt.Errorf("NVD API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var nvdResp NVDResponse
	if err := json.Unmarshal(body, &nvdResp); err != nil {
		return nil, fmt.Errorf("failed to parse NVD response: %w", err)
	}

	return &nvdResp, nil
}

// buildURL constructs the NVD API URL with query parameters
func (c *NVDClient) buildURL(params *FetchParams) string {
	var parts []string

	if !params.LastModStartDate.IsZero() {
		parts = append(parts, fmt.Sprintf("lastModStartDate=%s", params.LastModStartDate.Format(time.RFC3339)))
	}

	if !params.LastModEndDate.IsZero() {
		parts = append(parts, fmt.Sprintf("lastModEndDate=%s", params.LastModEndDate.Format(time.RFC3339)))
	}

	if params.ResultsPerPage > 0 {
		parts = append(parts, fmt.Sprintf("resultsPerPage=%d", params.ResultsPerPage))
	}

	if params.StartIndex > 0 {
		parts = append(parts, fmt.Sprintf("startIndex=%d", params.StartIndex))
	}

	url := c.BaseURL
	if len(parts) > 0 {
		url += "?" + strings.Join(parts, "&")
	}

	return url
}

// extractCVSSScore extracts the CVSS score from NVD metrics
func extractCVSSScore(metrics *NVDMetrics) *float64 {
	if metrics == nil || len(metrics.CVSSMetricV31) == 0 {
		return nil
	}
	score := metrics.CVSSMetricV31[0].CVSSData.Score
	if score == 0 {
		return nil
	}
	return &score
}

// extractCVSSVector extracts the CVSS vector string from NVD metrics
func extractCVSSVector(metrics *NVDMetrics) string {
	if metrics == nil || len(metrics.CVSSMetricV31) == 0 {
		return ""
	}
	return metrics.CVSSMetricV31[0].CVSSData.VectorString
}

// extractDescriptions extracts English descriptions from the list
func extractDescriptions(descriptions []Description) []string {
	var result []string
	for _, desc := range descriptions {
		if desc.Lang == "en" {
			result = append(result, desc.Value)
		}
	}
	return result
}

// extractReferences extracts URLs from references
func extractReferences(references []Reference) []string {
	var result []string
	for _, ref := range references {
		if ref.URL != "" {
			result = append(result, ref.URL)
		}
	}
	return result
}

// ConvertNVDCVEToIntelCVE converts an NVD CVE to IntelCVE format
func ConvertNVDCVEToIntelCVE(nvdCVE *NVDCVE) *IntelCVE {
	if nvdCVE == nil {
		return nil
	}

	intelCVE := &IntelCVE{
		CVE: nvdCVE.ID,
	}

	if nvdCVE.Metrics != nil {
		intelCVE.CVSSScore = extractCVSSScore(nvdCVE.Metrics)
		intelCVE.CVSSVector = extractCVSSVector(nvdCVE.Metrics)
	}

	// Join multiple descriptions with newlines
	descriptions := extractDescriptions(nvdCVE.Descriptions)
	if len(descriptions) > 0 {
		intelCVE.Description = strings.Join(descriptions, "\n\n")
	}

	return intelCVE
}

// IntelCVE represents the shape we use in our database
// This is defined here to avoid circular imports with repository package
type IntelCVE struct {
	CVE            string
	Description    string
	CVSSScore      *float64
	CVSSVector     string
	EPSSScore      *float64
	EPSSPercentile *float64
	IsKEV          bool
	KEVDateAdded   *string
	KEVDueDate     *string
}
