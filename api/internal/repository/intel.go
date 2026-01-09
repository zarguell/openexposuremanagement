package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

// IntelCVE represents threat intelligence data for a CVE
type IntelCVE struct {
	CVE            string    `db:"cve" json:"cve"`
	Description    string    `db:"description" json:"description,omitempty"`
	CVSSScore      *float64  `db:"cvss_score" json:"cvss_score,omitempty"`
	CVSSVector     string    `db:"cvss_vector" json:"cvss_vector,omitempty"`
	EPSSScore      *float64  `db:"epss_score" json:"epss_score,omitempty"`
	EPSSPercentile *float64  `db:"epss_percentile" json:"epss_percentile,omitempty"`
	IsKEV          bool      `db:"is_kev" json:"is_kev"`
	KEVDateAdded   *string   `db:"kev_date_added" json:"kev_date_added,omitempty"`
	KEVDueDate     *string   `db:"kev_due_date" json:"kev_due_date,omitempty"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// IntelRepository handles threat intelligence data access
type IntelRepository struct {
	db *sqlx.DB
}

// NewIntelRepository creates a new intel repository
func NewIntelRepository(db *sqlx.DB) *IntelRepository {
	return &IntelRepository{db: db}
}

// GetByCVE retrieves intel data for a single CVE
func (r *IntelRepository) GetByCVE(ctx context.Context, cve string) (*IntelCVE, error) {
	var intel IntelCVE
	err := r.db.GetContext(ctx, &intel, "SELECT * FROM intel_cve WHERE cve = $1", cve)
	if err != nil {
		return nil, err
	}
	return &intel, nil
}

// GetByCVEs retrieves intel data for multiple CVEs
func (r *IntelRepository) GetByCVEs(ctx context.Context, cves []string) (map[string]*IntelCVE, error) {
	if len(cves) == 0 {
		return make(map[string]*IntelCVE), nil
	}

	query, args, err := sqlx.In("SELECT * FROM intel_cve WHERE cve IN (?)", cves)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]*IntelCVE)
	for rows.Next() {
		var intel IntelCVE
		if err := rows.StructScan(&intel); err != nil {
			return nil, err
		}
		result[intel.CVE] = &intel
	}

	return result, rows.Err()
}

// UpsertCVE upserts intel data for a single CVE
func (r *IntelRepository) UpsertCVE(ctx context.Context, intel *IntelCVE) error {
	query := `
		INSERT INTO intel_cve (
			cve, description, cvss_score, cvss_vector,
			epss_score, epss_percentile, is_kev, kev_date_added, kev_due_date,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()
		)
		ON CONFLICT (cve) DO UPDATE SET
			description = EXCLUDED.description,
			cvss_score = EXCLUDED.cvss_score,
			cvss_vector = EXCLUDED.cvss_vector,
			epss_score = EXCLUDED.epss_score,
			epss_percentile = EXCLUDED.epss_percentile,
			is_kev = EXCLUDED.is_kev,
			kev_date_added = EXCLUDED.kev_date_added,
			kev_due_date = EXCLUDED.kev_due_date,
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		intel.CVE, intel.Description, intel.CVSSScore, intel.CVSSVector,
		intel.EPSSScore, intel.EPSSPercentile, intel.IsKEV,
		intel.KEVDateAdded, intel.KEVDueDate,
	)
	return err
}

// UpsertCVEsBatch upserts multiple intel CVE records in a transaction
func (r *IntelRepository) UpsertCVEsBatch(ctx context.Context, intels []*IntelCVE) error {
	if len(intels) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO intel_cve (
			cve, description, cvss_score, cvss_vector,
			epss_score, epss_percentile, is_kev, kev_date_added, kev_due_date,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()
		)
		ON CONFLICT (cve) DO UPDATE SET
			description = EXCLUDED.description,
			cvss_score = EXCLUDED.cvss_score,
			cvss_vector = EXCLUDED.cvss_vector,
			epss_score = EXCLUDED.epss_score,
			epss_percentile = EXCLUDED.epss_percentile,
			is_kev = EXCLUDED.is_kev,
			kev_date_added = EXCLUDED.kev_date_added,
			kev_due_date = EXCLUDED.kev_due_date,
			updated_at = NOW()
	`

	stmt, err := tx.PreparexContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, intel := range intels {
		if _, err := stmt.ExecContext(ctx,
			intel.CVE, intel.Description, intel.CVSSScore, intel.CVSSVector,
			intel.EPSSScore, intel.EPSSPercentile, intel.IsKEV,
			intel.KEVDateAdded, intel.KEVDueDate,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SyncRun represents a threat intel sync run
type SyncRun struct {
	ID          int       `db:"id"`
	StartedAt   time.Time `db:"started_at"`
	FinishedAt  *time.Time `db:"finished_at"`
	Status      string    `db:"status"` // "running", "completed", "failed"
	ErrorText   *string   `db:"error_text"`
	Source      string    `db:"source"` // "nvd", "epss", "kev"
	RecordsProcessed int  `db:"records_processed"`
	RecordsUpdated    int  `db:"records_updated"`
}

// StartSyncRun creates a new sync run record with status "running"
func (r *IntelRepository) StartSyncRun(ctx context.Context, source string) (*SyncRun, error) {
	query := `
		INSERT INTO intel_sync_runs (started_at, status, source, records_processed, records_updated)
		VALUES (NOW(), 'running', $1, 0, 0)
		RETURNING id, started_at, status, source
	`
	var syncRun SyncRun
	err := r.db.GetContext(ctx, &syncRun, query, source)
	if err != nil {
		return nil, err
	}
	return &syncRun, nil
}

// CompleteSyncRun marks a sync run as completed
func (r *IntelRepository) CompleteSyncRun(ctx context.Context, syncRunID int, recordsProcessed, recordsUpdated int) error {
	query := `
		UPDATE intel_sync_runs
		SET finished_at = NOW(),
		    status = 'completed',
		    records_processed = $2,
		    records_updated = $3
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, syncRunID, recordsProcessed, recordsUpdated)
	return err
}

// FailSyncRun marks a sync run as failed with an error message
func (r *IntelRepository) FailSyncRun(ctx context.Context, syncRunID int, errMsg string) error {
	query := `
		UPDATE intel_sync_runs
		SET finished_at = NOW(),
		    status = 'failed',
		    error_text = $2
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, syncRunID, errMsg)
	return err
}

// GetLatestSyncRun retrieves the most recent sync run for a source
func (r *IntelRepository) GetLatestSyncRun(ctx context.Context, source string) (*SyncRun, error) {
	query := `
		SELECT id, started_at, finished_at, status, error_text, source,
		       records_processed, records_updated
		FROM intel_sync_runs
		WHERE source = $1
		ORDER BY started_at DESC
		LIMIT 1
	`
	var syncRun SyncRun
	err := r.db.GetContext(ctx, &syncRun, query, source)
	if err != nil {
		return nil, err
	}
	return &syncRun, nil
}
