package interpreter

import (
	"fmt"

	"example.com/golox/lox/ast"
)

type returnValue struct {
	Value any
}

type LoxFunction struct {
	declaration *ast.Function
	closure		*Environment
}

func NewLoxFunction(declaration *ast.Function, closure *Environment) *LoxFunction {
	return &LoxFunction{
		declaration: declaration,
		closure: closure,
	}
}

func (f *LoxFunction) Arity() int {
	return len(f.declaration.Params)
}

func (f *LoxFunction) Call(in *Interpreter, args []any) (result any) {
	environment := NewEnclosedEnvironment(f.closure)

	for i, param := range f.declaration.Params {
		environment.Define(param.Lexeme, args[i])
	}

	defer func() {
		if r := recover(); r != nil {
			if ret, ok := r.(returnValue); ok {
				result = ret.Value
			} else {
				panic(r)
			}
		}
	}()

	in.executeBlock(f.declaration.Body, environment)

	return nil
}

func (f *LoxFunction) String() string {
	return fmt.Sprintf("<fn %s>", f.declaration.Name.Lexeme)
}

