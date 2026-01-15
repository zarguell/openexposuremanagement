package query_test

import (
	"testing"

	"github.com/openexposuremanagement/oem/internal/services/query"
)

func TestDotWalkingTranslator(t *testing.T) {
	translator := query.NewTranslator()

	t.Run("simple primary entity filter", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "is_active", Operator: "eq", Value: true},
			},
		}
		sql, args, err := translator.Translate("assets", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(sql, "SELECT * FROM assets_extended") {
			t.Errorf("expected simple SELECT, got: %s", sql)
		}
		if !contains(sql, "WHERE assets.is_active = $1") {
			t.Errorf("expected WHERE clause, got: %s", sql)
		}
		if len(args) != 1 {
			t.Errorf("expected 1 arg, got %d", len(args))
		}
	})

	t.Run("dot notation for software filter (INNER JOIN)", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "software.vendor", Operator: "eq", Value: "Microsoft"},
			},
		}
		sql, args, err := translator.Translate("assets", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(sql, "SELECT DISTINCT assets.*") {
			t.Errorf("expected SELECT DISTINCT, got: %s", sql)
		}
		if !contains(sql, "INNER JOIN software_inventory") {
			t.Errorf("expected INNER JOIN, got: %s", sql)
		}
		if !contains(sql, "ON assets.id = software_inventory.asset_id") {
			t.Errorf("expected ON clause, got: %s", sql)
		}
		if !contains(sql, "software_inventory.vendor = $1") {
			t.Errorf("expected filter on joined field, got: %s", sql)
		}
		if len(args) != 1 {
			t.Errorf("expected 1 arg, got %d", len(args))
		}
	})

	t.Run("dot notation with negate (NOT EXISTS)", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "is_active", Operator: "eq", Value: true},
				{Field: "software.vendor", Operator: "eq", Value: "CrowdStrike", Negate: true},
			},
		}
		sql, args, err := translator.Translate("assets", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(sql, "SELECT * FROM assets_extended") {
			t.Errorf("expected SELECT (not DISTINCT), got: %s", sql)
		}
		if !contains(sql, "NOT EXISTS") {
			t.Errorf("expected NOT EXISTS, got: %s", sql)
		}
		if !contains(sql, "software_inventory.vendor = $2") {
			t.Errorf("expected subquery filter, got: %s", sql)
		}
		if len(args) != 2 {
			t.Errorf("expected 2 args, got %d", len(args))
		}
	})

	t.Run("assets with critical CVEs", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "findings.severity", Operator: "eq", Value: "critical"},
				{Field: "findings.effective_status", Operator: "eq", Value: "open"},
			},
		}
		sql, args, err := translator.Translate("assets", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(sql, "SELECT DISTINCT assets.*") {
			t.Errorf("expected SELECT DISTINCT, got: %s", sql)
		}
		if !contains(sql, "INNER JOIN findings") {
			t.Errorf("expected INNER JOIN findings, got: %s", sql)
		}
		if !contains(sql, "ON assets.id = findings.asset_id") {
			t.Errorf("expected ON clause, got: %s", sql)
		}
		if len(args) != 2 {
			t.Errorf("expected 2 args, got %d", len(args))
		}
	})

	t.Run("multiple entity joins (software + findings)", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "software.vendor", Operator: "eq", Value: "Apache"},
				{Field: "findings.cve", Operator: "eq", Value: "CVE-2021-44228"},
			},
		}
		sql, args, err := translator.Translate("assets", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(sql, "INNER JOIN software_inventory") {
			t.Errorf("expected software_inventory join, got: %s", sql)
		}
		if !contains(sql, "INNER JOIN findings") {
			t.Errorf("expected findings join, got: %s", sql)
		}
		if len(args) != 2 {
			t.Errorf("expected 2 args, got %d", len(args))
		}
	})

	t.Run("mixed primary and related entity filters", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "is_active", Operator: "eq", Value: true},
				{Field: "software.vendor", Operator: "eq", Value: "CrowdStrike", Negate: true},
			},
			Limit: query.IntPtr(50),
		}
		sql, args, err := translator.Translate("assets", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(sql, "assets.is_active = $1") {
			t.Errorf("expected primary entity filter, got: %s", sql)
		}
		if !contains(sql, "NOT EXISTS") {
			t.Errorf("expected NOT EXISTS, got: %s", sql)
		}
		if len(args) != 3 { // is_active, vendor, limit
			t.Errorf("expected 3 args, got %d: %v", len(args), args)
		}
	})

	t.Run("aggregation on primary entity", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "is_active", Operator: "eq", Value: true},
			},
			Aggregations: []query.Aggregation{
				{Type: "group_by", Field: "hostname_norm"},
				{Type: "count"},
			},
		}
		sql, _, err := translator.Translate("assets", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(sql, "GROUP BY") {
			t.Errorf("expected GROUP BY, got: %s", sql)
		}
		if !contains(sql, "COUNT") {
			t.Errorf("expected COUNT, got: %s", sql)
		}
	})

	t.Run("sort on primary entity field", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "is_active", Operator: "eq", Value: true},
			},
			Sort: []query.Sort{
				{Field: "canonical_name", Order: "asc"},
			},
		}
		sql, _, err := translator.Translate("assets", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(sql, "ORDER BY assets.canonical_name ASC") {
			t.Errorf("expected ORDER BY, got: %s", sql)
		}
	})

	t.Run("sort on related entity field", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "software.vendor", Operator: "eq", Value: "Microsoft"},
			},
			Sort: []query.Sort{
				{Field: "software.product_name", Order: "asc"},
			},
		}
		sql, _, err := translator.Translate("assets", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(sql, "ORDER BY software_inventory.product_name ASC") {
			t.Errorf("expected ORDER BY on joined field, got: %s", sql)
		}
	})

	t.Run("unknown entity prefix returns error", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "unknown.field", Operator: "eq", Value: "test"},
			},
		}
		_, _, err := translator.Translate("assets", q)
		if err == nil {
			t.Error("expected error for unknown entity prefix")
		}
		if !contains(err.Error(), "unknown entity prefix") {
			t.Errorf("expected 'unknown entity prefix' error, got: %v", err)
		}
	})
}

func TestDotWalkingValidator(t *testing.T) {
	validator := query.NewValidator()

	t.Run("valid dot notation field", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "software.vendor", Operator: "eq", Value: "CrowdStrike"},
			},
		}
		err := validator.Validate("assets", q)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid entity prefix", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "unknown.field", Operator: "eq", Value: "test"},
			},
		}
		err := validator.Validate("assets", q)
		if err == nil {
			t.Error("expected error for unknown entity prefix")
		}
	})

	t.Run("invalid field for entity", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "software.unknown_field", Operator: "eq", Value: "test"},
			},
		}
		err := validator.Validate("assets", q)
		if err == nil {
			t.Error("expected error for invalid field")
		}
	})

	t.Run("valid aggregation with dot notation", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "is_active", Operator: "eq", Value: true},
			},
			Aggregations: []query.Aggregation{
				{Type: "count", Field: "software.product_name"},
			},
		}
		err := validator.Validate("assets", q)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid sort with dot notation", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "is_active", Operator: "eq", Value: true},
			},
			Sort: []query.Sort{
				{Field: "findings.severity", Order: "desc"},
			},
		}
		err := validator.Validate("assets", q)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
