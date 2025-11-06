package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"example.com/golox/lox/ast"
	"example.com/golox/lox/interpreter"
	"example.com/golox/lox/parser"
	"example.com/golox/lox/scanner"
	"example.com/golox/lox/shared"
)

const (
	exitUsage       = 64
	exitDataError   = 65 // hadError
	exitRuntimeError = 70 // hadRuntimeError
)

var interp = &interpreter.Interpreter{}

func main() {
	args := os.Args[1:]
	if len(args) > 1 {
		fmt.Println("Usage: glox [script]")
	} else if len(args) == 1 {
		shared.ResetErrors()
		if err := runFile(args[0]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(exitDataError)
		}
		
		if shared.HadError {
			os.Exit(exitDataError) 
		}
		if shared.HadRuntimeError {
			os.Exit(exitRuntimeError) 
		}
	} else {
		runPrompt()
	}
}

func runFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", path, err)
	}
	return run(string(data))
}

func runPrompt() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err == io.EOF { fmt.Println(); return}
		if err != nil { fmt.Fprintln(os.Stderr, "Read error: ", err); return}

		shared.ResetErrors()
		if err := run(line); err != nil {
			fmt.Fprintln(os.Stderr, "Error: ", err)
		}
	}
}

func run(source string) error {
	scanner := scanner.NewScanner(source)
	tokens := scanner.ScanTokens()

	parser := parser.NewParser(tokens)
	expression := parser.Parse()

	if shared.HadError || expression == nil {
		return nil
	}

	interp.Interpret(expression)
	return nil
}
