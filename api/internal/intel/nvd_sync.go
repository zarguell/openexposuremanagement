package intel

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/rs/zerolog/log"
)

// NVDSyncer handles syncing CVE data from NVD to the database
type NVDSyncer struct {
	client    *NVDClient
	repo      *repository.IntelRepository
	batchSize int
}

// NewNVDSyncer creates a new NVD syncer
func NewNVDSyncer(db *sqlx.DB) *NVDSyncer {
	return &NVDSyncer{
		client:    NewNVDClient(),
		repo:      repository.NewIntelRepository(db),
		batchSize: 100, // Process 100 CVEs per batch
	}
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	TotalProcessed int
	TotalUpdated   int
	Failed         int
	Errors         []string
	Duration       time.Duration
}

// SyncRecent syncs CVEs modified in the last N days
func (s *NVDSyncer) SyncRecent(ctx context.Context, days int) (*SyncResult, error) {
	log.Info().Int("days", days).Msg("Starting NVD sync for recent CVEs")

	startTime := time.Now()
	result := &SyncResult{}

	// Start sync run record
	syncRun, err := s.repo.StartSyncRun(ctx, "nvd")
	if err != nil {
		return nil, fmt.Errorf("failed to start sync run: %w", err)
	}

	// Calculate date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Fetch CVEs from NVD
	params := &FetchParams{
		LastModStartDate: startDate,
		LastModEndDate:   endDate,
		ResultsPerPage:   s.batchSize,
	}

	allCVEs := make([]*repository.IntelCVE, 0)
	startIndex := 0

	for {
		params.StartIndex = startIndex
		resp, err := s.client.Fetch(params)
		if err != nil {
			s.failSync(ctx, syncRun.ID, err)
			return nil, fmt.Errorf("failed to fetch CVEs: %w", err)
		}

		// Convert and collect CVEs
		for _, entry := range resp.Vulnerabilities {
			intelCVE := convertToRepoIntelCVE(entry.CVE)
			allCVEs = append(allCVEs, intelCVE)
		}

		result.TotalProcessed = len(allCVEs)

		// Check if we've fetched all results
		if len(resp.Vulnerabilities) == 0 || len(allCVEs) >= resp.TotalResults {
			break
		}

		startIndex += len(resp.Vulnerabilities)

		log.Debug().
			Int("fetched", len(allCVEs)).
			Int("total", resp.TotalResults).
			Msg("Fetched batch from NVD")
	}

	// Upsert to database in batches
	if err := s.upsertBatches(ctx, allCVEs, result); err != nil {
		s.failSync(ctx, syncRun.ID, err)
		return nil, fmt.Errorf("failed to upsert CVEs: %w", err)
	}

	// Complete sync run
	if err := s.repo.CompleteSyncRun(ctx, syncRun.ID, result.TotalProcessed, result.TotalUpdated); err != nil {
		log.Error().Err(err).Msg("Failed to complete sync run")
	}

	result.Duration = time.Since(startTime)

	log.Info().
		Int("processed", result.TotalProcessed).
		Int("updated", result.TotalUpdated).
		Int("failed", result.Failed).
		Dur("duration", result.Duration).
		Msg("NVD sync completed")

	return result, nil
}

// SyncSpecificCVEs syncs specific CVEs by ID
func (s *NVDSyncer) SyncSpecificCVEs(ctx context.Context, cveIDs []string) (*SyncResult, error) {
	log.Info().Int("count", len(cveIDs)).Msg("Starting NVD sync for specific CVEs")

	startTime := time.Now()
	result := &SyncResult{}

	// Start sync run record
	syncRun, err := s.repo.StartSyncRun(ctx, "nvd")
	if err != nil {
		return nil, fmt.Errorf("failed to start sync run: %w", err)
	}

	allCVEs := make([]*repository.IntelCVE, 0, len(cveIDs))

	// Fetch each CVE
	for _, cveID := range cveIDs {
		detail, err := s.client.FetchCVE(cveID)
		if err != nil {
			log.Error().Err(err).Str("cve", cveID).Msg("Failed to fetch CVE")
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", cveID, err))
			continue
		}

		intelCVE := &repository.IntelCVE{
			CVE:         detail.ID,
			Description: firstOrEmpty(detail.Descriptions),
			CVSSScore:   detail.CVSSScore,
			CVSSVector:  detail.CVSSVector,
		}

		allCVEs = append(allCVEs, intelCVE)
	}

	result.TotalProcessed = len(allCVEs)

	// Upsert to database
	if err := s.upsertBatches(ctx, allCVEs, result); err != nil {
		s.failSync(ctx, syncRun.ID, err)
		return nil, fmt.Errorf("failed to upsert CVEs: %w", err)
	}

	// Complete sync run
	if err := s.repo.CompleteSyncRun(ctx, syncRun.ID, result.TotalProcessed, result.TotalUpdated); err != nil {
		log.Error().Err(err).Msg("Failed to complete sync run")
	}

	result.Duration = time.Since(startTime)

	log.Info().
		Int("processed", result.TotalProcessed).
		Int("updated", result.TotalUpdated).
		Int("failed", result.Failed).
		Dur("duration", result.Duration).
		Msg("NVD sync completed")

	return result, nil
}

// upsertBatches upserts CVEs in batches
func (s *NVDSyncer) upsertBatches(ctx context.Context, cves []*repository.IntelCVE, result *SyncResult) error {
	for i := 0; i < len(cves); i += s.batchSize {
		end := i + s.batchSize
		if end > len(cves) {
			end = len(cves)
		}

		batch := cves[i:end]
		if err := s.repo.UpsertCVEsBatch(ctx, batch); err != nil {
			log.Error().Err(err).
				Int("batch_start", i).
				Int("batch_end", end).
				Msg("Failed to upsert batch")
			result.Failed += len(batch)
			result.Errors = append(result.Errors, fmt.Sprintf("batch %d-%d: %v", i, end, err))
			continue
		}

		result.TotalUpdated += len(batch)

		log.Debug().
			Int("batch_start", i).
			Int("batch_end", end).
			Msg("Upserted batch")
	}

	return nil
}

// failSync marks a sync run as failed
func (s *NVDSyncer) failSync(ctx context.Context, syncRunID int, err error) {
	failErr := s.repo.FailSyncRun(ctx, syncRunID, err.Error())
	if failErr != nil {
		log.Error().Err(failErr).Msg("Failed to mark sync run as failed")
	}
}

// convertToRepoIntelCVE converts an NVDCVE to repository.IntelCVE
func convertToRepoIntelCVE(nvdCVE NVDCVE) *repository.IntelCVE {
	intelCVE := &CVE{
		CVE: nvdCVE.ID,
	}

	if nvdCVE.Metrics != nil && len(nvdCVE.Metrics.CVSSMetricV31) > 0 {
		score := nvdCVE.Metrics.CVSSMetricV31[0].CVSSData.Score
		vector := nvdCVE.Metrics.CVSSMetricV31[0].CVSSData.VectorString

		if score > 0 {
			intelCVE.CVSSScore = &score
		}
		if vector != "" {
			intelCVE.CVSSVector = vector
		}
	}

	// Extract English descriptions
	var descriptions []string
	for _, desc := range nvdCVE.Descriptions {
		if desc.Lang == "en" {
			descriptions = append(descriptions, desc.Value)
		}
	}

	if len(descriptions) > 0 {
		intelCVE.Description = joinStrings(descriptions, "\n\n")
	}

	// Convert to repository format
	return &repository.IntelCVE{
		CVE:         intelCVE.CVE,
		Description: intelCVE.Description,
		CVSSScore:   intelCVE.CVSSScore,
		CVSSVector:  intelCVE.CVSSVector,
	}
}

// firstOrEmpty returns the first string or empty string
func firstOrEmpty(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[0]
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
