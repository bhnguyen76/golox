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

func (p *AstPrinter) VisitAssignExpr(expr *Assign) any {
	return p.parenthesize("assign "+expr.Name.Lexeme, expr.Value)
}

func (p *AstPrinter) VisitVariableExpr(expr *Variable) any {
	return expr.Name.Lexeme
}

func (p *AstPrinter) VisitLogicalExpr(expr *Logical) any {
	return p.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (p *AstPrinter) VisitCallExpr(expr *Call) any {
	parts := make([]Expr, 0, 1+len(expr.Arguments))
	parts = append(parts, expr.Callee)
	parts = append(parts, expr.Arguments...)
	return p.parenthesize("call", parts...)
}

func (p *AstPrinter) VisitGetExpr(expr *Get) any {
	return p.parenthesize("get "+expr.Name.Lexeme, expr.Object)
}

func (p *AstPrinter) VisitSetExpr(expr *Set) any {
	return p.parenthesize("set "+expr.Name.Lexeme, expr.Object, expr.Value)
}

func (p *AstPrinter) VisitThisExpr(expr *This) any {
	return "this"
}

func (p *AstPrinter) VisitSuperExpr(expr *Super) any {
	return p.parenthesize("super "+expr.Method.Lexeme)
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
