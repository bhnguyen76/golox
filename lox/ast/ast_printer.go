package ast

import (
	"fmt"
	"strings"
)

type AstPrinter struct{}

func (p *AstPrinter) Print(expr Expr) string {
	if expr == nil {
		return ""
	}
	result := expr.Accept(p)

	if s, ok := result.(string); ok {
		return s
	}
	return fmt.Sprint(result)
}

func (p *AstPrinter) VisitBinaryExpr(expr *Binary) any {
	return p.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (p *AstPrinter) VisitGroupingExpr(expr *Grouping) any {
	return p.parenthesize("group", expr.Expression)
}

func (p *AstPrinter) VisitLiteralExpr(expr *Literal) any {
	if expr.Value == nil {
		return "nil"
	}
	return fmt.Sprint(expr.Value)
}

func (p *AstPrinter) VisitUnaryExpr(expr *Unary) any {
	return p.parenthesize(expr.Operator.Lexeme, expr.Right)
}

// ---- Helper ----

func (p *AstPrinter) parenthesize(name string, exprs ...Expr) string {
	var b strings.Builder

	b.WriteString("(")
	b.WriteString(name)

	for _, expr := range exprs {
		b.WriteString(" ")
		val := expr.Accept(p)
		if s, ok := val.(string); ok {
			b.WriteString(s)
		} else {
			b.WriteString(fmt.Sprint(val))
		}
	}

	b.WriteString(")")
	return b.String()
}
