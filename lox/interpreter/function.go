package interpreter

import (
	"fmt"

	"example.com/golox/lox/ast"
)

type returnValue struct {
	Value any
}

type LoxFunction struct {
	Declaration *ast.Function
	Closure		*Environment
	IsInitializer bool
}

func NewLoxFunction(declaration *ast.Function, closure *Environment, isInitializer bool) *LoxFunction {
	return &LoxFunction{
		Declaration: declaration,
		Closure: closure,
		IsInitializer: isInitializer,
	}
}

func (f *LoxFunction) Arity() int {
	return len(f.Declaration.Params)
}

func (f *LoxFunction) Call(in *Interpreter, arguments []any) (result any) {
    env := NewEnclosedEnvironment(f.Closure)

    for i, param := range f.Declaration.Params {
        env.Define(param.Lexeme, arguments[i])
    }

    defer func() {
        if r := recover(); r != nil {
            if rv, ok := r.(returnValue); ok {
                if f.IsInitializer {
                    result = f.Closure.GetAt(0, "this")
                } else {
                    result = rv.Value
                }
            } else {
                panic(r)
            }
        } else {
            if f.IsInitializer {
                result = f.Closure.GetAt(0, "this")
            } else {
                result = nil
            }
        }
    }()

    in.executeBlock(f.Declaration.Body, env)
    return 
}

func (f *LoxFunction) Bind(instance *LoxInstance) *LoxFunction {
    env := NewEnclosedEnvironment(f.Closure)
    env.Define("this", instance)
    return &LoxFunction{
        Declaration:   f.Declaration,
        Closure:       env,
        IsInitializer: f.IsInitializer,
    }
}

func (f *LoxFunction) String() string {
	return fmt.Sprintf("<fn %s>", f.Declaration.Name.Lexeme)
}

