# Makefile for golox

BIN_DIR := bin
BINARY  := $(BIN_DIR)/glox

.PHONY: all
all: build

# Build the glox binary
.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY) .

# Run the REPL
.PHONY: repl
repl: build
	$(BINARY)

# Run a single script: make run-script SCRIPT=examples/01_smoke_all_features.lox
.PHONY: run-script
run-script: build
	@if [ -z "$(SCRIPT)" ]; then \
		echo "Usage: make run-script SCRIPT=path/to/file.lox"; \
		exit 1; \
	fi
	$(BINARY) "$(SCRIPT)"

# Run all .lox examples in the examples/ directory. Need to treat errors.lox differently so it doesn't crash the test run
.PHONY: examples
examples: build
	@if [ -z "$$(ls examples/*.lox 2>/dev/null)" ]; then \
		echo "No .lox files found in examples/"; \
		exit 1; \
	fi
	@for f in examples/*.lox; do \
		echo "=== Running $$f ==="; \
		case "$$f" in \
			*errors.lox) \
				echo "(expecting errors and non-zero exit code)"; \
				$(BINARY) "$$f" >/dev/null 2>&1 || true; \
				;; \
			*) \
				$(BINARY) "$$f" || exit $$?; \
				;; \
		esac; \
		echo; \
	done
	@echo "All examples finished."


# Run all Go tests
.PHONY: test
test:
	go test ./...

# Run tests with coverage
.PHONY: cover
cover:
	go test -cover ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
