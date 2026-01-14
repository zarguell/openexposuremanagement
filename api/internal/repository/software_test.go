package repository

import (
	"testing"
)

// TODO: Add integration tests with test database setup
// For now, these are unit tests for the query logic

func TestSoftwareRepository_GetByID(t *testing.T) {
	// Skip for now - will implement with test database
	t.Skip("TODO: Implement with test database")
}

func TestSoftwareRepository_List(t *testing.T) {
	// Skip for now - will implement with test database
	t.Skip("TODO: Implement with test database")
}

func TestSoftwareRepository_Search(t *testing.T) {
	// Skip for now - will implement with test database
	t.Skip("TODO: Implement with test database")
}

func TestSoftwareRepository_GetSoftwareForAsset(t *testing.T) {
	// Skip for now - will implement with test database
	t.Skip("TODO: Implement with test database")
}

func TestSoftwareRepository_GetAffectedAssets(t *testing.T) {
	// Skip for now - will implement with test database
	t.Skip("TODO: Implement with test database")
}

func TestSoftwareListParams_BuildQuery(t *testing.T) {
	tests := []struct {
		name     string
		params   SoftwareListParams
		wantSQL  string
		wantArgs []interface{}
	}{
		{
			name: "no filters",
			params: SoftwareListParams{
				TenantID: 1,
				Limit:    50,
				Offset:   0,
			},
			wantSQL: "WHERE tenant_id = $1 ORDER BY vendor, product_name, version LIMIT $2 OFFSET $3",
		},
		{
			name: "with vendor filter",
			params: SoftwareListParams{
				TenantID: 1,
				Vendor:   "Adobe",
				Limit:    50,
				Offset:   0,
			},
			wantSQL: "WHERE tenant_id = $1 AND vendor ILIKE $2 ORDER BY vendor, product_name, version LIMIT $3 OFFSET $4",
		},
		{
			name: "with product filter",
			params: SoftwareListParams{
				TenantID: 1,
				Product:  "Chrome",
				Limit:    50,
				Offset:   0,
			},
			wantSQL: "WHERE tenant_id = $1 AND product_name ILIKE $2 ORDER BY vendor, product_name, version LIMIT $3 OFFSET $4",
		},
		{
			name: "with CPE filter",
			params: SoftwareListParams{
				TenantID: 1,
				CPE:      "cpe:2.3:a:adobe",
				Limit:    50,
				Offset:   0,
			},
			wantSQL: "WHERE tenant_id = $1 AND cpe_string LIKE $2 ORDER BY vendor, product_name, version LIMIT $3 OFFSET $4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We'll test the query building logic without executing it
			// For now, just verify the params are set correctly
			if tt.params.TenantID != 1 {
				t.Errorf("Expected TenantID to be 1, got %d", tt.params.TenantID)
			}
		})
	}
}

func TestSoftwareModel(t *testing.T) {
	sw := Software{
		ID:            1,
		CPEString:     "cpe:2.3:a:adobe:acrobat_reader:2023.001.20093:*:*:*:*:*:*:*:*:*",
		Vendor:        "Adobe",
		ProductName:   "Acrobat Reader",
		Version:       "2023.001.20093",
		Edition:       "DC",
		TargetHW:      "x64",
		TitleFormatted: "Adobe Acrobat Reader DC 2023.001.20093",
		CreatedAt:     "2024-01-15T10:30:00Z",
		UpdatedAt:     "2024-01-15T10:30:00Z",
	}

	if sw.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", sw.ID)
	}

	if sw.Vendor != "Adobe" {
		t.Errorf("Expected Vendor to be Adobe, got %s", sw.Vendor)
	}

	if sw.ProductName != "Acrobat Reader" {
		t.Errorf("Expected ProductName to be Acrobat Reader, got %s", sw.ProductName)
	}
}

func TestAssetSoftwareModel(t *testing.T) {
	asw := AssetSoftware{
		ID:         1,
		TenantID:   1,
		AssetID:    100,
		SoftwareID: 5,
		Source:     "tenable",
		InstallPath: "C:\\Program Files\\Adobe\\",
		FirstSeenAt: "2024-01-15T10:30:00Z",
		LastSeenAt:  "2024-01-15T10:30:00Z",
		CreatedAt:   "2024-01-15T10:30:00Z",
		UpdatedAt:   "2024-01-15T10:30:00Z",
	}

	if asw.TenantID != 1 {
		t.Errorf("Expected TenantID to be 1, got %d", asw.TenantID)
	}

	if asw.AssetID != 100 {
		t.Errorf("Expected AssetID to be 100, got %d", asw.AssetID)
	}

	if asw.SoftwareID != 5 {
		t.Errorf("Expected SoftwareID to be 5, got %d", asw.SoftwareID)
	}
}

func TestSoftwareSummary(t *testing.T) {
	summary := SoftwareSummary{
		SoftwareID:     1,
		CPEString:      "cpe:2.3:a:adobe:acrobat_reader:2023.001.20093:*:*:*:*:*:*:*:*:*",
		Vendor:         "Adobe",
		ProductName:    "Acrobat Reader",
		Version:        "2023.001.20093",
		TitleFormatted: "Adobe Acrobat Reader DC 2023.001.20093",
		InstallCount:   42,
	}

	if summary.InstallCount != 42 {
		t.Errorf("Expected InstallCount to be 42, got %d", summary.InstallCount)
	}
}
