package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

// DashboardCounts represents finding counts by effective status
type DashboardCounts struct {
	TenantID        int64  `db:"tenant_id"`
	EffectiveStatus string `db:"effective_status"`
	Count           int    `db:"count"`
}

// DashboardAssets represents asset summary statistics
type DashboardAssets struct {
	TenantID                int64      `db:"tenant_id"`
	TotalAssets             int        `db:"total_assets"`
	ActiveAssets            int        `db:"active_assets"`
	MostRecentAssetActivity *time.Time `db:"most_recent_asset_activity"`
}

// DashboardOpenFindings represents open findings summary by severity
type DashboardOpenFindings struct {
	TenantID          int64      `db:"tenant_id"`
	OpenCount         int        `db:"open_count"`
	SuppressedCount   int        `db:"suppressed_count"`
	CriticalOpenCount int        `db:"critical_open_count"`
	HighOpenCount     int        `db:"high_open_count"`
	MostRecentFinding *time.Time `db:"most_recent_finding"`
}

// DashboardData aggregates all dashboard statistics for a tenant
type DashboardData struct {
	Assets    DashboardAssets
	Findings  DashboardOpenFindings
	IntelSync *IntelSyncStatus
}

// DashboardRepository handles dashboard data access
type DashboardRepository struct {
	db *sqlx.DB
}

// NewDashboardRepository creates a new dashboard repository
func NewDashboardRepository(db *sqlx.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

// GetTenantData retrieves all dashboard data for a tenant
func (r *DashboardRepository) GetTenantData(ctx context.Context, tenantID int64) (*DashboardData, error) {
	assets, err := r.getAssets(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	findings, err := r.getOpenFindings(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	intelSync, err := r.getIntelSyncStatus(ctx)
	if err != nil {
		return nil, err
	}

	data := &DashboardData{
		Assets:    *assets,
		Findings:  *findings,
		IntelSync: intelSync,
	}

	// Log telemetry for debugging
	if assets.TotalAssets == 0 {
		// This might indicate materialized view is not populated
		// Log warning but don't fail
	}

	if findings.OpenCount == 0 {
		// This might be expected if no findings exist
		// But log for visibility
	}

	return data, nil
}

// getAssets retrieves asset statistics from materialized view
func (r *DashboardRepository) getAssets(ctx context.Context, tenantID int64) (*DashboardAssets, error) {
	var assets DashboardAssets
	err := r.db.GetContext(ctx, &assets,
		"SELECT * FROM mv_dashboard_assets WHERE tenant_id = $1", tenantID)
	if err != nil {
		return nil, err
	}
	return &assets, nil
}

// getOpenFindings retrieves open findings statistics from materialized view
func (r *DashboardRepository) getOpenFindings(ctx context.Context, tenantID int64) (*DashboardOpenFindings, error) {
	var findings DashboardOpenFindings
	err := r.db.GetContext(ctx, &findings,
		"SELECT * FROM mv_dashboard_open_findings WHERE tenant_id = $1", tenantID)
	if err != nil {
		return nil, err
	}
	return &findings, nil
}

// IntelSyncStatus represents the status of intel sync runs
type IntelSyncStatus struct {
	LastSyncTime *time.Time `db:"last_sync_time" json:"last_sync_time,omitempty"`
	Status       string     `db:"status" json:"status"`
	ErrorText    *string    `db:"error_text" json:"error_text,omitempty"`
}

// getIntelSyncStatus retrieves the latest intel sync status
func (r *DashboardRepository) getIntelSyncStatus(ctx context.Context) (*IntelSyncStatus, error) {
	var status IntelSyncStatus
	err := r.db.GetContext(ctx, &status, `
		SELECT
			MAX(finished_at) as last_sync_time,
			CASE
				WHEN MAX(started_at) > COALESCE(MAX(finished_at), '1970-01-01'::timestamptz) THEN 'running'
				WHEN MAX(status) = 'failed' THEN 'error'
				ELSE 'ok'
			END as status,
			MAX(CASE WHEN status = 'failed' THEN error_text END) as error_text
		FROM intel_sync_runs
	`)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// RefreshMaterializedViews refreshes all dashboard materialized views concurrently
func (r *DashboardRepository) RefreshMaterializedViews(ctx context.Context) error {
	views := []string{
		"mv_dashboard_counts",
		"mv_dashboard_assets",
		"mv_dashboard_open_findings",
	}

	for _, view := range views {
		_, err := r.db.ExecContext(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY "+view)
		if err != nil {
			return err
		}
	}

	return nil
}
