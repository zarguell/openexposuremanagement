package oql

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEmptyInput       = errors.New("empty input")
	ErrUnexpectedToken  = errors.New("unexpected token")
	ErrUnmatchedParen   = errors.New("unmatched parenthesis")
	ErrMissingOperand   = errors.New("missing operand")
	ErrInvalidExpression = errors.New("invalid expression")
)

// Parser represents an OQL parser
type Parser struct {
	tokens  []Token
	pos     int
	current Token
}

// Parse parses an OQL query string into an AST
func Parse(input string) (*AST, error) {
	tokens, err := Tokenize(input)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, ErrEmptyInput
	}

	p := &Parser{
		tokens: tokens,
		pos:    -1, // Start before first token
	}
	p.advance() // This will set pos to 0 and current to tokens[0]

	ast := &AST{
		SourceQuery: input,
	}

	// Parse filters
	ast.Filters, err = p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Parse optional clauses (SORT, LIMIT, OFFSET)
	for {
		if p.isAtEnd() {
			break
		}

		if p.matchKeyword("limit") {
			limit, err := p.parseLimit()
			if err != nil {
				return nil, err
			}
			ast.Limit = limit
		} else if p.matchKeyword("offset") {
			offset, err := p.parseOffset()
			if err != nil {
				return nil, err
			}
			ast.Offset = offset
		} else if p.matchKeyword("sort") {
			sorts, err := p.parseSort()
			if err != nil {
				return nil, err
			}
			ast.Sort = sorts
		} else {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedToken, p.current)
		}
	}

	return ast, nil
}

// parseExpression parses expressions with operator precedence
// Precedence (lowest to highest): OR < AND < NOT < comparison
func (p *Parser) parseExpression() ([]*Node, error) {
	return p.parseOrExpression()
}

// parseOrExpression handles OR operators (lowest precedence)
func (p *Parser) parseOrExpression() ([]*Node, error) {
	left, err := p.parseAndExpression()
	if err != nil {
		return nil, err
	}

	for p.matchKeyword("or") {
		op := p.previous()
		right, err := p.parseAndExpression()
		if err != nil {
			return nil, err
		}

		// Create new OR expression node
		orExpr := &Node{
			Type: NodeTypeExpression,
			Children: []*Node{
				left[0],
				{Type: NodeTypeLogicalOp, Value: op.Value},
				right[0],
			},
		}
		left = []*Node{orExpr}
	}

	return left, nil
}

// parseAndExpression handles AND operators
func (p *Parser) parseAndExpression() ([]*Node, error) {
	left, err := p.parseNotExpression()
	if err != nil {
		return nil, err
	}

	for p.matchKeyword("and") {
		op := p.previous()
		right, err := p.parseNotExpression()
		if err != nil {
			return nil, err
		}

		// Create new AND expression node
		andExpr := &Node{
			Type: NodeTypeExpression,
			Children: []*Node{
				left[0],
				{Type: NodeTypeLogicalOp, Value: op.Value},
				right[0],
			},
		}
		left = []*Node{andExpr}
	}

	return left, nil
}

// parseNotExpression handles NOT operators
func (p *Parser) parseNotExpression() ([]*Node, error) {
	if p.matchKeyword("not") {
		op := p.previous()
		// Recursively parse NOT to allow "NOT NOT x"
		// But we need to call parseNotExpression again to catch chained NOTs
		// After NOT, we expect either another NOT or a primary expression
		var operand []*Node
		var err error

		// Check if another NOT follows
		if p.check(NodeTypeKeyword) && strings.EqualFold(p.current.Value, "not") {
			operand, err = p.parseNotExpression()
		} else {
			operand, err = p.parsePrimary()
		}

		if err != nil {
			return nil, err
		}

		// Create NOT expression with operand as children
		return []*Node{
			{
				Type: NodeTypeExpression,
				Children: append([]*Node{
					{Type: NodeTypeLogicalOp, Value: op.Value},
				}, operand[0].Children...),
			},
		}, nil
	}

	return p.parsePrimary()
}

// parsePrimary handles primary expressions: comparisons, parentheses
func (p *Parser) parsePrimary() ([]*Node, error) {
	// Parenthesized expression
	if p.match(NodeTypeLeftParen) {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if !p.match(NodeTypeRightParen) {
			return nil, ErrUnmatchedParen
		}
		return expr, nil
	}

	// Comparison expression
	return p.parseComparison()
}

// parseComparison parses comparison expressions
func (p *Parser) parseComparison() ([]*Node, error) {
	// Left side: identifier or dot-walked field
	left, err := p.parseField()
	if err != nil {
		return nil, err
	}

	// Operator
	if !p.match(NodeTypeOperator, NodeTypeKeyword) {
		return nil, fmt.Errorf("%w: expected operator, got %v", ErrMissingOperand, p.current)
	}

	op := p.previous()
	opValue := op.Value

	// Handle IS NULL and IS NOT NULL
	if strings.EqualFold(opValue, "is") {
		if p.matchKeyword("not") {
			opValue = "IS NOT NULL"
			if !p.matchKeyword("null") {
				return nil, fmt.Errorf("%w: expected NULL after IS NOT", ErrInvalidExpression)
			}
		} else if p.matchKeyword("null") {
			opValue = "IS NULL"
		} else {
			return nil, fmt.Errorf("%w: expected NULL or NOT NULL after IS", ErrInvalidExpression)
		}

		return []*Node{
			{
				Type:  NodeTypeExpression,
				Children: []*Node{
					left,
					{Type: NodeTypeOperator, Value: opValue},
				},
			},
		}, nil
	}

	// Handle IN operator with list
	if strings.EqualFold(opValue, "in") {
		if !p.match(NodeTypeLeftParen) {
			return nil, fmt.Errorf("%w: expected ( after IN", ErrInvalidExpression)
		}

		values := []*Node{}
		for !p.check(NodeTypeRightParen) && !p.isAtEnd() {
			val, err := p.parseValue()
			if err != nil {
				return nil, err
			}
			values = append(values, val)

			// Consume comma if present
			p.match(NodeTypeComma)
		}

		if !p.match(NodeTypeRightParen) {
			return nil, ErrUnmatchedParen
		}

		// Create list node
		listNode := &Node{
			Type:     NodeTypeExpression,
			Children: values,
		}

		return []*Node{
			{
				Type:  NodeTypeExpression,
				Children: []*Node{
					left,
					{Type: NodeTypeOperator, Value: "IN"},
					listNode,
				},
			},
		}, nil
	}

	// Right side: value or field
	right, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	return []*Node{
		{
			Type:  NodeTypeExpression,
			Children: []*Node{
				left,
				{Type: NodeTypeOperator, Value: opValue},
				right,
			},
		},
	}, nil
}

// parseField parses an identifier or dot-walked field
func (p *Parser) parseField() (*Node, error) {
	if !p.match(NodeTypeIdentifier) {
		return nil, fmt.Errorf("%w: expected identifier, got %v", ErrUnexpectedToken, p.current)
	}

	field := p.previous().Value

	// Handle dot-walking
	for p.match(NodeTypeDot) {
		if !p.match(NodeTypeIdentifier) {
			return nil, fmt.Errorf("%w: expected identifier after dot", ErrUnexpectedToken)
		}
		field += "." + p.previous().Value
	}

	return &Node{
		Type:  NodeTypeIdentifier,
		Value: field,
	}, nil
}

// parseValue parses a literal value
func (p *Parser) parseValue() (*Node, error) {
	if p.match(NodeTypeStringLiteral) {
		return &Node{
			Type:  NodeTypeStringLiteral,
			Value: p.previous().Value,
		}, nil
	}

	if p.match(NodeTypeNumberLiteral) {
		return &Node{
			Type:  NodeTypeNumberLiteral,
			Value: p.previous().Value,
		}, nil
	}

	if p.match(NodeTypeBooleanLiteral) {
		return &Node{
			Type:  NodeTypeBooleanLiteral,
			Value: p.previous().Value,
		}, nil
	}

	return nil, fmt.Errorf("%w: expected value, got %v", ErrUnexpectedToken, p.current)
}

// parseLimit parses LIMIT clause
func (p *Parser) parseLimit() (*int, error) {
	if !p.match(NodeTypeNumberLiteral) {
		return nil, fmt.Errorf("%w: expected number after LIMIT", ErrUnexpectedToken)
	}

	limit := parseInt(p.previous().Value)
	return &limit, nil
}

// parseOffset parses OFFSET clause
func (p *Parser) parseOffset() (*int, error) {
	if !p.match(NodeTypeNumberLiteral) {
		return nil, fmt.Errorf("%w: expected number after OFFSET", ErrUnexpectedToken)
	}

	offset := parseInt(p.previous().Value)
	return &offset, nil
}

// parseSort parses SORT clause(s)
func (p *Parser) parseSort() ([]*Node, error) {
	var sorts []*Node

	for {
		// Parse field
		if !p.match(NodeTypeIdentifier) {
			return nil, fmt.Errorf("%w: expected field after SORT", ErrUnexpectedToken)
		}
		field := p.previous().Value

		// Handle dot-walking in sort
		for p.match(NodeTypeDot) {
			if !p.match(NodeTypeIdentifier) {
				return nil, fmt.Errorf("%w: expected identifier after dot", ErrUnexpectedToken)
			}
			field += "." + p.previous().Value
		}

		// Parse direction (optional, defaults to ASC)
		var direction string = "ASC"
		if p.matchKeyword("asc") {
			direction = "ASC"
		} else if p.matchKeyword("desc") {
			direction = "DESC"
		}

		sorts = append(sorts, &Node{
			Type: NodeTypeSortClause,
			Children: []*Node{
				{Type: NodeTypeIdentifier, Value: field},
				{Type: NodeTypeKeyword, Value: direction},
			},
		})

		// Check for comma (multiple sort clauses)
		if !p.match(NodeTypeComma) {
			break
		}
	}

	return sorts, nil
}

// Helper methods

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.pos++
	}
	if p.pos < len(p.tokens) {
		p.current = p.tokens[p.pos]
	}
	return p.current
}

func (p *Parser) match(types ...NodeType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) matchKeyword(keyword string) bool {
	if p.check(NodeTypeKeyword) && strings.EqualFold(p.current.Value, keyword) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) check(t NodeType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.current.Type == t
}

func (p *Parser) isAtEnd() bool {
	return p.pos >= len(p.tokens)
}

func (p *Parser) previous() Token {
	if p.pos > 0 {
		return p.tokens[p.pos-1]
	}
	return Token{}
}

func parseInt(s string) int {
	var result int
	for _, ch := range s {
		result = result*10 + int(ch-'0')
	}
	return result
}
