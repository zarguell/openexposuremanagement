package oql

import (
	"testing"
)

func TestParseAndTranslateSimpleQuery(t *testing.T) {
	oql := "is_active = true"
	q, err := ParseOQL(oql)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(q.Filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(q.Filters))
	}

	filter := q.Filters[0]
	if filter.Field != "is_active" {
		t.Errorf("expected field 'is_active', got '%s'", filter.Field)
	}
	if filter.Operator != "eq" {
		t.Errorf("expected operator 'eq', got '%s'", filter.Operator)
	}
	if filter.Value != true {
		t.Errorf("expected value true, got %v", filter.Value)
	}

	if q.PrimaryEntity != "assets" {
		t.Errorf("expected primary entity 'assets', got '%s'", q.PrimaryEntity)
	}
}

func TestParseAndTranslateComplexQuery(t *testing.T) {
	oql := "is_active = true and not software.vendor = 'CrowdStrike' limit 100"
	q, err := ParseOQL(oql)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// AND expression creates a single filter with nested filters
	if len(q.Filters) != 1 {
		t.Errorf("expected 1 filter (AND expression), got %d", len(q.Filters))
	}

	// Check that it's a logical AND filter
	if q.Filters[0].Logic != "and" {
		t.Errorf("expected filter logic 'and', got '%s'", q.Filters[0].Logic)
	}

	// Check nested filters
	if len(q.Filters[0].Filters) != 2 {
		t.Errorf("expected 2 nested filters, got %d", len(q.Filters[0].Filters))
	}

	if q.Limit != 100 {
		t.Errorf("expected limit 100, got %d", q.Limit)
	}

	if q.PrimaryEntity != "assets" {
		t.Errorf("expected primary entity 'assets', got '%s'", q.PrimaryEntity)
	}
}

func TestParseOQLEmptyInput(t *testing.T) {
	_, err := ParseOQL("")
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

func TestParseOQLInvalidSyntax(t *testing.T) {
	_, err := ParseOQL("is_active = = true")
	if err == nil {
		t.Fatal("expected error for invalid syntax, got nil")
	}
}

func TestParseOQLWithDotWalking(t *testing.T) {
	oql := "software.vendor = 'Microsoft'"
	q, err := ParseOQL(oql)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(q.Filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(q.Filters))
	}

	filter := q.Filters[0]
	if filter.Field != "software.vendor" {
		t.Errorf("expected field 'software.vendor', got '%s'", filter.Field)
	}
	if filter.Value != "Microsoft" {
		t.Errorf("expected value 'Microsoft', got %v", filter.Value)
	}
}

func TestParseOQLWithNotOperator(t *testing.T) {
	oql := "not is_active = false"
	q, err := ParseOQL(oql)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(q.Filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(q.Filters))
	}

	filter := q.Filters[0]
	if !filter.Negate {
		t.Error("expected Negate flag to be true for NOT expression")
	}
}

func TestParseOQLWithLimitAndOffset(t *testing.T) {
	oql := "is_active = true limit 50 offset 10"
	q, err := ParseOQL(oql)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if q.Limit != 50 {
		t.Errorf("expected limit 50, got %d", q.Limit)
	}
	if q.Offset != 10 {
		t.Errorf("expected offset 10, got %d", q.Offset)
	}
}

func TestParseOQLWithSort(t *testing.T) {
	oql := "is_active = true sort hostname desc"
	q, err := ParseOQL(oql)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if q.Sort == nil {
		t.Fatal("expected Sort to be non-nil")
	}
	if q.Sort.Field != "hostname" {
		t.Errorf("expected sort field 'hostname', got '%s'", q.Sort.Field)
	}
	if q.Sort.Order != "desc" {
		t.Errorf("expected sort order 'desc', got '%s'", q.Sort.Order)
	}
}

func TestParseOQLWithInOperator(t *testing.T) {
	oql := "hostname in ('web-01', 'web-02', 'web-03')"
	q, err := ParseOQL(oql)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(q.Filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(q.Filters))
	}

	filter := q.Filters[0]
	if filter.Operator != "in" {
		t.Errorf("expected operator 'in', got '%s'", filter.Operator)
	}

	values, ok := filter.Value.([]interface{})
	if !ok {
		t.Fatalf("expected value to be []interface{}, got %T", filter.Value)
	}
	if len(values) != 3 {
		t.Errorf("expected 3 values in IN list, got %d", len(values))
	}
}
