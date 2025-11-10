package resolver

import (
	"fmt"

	"example.com/golox/lox/ast"
	"example.com/golox/lox/interpreter"
	"example.com/golox/lox/scanner"
	"example.com/golox/lox/shared"
)

type FunctionType int

const (
    FunctionNone FunctionType = iota
    FunctionFunction
    FunctionInitializer
    FunctionMethod
)

type ClassType int

const (
    ClassNone ClassType = iota
    ClassClass
    ClassSubClass
)

type Resolver struct {
	interpreter *interpreter.Interpreter
	scopes []map[string]bool
    currentFunction FunctionType
    currentClass ClassType
}

func (r *Resolver) errorToken(token scanner.Token, message string) {
    where := fmt.Sprintf(" at '%s'", token.Lexeme)
    shared.Report(token.Line, where, message)
}

func NewResolver(interpreter *interpreter.Interpreter) *Resolver {
	return &Resolver{
		interpreter: interpreter,
        scopes: nil,
        currentFunction: FunctionNone,
        currentClass:    ClassNone,
	}
}

func (r *Resolver) VisitBlockStmt(stmt *ast.Block) any {
	r.beginScope()
	r.resolveStmts(stmt.Statements)
	r.endScope()
	return nil
}

func (r *Resolver) VisitVarStmt(stmt *ast.Var) any {
	r.declare(stmt.Name)
	if stmt.Initializer != nil {
		r.resolveExpr(stmt.Initializer)
	}
	r.define(stmt.Name)
	return nil
}

func (r *Resolver) VisitVariableExpr(expr *ast.Variable) any {
    if len(r.scopes) > 0 {
        top := r.scopes[len(r.scopes)-1]

        if defined, ok := top[expr.Name.Lexeme]; ok && !defined {
            r.errorToken(expr.Name, "Can't read local variable in its own initializer.")
        }
    }

    r.resolveLocal(expr, expr.Name)
    return nil
}

func (r *Resolver) VisitAssignExpr(expr *ast.Assign) any {
    r.resolveExpr(expr.Value)
    r.resolveLocal(expr, expr.Name)
    return nil
}

func (r *Resolver) VisitFunctionStmt(stmt *ast.Function) any {
	r.declare(stmt.Name)
	r.define(stmt.Name)

	r.resolveFunction(stmt, FunctionFunction)
	return nil
}

func (r *Resolver) VisitExpressionStmt(stmt *ast.Expression) any {
	r.resolveExpr(stmt.Expression)
	return nil
}

func (r *Resolver) VisitIfStmt(stmt *ast.If) any {
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.ThenBranch)
    if stmt.ElseBranch != nil {
        r.resolveStmt(stmt.ElseBranch)
    }
    return nil
}

func (r *Resolver) VisitPrintStmt(stmt *ast.Print) any {
    r.resolveExpr(stmt.Expression) 
    return nil
}

func (r *Resolver) VisitReturnStmt(stmt *ast.Return) any {
    if r.currentFunction == FunctionNone {
        r.errorToken(stmt.Keyword, "Can't return from top-level code.")
    }

    if stmt.Value != nil {
        if r.currentFunction == FunctionInitializer {
            r.errorToken(stmt.Keyword, "Can't return a value from an initializer.")
        }
        r.resolveExpr(stmt.Value)
    }

    return nil
}


func (r *Resolver) VisitWhileStmt(stmt *ast.While) any {
    r.resolveExpr(stmt.Condition)
    r.resolveStmt(stmt.Body)
    return nil
}


func (r *Resolver) VisitBinaryExpr(expr *ast.Binary) any {
    r.resolveExpr(expr.Left)
    r.resolveExpr(expr.Right)
    return nil
}

func (r *Resolver) VisitCallExpr(expr *ast.Call) any {
    r.resolveExpr(expr.Callee)

    for _, arg := range expr.Arguments {
        r.resolveExpr(arg)
    }
    return nil
}

func (r *Resolver) VisitGroupingExpr(expr *ast.Grouping) any {
    r.resolveExpr(expr.Expression)
    return nil
}

func (r *Resolver) VisitLiteralExpr(expr *ast.Literal) any {
    return nil
}

func (r *Resolver) VisitLogicalExpr(expr *ast.Logical) any {
    r.resolveExpr(expr.Left)
    r.resolveExpr(expr.Right)
    return nil
}

func (r *Resolver) VisitUnaryExpr(expr *ast.Unary) any {
    r.resolveExpr(expr.Right)
    return nil
}

func (r *Resolver) VisitClassStmt(stmt *ast.Class) any {
    enclosingClass := r.currentClass
    r.currentClass = ClassClass

    r.declare(stmt.Name)
    r.define(stmt.Name)

    if stmt.Superclass != nil {
        if superVar, ok := stmt.Superclass.(*ast.Variable); ok {
            if stmt.Name.Lexeme == superVar.Name.Lexeme {
                r.errorToken(superVar.Name, "A class can't inherit from itself.")
            }
        }
    }

    if stmt.Superclass != nil {
        r.currentClass = ClassSubClass
        r.resolveExpr(stmt.Superclass)
    }

    if stmt.Superclass != nil {
        r.beginScope()
        r.scopes[len(r.scopes)-1]["super"] = true
    }

    r.beginScope()

    r.scopes[len(r.scopes)-1]["this"] = true

    for _, method := range stmt.Methods {
        fnType := FunctionMethod
        if method.Name.Lexeme == "init" {
            fnType = FunctionInitializer
        }
        r.resolveFunction(method, fnType)
    }

    r.endScope()

    if stmt.Superclass != nil {
        r.endScope()
    }

    r.currentClass = enclosingClass
    return nil
}


func (r *Resolver) VisitGetExpr(expr *ast.Get) any {
    r.resolveExpr(expr.Object)
    return nil
}

func (r *Resolver) VisitSetExpr(expr *ast.Set) any {
    r.resolveExpr(expr.Value)
    r.resolveExpr(expr.Object)
    return nil
}

func (r *Resolver) VisitThisExpr(expr *ast.This) any {
    if r.currentClass == ClassNone {
        r.errorToken(expr.Keyword, "Can't use 'this' outside of a class.")
        return nil
    }

    r.resolveLocal(expr, expr.Keyword)
    return nil
}

func (r *Resolver) VisitSuperExpr(expr *ast.Super) any {
    if r.currentClass == ClassNone {
        r.errorToken(expr.Keyword, "Can't use 'super' outside of a class.")
    } else if r.currentClass != ClassSubClass {
        r.errorToken(expr.Keyword, "Can't use 'super' in a class with no superclass.")
    }

    r.resolveLocal(expr, expr.Keyword)
    return nil
}

func (r *Resolver) Resolve(statements []ast.Stmt) {
	r.resolveStmts(statements)
}

func (r *Resolver) resolveStmts(stmts []ast.Stmt) {
	for _, s := range stmts {
		if s != nil {
			s.Accept(r)
		}
	}
}

func (r *Resolver) resolveStmt(stmt ast.Stmt) {
    if stmt != nil {
        stmt.Accept(r)
    }
}

func (r *Resolver) resolveExpr(expr ast.Expr) {
	if expr != nil {
		expr.Accept(r)
	}
}

func (r *Resolver) beginScope() {
	r.scopes = append(r.scopes, make(map[string]bool))
}

func (r *Resolver) endScope() {
	r.scopes = r.scopes[:len(r.scopes)-1]
}

func (r *Resolver) declare(name scanner.Token) {
    if len(r.scopes) == 0 {
        return
    }

    scope := r.scopes[len(r.scopes)-1]

    if _, exists := scope[name.Lexeme]; exists {
        r.errorToken(name, "Already a variable with this name in this scope.")
    }

    scope[name.Lexeme] = false
}


func (r *Resolver) define(name scanner.Token) {
	if len(r.scopes) == 0 {
		return
	}
	scope := r.scopes[len(r.scopes)-1]
	scope[name.Lexeme] = true
}

func (r *Resolver) resolveLocal(expr ast.Expr, name scanner.Token) {
    for i := len(r.scopes) - 1; i >= 0; i-- {
        scope := r.scopes[i]
        if _, ok := scope[name.Lexeme]; ok {
            distance := len(r.scopes) - 1 - i
            r.interpreter.Resolve(expr, distance)
            return
        }
    }
}

func (r *Resolver) resolveFunction(function *ast.Function, fnType FunctionType) {
    enclosingFunction := r.currentFunction
    r.currentFunction = fnType

    r.beginScope()
    for _, param := range function.Params {
        r.declare(param)
        r.define(param)
    }
    r.resolveStmts(function.Body)
    r.endScope()
    r.currentFunction = enclosingFunction
}

