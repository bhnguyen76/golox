package interpreter

import (
	"fmt"
	"time"
)

type ClockFn struct{}

func (ClockFn) Arity() int { return 0 }

func (ClockFn) Call(in *Interpreter, arguments []any) any {
	return float64(time.Now().UnixNano()) / 1e9
}

func (ClockFn) String() string { return "<native fn>"}

var _ fmt.Stringer = ClockFn{}