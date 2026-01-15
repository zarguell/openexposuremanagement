package oql

import (
	"encoding/json"
	"testing"

	"github.com/openexposuremanagement/oem/internal/services/query"
)

func TestTranslateSimpleComparison(t *testing.T) {
	ast := &BinaryExpr{
		Left:  &Identifier{Name: "hostname"},
		Op:    OpEq,
		Right: &StringLiteral{Value: "web-server-01"},
	}

	result, err := Translate(ast, "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.PrimaryEntity != "assets" {
		t.Errorf("expected primary entity 'assets', got '%s'", result.PrimaryEntity)
	}

	if len(result.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(result.Filters))
	}

	filter := result.Filters[0]
	if filter.Field != "hostname" {
		t.Errorf("expected field 'hostname', got '%s'", filter.Field)
	}
	if filter.Operator != "eq" {
		t.Errorf("expected operator 'eq', got '%s'", filter.Operator)
	}
	if filter.Value != "web-server-01" {
		t.Errorf("expected value 'web-server-01', got '%v'", filter.Value)
	}
}

func TestTranslateDotWalking(t *testing.T) {
	ast := &BinaryExpr{
		Left: &DotExpr{
			Left:  &Identifier{Name: "software"},
			Right: &Identifier{Name: "vendor"},
		},
		Op:    OpEq,
		Right: &StringLiteral{Value: "Microsoft"},
	}

	result, err := Translate(ast, "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(result.Filters))
	}

	filter := result.Filters[0]
	if filter.Field != "software.vendor" {
		t.Errorf("expected field 'software.vendor', got '%s'", filter.Field)
	}
	if filter.Value != "Microsoft" {
		t.Errorf("expected value 'Microsoft', got '%v'", filter.Value)
	}
}

func TestTranslateNotExpression(t *testing.T) {
	ast := &UnaryExpr{
		Op: OpNot,
		Expr: &BinaryExpr{
			Left:  &Identifier{Name: "is_active"},
			Op:    OpEq,
			Right: &BooleanLiteral{Value: true},
		},
	}

	result, err := Translate(ast, "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(result.Filters))
	}

	filter := result.Filters[0]
	if !filter.Negate {
		t.Error("expected Negate flag to be true")
	}
}

func TestTranslateAndExpression(t *testing.T) {
	ast := &BinaryExpr{
		Left: &BinaryExpr{
			Left:  &Identifier{Name: "hostname"},
			Op:    OpEq,
			Right: &StringLiteral{Value: "web-server"},
		},
		Op: OpAnd,
		Right: &BinaryExpr{
			Left:  &Identifier{Name: "is_active"},
			Op:    OpEq,
			Right: &BooleanLiteral{Value: true},
		},
	}

	result, err := Translate(ast, "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Filters) != 1 {
		t.Fatalf("expected 1 filter with logical group, got %d", len(result.Filters))
	}

	if result.Filters[0].Logic != "and" {
		t.Errorf("expected logic 'and', got '%s'", result.Filters[0].Logic)
	}

	if len(result.Filters[0].Filters) != 2 {
		t.Fatalf("expected 2 nested filters, got %d", len(result.Filters[0].Filters))
	}
}

func TestTranslateOrExpression(t *testing.T) {
	ast := &BinaryExpr{
		Left: &BinaryExpr{
			Left:  &Identifier{Name: "status"},
			Op:    OpEq,
			Right: &StringLiteral{Value: "open"},
		},
		Op: OpOr,
		Right: &BinaryExpr{
			Left:  &Identifier{Name: "status"},
			Op:    OpEq,
			Right: &StringLiteral{Value: "in_progress"},
		},
	}

	result, err := Translate(ast, "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Filters) != 1 {
		t.Fatalf("expected 1 filter with logical group, got %d", len(result.Filters))
	}

	if result.Filters[0].Logic != "or" {
		t.Errorf("expected logic 'or', got '%s'", result.Filters[0].Logic)
	}

	if len(result.Filters[0].Filters) != 2 {
		t.Fatalf("expected 2 nested filters, got %d", len(result.Filters[0].Filters))
	}
}

func TestTranslateLikeOperator(t *testing.T) {
	ast := &BinaryExpr{
		Left:  &Identifier{Name: "hostname"},
		Op:    OpLike,
		Right: &StringLiteral{Value: "web-%"},
	}

	result, err := Translate(ast, "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter := result.Filters[0]
	if filter.Operator != "like" {
		t.Errorf("expected operator 'like', got '%s'", filter.Operator)
	}
}

func TestTranslateInOperator(t *testing.T) {
	ast := &BinaryExpr{
		Left: &Identifier{Name: "severity"},
		Op:   OpIn,
		Right: &ArrayLiteral{
			Elements: []Expr{
				&StringLiteral{Value: "critical"},
				&StringLiteral{Value: "high"},
			},
		},
	}

	result, err := Translate(ast, "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter := result.Filters[0]
	if filter.Operator != "in" {
		t.Errorf("expected operator 'in', got '%s'", filter.Operator)
	}

	expectedValues := []interface{}{"critical", "high"}
	actualValues, ok := filter.Value.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{} for value, got %T", filter.Value)
	}

	if len(actualValues) != len(expectedValues) {
		t.Fatalf("expected %d values, got %d", len(expectedValues), len(actualValues))
	}
}

func TestTranslateNullCheck(t *testing.T) {
	ast := &UnaryExpr{
		Op:    OpNotNull,
		Expr:  &Identifier{Name: "owner_team_id"},
	}

	result, err := Translate(ast, "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter := result.Filters[0]
	if filter.Operator != "is_not_null" {
		t.Errorf("expected operator 'is_not_null', got '%s'", filter.Operator)
	}
}

func TestTranslateLimit(t *testing.T) {
	ast := &Query{
		Filter: &BinaryExpr{
			Left:  &Identifier{Name: "hostname"},
			Op:    OpLike,
			Right: &StringLiteral{Value: "web-%"},
		},
		Limit: 10,
	}

	result, err := Translate(ast, "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Limit != 10 {
		t.Errorf("expected limit 10, got %d", result.Limit)
	}
}

func TestTranslateOffset(t *testing.T) {
	ast := &Query{
		Filter: &BinaryExpr{
			Left:  &Identifier{Name: "hostname"},
			Op:    OpLike,
			Right: &StringLiteral{Value: "web-%"},
		},
		Offset: 20,
	}

	result, err := Translate(ast, "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Offset != 20 {
		t.Errorf("expected offset 20, got %d", result.Offset)
	}
}

func TestTranslateSort(t *testing.T) {
	ast := &Query{
		Filter: &BinaryExpr{
			Left:  &Identifier{Name: "hostname"},
			Op:    OpEq,
			Right: &StringLiteral{Value: "web-server"},
		},
		Sort: &SortClause{
			Field: &Identifier{Name: "hostname"},
			Order: SortAsc,
		},
	}

	result, err := Translate(ast, "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Sort == nil {
		t.Fatal("expected sort to be set")
	}

	if result.Sort.Field != "hostname" {
		t.Errorf("expected sort field 'hostname', got '%s'", result.Sort.Field)
	}

	if result.Sort.Order != "asc" {
		t.Errorf("expected sort order 'asc', got '%s'", result.Sort.Order)
	}
}

func TestTranslateComplexQuery(t *testing.T) {
	ast := &Query{
		Filter: &BinaryExpr{
			Left: &BinaryExpr{
				Left: &Identifier{Name: "hostname"},
				Op:   OpLike,
				Right: &StringLiteral{Value: "web-%"},
			},
			Op: OpAnd,
			Right: &UnaryExpr{
				Op: OpNot,
				Expr: &BinaryExpr{
					Left:  &Identifier{Name: "is_active"},
					Op:    OpEq,
					Right: &BooleanLiteral{Value: true},
				},
			},
		},
		Limit: 100,
		Sort: &SortClause{
			Field: &Identifier{Name: "hostname"},
			Order: SortAsc,
		},
	}

	result, err := Translate(ast, "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify filter structure
	if len(result.Filters) != 1 {
		t.Fatalf("expected 1 filter with logical group, got %d", len(result.Filters))
	}

	if result.Filters[0].Logic != "and" {
		t.Errorf("expected logic 'and', got '%s'", result.Filters[0].Logic)
	}

	if len(result.Filters[0].Filters) != 2 {
		t.Fatalf("expected 2 nested filters, got %d", len(result.Filters[0].Filters))
	}

	// Verify limit
	if result.Limit != 100 {
		t.Errorf("expected limit 100, got %d", result.Limit)
	}

	// Verify sort
	if result.Sort == nil {
		t.Fatal("expected sort to be set")
	}

	if result.Sort.Field != "hostname" {
		t.Errorf("expected sort field 'hostname', got '%s'", result.Sort.Field)
	}

	// Verify JSON serialization
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal result to JSON: %v", err)
	}

	var unmarshaled query.UnifiedQuery
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if unmarshaled.PrimaryEntity != "assets" {
		t.Errorf("JSON round-trip failed: expected primary entity 'assets', got '%s'", unmarshaled.PrimaryEntity)
	}
}

func TestTranslateNumericComparisons(t *testing.T) {
	tests := []struct {
		name           string
		op             Operator
		expectedOp     string
		expectedValue  interface{}
	}{
		{"less than", OpLt, "lt", 5},
		{"less than or equal", OpLte, "lte", 5},
		{"greater than", OpGt, "gt", 10},
		{"greater than or equal", OpGte, "gte", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast := &BinaryExpr{
				Left:  &Identifier{Name: "epss_score"},
				Op:    tt.op,
				Right: &NumericLiteral{Value: tt.expectedValue},
			}

			result, err := Translate(ast, "assets")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			filter := result.Filters[0]
			if filter.Operator != tt.expectedOp {
				t.Errorf("expected operator '%s', got '%s'", tt.expectedOp, filter.Operator)
			}

			if filter.Value != tt.expectedValue {
				t.Errorf("expected value %v, got %v", tt.expectedValue, filter.Value)
			}
		})
	}
}

func TestTranslateDotWalkingWithNot(t *testing.T) {
	ast := &UnaryExpr{
		Op: OpNot,
		Expr: &BinaryExpr{
			Left: &DotExpr{
				Left:  &Identifier{Name: "software"},
				Right: &Identifier{Name: "vendor"},
			},
			Op:    OpEq,
			Right: &StringLiteral{Value: "Microsoft"},
		},
	}

	result, err := Translate(ast, "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter := result.Filters[0]
	if filter.Field != "software.vendor" {
		t.Errorf("expected field 'software.vendor', got '%s'", filter.Field)
	}
	if !filter.Negate {
		t.Error("expected Negate flag to be true")
	}
}

func TestTranslateNestedLogicalExpressions(t *testing.T) {
	// (status = 'open' OR status = 'in_progress') AND severity = 'critical'
	ast := &BinaryExpr{
		Left: &BinaryExpr{
			Left: &BinaryExpr{
				Left:  &Identifier{Name: "status"},
				Op:    OpEq,
				Right: &StringLiteral{Value: "open"},
			},
			Op: OpOr,
			Right: &BinaryExpr{
				Left:  &Identifier{Name: "status"},
				Op:    OpEq,
				Right: &StringLiteral{Value: "in_progress"},
			},
		},
		Op: OpAnd,
		Right: &BinaryExpr{
			Left:  &Identifier{Name: "severity"},
			Op:    OpEq,
			Right: &StringLiteral{Value: "critical"},
		},
	}

	result, err := Translate(ast, "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Filters) != 1 {
		t.Fatalf("expected 1 filter with logical group, got %d", len(result.Filters))
	}

	if result.Filters[0].Logic != "and" {
		t.Errorf("expected logic 'and', got '%s'", result.Filters[0].Logic)
	}

	if len(result.Filters[0].Filters) != 2 {
		t.Fatalf("expected 2 nested filters, got %d", len(result.Filters[0].Filters))
	}

	// First filter should be the OR expression
	orFilter := result.Filters[0].Filters[0]
	if orFilter.Logic != "or" {
		t.Errorf("expected nested logic 'or', got '%s'", orFilter.Logic)
	}
}
