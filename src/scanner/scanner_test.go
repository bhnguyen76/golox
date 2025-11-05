package scanner

import (
	"testing"
	"example.com/golox/src/shared"
)

func TestScanSimpleParens(t *testing.T) {
	s := NewScanner("()")
	tokens := s.ScanTokens()

	// Expect: LEFT_PAREN, RIGHT_PAREN, EOF
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d: %v", len(tokens), tokens)
	}

	if tokens[0].Type != LEFT_PAREN {
		t.Errorf("expected first token LEFT_PAREN, got %v", tokens[0].Type)
	}
	if tokens[1].Type != RIGHT_PAREN {
		t.Errorf("expected second token RIGHT_PAREN, got %v", tokens[1].Type)
	}
	if tokens[2].Type != EOF {
		t.Errorf("expected third token EOF, got %v", tokens[2].Type)
	}
}

type expectedToken struct {
	typ    TokenType
	lexeme string
	lit    any
	line   int
}

func checkTokens(t *testing.T, src string, want []expectedToken) {
	t.Helper()

	s := NewScanner(src)
	toks := s.ScanTokens()

	// We usually expect an implicit EOF, so length should be len(want)+1.
	if len(toks) != len(want)+1 {
		t.Fatalf("for source %q expected %d tokens (+ EOF), got %d (%v)",
			src, len(want)+1, len(toks), toks)
	}

	for i, w := range want {
		got := toks[i]
		if got.Type != w.typ || got.Lexeme != w.lexeme || got.Line != w.line {
			t.Errorf("token %d mismatch for %q: got (%v, %q, line %d), want (%v, %q, line %d)",
				i, src, got.Type, got.Lexeme, got.Line, w.typ, w.lexeme, w.line)
		}
		// Literal checks only when not nil (numbers/strings)
		if w.lit != nil {
			if got.Literal != w.lit {
				t.Errorf("token %d literal mismatch for %q: got %v, want %v",
					i, src, got.Literal, w.lit)
			}
		}
	}

	if toks[len(toks)-1].Type != EOF {
		t.Errorf("last token type = %v, want EOF", toks[len(toks)-1].Type)
	}
}

func TestScanNumber(t *testing.T) {
	checkTokens(t, "123.45", []expectedToken{
		{typ: NUMBER, lexeme: "123.45", lit: 123.45, line: 1},
	})
}

func TestScanString(t *testing.T) {
	checkTokens(t, `"hello"`, []expectedToken{
		{typ: STRING, lexeme: `"hello"`, lit: "hello", line: 1},
	})
}

func TestIdentifierAndKeyword(t *testing.T) {
	// "or" should be keyword OR, "foo" should be IDENTIFIER
	checkTokens(t, "or foo", []expectedToken{
		{typ: OR,        lexeme: "or",  lit: nil, line: 1},
		{typ: IDENTIFIER, lexeme: "foo", lit: nil, line: 1},
	})
}

func TestCommentAndWhitespace(t *testing.T) {
	checkTokens(t, "  // this is a comment\n()", []expectedToken{
		{typ: LEFT_PAREN,  lexeme: "(", lit: nil, line: 2},
		{typ: RIGHT_PAREN, lexeme: ")", lit: nil, line: 2},
	})
}

func TestUnterminatedStringSetsError(t *testing.T) {
	shared.ResetErrors()
	s := NewScanner(`"unterminated`)
	_ = s.ScanTokens()

	if !shared.HadError {
		t.Fatalf("expected HadError to be true for unterminated string")
	}
}


