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

type Interpreter struct{
	globals *Environment
	environment *Environment
	locals map[ast.Expr]int
}

func NewInterpreter() *Interpreter {
	globals := NewEnvironment()
	globals.Define("clock", ClockFn{})

	return &Interpreter{
		globals: globals,
		environment: globals,
	}
}


func (in *Interpreter) evaluate(expr ast.Expr) any {
	if expr == nil {
		return nil
	}
	return expr.Accept(in)
}

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

func (in *Interpreter) VisitExpressionStmt(stmt *ast.Expression) any {
	in.evaluate(stmt.Expression)
	return nil
}

func (in *Interpreter) VisitPrintStmt(stmt *ast.Print) any {
	value := in.evaluate(stmt.Expression)
	fmt.Println(stringify(value))
	return nil
}

func (in *Interpreter) VisitVarStmt(stmt *ast.Var) any { 
	var value any = nil
	if stmt.Initializer != nil {
		value = in.evaluate(stmt.Initializer)
	}
	in.environment.Define(stmt.Name.Lexeme, value)
	return nil
}

func (in *Interpreter) VisitAssignExpr(expr *ast.Assign) any {
	value := in.evaluate(expr.Value)

	if distance, ok := in.locals[expr]; ok {
        in.environment.AssignAt(distance, expr.Name, value)
    } else {
        in.globals.Assign(expr.Name, value)
    }

	return value
}

func (in *Interpreter) VisitVariableExpr(expr *ast.Variable) any {
	return in.lookUpVariable(expr.Name, expr)
}

func (in *Interpreter) VisitBlockStmt(stmt *ast.Block) any {
	newEnv := NewEnclosedEnvironment(in.environment)
	in.executeBlock(stmt.Statements, newEnv)
	return nil
}

func (in *Interpreter) VisitIfStmt(stmt *ast.If) any {
	if isTruthy(in.evaluate(stmt.Condition)) {
		in.execute(stmt.ThenBranch)
	} else if stmt.ElseBranch != nil {
		in.execute(stmt.ElseBranch)
	}
	return nil
}

func (in *Interpreter) VisitLogicalExpr(expr *ast.Logical) any {
	left := in.evaluate(expr.Left)

	if expr.Operator.Type == scanner.OR {
		if isTruthy(left) {
			return left
		}
	} else {
		if !isTruthy(left) {
			return left
		}
	}

	return in.evaluate(expr.Right)
}

func (in *Interpreter) VisitWhileStmt(stmt *ast.While) any {
	for isTruthy(in.evaluate(stmt.Condition)) {
		in.execute(stmt.Body)
	}
	return nil
}

func (in *Interpreter) VisitCallExpr(expr *ast.Call) any {
	callee := in.evaluate(expr.Callee)

	var arguments []any
	for _, arguement := range expr.Arguments {
		arguments = append(arguments, in.evaluate(arguement))
	}

	fn, ok := callee.(LoxCallable)
	if !ok {
		panic(RuntimeError{
			Token: expr.Paren,
			Message: "Can only call functions and classes.",
		})
	}

	if len(arguments) != fn.Arity() {
		panic(RuntimeError{
			Token: expr.Paren,
			Message: fmt.Sprintf("Expected %d arguments but got %d.", fn.Arity(), len(arguments)),
		})
	}

	return fn.Call(in, arguments)
}

func (in *Interpreter) VisitFunctionStmt(stmt *ast.Function) any {
    function := NewLoxFunction(stmt, in.environment, false)
    in.environment.Define(stmt.Name.Lexeme, function)
    return nil
}

func (in *Interpreter) VisitReturnStmt(stmt *ast.Return) any {
	var value any = nil
	if stmt.Value != nil {
		value = in.evaluate(stmt.Value)
	}

	panic(returnValue{Value: value})
}

func (in *Interpreter) VisitClassStmt(stmt *ast.Class) any {
	var superclass *LoxClass
    if stmt.Superclass != nil {
        value := in.evaluate(stmt.Superclass)

        var ok bool
        superclass, ok = value.(*LoxClass)
        if !ok {
            if superVar, ok2 := stmt.Superclass.(*ast.Variable); ok2 {
                panic(RuntimeError{
                    Token:   superVar.Name,
                    Message: "Superclass must be a class.",
                })
            }
            panic(RuntimeError{
                Token:   stmt.Name,
                Message: "Superclass must be a class.",
            })
        }
    }

    in.environment.Define(stmt.Name.Lexeme, nil)

    var previousEnv *Environment
    if superclass != nil {
        previousEnv = in.environment
        in.environment = NewEnclosedEnvironment(in.environment)
        in.environment.Define("super", superclass)
    }

    methods := make(map[string]*LoxFunction)
    for _, method := range stmt.Methods {
        isInitializer := method.Name.Lexeme == "init"
        function := NewLoxFunction(method, in.environment, isInitializer)
        methods[method.Name.Lexeme] = function
    }

    klass := NewLoxClass(stmt.Name.Lexeme, superclass, methods)

	if superclass != nil {
        in.environment = previousEnv
    }

    in.environment.Assign(stmt.Name, klass)
    return nil
}

func (in *Interpreter) VisitGetExpr(expr *ast.Get) any {
    object := in.evaluate(expr.Object)

    if instance, ok := object.(*LoxInstance); ok {
        return instance.Get(expr.Name)
    }

    panic(RuntimeError{
        Token:   expr.Name,
        Message: "Only instances have properties.",
    })
}

func (in *Interpreter) VisitSetExpr(expr *ast.Set) any {
    object := in.evaluate(expr.Object)

    instance, ok := object.(*LoxInstance)
    if !ok {
        panic(RuntimeError{
            Token:   expr.Name,
            Message: "Only instances have fields.",
        })
    }

    value := in.evaluate(expr.Value)
    instance.Set(expr.Name, value)
    return value
}

func (in *Interpreter) VisitThisExpr(expr *ast.This) any {
	return in.lookUpVariable(expr.Keyword, expr)
}

func (in *Interpreter) VisitSuperExpr(expr *ast.Super) any {
    distance, ok := in.locals[expr]
    if !ok {
        panic(RuntimeError{
            Token:   expr.Keyword,
            Message: "Internal error: no local distance for 'super'.",
        })
    }

    superVal := in.environment.GetAt(distance, "super")
    superclass, ok := superVal.(*LoxClass)
    if !ok {
        panic(RuntimeError{
            Token:   expr.Keyword,
            Message: "Superclass must be a class.",
        })
    }

    thisVal := in.environment.GetAt(distance-1, "this")
    object, ok := thisVal.(*LoxInstance)
    if !ok {
        panic(RuntimeError{
            Token:   expr.Keyword,
            Message: "Internal error: 'this' is not an instance.",
        })
    }

    method := superclass.FindMethod(expr.Method.Lexeme)
    if method == nil {
        panic(RuntimeError{
            Token:   expr.Method,
            Message: fmt.Sprintf("Undefined property '%s'.", expr.Method.Lexeme),
        })
    }

    return method.Bind(object)
}


func (in *Interpreter) Interpret(statements []ast.Stmt) {
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

	for _, statement := range statements {
		in.execute(statement)
	}
}

func (in *Interpreter) execute(stmt ast.Stmt) {
	stmt.Accept(in)
}

func (in *Interpreter) Resolve(expr ast.Expr, depth int) {
	if in.locals == nil {
        in.locals = make(map[ast.Expr]int)
    }
    in.locals[expr] = depth
}

func (in *Interpreter) executeBlock(statements []ast.Stmt, environment *Environment) {
	previous := in.environment
	in.environment = environment
	defer func() { in.environment = previous }()

	for _, stmt := range statements {
		in.execute(stmt)
	}
}

func (in *Interpreter) lookUpVariable(name scanner.Token, expr ast.Expr) any {
    if distance, ok := in.locals[expr]; ok {
        return in.environment.GetAt(distance, name.Lexeme)
    }
    return in.globals.Get(name)
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