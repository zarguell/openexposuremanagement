package intel

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNVDClient_FetchCVE(t *testing.T) {
	t.Run("successfully fetches CVE details", func(t *testing.T) {
		server := mockNVDServer(`{
			"totalResults": 1,
			"vulnerabilities": [{
				"cve": {
					"id": "CVE-2021-44228",
					"descriptions": [{"lang": "en", "value": "Apache Log4j2 vulnerability"}],
					"metrics": {
						"cvssMetricV31": [{
							"cvssData": {
								"baseScore": 10.0,
								"vectorString": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H"
							}
						}]
					},
					"references": [{"url": "https://nvd.nist.gov/vuln/detail/CVE-2021-44228"}]
				}
			}]
		}`)
		defer server.Close()

		client := &NVDClient{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
			RateLimit:  5 * time.Second,
		}

		detail, err := client.FetchCVE("CVE-2021-44228")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if detail == nil {
			t.Fatal("expected CVE detail, got nil")
		}

		if detail.ID != "CVE-2021-44228" {
			t.Errorf("expected CVE ID CVE-2021-44228, got %s", detail.ID)
		}

		if detail.CVSSScore == nil {
			t.Error("expected CVSS score, got nil")
		} else if *detail.CVSSScore != 10.0 {
			t.Errorf("expected CVSS score 10.0, got %v", *detail.CVSSScore)
		}

		if len(detail.Descriptions) != 1 {
			t.Errorf("expected 1 description, got %d", len(detail.Descriptions))
		}
	})

	t.Run("handles rate limiting", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Rate limited"}`))
		}))
		defer server.Close()

		client := &NVDClient{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
			RateLimit:  5 * time.Second,
		}

		_, err := client.FetchCVE("CVE-2021-44228")

		if err == nil {
			t.Error("expected error for rate limit, got nil")
		}
	})

	t.Run("handles not found error", func(t *testing.T) {
		server := mockNVDServer(`{
			"totalResults": 0,
			"vulnerabilities": []
		}`)
		defer server.Close()

		client := &NVDClient{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
			RateLimit:  5 * time.Second,
		}

		_, err := client.FetchCVE("CVE-2021-44228")

		if err == nil {
			t.Error("expected error for not found, got nil")
		}
	})
}

func TestNVDClient_FetchWithPagination(t *testing.T) {
	t.Run("fetches multiple pages of results", func(t *testing.T) {
		server := mockNVDServer(`{
			"totalResults": 2,
			"vulnerabilities": [
				{
					"cve": {
						"id": "CVE-2021-44228",
						"descriptions": [{"lang": "en", "value": "Apache Log4j2 vulnerability"}],
						"metrics": {
							"cvssMetricV31": [{
								"cvssData": {
									"baseScore": 10.0,
									"vectorString": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H"
								}
							}]
						}
					}
				},
				{
					"cve": {
						"id": "CVE-2021-45046",
						"descriptions": [{"lang": "en", "value": "Another Log4j vulnerability"}],
						"metrics": {
							"cvssMetricV31": [{
								"cvssData": {
									"baseScore": 9.0,
									"vectorString": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H"
								}
							}]
						}
					}
				}
			]
		}`)
		defer server.Close()

		client := &NVDClient{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
			RateLimit:  100 * time.Millisecond,
		}

		params := &FetchParams{
			ResultsPerPage: 20,
		}

		resp, err := client.Fetch(params)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if resp == nil {
			t.Fatal("expected response, got nil")
		}

		if len(resp.Vulnerabilities) != 2 {
			t.Errorf("expected 2 vulnerabilities, got %d", len(resp.Vulnerabilities))
		}
	})

	t.Run("respects rate limit between requests", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"totalResults": 1, "vulnerabilities": []}`))
		}))
		defer server.Close()

		client := &NVDClient{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
			RateLimit:  50 * time.Millisecond,
		}

		params := &FetchParams{
			ResultsPerPage: 20,
		}

		start := time.Now()
		_, _ = client.Fetch(params)
		_, _ = client.Fetch(params)
		duration := time.Since(start)

		// Should have been called twice
		if callCount != 2 {
			t.Errorf("expected 2 calls, got %d", callCount)
		}

		// Should have taken at least the rate limit time
		if duration < client.RateLimit {
			t.Errorf("expected rate limiting of %v, but two requests took %v", client.RateLimit, duration)
		}
	})
}

func TestConvertNVDCVEToIntelCVE(t *testing.T) {
	t.Run("converts NVD CVE to IntelCVE format", func(t *testing.T) {
		nvdCVE := &NVDCVE{
			ID: "CVE-2021-44228",
			Metrics: &NVDMetrics{
				CVSSMetricV31: []CVSSMetricV31{
					{
						CVSSData: CVSSData{
							Score:       10.0,
							VectorString: "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H",
						},
					},
				},
			},
			Descriptions: []Description{
				{Lang: "en", Value: "Apache Log4j2 2.0-beta9 through 2.15.0 JNDI features..."},
			},
		}

		intelCVE := ConvertNVDCVEToIntelCVE(nvdCVE)

		if intelCVE.CVE != nvdCVE.ID {
			t.Errorf("expected CVE %s, got %s", nvdCVE.ID, intelCVE.CVE)
		}

		if intelCVE.CVSSScore == nil {
			t.Error("expected CVSS score, got nil")
		} else if *intelCVE.CVSSScore != 10.0 {
			t.Errorf("expected CVSS score 10.0, got %v", *intelCVE.CVSSScore)
		}

		if intelCVE.CVSSVector != "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H" {
			t.Errorf("unexpected vector: %s", intelCVE.CVSSVector)
		}

		if intelCVE.Description != "Apache Log4j2 2.0-beta9 through 2.15.0 JNDI features..." {
			t.Errorf("unexpected description: %s", intelCVE.Description)
		}
	})

	t.Run("handles CVE with no CVSS score", func(t *testing.T) {
		nvdCVE := &NVDCVE{
			ID:          "CVE-2021-0001",
			Metrics:     nil,
			Descriptions: []Description{{Lang: "en", Value: "A vulnerability..."}},
		}

		intelCVE := ConvertNVDCVEToIntelCVE(nvdCVE)

		if intelCVE.CVSSScore != nil {
			t.Error("expected nil CVSS score, got value")
		}

		if intelCVE.CVSSVector != "" {
			t.Error("expected empty CVSS vector, got value")
		}
	})

	t.Run("handles CVE with multiple descriptions", func(t *testing.T) {
		nvdCVE := &NVDCVE{
			ID:  "CVE-2021-0001",
			Metrics: &NVDMetrics{
				CVSSMetricV31: []CVSSMetricV31{
					{
						CVSSData: CVSSData{
							Score:       7.5,
							VectorString: "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N",
						},
					},
				},
			},
			Descriptions: []Description{
				{Lang: "en", Value: "First description of the vulnerability"},
				{Lang: "en", Value: "Second description of the vulnerability"},
			},
		}

		intelCVE := ConvertNVDCVEToIntelCVE(nvdCVE)

		// Should concatenate descriptions
		if intelCVE.Description == "" {
			t.Error("expected non-empty description")
		}

		if !strings.Contains(intelCVE.Description, "First description") {
			t.Error("description should contain first part")
		}

		if !strings.Contains(intelCVE.Description, "Second description") {
			t.Error("description should contain second part")
		}
	})

	t.Run("handles non-English descriptions", func(t *testing.T) {
		nvdCVE := &NVDCVE{
			ID:  "CVE-2021-0001",
			Metrics: nil,
			Descriptions: []Description{
				{Lang: "es", Value: "Descripción en español"},
				{Lang: "en", Value: "English description"},
			},
		}

		intelCVE := ConvertNVDCVEToIntelCVE(nvdCVE)

		// Should only include English description
		if strings.Contains(intelCVE.Description, "español") {
			t.Error("should not include non-English descriptions")
		}

		if !strings.Contains(intelCVE.Description, "English description") {
			t.Error("should include English description")
		}
	})
}

// mockNVDServer creates a test server that returns NVD-like responses
func mockNVDServer(response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	}))
}
