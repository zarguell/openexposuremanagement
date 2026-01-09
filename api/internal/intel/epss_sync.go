package intel

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/rs/zerolog/log"
)

// EPSSSyncer handles syncing EPSS scores to the database
type EPSSSyncer struct {
	client    *EPSSClient
	repo      *repository.IntelRepository
	batchSize int
}

// NewEPSSSyncer creates a new EPSS syncer
func NewEPSSSyncer(db *sqlx.DB) *EPSSSyncer {
	return &EPSSSyncer{
		client:    NewEPSSClient(),
		repo:      repository.NewIntelRepository(db),
		batchSize: 500, // Process 500 EPSS records per batch
	}
}

// EPSSSyncResult represents the result of an EPSS sync operation
type EPSSSyncResult struct {
	TotalProcessed int
	TotalUpdated   int
	Failed         int
	Skipped        int
	Duration       time.Duration
}

// Sync fetches EPSS data and syncs it to the database
func (s *EPSSSyncer) Sync(ctx context.Context) (*EPSSSyncResult, error) {
	log.Info().Msg("Starting EPSS sync")

	startTime := time.Now()
	result := &EPSSSyncResult{}

	// Start sync run record
	syncRun, err := s.repo.StartSyncRun(ctx, "epss")
	if err != nil {
		return nil, fmt.Errorf("failed to start sync run: %w", err)
	}

	// Fetch EPSS data (streaming)
	recordCh, errCh := s.client.Fetch(ctx)

	// Process records in batches
	batch := make([]*repository.IntelCVE, 0, s.batchSize)

	for record := range recordCh {
		// Convert EPSS record to IntelCVE
		intelCVE := &repository.IntelCVE{
			CVE:            record.CVE,
			EPSSScore:      &record.Score,
			EPSSPercentile: &record.Percentile,
		}

		batch = append(batch, intelCVE)

		// Process batch when full
		if len(batch) >= s.batchSize {
			if err := s.processBatch(ctx, batch, result); err != nil {
				s.failSync(ctx, syncRun.ID, err)
				return nil, fmt.Errorf("failed to process batch: %w", err)
			}
			batch = batch[:0] // Clear batch
		}
	}

	// Process remaining records in batch
	if len(batch) > 0 {
		if err := s.processBatch(ctx, batch, result); err != nil {
			s.failSync(ctx, syncRun.ID, err)
			return nil, fmt.Errorf("failed to process final batch: %w", err)
		}
	}

	// Check for errors from fetch
	if err := <-errCh; err != nil {
		s.failSync(ctx, syncRun.ID, err)
		return nil, fmt.Errorf("error fetching EPSS data: %w", err)
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
		Int("skipped", result.Skipped).
		Dur("duration", result.Duration).
		Msg("EPSS sync completed")

	return result, nil
}

// processBatch processes a batch of EPSS records
func (s *EPSSSyncer) processBatch(ctx context.Context, batch []*repository.IntelCVE, result *EPSSSyncResult) error {
	for _, intelCVE := range batch {
		// Check if CVE already exists in intel_cve table
		existing, err := s.repo.GetByCVE(ctx, intelCVE.CVE)
		if err != nil {
			// CVE doesn't exist yet, skip it
			// (it will be created when NVD sync runs)
			result.Skipped++
			continue
		}

		// Update existing record with EPSS data
		existing.EPSSScore = intelCVE.EPSSScore
		existing.EPSSPercentile = intelCVE.EPSSPercentile

		if err := s.repo.UpsertCVE(ctx, existing); err != nil {
			log.Error().Err(err).Str("cve", intelCVE.CVE).Msg("Failed to upsert EPSS data")
			result.Failed++
			continue
		}

		result.TotalUpdated++
		result.TotalProcessed++
	}

	return nil
}

// failSync marks a sync run as failed
func (s *EPSSSyncer) failSync(ctx context.Context, syncRunID int, err error) {
	failErr := s.repo.FailSyncRun(ctx, syncRunID, err.Error())
	if failErr != nil {
		log.Error().Err(failErr).Msg("Failed to mark sync run as failed")
	}
}
