package oql

// NodeType identifies the type of AST node
type NodeType int

const (
	NodeTypeIdentifier NodeType = iota
	NodeTypeStringLiteral
	NodeTypeNumberLiteral
	NodeTypeBooleanLiteral
	NodeTypeOperator
	NodeTypeKeyword
	NodeTypeDot
	NodeTypeComma
	NodeTypeLeftParen
	NodeTypeRightParen
	NodeTypeExpression
	NodeTypeComparison
	NodeTypeLogicalOp
	NodeTypeSortClause
	NodeTypeLimitClause
	NodeTypeOffsetClause
)

// Node represents a node in the Abstract Syntax Tree
type Node struct {
	Type     NodeType
	Value    string
	Children []*Node
	Position Position // For error reporting
}

// Position tracks location in source for error messages
type Position struct {
	Line   int
	Column int
	Offset int
}

// AST represents the complete parsed query
type AST struct {
	Filters     []*Node
	Sort        []*Node
	Limit       *int
	Offset      *int
	SourceQuery string // Original OQL for error messages
}
