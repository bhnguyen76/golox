package shared

import (
	"fmt"
	"os"
)

// HadError is set to true when any error is reported.
var HadError bool
var HadRuntimeError bool

// ErrorAt reports an error at a given line with a message.
func ErrorAt(line int, message string) {
	Report(line, "", message)
}

// Report prints a formatted error message and marks HadError.
func Report(line int, where string, message string) {
	fmt.Fprintf(os.Stderr, "[Line %d] Error%s: %s\n", line, where, message)
	HadError = true
}

func ResetErrors() {
    HadError = false
    HadRuntimeError = false
}
