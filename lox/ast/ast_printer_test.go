package ast

import (
	"testing"

	"example.com/golox/lox/scanner"
)

func TestAstPrinterExample(t *testing.T) {
	expr := &Binary{
		Left: &Unary{
			Operator: scanner.Token{Type: scanner.MINUS, Lexeme: "-", Line: 1},
			Right:    &Literal{Value: 123.0},
		},
		Operator: scanner.Token{Type: scanner.STAR, Lexeme: "*", Line: 1},
		Right: &Grouping{
			Expression: &Literal{Value: 45.67},
		},
	}

	got := (&AstPrinter{}).Print(expr)
	want := "(* (- 123) (group 45.67))"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
