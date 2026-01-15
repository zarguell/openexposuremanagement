package oql

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var (
	ErrUnterminatedString = errors.New("unterminated string literal")
	ErrInvalidCharacter   = errors.New("invalid character")
)

// Token represents a single token in the OQL query string
type Token struct {
	Type     NodeType
	Value    string
	Position int
}

// Tokenize breaks an OQL query string into tokens
func Tokenize(input string) ([]Token, error) {
	var tokens []Token
	var pos int

	input = strings.TrimSpace(input)

	for pos < len(input) {
		ch := input[pos]

		// Skip whitespace
		if unicode.IsSpace(rune(ch)) {
			pos++
			continue
		}

		// Single-character tokens
		switch ch {
		case '=':
			tokens = append(tokens, Token{Type: NodeTypeOperator, Value: "=", Position: pos})
			pos++
			continue
		case '(':
			tokens = append(tokens, Token{Type: NodeTypeLeftParen, Value: "(", Position: pos})
			pos++
			continue
		case ')':
			tokens = append(tokens, Token{Type: NodeTypeRightParen, Value: ")", Position: pos})
			pos++
			continue
		case ',':
			tokens = append(tokens, Token{Type: NodeTypeComma, Value: ",", Position: pos})
			pos++
			continue
		case '.':
			tokens = append(tokens, Token{Type: NodeTypeDot, Value: ".", Position: pos})
			pos++
			continue
		}

		// Multi-character operators
		if pos+1 < len(input) {
			twoChar := input[pos : pos+2]
			switch twoChar {
			case "==", "!=", "<=", ">=", "&&", "||":
				tokens = append(tokens, Token{Type: NodeTypeOperator, Value: twoChar, Position: pos})
				pos += 2
				continue
			}
		}

		// Single-char operators
		if ch == '<' || ch == '>' || ch == '!' {
			tokens = append(tokens, Token{Type: NodeTypeOperator, Value: string(ch), Position: pos})
			pos++
			continue
		}

		// String literal
		if ch == '\'' || ch == '"' {
			strValue, endPos, err := readStringLiteral(input, pos)
			if err != nil {
				return nil, fmt.Errorf("%w at position %d", err, pos)
			}
			tokens = append(tokens, Token{Type: NodeTypeStringLiteral, Value: strValue, Position: pos})
			pos = endPos + 1
			continue
		}

		// Number literal
		if unicode.IsDigit(rune(ch)) || (ch == '.' && pos+1 < len(input) && unicode.IsDigit(rune(input[pos+1]))) {
			numValue, endPos, err := readNumberLiteral(input, pos)
			if err != nil {
				return nil, fmt.Errorf("%w at position %d", err, pos)
			}
			tokens = append(tokens, Token{Type: NodeTypeNumberLiteral, Value: numValue, Position: pos})
			pos = endPos
			continue
		}

		// Identifier or keyword or boolean
		if unicode.IsLetter(rune(ch)) || ch == '_' {
			value, endPos := readIdentifier(input, pos)

			// Check if it's a boolean literal
			if value == "true" || value == "false" {
				tokens = append(tokens, Token{Type: NodeTypeBooleanLiteral, Value: value, Position: pos})
			} else if isKeyword(value) {
				tokens = append(tokens, Token{Type: NodeTypeKeyword, Value: value, Position: pos})
			} else {
				tokens = append(tokens, Token{Type: NodeTypeIdentifier, Value: value, Position: pos})
			}
			pos = endPos
			continue
		}

		return nil, fmt.Errorf("%w: '%c' at position %d", ErrInvalidCharacter, ch, pos)
	}

	return tokens, nil
}

// readStringLiteral reads a string literal (single or double quoted)
func readStringLiteral(input string, start int) (string, int, error) {
	quote := input[start]
	pos := start + 1
	var sb strings.Builder

	for pos < len(input) {
		ch := input[pos]

		// Check for escape sequence
		if ch == '\\' && pos+1 < len(input) {
			pos++
			ch = input[pos]
			switch ch {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\', '\'', '"':
				sb.WriteRune(rune(ch))
			default:
				sb.WriteRune('\\')
				sb.WriteRune(rune(ch))
			}
			pos++
			continue
		}

		if ch == quote {
			return sb.String(), pos, nil
		}

		sb.WriteRune(rune(ch))
		pos++
	}

	return "", -1, ErrUnterminatedString
}

// readNumberLiteral reads a number literal (integer or float)
func readNumberLiteral(input string, start int) (string, int, error) {
	pos := start
	hasDecimal := false

	for pos < len(input) {
		ch := input[pos]

		if ch == '.' {
			if hasDecimal {
				break // Second decimal point
			}
			hasDecimal = true
		} else if !unicode.IsDigit(rune(ch)) {
			break
		}

		pos++
	}

	return input[start:pos], pos, nil
}

// readIdentifier reads an identifier (alphanumeric + underscore)
func readIdentifier(input string, start int) (string, int) {
	pos := start

	for pos < len(input) {
		ch := input[pos]
		if !(unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_') {
			break
		}
		pos++
	}

	return input[start:pos], pos
}

// isKeyword checks if the given word is a reserved keyword
func isKeyword(word string) bool {
	switch strings.ToLower(word) {
	case "and", "or", "not", "like", "in", "is", "null",
		"limit", "sort", "asc", "desc":
		return true
	default:
		return false
	}
}
