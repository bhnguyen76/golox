package ast

import (
    "testing"

    "example.com/golox/lox/scanner"
)

func tok(typ scanner.TokenType, lexeme string) scanner.Token {
    return scanner.Token{
        Type:   typ,
        Lexeme: lexeme,
        Line:   1,
    }
}

func TestPrintNilExprReturnsEmptyString(t *testing.T) {
    p := &AstPrinter{}
    got := p.Print(nil)
    if got != "" {
        t.Fatalf("expected empty string for nil expr, got %q", got)
    }
}

func TestLiteralPrintingNumbersStringsBoolsNil(t *testing.T) {
    p := &AstPrinter{}

    tests := []struct {
        name string
        val  any
        want string
    }{
        {"number", 123.0, "123"},
        {"string", "hello", "hello"},
        {"true", true, "true"},
        {"false", false, "false"},
        {"nil", nil, "nil"},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            expr := &Literal{Value: tc.val}
            got := p.Print(expr)
            if got != tc.want {
                t.Fatalf("literal %s: expected %q, got %q", tc.name, tc.want, got)
            }
        })
    }
}

func TestUnaryPrinting(t *testing.T) {
    expr := &Unary{
        Operator: tok(scanner.MINUS, "-"),
        Right:    &Literal{Value: 123.0},
    }

    got := (&AstPrinter{}).Print(expr)
    want := "(- 123)"

    if got != want {
        t.Fatalf("expected %q, got %q", want, got)
    }
}

func TestBinaryPrinting(t *testing.T) {
    expr := &Binary{
        Left:     &Literal{Value: 1.0},
        Operator: tok(scanner.PLUS, "+"),
        Right:    &Literal{Value: 2.0},
    }

    got := (&AstPrinter{}).Print(expr)
    want := "(+ 1 2)"

    if got != want {
        t.Fatalf("expected %q, got %q", want, got)
    }
}

func TestGroupingPrinting(t *testing.T) {
    expr := &Grouping{
        Expression: &Literal{Value: 42.0},
    }

    got := (&AstPrinter{}).Print(expr)
    want := "(group 42)"

    if got != want {
        t.Fatalf("expected %q, got %q", want, got)
    }
}

func TestVariablePrinting(t *testing.T) {
    expr := &Variable{
        Name: tok(scanner.IDENTIFIER, "foo"),
    }

    got := (&AstPrinter{}).Print(expr)
    want := "foo"

    if got != want {
        t.Fatalf("expected %q, got %q", want, got)
    }
}

func TestAssignPrinting(t *testing.T) {
    expr := &Assign{
        Name: tok(scanner.IDENTIFIER, "a"),
        Value: &Binary{
            Left:     &Literal{Value: 1.0},
            Operator: tok(scanner.PLUS, "+"),
            Right:    &Literal{Value: 2.0},
        },
    }

    got := (&AstPrinter{}).Print(expr)
    want := "(assign a (+ 1 2))"

    if got != want {
        t.Fatalf("expected %q, got %q", want, got)
    }
}

func TestLogicalPrinting(t *testing.T) {
    expr := &Logical{
        Left:     &Literal{Value: true},
        Operator: tok(scanner.OR, "or"),
        Right:    &Literal{Value: false},
    }

    got := (&AstPrinter{}).Print(expr)
    want := "(or true false)"

    if got != want {
        t.Fatalf("expected %q, got %q", want, got)
    }
}

func TestCallPrinting(t *testing.T) {
    expr := &Call{
        Callee: &Variable{Name: tok(scanner.IDENTIFIER, "foo")},
        Paren:  tok(scanner.RIGHT_PAREN, ")"),
        Arguments: []Expr{
            &Literal{Value: 1.0},
            &Literal{Value: 2.0},
        },
    }

    got := (&AstPrinter{}).Print(expr)
    want := "(call foo 1 2)"

    if got != want {
        t.Fatalf("expected %q, got %q", want, got)
    }
}

func TestGetPrinting(t *testing.T) {
    expr := &Get{
        Object: &Variable{Name: tok(scanner.IDENTIFIER, "foo")},
        Name:   tok(scanner.IDENTIFIER, "bar"),
    }

    got := (&AstPrinter{}).Print(expr)
    want := "(get bar foo)"

    if got != want {
        t.Fatalf("expected %q, got %q", want, got)
    }
}

func TestSetPrinting(t *testing.T) {
    expr := &Set{
        Object: &Variable{Name: tok(scanner.IDENTIFIER, "foo")},
        Name:   tok(scanner.IDENTIFIER, "bar"),
        Value:  &Literal{Value: 123.0},
    }

    got := (&AstPrinter{}).Print(expr)
    want := "(set bar foo 123)"

    if got != want {
        t.Fatalf("expected %q, got %q", want, got)
    }
}

func TestThisPrinting(t *testing.T) {
    expr := &This{
        Keyword: tok(scanner.THIS, "this"),
    }

    got := (&AstPrinter{}).Print(expr)
    want := "this"

    if got != want {
        t.Fatalf("expected %q, got %q", want, got)
    }
}

func TestSuperPrinting(t *testing.T) {
    expr := &Super{
        Keyword: tok(scanner.SUPER, "super"),
        Method:  tok(scanner.IDENTIFIER, "method"),
    }

    got := (&AstPrinter{}).Print(expr)
    want := "(super method)"

    if got != want {
        t.Fatalf("expected %q, got %q", want, got)
    }
}

type fakeExpr struct{}

func (fakeExpr) Accept(v ExprVisitor) any {
    return 99 
}

func TestPrintFallbackToSprintForNonStringResult(t *testing.T) {
    p := &AstPrinter{}
    got := p.Print(fakeExpr{})
    want := "99"

    if got != want {
        t.Fatalf("expected %q from fallback fmt.Sprint, got %q", want, got)
    }
}

func TestAstPrinterCompositeExample(t *testing.T) {
    expr := &Binary{
        Left: &Unary{
            Operator: tok(scanner.MINUS, "-"),
            Right:    &Literal{Value: 123.0},
        },
        Operator: tok(scanner.STAR, "*"),
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
