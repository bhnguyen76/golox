package ast

import "example.com/golox/lox/scanner"

type Stmt interface {
	Accept(v StmtVisitor) any
}

type StmtVisitor interface {
	VisitBlockStmt(*Block) any
	VisitClassStmt(*Class) any
	VisitExpressionStmt(*Expression) any
	VisitFunctionStmt(*Function) any
	VisitPrintStmt(*Print) any
	VisitIfStmt(*If) any
	VisitReturnStmt(*Return) any
	VisitVarStmt(*Var) any
	VisitWhileStmt(*While) any
}

type Block struct {
	Statements []Stmt
}

func (n *Block) Accept(v StmtVisitor) any {
	return v.VisitBlockStmt(n)
}

type Class struct {
	Name scanner.Token
	Superclass Expr
	Methods []*Function
}

func (n *Class) Accept(v StmtVisitor) any {
	return v.VisitClassStmt(n)
}

type Expression struct {
	Expression Expr
}

func (n *Expression) Accept(v StmtVisitor) any {
	return v.VisitExpressionStmt(n)
}

type Function struct {
	Name scanner.Token
	Params []scanner.Token
	Body []Stmt
}

func (n *Function) Accept(v StmtVisitor) any {
	return v.VisitFunctionStmt(n)
}

type Print struct {
	Expression Expr
}

func (n *Print) Accept(v StmtVisitor) any {
	return v.VisitPrintStmt(n)
}

type If struct {
	Condition Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (n *If) Accept(v StmtVisitor) any {
	return v.VisitIfStmt(n)
}

type Return struct {
	Keyword scanner.Token
	Value Expr
}

func (n *Return) Accept(v StmtVisitor) any {
	return v.VisitReturnStmt(n)
}

type Var struct {
	Name scanner.Token
	Initializer Expr
}

func (n *Var) Accept(v StmtVisitor) any {
	return v.VisitVarStmt(n)
}

type While struct {
	Condition Expr
	Body Stmt
}

func (n *While) Accept(v StmtVisitor) any {
	return v.VisitWhileStmt(n)
}

