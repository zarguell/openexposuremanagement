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
