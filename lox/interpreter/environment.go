package interpreter

import (
	"fmt"

	"example.com/golox/lox/scanner"
)


type Environment struct {
	enclosing *Environment
	values map[string]any
}

func NewEnvironment() *Environment {
	return &Environment{
		enclosing: nil,
		values: make(map[string]any),
	}
}

func NewEnclosedEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		enclosing: enclosing,
		values:    make(map[string]any),
	}
}

func (env *Environment) Get(name scanner.Token) any {
	if val, ok := env.values[name.Lexeme]; ok {
		return val
	}
	
	if env.enclosing != nil {
		return env.enclosing.Get(name)
	}

	panic(RuntimeError{
		Token:   name,
		Message: fmt.Sprintf("Undefined variable '%s'.", name.Lexeme),
	})
}

func (env *Environment) GetAt(distance int, name string) any {
	environment := env.ancestor(distance)
    if value, ok := environment.values[name]; ok {
        return value
    }
    
    return nil
}

func (env *Environment) ancestor(distance int) *Environment {
    environment := env
    for i := 0; i < distance; i++ {
        if environment.enclosing == nil {
            break 
        }
        environment = environment.enclosing
    }
    return environment
}

func (env *Environment) Assign(name scanner.Token, value any) {
	if _, ok := env.values[name.Lexeme]; ok {
		env.values[name.Lexeme] = value
		return
	}

	if env.enclosing != nil {
		env.enclosing.Assign(name, value)
		return
	}

	panic(RuntimeError{
		Token:   name,
		Message: fmt.Sprintf("Undefined variable '%s'.", name.Lexeme),
	})
}

func (env *Environment) AssignAt(distance int, name scanner.Token, value any) {
    environment := env.ancestor(distance)
    environment.values[name.Lexeme] = value
}

func (env *Environment) Define(name string, value any) {
	env.values[name] = value
}