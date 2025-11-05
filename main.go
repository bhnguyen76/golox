package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"example.com/golox/lox/scanner"
	"example.com/golox/lox/shared"
)

// var hadError bool

func main() {
	args := os.Args[1:]
	if len(args) > 1 {
		fmt.Println("Usage: glox [script]")
	} else if len(args) == 1 {
		sourcePath, err := filepath.Abs(os.Args[1])
		if err != nil {
			fmt.Println("Error running file")
			os.Exit(-1)
		}
		runFile(sourcePath)
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

	// For now, just print the tokens.
	for token := range tokens {
		fmt.Println(token) 
	}
	return nil
}

// func errorAt(line int, message string) {
// 	report(line, "",message)
// }

// func report(line int, where string, message string) {
// 	fmt.Fprint(os.Stderr, "[Line %d] Error%s: %s\n", line, where, message)
// 	hadError = true
// }