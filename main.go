package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"example.com/golox/lox/interpreter"
	"example.com/golox/lox/parser"
	"example.com/golox/lox/resolver"
	"example.com/golox/lox/scanner"
	"example.com/golox/lox/shared"
)

const (
	exitUsage       = 64
	exitDataError   = 65 // hadError
	exitRuntimeError = 70 // hadRuntimeError
)

var interp = interpreter.NewInterpreter()

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
	sc := scanner.NewScanner(source)
	tokens := sc.ScanTokens()

	p := parser.NewParser(tokens)
	statements := p.Parse()

	if shared.HadError || statements == nil {
		return nil
	}

	res := resolver.NewResolver(interp)
	res.Resolve(statements)

	if shared.HadError {
		return nil
	}

	interp.Interpret(statements)
	return nil
}

