package ast

import "example.com/golox/lox/scanner"

type Expr interface {
	Accept(v Visitor) any
}

type Visitor interface {
	VisitBinaryExpr(*Binary) any
	VisitGroupingExpr(*Grouping) any
	VisitLiteralExpr(*Literal) any
	VisitUnaryExpr(*Unary) any
}

type Binary struct {
	Left Expr
	Operator scanner.Token
	Right Expr
}

func (b *Binary) Accept(v Visitor) any {
	return v.VisitBinaryExpr(b)
}

type Grouping struct {
	Expression Expr
}

func (g *Grouping) Accept(v Visitor) any {
	return v.VisitGroupingExpr(g)
}

type Literal struct {
	Value any
}

func (l *Literal) Accept(v Visitor) any {
	return v.VisitLiteralExpr(l)
}

type Unary struct {
	Operator scanner.Token
	Right Expr
}

func (u *Unary) Accept(v Visitor) any {
	return v.VisitUnaryExpr(u)
}

