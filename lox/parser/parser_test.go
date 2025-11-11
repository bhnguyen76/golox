package parser

import (
    "testing"

    "example.com/golox/lox/ast"
    "example.com/golox/lox/scanner"
    "example.com/golox/lox/shared"
)

func scanAndParse(t *testing.T, src string) []ast.Stmt {
    t.Helper()
    shared.ResetErrors()

    s := scanner.NewScanner(src)
    tokens := s.ScanTokens()

    p := NewParser(tokens)
    stmts := p.Parse()

    if shared.HadError {
        t.Fatalf("parse reported HadError for source: %q", src)
    }
    return stmts
}

func TestVarDeclarationParses(t *testing.T) {
    src := `var a = 123;`
    stmts := scanAndParse(t, src)

    if len(stmts) != 1 {
        t.Fatalf("expected 1 statement, got %d", len(stmts))
    }

    v, ok := stmts[0].(*ast.Var)
    if !ok {
        t.Fatalf("expected *ast.Var, got %T", stmts[0])
    }

    if v.Name.Lexeme != "a" {
        t.Errorf("expected var name 'a', got %q", v.Name.Lexeme)
    }

    lit, ok := v.Initializer.(*ast.Literal)
    if !ok {
        t.Fatalf("expected initializer to be *ast.Literal, got %T", v.Initializer)
    }

    if lit.Value != float64(123) { 
        t.Errorf("expected literal value 123, got %#v", lit.Value)
    }
}

func TestBinaryPrecedence(t *testing.T) {
    src := `1 + 2 * 3;`
    stmts := scanAndParse(t, src)

    if len(stmts) != 1 {
        t.Fatalf("expected 1 statement, got %d", len(stmts))
    }

    exprStmt, ok := stmts[0].(*ast.Expression)
    if !ok {
        t.Fatalf("expected *ast.Expression, got %T", stmts[0])
    }

    plus, ok := exprStmt.Expression.(*ast.Binary)
    if !ok {
        t.Fatalf("expected top expr to be *ast.Binary, got %T", exprStmt.Expression)
    }

    if plus.Operator.Type != scanner.PLUS {
        t.Errorf("expected top operator '+', got %v", plus.Operator.Type)
    }

    leftLit, ok := plus.Left.(*ast.Literal)
    if !ok || leftLit.Value != float64(1) {
        t.Errorf("expected left literal 1, got %T (%#v)", plus.Left, leftLit)
    }

    mul, ok := plus.Right.(*ast.Binary)
    if !ok {
        t.Fatalf("expected right expr to be *ast.Binary, got %T", plus.Right)
    }

    if mul.Operator.Type != scanner.STAR {
        t.Errorf("expected '*' as inner operator, got %v", mul.Operator.Type)
    }

    rLeft, ok := mul.Left.(*ast.Literal)
    if !ok || rLeft.Value != float64(2) {
        t.Errorf("expected left inner literal 2, got %T (%#v)", mul.Left, mul.Left)
    }

    rRight, ok := mul.Right.(*ast.Literal)
    if !ok || rRight.Value != float64(3) {
        t.Errorf("expected right inner literal 3, got %T (%#v)", mul.Right, mul.Right)
    }
}

func TestIfElseParses(t *testing.T) {
    src := `
        if (true) print 1;
        else print 2;
    `
    stmts := scanAndParse(t, src)

    if len(stmts) != 1 {
        t.Fatalf("expected 1 statement, got %d", len(stmts))
    }

    ifStmt, ok := stmts[0].(*ast.If)
    if !ok {
        t.Fatalf("expected *ast.If, got %T", stmts[0])
    }

    if _, ok := ifStmt.Condition.(*ast.Literal); !ok {
        t.Errorf("expected condition to be *ast.Literal, got %T", ifStmt.Condition)
    }

    if _, ok := ifStmt.ThenBranch.(*ast.Print); !ok {
        t.Errorf("expected then branch to be *ast.Print, got %T", ifStmt.ThenBranch)
    }

    if _, ok := ifStmt.ElseBranch.(*ast.Print); !ok {
        t.Errorf("expected else branch to be *ast.Print, got %T", ifStmt.ElseBranch)
    }
}

func TestWhileParses(t *testing.T) {
    src := `
        while (i < 10) print i;
    `
    stmts := scanAndParse(t, src)

    if len(stmts) != 1 {
        t.Fatalf("expected 1 statement, got %d", len(stmts))
    }

    w, ok := stmts[0].(*ast.While)
    if !ok {
        t.Fatalf("expected *ast.While, got %T", stmts[0])
    }

    if _, ok := w.Condition.(*ast.Binary); !ok {
        t.Errorf("expected condition to be *ast.Binary, got %T", w.Condition)
    }

    if _, ok := w.Body.(*ast.Print); !ok {
        t.Errorf("expected body to be *ast.Print, got %T", w.Body)
    }
}

func TestForDesugarsToWhileWithInitializerAndIncrement(t *testing.T) {
    src := `
        for (var i = 0; i < 3; i = i + 1) print i;
    `
    stmts := scanAndParse(t, src)

    if len(stmts) != 1 {
        t.Fatalf("expected 1 top-level statement, got %d", len(stmts))
    }

    block, ok := stmts[0].(*ast.Block)
    if !ok {
        t.Fatalf("expected top-level *ast.Block from desugared for, got %T", stmts[0])
    }

    if len(block.Statements) != 2 {
        t.Fatalf("expected Block with 2 statements (initializer, while), got %d", len(block.Statements))
    }

    if _, ok := block.Statements[0].(*ast.Var); !ok {
        t.Errorf("expected first stmt to be *ast.Var (initializer), got %T", block.Statements[0])
    }

    w, ok := block.Statements[1].(*ast.While)
    if !ok {
        t.Fatalf("expected second stmt to be *ast.While, got %T", block.Statements[1])
    }

    innerBlock, ok := w.Body.(*ast.Block)
    if !ok {
        t.Fatalf("expected while body to be *ast.Block, got %T", w.Body)
    }

    if len(innerBlock.Statements) != 2 {
        t.Fatalf("expected inner Block with 2 statements (body, increment), got %d", len(innerBlock.Statements))
    }

    if _, ok := innerBlock.Statements[0].(*ast.Print); !ok {
        t.Errorf("expected first inner stmt to be *ast.Print, got %T", innerBlock.Statements[0])
    }
    if _, ok := innerBlock.Statements[1].(*ast.Expression); !ok {
        t.Errorf("expected second inner stmt to be *ast.Expression (increment), got %T", innerBlock.Statements[1])
    }
}

func TestFunctionDeclarationParses(t *testing.T) {
    src := `
        fun add(a, b) {
            return a + b;
        }
    `
    stmts := scanAndParse(t, src)

    if len(stmts) != 1 {
        t.Fatalf("expected 1 statement, got %d", len(stmts))
    }

    fn, ok := stmts[0].(*ast.Function)
    if !ok {
        t.Fatalf("expected *ast.Function, got %T", stmts[0])
    }

    if fn.Name.Lexeme != "add" {
        t.Errorf("expected function name 'add', got %q", fn.Name.Lexeme)
    }

    if len(fn.Params) != 2 {
        t.Errorf("expected 2 parameters, got %d", len(fn.Params))
    }

    if len(fn.Body) != 1 {
        t.Fatalf("expected function body with 1 statement, got %d", len(fn.Body))
    }

    if _, ok := fn.Body[0].(*ast.Return); !ok {
        t.Errorf("expected body[0] to be *ast.Return, got %T", fn.Body[0])
    }
}

func TestClassDeclarationParses(t *testing.T) {
    src := `
        class Foo < Bar {
            init(x) { this.x = x; }
            getX() { return this.x; }
        }
    `
    stmts := scanAndParse(t, src)

    if len(stmts) != 1 {
        t.Fatalf("expected 1 statement, got %d", len(stmts))
    }

    cls, ok := stmts[0].(*ast.Class)
    if !ok {
        t.Fatalf("expected *ast.Class, got %T", stmts[0])
    }

    if cls.Name.Lexeme != "Foo" {
        t.Errorf("expected class name 'Foo', got %q", cls.Name.Lexeme)
    }

    if cls.Superclass == nil {
        t.Fatalf("expected superclass, got nil")
    }
    if superVar, ok := cls.Superclass.(*ast.Variable); !ok || superVar.Name.Lexeme != "Bar" {
        t.Errorf("expected superclass Variable 'Bar', got %T (%#v)", cls.Superclass, cls.Superclass)
    }

    if len(cls.Methods) != 2 {
        t.Fatalf("expected 2 methods, got %d", len(cls.Methods))
    }
}

func TestLogicalPrecedence(t *testing.T) {
    src := `true or false and true;`
    stmts := scanAndParse(t, src)

    if len(stmts) != 1 {
        t.Fatalf("expected 1 statement, got %d", len(stmts))
    }

    exprStmt, ok := stmts[0].(*ast.Expression)
    if !ok {
        t.Fatalf("expected *ast.Expression, got %T", stmts[0])
    }

    orExpr, ok := exprStmt.Expression.(*ast.Logical)
    if !ok {
        t.Fatalf("expected top expr to be *ast.Logical (or), got %T", exprStmt.Expression)
    }

    if orExpr.Operator.Type != scanner.OR {
        t.Errorf("expected top operator OR, got %v", orExpr.Operator.Type)
    }

    andExpr, ok := orExpr.Right.(*ast.Logical)
    if !ok {
        t.Fatalf("expected right side of OR to be *ast.Logical (and), got %T", orExpr.Right)
    }

    if andExpr.Operator.Type != scanner.AND {
        t.Errorf("expected inner operator AND, got %v", andExpr.Operator.Type)
    }
}

func TestCallAndGetParse(t *testing.T) {
    src := `foo.bar(1, 2);`
    stmts := scanAndParse(t, src)

    if len(stmts) != 1 {
        t.Fatalf("expected 1 statement, got %d", len(stmts))
    }

    exprStmt, ok := stmts[0].(*ast.Expression)
    if !ok {
        t.Fatalf("expected *ast.Expression, got %T", stmts[0])
    }

    call, ok := exprStmt.Expression.(*ast.Call)
    if !ok {
        t.Fatalf("expected expr to be *ast.Call, got %T", exprStmt.Expression)
    }

    get, ok := call.Callee.(*ast.Get)
    if !ok {
        t.Fatalf("expected callee to be *ast.Get, got %T", call.Callee)
    }

    if get.Name.Lexeme != "bar" {
        t.Errorf("expected property name 'bar', got %q", get.Name.Lexeme)
    }

    if v, ok := get.Object.(*ast.Variable); !ok || v.Name.Lexeme != "foo" {
        t.Errorf("expected object Variable 'foo', got %T (%#v)", get.Object, get.Object)
    }

    if len(call.Arguments) != 2 {
        t.Errorf("expected 2 arguments, got %d", len(call.Arguments))
    }
}

func scanAndParseAllowError(t *testing.T, src string) ([]ast.Stmt, bool) {
    t.Helper()
    shared.ResetErrors()

    s := scanner.NewScanner(src)
    tokens := s.ScanTokens()

    p := NewParser(tokens)
    stmts := p.Parse()
    return stmts, shared.HadError
}

func TestBadVarDeclarationThenRecovery(t *testing.T) {
    src := `
        var a = ;
        print 1;
    `
    stmts, hadErr := scanAndParseAllowError(t, src)

    if !hadErr {
        t.Fatalf("expected HadError to be true for invalid var declaration")
    }

    if len(stmts) == 0 {
        t.Fatalf("expected at least one statement after recovery")
    }

    if _, ok := stmts[len(stmts)-1].(*ast.Print); !ok {
        t.Errorf("expected last statement to be *ast.Print, got %T", stmts[len(stmts)-1])
    }
}

