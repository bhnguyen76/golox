package ast

import "example.com/golox/lox/scanner"

type Expr interface {
	Accept(v ExprVisitor) any
}

type ExprVisitor interface {
	VisitAssignExpr(*Assign) any
	VisitBinaryExpr(*Binary) any
	VisitCallExpr(*Call) any
	VisitGroupingExpr(*Grouping) any
	VisitLiteralExpr(*Literal) any
	VisitLogicalExpr(*Logical) any
	VisitUnaryExpr(*Unary) any
	VisitVariableExpr(*Variable) any
}

type Assign struct {
	Name scanner.Token
	Value Expr
}

func (n *Assign) Accept(v ExprVisitor) any {
	return v.VisitAssignExpr(n)
}

type Binary struct {
	Left Expr
	Operator scanner.Token
	Right Expr
}

func (n *Binary) Accept(v ExprVisitor) any {
	return v.VisitBinaryExpr(n)
}

type Call struct {
	Callee Expr
	Paren scanner.Token
	Arguments []Expr
}

func (n *Call) Accept(v ExprVisitor) any {
	return v.VisitCallExpr(n)
}

type Grouping struct {
	Expression Expr
}

func (n *Grouping) Accept(v ExprVisitor) any {
	return v.VisitGroupingExpr(n)
}

type Literal struct {
	Value any
}

func (n *Literal) Accept(v ExprVisitor) any {
	return v.VisitLiteralExpr(n)
}

type Logical struct {
	Left Expr
	Operator scanner.Token
	Right Expr
}

func (n *Logical) Accept(v ExprVisitor) any {
	return v.VisitLogicalExpr(n)
}

type Unary struct {
	Operator scanner.Token
	Right Expr
}

func (n *Unary) Accept(v ExprVisitor) any {
	return v.VisitUnaryExpr(n)
}

type Variable struct {
	Name scanner.Token
}

func (n *Variable) Accept(v ExprVisitor) any {
	return v.VisitVariableExpr(n)
}

