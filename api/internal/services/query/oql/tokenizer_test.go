package oql

import (
	"testing"
)

func TestTokenizeSimpleIdentifier(t *testing.T) {
	tokens, err := Tokenize("is_active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 1 {
		t.Errorf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != NodeTypeIdentifier {
		t.Errorf("expected identifier, got %v", tokens[0].Type)
	}
	if tokens[0].Value != "is_active" {
		t.Errorf("expected 'is_active', got '%s'", tokens[0].Value)
	}
}

func TestTokenizeOperator(t *testing.T) {
	tokens, err := Tokenize("is_active = true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[1].Type != NodeTypeOperator || tokens[1].Value != "=" {
		t.Errorf("expected '=' operator, got %v:%s", tokens[1].Type, tokens[1].Value)
	}
}

func TestTokenizeStringLiteral(t *testing.T) {
	tokens, err := Tokenize("vendor = 'CrowdStrike'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[2].Type != NodeTypeStringLiteral {
		t.Errorf("expected string literal, got %v", tokens[2].Type)
	}
	if tokens[2].Value != "CrowdStrike" {
		t.Errorf("expected 'CrowdStrike', got '%s'", tokens[2].Value)
	}
}

func TestTokenizeDotWalking(t *testing.T) {
	tokens, err := Tokenize("software.vendor = 'Microsoft'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 5 {
		t.Errorf("expected 5 tokens, got %d", len(tokens))
	}
	if tokens[1].Type != NodeTypeDot {
		t.Errorf("expected dot token, got %v", tokens[1].Type)
	}
}

func TestTokenizeKeywords(t *testing.T) {
	tokens, err := Tokenize("is_active = true and not software.vendor = 'X'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	foundAnd := false
	foundNot := false
	for _, tok := range tokens {
		if tok.Value == "and" && tok.Type == NodeTypeKeyword {
			foundAnd = true
		}
		if tok.Value == "not" && tok.Type == NodeTypeKeyword {
			foundNot = true
		}
	}
	if !foundAnd {
		t.Error("expected 'and' keyword")
	}
	if !foundNot {
		t.Error("expected 'not' keyword")
	}
}

func TestTokenizeLikeOperator(t *testing.T) {
	tokens, err := Tokenize("hostname like 'web%'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[1].Value != "like" || tokens[1].Type != NodeTypeKeyword {
		t.Errorf("expected 'like' keyword, got %v:%s", tokens[1].Type, tokens[1].Value)
	}
}

func TestTokenizeInOperator(t *testing.T) {
	tokens, err := Tokenize("severity in ('critical', 'high')")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 7 {
		t.Errorf("expected 7 tokens, got %d", len(tokens))
	}
}

func TestTokenizeNullCheck(t *testing.T) {
	tokens, err := Tokenize("ipv4 is null")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(tokens))
	}
}

func TestTokenizeParentheses(t *testing.T) {
	tokens, err := Tokenize("(is_active = true)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 5 {
		t.Errorf("expected 5 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != NodeTypeLeftParen {
		t.Errorf("expected left paren, got %v", tokens[0].Type)
	}
	if tokens[4].Type != NodeTypeRightParen {
		t.Errorf("expected right paren, got %v", tokens[4].Type)
	}
}

func TestTokenizeComplexQuery(t *testing.T) {
	oql := "is_active = true and not software.vendor = 'CrowdStrike' limit 100 sort canonical_name asc"
	tokens, err := Tokenize(oql)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) < 10 {
		t.Errorf("expected at least 10 tokens, got %d", len(tokens))
	}
}

func TestTokenizeNumberLiteral(t *testing.T) {
	tokens, err := Tokenize("epss_score > 0.9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[2].Type != NodeTypeNumberLiteral {
		t.Errorf("expected number literal, got %v", tokens[2].Type)
	}
}

func TestTokenizeBooleanLiteral(t *testing.T) {
	tokens, err := Tokenize("is_active = true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens[2].Type != NodeTypeBooleanLiteral {
		t.Errorf("expected boolean literal, got %v", tokens[2].Type)
	}
}

func TestTokenizeErrorUnterminatedString(t *testing.T) {
	_, err := Tokenize("vendor = 'unterminated")
	if err == nil {
		t.Error("expected error for unterminated string")
	}
}

func TestTokenizeWhitespace(t *testing.T) {
	tokens, err := Tokenize("  is_active   =   true  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(tokens))
	}
}

func TestTokenizeEmpty(t *testing.T) {
	tokens, err := Tokenize("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens, got %d", len(tokens))
	}
}
