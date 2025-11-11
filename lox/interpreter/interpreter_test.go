package interpreter_test

import (
    "bytes"
    "io"
    "os"
    "strings"
    "testing"

    "example.com/golox/lox/interpreter"
    "example.com/golox/lox/parser"
    "example.com/golox/lox/resolver"
    "example.com/golox/lox/scanner"
    "example.com/golox/lox/shared"
)

func runLox(t *testing.T, src string) (stdout string, hadError, hadRuntimeError bool) {
    t.Helper()

    shared.ResetErrors()
    shared.HadRuntimeError = false

    oldStdout := os.Stdout
    rOut, wOut, _ := os.Pipe()
    os.Stdout = wOut

    outCh := make(chan string)
    go func() {
        var buf bytes.Buffer
        _, _ = io.Copy(&buf, rOut)
        outCh <- buf.String()
    }()

    s := scanner.NewScanner(src)
    tokens := s.ScanTokens()

    p := parser.NewParser(tokens)
    stmts := p.Parse()

    if !shared.HadError {
        in := interpreter.NewInterpreter()
        res := resolver.NewResolver(in)
        res.Resolve(stmts)

        if !shared.HadError {
            in.Interpret(stmts)
        }
    }

    // stop capturing
    wOut.Close()
    os.Stdout = oldStdout
    stdout = strings.TrimSpace(<-outCh)

    hadError = shared.HadError
    hadRuntimeError = shared.HadRuntimeError
    return
}

func TestArithmeticAndPrint(t *testing.T) {
    src := `
        print 1 + 2 * 3;
        print (1 + 2) * 3;
    `
    out, hadErr, hadRt := runLox(t, src)

    if hadErr || hadRt {
        t.Fatalf("unexpected error flags: hadError=%v, hadRuntimeError=%v", hadErr, hadRt)
    }

    lines := strings.Split(out, "\n")
    if len(lines) != 2 {
        t.Fatalf("expected 2 lines of output, got %d (%q)", len(lines), out)
    }
    if strings.TrimSpace(lines[0]) != "7" {
        t.Errorf("expected first line 7, got %q", lines[0])
    }
    if strings.TrimSpace(lines[1]) != "9" {
        t.Errorf("expected second line 9, got %q", lines[1])
    }
}

func TestFunctionAndClosure(t *testing.T) {
    src := `
        fun makeCounter() {
            var i = 0;
            fun count() {
                i = i + 1;
                print i;
            }
            return count;
        }

        var c = makeCounter();
        c();
        c();
        c();
    `
    out, hadErr, hadRt := runLox(t, src)

    if hadErr || hadRt {
        t.Fatalf("unexpected error flags: hadError=%v, hadRuntimeError=%v", hadErr, hadRt)
    }

    lines := strings.Split(out, "\n")
    want := []string{"1", "2", "3"}
    if len(lines) != len(want) {
        t.Fatalf("expected %d lines, got %d (%q)", len(want), len(lines), out)
    }
    for i, w := range want {
        if strings.TrimSpace(lines[i]) != w {
            t.Errorf("line %d: expected %q, got %q", i, w, lines[i])
        }
    }
}

func TestClassThisAndSuper(t *testing.T) {
    src := `
        class A {
            method() {
                return "A.method";
            }
        }

        class B < A {
            method() {
                return "B.method";
            }

            test() {
                return super.method();
            }
        }

        var b = B();
        print b.method();
        print b.test();
    `
    out, hadErr, hadRt := runLox(t, src)

    if hadErr || hadRt {
        t.Fatalf("unexpected error flags: hadError=%v, hadRuntimeError=%v", hadErr, hadRt)
    }

    lines := strings.Split(out, "\n")
    if len(lines) != 2 {
        t.Fatalf("expected 2 lines, got %d (%q)", len(lines), out)
    }
    if strings.TrimSpace(lines[0]) != "B.method" {
        t.Errorf("expected first line B.method, got %q", lines[0])
    }
    if strings.TrimSpace(lines[1]) != "A.method" {
        t.Errorf("expected second line A.method, got %q", lines[1])
    }
}

func TestUnaryBangAndMinus(t *testing.T) {
    src := `
        print !true;   // false
        print !false;  // true
        print !nil;    // true
        print -3;      // -3
    `
    out, hadErr, hadRt := runLox(t, src)
    if hadErr || hadRt {
        t.Fatalf("unexpected error flags: hadError=%v, hadRuntimeError=%v", hadErr, hadRt)
    }

    lines := strings.Split(out, "\n")
    want := []string{"false", "true", "true", "-3"}
    if len(lines) != len(want) {
        t.Fatalf("expected %d lines, got %d (%q)", len(want), len(lines), out)
    }
    for i, w := range want {
        if strings.TrimSpace(lines[i]) != w {
            t.Errorf("line %d: expected %q, got %q", i, w, lines[i])
        }
    }
}

func TestUnaryMinusTypeError(t *testing.T) {
    src := `print -"hi";`
    out, hadErr, hadRt := runLox(t, src)

    if hadErr {
        t.Fatalf("expected runtime error only, got hadError=true")
    }
    if !hadRt {
        t.Fatalf("expected runtime error for unary minus on string")
    }
    if out != "" {
        t.Errorf("expected no output, got %q", out)
    }
}

func TestComparisonsAndEquality(t *testing.T) {
    src := `
        print 3 > 2;
        print 3 >= 3;
        print 1 < 2;
        print 2 <= 2;
        print 1 == 1;
        print 1 != 2;
        print nil == nil;
        print nil != 1;
    `
    out, hadErr, hadRt := runLox(t, src)
    if hadErr || hadRt {
        t.Fatalf("unexpected error flags: hadError=%v, hadRuntimeError=%v", hadErr, hadRt)
    }

    lines := strings.Split(out, "\n")
    want := []string{"true", "true", "true", "true", "true", "true", "true", "true"}
    if len(lines) != len(want) {
        t.Fatalf("expected %d lines, got %d (%q)", len(want), len(lines), out)
    }
    for i, w := range want {
        if strings.TrimSpace(lines[i]) != w {
            t.Errorf("line %d: expected %q, got %q", i, w, lines[i])
        }
    }
}

func TestStringConcatenation(t *testing.T) {
    src := `print "hi " + "there";`
    out, hadErr, hadRt := runLox(t, src)
    if hadErr || hadRt {
        t.Fatalf("unexpected error flags: hadError=%v, hadRuntimeError=%v", hadErr, hadRt)
    }
    if strings.TrimSpace(out) != "hi there" {
        t.Errorf("expected 'hi there', got %q", out)
    }
}

func TestPlusTypeErrors(t *testing.T) {
    cases := []string{
        `print 1 + "a";`,   // right operand must be a number
        `print "a" + 1;`,   // right operand must be a string
        `print true + true;`, // operands must be two numbers or two strings
    }

    for i, src := range cases {
        out, hadErr, hadRt := runLox(t, src)
        if hadErr {
            t.Fatalf("case %d: expected runtime error only, got hadError=true", i)
        }
        if !hadRt {
            t.Fatalf("case %d: expected runtime error from bad '+' operands", i)
        }
        if out != "" {
            t.Errorf("case %d: expected no output, got %q", i, out)
        }
    }
}

func TestLogicalShortCircuit(t *testing.T) {
    src := `
        var a = false or true;  // evaluate right
        print a;
        var b = true or false;  // short-circuit
        print b;
        var c = false and true; // short-circuit
        print c;
        var d = true and false; // evaluate right
        print d;
    `
    out, hadErr, hadRt := runLox(t, src)
    if hadErr || hadRt {
        t.Fatalf("unexpected error flags: hadError=%v, hadRuntimeError=%v", hadErr, hadRt)
    }

    lines := strings.Split(out, "\n")
    want := []string{"true", "true", "false", "false"}
    if len(lines) != len(want) {
        t.Fatalf("expected %d lines, got %d (%q)", len(want), len(lines), out)
    }
    for i, w := range want {
        if strings.TrimSpace(lines[i]) != w {
            t.Errorf("line %d: expected %q, got %q", i, w, lines[i])
        }
    }
}

func TestWhileLoop(t *testing.T) {
    src := `
        var i = 0;
        while (i < 3) {
            print i;
            i = i + 1;
        }
    `
    out, hadErr, hadRt := runLox(t, src)
    if hadErr || hadRt {
        t.Fatalf("unexpected error flags: hadError=%v, hadRuntimeError=%v", hadErr, hadRt)
    }

    lines := strings.Split(out, "\n")
    want := []string{"0", "1", "2"}
    if len(lines) != len(want) {
        t.Fatalf("expected %d lines, got %d (%q)", len(want), len(lines), out)
    }
    for i, w := range want {
        if strings.TrimSpace(lines[i]) != w {
            t.Errorf("line %d: expected %q, got %q", i, w, lines[i])
        }
    }
}

func TestUndefinedVariableRuntimeError(t *testing.T) {
    src := `print notDefined;`
    out, hadErr, hadRt := runLox(t, src)

    if hadErr {
        t.Fatalf("expected runtime error only, got hadError=true")
    }
    if !hadRt {
        t.Fatalf("expected runtime error for undefined variable")
    }
    if out != "" {
        t.Errorf("expected no output, got %q", out)
    }
}

func TestCallNonCallableRuntimeError(t *testing.T) {
    src := `
        var x = 123;
        x();
    `
    out, hadErr, hadRt := runLox(t, src)

    if hadErr {
        t.Fatalf("expected runtime error only, got hadError=true")
    }
    if !hadRt {
        t.Fatalf("expected runtime error when calling non-callable value")
    }
    if out != "" {
        t.Errorf("expected no output, got %q", out)
    }
}

func TestFunctionArityMismatchRuntimeError(t *testing.T) {
    src := `
        fun f(a, b) {
            print a;
            print b;
        }
        f(1);
    `
    out, hadErr, hadRt := runLox(t, src)

    if hadErr {
        t.Fatalf("expected runtime error only, got hadError=true")
    }
    if !hadRt {
        t.Fatalf("expected runtime error for wrong argument count")
    }
    if out != "" {
        t.Errorf("expected no output (call should fail before prints), got %q", out)
    }
}

func TestInstancePropertiesAndToString(t *testing.T) {
    src := `
        class Foo {
            init(x) {
                this.x = x;
            }
            getX() {
                return this.x;
            }
        }

        var f = Foo(42);
        print f;        // calls instance String()
        print f.getX(); // method call using 'this'
        f.y = 10;       // set field
        print f.y;      // get field
    `
    out, hadErr, hadRt := runLox(t, src)

    if hadErr || hadRt {
        t.Fatalf("unexpected error flags: hadError=%v, hadRuntimeError=%v", hadErr, hadRt)
    }

    lines := strings.Split(out, "\n")
    if len(lines) != 3 {
        t.Fatalf("expected 3 lines, got %d (%q)", len(lines), out)
    }

    if !strings.Contains(strings.TrimSpace(lines[0]), "<Foo instance>") {
        t.Errorf("expected first line to contain '<Foo instance>', got %q", lines[0])
    }
    if strings.TrimSpace(lines[1]) != "42" {
        t.Errorf("expected second line 42, got %q", lines[1])
    }
    if strings.TrimSpace(lines[2]) != "10" {
        t.Errorf("expected third line 10, got %q", lines[2])
    }
}

func TestClassToString(t *testing.T) {
    src := `
        class Bar {}
        print Bar;
    `
    out, hadErr, hadRt := runLox(t, src)
    if hadErr || hadRt {
        t.Fatalf("unexpected error flags: hadError=%v, hadRuntimeError=%v", hadErr, hadRt)
    }
    if !strings.Contains(strings.TrimSpace(out), "<class Bar>") {
        t.Errorf("expected output to contain '<class Bar>', got %q", out)
    }
}

func TestNativeClockFunction(t *testing.T) {
    in := interpreter.NewInterpreter()
    cf := interpreter.ClockFn{}

    if cf.Arity() != 0 {
        t.Fatalf("expected clock arity 0, got %d", cf.Arity())
    }

    t1, ok1 := cf.Call(in, nil).(float64)
    t2, ok2 := cf.Call(in, nil).(float64)
    if !ok1 || !ok2 {
        t.Fatalf("expected clock() to return float64, got %T and %T", t1, t2)
    }
    if t2 < t1 {
        t.Errorf("expected non-decreasing clock values, got t1=%v, t2=%v", t1, t2)
    }

    if s := cf.String(); !strings.Contains(s, "<native fn>") {
        t.Errorf("expected String() to contain '<native fn>', got %q", s)
    }
}
