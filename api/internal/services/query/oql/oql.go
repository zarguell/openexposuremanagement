package oql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/openexposuremanagement/oem/internal/services/query"
)

// ParseOQL is the main entry point for parsing OQL queries
// It returns a UnifiedQuery struct that can be executed by the existing query infrastructure
func ParseOQL(input string) (*query.UnifiedQuery, error) {
	// Tokenize and parse into AST
	ast, err := Parse(input)
	if err != nil {
		return nil, err
	}

	// Convert parser AST to translator Expr types
	expr, err := convertNodeToExpr(ast.Filters)
	if err != nil {
		return nil, fmt.Errorf("converting AST to expression: %w", err)
	}

	// Create Query struct with all clauses
	queryNode := &Query{
		Filter: expr,
	}

	// Add LIMIT if present
	if ast.Limit != nil {
		queryNode.Limit = *ast.Limit
	}

	// Add OFFSET if present
	if ast.Offset != nil {
		queryNode.Offset = *ast.Offset
	}

	// Add SORT if present
	if len(ast.Sort) > 0 {
		sortClause, err := convertSortClause(ast.Sort[0])
		if err != nil {
			return nil, fmt.Errorf("converting sort clause: %w", err)
		}
		queryNode.Sort = sortClause
	}

	// Translate to UnifiedQuery
	// Use "assets" as the default primary entity for MVP
	jsonQuery, err := Translate(queryNode, "assets")
	if err != nil {
		return nil, fmt.Errorf("translating to unified query: %w", err)
	}

	return jsonQuery, nil
}

// convertNodeToExpr converts a parser Node to a translator Expr
func convertNodeToExpr(nodes []*Node) (Expr, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("empty node list")
	}

	// For single expression node, convert it directly
	if len(nodes) == 1 && nodes[0].Type == NodeTypeExpression {
		return convertExpressionNode(nodes[0])
	}

	// For comparison node, convert it directly
	if len(nodes) == 1 && nodes[0].Type == NodeTypeComparison {
		return convertExpressionNode(nodes[0])
	}

	// Multiple nodes at top level - wrap in implicit AND
	if len(nodes) > 1 {
		var left Expr
		var err error

		left, err = convertExpressionNode(nodes[0])
		if err != nil {
			return nil, err
		}

		for i := 1; i < len(nodes); i++ {
			right, err := convertExpressionNode(nodes[i])
			if err != nil {
				return nil, err
			}

			left = &BinaryExpr{
				Left:  left,
				Op:    OpAnd,
				Right: right,
			}
		}

		return left, nil
	}

	return convertExpressionNode(nodes[0])
}

// convertExpressionNode converts an expression node to an Expr
func convertExpressionNode(node *Node) (Expr, error) {
	if node == nil {
		return nil, fmt.Errorf("nil node")
	}

	if node.Type != NodeTypeExpression && node.Type != NodeTypeComparison {
		return nil, fmt.Errorf("expected expression or comparison node, got %v", node.Type)
	}

	if len(node.Children) == 0 {
		return nil, fmt.Errorf("expression node has no children")
	}

	// Check for logical operator (AND, OR, NOT) at any position
	for _, child := range node.Children {
		if child.Type == NodeTypeLogicalOp {
			return convertLogicalExpression(node)
		}
	}

	// Check for comparison operator (typically at position 1)
	if len(node.Children) >= 3 && node.Children[1].Type == NodeTypeOperator {
		return convertComparisonExpression(node)
	}

	// Single child - recurse
	if len(node.Children) == 1 {
		return convertExpressionNode(node.Children[0])
	}

	return nil, fmt.Errorf("unsupported expression structure: %d children, types: %v",
		len(node.Children), getNodeTypes(node.Children))
}

// convertLogicalExpression converts a logical expression (AND, OR, NOT)
func convertLogicalExpression(node *Node) (Expr, error) {
	if len(node.Children) < 2 {
		return nil, fmt.Errorf("logical expression must have at least 2 children")
	}

	// Find the operator node
	var opIndex int = -1
	for i, child := range node.Children {
		if child.Type == NodeTypeLogicalOp {
			opIndex = i
			break
		}
	}

	if opIndex == -1 {
		return nil, fmt.Errorf("no logical operator found in expression")
	}

	opNode := node.Children[opIndex]
	op := strings.ToLower(opNode.Value)

	// Handle NOT operator
	if op == "not" {
		// NOT expressions have structure: [NOT, operand, ...]
		// where operand is at index 1
		if len(node.Children) < 2 {
			return nil, fmt.Errorf("NOT expression must have an operand")
		}

		// The operand is the next child after NOT
		operandIndex := opIndex + 1
		if operandIndex >= len(node.Children) {
			return nil, fmt.Errorf("NOT expression missing operand")
		}

		var operand *Node
		if len(node.Children) == 2 {
			// Simple case: NOT followed by single expression
			operand = node.Children[1]
		} else {
			// Multiple children after NOT - wrap them in an expression node
			operand = &Node{
				Type:     NodeTypeExpression,
				Children: node.Children[operandIndex:],
			}
		}

		expr, err := convertSingleNode(operand)
		if err != nil {
			return nil, fmt.Errorf("converting NOT operand: %w", err)
		}

		return &UnaryExpr{
			Op:   OpNot,
			Expr: expr,
		}, nil
	}

	// Handle AND/OR operators
	if len(node.Children) != 3 {
		return nil, fmt.Errorf("%s expression must have 3 children (left, op, right)", op)
	}

	left, err := convertSingleNode(node.Children[0])
	if err != nil {
		return nil, fmt.Errorf("converting left operand: %w", err)
	}

	right, err := convertSingleNode(node.Children[2])
	if err != nil {
		return nil, fmt.Errorf("converting right operand: %w", err)
	}

	var binaryOp Operator
	if op == "and" {
		binaryOp = OpAnd
	} else if op == "or" {
		binaryOp = OpOr
	} else {
		return nil, fmt.Errorf("unsupported logical operator: %s", op)
	}

	return &BinaryExpr{
		Left:  left,
		Op:    binaryOp,
		Right: right,
	}, nil
}

// convertComparisonExpression converts a comparison expression
func convertComparisonExpression(node *Node) (Expr, error) {
	if len(node.Children) < 2 {
		return nil, fmt.Errorf("comparison must have at least 2 children")
	}

	left, err := convertSingleNode(node.Children[0])
	if err != nil {
		return nil, fmt.Errorf("converting left side: %w", err)
	}

	opNode := node.Children[1]
	if opNode.Type != NodeTypeOperator {
		return nil, fmt.Errorf("expected operator node, got %v", opNode.Type)
	}

	op, err := parseOperator(opNode.Value)
	if err != nil {
		return nil, err
	}

	// Handle IS NULL and IS NOT NULL (unary operators)
	if op == OpIsNull || op == OpNotNull {
		return &UnaryExpr{
			Op:   op,
			Expr: left,
		}, nil
	}

	// Handle IN operator with list
	if op == OpIn && len(node.Children) >= 3 {
		// Third child should be the list expression
		if node.Children[2].Type != NodeTypeExpression {
			return nil, fmt.Errorf("expected list expression for IN operator")
		}

		elements := make([]Expr, 0)
		for _, child := range node.Children[2].Children {
			expr, err := convertLiteralNode(child)
			if err != nil {
				return nil, fmt.Errorf("converting IN list element: %w", err)
			}
			elements = append(elements, expr)
		}

		return &BinaryExpr{
			Left:  left,
			Op:    op,
			Right: &ArrayLiteral{Elements: elements},
		}, nil
	}

	// Handle regular binary comparison
	if len(node.Children) < 3 {
		return nil, fmt.Errorf("comparison must have a right operand")
	}

	right, err := convertSingleNode(node.Children[2])
	if err != nil {
		return nil, fmt.Errorf("converting right side: %w", err)
	}

	return &BinaryExpr{
		Left:  left,
		Op:    op,
		Right: right,
	}, nil
}

// convertSingleNode converts a single node to an Expr
func convertSingleNode(node *Node) (Expr, error) {
	if node == nil {
		return nil, fmt.Errorf("nil node")
	}

	switch node.Type {
	case NodeTypeIdentifier:
		return convertIdentifier(node), nil

	case NodeTypeStringLiteral:
		return &StringLiteral{Value: unquoteString(node.Value)}, nil

	case NodeTypeNumberLiteral:
		return convertNumberLiteral(node.Value)

	case NodeTypeBooleanLiteral:
		return &BooleanLiteral{Value: strings.ToLower(node.Value) == "true"}, nil

	case NodeTypeExpression, NodeTypeComparison:
		return convertExpressionNode(node)

	default:
		return nil, fmt.Errorf("unsupported node type: %v", node.Type)
	}
}

// convertIdentifier converts an identifier node to an Identifier Expr
func convertIdentifier(node *Node) Expr {
	// Handle dot-walking
	if strings.Contains(node.Value, ".") {
		parts := strings.Split(node.Value, ".")
		if len(parts) == 2 {
			return &DotExpr{
				Left:  &Identifier{Name: parts[0]},
				Right: &Identifier{Name: parts[1]},
			}
		}
		// For deeper nesting, create nested DotExpr
		var left Expr = &Identifier{Name: parts[0]}
		for i := 1; i < len(parts); i++ {
			left = &DotExpr{
				Left:  left,
				Right: &Identifier{Name: parts[i]},
			}
		}
		return left
	}

	return &Identifier{Name: node.Value}
}

// convertLiteralNode converts a literal node to a literal Expr
func convertLiteralNode(node *Node) (Expr, error) {
	switch node.Type {
	case NodeTypeIdentifier:
		return convertIdentifier(node), nil
	case NodeTypeStringLiteral:
		return &StringLiteral{Value: unquoteString(node.Value)}, nil
	case NodeTypeNumberLiteral:
		return convertNumberLiteral(node.Value)
	case NodeTypeBooleanLiteral:
		return &BooleanLiteral{Value: strings.ToLower(node.Value) == "true"}, nil
	default:
		return nil, fmt.Errorf("expected literal node, got %v", node.Type)
	}
}

// convertNumberLiteral converts a number string to appropriate numeric type
func convertNumberLiteral(value string) (Expr, error) {
	// Try integer first
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return &NumericLiteral{Value: int(i)}, nil
	}

	// Try float
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return &NumericLiteral{Value: f}, nil
	}

	return nil, fmt.Errorf("invalid number literal: %s", value)
}

// parseOperator converts operator string to Operator enum
func parseOperator(op string) (Operator, error) {
	switch strings.ToLower(op) {
	case "=", "==":
		return OpEq, nil
	case "!=":
		return OpNe, nil
	case "<":
		return OpLt, nil
	case "<=":
		return OpLte, nil
	case ">":
		return OpGt, nil
	case ">=":
		return OpGte, nil
	case "like":
		return OpLike, nil
	case "in":
		return OpIn, nil
	case "is null":
		return OpIsNull, nil
	case "is not null":
		return OpNotNull, nil
	default:
		return -1, fmt.Errorf("unsupported operator: %s", op)
	}
}

// unquoteString removes quotes from a string literal
func unquoteString(s string) string {
	if len(s) >= 2 && (s[0] == '\'' || s[0] == '"') && s[0] == s[len(s)-1] {
		return s[1 : len(s)-1]
	}
	return s
}

// convertSortClause converts a parser sort node to a translator SortClause
func convertSortClause(node *Node) (*SortClause, error) {
	if node == nil || node.Type != NodeTypeSortClause {
		return nil, fmt.Errorf("expected sort clause node, got %v", node.Type)
	}

	if len(node.Children) < 2 {
		return nil, fmt.Errorf("sort clause must have at least 2 children (field and order)")
	}

	// First child is the field
	fieldExpr, err := convertSingleNode(node.Children[0])
	if err != nil {
		return nil, fmt.Errorf("converting sort field: %w", err)
	}

	// Second child is the direction (ASC/DESC)
	orderStr := strings.ToUpper(node.Children[1].Value)
	var order SortOrder
	if orderStr == "ASC" {
		order = SortAsc
	} else if orderStr == "DESC" {
		order = SortDesc
	} else {
		return nil, fmt.Errorf("invalid sort order: %s", orderStr)
	}

	return &SortClause{
		Field: fieldExpr,
		Order: order,
	}, nil
}

// getNodeTypes returns a slice of node types for debugging
func getNodeTypes(nodes []*Node) []NodeType {
	types := make([]NodeType, len(nodes))
	for i, node := range nodes {
		types[i] = node.Type
	}
	return types
}
