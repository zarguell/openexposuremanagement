package intel

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/rs/zerolog/log"
)

// KEVSyncer handles syncing CISA KEV data to the database
type KEVSyncer struct {
	client    *KEVClient
	repo      *repository.IntelRepository
	batchSize int
}

// NewKEVSyncer creates a new KEV syncer
func NewKEVSyncer(db *sqlx.DB) *KEVSyncer {
	return &KEVSyncer{
		client:    NewKEVClient(),
		repo:      repository.NewIntelRepository(db),
		batchSize: 100, // Process 100 KEV records per batch
	}
}

// KEVSyncResult represents the result of a KEV sync operation
type KEVSyncResult struct {
	TotalProcessed int
	TotalUpdated   int
	TotalAdded     int
	Failed         int
	Duration       time.Duration
}

// Sync fetches KEV data and syncs it to the database
func (s *KEVSyncer) Sync(ctx context.Context) (*KEVSyncResult, error) {
	log.Info().Msg("Starting CISA KEV sync")

	startTime := time.Now()
	result := &KEVSyncResult{}

	// Start sync run record
	syncRun, err := s.repo.StartSyncRun(ctx, "kev")
	if err != nil {
		return nil, fmt.Errorf("failed to start sync run: %w", err)
	}

	// Fetch KEV catalog
	catalog, err := s.client.Fetch(ctx)
	if err != nil {
		s.failSync(ctx, syncRun.ID, err)
		return nil, fmt.Errorf("failed to fetch KEV catalog: %w", err)
	}

	// Convert to KEV records
	records := s.client.ToEPSSRecords(catalog)

	// Process in batches
	for i := 0; i < len(records); i += s.batchSize {
		end := i + s.batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		if err := s.processBatch(ctx, batch, result); err != nil {
			s.failSync(ctx, syncRun.ID, err)
			return nil, fmt.Errorf("failed to process batch %d-%d: %w", i, end, err)
		}

		log.Debug().
			Int("batch_start", i).
			Int("batch_end", end).
			Msg("Processed KEV batch")
	}

	// Complete sync run
	if err := s.repo.CompleteSyncRun(ctx, syncRun.ID, result.TotalProcessed, result.TotalUpdated); err != nil {
		log.Error().Err(err).Msg("Failed to complete sync run")
	}

	result.Duration = time.Since(startTime)

	log.Info().
		Int("processed", result.TotalProcessed).
		Int("updated", result.TotalUpdated).
		Int("added", result.TotalAdded).
		Int("failed", result.Failed).
		Dur("duration", result.Duration).
		Msg("KEV sync completed")

	return result, nil
}

// processBatch processes a batch of KEV records
func (s *KEVSyncer) processBatch(ctx context.Context, batch []KEVRecord, result *KEVSyncResult) error {
	for _, record := range batch {
		// Check if CVE already exists
		existing, err := s.repo.GetByCVE(ctx, record.CVE)

		if err != nil {
			// CVE doesn't exist, skip it
			// It will be created when NVD sync runs
			result.TotalProcessed++
			continue
		}

		// Update existing record with KEV data
		existing.IsKEV = record.IsKEV
		existing.KEVDateAdded = parseDateAdded(record)
		existing.KEVDueDate = record.DueDate

		if err := s.repo.UpsertCVE(ctx, existing); err != nil {
			log.Error().Err(err).Str("cve", record.CVE).Msg("Failed to upsert KEV data")
			result.Failed++
			continue
		}

		result.TotalUpdated++
		result.TotalProcessed++
	}

	return nil
}

// parseDateAdded parses the date added from a KEV record
func parseDateAdded(record KEVRecord) *string {
	if record.DateAdded != "" {
		return &record.DateAdded
	}
	return nil
}

// failSync marks a sync run as failed
func (s *KEVSyncer) failSync(ctx context.Context, syncRunID int, err error) {
	failErr := s.repo.FailSyncRun(ctx, syncRunID, err.Error())
	if failErr != nil {
		log.Error().Err(failErr).Msg("Failed to mark sync run as failed")
	}
}
