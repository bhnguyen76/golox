package interpreter

import (
	"fmt"

	"example.com/golox/lox/scanner"
)

type LoxInstance struct {
    Class  *LoxClass
    Fields map[string]any
}

func NewLoxInstance(class *LoxClass) *LoxInstance {
    return &LoxInstance{
        Class:  class,
        Fields: make(map[string]any),
    }
}

func (i *LoxInstance) Get(name scanner.Token) any {
    if value, ok := i.Fields[name.Lexeme]; ok {
        return value
    }

	if method := i.Class.FindMethod(name.Lexeme); method != nil {
        return method.Bind(i) 
    }

    panic(RuntimeError{
        Token:   name,
        Message: fmt.Sprintf("Undefined property '%s'.", name.Lexeme),
    })
}

func (i *LoxInstance) Set(name scanner.Token, value any) {
    i.Fields[name.Lexeme] = value
}

func (i *LoxInstance) String() string {
    return fmt.Sprintf("<%s instance>", i.Class.Name)
}
