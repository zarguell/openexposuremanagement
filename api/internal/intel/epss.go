package intel

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// EPSS data URL from FIRST.org
	epssDataURL = "https://epss.cyentia.com/epss_scores-current.csv.gz"
	// EPSS data is typically updated daily
	epssTimeout = 5 * time.Minute
)

// EPSSClient fetches EPSS exploit prediction scores
type EPSSClient struct {
	DataURL  string
	Timeout  time.Duration
}

// NewEPSSClient creates a new EPSS client
func NewEPSSClient() *EPSSClient {
	return &EPSSClient{
		DataURL: epssDataURL,
		Timeout: epssTimeout,
	}
}

// EPSSRecord represents a single EPSS score record
type EPSSRecord struct {
	CVE        string
	Score      float64
	Percentile float64
}

// Fetch fetches and parses EPSS data from the source
// Returns a channel of records to allow streaming processing
func (c *EPSSClient) Fetch(ctx context.Context) (<-chan EPSSRecord, <-chan error) {
	recordCh := make(chan EPSSRecord, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(recordCh)
		defer close(errCh)

		log.Info().Str("url", c.DataURL).Msg("Fetching EPSS data")

		// Create HTTP client with timeout
		client := &http.Client{Timeout: c.Timeout}

		req, err := http.NewRequestWithContext(ctx, "GET", c.DataURL, nil)
		if err != nil {
			errCh <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			errCh <- fmt.Errorf("failed to fetch EPSS data: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errCh <- fmt.Errorf("EPSS data fetch failed with status %d", resp.StatusCode)
			return
		}

		// Note: The data is gzipped, but Go's http.Client automatically
		// handles gzip decompression if the server sends Content-Encoding: gzip

		// Parse CSV
		reader := csv.NewReader(resp.Body)
		reader.Comma = ',' // EPSS uses comma-separated values

		// Read header
		header, err := reader.Read()
		if err != nil {
			errCh <- fmt.Errorf("failed to read CSV header: %w", err)
			return
		}

		// Validate header
		if len(header) < 3 || header[0] != "cve" || header[1] != "epss" || header[2] != "percentile" {
			errCh <- fmt.Errorf("unexpected CSV header format: %v", header)
			return
		}

		recordCount := 0
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				errCh <- fmt.Errorf("failed to read CSV record: %w", err)
				return
			}

			// Parse record
			if len(record) < 3 {
				log.Warn().Strs("record", record).Msg("Skipping malformed EPSS record")
				continue
			}

			cve := record[0]
			score, err := strconv.ParseFloat(record[1], 64)
			if err != nil {
				log.Warn().Str("cve", cve).Str("score", record[1]).Err(err).Msg("Failed to parse EPSS score")
				continue
			}

			percentile, err := strconv.ParseFloat(record[2], 64)
			if err != nil {
				log.Warn().Str("cve", cve).Str("percentile", record[2]).Err(err).Msg("Failed to parse EPSS percentile")
				continue
			}

			select {
			case recordCh <- EPSSRecord{
				CVE:        cve,
				Score:      score,
				Percentile: percentile,
			}:
				recordCount++
				if recordCount%10000 == 0 {
					log.Debug().Int("count", recordCount).Msg("Processed EPSS records")
				}
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}
		}

		log.Info().Int("total_records", recordCount).Msg("Finished processing EPSS data")
	}()

	return recordCh, errCh
}

// FetchAll fetches all EPSS records into memory
// Use this for smaller datasets or when you need all records at once
func (c *EPSSClient) FetchAll(ctx context.Context) ([]EPSSRecord, error) {
	recordCh, errCh := c.Fetch(ctx)

	var records []EPSSRecord
	for record := range recordCh {
		records = append(records, record)
	}

	// Check for errors
	if err := <-errCh; err != nil {
		return nil, err
	}

	return records, nil
}
