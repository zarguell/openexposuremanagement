package intel

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/rs/zerolog/log"
)

// Syncer orchestrates all threat intelligence syncing
type Syncer struct {
	nvd      *NVDSyncer
	epss     *EPSSSyncer
	kev      *KEVSyncer
	repo     *repository.IntelRepository
}

// NewSyncer creates a new unified syncer
func NewSyncer(db *sqlx.DB) *Syncer {
	return &Syncer{
		nvd:  NewNVDSyncer(db),
		epss: NewEPSSSyncer(db),
		kev:  NewKEVSyncer(db),
		repo: repository.NewIntelRepository(db),
	}
}

// FullSyncResult represents the result of a full sync operation
type FullSyncResult struct {
	NVDResult  *SyncResult   `json:"nvd_result"`
	EPSSResult *EPSSSyncResult `json:"epss_result"`
	KEVResult  *KEVSyncResult  `json:"kev_result"`
	Duration   time.Duration  `json:"duration"`
}

// SyncAll performs a full sync of all threat intel sources
// Order: NVD (create records) → EPSS (enrich with scores) → KEV (mark exploited)
func (s *Syncer) SyncAll(ctx context.Context) (*FullSyncResult, error) {
	log.Info().Msg("Starting full threat intelligence sync")

	startTime := time.Now()
	result := &FullSyncResult{}

	// Step 1: Sync NVD data (creates base CVE records)
	log.Info().Msg("Step 1: Syncing NVD data")
	nvdResult, err := s.nvd.SyncRecent(ctx, 30) // Sync last 30 days
	if err != nil {
		return nil, fmt.Errorf("NVD sync failed: %w", err)
	}
	result.NVDResult = nvdResult

	log.Info().
		Int("processed", nvdResult.TotalProcessed).
		Int("updated", nvdResult.TotalUpdated).
		Msg("NVD sync completed")

	// Step 2: Sync EPSS data (enriches with exploit probability)
	log.Info().Msg("Step 2: Syncing EPSS data")
	epssResult, err := s.epss.Sync(ctx)
	if err != nil {
		// EPSS failure is not fatal - continue with KEV
		log.Error().Err(err).Msg("EPSS sync failed, continuing with KEV")
		result.EPSSResult = &EPSSSyncResult{Failed: 1} // Mark as failed
	} else {
		result.EPSSResult = epssResult

		log.Info().
			Int("processed", epssResult.TotalProcessed).
			Int("updated", epssResult.TotalUpdated).
			Msg("EPSS sync completed")
	}

	// Step 3: Sync KEV data (marks known exploited vulnerabilities)
	log.Info().Msg("Step 3: Syncing CISA KEV data")
	kevResult, err := s.kev.Sync(ctx)
	if err != nil {
		// KEV failure is not fatal - report results
		log.Error().Err(err).Msg("KEV sync failed")
		result.KEVResult = &KEVSyncResult{Failed: 1} // Mark as failed
	} else {
		result.KEVResult = kevResult

		log.Info().
			Int("processed", kevResult.TotalProcessed).
			Int("updated", kevResult.TotalUpdated).
			Msg("KEV sync completed")
	}

	result.Duration = time.Since(startTime)

	log.Info().
		Dur("duration", result.Duration).
		Msg("Full threat intelligence sync completed")

	return result, nil
}

// SyncNVD syncs only NVD data
func (s *Syncer) SyncNVD(ctx context.Context, days int) (*SyncResult, error) {
	return s.nvd.SyncRecent(ctx, days)
}

// SyncEPSS syncs only EPSS data
func (s *Syncer) SyncEPSS(ctx context.Context) (*EPSSSyncResult, error) {
	return s.epss.Sync(ctx)
}

// SyncKEV syncs only KEV data
func (s *Syncer) SyncKEV(ctx context.Context) (*KEVSyncResult, error) {
	return s.kev.Sync(ctx)
}

// GetSyncStatus returns the status of the most recent sync runs for all sources
func (s *Syncer) GetSyncStatus(ctx context.Context) (*SyncStatus, error) {
	sources := []string{"nvd", "epss", "kev"}
	status := &SyncStatus{
		Sources: make(map[string]SourceStatus),
	}

	for _, source := range sources {
		run, err := s.repo.GetLatestSyncRun(ctx, source)
		if err != nil {
			// No sync run yet for this source
			status.Sources[source] = SourceStatus{
				Source:     source,
				Status:     "never_run",
				LastSync:   nil,
				Records:    0,
				Error:      nil,
			}
			continue
		}

		status.Sources[source] = SourceStatus{
			Source:        source,
			Status:        run.Status,
			LastSync:      &run.StartedAt,
			FinishedAt:    run.FinishedAt,
			Records:       run.RecordsProcessed,
			RecordsUpdated: run.RecordsUpdated,
			Error:         run.ErrorText,
		}
	}

	return status, nil
}

// SyncStatus represents the sync status for all sources
type SyncStatus struct {
	Sources map[string]SourceStatus `json:"sources"`
}

// SourceStatus represents the sync status for a single source
type SourceStatus struct {
	Source         string     `json:"source"`
	Status         string     `json:"status"` // running, completed, failed, never_run
	LastSync       *time.Time `json:"last_sync,omitempty"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	Records        int        `json:"records"`
	RecordsUpdated int        `json:"records_updated,omitempty"`
	Error          *string    `json:"error,omitempty"`
}
