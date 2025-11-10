package interpreter

import "fmt"

type LoxClass struct {
    Name string
    Superclass *LoxClass
	Methods map[string]*LoxFunction
}

func NewLoxClass(name string, superclass *LoxClass, methods map[string]*LoxFunction) *LoxClass {
    return &LoxClass{
        Name:       name,
        Superclass: superclass,
        Methods:    methods,
    }
}

func (c *LoxClass) String() string {
    return fmt.Sprintf("<class %s>", c.Name)
}

func (c *LoxClass) Call(in *Interpreter, arguments []any) any {
    instance := NewLoxInstance(c)

    if initializer := c.FindMethod("init"); initializer != nil {
        initializer.Bind(instance).Call(in, arguments)
    }

    return instance
}

func (c *LoxClass) Arity() int {
    if initializer := c.FindMethod("init"); initializer != nil {
        return initializer.Arity()
    }
    return 0
}

func (c *LoxClass) FindMethod(name string) *LoxFunction {
    if m, ok := c.Methods[name]; ok {
        return m
    }

    if c.Superclass != nil {
        return c.Superclass.FindMethod(name)
    }
    
    return nil
}