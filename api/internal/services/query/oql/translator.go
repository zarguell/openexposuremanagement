package oql

import (
	"fmt"

	"github.com/openexposuremanagement/oem/internal/services/query"
)

// AST Node Types

// Expr represents an expression in the OQL AST
type Expr interface {
	exprNode()
}

// BinaryExpr represents a binary expression (e.g., a = b, a AND b)
type BinaryExpr struct {
	Left  Expr
	Op    Operator
	Right Expr
}

func (e *BinaryExpr) exprNode() {}

// UnaryExpr represents a unary expression (e.g., NOT a, IS NULL a)
type UnaryExpr struct {
	Op   Operator
	Expr Expr
}

func (e *UnaryExpr) exprNode() {}

// Identifier represents a field or variable name
type Identifier struct {
	Name string
}

func (e *Identifier) exprNode() {}

// DotExpr represents a dot-walking expression (e.g., software.vendor)
type DotExpr struct {
	Left  Expr
	Right Expr
}

func (e *DotExpr) exprNode() {}

// StringLiteral represents a string value
type StringLiteral struct {
	Value string
}

func (e *StringLiteral) exprNode() {}

// NumericLiteral represents a numeric value
type NumericLiteral struct {
	Value interface{} // int or float64
}

func (e *NumericLiteral) exprNode() {}

// BooleanLiteral represents a boolean value
type BooleanLiteral struct {
	Value bool
}

func (e *BooleanLiteral) exprNode() {}

// ArrayLiteral represents an array of values (for IN operator)
type ArrayLiteral struct {
	Elements []Expr
}

func (e *ArrayLiteral) exprNode() {}

// Query represents the complete query structure
type Query struct {
	Filter Expr
	Limit  int
	Offset int
	Sort   *SortClause
}

func (e *Query) exprNode() {}

// SortClause represents a sort order
type SortClause struct {
	Field Expr
	Order SortOrder
}

// SortOrder represents the sort direction
type SortOrder int

const (
	SortAsc SortOrder = iota
	SortDesc
)

// Operator represents an operator token
type Operator int

const (
	OpEq Operator = iota
	OpNe
	OpLt
	OpLte
	OpGt
	OpGte
	OpLike
	OpIn
	OpAnd
	OpOr
	OpNot
	OpIsNull
	OpNotNull
)

// Translate converts an OQL AST into a unified query JSON structure
func Translate(ast Expr, primaryEntity string) (*query.UnifiedQuery, error) {
	q := &query.UnifiedQuery{
		PrimaryEntity: primaryEntity,
	}

	// Handle Query node (top-level)
	if queryNode, ok := ast.(*Query); ok {
		if queryNode.Filter != nil {
			filter, err := translateExpr(queryNode.Filter)
			if err != nil {
				return nil, fmt.Errorf("translating filter: %w", err)
			}
			q.Filters = []query.Filter{*filter}
		}

		if queryNode.Limit > 0 {
			q.Limit = queryNode.Limit
		}

		if queryNode.Offset > 0 {
			q.Offset = queryNode.Offset
		}

		if queryNode.Sort != nil {
			fieldName, err := extractFieldName(queryNode.Sort.Field)
			if err != nil {
				return nil, fmt.Errorf("translating sort field: %w", err)
			}
			order := "asc"
			if queryNode.Sort.Order == SortDesc {
				order = "desc"
			}
			q.Sort = &query.Sort{
				Field: fieldName,
				Order: order,
			}
		}

		return q, nil
	}

	// Handle direct expression (without Query wrapper)
	filter, err := translateExpr(ast)
	if err != nil {
		return nil, fmt.Errorf("translating expression: %w", err)
	}
	q.Filters = []query.Filter{*filter}

	return q, nil
}

// translateExpr converts an expression node to a filter
func translateExpr(expr Expr) (*query.Filter, error) {
	switch e := expr.(type) {
	case *BinaryExpr:
		return translateBinaryExpr(e)
	case *UnaryExpr:
		return translateUnaryExpr(e)
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// translateBinaryExpr converts a binary expression to a filter
func translateBinaryExpr(expr *BinaryExpr) (*query.Filter, error) {
	switch expr.Op {
	case OpAnd, OpOr:
		// Logical operator - translate both sides
		leftFilter, err := translateExpr(expr.Left)
		if err != nil {
			return nil, fmt.Errorf("translating left operand: %w", err)
		}

		rightFilter, err := translateExpr(expr.Right)
		if err != nil {
			return nil, fmt.Errorf("translating right operand: %w", err)
		}

		logic := "and"
		if expr.Op == OpOr {
			logic = "or"
		}

		return &query.Filter{
			Logic: logic,
			Filters: []query.Filter{
				*leftFilter,
				*rightFilter,
			},
		}, nil

	case OpEq, OpNe, OpLt, OpLte, OpGt, OpGte, OpLike, OpIn:
		// Comparison operator
		field, err := extractFieldName(expr.Left)
		if err != nil {
			return nil, fmt.Errorf("extracting field name: %w", err)
		}

		value, err := extractValue(expr.Right)
		if err != nil {
			return nil, fmt.Errorf("extracting value: %w", err)
		}

		op := operatorToString(expr.Op)

		return &query.Filter{
			Field:    field,
			Operator: op,
			Value:    value,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported binary operator: %v", expr.Op)
	}
}

// translateUnaryExpr converts a unary expression to a filter
func translateUnaryExpr(expr *UnaryExpr) (*query.Filter, error) {
	switch expr.Op {
	case OpNot:
		// NOT expression - translate inner expression and set Negate flag
		filter, err := translateExpr(expr.Expr)
		if err != nil {
			return nil, fmt.Errorf("translating NOT operand: %w", err)
		}
		filter.Negate = true
		return filter, nil

	case OpIsNull, OpNotNull:
		// IS NULL / IS NOT NULL
		field, err := extractFieldName(expr.Expr)
		if err != nil {
			return nil, fmt.Errorf("extracting field name: %w", err)
		}

		op := "is_null"
		if expr.Op == OpNotNull {
			op = "is_not_null"
		}

		return &query.Filter{
			Field:    field,
			Operator: op,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported unary operator: %v", expr.Op)
	}
}

// extractFieldName extracts the field name from an expression
// Handles dot-walking for nested fields
func extractFieldName(expr Expr) (string, error) {
	switch e := expr.(type) {
	case *Identifier:
		return e.Name, nil
	case *DotExpr:
		left, err := extractFieldName(e.Left)
		if err != nil {
			return "", err
		}
		right, err := extractFieldName(e.Right)
		if err != nil {
			return "", err
		}
		return left + "." + right, nil
	default:
		return "", fmt.Errorf("expected identifier or dot expression, got %T", expr)
	}
}

// extractValue extracts the value from an expression literal
func extractValue(expr Expr) (interface{}, error) {
	switch e := expr.(type) {
	case *StringLiteral:
		return e.Value, nil
	case *NumericLiteral:
		return e.Value, nil
	case *BooleanLiteral:
		return e.Value, nil
	case *ArrayLiteral:
		values := make([]interface{}, len(e.Elements))
		for i, elem := range e.Elements {
			val, err := extractValue(elem)
			if err != nil {
				return nil, fmt.Errorf("extracting array element %d: %w", i, err)
			}
			values[i] = val
		}
		return values, nil
	default:
		return nil, fmt.Errorf("expected literal value, got %T", expr)
	}
}

// operatorToString converts an operator to its string representation
func operatorToString(op Operator) string {
	switch op {
	case OpEq:
		return "eq"
	case OpNe:
		return "ne"
	case OpLt:
		return "lt"
	case OpLte:
		return "lte"
	case OpGt:
		return "gt"
	case OpGte:
		return "gte"
	case OpLike:
		return "like"
	case OpIn:
		return "in"
	default:
		return ""
	}
}
