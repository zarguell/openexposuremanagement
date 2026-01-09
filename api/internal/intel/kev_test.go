package intel

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestKEVClient_Fetch(t *testing.T) {
	// Sample CISA KEV catalog JSON
	jsonData := `{
		"title": "CISA Known Exploited Vulnerabilities Catalog",
		"catalogVersion": "1.0",
		"dateReleased": "2024-01-01",
		"count": 2,
		"vulnerabilities": [
			{
				"cveID": "CVE-2021-44228",
				"vendorProject": "Apache",
				"product": "Log4j",
				"vulnerabilityName": "Apache Log4j2 Remote Code Execution",
				"shortDescription": "Apache Log4j2...",
				"requiredAction": "Apply updates",
				"dueDate": "2021-12-24",
				"knownRansomwareCampaignUse": true,
				"notes": "Log4Shell"
			},
			{
				"cveID": "CVE-2021-34527",
				"vendorProject": "Microsoft",
				"product": "Windows Print Spooler",
				"vulnerabilityName": "Windows Print Spooler RCE",
				"shortDescription": "PrintNightmare...",
				"requiredAction": "Disable service",
				"dueDate": "2021-08-23",
				"knownRansomwareCampaignUse": false,
				"notes": ""
			}
		]
	}`

	t.Run("successfully fetches and parses KEV catalog", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(jsonData))
		}))
		defer server.Close()

		client := &KEVClient{
			DataURL: server.URL,
			Timeout: 5 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		catalog, err := client.Fetch(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if catalog == nil {
			t.Fatal("expected catalog, got nil")
		}

		if catalog.Count != 2 {
			t.Errorf("expected count 2, got %d", catalog.Count)
		}

		if len(catalog.Vulnerabilities) != 2 {
			t.Errorf("expected 2 vulnerabilities, got %d", len(catalog.Vulnerabilities))
		}

		// Check first vulnerability
		if catalog.Vulnerabilities[0].CVEID != "CVE-2021-44228" {
			t.Errorf("expected CVE-2021-44228, got %s", catalog.Vulnerabilities[0].CVEID)
		}

		if !catalog.Vulnerabilities[0].KnownRansomwareCampaignUse {
			t.Error("expected KnownRansomwareCampaignUse to be true")
		}

		if catalog.Vulnerabilities[0].DueDate != "2021-12-24" {
			t.Errorf("expected due date 2021-12-24, got %s", catalog.Vulnerabilities[0].DueDate)
		}
	})

	t.Run("handles HTTP errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
		}))
		defer server.Close()

		client := &KEVClient{
			DataURL: server.URL,
			Timeout: 5 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := client.Fetch(ctx)
		if err == nil {
			t.Error("expected error for HTTP 500, got nil")
		}
	})

	t.Run("handles malformed JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{invalid json}`))
		}))
		defer server.Close()

		client := &KEVClient{
			DataURL: server.URL,
			Timeout: 5 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := client.Fetch(ctx)
		if err == nil {
			t.Error("expected error for malformed JSON, got nil")
		}
	})

	t.Run("handles context timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(jsonData))
		}))
		defer server.Close()

		client := &KEVClient{
			DataURL: server.URL,
			Timeout: 5 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := client.Fetch(ctx)
		if err == nil {
			t.Error("expected context deadline exceeded error, got nil")
		}
	})
}

func TestKEVClient_ToEPSSRecords(t *testing.T) {
	t.Run("converts KEV catalog to records", func(t *testing.T) {
		catalog := &KEVCatalog{
			Vulnerabilities: []KEVVuln{
				{
					CVEID:             "CVE-2021-44228",
					DueDate:           "2021-12-24",
				},
				{
					CVEID:             "CVE-2021-34527",
					DueDate:           "2021-08-23",
				},
			},
		}

		client := NewKEVClient()
		records := client.ToEPSSRecords(catalog)

		if len(records) != 2 {
			t.Errorf("expected 2 records, got %d", len(records))
		}

		if records[0].CVE != "CVE-2021-44228" {
			t.Errorf("expected CVE-2021-44228, got %s", records[0].CVE)
		}

		if !records[0].IsKEV {
			t.Error("expected IsKEV to be true")
		}

		if records[0].DueDate == nil {
			t.Error("expected DueDate to be set")
		} else if *records[0].DueDate != "2021-12-24" {
			t.Errorf("expected due date 2021-12-24, got %s", *records[0].DueDate)
		}
	})
}

func TestParseDueDate(t *testing.T) {
	t.Run("parses valid due date", func(t *testing.T) {
		dueDate := "2021-12-24"
		result := parseDueDate(dueDate)

		if result == nil {
			t.Error("expected non-nil result")
		} else if *result != dueDate {
			t.Errorf("expected %s, got %s", dueDate, *result)
		}
	})

	t.Run("handles empty due date", func(t *testing.T) {
		result := parseDueDate("")

		if result != nil {
			t.Error("expected nil result for empty due date")
		}
	})
}

func TestKEVSyncer_Sync(t *testing.T) {
	t.Run("integration test - requires database", func(t *testing.T) {
		t.Skip("integration test - requires database")
	})
}
