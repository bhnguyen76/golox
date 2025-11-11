# goLox
Complete implementation of lox interpreter (Chapters 4-13)
/root
├── makefile – Defines make commands to build, test, and run the interpreter
├── README.md – Directory structure and make command descriptions
├── DOCUMENTATION.txt – Detailed testing plan, results, and limitations
├── bin/ – Compiled binary output (created after build)
├── examples/ – Lox stress tests (features, classes, control flow, etc.)
│   ├── classes_inheritance.lox
│   ├── control_flow.lox
│   ├── errors.lox
│   ├── features.lox
│   └── functions.lox
└── lox/
    ├── ast/ – Abstract Syntax Tree definitions
    │   ├── ast_printer.go
    │   ├── ast_printer_test.go
    │   ├── expr.go
    │   └── stmt.go
    │
    ├── scanner/ – Scans source text into tokens
    │   ├── scanner.go
    │   ├── token.go
    │   ├── keywords.go
    │   └── scanner_test.go
    │
    ├── parser/ – Parses tokens into an AST
    │   ├── parser.go
    │   └── parser_test.go
    │
    ├── resolver/ – Handles variable and class scope resolution
    │   ├── resolver.go
    │   └── resolver_test.go
    │
    ├── interpreter/ – Executes AST nodes at runtime
    │   ├── interpreter.go
    │   ├── environment.go
    │   ├── callable.go
    │   ├── native.go
    │   ├── class.go
    │   ├── instance.go
    │   ├── function.go
    │   └── interpreter_test.go
    │
    └── shared/ – Shared error reporting and runtime flags
        ├── report.go
        └── shared_test.go

# Build Process

To compile interpreter:
    make build

To run the REPL:
    make repl

To run all Go unit tests:
    make test

To run tests with coverage report:
    make cover

To run a single Lox script:
    make run-script SCRIPT=... (location of script Ex. 'examples/features.lox')

To run all example stress tests:
    make examples

To clean up build artifacts:
    make clean

Notes:
To exit reply, ctrl+D