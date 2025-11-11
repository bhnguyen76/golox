package scanner

import (
	"testing"
	"example.com/golox/lox/shared"
)

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

func TestScanSimpleParens(t *testing.T) {
	checkTokens(t, "()", []expectedToken{
		{typ: LEFT_PAREN,  lexeme: "(", lit: nil, line: 1},
		{typ: RIGHT_PAREN, lexeme: ")", lit: nil, line: 1},
	})
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

func TestScanSingleCharTokens(t *testing.T) {
	checkTokens(t, "(){}.,-+;*", []expectedToken{
		{typ: LEFT_PAREN,  lexeme: "(",  lit: nil, line: 1},
		{typ: RIGHT_PAREN, lexeme: ")",  lit: nil, line: 1},
		{typ: LEFT_BRACE,  lexeme: "{",  lit: nil, line: 1},
		{typ: RIGHT_BRACE, lexeme: "}",  lit: nil, line: 1},
		{typ: DOT,         lexeme: ".",  lit: nil, line: 1}, 
		{typ: COMMA,       lexeme: ",",  lit: nil, line: 1}, 
		{typ: MINUS,       lexeme: "-",  lit: nil, line: 1},
		{typ: PLUS,        lexeme: "+",  lit: nil, line: 1},
		{typ: SEMICOLON,   lexeme: ";",  lit: nil, line: 1},
		{typ: STAR,        lexeme: "*",  lit: nil, line: 1},
	})
}

func TestScanOperators(t *testing.T) {
	src := `! != = == < <= > >= /`
	checkTokens(t, src, []expectedToken{
		{typ: BANG,          lexeme: "!",   lit: nil, line: 1},
		{typ: BANG_EQUAL,    lexeme: "!=",  lit: nil, line: 1},
		{typ: EQUAL,         lexeme: "=",   lit: nil, line: 1},
		{typ: EQUAL_EQUAL,   lexeme: "==",  lit: nil, line: 1},
		{typ: LESS,          lexeme: "<",   lit: nil, line: 1},
		{typ: LESS_EQUAL,    lexeme: "<=",  lit: nil, line: 1},
		{typ: GREATER,       lexeme: ">",   lit: nil, line: 1},
		{typ: GREATER_EQUAL, lexeme: ">=",  lit: nil, line: 1},
		{typ: SLASH,         lexeme: "/",   lit: nil, line: 1},
	})
}

func TestNumberFollowedByDot(t *testing.T) {
	checkTokens(t, "123.", []expectedToken{
		{typ: NUMBER, lexeme: "123", lit: 123.0, line: 1},
		{typ: DOT,    lexeme: ".",   lit: nil,   line: 1},
	})
}

func TestIdentifiersUnderscoreAndMoreKeywords(t *testing.T) {
	src := "and class else false for fun if nil or print return super this true var while _foo Bar123"
	checkTokens(t, src, []expectedToken{
		{typ: AND,        lexeme: "and",    lit: nil, line: 1},
		{typ: CLASS,      lexeme: "class",  lit: nil, line: 1},
		{typ: ELSE,       lexeme: "else",   lit: nil, line: 1},
		{typ: FALSE,      lexeme: "false",  lit: nil, line: 1},
		{typ: FOR,        lexeme: "for",    lit: nil, line: 1},
		{typ: FUN,        lexeme: "fun",    lit: nil, line: 1},
		{typ: IF,         lexeme: "if",     lit: nil, line: 1},
		{typ: NIL,        lexeme: "nil",    lit: nil, line: 1},
		{typ: OR,         lexeme: "or",     lit: nil, line: 1},
		{typ: PRINT,      lexeme: "print",  lit: nil, line: 1},
		{typ: RETURN,     lexeme: "return", lit: nil, line: 1},
		{typ: SUPER,      lexeme: "super",  lit: nil, line: 1},
		{typ: THIS,       lexeme: "this",   lit: nil, line: 1},
		{typ: TRUE,       lexeme: "true",   lit: nil, line: 1},
		{typ: VAR,        lexeme: "var",    lit: nil, line: 1},
		{typ: WHILE,      lexeme: "while",  lit: nil, line: 1},
		{typ: IDENTIFIER, lexeme: "_foo",   lit: nil, line: 1},   // underscore start
		{typ: IDENTIFIER, lexeme: "Bar123", lit: nil, line: 1},   // alpha + digits
	})
}

func TestMultilineStringUpdatesLine(t *testing.T) {
	src := "\"hello\nworld\" ()"
	checkTokens(t, src, []expectedToken{
		{typ: STRING,      lexeme: "\"hello\nworld\"", lit: "hello\nworld", line: 2},
		{typ: LEFT_PAREN,  lexeme: "(", lit: nil, line: 2},
		{typ: RIGHT_PAREN, lexeme: ")", lit: nil, line: 2},
	})
}

func TestUnexpectedCharacterSetsError(t *testing.T) {
	shared.ResetErrors()
	s := NewScanner("@")
	toks := s.ScanTokens()

	if !shared.HadError {
		t.Fatalf("expected HadError to be true for unexpected character '@'")
	}

	// Scanner should still produce an EOF token.
	if len(toks) != 1 || toks[0].Type != EOF {
		t.Fatalf("expected only EOF token after error, got %+v", toks)
	}
}

func TestTokenTypeStringKnownValues(t *testing.T) {
    cases := []struct {
        tt   TokenType
        want string
    }{
        {LEFT_PAREN, "LEFT_PAREN"},
        {RIGHT_PAREN, "RIGHT_PAREN"},
        {LEFT_BRACE, "LEFT_BRACE"},
        {RIGHT_BRACE, "RIGHT_BRACE"},
        {COMMA, "COMMA"},
        {DOT, "DOT"},
        {MINUS, "MINUS"},
        {PLUS, "PLUS"},
        {SEMICOLON, "SEMICOLON"},
        {SLASH, "SLASH"},
        {STAR, "STAR"},
        {BANG, "BANG"},
        {BANG_EQUAL, "BANG_EQUAL"},
        {EQUAL, "EQUAL"},
        {EQUAL_EQUAL, "EQUAL_EQUAL"},
        {GREATER, "GREATER"},
        {GREATER_EQUAL, "GREATER_EQUAL"},
        {LESS, "LESS"},
        {LESS_EQUAL, "LESS_EQUAL"},
        {IDENTIFIER, "IDENTIFIER"},
        {STRING, "STRING"},
        {NUMBER, "NUMBER"},
        {AND, "AND"},
        {CLASS, "CLASS"},
        {ELSE, "ELSE"},
        {FALSE, "FALSE"},
        {FUN, "FUN"},
        {FOR, "FOR"},
        {IF, "IF"},
        {NIL, "NIL"},
        {OR, "OR"},
        {PRINT, "PRINT"},
        {RETURN, "RETURN"},
        {SUPER, "SUPER"},
        {THIS, "THIS"},
        {TRUE, "TRUE"},
        {VAR, "VAR"},
        {WHILE, "WHILE"},
        {EOF, "EOF"},
    }

    for _, tc := range cases {
        got := tc.tt.String()
        if got != tc.want {
            t.Errorf("TokenType(%v).String() = %q, want %q", tc.tt, got, tc.want)
        }
    }
}

func TestTokenTypeStringUnknownValue(t *testing.T) {
    tt := TokenType(-1)
    got := tt.String()
    if got != "UNKNOWN" {
        t.Fatalf("TokenType(-1).String() = %q, want %q", got, "UNKNOWN")
    }
}

