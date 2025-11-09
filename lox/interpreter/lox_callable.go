package interpreter

type LoxCallable interface {
	Call(in *Interpreter, arguments []any) any

	Arity() int
}