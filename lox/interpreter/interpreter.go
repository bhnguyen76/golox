package interpreter

import (
	"fmt"
	"os"

	"example.com/golox/lox/ast"
	"example.com/golox/lox/scanner"
	"example.com/golox/lox/shared"
)

type RuntimeError struct {
	Token   scanner.Token
	Message string
}

func (e RuntimeError) Error() string {
	return e.Message
}

type Interpreter struct{}

// Eval is like Java's "evaluate" from outside: given an Expr, produce a value.
func (in *Interpreter) evaluate(expr ast.Expr) any {
	if expr == nil {
		return nil
	}
	return expr.Accept(in)
}

// --- Visitor methods ---

func (in *Interpreter) VisitLiteralExpr(expr *ast.Literal) any {
	return expr.Value
}

func (in *Interpreter) VisitGroupingExpr(expr *ast.Grouping) any {
	return in.evaluate(expr.Expression)
}

func (in *Interpreter) VisitUnaryExpr(expr *ast.Unary) any {
	right := in.evaluate(expr.Right)

	switch expr.Operator.Type {
	case scanner.BANG:
		return !isTruthy(right)
	case scanner.MINUS:
		checkNumberOperand(expr.Operator, right)
		num := right.(float64)
		return -num
	}

	// Unreachable 
	return nil
}

func (in *Interpreter) VisitBinaryExpr(expr *ast.Binary) any {
	left := in.evaluate(expr.Left)
	right := in.evaluate(expr.Right)

	switch expr.Operator.Type {
	case scanner.GREATER:
		checkNumberOperands(expr.Operator, left, right)
		return left.(float64) > right.(float64)
	case scanner.GREATER_EQUAL:
		checkNumberOperands(expr.Operator, left, right)
		return left.(float64) >= right.(float64)
	case scanner.LESS:
		checkNumberOperands(expr.Operator, left, right)
		return left.(float64) < right.(float64)
	case scanner.LESS_EQUAL:
		checkNumberOperands(expr.Operator, left, right)
		return left.(float64) <= right.(float64)

	case scanner.MINUS:
		checkNumberOperands(expr.Operator, left, right)
		return left.(float64) - right.(float64)
	case scanner.PLUS:
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l + r
			}
			panic(RuntimeError{Token: expr.Operator, Message: "Right operand must be a number."})
		}

		if ls, ok := left.(string); ok {
			if rs, ok := right.(string); ok {
				return ls + rs
			}
			panic(RuntimeError{Token: expr.Operator, Message: "Right operand must be a string."})
		}
		panic(RuntimeError{Token: expr.Operator, Message: "Operands must be two numbers or two strings."})

	case scanner.SLASH:
		checkNumberOperands(expr.Operator, left, right)
		return left.(float64) / right.(float64)
	case scanner.STAR:
		checkNumberOperands(expr.Operator, left, right)
		return left.(float64) * right.(float64)

	case scanner.BANG_EQUAL:
		return !isEqual(left, right)
	case scanner.EQUAL_EQUAL:
		return isEqual(left, right)
	}

	// Unreachable.
	return nil
}

func (in *Interpreter) Interpret(expr ast.Expr) {
	defer func() {
		if r := recover(); r != nil {
			if rt, ok := r.(RuntimeError); ok {
				fmt.Fprintf(os.Stderr, "%s\n[line %d]\n", rt.Message, rt.Token.Line)
				shared.HadRuntimeError = true
			} else {
				panic(r)
			}
		}
	}()

	value := in.evaluate(expr)
	fmt.Println(stringify(value))
}

func isTruthy(object any) bool {
	if object == nil {
		return false
	}
	if b, ok := object.(bool); ok {
		return b
	}
	return true
}

func isEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil {
		return false
	}
	return a == b
}

func stringify(object any) string {
	if object == nil {
		return "nil"
	}

	if num, ok := object.(float64); ok {
		s := fmt.Sprintf("%g", num)
		return s
	}

	return fmt.Sprint(object)
}

func checkNumberOperand(operator scanner.Token, operand any) {
	if _, ok := operand.(float64); ok {
		return
	}
	panic(RuntimeError{
		Token:   operator,
		Message: "Operand must be a number.",
	})
}

func checkNumberOperands(operator scanner.Token, left, right any) {
	_, l := left.(float64)
	_, r := right.(float64)
	if l && r {
		return
	}

	panic(RuntimeError{
		Token:   operator,
		Message: "Operands must be numbers.",
	})
}