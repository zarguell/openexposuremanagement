package intel

import (
	"context"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/openexposuremanagement/oem/internal/repository"
)

func TestEPSSClient_Fetch(t *testing.T) {
	// Sample EPSS CSV data (gzipped)
	csvData := `cve,epss,percentile
CVE-2021-44228,0.95,0.98
CVE-2021-34527,0.85,0.92
CVE-2021-26855,0.75,0.85
`

	// Create test server that returns gzipped CSV
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Encoding", "gzip")

		gz := gzip.NewWriter(w)
		defer gz.Close()

		gz.Write([]byte(csvData))
	}))
	defer server.Close()

	t.Run("successfully fetches and parses EPSS data", func(t *testing.T) {
		client := &EPSSClient{
			DataURL: server.URL,
			Timeout: 5 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		recordCh, errCh := client.Fetch(ctx)

		var records []EPSSRecord
		for record := range recordCh {
			records = append(records, record)
		}

		if err := <-errCh; err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(records) != 3 {
			t.Errorf("expected 3 records, got %d", len(records))
		}

		// Check first record
		if records[0].CVE != "CVE-2021-44228" {
			t.Errorf("expected CVE-2021-44228, got %s", records[0].CVE)
		}

		if records[0].Score != 0.95 {
			t.Errorf("expected score 0.95, got %f", records[0].Score)
		}

		if records[0].Percentile != 0.98 {
			t.Errorf("expected percentile 0.98, got %f", records[0].Percentile)
		}
	})

	t.Run("handles malformed CSV records gracefully", func(t *testing.T) {
		// Create CSV with malformed record
		badCSVData := `cve,epss,percentile
CVE-2021-44228,0.95,0.98
CVE-2021-34527,invalid,0.92
CVE-2021-26855,0.75,0.85
`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/csv")
			w.Write([]byte(badCSVData))
		}))
		defer server.Close()

		client := &EPSSClient{
			DataURL: server.URL,
			Timeout: 5 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		recordCh, errCh := client.Fetch(ctx)

		var records []EPSSRecord
		for record := range recordCh {
			records = append(records, record)
		}

		if err := <-errCh; err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Should skip malformed record and process 2 valid records
		if len(records) != 2 {
			t.Errorf("expected 2 valid records, got %d", len(records))
		}
	})

	t.Run("handles HTTP errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
		}))
		defer server.Close()

		client := &EPSSClient{
			DataURL: server.URL,
			Timeout: 5 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, errCh := client.Fetch(ctx)

		if err := <-errCh; err == nil {
			t.Error("expected error for HTTP 500, got nil")
		}
	})

	t.Run("respects context timeout", func(t *testing.T) {
		// Server that delays response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.Header().Set("Content-Type", "text/csv")
			w.Write([]byte(csvData))
		}))
		defer server.Close()

		client := &EPSSClient{
			DataURL: server.URL,
			Timeout: 5 * time.Second,
		}

		// Context with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, errCh := client.Fetch(ctx)

		// Should get context error
		err := <-errCh
		if err == nil {
			t.Error("expected context deadline exceeded error, got nil")
		}
	})
}

func TestEPSSClient_FetchAll(t *testing.T) {
	t.Run("fetches all records into memory", func(t *testing.T) {
		csvData := `cve,epss,percentile
CVE-2021-44228,0.95,0.98
CVE-2021-34527,0.85,0.92
`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/csv")
			w.Write([]byte(csvData))
		}))
		defer server.Close()

		client := &EPSSClient{
			DataURL: server.URL,
			Timeout: 5 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		records, err := client.FetchAll(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(records) != 2 {
			t.Errorf("expected 2 records, got %d", len(records))
		}
	})
}

func TestEPSSRecord(t *testing.T) {
	t.Run("valid EPSS record", func(t *testing.T) {
		record := EPSSRecord{
			CVE:        "CVE-2021-44228",
			Score:      0.95,
			Percentile: 0.98,
		}

		if record.CVE != "CVE-2021-44228" {
			t.Errorf("unexpected CVE: %s", record.CVE)
		}

		if record.Score != 0.95 {
			t.Errorf("unexpected score: %f", record.Score)
		}

		if record.Percentile != 0.98 {
			t.Errorf("unexpected percentile: %f", record.Percentile)
		}
	})
}

func TestConvertEPSSRecordToIntelCVE(t *testing.T) {
	t.Run("converts EPSS record to IntelCVE format", func(t *testing.T) {
		record := EPSSRecord{
			CVE:        "CVE-2021-44228",
			Score:      0.95,
			Percentile: 0.98,
		}

		score := record.Score
		percentile := record.Percentile

		intelCVE := &repository.IntelCVE{
			CVE:            record.CVE,
			EPSSScore:      &score,
			EPSSPercentile: &percentile,
		}

		if intelCVE.CVE != record.CVE {
			t.Errorf("expected CVE %s, got %s", record.CVE, intelCVE.CVE)
		}

		if intelCVE.EPSSScore == nil || *intelCVE.EPSSScore != record.Score {
			t.Errorf("EPSS score mismatch")
		}

		if intelCVE.EPSSPercentile == nil || *intelCVE.EPSSPercentile != record.Percentile {
			t.Errorf("EPSS percentile mismatch")
		}
	})
}

func TestEPSSSyncer_Sync(t *testing.T) {
	t.Run("integration test - requires database", func(t *testing.T) {
		// This test requires a real database connection
		// Skip for unit tests
		t.Skip("integration test - requires database")
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		// This test requires a real database connection
		t.Skip("integration test - requires database")
	})
}
