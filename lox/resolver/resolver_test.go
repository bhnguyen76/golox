package resolver

import (
    "testing"

    "example.com/golox/lox/interpreter"
    "example.com/golox/lox/parser"
    "example.com/golox/lox/scanner"
    "example.com/golox/lox/shared"
)

// helper to run scanner + parser + resolver, and tell us if resolver reported an error.
func resolveSource(src string) (hadError bool) {
    shared.ResetErrors()
    shared.HadRuntimeError = false // just in case

    s := scanner.NewScanner(src)
    tokens := s.ScanTokens()

    p := parser.NewParser(tokens)
    stmts := p.Parse()

    // If parsing itself failed, just treat that as an error.
    if shared.HadError {
        return true
    }

    interp := interpreter.NewInterpreter()
    r := NewResolver(interp)
    r.Resolve(stmts)

    return shared.HadError
}

func TestReturnFromTopLevelIsError(t *testing.T) {
    src := `return 123;`

    if ok := resolveSource(src); !ok {
        // ok == true means HadError == true
        t.Fatalf("expected resolver to report error for top-level return, but it did not")
    }
}

func TestThisOutsideClassIsError(t *testing.T) {
    src := `
        fun f() {
            print this;
        }
        f();
    `

    if ok := resolveSource(src); !ok {
        t.Fatalf("expected resolver to report error for using 'this' outside of a class")
    }
}

func TestThisInsideClassIsOK(t *testing.T) {
    src := `
        class Foo {
            method() {
                print this;
            }
        }
        var f = Foo();
        f.method();
    `
    if ok := resolveSource(src); ok {
        t.Fatalf("did not expect resolver error when 'this' is used inside a class")
    }
}

func TestLocalVariableInOwnInitializerIsError(t *testing.T) {
    src := `
        {
            var a = a;
        }
    `
    if ok := resolveSource(src); !ok {
        t.Fatalf("expected resolver to report error for variable used in its own initializer")
    }
}

func TestDuplicateVariableInSameScopeIsError(t *testing.T) {
    src := `
        {
            var a = 1;
            var a = 2;
        }
    `
    if ok := resolveSource(src); !ok {
        t.Fatalf("expected resolver error for duplicate variable in same scope")
    }
}

func TestShadowingInInnerScopeIsOK(t *testing.T) {
    src := `
        var a = 1;
        {
            var a = 2;
            print a;
        }
    `
    if ok := resolveSource(src); ok {
        t.Fatalf("did not expect resolver error for shadowing in inner scope")
    }
}

func TestReturnValueFromInitializerIsError(t *testing.T) {
    src := `
        class Foo {
            init() {
                return 42;
            }
        }
    `
    if ok := resolveSource(src); !ok {
        t.Fatalf("expected resolver to report error for returning value from initializer")
    }
}

func TestSuperOutsideClassIsError(t *testing.T) {
    src := `
        super.foo();
    `
    if ok := resolveSource(src); !ok {
        t.Fatalf("expected resolver to report error for using 'super' outside of a class")
    }
}

func TestSuperInClassWithoutSuperclassIsError(t *testing.T) {
    src := `
        class A {
            method() {
                super.method();
            }
        }
    `
    if ok := resolveSource(src); !ok {
        t.Fatalf("expected resolver to report error for using 'super' in class with no superclass")
    }
}

func TestSuperInSubclassIsOK(t *testing.T) {
    src := `
        class A {
            method() {}
        }
        class B < A {
            method() {
                super.method();
            }
        }
    `
    if ok := resolveSource(src); ok {
        t.Fatalf("did not expect resolver error for using 'super' in a subclass")
    }
}

func TestClassCannotInheritFromItself(t *testing.T) {
    src := `
        class Foo < Foo {}
    `
    if ok := resolveSource(src); !ok {
        t.Fatalf("expected resolver to report error for class inheriting from itself")
    }
}

func TestGetAndSetAreResolved(t *testing.T) {
    src := `
        class Foo {}
        var f = Foo();
        f.x = 10;
        print f.x;
    `
    if ok := resolveSource(src); ok {
        t.Fatalf("did not expect resolver error for property get/set usage")
    }
}

func TestLogicalExpressionIsResolved(t *testing.T) {
    src := `
        var a = true or false and true;
    `
    if ok := resolveSource(src); ok {
        t.Fatalf("did not expect resolver error for logical expression")
    }
}
