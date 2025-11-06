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

	// TODO: implement the generator here, e.g.:
	defineAst(outputDir, "Expr", []string{
	    "Binary   : Expr left, Token operator, Expr right",
	    "Grouping : Expr expression",
	    "Literal  : any value",
	    "Unary    : Token operator, Expr right",
	})
}

func defineAst(outputDir, baseName string, types []string) error {
	filename := strings.ToLower(baseName) + ".go" // expr.go, stmt.go, etc.
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
	fmt.Fprintf(w, "type %s interface {\n", baseName)
	fmt.Fprintln(w, "\tAccept(v Visitor) any")
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
	// e.g. type Visitor interface { VisitBinaryExpr(*Binary) any; ... }
	fmt.Fprintln(w, "type Visitor interface {")
	for _, t := range types {
		typeName := strings.TrimSpace(strings.Split(t, ":")[0])
		// VisitBinaryExpr(*Binary) any
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
	recv := strings.ToLower(className[:1]) // n, g, l, u, etc.
	fmt.Fprintf(w, "func (%s *%s) Accept(v Visitor) any {\n", recv, className)
	fmt.Fprintf(w, "\treturn v.Visit%s%s(%s)\n", className, baseName, recv)
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
	default:
		// you can extend this later for List<Stmt> etc.
		return t
	}
}


