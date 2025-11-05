package ast

import "example.com/golox/lox/scanner"

type Expr interface {
	exprNode()
}
type Binary struct {
	Left Expr
	Operator scanner.Token
	Right Expr
}

func (Binary) exprNode() {}