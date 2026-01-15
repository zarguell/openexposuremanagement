package oql

import (
	"testing"
)

func TestParseSimpleComparison(t *testing.T) {
	input := "hostname = 'web01'"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(ast.Filters))
	}
	filter := ast.Filters[0]
	if filter.Type != NodeTypeExpression {
		t.Errorf("expected Expression node, got %v", filter.Type)
	}
	if len(filter.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(filter.Children))
	}
	if filter.Children[0].Value != "hostname" {
		t.Errorf("expected hostname, got %s", filter.Children[0].Value)
	}
	if filter.Children[1].Value != "=" {
		t.Errorf("expected =, got %s", filter.Children[1].Value)
	}
	if filter.Children[2].Value != "web01" {
		t.Errorf("expected web01, got %s", filter.Children[2].Value)
	}
}

func TestParseAndExpression(t *testing.T) {
	input := "hostname = 'web01' AND status = 'open'"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(ast.Filters))
	}
	filter := ast.Filters[0]
	if filter.Type != NodeTypeExpression {
		t.Errorf("expected Expression node, got %v", filter.Type)
	}
	// Should have: left expr, AND, right expr
	if len(filter.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(filter.Children))
	}
	if filter.Children[1].Value != "AND" {
		t.Errorf("expected AND operator, got %s", filter.Children[1].Value)
	}
}

func TestParseOrExpression(t *testing.T) {
	input := "status = 'open' OR status = 'fixed'"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(ast.Filters))
	}
	filter := ast.Filters[0]
	if filter.Type != NodeTypeExpression {
		t.Errorf("expected Expression node, got %v", filter.Type)
	}
	if len(filter.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(filter.Children))
	}
	if filter.Children[1].Value != "OR" {
		t.Errorf("expected OR operator, got %s", filter.Children[1].Value)
	}
}

func TestParseNotExpression(t *testing.T) {
	input := "NOT status = 'fixed'"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(ast.Filters))
	}
	filter := ast.Filters[0]
	// NOT should create a logical expression
	if filter.Children[0].Value != "NOT" {
		t.Errorf("expected NOT, got %s", filter.Children[0].Value)
	}
}

func TestParseGroupedExpression(t *testing.T) {
	input := "(hostname = 'web01' OR hostname = 'web02') AND status = 'open'"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(ast.Filters))
	}
	// The outer expression should be AND
	// Left child should be OR expression (from parens)
	filter := ast.Filters[0]
	if filter.Children[1].Value != "AND" {
		t.Errorf("expected AND as top-level operator, got %s", filter.Children[1].Value)
	}
}

func TestParseDotWalking(t *testing.T) {
	input := "software.vendor = 'Microsoft'"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(ast.Filters))
	}
	filter := ast.Filters[0]
	left := filter.Children[0]
	// Dot-walked field should have dot-separated components
	if left.Value != "software.vendor" {
		t.Errorf("expected software.vendor, got %s", left.Value)
	}
}

func TestParseLikeOperator(t *testing.T) {
	input := "hostname LIKE 'web%'"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(ast.Filters))
	}
	filter := ast.Filters[0]
	op := filter.Children[1]
	if op.Value != "LIKE" {
		t.Errorf("expected LIKE, got %s", op.Value)
	}
}

func TestParseInOperator(t *testing.T) {
	input := "status IN ('open', 'fixed', 'in_progress')"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(ast.Filters))
	}
	filter := ast.Filters[0]
	op := filter.Children[1]
	if op.Value != "IN" {
		t.Errorf("expected IN, got %s", op.Value)
	}
	// Should have list of values
	right := filter.Children[2]
	if len(right.Children) != 3 {
		t.Errorf("expected 3 values in IN list, got %d", len(right.Children))
	}
}

func TestParseNullCheck(t *testing.T) {
	input := "owner IS NULL"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(ast.Filters))
	}
	filter := ast.Filters[0]
	op := filter.Children[1]
	if op.Value != "IS NULL" {
		t.Errorf("expected IS NULL, got %s", op.Value)
	}
}

func TestParseNotNullCheck(t *testing.T) {
	input := "owner IS NOT NULL"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(ast.Filters))
	}
	filter := ast.Filters[0]
	op := filter.Children[1]
	if op.Value != "IS NOT NULL" {
		t.Errorf("expected IS NOT NULL, got %s", op.Value)
	}
}

func TestParseLimitClause(t *testing.T) {
	input := "hostname = 'web01' LIMIT 10"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ast.Limit == nil {
		t.Fatal("expected limit to be set")
	}
	if *ast.Limit != 10 {
		t.Errorf("expected limit 10, got %d", *ast.Limit)
	}
}

func TestParseOffsetClause(t *testing.T) {
	input := "hostname = 'web01' OFFSET 20"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ast.Offset == nil {
		t.Fatal("expected offset to be set")
	}
	if *ast.Offset != 20 {
		t.Errorf("expected offset 20, got %d", *ast.Offset)
	}
}

func TestParseSortClause(t *testing.T) {
	input := "status = 'open' SORT hostname ASC"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Sort) != 1 {
		t.Fatalf("expected 1 sort clause, got %d", len(ast.Sort))
	}
	sort := ast.Sort[0]
	if sort.Children[0].Value != "hostname" {
		t.Errorf("expected hostname field, got %s", sort.Children[0].Value)
	}
	if sort.Children[1].Value != "ASC" {
		t.Errorf("expected ASC direction, got %s", sort.Children[1].Value)
	}
}

func TestParseMultipleSortClauses(t *testing.T) {
	input := "status = 'open' SORT hostname ASC, last_seen DESC"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Sort) != 2 {
		t.Fatalf("expected 2 sort clauses, got %d", len(ast.Sort))
	}
	if ast.Sort[0].Children[0].Value != "hostname" {
		t.Errorf("expected first sort on hostname, got %s", ast.Sort[0].Children[0].Value)
	}
	if ast.Sort[1].Children[0].Value != "last_seen" {
		t.Errorf("expected second sort on last_seen, got %s", ast.Sort[1].Children[0].Value)
	}
}

func TestParseComplexQuery(t *testing.T) {
	input := "(hostname = 'web01' OR hostname = 'web02') AND status = 'open' SORT severity DESC LIMIT 10"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(ast.Filters))
	}
	if len(ast.Sort) != 1 {
		t.Fatalf("expected 1 sort clause, got %d", len(ast.Sort))
	}
	if ast.Limit == nil {
		t.Fatal("expected limit to be set")
	}
	if *ast.Limit != 10 {
		t.Errorf("expected limit 10, got %d", *ast.Limit)
	}
}

func TestParseErrorUnmatchedParen(t *testing.T) {
	input := "(hostname = 'web01' AND status = 'open'"
	_, err := Parse(input)
	if err == nil {
		t.Fatal("expected error for unmatched parenthesis")
	}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestParseErrorMissingOperator(t *testing.T) {
	input := "hostname 'web01'"
	_, err := Parse(input)
	if err == nil {
		t.Fatal("expected error for missing operator")
	}
}

func TestParseErrorEmpty(t *testing.T) {
	input := ""
	_, err := Parse(input)
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParsePrecedence(t *testing.T) {
	// NOT should bind tighter than AND
	input := "NOT status = 'fixed' AND hostname = 'web01'"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should parse as (NOT status = 'fixed') AND hostname = 'web01'
	filter := ast.Filters[0]
	if filter.Children[1].Value != "AND" {
		t.Errorf("expected AND at top level, got %s", filter.Children[1].Value)
	}
	// Left child should be NOT expression
	leftExpr := filter.Children[0]
	if len(leftExpr.Children) == 0 {
		t.Fatal("left expression has no children")
	}
	if leftExpr.Children[0].Value != "NOT" {
		t.Errorf("expected NOT in left expression, got %s", leftExpr.Children[0].Value)
	}
}

func TestParseAndOrPrecedence(t *testing.T) {
	// AND should bind tighter than OR
	input := "a = 1 OR b = 2 AND c = 3"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should parse as a = 1 OR (b = 2 AND c = 3)
	filter := ast.Filters[0]
	if filter.Children[1].Value != "OR" {
		t.Errorf("expected OR at top level, got %s", filter.Children[1].Value)
	}
	// Right child should be AND expression
	rightExpr := filter.Children[2]
	if rightExpr.Children[1].Value != "AND" {
		t.Errorf("expected AND in right expression, got %s", rightExpr.Children[1].Value)
	}
}

func TestParseDeepDotWalking(t *testing.T) {
	input := "software.vendor.title = 'Microsoft Windows'"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	filter := ast.Filters[0]
	left := filter.Children[0]
	if left.Value != "software.vendor.title" {
		t.Errorf("expected software.vendor.title, got %s", left.Value)
	}
}

func TestParseComplexBooleanLogic(t *testing.T) {
	input := "(a = 1 AND b = 2) OR (c = 3 AND d = 4)"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(ast.Filters))
	}
}

func TestParseMultipleNot(t *testing.T) {
	input := "NOT NOT status = 'fixed'"
	ast, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(ast.Filters))
	}
}
