package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const exitUsage = 64

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: generate_ast <output directory>")
		os.Exit(exitUsage)
	}

	outputDir := args[0]

	if err := defineAst(outputDir, "Expr", []string{
		"Assign   : Token name, Expr value",
		"Binary   : Expr left, Token operator, Expr right",
		"Call     : Expr callee, Token paren, List<Expr> arguments",
		"Get      : Expr object, Token name",
		"Grouping : Expr expression",
		"Literal  : any value",
		"Logical  : Expr left, Token operator, Expr right",
		"Set      : Expr object, Token name, Expr value",
		"This     : Token keyword",
		"Unary    : Token operator, Expr right",
		"Variable : Token name",
	}); err != nil {
		fmt.Fprintln(os.Stderr, "generate_ast error:", err)
		os.Exit(1)
	}

	if err := defineAst(outputDir, "Stmt", []string{
		"Block      : List<Stmt> statements",
		"Class      : Token name, List<Function> methods",
		"Expression : Expr expression",
		"Function	: Token name, List<Token> params," +
					" List<Stmt> body",
		"Print      : Expr expression",
		"If         : Expr condition, Stmt thenBranch," + " Stmt elseBranch",
		"Return		: Token keyword, Expr value",
		"Var		: Token name, Expr initializer",
		"While      : Expr condition, Stmt body",
	}); err != nil {
		fmt.Fprintln(os.Stderr, "generate_ast error:", err)
		os.Exit(1)
	}
}

func defineAst(outputDir, baseName string, types []string) error {
	filename := strings.ToLower(baseName) + ".go" 
	path := filepath.Join(outputDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	// TODO: adjust module path to match your go.mod!
	fmt.Fprintln(w, "package ast")
	fmt.Fprintln(w)
	fmt.Fprintln(w, `import "example.com/golox/lox/scanner"`)
	fmt.Fprintln(w)

	// Base interface: e.g. "type Expr interface { Accept(v Visitor) any }"
    visitorName := baseName + "Visitor" // ExprVisitor, StmtVisitor

    fmt.Fprintf(w, "type %s interface {\n", baseName)
    fmt.Fprintf(w, "\tAccept(v %s) any\n", visitorName)
    fmt.Fprintln(w, "}")
    fmt.Fprintln(w)

	// Visitor interface: VisitBinaryExpr(*Binary) any, etc.
	if err := defineVisitor(w, baseName, types); err != nil {
		return err
	}

	// Concrete AST node types.
	for _, t := range types {
		parts := strings.Split(t, ":")
		if len(parts) != 2 {
			return fmt.Errorf("bad type spec: %q", t)
		}
		className := strings.TrimSpace(parts[0])
		fieldList := strings.TrimSpace(parts[1])

		if err := defineType(w, baseName, className, fieldList); err != nil {
			return err
		}
	}

	return nil
}

func defineVisitor(w *bufio.Writer, baseName string, types []string) error {
    visitorName := baseName + "Visitor" // ExprVisitor or StmtVisitor

    fmt.Fprintf(w, "type %s interface {\n", visitorName)
    for _, t := range types {
        typeName := strings.TrimSpace(strings.Split(t, ":")[0])
        // e.g. VisitBinaryExpr(*Binary) any  or VisitPrintStmt(*Print) any
        fmt.Fprintf(w, "\tVisit%s%s(*%s) any\n", typeName, baseName, typeName)
    }
    fmt.Fprintln(w, "}")
    fmt.Fprintln(w)
    return nil
}

func defineType(w *bufio.Writer, baseName, className, fieldList string) error {
	// struct header: type Binary struct { ... }
	fmt.Fprintf(w, "type %s struct {\n", className)

	fields := strings.Split(fieldList, ", ")
	for _, f := range fields {
		parts := strings.Split(f, " ")
		if len(parts) != 2 {
			return fmt.Errorf("bad field: %q", f)
		}
		fieldType := strings.TrimSpace(parts[0])
		fieldName := strings.TrimSpace(parts[1])

		exportedName := strings.ToUpper(fieldName[:1]) + fieldName[1:]
		goType := mapFieldTypeToGo(fieldType)

		fmt.Fprintf(w, "\t%s %s\n", exportedName, goType)
	}
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w)

	// Accept method, e.g.:
	// func (n *Binary) Accept(v Visitor) any {
	//     return v.VisitBinaryExpr(n)
	// }
    visitorName := baseName + "Visitor" // ExprVisitor or StmtVisitor

    fmt.Fprintf(w, "func (n *%s) Accept(v %s) any {\n", className, visitorName)
    fmt.Fprintf(w, "\treturn v.Visit%s%s(n)\n", className, baseName)
    fmt.Fprintln(w, "}")
    fmt.Fprintln(w)

	return nil
}

func mapFieldTypeToGo(t string) string {
	switch t {
	case "Expr":
		return "Expr"
	case "Token":
		return "scanner.Token"
	case "Object", "any":
		return "any"
	case "List<Stmt>":
		return "[]Stmt"
	case "List<Expr>":
		return "[]Expr"
	case "List<Token>":
		return "[]scanner.Token"
	case "List<Function>":
		return "[]*Function"
	default:
		return t
	}
}


