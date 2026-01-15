package oql

import (
	"testing"
)

func TestNodeTypeValues(t *testing.T) {
	tests := []struct {
		name     string
		nodeType NodeType
		expected int
	}{
		{"NodeTypeIdentifier", NodeTypeIdentifier, 0},
		{"NodeTypeStringLiteral", NodeTypeStringLiteral, 1},
		{"NodeTypeNumberLiteral", NodeTypeNumberLiteral, 2},
		{"NodeTypeBooleanLiteral", NodeTypeBooleanLiteral, 3},
		{"NodeTypeOperator", NodeTypeOperator, 4},
		{"NodeTypeKeyword", NodeTypeKeyword, 5},
		{"NodeTypeDot", NodeTypeDot, 6},
		{"NodeTypeComma", NodeTypeComma, 7},
		{"NodeTypeLeftParen", NodeTypeLeftParen, 8},
		{"NodeTypeRightParen", NodeTypeRightParen, 9},
		{"NodeTypeExpression", NodeTypeExpression, 10},
		{"NodeTypeComparison", NodeTypeComparison, 11},
		{"NodeTypeLogicalOp", NodeTypeLogicalOp, 12},
		{"NodeTypeSortClause", NodeTypeSortClause, 13},
		{"NodeTypeLimitClause", NodeTypeLimitClause, 14},
		{"NodeTypeOffsetClause", NodeTypeOffsetClause, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.nodeType) != tt.expected {
				t.Errorf("NodeType %s = %v, want %v", tt.name, tt.nodeType, tt.expected)
			}
		})
	}
}

func TestNodeCreation(t *testing.T) {
	pos := Position{Line: 1, Column: 5, Offset: 10}

	node := &Node{
		Type:     NodeTypeIdentifier,
		Value:    "test_field",
		Children: nil,
		Position: pos,
	}

	if node.Type != NodeTypeIdentifier {
		t.Errorf("node.Type = %v, want %v", node.Type, NodeTypeIdentifier)
	}
	if node.Value != "test_field" {
		t.Errorf("node.Value = %q, want %q", node.Value, "test_field")
	}
	if node.Position.Line != 1 {
		t.Errorf("node.Position.Line = %v, want %v", node.Position.Line, 1)
	}
}

func TestNodeWithChildren(t *testing.T) {
	parent := &Node{
		Type:  NodeTypeExpression,
		Value: "",
	}

	child1 := &Node{
		Type:  NodeTypeIdentifier,
		Value: "field1",
	}

	child2 := &Node{
		Type:  NodeTypeIdentifier,
		Value: "field2",
	}

	parent.Children = append(parent.Children, child1, child2)

	if len(parent.Children) != 2 {
		t.Errorf("len(parent.Children) = %v, want %v", len(parent.Children), 2)
	}
	if parent.Children[0].Value != "field1" {
		t.Errorf("parent.Children[0].Value = %q, want %q", parent.Children[0].Value, "field1")
	}
	if parent.Children[1].Value != "field2" {
		t.Errorf("parent.Children[1].Value = %q, want %q", parent.Children[1].Value, "field2")
	}
}

func TestASTStructure(t *testing.T) {
	limit := 100
	offset := 10

	ast := &AST{
		Filters: []*Node{
			{
				Type:  NodeTypeComparison,
				Value: "eq",
			},
		},
		Sort: []*Node{
			{
				Type:  NodeTypeSortClause,
				Value: "asc",
			},
		},
		Limit:       &limit,
		Offset:      &offset,
		SourceQuery: "status = 'open' sort asc limit 100 offset 10",
	}

	if len(ast.Filters) != 1 {
		t.Errorf("len(ast.Filters) = %v, want %v", len(ast.Filters), 1)
	}
	if len(ast.Sort) != 1 {
		t.Errorf("len(ast.Sort) = %v, want %v", len(ast.Sort), 1)
	}
	if ast.Limit == nil || *ast.Limit != 100 {
		t.Errorf("ast.Limit = %v, want %v", ast.Limit, 100)
	}
	if ast.Offset == nil || *ast.Offset != 10 {
		t.Errorf("ast.Offset = %v, want %v", ast.Offset, 10)
	}
	if ast.SourceQuery != "status = 'open' sort asc limit 100 offset 10" {
		t.Errorf("ast.SourceQuery = %q, want %q", ast.SourceQuery, "status = 'open' sort asc limit 100 offset 10")
	}
}

func TestASTWithNilLimits(t *testing.T) {
	ast := &AST{
		Filters:     []*Node{},
		Sort:        nil,
		Limit:       nil,
		Offset:      nil,
		SourceQuery: "field = 'value'",
	}

	if ast.Limit != nil {
		t.Errorf("ast.Limit = %v, want nil", ast.Limit)
	}
	if ast.Offset != nil {
		t.Errorf("ast.Offset = %v, want nil", ast.Offset)
	}
	if ast.Sort != nil {
		t.Errorf("ast.Sort = %v, want nil", ast.Sort)
	}
}

func TestPositionTracking(t *testing.T) {
	tests := []struct {
		name     string
		pos      Position
		wantLine int
		wantCol  int
		wantOff  int
	}{
		{
			name:     "zero position",
			pos:      Position{},
			wantLine: 0,
			wantCol:  0,
			wantOff:  0,
		},
		{
			name:     "middle of file",
			pos:      Position{Line: 5, Column: 10, Offset: 42},
			wantLine: 5,
			wantCol:  10,
			wantOff:  42,
		},
		{
			name:     "large file",
			pos:      Position{Line: 1000, Column: 1, Offset: 50000},
			wantLine: 1000,
			wantCol:  1,
			wantOff:  50000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pos.Line != tt.wantLine {
				t.Errorf("Position.Line = %v, want %v", tt.pos.Line, tt.wantLine)
			}
			if tt.pos.Column != tt.wantCol {
				t.Errorf("Position.Column = %v, want %v", tt.pos.Column, tt.wantCol)
			}
			if tt.pos.Offset != tt.wantOff {
				t.Errorf("Position.Offset = %v, want %v", tt.pos.Offset, tt.wantOff)
			}
		})
	}
}
