package server

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/rs/zerolog/log"
)

// ViewRefresher handles periodic refresh of materialized views
type ViewRefresher struct {
	db              *sqlx.DB
	refreshInterval time.Duration
	stopCh          chan struct{}
}

// NewViewRefresher creates a new materialized view refresher
func NewViewRefresher(db *sqlx.DB, refreshInterval time.Duration) *ViewRefresher {
	return &ViewRefresher{
		db:              db,
		refreshInterval: refreshInterval,
		stopCh:          make(chan struct{}),
	}
}

// Start begins the periodic refresh loop
func (vr *ViewRefresher) Start(ctx context.Context) {
	log.Info().Dur("interval", vr.refreshInterval).Msg("Starting materialized view refresher")

	ticker := time.NewTicker(vr.refreshInterval)
	defer ticker.Stop()

	// Run once immediately on start
	vr.refresh(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Context cancelled, stopping view refresher")
			return
		case <-vr.stopCh:
			log.Info().Msg("Stop signal received, stopping view refresher")
			return
		case <-ticker.C:
			vr.refresh(ctx)
		}
	}
}

// Stop gracefully stops the refresher
func (vr *ViewRefresher) Stop() {
	close(vr.stopCh)
}

// refresh performs the actual refresh of materialized views
func (vr *ViewRefresher) refresh(ctx context.Context) {
	log.Debug().Msg("Refreshing materialized views")

	start := time.Now()
	dashRepo := repository.NewDashboardRepository(vr.db)
	err := dashRepo.RefreshMaterializedViews(ctx)
	duration := time.Since(start)

	if err != nil {
		log.Error().Err(err).Dur("duration", duration).Msg("Failed to refresh materialized views")
		return
	}

	log.Debug().Dur("duration", duration).Msg("Successfully refreshed materialized views")
}

// RefreshNow performs an immediate one-time refresh
func (vr *ViewRefresher) RefreshNow(ctx context.Context) error {
	dashRepo := repository.NewDashboardRepository(vr.db)
	return dashRepo.RefreshMaterializedViews(ctx)
}
